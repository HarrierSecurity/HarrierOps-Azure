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

type persistenceVMExtensionStepDefinition struct {
	Action     string
	APISurface string
}

var persistenceVMExtensionSteps = []persistenceVMExtensionStepDefinition{
	{Action: "modify VM extension configuration", APISurface: "Microsoft.Compute VM or VMSS extension write"},
	{Action: "reuse VM or VMSS target", APISurface: "Microsoft.Compute/virtualMachines or virtualMachineScaleSets"},
	{Action: "add or modify extension attachment", APISurface: "Microsoft.Compute extension child resource"},
	{Action: "provide script or command source", APISurface: "settings.fileUris / commandToExecute / scriptUri"},
	{Action: "configure extension execution", APISurface: "extension settings / protected settings posture"},
	{Action: "deliver config to VM agent", APISurface: "Azure extension lifecycle / instance view"},
	{Action: "hand off extension execution to VM agent", APISurface: "guest-side handler boundary"},
	{Action: "update extension later", APISurface: "forceUpdateTag / timestamp / extension update"},
	{Action: "preserve control-plane execution path", APISurface: "extension attachment + stored settings"},
}

func buildPersistenceVMExtensionsOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.PersistenceSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	extensionsFuture := runGroupedCommandOutput[models.VMExtensionsOutput](group, ctx, request, vmExtensionsHandler(provider, now), "vm-extensions")
	backingFutures := startPersistenceBackingFutures(group, ctx, provider, now, request)

	extensions, err := extensionsFuture.wait()
	if err != nil {
		return nil, err
	}
	backing, err := backingFutures.wait(request, extensions.Metadata.TenantID, extensions.Metadata.SubscriptionID, extensions.Issues)

	if err != nil {
		return nil, err
	}
	managedIdentitiesByID := persistenceVMExtensionsManagedIdentitiesByID(backing.managedIdentities.Identities)
	items := sortedByLess(extensions.VMExtensions, vmExtensionLess)
	rows := make([]models.PersistenceVMExtension, 0, len(items))
	for _, extension := range items {
		control, controlOK := persistenceVMExtensionControl(extension, backing.evidence.currentIdentityAssignments)
		currentContext := persistenceCurrentIdentityContext(backing.evidence.currentIdentity, control, controlOK)
		capabilitySteps := persistenceVMExtensionCapabilitySteps(controlOK)
		executionOptions := persistenceVMExtensionExecutionContextOptions(extension)
		strongestContext, strongestContextHasAzureControl := persistenceVMExtensionExecutionContext(
			extension,
			managedIdentitiesByID,
			backing.evidence.permissionsByPrincipal,
			backing.evidence.assignmentsByPrincipal,
		)
		nearbyNames := persistenceVMExtensionNearbyNames(items, extension.Name)

		rows = append(rows, models.PersistenceVMExtension{
			ID:                      extension.ID,
			Name:                    extension.Name,
			ResourceGroup:           extension.ResourceGroup,
			Location:                extension.Location,
			CapabilitySteps:         capabilitySteps,
			CurrentIdentityContext:  currentContext,
			ExecutionContextOptions: executionOptions,
			CurrentState:            persistenceVMExtensionStateFromAsset(extension, strongestContext, nearbyNames),
			StillUnmapped:           persistenceVMExtensionStillUnmapped(),
			Summary:                 persistenceVMExtensionSummary(extension, controlOK, strongestContext, strongestContextHasAzureControl),
			RelatedIDs:              mergeRelatedIDs(append(extension.RelatedIDs, persistenceVMExtensionContextPrincipalID(strongestContext))),
		})
	}

	return models.PersistenceVMExtensionsOutput{
		Metadata:           scopedMetadata(now, request, backing.tenantID, backing.subscriptionID, "persistence"),
		GroupedCommandName: "persistence",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		VMExtensions:       rows,
		Issues:             backing.issues,
	}, nil
}

