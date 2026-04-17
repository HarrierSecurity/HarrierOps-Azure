package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

type persistenceLogicAppStepDefinition struct {
	Action     string
	APISurface string
}

var persistenceLogicAppSteps = []persistenceLogicAppStepDefinition{
	{Action: "create or modify workflow", APISurface: "Microsoft.Logic/workflows"},
	{Action: "edit workflow definition", APISurface: "workflow definition"},
	{Action: "attach or reuse exec ctx", APISurface: "workflow identity / connections"},
	{Action: "define or modify trigger", APISurface: "request, recurrence, or event trigger"},
	{Action: "enable workflow", APISurface: "workflow state"},
	{Action: "add or repurpose downstream actions", APISurface: "workflow actions / connectors"},
}

func buildPersistenceLogicAppsOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.PersistenceSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	logicAppsFuture := runGroupedCommandOutput[models.LogicAppsOutput](group, ctx, request, logicAppsHandler(provider, now), "logic-apps")
	permissionsFuture := runGroupedCommandOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions")
	rbacFuture := runGroupedCommandOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac")

	logicApps, err := logicAppsFuture.wait()
	if err != nil {
		return nil, err
	}
	permissions, err := permissionsFuture.wait()
	if err != nil {
		return nil, err
	}
	rbac, err := rbacFuture.wait()
	if err != nil {
		return nil, err
	}

	subscriptionID := firstNonEmpty(
		request.Subscription,
		stringPtrValue(logicApps.Metadata.SubscriptionID),
		stringPtrValue(permissions.Metadata.SubscriptionID),
	)
	tenantID := firstNonEmpty(
		request.Tenant,
		stringPtrValue(logicApps.Metadata.TenantID),
		stringPtrValue(permissions.Metadata.TenantID),
	)

	permissionsByPrincipal := make(map[string]models.PermissionRow, len(permissions.Permissions))
	currentIdentity := models.PermissionRow{}
	currentIdentityVisible := false
	for _, permission := range permissions.Permissions {
		if permission.PrincipalID == "" {
			continue
		}
		permissionsByPrincipal[permission.PrincipalID] = permission
		if permission.IsCurrentIdentity && !currentIdentityVisible {
			currentIdentity = permission
			currentIdentityVisible = true
		}
	}

	currentIdentityAssignments := make([]models.RoleAssignment, 0)
	for _, assignment := range rbac.RoleAssignments {
		if currentIdentityVisible && assignment.PrincipalID == currentIdentity.PrincipalID {
			currentIdentityAssignments = append(currentIdentityAssignments, assignment)
		}
	}

	workflows := sortedByLess(logicApps.Workflows, logicAppLess)
	rows := make([]models.PersistenceLogicAppWorkflow, 0, len(workflows))
	for _, workflow := range workflows {
		control, controlOK := persistenceAutomationControl(workflow.ID, currentIdentityAssignments)
		currentContext := persistenceCurrentIdentityContext(currentIdentity, control, controlOK)
		capabilitySteps := persistenceLogicAppCapabilitySteps(controlOK)
		executionContextOptions := persistenceLogicAppExecutionContextOptions(workflow)
		strongestContext, strongestContextHasAzureControl := persistenceLogicAppExecutionContext(workflow, permissionsByPrincipal)
		nearbyNames := persistenceLogicAppNearbyNames(workflows, workflow.Name)

		rows = append(rows, models.PersistenceLogicAppWorkflow{
			ID:                      workflow.ID,
			Name:                    workflow.Name,
			ResourceGroup:           workflow.ResourceGroup,
			Location:                workflow.Location,
			CapabilitySteps:         capabilitySteps,
			CurrentIdentityContext:  currentContext,
			ExecutionContextOptions: executionContextOptions,
			CurrentState: models.PersistenceLogicAppState{
				Classification:                   workflow.Classification,
				Platform:                         workflow.Platform,
				WorkflowKind:                     workflow.WorkflowKind,
				State:                            workflow.State,
				TriggerTypes:                     append([]string{}, workflow.TriggerTypes...),
				ExternallyCallableRequestTrigger: workflow.ExternallyCallableRequestTrigger,
				RecurrenceSummary:                workflow.RecurrenceSummary,
				IdentityType:                     workflow.IdentityType,
				StrongestVisibleExecutionContext: strongestContext,
				NearbyThematicNames:              nearbyNames,
				DownstreamActionKinds:            append([]string{}, workflow.DownstreamActionKinds...),
			},
			StillUnmapped: persistenceLogicAppStillUnmapped(workflow),
			Summary:       persistenceLogicAppSummary(workflow, controlOK, strongestContext, strongestContextHasAzureControl),
			RelatedIDs:    mergeRelatedIDs(workflow.RelatedIDs),
		})
	}

	issues := append([]models.Issue{}, logicApps.Issues...)
	issues = append(issues, permissions.Issues...)
	issues = append(issues, rbac.Issues...)

	return models.PersistenceLogicAppsOutput{
		Metadata:           scopedMetadata(now, request, tenantID, subscriptionID, "persistence"),
		GroupedCommandName: "persistence",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		Workflows:          rows,
		Issues:             issues,
	}, nil
}

