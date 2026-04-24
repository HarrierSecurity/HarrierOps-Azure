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

	evidence := buildPersistencePrincipalEvidence(permissions.Permissions, rbac.RoleAssignments)

	workflows := sortedByLess(logicApps.Workflows, logicAppLess)
	rows := make([]models.PersistenceLogicAppWorkflow, 0, len(workflows))
	for _, workflow := range workflows {
		control, controlOK := persistenceAutomationControl(workflow.ID, evidence.currentIdentityAssignments)
		currentContext := persistenceCurrentIdentityContext(evidence.currentIdentity, control, controlOK)
		capabilitySteps := persistenceLogicAppCapabilitySteps(controlOK)
		executionContextOptions := persistenceLogicAppExecutionContextOptions(workflow)
		strongestContext, strongestContextHasAzureControl := persistenceLogicAppExecutionContext(workflow, evidence.permissionsByPrincipal, evidence.assignmentsByPrincipal)
		nearbyNames := persistenceLogicAppNearbyNames(workflows, workflow.Name)

		rows = append(rows, models.PersistenceLogicAppWorkflow{
			ID:                      workflow.ID,
			Name:                    workflow.Name,
			ResourceGroup:           workflow.ResourceGroup,
			Location:                workflow.Location,
			CapabilitySteps:         capabilitySteps,
			CurrentIdentityContext:  currentContext,
			ExecutionContextOptions: executionContextOptions,
			CurrentState: models.PersistenceLogicAppWorkflowState{
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
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool) {
	context, carriesAzureControl, _ := persistencePrincipalRoleContext(persistencePrincipalRoleContextOptions{
		fallbackName:           persistenceLogicAppIdentityName(workflow),
		kind:                   "logic-app-execution-context",
		principalID:            workflow.PrincipalID,
		identityType:           workflow.IdentityType,
		permissionsByPrincipal: permissionsByPrincipal,
		assignmentsByPrincipal: assignmentsByPrincipal,
		resolvedSummary: func(name string, roleSummary string) string {
			return fmt.Sprintf("The strongest visible execution context here is the Logic App identity `%s`, which already holds %s.", name, roleSummary)
		},
		lowerImpactSummary: func(name string) string {
			return fmt.Sprintf("Logic App identity `%s` is visible here, but only lower-impact Azure role assignments are visible from current scope.", name)
		},
		unresolvedPrivilegedSummary: func(name string, roleSummary string) string {
			return fmt.Sprintf("The strongest visible execution context here is the Logic App identity `%s`, which already holds %s.", name, roleSummary)
		},
		noAssignmentsSummary: func(name string) string {
			return fmt.Sprintf("Logic App identity `%s` is visible here, but no Azure role-assignment rows are found for its principal ID.", name)
		},
		rbacOnlyCarriesAzureControl: true,
	})
	return context, carriesAzureControl
}

func persistenceLogicAppIdentityName(workflow models.LogicAppWorkflowAsset) string {
	if strings.Contains(strings.ToLower(stringPtrValue(workflow.IdentityType)), "userassigned") && len(workflow.IdentityIDs) > 0 {
		return persistenceResourceNameFromID(workflow.IdentityIDs[0])
	}
	return firstNonEmpty(workflow.Name+"-identity", "logic-app identity")
}

func persistenceLogicAppStillUnmapped(workflow models.LogicAppWorkflowAsset) []string {
	items := []string{
		"the current command does not print full workflow definitions, connector secret material, or connection credential values, so operator intent is not inferred from hidden workflow content here",
		"the current command does not print callback URLs, access signatures, or other trigger secret material behind the visible trigger posture",
		"the current command does not invoke request triggers, validate upstream caller auth, or prove runtime-side trigger success",
		"the current command does not resolve exact downstream payloads or high-value target impact without deeper workflow and connector inspection",
	}
	if !persistenceLogicAppIsDurable(workflow) {
		items = append(items, "the current command does not prove whether a later definition change would turn this workflow into durable request or recurrence-backed re-entry")
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
