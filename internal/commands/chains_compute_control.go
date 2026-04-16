package commands

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

var computeControlUUID = regexp.MustCompile(`^[0-9a-fA-F-]{36}$`)

type computeControlIdentityBinding struct {
	PrincipalID          string
	IdentityName         string
	IdentityID           string
	BindingSource        string
	IdentityChoiceBasis  string
	IdentityChoiceDetail string
}

type computeControlCandidateBinding struct {
	computeControlIdentityBinding
	StrongerOutcome string
	ControlBasis    string
}

type computeControlCorroboration struct {
	IdentityChoice string
	IdentityID     string
	Basis          string
	Detail         string
}

func buildComputeControlOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	family contracts.FamilyContract,
) (models.ChainsOutput, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	tokenSurfacesFuture := runGroupedCommandOutput[models.TokensCredentialsOutput](group, ctx, request, tokensCredentialsHandler(provider, now), "tokens-credentials")
	envVarsFuture := runGroupedCommandOutput[models.EnvVarsOutput](group, ctx, request, envVarsHandler(provider, now), "env-vars")
	managedIdentitiesFuture := runGroupedCommandOutput[models.ManagedIdentitiesOutput](group, ctx, request, managedIdentitiesHandler(provider, now), "managed-identities")
	permissionsFuture := runGroupedCommandOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions")
	workloadsFuture := runGroupedCommandOutput[models.WorkloadsOutput](group, ctx, request, workloadsHandler(provider, now), "workloads")

	tokenSurfaces, err := tokenSurfacesFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	envVars, err := envVarsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	managedIdentities, err := managedIdentitiesFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	permissions, err := permissionsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	workloads, err := workloadsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}

	workloadsByAsset := make(map[string]models.WorkloadSummary, len(workloads.Workloads))
	for _, workload := range workloads.Workloads {
		if workload.AssetID == "" {
			continue
		}
		workloadsByAsset[workload.AssetID] = workload
	}

	managedByID := make(map[string]models.ManagedIdentity, len(managedIdentities.Identities))
	managedByPrincipal := map[string][]models.ManagedIdentity{}
	for _, identity := range managedIdentities.Identities {
		if identity.ID != "" {
			managedByID[identity.ID] = identity
		}
		if identity.PrincipalID != nil && strings.TrimSpace(*identity.PrincipalID) != "" {
			managedByPrincipal[*identity.PrincipalID] = append(managedByPrincipal[*identity.PrincipalID], identity)
		}
	}

	permissionsByPrincipal := map[string]models.PermissionRow{}
	for _, permission := range permissions.Permissions {
		if permission.PrincipalID == "" || !permission.Privileged {
			continue
		}
		permissionsByPrincipal[permission.PrincipalID] = permission
	}

	envRowsByAsset := map[string][]models.EnvVarSummary{}
	for _, envVar := range envVars.EnvVars {
		if envVar.AssetID == "" {
			continue
		}
		envRowsByAsset[envVar.AssetID] = append(envRowsByAsset[envVar.AssetID], envVar)
	}

	roleAssignmentsByPrincipal := map[string][]models.ManagedIdentityRoleAssignment{}
	for _, assignment := range managedIdentities.RoleAssignments {
		if assignment.PrincipalID == "" {
			continue
		}
		roleAssignmentsByPrincipal[assignment.PrincipalID] = append(roleAssignmentsByPrincipal[assignment.PrincipalID], assignment)
	}

	paths := make([]models.ChainPathRecord, 0)
	for _, surface := range tokenSurfaces.Surfaces {
		if surface.SurfaceType != models.TokenCredentialSurfaceManagedIdentityToken {
			continue
		}

		workload, ok := workloadsByAsset[surface.AssetID]
		if !ok {
			continue
		}

		envRows := envRowsByAsset[surface.AssetID]
		if computeControlIsMixedIdentityWorkload(workload) {
			binding := computeControlResolveMixedIdentityBinding(surface, workload, envRows, managedByID, managedByPrincipal)
			if binding != nil {
				permission, hasPermission := permissionsByPrincipal[binding.PrincipalID]
				assignmentSummary := computeControlAssignmentSummary(binding.PrincipalID, roleAssignmentsByPrincipal)
				if hasPermission || assignmentSummary != "" {
					paths = append(paths, buildComputeControlRecord(surface, workload, *binding, permission, hasPermission, assignmentSummary))
					continue
				}
			}

			candidates := computeControlMixedIdentityCandidates(surface, workload, managedByID, managedByPrincipal, permissionsByPrincipal, roleAssignmentsByPrincipal)
			if len(candidates) > 0 {
				paths = append(paths, buildComputeControlCandidateRecord(surface, workload, candidates))
			}
			continue
		}

		binding := computeControlResolveIdentityBinding(surface, workload, managedByID, managedByPrincipal)
		if binding == nil {
			continue
		}
		permission, hasPermission := permissionsByPrincipal[binding.PrincipalID]
		assignmentSummary := computeControlAssignmentSummary(binding.PrincipalID, roleAssignmentsByPrincipal)
		if !hasPermission && assignmentSummary == "" {
			continue
		}
		paths = append(paths, buildComputeControlRecord(surface, workload, *binding, permission, hasPermission, assignmentSummary))
	}

	sort.SliceStable(paths, func(i int, j int) bool {
		left := paths[i]
		right := paths[j]
		if prioritySortValue(left.Priority) != prioritySortValue(right.Priority) {
			return prioritySortValue(left.Priority) < prioritySortValue(right.Priority)
		}
		if computeControlUrgencyRank(stringPtrValue(left.Urgency)) != computeControlUrgencyRank(stringPtrValue(right.Urgency)) {
			return computeControlUrgencyRank(stringPtrValue(left.Urgency)) < computeControlUrgencyRank(stringPtrValue(right.Urgency))
		}
		if left.AssetName != right.AssetName {
			return left.AssetName < right.AssetName
		}
		return stringPtrValue(left.InsertionPoint) < stringPtrValue(right.InsertionPoint)
	})

	issues := append([]models.Issue{}, tokenSurfaces.Issues...)
	issues = append(issues, envVars.Issues...)
	issues = append(issues, managedIdentities.Issues...)
	issues = append(issues, permissions.Issues...)
	issues = append(issues, workloads.Issues...)

	return models.ChainsOutput{
		Metadata:                scopedMetadata(now, request, request.Tenant, request.Subscription, "chains"),
		GroupedCommandName:      "chains",
		Family:                  family.Name,
		InputMode:               "live",
		CommandState:            "extraction-only",
		Summary:                 family.Summary,
		ClaimBoundary:           family.AllowedClaim,
		CurrentGap:              models.StringPtr(family.CurrentGap),
		ArtifactPreferenceOrder: []string{},
		BackingCommands:         append([]string{}, family.BackingCommands...),
		SourceArtifacts:         []models.ChainSourceArtifact{},
		Paths:                   paths,
		Issues:                  issues,
	}, nil
}