func persistenceLogicAppNearbyNames(
	workflows []models.LogicAppWorkflowAsset,
	currentWorkflowName string,
) []string {
	seen := map[string]struct{}{}
	candidates := []string{}
	for _, workflow := range workflows {
		name := strings.TrimSpace(workflow.Name)
		if name == "" || strings.EqualFold(name, currentWorkflowName) {
			continue
		}
		if persistenceAutomationNameScore(name) == 0 {
			continue
		}
		if _, ok := seen[strings.ToLower(name)]; ok {
			continue
		}
		seen[strings.ToLower(name)] = struct{}{}
		candidates = append(candidates, name)
	}
	sort.SliceStable(candidates, func(i int, j int) bool {
		leftScore := persistenceAutomationNameScore(candidates[i])
		rightScore := persistenceAutomationNameScore(candidates[j])
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return false
	})
	if len(candidates) > 4 {
		return append([]string{}, candidates[:4]...)
	}
	return candidates
}

func persistenceLogicAppCapabilitySteps(controlOK bool) []models.PersistenceCapabilityStep {
	steps := make([]models.PersistenceCapabilityStep, 0, len(persistenceLogicAppSteps))
	for _, step := range persistenceLogicAppSteps {
		status := "not proven"
		if controlOK {
			status = "yes"
		}
		steps = append(steps, models.PersistenceCapabilityStep{
			Action:     step.Action,
			APISurface: step.APISurface,
			Status:     status,
		})
	}
	return steps
}

func persistenceLogicAppExecutionContextOptions(workflow models.LogicAppWorkflowAsset) []string {
	options := []string{}
	if strings.TrimSpace(stringPtrValue(workflow.IdentityType)) != "" {
		options = append(options, "managed identity")
	}
	if persistenceContainsString(workflow.DownstreamActionKinds, "connector") {
		options = append(options, "connector-backed actions")
	}
	return dedupeStrings(options)
}

func persistenceLogicAppExecutionContext(
	workflow models.LogicAppWorkflowAsset,
	permissionsByPrincipal map[string]models.PermissionRow,
) (*models.PersistenceRoleContext, bool) {
	if workflow.PrincipalID == nil || strings.TrimSpace(*workflow.PrincipalID) == "" {
		return nil, false
	}

	name := persistenceLogicAppIdentityName(workflow)
	permission, ok := permissionsByPrincipal[*workflow.PrincipalID]
	if !ok {
		return &models.PersistenceRoleContext{
			Name:         name,
			Kind:         "logic-app-execution-context",
			PrincipalID:  workflow.PrincipalID,
			IdentityType: workflow.IdentityType,
			RoleNames:    []string{},
			ScopeIDs:     []string{},
			Summary:      fmt.Sprintf("Logic App identity `%s` is visible here, but no matching Azure role context is confirmed from current scope.", name),
		}, false
	}

	roleNames := append([]string{}, permission.HighImpactRoles...)
	if len(roleNames) == 0 {
		roleNames = append(roleNames, permission.AllRoleNames...)
	}

	summary := fmt.Sprintf("The strongest visible execution context here is the Logic App identity `%s`, which already holds %s.", name, persistenceRoleSummary(roleNames, permission.ScopeIDs))
	if !permission.Privileged {
		summary = fmt.Sprintf("Logic App identity `%s` is visible here, but no high-impact Azure role assignments are confirmed from current scope.", name)
	}

	return &models.PersistenceRoleContext{
		Name:         name,
		Kind:         "logic-app-execution-context",
		PrincipalID:  workflow.PrincipalID,
		IdentityType: workflow.IdentityType,
		RoleNames:    dedupeStrings(roleNames),
		ScopeIDs:     dedupeStrings(permission.ScopeIDs),
		Summary:      summary,
	}, permission.Privileged
}

