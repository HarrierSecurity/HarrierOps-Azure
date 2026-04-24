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

type persistenceFunctionStepDefinition struct {
	Action     string
	APISurface string
}

var persistenceFunctionSteps = []persistenceFunctionStepDefinition{
	{Action: "create or modify function app", APISurface: "Microsoft.Web/sites"},
	{Action: "deploy or replace code", APISurface: "zip deploy / run-from-package / publish"},
	{Action: "repurpose trigger posture", APISurface: "HTTP, timer, queue, or event trigger"},
	{Action: "change app settings or deployment config", APISurface: "app settings / site config"},
	{Action: "attach or reuse exec ctx", APISurface: "function identity / app settings"},
	{Action: "restart or enable function host", APISurface: "site state / restart action"},
}

func buildPersistenceFunctionsOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.PersistenceSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	functionsFuture := runGroupedCommandOutput[models.FunctionsOutput](group, ctx, request, functionsHandler(provider, now), "functions")
	permissionsFuture := runGroupedCommandOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions")
	rbacFuture := runGroupedCommandOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac")

	functions, err := functionsFuture.wait()
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
		stringPtrValue(functions.Metadata.SubscriptionID),
		stringPtrValue(permissions.Metadata.SubscriptionID),
	)
	tenantID := firstNonEmpty(
		request.Tenant,
		stringPtrValue(functions.Metadata.TenantID),
		stringPtrValue(permissions.Metadata.TenantID),
	)

	evidence := buildPersistencePrincipalEvidence(permissions.Permissions, rbac.RoleAssignments)

	functionApps := sortedByLess(functions.FunctionApps, functionAppLess)
	rows := make([]models.PersistenceFunctionApp, 0, len(functionApps))
	for _, app := range functionApps {
		control, controlOK := persistenceAutomationControl(app.ID, evidence.currentIdentityAssignments)
		currentContext := persistenceCurrentIdentityContext(evidence.currentIdentity, control, controlOK)
		capabilitySteps := persistenceFunctionCapabilitySteps(controlOK)
		executionContextOptions := persistenceFunctionExecutionContextOptions(app)
		strongestContext, strongestContextHasAzureControl := persistenceFunctionExecutionContext(app, evidence.permissionsByPrincipal, evidence.assignmentsByPrincipal)
		nearbyNames := persistenceFunctionNearbyNames(functionApps, app.Name)

		rows = append(rows, models.PersistenceFunctionApp{
			ID:                      app.ID,
			Name:                    app.Name,
			ResourceGroup:           app.ResourceGroup,
			Location:                app.Location,
			CapabilitySteps:         capabilitySteps,
			CurrentIdentityContext:  currentContext,
			ExecutionContextOptions: executionContextOptions,
			CurrentState: models.PersistenceFunctionAppState{
				State:                            app.State,
				Hostname:                         app.DefaultHostname,
				PublicNetworkAccess:              app.PublicNetworkAccess,
				Runtime:                          app.Runtime,
				Deployment:                       app.Deployment,
				IdentityType:                     app.WorkloadIdentityType,
				AlwaysOn:                         app.AlwaysOn,
				AzureWebJobsStorageValueType:     app.AzureWebJobsStorageValueType,
				KeyVaultReferenceCount:           app.KeyVaultReferenceCount,
				RunFromPackage:                   app.RunFromPackage,
				TriggerTypes:                     append([]string{}, app.TriggerTypes...),
				VisibleFunctions:                 append([]models.FunctionChildAsset{}, app.VisibleFunctions...),
				StrongestVisibleExecutionContext: strongestContext,
				NearbyThematicNames:              nearbyNames,
			},
			StillUnmapped: persistenceFunctionStillUnmapped(app),
			Summary:       persistenceFunctionSummary(app, controlOK, strongestContext, strongestContextHasAzureControl),
			RelatedIDs:    mergeRelatedIDs(app.RelatedIDs),
		})
	}

	issues := append([]models.Issue{}, functions.Issues...)
	issues = append(issues, permissions.Issues...)
	issues = append(issues, rbac.Issues...)

	return models.PersistenceFunctionsOutput{
		Metadata:           scopedMetadata(now, request, tenantID, subscriptionID, "persistence"),
		GroupedCommandName: "persistence",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		FunctionApps:       rows,
		Issues:             issues,
	}, nil
}