func persistenceVMExtensionStateFromAsset(
	extension models.VMExtensionAsset,
	strongestContext *models.PersistenceRoleContext,
	nearbyNames []string,
) models.PersistenceVMExtensionState {
	return models.PersistenceVMExtensionState{
		TargetKind:                       extension.TargetKind,
		TargetName:                       extension.TargetName,
		TargetID:                         extension.TargetID,
		Publisher:                        extension.Publisher,
		ExtensionType:                    extension.ExtensionType,
		TypeHandlerVersion:               extension.TypeHandlerVersion,
		AutoUpgradeMinorVersion:          extension.AutoUpgradeMinorVersion,
		EnableAutomaticUpgrade:           extension.EnableAutomaticUpgrade,
		FileURIHosts:                     append([]string{}, extension.FileURIHosts...),
		FileURICount:                     extension.FileURICount,
		CommandClue:                      extension.CommandClue,
		PublicSettingKeys:                append([]string{}, extension.PublicSettingKeys...),
		ProtectedSettingsPresent:         extension.ProtectedSettingsPresent,
		KeyVaultProtectedSettings:        extension.KeyVaultProtectedSettings,
		SuppressFailures:                 extension.SuppressFailures,
		ForceUpdateTag:                   extension.ForceUpdateTag,
		RerunClues:                       append([]string{}, extension.RerunClues...),
		ProvisionAfterExtensions:         append([]string{}, extension.ProvisionAfterExtensions...),
		ProvisioningState:                extension.ProvisioningState,
		InstanceViewStatuses:             append([]string{}, extension.InstanceViewStatuses...),
		TargetIdentityIDs:                append([]string{}, extension.TargetIdentityIDs...),
		StrongestVisibleExecutionContext: strongestContext,
		VMSSOrchestrationMode:            extension.VMSSOrchestrationMode,
		VMSSUpgradeMode:                  extension.VMSSUpgradeMode,
		NearbyThematicNames:              append([]string{}, nearbyNames...),
	}
}

func persistenceVMExtensionControl(extension models.VMExtensionAsset, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	targetAction := "Microsoft.Compute/virtualMachines/extensions/write"
	if strings.EqualFold(extension.TargetKind, "vmss") {
		targetAction = "Microsoft.Compute/virtualMachineScaleSets/extensions/write"
	}
	targets := []string{extension.ID, extension.TargetID}
	bestRank := 99
	best := persistenceCurrentIdentityControl{}
	for _, assignment := range assignments {
		if !persistenceRoleAssignmentAllowsNamedOrActionControl(
			assignment,
			targetAction,
			"owner",
			"contributor",
			"virtual machine contributor",
		) {
			continue
		}
		for _, target := range targets {
			rank, ok := persistenceScopeRank(assignment.ScopeID, target)
			if !ok || rank >= bestRank {
				continue
			}
			bestRank = rank
			best = persistenceCurrentIdentityControl{
				RoleName: fmt.Sprintf("%s at %s", assignment.RoleName, persistenceScopeLabel(assignment.ScopeID)),
				ScopeID:  assignment.ScopeID,
			}
		}
	}
	return best, bestRank != 99
}

func persistenceVMExtensionCapabilitySteps(controlOK bool) []models.PersistenceCapabilityStep {
	steps := make([]models.PersistenceCapabilityStep, 0, len(persistenceVMExtensionSteps))
	for _, step := range persistenceVMExtensionSteps {
		steps = append(steps, models.PersistenceCapabilityStep{
			Action:     step.Action,
			APISurface: step.APISurface,
			Status:     persistenceVMExtensionStepStatus(step.Action, controlOK),
		})
	}
	return steps
}

func persistenceVMExtensionStepStatus(action string, controlOK bool) string {
	if !controlOK {
		return "not proven"
	}
	switch action {
	case "modify VM extension configuration",
		"reuse VM or VMSS target",
		"add or modify extension attachment",
		"provide script or command source",
		"configure extension execution",
		"update extension later",
		"preserve control-plane execution path":
		return "yes"
	default:
		return "not proven"
	}
}

func persistenceVMExtensionExecutionContextOptions(extension models.VMExtensionAsset) []string {
	options := []string{}
	if vmExtensionIsCustomScript(extension) {
		options = append(options, "Custom Script-style handler")
	}
	if len(extension.FileURIHosts) > 0 {
		options = append(options, "reachable source clue")
	}
	if strings.TrimSpace(stringPtrValue(extension.CommandClue)) != "" {
		options = append(options, "public command clue")
	}
	if len(extension.PublicSettingKeys) > 0 {
		options = append(options, "public settings visible")
	}
	if persistenceVMExtensionBoolPtrValue(extension.ProtectedSettingsPresent) {
		options = append(options, "protected settings present")
	}
	if persistenceVMExtensionBoolPtrValue(extension.KeyVaultProtectedSettings) {
		options = append(options, "Key Vault-referenced protected settings")
	}
	if len(extension.RerunClues) > 0 {
		options = append(options, "rerun clues visible")
	}
	if len(extension.TargetIdentityIDs) > 0 {
		options = append(options, "target identity attached")
	}
	return dedupeStrings(options)
}

func persistenceVMExtensionsManagedIdentitiesByID(identities []models.ManagedIdentity) map[string]models.ManagedIdentity {
	byID := make(map[string]models.ManagedIdentity, len(identities))
	for _, identity := range identities {
		key := persistenceVMExtensionsArmIDJoinKey(identity.ID)
		if key == "" {
			continue
		}
		byID[key] = identity
	}
	return byID
}