func computeControlResolveIdentityBinding(
	surface models.TokenCredentialSurfaceSummary,
	workload models.WorkloadSummary,
	managedByID map[string]models.ManagedIdentity,
	managedByPrincipal map[string][]models.ManagedIdentity,
) *computeControlIdentityBinding {
	managedMatches := make([]models.ManagedIdentity, 0)
	seen := map[string]struct{}{}
	for _, value := range append(append([]string{}, surface.RelatedIDs...), workload.IdentityIDs...) {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		match, ok := managedByID[value]
		if ok {
			managedMatches = append(managedMatches, match)
		}
	}
	if len(managedMatches) == 1 {
		return computeControlBindingFromManagedIdentity(managedMatches[0])
	}
	if len(managedMatches) > 1 {
		return nil
	}

	principalID := strings.TrimSpace(stringPtrValue(workload.IdentityPrincipalID))
	if principalID == "" {
		principalIDs := map[string]struct{}{}
		for _, relatedID := range surface.RelatedIDs {
			if computeControlUUID.MatchString(relatedID) {
				principalIDs[relatedID] = struct{}{}
			}
		}
		if len(principalIDs) == 1 {
			for candidate := range principalIDs {
				principalID = candidate
			}
		}
	}
	if principalID == "" {
		return nil
	}

	if binding := computeControlAttachedIdentityBinding(principalID, surface.AssetID, managedByPrincipal); binding != nil {
		return binding
	}
	return computeControlSystemIdentityBinding(principalID, workload, surface)
}

func computeControlIsMixedIdentityWorkload(workload models.WorkloadSummary) bool {
	return strings.TrimSpace(stringPtrValue(workload.IdentityPrincipalID)) != "" && len(workload.IdentityIDs) > 0
}