func persistenceFunctionCapabilitySteps(controlOK bool) []models.PersistenceCapabilityStep {
	steps := make([]models.PersistenceCapabilityStep, 0, len(persistenceFunctionSteps))
	for _, step := range persistenceFunctionSteps {
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

func persistenceFunctionExecutionContextOptions(app models.FunctionAppAsset) []string {
	options := []string{}
	if strings.TrimSpace(stringPtrValue(app.WorkloadIdentityType)) != "" {
		options = append(options, "managed identity")
	}
	if intPtrValue(app.KeyVaultReferenceCount) > 0 {
		options = append(options, "Key Vault-backed settings")
	}
	if value := strings.TrimSpace(stringPtrValue(app.AzureWebJobsStorageValueType)); value != "" {
		options = append(options, "AzureWebJobsStorage="+value)
	}
	return dedupeStrings(options)
}

func persistenceFunctionExecutionContext(
	app models.FunctionAppAsset,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool) {
	candidates := []persistenceFunctionExecutionCandidate{}
	if app.WorkloadPrincipalID != nil && strings.TrimSpace(*app.WorkloadPrincipalID) != "" {
		candidate := persistenceFunctionExecutionCandidate{
			name:            firstNonEmpty(app.Name+"-system", "function-app identity"),
			principalID:     app.WorkloadPrincipalID,
			identityType:    app.WorkloadIdentityType,
			attachedContext: "Function App identity",
		}
		if context, privileged, ok := persistenceFunctionRoleContextCandidate(candidate, permissionsByPrincipal, assignmentsByPrincipal); ok {
			candidates = append(candidates, persistenceFunctionExecutionCandidate{
				name:            candidate.name,
				principalID:     app.WorkloadPrincipalID,
				identityType:    app.WorkloadIdentityType,
				attachedContext: candidate.attachedContext,
				context:         context,
				privileged:      privileged,
			})
		}
	}

	for _, identity := range app.UserAssignedIdentities {
		if identity.PrincipalID == nil || strings.TrimSpace(*identity.PrincipalID) == "" {
			continue
		}
		identityType := "UserAssigned"
		candidate := persistenceFunctionExecutionCandidate{
			name:            firstNonEmpty(identity.Name, persistenceFunctionIdentityName(identity.ID), "attached user-assigned identity"),
			principalID:     identity.PrincipalID,
			identityType:    &identityType,
			attachedContext: "Attached user-assigned identity",
		}
		if context, privileged, ok := persistenceFunctionRoleContextCandidate(candidate, permissionsByPrincipal, assignmentsByPrincipal); ok {
			candidates = append(candidates, persistenceFunctionExecutionCandidate{
				name:            candidate.name,
				principalID:     identity.PrincipalID,
				identityType:    candidate.identityType,
				attachedContext: candidate.attachedContext,
				context:         context,
				privileged:      privileged,
			})
		}
	}

	if len(candidates) == 0 {
		return nil, false
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return persistenceFunctionExecutionCandidateLess(candidates[i], candidates[j])
	})
	return candidates[0].context, candidates[0].privileged
}

type persistenceFunctionExecutionCandidate struct {
	name            string
	principalID     *string
	identityType    *string
	attachedContext string
	context         *models.PersistenceRoleContext
	privileged      bool
}

func persistenceFunctionRoleContextCandidate(
	candidate persistenceFunctionExecutionCandidate,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool, bool) {
	return persistencePrincipalRoleContext(persistencePrincipalRoleContextOptions{
		fallbackName:           candidate.name,
		kind:                   "function-app-execution-context",
		principalID:            candidate.principalID,
		identityType:           candidate.identityType,
		permissionsByPrincipal: permissionsByPrincipal,
		assignmentsByPrincipal: assignmentsByPrincipal,
		resolvedSummary: func(name string, roleSummary string) string {
			return fmt.Sprintf("The strongest visible execution context here is the %s `%s`, which already holds %s.", strings.ToLower(candidate.attachedContext), name, roleSummary)
		},
		lowerImpactSummary: func(name string) string {
			return fmt.Sprintf("%s `%s` is visible here, but only lower-impact Azure role assignments are visible from current scope.", candidate.attachedContext, name)
		},
		unresolvedPrivilegedSummary: func(name string, _ string) string {
			return fmt.Sprintf("%s `%s` is visible here, and raw Azure role-assignment rows for its principal ID suggest stronger Azure control, but that principal is not resolved as a standalone permissions row here.", candidate.attachedContext, name)
		},
		noAssignmentsSummary: func(name string) string {
			return fmt.Sprintf("%s `%s` is visible here, but no Azure role-assignment rows are found for its principal ID.", candidate.attachedContext, name)
		},
		rbacOnlyCarriesAzureControl: false,
	})
}

func persistenceFunctionAssignmentsRoleContext(assignments []models.RoleAssignment) ([]string, []string, bool) {
	allRoles := []string{}
	highImpactRoles := []string{}
	scopeIDs := []string{}
	for _, assignment := range assignments {
		roleName := strings.TrimSpace(assignment.RoleName)
		if roleName != "" {
			allRoles = append(allRoles, roleName)
			switch strings.ToLower(roleName) {
			case "owner", "user access administrator", "contributor":
				highImpactRoles = append(highImpactRoles, roleName)
			}
		}
		if scopeID := strings.TrimSpace(assignment.ScopeID); scopeID != "" {
			scopeIDs = append(scopeIDs, scopeID)
		}
	}
	roleNames := dedupeStrings(highImpactRoles)
	if len(roleNames) == 0 {
		roleNames = dedupeStrings(allRoles)
	}
	return roleNames, dedupeStrings(scopeIDs), len(highImpactRoles) > 0
}

func persistenceFunctionExecutionCandidateLess(left, right persistenceFunctionExecutionCandidate) bool {
	leftContext := left.context
	rightContext := right.context
	if left.privileged != right.privileged {
		return left.privileged
	}
	leftRoleRank := permissionRoleRank(leftContext.RoleNames)
	rightRoleRank := permissionRoleRank(rightContext.RoleNames)
	if leftRoleRank != rightRoleRank {
		return leftRoleRank < rightRoleRank
	}
	leftScopeRank := persistenceFunctionScopeBreadthRank(leftContext.ScopeIDs)
	rightScopeRank := persistenceFunctionScopeBreadthRank(rightContext.ScopeIDs)
	if leftScopeRank != rightScopeRank {
		return leftScopeRank < rightScopeRank
	}
	if len(leftContext.ScopeIDs) != len(rightContext.ScopeIDs) {
		return len(leftContext.ScopeIDs) > len(rightContext.ScopeIDs)
	}
	return leftContext.Name < rightContext.Name
}

func persistenceFunctionScopeBreadthRank(scopeIDs []string) int {
	best := 99
	for _, scopeID := range scopeIDs {
		scopeLower := strings.ToLower(strings.TrimSpace(scopeID))
		rank := 3
		switch {
		case strings.Contains(scopeLower, "/subscriptions/") && !strings.Contains(scopeLower, "/resourcegroups/"):
			rank = 0
		case strings.Contains(scopeLower, "/resourcegroups/"):
			rank = 1
		case scopeLower != "":
			rank = 2
		}
		if rank < best {
			best = rank
		}
	}
	if best == 99 {
		return 9
	}
	return best
}

func persistenceFunctionIdentityName(identityID string) string {
	parts := strings.Split(strings.TrimSpace(identityID), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func persistenceFunctionStillUnmapped(app models.FunctionAppAsset) []string {
	items := []string{}
	if len(app.VisibleFunctions) > 0 || len(app.TriggerTypes) > 0 {
		items = append(items, "the current default collector does not perform data-plane validation of function keys, caller auth, upstream bus or storage access, or other runtime-side restrictions behind the visible management-plane trigger metadata")
	} else {
		items = append(items, "the current default collector did not confirm child-function trigger definitions from the management-plane read path for this Function App")
	}
	items = append(items,
		"the current command does not retrieve deployed function packages or project contents, so operator intent is not inferred from code here",
		"the current command does not infer downstream resource actions, secret usage, or service connections without reading deployed code or exercising the function path",
	)
	return dedupeStrings(items)
}

func persistenceFunctionSummary(
	app models.FunctionAppAsset,
	controlOK bool,
	strongestContext *models.PersistenceRoleContext,
	strongestContextHasAzureControl bool,
) string {
	triggerTruth := persistenceFunctionTriggerSummary(app.TriggerTypes)
	if controlOK && persistenceFunctionShowsReusablePosture(app) && strongestContext != nil && strongestContextHasAzureControl {
		if triggerTruth != "none" {
			return fmt.Sprintf("Current identity can repurpose Function App '%s' as reusable Azure Functions persistence, visible child functions already show %s trigger paths, and the strongest visible execution context already carries Azure control.", app.Name, triggerTruth)
		}
		return fmt.Sprintf("Current identity can repurpose Function App '%s' as reusable Azure Functions persistence, and the strongest visible execution context already carries Azure control.", app.Name)
	}
	if controlOK && persistenceFunctionShowsReusablePosture(app) {
		if triggerTruth != "none" {
			return fmt.Sprintf("Current identity can repurpose Function App '%s' as reusable Azure Functions persistence, and visible child functions already show %s trigger paths from the current read path.", app.Name, triggerTruth)
		}
		return fmt.Sprintf("Current identity can repurpose Function App '%s' as reusable Azure Functions persistence from current RBAC evidence.", app.Name)
	}
	if controlOK {
		if triggerTruth != "none" {
			return fmt.Sprintf("Current identity can build or repurpose Function App '%s', and visible child functions already show %s trigger paths from the current read path.", app.Name, triggerTruth)
		}
		return fmt.Sprintf("Current identity can build or repurpose Function App '%s', but the current read path does not yet prove exact per-function trigger posture beyond the visible host, identity, and deployment signals.", app.Name)
	}
	if persistenceFunctionShowsReusablePosture(app) {
		return fmt.Sprintf("Function App '%s' already shows reusable host, identity, and deployment posture, but the current identity does not yet have a proven path to repurpose it here.", app.Name)
	}
	return fmt.Sprintf("Function App '%s' is visible, but the current identity does not yet have a proven path to turn it into reusable Azure Functions persistence.", app.Name)
}

func persistenceFunctionTriggerSummary(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}

func persistenceFunctionShowsReusablePosture(app models.FunctionAppAsset) bool {
	if strings.TrimSpace(stringPtrValue(app.DefaultHostname)) != "" {
		return true
	}
	if strings.EqualFold(stringPtrValue(app.PublicNetworkAccess), "Enabled") {
		return true
	}
	if strings.TrimSpace(stringPtrValue(app.Runtime)) != "" {
		return true
	}
	return strings.TrimSpace(stringPtrValue(app.Deployment)) != ""
}

func persistenceFunctionNearbyNames(
	apps []models.FunctionAppAsset,
	currentAppName string,
) []string {
	seen := map[string]struct{}{}
	candidates := []string{}
	for _, app := range apps {
		name := strings.TrimSpace(app.Name)
		if name == "" || strings.EqualFold(name, currentAppName) {
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