func persistenceVMExtensionExecutionContext(
	extension models.VMExtensionAsset,
	managedIdentitiesByID map[string]models.ManagedIdentity,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool) {
	candidates := []persistenceVMExtensionExecutionCandidate{}
	for _, identityID := range extension.TargetIdentityIDs {
		identity, ok := managedIdentitiesByID[persistenceVMExtensionsArmIDJoinKey(identityID)]
		if !ok || strings.TrimSpace(stringPtrValue(identity.PrincipalID)) == "" {
			continue
		}
		name := firstNonEmpty(identity.Name, persistenceResourceNameFromID(identity.ID), "target managed identity")
		candidates = append(candidates, persistenceVMExtensionExecutionCandidate{
			name:         name,
			principalID:  identity.PrincipalID,
			identityType: stringPtrIf(identity.IdentityType),
		})
	}
	ranked := []persistenceVMExtensionRankedExecutionCandidate{}
	for _, candidate := range candidates {
		context, privileged, ok := persistenceVMExtensionRoleContext(extension, candidate, permissionsByPrincipal, assignmentsByPrincipal)
		if !ok {
			continue
		}
		ranked = append(ranked, persistenceVMExtensionRankedExecutionCandidate{
			candidate:  candidate,
			context:    context,
			privileged: privileged,
		})
	}
	if len(ranked) == 0 {
		return nil, false
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		left := ranked[i]
		right := ranked[j]
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
		return left.candidate.name < right.candidate.name
	})
	return ranked[0].context, ranked[0].privileged
}

type persistenceVMExtensionExecutionCandidate struct {
	name         string
	principalID  *string
	identityType *string
}

type persistenceVMExtensionRankedExecutionCandidate struct {
	candidate  persistenceVMExtensionExecutionCandidate
	context    *models.PersistenceRoleContext
	privileged bool
}

func persistenceVMExtensionRoleContext(
	extension models.VMExtensionAsset,
	candidate persistenceVMExtensionExecutionCandidate,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool, bool) {
	if candidate.principalID == nil || stringPtrValue(candidate.principalID) == "" {
		return nil, false, false
	}
	targetKind := persistenceVMExtensionTargetKindPhrase(extension.TargetKind)
	name := firstNonEmpty(candidate.name, "target managed identity")
	if permission, ok := permissionsByPrincipal[*candidate.principalID]; ok {
		roleNames := append([]string{}, permission.HighImpactRoles...)
		if len(roleNames) == 0 {
			roleNames = append(roleNames, permission.AllRoleNames...)
		}
		summary := fmt.Sprintf("Target managed identity `%s` is attached to the %s target for this extension, and it already holds %s.", name, targetKind, persistenceRoleSummary(roleNames, permission.ScopeIDs))
		if !permission.Privileged {
			summary = fmt.Sprintf("Target managed identity `%s` is attached to the %s target for this extension, but only lower-impact Azure role assignments are visible from current scope.", name, targetKind)
		}
		return &models.PersistenceRoleContext{
			Name:         name,
			Kind:         "vm-extension-target-identity-context",
			PrincipalID:  candidate.principalID,
			IdentityType: candidate.identityType,
			RoleNames:    dedupeStrings(roleNames),
			ScopeIDs:     dedupeStrings(permission.ScopeIDs),
			Summary:      summary,
		}, permission.Privileged, true
	}

	assignments := assignmentsByPrincipal[*candidate.principalID]
	roleNames, scopeIDs, privileged := persistenceFunctionAssignmentsRoleContext(assignments)
	summary := fmt.Sprintf("Target managed identity `%s` is attached to the %s target for this extension, but no Azure role-assignment rows are found for its principal ID.", name, targetKind)
	switch {
	case len(assignments) == 0:
	case !privileged:
		summary = fmt.Sprintf("Target managed identity `%s` is attached to the %s target for this extension, but only lower-impact Azure role assignments are visible from current scope.", name, targetKind)
	default:
		summary = fmt.Sprintf("Target managed identity `%s` is attached to the %s target for this extension, and raw Azure role-assignment rows for its principal ID show %s.", name, targetKind, persistenceRoleSummary(roleNames, scopeIDs))
	}
	return &models.PersistenceRoleContext{
		Name:         name,
		Kind:         "vm-extension-target-identity-context",
		PrincipalID:  candidate.principalID,
		IdentityType: candidate.identityType,
		RoleNames:    dedupeStrings(roleNames),
		ScopeIDs:     dedupeStrings(scopeIDs),
		Summary:      summary,
	}, privileged, true
}

func persistenceVMExtensionContextPrincipalID(context *models.PersistenceRoleContext) string {
	if context == nil || context.PrincipalID == nil {
		return ""
	}
	return strings.TrimSpace(*context.PrincipalID)
}

