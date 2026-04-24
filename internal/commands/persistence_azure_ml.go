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

type persistenceAzureMLStepDefinition struct {
	Action     string
	APISurface string
}

var persistenceAzureMLSteps = []persistenceAzureMLStepDefinition{
	{Action: "create or modify workspace", APISurface: "Microsoft.MachineLearningServices/workspaces"},
	{Action: "attach or reuse compute", APISurface: "computes"},
	{Action: "add or modify jobs or pipelines", APISurface: "jobs / pipelines"},
	{Action: "attach or reuse exec ctx", APISurface: "workspace identity"},
	{Action: "create or modify schedule", APISurface: "schedules"},
	{Action: "expose or reuse endpoint", APISurface: "onlineEndpoints"},
}

func buildPersistenceAzureMLOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.PersistenceSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	azureMLFuture := runGroupedCommandOutput[models.AzureMLOutput](group, ctx, request, azureMLHandler(provider, now), "azure-ml")
	managedIdentitiesFuture := runGroupedCommandOutput[models.ManagedIdentitiesOutput](group, ctx, request, managedIdentitiesHandler(provider, now), "managed-identities")
	permissionsFuture := runGroupedCommandOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions")
	rbacFuture := runGroupedCommandOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac")

	azureML, err := azureMLFuture.wait()
	if err != nil {
		return nil, err
	}
	managedIdentities, err := managedIdentitiesFuture.wait()
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
		stringPtrValue(azureML.Metadata.SubscriptionID),
		stringPtrValue(permissions.Metadata.SubscriptionID),
	)
	tenantID := firstNonEmpty(
		request.Tenant,
		stringPtrValue(azureML.Metadata.TenantID),
		stringPtrValue(permissions.Metadata.TenantID),
	)

	evidence := buildPersistencePrincipalEvidence(permissions.Permissions, rbac.RoleAssignments)

	managedIdentitiesByAttachment := persistenceAzureMLManagedIdentitiesByAttachment(managedIdentities.Identities)
	workspaces := sortedByLess(azureML.Workspaces, persistenceAzureMLWorkspaceLess)
	rows := make([]models.PersistenceAzureMLWorkspace, 0, len(workspaces))
	for _, workspace := range workspaces {
		control, controlOK := persistenceAzureMLControl(workspace.ID, evidence.currentIdentityAssignments)
		currentContext := persistenceCurrentIdentityContext(evidence.currentIdentity, control, controlOK)
		attachedManagedIdentities := persistenceAzureMLAttachedManagedIdentities(workspace, managedIdentitiesByAttachment)
		capabilitySteps := persistenceAzureMLCapabilitySteps(controlOK)
		executionContextOptions := persistenceAzureMLExecutionContextOptions(workspace)
		strongestContext, strongestContextHasAzureControl := persistenceAzureMLExecutionContext(workspace, attachedManagedIdentities, evidence.permissionsByPrincipal, evidence.assignmentsByPrincipal)
		nearbyNames := persistenceAzureMLNearbyNames(workspaces, workspace.Name)

		rows = append(rows, models.PersistenceAzureMLWorkspace{
			ID:                      workspace.ID,
			Name:                    workspace.Name,
			ResourceGroup:           workspace.ResourceGroup,
			Location:                workspace.Location,
			CapabilitySteps:         capabilitySteps,
			CurrentIdentityContext:  currentContext,
			ExecutionContextOptions: executionContextOptions,
			CurrentState: models.PersistenceAzureMLWorkspaceState{
				Classification:                   workspace.Classification,
				State:                            workspace.State,
				PublicNetworkAccess:              workspace.PublicNetworkAccess,
				IdentityType:                     workspace.IdentityType,
				VisibleIdentityNames:             persistenceAzureMLVisibleIdentityNames(workspace, attachedManagedIdentities),
				ComputeCount:                     persistenceAzureMLIntPtr(workspace.ComputeCount),
				ComputeTypes:                     append([]string{}, workspace.ComputeTypes...),
				JobCount:                         persistenceAzureMLIntPtr(workspace.JobCount),
				JobTypes:                         append([]string{}, workspace.JobTypes...),
				ScheduleCount:                    persistenceAzureMLIntPtr(workspace.ScheduleCount),
				ScheduleTriggerTypes:             append([]string{}, workspace.ScheduleTriggerTypes...),
				EndpointCount:                    persistenceAzureMLIntPtr(workspace.EndpointCount),
				EndpointAuthModes:                append([]string{}, workspace.EndpointAuthModes...),
				EndpointPublicAccess:             append([]string{}, workspace.EndpointPublicAccess...),
				DatastoreCount:                   persistenceAzureMLIntPtr(workspace.DatastoreCount),
				DatastoreTypes:                   append([]string{}, workspace.DatastoreTypes...),
				StrongestVisibleExecutionContext: strongestContext,
				NearbyThematicNames:              nearbyNames,
			},
			StillUnmapped: persistenceAzureMLStillUnmapped(workspace, attachedManagedIdentities),
			Summary:       persistenceAzureMLSummary(workspace, controlOK, strongestContext, strongestContextHasAzureControl),
			RelatedIDs:    mergeRelatedIDs(workspace.RelatedIDs),
		})
	}

	issues := append([]models.Issue{}, azureML.Issues...)
	issues = append(issues, managedIdentities.Issues...)
	issues = append(issues, permissions.Issues...)
	issues = append(issues, rbac.Issues...)

	return models.PersistenceAzureMLOutput{
		Metadata:           scopedMetadata(now, request, tenantID, subscriptionID, "persistence"),
		GroupedCommandName: "persistence",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		Workspaces:         rows,
		Issues:             issues,
	}, nil
}