func computeControlBindingFromManagedIdentity(identity models.ManagedIdentity) *computeControlIdentityBinding {
	if identity.PrincipalID == nil || strings.TrimSpace(*identity.PrincipalID) == "" {
		return nil
	}
	return &computeControlIdentityBinding{
		PrincipalID:   *identity.PrincipalID,
		IdentityName:  firstNonEmpty(identity.Name, *identity.PrincipalID),
		IdentityID:    firstNonEmpty(identity.ID, *identity.PrincipalID),
		BindingSource: "managed-identity",
	}
}

func computeControlAttachedIdentityBinding(
	principalID string,
	assetID string,
	managedByPrincipal map[string][]models.ManagedIdentity,
) *computeControlIdentityBinding {
	matches := make([]models.ManagedIdentity, 0)
	for _, identity := range managedByPrincipal[principalID] {
		for _, attached := range identity.AttachedTo {
			if attached == assetID {
				matches = append(matches, identity)
				break
			}
		}
	}
	if len(matches) != 1 {
		return nil
	}
	return computeControlBindingFromManagedIdentity(matches[0])
}

func computeControlResolveMixedIdentityBinding(
	surface models.TokenCredentialSurfaceSummary,
	workload models.WorkloadSummary,
	envRows []models.EnvVarSummary,
	managedByID map[string]models.ManagedIdentity,
	managedByPrincipal map[string][]models.ManagedIdentity,
) *computeControlIdentityBinding {
	corroboration := computeControlIdentityChoiceCorroboration(workload, envRows)
	if corroboration == nil {
		return nil
	}

	if corroboration.IdentityChoice == "systemAssigned" {
		principalID := strings.TrimSpace(stringPtrValue(workload.IdentityPrincipalID))
		if principalID == "" {
			return nil
		}
		binding := computeControlAttachedIdentityBinding(principalID, surface.AssetID, managedByPrincipal)
		if binding == nil {
			binding = computeControlSystemIdentityBinding(principalID, workload, surface)
		}
		binding.IdentityChoiceBasis = corroboration.Basis
		binding.IdentityChoiceDetail = corroboration.Detail
		return binding
	}

	if corroboration.IdentityChoice == "userAssigned" {
		identity, ok := managedByID[corroboration.IdentityID]
		if !ok {
			return nil
		}
		binding := computeControlBindingFromManagedIdentity(identity)
		if binding == nil {
			return nil
		}
		binding.IdentityChoiceBasis = corroboration.Basis
		binding.IdentityChoiceDetail = corroboration.Detail
		return binding
	}
	return nil
}

func computeControlIdentityChoiceCorroboration(workload models.WorkloadSummary, envRows []models.EnvVarSummary) *computeControlCorroboration {
	identityNames := map[string][]string{}
	identityIDs := map[string]struct{}{}
	for _, identityID := range workload.IdentityIDs {
		identityIDs[identityID] = struct{}{}
		normalized := computeControlNormalizeIdentitySelector(identityID)
		if normalized == "" {
			continue
		}
		identityNames[normalized] = append(identityNames[normalized], identityID)
	}

	corroborations := map[string]computeControlCorroboration{}
	for _, row := range envRows {
		explicitIdentity := strings.TrimSpace(stringPtrValue(row.KeyVaultReferenceIdentity))
		if explicitIdentity == "" {
			continue
		}
		basis := "env-vars:" + firstNonEmpty(row.SettingName, "unknown-setting")
		if strings.EqualFold(explicitIdentity, "systemAssigned") {
			corroborations["systemAssigned"] = computeControlCorroboration{
				IdentityChoice: "systemAssigned",
				Basis:          basis,
				Detail:         "current app configuration explicitly names SystemAssigned for a collected workload behavior.",
			}
			continue
		}
		if _, ok := identityIDs[explicitIdentity]; ok {
			corroborations["userAssigned::"+explicitIdentity] = computeControlCorroboration{
				IdentityChoice: "userAssigned",
				IdentityID:     explicitIdentity,
				Basis:          basis,
				Detail:         "current app configuration explicitly names the attached user-assigned identity '" + computeControlDisplayIdentitySelector(explicitIdentity) + "' for a collected workload behavior.",
			}
			continue
		}
		normalized := computeControlNormalizeIdentitySelector(explicitIdentity)
		matchedIDs := identityNames[normalized]
		if len(matchedIDs) == 1 {
			corroborations["userAssigned::"+matchedIDs[0]] = computeControlCorroboration{
				IdentityChoice: "userAssigned",
				IdentityID:     matchedIDs[0],
				Basis:          basis,
				Detail:         "current app configuration explicitly names the attached user-assigned identity '" + computeControlDisplayIdentitySelector(explicitIdentity) + "' for a collected workload behavior.",
			}
		}
	}

	if len(corroborations) != 1 {
		return nil
	}
	for _, corroboration := range corroborations {
		copy := corroboration
		return &copy
	}
	return nil
}