func persistenceLogicAppIdentityName(workflow models.LogicAppWorkflowAsset) string {
	if strings.Contains(strings.ToLower(stringPtrValue(workflow.IdentityType)), "userassigned") && len(workflow.IdentityIDs) > 0 {
		return persistenceResourceNameFromID(workflow.IdentityIDs[0])
	}
	return firstNonEmpty(workflow.Name+"-identity", "logic-app identity")
}

func persistenceLogicAppStillUnmapped(workflow models.LogicAppWorkflowAsset) []string {
	items := []string{
		"the exact workflow definition, connector secret material, or operator intent behind this Logic App",
		"the exact callback URL, access signature, or upstream caller identity behind the visible trigger posture",
		"the exact downstream payloads, connection credentials, or whether each visible action category reaches a high-value target",
	}
	if !persistenceLogicAppIsDurable(workflow) {
		items = append(items, "whether a later definition change would turn this workflow into durable request or recurrence-backed re-entry")
	}
	return dedupeStrings(items)
}

func persistenceLogicAppSummary(
	workflow models.LogicAppWorkflowAsset,
	controlOK bool,
	strongestContext *models.PersistenceRoleContext,
	strongestContextHasAzureControl bool,
) string {
	if controlOK && persistenceLogicAppIsDurable(workflow) && strongestContext != nil && strongestContextHasAzureControl {
		return fmt.Sprintf("Current identity can set up Logic App '%s' as durable workflow persistence, and the strongest visible execution context already carries Azure control.", workflow.Name)
	}
	if controlOK && persistenceLogicAppIsDurable(workflow) {
		return fmt.Sprintf("Current identity can set up Logic App '%s' as durable workflow persistence from current RBAC evidence.", workflow.Name)
	}
	if controlOK {
		return fmt.Sprintf("Current identity can build or repurpose Logic App '%s', but the current workflow definition does not yet show a durable request or recurrence trigger.", workflow.Name)
	}
	if persistenceLogicAppIsDurable(workflow) {
		return fmt.Sprintf("Logic App '%s' already shows durable trigger posture, but the current identity does not yet have a proven path to set up or repurpose it here.", workflow.Name)
	}
	return fmt.Sprintf("Logic App '%s' is visible, but the current identity does not yet have a proven path to set it up as durable workflow persistence.", workflow.Name)
}

func persistenceLogicAppIsDurable(workflow models.LogicAppWorkflowAsset) bool {
	state := strings.TrimSpace(strings.ToLower(stringPtrValue(workflow.State)))
	if state != "" && state != "enabled" {
		return false
	}
	if workflow.Classification == "persistence-capable" {
		return true
	}
	if workflow.ExternallyCallableRequestTrigger {
		return true
	}
	return strings.TrimSpace(stringPtrValue(workflow.RecurrenceSummary)) != ""
}

func persistenceContainsString(values []string, needle string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), needle) {
			return true
		}
	}
	return false
}

func persistenceResourceNameFromID(resourceID string) string {
	parts := strings.Split(strings.TrimRight(strings.TrimSpace(resourceID), "/"), "/")
	for index := len(parts) - 1; index >= 0; index-- {
		if strings.TrimSpace(parts[index]) != "" {
			return parts[index]
		}
	}
	return "unknown"
}