func persistenceAzureMLIntPtr(value int) *int {
	copied := value
	return &copied
}

func persistenceAzureMLControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	bestRank := 99
	best := persistenceCurrentIdentityControl{}
	for _, assignment := range assignments {
		role := strings.ToLower(strings.TrimSpace(assignment.RoleName))
		if role != "owner" && role != "contributor" {
			continue
		}
		rank, ok := persistenceScopeRank(assignment.ScopeID, resourceID)
		if !ok || rank >= bestRank {
			continue
		}
		bestRank = rank
		best = persistenceCurrentIdentityControl{
			RoleName: fmt.Sprintf("%s at %s", assignment.RoleName, persistenceScopeLabel(assignment.ScopeID)),
			ScopeID:  assignment.ScopeID,
		}
	}
	return best, bestRank != 99
}

func persistenceAzureMLCapabilitySteps(controlOK bool) []models.PersistenceCapabilityStep {
	steps := make([]models.PersistenceCapabilityStep, 0, len(persistenceAzureMLSteps))
	for _, step := range persistenceAzureMLSteps {
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

func persistenceAzureMLExecutionContextOptions(workspace models.AzureMLWorkspaceAsset) []string {
	options := []string{}
	if strings.TrimSpace(stringPtrValue(workspace.IdentityType)) != "" {
		options = append(options, "managed identity")
	}
	if strings.TrimSpace(stringPtrValue(workspace.StorageAccountID)) != "" {
		options = append(options, "workspace-linked storage")
	}
	if strings.TrimSpace(stringPtrValue(workspace.KeyVaultID)) != "" {
		options = append(options, "workspace-linked key vault")
	}
	if strings.TrimSpace(stringPtrValue(workspace.ContainerRegistryID)) != "" {
		options = append(options, "workspace-linked container registry")
	}
	return dedupeStrings(options)
}

func persistenceAzureMLExecutionContext(
	workspace models.AzureMLWorkspaceAsset,
	attachedManagedIdentities []models.ManagedIdentity,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool) {
	candidates := []persistenceAzureMLExecutionCandidate{}
	for _, identity := range attachedManagedIdentities {
		candidate, ok := persistenceAzureMLExecutionCandidateFromIdentity(identity, permissionsByPrincipal, assignmentsByPrincipal)
		if !ok {
			continue
		}
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 && workspace.PrincipalID != nil && strings.TrimSpace(*workspace.PrincipalID) != "" {
		fallback := models.ManagedIdentity{
			ID:           firstNonEmpty(workspace.ID+"/identities/system", workspace.ID),
			Name:         persistenceAzureMLIdentityName(workspace),
			IdentityType: "systemAssigned",
			PrincipalID:  workspace.PrincipalID,
			AttachedTo:   []string{workspace.ID},
			ScopeIDs:     []string{},
		}
		candidate, ok := persistenceAzureMLExecutionCandidateFromIdentity(fallback, permissionsByPrincipal, assignmentsByPrincipal)
		if ok {
			candidates = append(candidates, candidate)
		}
	}
	if len(candidates) == 0 {
		return nil, false
	}
	sort.SliceStable(candidates, func(i int, j int) bool {
		return persistenceAzureMLExecutionCandidateLess(candidates[i], candidates[j])
	})
	best := candidates[0]
	return &best.context, best.privileged
}

type persistenceAzureMLExecutionCandidate struct {
	context    models.PersistenceRoleContext
	privileged bool
}

func persistenceAzureMLExecutionCandidateFromIdentity(
	identity models.ManagedIdentity,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (persistenceAzureMLExecutionCandidate, bool) {
	name := strings.TrimSpace(identity.Name)
	if name == "" {
		name = firstNonEmpty(persistenceAzureMLResourceNameFromID(identity.ID), "azure-ml execution identity")
	}

	context, privileged, ok := persistencePrincipalRoleContext(persistencePrincipalRoleContextOptions{
		fallbackName:           name,
		kind:                   "azure-ml-execution-context",
		principalID:            identity.PrincipalID,
		identityType:           stringPtrIf(identity.IdentityType),
		permissionsByPrincipal: permissionsByPrincipal,
		assignmentsByPrincipal: assignmentsByPrincipal,
		resolvedSummary: func(name string, roleSummary string) string {
			return fmt.Sprintf("The strongest visible execution context here is the Azure ML identity `%s`, which already holds %s.", name, roleSummary)
		},
		lowerImpactSummary: func(name string) string {
			return fmt.Sprintf("Azure ML execution identity `%s` is visible here, but only lower-impact Azure role assignments are visible from current scope.", name)
		},
		unresolvedPrivilegedSummary: func(name string, _ string) string {
			return fmt.Sprintf("Azure ML execution identity `%s` is visible here, and raw Azure role-assignment rows for its principal ID suggest stronger Azure control, but that principal is not resolved as a standalone permissions row here.", name)
		},
		noAssignmentsSummary: func(name string) string {
			return fmt.Sprintf("Azure ML execution identity `%s` is visible here, but no Azure role-assignment rows are found for its principal ID.", name)
		},
		rbacOnlyCarriesAzureControl: false,
	})
	if !ok {
		context = &models.PersistenceRoleContext{
			Name:         name,
			Kind:         "azure-ml-execution-context",
			PrincipalID:  identity.PrincipalID,
			IdentityType: stringPtrIf(identity.IdentityType),
			Summary:      fmt.Sprintf("Azure ML execution identity `%s` is visible here, but no Azure role-assignment rows are found for its principal ID.", name),
		}
	}
	context.ScopeIDs = dedupeStrings(append(append([]string{}, context.ScopeIDs...), identity.ScopeIDs...))

	return persistenceAzureMLExecutionCandidate{
		context:    *context,
		privileged: privileged,
	}, true
}

func persistenceAzureMLIdentityName(workspace models.AzureMLWorkspaceAsset) string {
	return firstNonEmpty(workspace.Name+"-workspace-identity", workspace.Name+"-identity", "azure-ml workspace identity")
}

func persistenceAzureMLExecutionCandidateLess(left, right persistenceAzureMLExecutionCandidate) bool {
	if left.privileged != right.privileged {
		return left.privileged
	}
	leftRoleRank := permissionRoleRank(left.context.RoleNames)
	rightRoleRank := permissionRoleRank(right.context.RoleNames)
	if leftRoleRank != rightRoleRank {
		return leftRoleRank < rightRoleRank
	}
	leftScopeRank := persistenceFunctionScopeBreadthRank(left.context.ScopeIDs)
	rightScopeRank := persistenceFunctionScopeBreadthRank(right.context.ScopeIDs)
	if leftScopeRank != rightScopeRank {
		return leftScopeRank < rightScopeRank
	}
	leftIdentityRank := persistenceAzureMLIdentityTypeRank(left.context.IdentityType)
	rightIdentityRank := persistenceAzureMLIdentityTypeRank(right.context.IdentityType)
	if leftIdentityRank != rightIdentityRank {
		return leftIdentityRank < rightIdentityRank
	}
	return left.context.Name < right.context.Name
}

func persistenceAzureMLIdentityTypeRank(identityType *string) int {
	value := strings.ToLower(strings.TrimSpace(stringPtrValue(identityType)))
	switch {
	case strings.Contains(value, "userassigned"):
		return 0
	case strings.Contains(value, "systemassigned"):
		return 1
	default:
		return 2
	}
}

func persistenceAzureMLWorkspaceLess(left, right models.AzureMLWorkspaceAsset) bool {
	leftRank := persistenceAzureMLClassificationRank(left.Classification)
	rightRank := persistenceAzureMLClassificationRank(right.Classification)
	if leftRank != rightRank {
		return leftRank < rightRank
	}
	if left.ScheduleCount != right.ScheduleCount {
		return left.ScheduleCount > right.ScheduleCount
	}
	if left.JobCount != right.JobCount {
		return left.JobCount > right.JobCount
	}
	if left.ComputeCount != right.ComputeCount {
		return left.ComputeCount > right.ComputeCount
	}
	if left.EndpointCount != right.EndpointCount {
		return left.EndpointCount > right.EndpointCount
	}
	leftIdentity := strings.TrimSpace(stringPtrValue(left.IdentityType)) != ""
	rightIdentity := strings.TrimSpace(stringPtrValue(right.IdentityType)) != ""
	if leftIdentity != rightIdentity {
		return leftIdentity
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func persistenceAzureMLClassificationRank(classification string) int {
	switch classification {
	case "execution-capable":
		return 0
	case "supporting-persistence-context":
		return 1
	default:
		return 2
	}
}

func persistenceAzureMLNearbyNames(workspaces []models.AzureMLWorkspaceAsset, currentName string) []string {
	candidates := []string{}
	for _, workspace := range workspaces {
		name := strings.TrimSpace(workspace.Name)
		if name == "" || strings.EqualFold(name, currentName) {
			continue
		}
		if workspace.ScheduleCount == 0 && workspace.JobCount == 0 && workspace.ComputeCount == 0 {
			continue
		}
		candidates = append(candidates, name)
	}
	sort.Strings(candidates)
	if len(candidates) > 4 {
		return append([]string{}, candidates[:4]...)
	}
	return candidates
}

func persistenceAzureMLStillUnmapped(workspace models.AzureMLWorkspaceAsset, attachedManagedIdentities []models.ManagedIdentity) []string {
	items := []string{
		"the current command does not retrieve notebook content, model content, environment definitions, or job or pipeline payloads from Azure ML workspaces, so operator intent is not inferred from Azure ML content here",
		"the current command does not inspect run history, job outputs, or schedule execution history, so it does not prove what has already run here over time",
		"the current command does not invoke online endpoints or test live runtime behavior, so endpoint conclusions here stop at visible management-plane posture",
	}
	if persistenceAzureMLHasUserAssignedIdentityGap(workspace, attachedManagedIdentities) {
		items = append(items, "attached user-assigned identities are visible by resource ID on this Azure ML workspace, but the current output does not yet resolve their backing principals into the strongest visible execution-context ranking")
	}
	return items
}

func persistenceAzureMLHasUserAssignedIdentityGap(workspace models.AzureMLWorkspaceAsset, attachedManagedIdentities []models.ManagedIdentity) bool {
	if !strings.Contains(strings.ToLower(stringPtrValue(workspace.IdentityType)), "userassigned") {
		return false
	}
	resolved := map[string]struct{}{}
	for _, identity := range attachedManagedIdentities {
		if strings.TrimSpace(identity.ID) == "" || identity.PrincipalID == nil || strings.TrimSpace(*identity.PrincipalID) == "" {
			continue
		}
		resolved[persistenceAzureMLArmIDJoinKey(identity.ID)] = struct{}{}
	}
	for _, identityID := range workspace.IdentityIDs {
		if !strings.Contains(strings.ToLower(identityID), "/userassignedidentities/") {
			continue
		}
		if _, ok := resolved[persistenceAzureMLArmIDJoinKey(identityID)]; !ok {
			return true
		}
	}
	return false
}

func persistenceAzureMLManagedIdentitiesByAttachment(identities []models.ManagedIdentity) map[string][]models.ManagedIdentity {
	index := map[string][]models.ManagedIdentity{}
	for _, identity := range identities {
		for _, attachedTo := range identity.AttachedTo {
			key := persistenceAzureMLArmIDJoinKey(attachedTo)
			if key == "" {
				continue
			}
			index[key] = append(index[key], identity)
		}
	}
	return index
}

func persistenceAzureMLAttachedManagedIdentities(
	workspace models.AzureMLWorkspaceAsset,
	identitiesByAttachment map[string][]models.ManagedIdentity,
) []models.ManagedIdentity {
	items := append([]models.ManagedIdentity{}, identitiesByAttachment[persistenceAzureMLArmIDJoinKey(workspace.ID)]...)
	seen := map[string]struct{}{}
	for _, item := range items {
		seen[persistenceAzureMLArmIDJoinKey(item.ID)] = struct{}{}
	}
	if persistenceAzureMLIdentityIncludesType(workspace.IdentityType, "SystemAssigned") {
		systemID := persistenceAzureMLArmIDJoinKey(workspace.ID + "/identities/system")
		if _, ok := seen[systemID]; !ok {
			items = append(items, models.ManagedIdentity{
				ID:           workspace.ID + "/identities/system",
				Name:         persistenceAzureMLIdentityName(workspace),
				IdentityType: "systemAssigned",
				PrincipalID:  workspace.PrincipalID,
				AttachedTo:   []string{workspace.ID},
			})
		}
	}
	sort.SliceStable(items, func(i int, j int) bool {
		left := items[i]
		right := items[j]
		leftRank := persistenceAzureMLIdentityTypeTextRank(left.IdentityType)
		rightRank := persistenceAzureMLIdentityTypeTextRank(right.IdentityType)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		return left.Name < right.Name
	})
	return items
}

func persistenceAzureMLVisibleIdentityNames(
	workspace models.AzureMLWorkspaceAsset,
	attachedManagedIdentities []models.ManagedIdentity,
) []string {
	names := []string{}
	for _, identity := range attachedManagedIdentities {
		name := strings.TrimSpace(identity.Name)
		if name == "" {
			name = persistenceAzureMLResourceNameFromID(identity.ID)
		}
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 && persistenceAzureMLIdentityIncludesType(workspace.IdentityType, "SystemAssigned") {
		names = append(names, persistenceAzureMLIdentityName(workspace))
	}
	return dedupeStrings(names)
}

func persistenceAzureMLIdentityTypeTextRank(identityType string) int {
	value := strings.ToLower(strings.TrimSpace(identityType))
	switch {
	case strings.Contains(value, "userassigned"):
		return 0
	case strings.Contains(value, "systemassigned"):
		return 1
	default:
		return 2
	}
}

func persistenceAzureMLIdentityIncludesType(identityType *string, expected string) bool {
	value := strings.ToLower(strings.TrimSpace(stringPtrValue(identityType)))
	if value == "" {
		return false
	}
	return strings.Contains(value, strings.ToLower(expected))
}

func persistenceAzureMLResourceNameFromID(resourceID string) string {
	resourceID = strings.TrimSpace(resourceID)
	if resourceID == "" {
		return ""
	}
	parts := strings.Split(strings.Trim(resourceID, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func persistenceAzureMLArmIDJoinKey(resourceID string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimRight(resourceID, "/")))
}

func persistenceAzureMLSummary(
	workspace models.AzureMLWorkspaceAsset,
	controlOK bool,
	strongestContext *models.PersistenceRoleContext,
	strongestContextHasAzureControl bool,
) string {
	if !controlOK {
		return fmt.Sprintf("Current identity can see Azure ML workspace '%s', but current RBAC evidence does not yet prove workspace-level persistence control.", workspace.Name)
	}
	if strongestContextHasAzureControl {
		return fmt.Sprintf("Current identity can repurpose Azure ML workspace '%s' as reusable ML compute, schedule, and endpoint-backed persistence, and the strongest visible execution context already carries Azure control.", workspace.Name)
	}
	if strongestContext != nil {
		return fmt.Sprintf("Current identity can repurpose Azure ML workspace '%s' as reusable ML compute, schedule, and endpoint-backed persistence from current RBAC evidence.", workspace.Name)
	}
	return fmt.Sprintf("Current identity can repurpose Azure ML workspace '%s' as reusable ML compute, schedule, and endpoint-backed persistence from current RBAC evidence.", workspace.Name)
}