func computeControlMixedIdentityCandidates(
	surface models.TokenCredentialSurfaceSummary,
	workload models.WorkloadSummary,
	managedByID map[string]models.ManagedIdentity,
	managedByPrincipal map[string][]models.ManagedIdentity,
	permissionsByPrincipal map[string]models.PermissionRow,
	roleAssignmentsByPrincipal map[string][]models.ManagedIdentityRoleAssignment,
) []computeControlCandidateBinding {
	candidates := make([]computeControlIdentityBinding, 0)
	systemPrincipalID := strings.TrimSpace(stringPtrValue(workload.IdentityPrincipalID))
	if systemPrincipalID != "" {
		binding := computeControlAttachedIdentityBinding(systemPrincipalID, surface.AssetID, managedByPrincipal)
		if binding == nil {
			binding = computeControlSystemIdentityBinding(systemPrincipalID, workload, surface)
		}
		candidates = append(candidates, *binding)
	}

	for _, identityID := range workload.IdentityIDs {
		identity, ok := managedByID[identityID]
		if !ok {
			candidates = append(candidates, computeControlIdentityBinding{
				IdentityName: computeControlDisplayIdentitySelector(identityID),
				IdentityID:   identityID,
			})
			continue
		}
		binding := computeControlBindingFromManagedIdentity(identity)
		if binding != nil {
			candidates = append(candidates, *binding)
		}
	}

	visible := make([]computeControlCandidateBinding, 0, len(candidates))
	seen := map[string]struct{}{}
	anyControl := false
	for _, candidate := range candidates {
		dedupeKey := candidate.PrincipalID + "::" + candidate.IdentityID
		if _, ok := seen[dedupeKey]; ok {
			continue
		}
		seen[dedupeKey] = struct{}{}

		outcome, basis := computeControlCandidateSummary(candidate, permissionsByPrincipal, roleAssignmentsByPrincipal)
		candidateRow := computeControlCandidateBinding{
			computeControlIdentityBinding: candidate,
			StrongerOutcome:               outcome,
			ControlBasis:                  basis,
		}
		if outcome != "" {
			anyControl = true
		}
		visible = append(visible, candidateRow)
	}
	if !anyControl {
		return nil
	}
	return visible
}

func computeControlSystemIdentityBinding(
	principalID string,
	workload models.WorkloadSummary,
	surface models.TokenCredentialSurfaceSummary,
) *computeControlIdentityBinding {
	return &computeControlIdentityBinding{
		PrincipalID:   principalID,
		IdentityName:  firstNonEmpty(workload.AssetName+" system identity", surface.AssetName+" system identity", principalID),
		IdentityID:    principalID,
		BindingSource: "workload-principal",
	}
}

func computeControlCandidateSummary(
	candidate computeControlIdentityBinding,
	permissionsByPrincipal map[string]models.PermissionRow,
	roleAssignmentsByPrincipal map[string][]models.ManagedIdentityRoleAssignment,
) (string, string) {
	principalID := strings.TrimSpace(candidate.PrincipalID)
	if principalID == "" {
		return "", ""
	}
	if permission, ok := permissionsByPrincipal[principalID]; ok {
		return computeControlPermissionSummary(&permission), "permissions"
	}
	if assignmentSummary := computeControlAssignmentSummary(principalID, roleAssignmentsByPrincipal); assignmentSummary != "" {
		return assignmentSummary, "role-assignment"
	}
	return "", ""
}