func persistenceVMExtensionTargetKindPhrase(kind string) string {
	if strings.EqualFold(kind, "vmss") {
		return "VMSS"
	}
	if strings.EqualFold(kind, "vm") {
		return "VM"
	}
	return "VM or VMSS"
}

func persistenceVMExtensionsArmIDJoinKey(resourceID string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimRight(resourceID, "/")))
}

func persistenceVMExtensionStillUnmapped() []string {
	return []string{
		"the current command does not print protected setting values or Key Vault-referenced protected setting material",
		"the current command does not fetch script bodies, GitHub files, storage blobs, or other remote payloads, so operator intent is not inferred from source clues alone",
		"the current command does not inspect guest filesystem state, guest logs, running processes, or runtime-side script effects",
		"the current command does not prove that a previously configured command or remote payload still succeeds if Azure applies the extension again",
	}
}

func persistenceVMExtensionSummary(
	extension models.VMExtensionAsset,
	controlOK bool,
	strongestContext *models.PersistenceRoleContext,
	strongestContextHasAzureControl bool,
) string {
	if controlOK && persistenceVMExtensionShowsReusablePosture(extension) && strongestContext != nil && strongestContextHasAzureControl {
		return fmt.Sprintf("Current identity can repurpose VM Extension '%s' on %s '%s' as VM Extensions persistence, and the target managed identity already carries visible Azure control.", extension.Name, strings.ToUpper(extension.TargetKind), extension.TargetName)
	}
	if controlOK && persistenceVMExtensionShowsReusablePosture(extension) {
		return fmt.Sprintf("Current identity can repurpose VM Extension '%s' on %s '%s' as VM Extensions persistence, with visible handler, source, settings, and rerun posture from the current read path.", extension.Name, strings.ToUpper(extension.TargetKind), extension.TargetName)
	}
	if controlOK {
		return fmt.Sprintf("Current identity can modify VM Extension '%s' on %s '%s', but the current read path only confirms part of the extension persistence story.", extension.Name, strings.ToUpper(extension.TargetKind), extension.TargetName)
	}
	if persistenceVMExtensionShowsReusablePosture(extension) {
		return fmt.Sprintf("VM Extension '%s' on %s '%s' already shows reusable extension posture, but the current identity does not yet have a proven path to repurpose it here.", extension.Name, strings.ToUpper(extension.TargetKind), extension.TargetName)
	}
	return fmt.Sprintf("VM Extension '%s' on %s '%s' is visible, but the current identity does not yet have a proven path to turn it into reusable VM Extensions persistence.", extension.Name, strings.ToUpper(extension.TargetKind), extension.TargetName)
}

func persistenceVMExtensionShowsReusablePosture(extension models.VMExtensionAsset) bool {
	if vmExtensionIsCustomScript(extension) {
		return true
	}
	if len(extension.FileURIHosts) > 0 || strings.TrimSpace(stringPtrValue(extension.CommandClue)) != "" {
		return true
	}
	if len(extension.PublicSettingKeys) > 0 || persistenceVMExtensionBoolPtrValue(extension.ProtectedSettingsPresent) || persistenceVMExtensionBoolPtrValue(extension.KeyVaultProtectedSettings) {
		return true
	}
	return len(extension.RerunClues) > 0
}

func persistenceVMExtensionBoolPtrValue(value *bool) bool {
	return value != nil && *value
}

func persistenceVMExtensionNearbyNames(
	extensions []models.VMExtensionAsset,
	currentName string,
) []string {
	seen := map[string]struct{}{}
	candidates := []string{}
	for _, extension := range extensions {
		name := strings.TrimSpace(extension.Name)
		if name == "" || strings.EqualFold(name, currentName) {
			continue
		}
		if persistenceVMExtensionNameScore(name) == 0 {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		candidates = append(candidates, name)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		leftScore := persistenceVMExtensionNameScore(candidates[i])
		rightScore := persistenceVMExtensionNameScore(candidates[j])
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return candidates[i] < candidates[j]
	})
	if len(candidates) > 4 {
		return append([]string{}, candidates[:4]...)
	}
	return candidates
}

func persistenceVMExtensionNameScore(name string) int {
	lower := strings.ToLower(name)
	strongKeywords := []string{"maintenance", "bootstrap", "dependency", "monitoring", "config", "customscript", "custom-script", "patch"}
	for _, keyword := range strongKeywords {
		if strings.Contains(lower, keyword) {
			return 2
		}
	}
	weakKeywords := []string{"agent", "script", "setup", "baseline", "install", "reapply"}
	for _, keyword := range weakKeywords {
		if strings.Contains(lower, keyword) {
			return 1
		}
	}
	return 0
}