func computeControlNormalizeIdentitySelector(value string) string {
	trimmed := strings.TrimSuffix(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	return strings.ToLower(parts[len(parts)-1])
}

func computeControlDisplayIdentitySelector(value string) string {
	trimmed := strings.TrimSuffix(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	return parts[len(parts)-1]
}

func computeControlAssignmentSummary(
	principalID string,
	roleAssignmentsByPrincipal map[string][]models.ManagedIdentityRoleAssignment,
) string {
	assignments := roleAssignmentsByPrincipal[principalID]
	highImpact := map[string]struct{}{}
	scopes := map[string]struct{}{}
	for _, assignment := range assignments {
		roleName := strings.ToLower(strings.TrimSpace(assignment.RoleName))
		switch roleName {
		case "owner", "contributor", "user access administrator":
			highImpact[assignment.RoleName] = struct{}{}
		}
		if assignment.ScopeID != "" {
			scopes[assignment.ScopeID] = struct{}{}
		}
	}
	if len(highImpact) == 0 {
		return ""
	}
	roleNames := make([]string, 0, len(highImpact))
	for roleName := range highImpact {
		roleNames = append(roleNames, roleName)
	}
	sort.Strings(roleNames)
	scopeText := "subscription-wide scope"
	if len(scopes) > 1 {
		scopeText = fmt.Sprintf("%d visible scopes", len(scopes))
	}
	return fmt.Sprintf("%s across %s", strings.Join(roleNames, ", "), scopeText)
}

func buildComputeControlRecord(
	surface models.TokenCredentialSurfaceSummary,
	workload models.WorkloadSummary,
	binding computeControlIdentityBinding,
	permission models.PermissionRow,
	hasPermission bool,
	assignmentSummary string,
) models.ChainPathRecord {
	strongerOutcome := computeControlPermissionSummary(func() *models.PermissionRow {
		if !hasPermission {
			return nil
		}
		return &permission
	}())
	if strongerOutcome == "" {
		strongerOutcome = assignmentSummary
	}
	if strongerOutcome == "" {
		strongerOutcome = "-"
	}

	publicFoothold := computeControlHasPublicSignal(workload)
	priority := "medium"
	urgency := "review-soon"
	if publicFoothold {
		priority = "high"
		urgency = "pivot-now"
	}

	mixed := binding.IdentityChoiceBasis != ""
	confidenceBoundary := computeControlConfidenceBoundary(binding, hasPermission, permission, publicFoothold)
	nextReview := computeControlNextReview(workload, binding.IdentityChoiceBasis != "")
	whyCare := computeControlWhyCare(surface, workload, binding, strongerOutcome, mixed)

	evidenceCommands := []string{"tokens-credentials", "workloads"}
	joinedSurfaceTypes := []string{"managed-identity-token", "workload"}
	if mixed {
		evidenceCommands = append(evidenceCommands, "env-vars")
		joinedSurfaceTypes = append(joinedSurfaceTypes, "identity-choice-corroboration")
	}
	if binding.BindingSource == "managed-identity" {
		evidenceCommands = append(evidenceCommands, "managed-identities")
		joinedSurfaceTypes = append(joinedSurfaceTypes, "identity-anchor")
	} else {
		joinedSurfaceTypes = append(joinedSurfaceTypes, "workload-principal")
	}
	confirmationBasis := "role-assignment-join"
	if hasPermission {
		evidenceCommands = append(evidenceCommands, "permissions")
		joinedSurfaceTypes = append(joinedSurfaceTypes, "permissions")
		confirmationBasis = "permission-join"
	} else {
		joinedSurfaceTypes = append(joinedSurfaceTypes, "role-assignment")
	}
	if mixed {
		if hasPermission {
			confirmationBasis = "mixed-identity-corroborated-permission-join"
		} else {
			confirmationBasis = "mixed-identity-corroborated-role-assignment-join"
		}
	}

	targetResolution := "path-confirmed"
	missingConfirmation := ""
	if mixed {
		targetResolution = "identity-choice-corroborated"
		missingConfirmation = "Current foothold does not directly verify which attached identity the raw token path will choose on every request."
	}

	insertionPoint := computeControlInsertionPoint(surface, workload)
	sourceCommand := "tokens-credentials"
	sourceContext := surface.AccessPath
	pathConcept := "direct-token-opportunity"
	urgencyPtr := models.StringPtr(urgency)
	confirmationPtr := models.StringPtr(confirmationBasis)

	return models.ChainPathRecord{
		ChainID:             fmt.Sprintf("compute-control::%s::%s", surface.AssetID, binding.PrincipalID),
		AssetID:             surface.AssetID,
		AssetName:           firstNonEmpty(surface.AssetName, surface.AssetID),
		AssetKind:           firstNonEmpty(surface.AssetKind, workload.AssetKind),
		Location:            models.StringPtr(firstNonEmpty(stringPtrValue(surface.Location), workload.Location)),
		SourceCommand:       &sourceCommand,
		SourceContext:       &sourceContext,
		ClueType:            string(surface.SurfaceType),
		ConfirmationBasis:   confirmationPtr,
		Priority:            priority,
		Urgency:             urgencyPtr,
		VisiblePath:         surface.Summary,
		InsertionPoint:      &insertionPoint,
		PathConcept:         &pathConcept,
		StrongerOutcome:     models.StringPtr(strongerOutcome),
		WhyCare:             models.StringPtr(whyCare),
		LikelyImpact:        models.StringPtr(strongerOutcome),
		ConfidenceBoundary:  models.StringPtr(confidenceBoundary),
		TargetService:       "azure-control",
		TargetResolution:    targetResolution,
		EvidenceCommands:    evidenceCommands,
		JoinedSurfaceTypes:  joinedSurfaceTypes,
		TargetCount:         1,
		TargetIDs:           []string{binding.IdentityID},
		TargetNames:         []string{binding.IdentityName},
		NextReview:          nextReview,
		Summary:             strings.TrimSpace(confidenceBoundary + " " + nextReview),
		MissingConfirmation: missingConfirmation,
		RelatedIDs:          mergeRelatedIDs(surface.RelatedIDs, workload.RelatedIDs, []string{binding.IdentityID}),
	}
}

func buildComputeControlCandidateRecord(
	surface models.TokenCredentialSurfaceSummary,
	workload models.WorkloadSummary,
	candidates []computeControlCandidateBinding,
) models.ChainPathRecord {
	controlCandidates := make([]computeControlCandidateBinding, 0, len(candidates))
	controlBases := map[string]struct{}{}
	bindingSources := map[string]struct{}{}
	targetIDs := make([]string, 0, len(candidates))
	targetNames := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.IdentityID != "" {
			targetIDs = append(targetIDs, candidate.IdentityID)
		}
		if candidate.IdentityName != "" {
			targetNames = append(targetNames, candidate.IdentityName)
		}
		if candidate.StrongerOutcome != "" {
			controlCandidates = append(controlCandidates, candidate)
			controlBases[candidate.ControlBasis] = struct{}{}
		}
		if candidate.BindingSource != "" {
			bindingSources[candidate.BindingSource] = struct{}{}
		}
	}

	outcomeParts := make([]string, 0, len(controlCandidates))
	for _, candidate := range controlCandidates {
		outcomeParts = append(outcomeParts, fmt.Sprintf("%s=%s", candidate.IdentityName, candidate.StrongerOutcome))
	}
	strongerOutcome := strings.Join(outcomeParts, "; ")
	confidenceBoundary := "Based on the current evidence, this workload can request tokens through mixed attached identities, but HO-Azure cannot directly verify which attached identity the raw token path will choose on every request. The attached identities currently in play are listed here instead of a single chosen lead."
	nextReview := "The current foothold bounds this path to the attached identities shown here; exact per-request identity choice remains unconfirmed."
	whyCare := computeControlRequiredFoothold(surface, workload, fmt.Sprintf("%s '%s' carries mixed identities. HO-Azure cannot yet defend one chosen identity, but visible Azure control currently maps to %s.", workload.AssetKind, workload.AssetName, strongerOutcome))

	evidenceCommands := []string{"tokens-credentials", "workloads"}
	joinedSurfaceTypes := []string{"managed-identity-token", "workload"}
	if _, ok := bindingSources["managed-identity"]; ok {
		evidenceCommands = append(evidenceCommands, "managed-identities")
		joinedSurfaceTypes = append(joinedSurfaceTypes, "identity-anchor")
	}
	if _, ok := bindingSources["workload-principal"]; ok {
		joinedSurfaceTypes = append(joinedSurfaceTypes, "workload-principal")
	}
	if _, ok := controlBases["permissions"]; ok {
		evidenceCommands = append(evidenceCommands, "permissions")
		joinedSurfaceTypes = append(joinedSurfaceTypes, "permissions")
	}
	if _, ok := controlBases["role-assignment"]; ok {
		joinedSurfaceTypes = append(joinedSurfaceTypes, "role-assignment")
	}

	priority := "medium"
	urgency := "review-soon"
	if computeControlHasPublicSignal(workload) {
		priority = "high"
		urgency = "pivot-now"
	}
	insertionPoint := computeControlInsertionPoint(surface, workload)
	sourceCommand := "tokens-credentials"
	sourceContext := surface.AccessPath
	pathConcept := "direct-token-opportunity"
	confirmationBasis := "mixed-identity-attached-candidates"

	return models.ChainPathRecord{
		ChainID:             fmt.Sprintf("compute-control::%s::mixed-identities", surface.AssetID),
		AssetID:             surface.AssetID,
		AssetName:           firstNonEmpty(surface.AssetName, surface.AssetID),
		AssetKind:           firstNonEmpty(surface.AssetKind, workload.AssetKind),
		Location:            models.StringPtr(firstNonEmpty(stringPtrValue(surface.Location), workload.Location)),
		SourceCommand:       &sourceCommand,
		SourceContext:       &sourceContext,
		ClueType:            string(surface.SurfaceType),
		ConfirmationBasis:   &confirmationBasis,
		Priority:            priority,
		Urgency:             models.StringPtr(urgency),
		VisiblePath:         surface.Summary,
		InsertionPoint:      &insertionPoint,
		PathConcept:         &pathConcept,
		StrongerOutcome:     models.StringPtr(strongerOutcome),
		WhyCare:             models.StringPtr(whyCare),
		LikelyImpact:        models.StringPtr(strongerOutcome),
		ConfidenceBoundary:  models.StringPtr(confidenceBoundary),
		TargetService:       "azure-control",
		TargetResolution:    "narrowed candidates",
		EvidenceCommands:    evidenceCommands,
		JoinedSurfaceTypes:  joinedSurfaceTypes,
		TargetCount:         len(targetIDs),
		TargetIDs:           targetIDs,
		TargetNames:         targetNames,
		NextReview:          nextReview,
		Summary:             strings.TrimSpace(confidenceBoundary + " " + nextReview),
		MissingConfirmation: "Current foothold does not directly verify which attached identity the raw token path will choose on every request.",
		RelatedIDs:          mergeRelatedIDs(surface.RelatedIDs, workload.RelatedIDs, targetIDs),
	}
}

func computeControlConfidenceBoundary(
	binding computeControlIdentityBinding,
	hasPermission bool,
	permission models.PermissionRow,
	publicFoothold bool,
) string {
	switch {
	case binding.IdentityChoiceBasis != "" && hasPermission:
		return "Due to mixed identities and the current foothold, HO-Azure cannot directly verify which attached identity the raw token path will choose on every request. Another collected workload surface currently points to this identity as the best current lead, and the stronger Azure control behind it is visible. Specifically, " + binding.IdentityChoiceDetail
	case binding.IdentityChoiceBasis != "":
		return "Due to mixed identities and the current foothold, HO-Azure cannot directly verify which attached identity the raw token path will choose on every request. Another collected workload surface currently points to this identity as the best current lead. Specifically, " + binding.IdentityChoiceDetail
	case hasPermission && binding.BindingSource == "managed-identity":
		return "HO-Azure can name the token-capable compute foothold, the attached identity, and the stronger Azure control behind it from current scope."
	case hasPermission:
		return "HO-Azure can name the token-capable compute foothold and the workload principal that maps to stronger Azure control from current scope. The explicit managed identity anchor is inferred from workload metadata rather than a separate managed-identities row."
	case binding.BindingSource == "managed-identity":
		return "HO-Azure can name the token-capable compute foothold and the attached identity, and can see a high-impact role signal on that identity. The fuller permission story still needs confirmation."
	default:
		_ = publicFoothold
		return "HO-Azure can name the token-capable compute foothold and a high-impact role signal on the workload principal that identity uses, but the explicit managed identity anchor and fuller permission story still need confirmation."
	}
}

func computeControlWhyCare(
	surface models.TokenCredentialSurfaceSummary,
	workload models.WorkloadSummary,
	binding computeControlIdentityBinding,
	strongerOutcome string,
	mixed bool,
) string {
	base := ""
	if mixed {
		base = fmt.Sprintf("%s '%s' carries mixed identities. Current collected workload behavior points to %s as the best current lead, and that identity already maps to %s.", workload.AssetKind, workload.AssetName, binding.IdentityName, strongerOutcome)
	} else {
		base = fmt.Sprintf("%s '%s' can request tokens as %s; that identity already maps to %s.", workload.AssetKind, workload.AssetName, binding.IdentityName, strongerOutcome)
	}
	return computeControlRequiredFoothold(surface, workload, base)
}

func computeControlRequiredFoothold(
	surface models.TokenCredentialSurfaceSummary,
	workload models.WorkloadSummary,
	base string,
) string {
	accessPath := surface.AccessPath
	assetKind := firstNonEmpty(workload.AssetKind, "workload")
	publicSignal := computeControlHasPublicSignal(workload)
	publicComputeLabel := "this public-facing service"
	if assetKind == "ContainerInstance" {
		publicComputeLabel = "this public-facing container group"
	}
	publicTokenRequestLabel := "make this public-facing service ask Azure for its own token"
	if assetKind == "ContainerInstance" {
		publicTokenRequestLabel = "make this public-facing container group ask Azure for its own token"
	}
	internalComputeLabel := "this workload"
	if assetKind == "ContainerInstance" {
		internalComputeLabel = "this container group"
	}

	switch accessPath {
	case "workload-identity":
		if publicSignal {
			return base + " To turn this into downstream Azure access, an operator would need a way to " + publicTokenRequestLabel + ". HO-Azure shows that " + publicComputeLabel + " is public and token-capable, but public reachability alone does not prove that path."
		}
		return base + " To turn this into downstream Azure access, an operator would need a service-side foothold that can run inside " + internalComputeLabel + " and invoke its token request path. HO-Azure does not yet show that start from the current foothold."
	case "imds":
		if publicSignal {
			return base + " To turn this into downstream Azure access, an operator would need a way to make this public-facing workload reach the Azure VM metadata service. HO-Azure shows that the workload is public and IMDS-backed, but public reachability alone does not prove that path."
		}
		return base + " To turn this into downstream Azure access, an operator would need host-level execution or admin access on this " + assetKind + " so the Azure VM metadata token path is reachable. HO-Azure does not yet show that start from the current foothold."
	default:
		return base + " To turn this into downstream Azure access, an operator would need a foothold that can reach the workload-side token path. HO-Azure does not yet show that start from the current foothold."
	}
}

func computeControlInsertionPoint(surface models.TokenCredentialSurfaceSummary, workload models.WorkloadSummary) string {
	switch surface.AccessPath {
	case "imds":
		if computeControlHasPublicSignal(workload) {
			return "public IMDS token path"
		}
		return "IMDS token path"
	case "workload-identity":
		if computeControlHasPublicSignal(workload) {
			return "reachable service token request path"
		}
		return "service token request path"
	default:
		return firstNonEmpty(surface.AccessPath, "token-capable compute path")
	}
}

func computeControlNextReview(workload models.WorkloadSummary, identityChoiceBasis bool) string {
	switch workload.AssetKind {
	case "VM":
		return "Check vms for the host foothold, then permissions for exact scope on the attached identity."
	case "VMSS":
		return "Check vmss for the fleet foothold, then permissions for exact scope on the attached identity."
	case "AppService":
		return "Check app-services for the running service foothold, then permissions for exact scope on the attached identity."
	case "FunctionApp":
		if identityChoiceBasis {
			return "Current collected workload configuration already narrows this path to the identity shown here; exact per-request token choice remains bounded by the current foothold."
		}
		return "Check functions for the running service foothold, then permissions for exact scope on the attached identity."
	default:
		return "Check workloads for the compute foothold, then permissions for exact scope on the attached identity."
	}
}

func computeControlPermissionSummary(permission *models.PermissionRow) string {
	if permission == nil {
		return ""
	}
	roles := append([]string{}, permission.HighImpactRoles...)
	if len(roles) == 0 {
		roles = []string{"high-impact roles"}
	}
	scopeCount := permission.ScopeCount
	if scopeCount == 0 {
		scopeCount = len(permission.ScopeIDs)
	}
	scopeText := "subscription-wide scope"
	if scopeCount > 1 {
		scopeText = fmt.Sprintf("%d visible scopes", scopeCount)
	}
	return fmt.Sprintf("%s across %s", strings.Join(roles, ", "), scopeText)
}

func computeControlHasPublicSignal(workload models.WorkloadSummary) bool {
	if len(workload.Endpoints) > 0 {
		return true
	}
	for _, family := range workload.ExposureFamilies {
		if strings.EqualFold(family, "public-ip") {
			return true
		}
	}
	return false
}

func computeControlUrgencyRank(urgency string) int {
	switch urgency {
	case "pivot-now":
		return 0
	case "review-soon":
		return 1
	case "bookmark":
		return 2
	default:
		return 9
	}
}
