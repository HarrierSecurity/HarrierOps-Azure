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

const (
	escalationPathResolutionConfirmed = "path-confirmed"
)

type escalationFootholdContext struct {
	PrincipalID       string
	Principal         string
	PrincipalType     string
	StartingFoothold  string
	RelatedIDs        []string
	CurrentPermission *models.PermissionRow
}

func buildEscalationPathOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	family contracts.FamilyContract,
) (models.ChainsOutput, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	permissionsFuture := runGroupedCommandOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions")
	roleTrustsFuture := runGroupedCommandOutput[models.RoleTrustsOutput](group, ctx, request, roleTrustsHandler(provider, now), "role-trusts")

	permissions, err := permissionsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	roleTrusts, err := roleTrustsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}

	permissionsByPrincipal := map[string]models.PermissionRow{}
	for _, permission := range permissions.Permissions {
		if strings.TrimSpace(permission.PrincipalID) == "" {
			continue
		}
		permissionsByPrincipal[permission.PrincipalID] = permission
	}

	paths := make([]models.ChainPathRecord, 0)
	for _, foothold := range escalationCurrentFootholds(permissions.Permissions) {
		if foothold.CurrentPermission != nil && foothold.CurrentPermission.Privileged {
			paths = append(paths, buildEscalationDirectControlRecord(foothold, *foothold.CurrentPermission))
		}
		paths = append(paths, buildEscalationTrustRecords(foothold, roleTrusts.Trusts, permissionsByPrincipal)...)
	}

	sort.SliceStable(paths, func(i int, j int) bool {
		return escalationPathLess(paths[i], paths[j], permissionsByPrincipal)
	})

	issues := append([]models.Issue{}, permissions.Issues...)
	issues = append(issues, roleTrusts.Issues...)

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

func escalationCurrentFootholds(rows []models.PermissionRow) []escalationFootholdContext {
	candidates := make([]escalationFootholdContext, 0)
	for _, row := range rows {
		if !row.IsCurrentIdentity || strings.TrimSpace(row.PrincipalID) == "" {
			continue
		}
		name := firstNonEmpty(row.DisplayName, row.PrincipalID, "unknown")
		rowCopy := row
		candidates = append(candidates, escalationFootholdContext{
			PrincipalID:       row.PrincipalID,
			Principal:         name,
			PrincipalType:     firstNonEmpty(row.PrincipalType, "Principal"),
			StartingFoothold:  name + " (current foothold)",
			RelatedIDs:        mergeRelatedIDs([]string{row.PrincipalID}, row.ScopeIDs),
			CurrentPermission: &rowCopy,
		})
	}

	sort.SliceStable(candidates, func(i int, j int) bool {
		left := candidates[i]
		right := candidates[j]
		leftPrivileged := left.CurrentPermission != nil && left.CurrentPermission.Privileged
		rightPrivileged := right.CurrentPermission != nil && right.CurrentPermission.Privileged
		if leftPrivileged != rightPrivileged {
			return leftPrivileged
		}
		leftPriority := "low"
		rightPriority := "low"
		leftScopes := 0
		rightScopes := 0
		if left.CurrentPermission != nil {
			leftPriority = left.CurrentPermission.Priority
			leftScopes = max(left.CurrentPermission.ScopeCount, len(left.CurrentPermission.ScopeIDs))
		}
		if right.CurrentPermission != nil {
			rightPriority = right.CurrentPermission.Priority
			rightScopes = max(right.CurrentPermission.ScopeCount, len(right.CurrentPermission.ScopeIDs))
		}
		if leftPriority != rightPriority {
			return prioritySortValue(leftPriority) < prioritySortValue(rightPriority)
		}
		if leftScopes != rightScopes {
			return leftScopes > rightScopes
		}
		return left.Principal < right.Principal
	})

	return candidates
}

func buildEscalationDirectControlRecord(
	foothold escalationFootholdContext,
	permission models.PermissionRow,
) models.ChainPathRecord {
	scopeText := escalationPermissionScopeText(permission)
	strongerOutcome := escalationPermissionControlSummary(permission)
	impactRoles := strings.Join(permission.HighImpactRoles, ", ")
	if impactRoles == "" {
		impactRoles = "high-impact RBAC"
	}
	provenPath := fmt.Sprintf(
		"Current foothold '%s' already holds high-impact RBAC (%s) on visible scope.",
		foothold.Principal,
		impactRoles,
	)
	missingProof := "HO-Azure does not prove which exact abuse action is the best next step from this row alone."
	confidenceBoundary := strings.TrimSpace(provenPath + " " + missingProof)
	nextReview := "Check rbac for the exact assignment evidence behind the current foothold."
	whyCare := fmt.Sprintf(
		"The current foothold already has %s. This row is already direct Azure control, not a separate pivot hunt. HO-Azure is not narrowing one exact downstream action beyond the control already shown here.",
		strongerOutcome,
	)
	pathConcept := "current-foothold-direct-control"
	pathType := escalationPathTypeLabel(pathConcept)
	sourceCommand := "permissions"
	sourceContext := foothold.Principal
	urgency := "pivot-now"
	insertionPoint := "Current foothold already holds high-impact RBAC on visible scope."
	confirmationBasis := "current-identity-rooted"

	return models.ChainPathRecord{
		ChainID:             fmt.Sprintf("escalation-path::%s::current-foothold-direct-control", foothold.PrincipalID),
		AssetID:             foothold.PrincipalID,
		AssetName:           foothold.StartingFoothold,
		AssetKind:           foothold.PrincipalType,
		StartingFoothold:    models.StringPtr(foothold.StartingFoothold),
		SourceCommand:       &sourceCommand,
		SourceContext:       &sourceContext,
		ClueType:            "direct-role-abuse",
		ConfirmationBasis:   &confirmationBasis,
		Priority:            "high",
		Urgency:             &urgency,
		VisiblePath:         "Current foothold -> high-impact RBAC already visible",
		InsertionPoint:      &insertionPoint,
		PathConcept:         &pathConcept,
		PathType:            &pathType,
		StrongerOutcome:     models.StringPtr(strongerOutcome),
		WhyCare:             models.StringPtr(whyCare),
		LikelyImpact:        models.StringPtr(strongerOutcome),
		ConfidenceBoundary:  models.StringPtr(confidenceBoundary),
		TargetService:       "azure-control",
		TargetResolution:    escalationPathResolutionConfirmed,
		EvidenceCommands:    []string{"permissions"},
		JoinedSurfaceTypes:  []string{"current-foothold", "permission-summary"},
		TargetCount:         max(1, len(permission.ScopeIDs)),
		TargetIDs:           append([]string{}, permission.ScopeIDs...),
		TargetNames:         []string{scopeText},
		NextReview:          nextReview,
		Summary:             strings.TrimSpace(confidenceBoundary + " " + nextReview),
		MissingConfirmation: missingProof,
		RelatedIDs:          append([]string{}, foothold.RelatedIDs...),
	}
}

func buildEscalationTrustRecords(
	foothold escalationFootholdContext,
	trusts []models.RoleTrustSummary,
	permissionsByPrincipal map[string]models.PermissionRow,
) []models.ChainPathRecord {
	if strings.TrimSpace(foothold.PrincipalID) == "" {
		return nil
	}

	federatedByApplication := map[string][]models.RoleTrustSummary{}
	for _, trust := range trusts {
		if strings.TrimSpace(trust.TrustType) != "federated-credential" {
			continue
		}
		applicationID := strings.TrimSpace(trust.SourceObjectID)
		if applicationID == "" {
			continue
		}
		federatedByApplication[applicationID] = append(federatedByApplication[applicationID], trust)
	}

	type candidate struct {
		record    models.ChainPathRecord
		current   *models.PermissionRow
		target    *models.PermissionRow
		trustType string
		targetKey string
	}
	candidates := make([]candidate, 0)

	for _, trust := range trusts {
		if trust.SourceObjectID != foothold.PrincipalID {
			continue
		}

		if trust.TrustType == "app-owner" {
			appID := strings.TrimSpace(trust.TargetObjectID)
			for _, federated := range federatedByApplication[appID] {
				record, targetPermission, ok := buildEscalationFederatedTakeoverRecord(foothold, trust, federated, permissionsByPrincipal)
				if ok {
					candidates = append(candidates, candidate{
						record:    record,
						current:   foothold.CurrentPermission,
						target:    targetPermission,
						trustType: "federated-credential",
						targetKey: strings.Join(record.TargetIDs, "|"),
					})
				}
			}
		}

		record, targetPermission, ok := buildEscalationSingleTrustRecord(foothold, trust, permissionsByPrincipal)
		if ok {
			candidates = append(candidates, candidate{
				record:    record,
				current:   foothold.CurrentPermission,
				target:    targetPermission,
				trustType: trust.TrustType,
				targetKey: strings.Join(record.TargetIDs, "|"),
			})
		}
	}

	sort.SliceStable(candidates, func(i int, j int) bool {
		left := candidates[i]
		right := candidates[j]
		if left.record.Priority != right.record.Priority {
			return prioritySortValue(left.record.Priority) < prioritySortValue(right.record.Priority)
		}
		leftUpgrade := escalationPermissionUpgradeScore(left.current, left.target)
		rightUpgrade := escalationPermissionUpgradeScore(right.current, right.target)
		if leftUpgrade != rightUpgrade {
			return leftUpgrade > rightUpgrade
		}
		leftNewScopes := len(escalationPermissionNewScopeIDs(left.current, left.target))
		rightNewScopes := len(escalationPermissionNewScopeIDs(right.current, right.target))
		if leftNewScopes != rightNewScopes {
			return leftNewScopes > rightNewScopes
		}
		leftStrength := escalationPermissionRoleStrength(left.target)
		rightStrength := escalationPermissionRoleStrength(right.target)
		if leftStrength != rightStrength {
			return leftStrength > rightStrength
		}
		leftEffort := escalationPathEffortRank(stringPtrValue(left.record.PathConcept), left.trustType)
		rightEffort := escalationPathEffortRank(stringPtrValue(right.record.PathConcept), right.trustType)
		if leftEffort != rightEffort {
			return leftEffort < rightEffort
		}
		return left.record.ChainID < right.record.ChainID
	})

	selected := make([]models.ChainPathRecord, 0, len(candidates))
	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		key := stringPtrValue(candidate.record.PathConcept) + "::" + candidate.targetKey
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		selected = append(selected, candidate.record)
	}
	return selected
}

func buildEscalationSingleTrustRecord(
	foothold escalationFootholdContext,
	trust models.RoleTrustSummary,
	permissionsByPrincipal map[string]models.PermissionRow,
) (models.ChainPathRecord, *models.PermissionRow, bool) {
	targetPermissionID := strings.TrimSpace(trust.TargetObjectID)
	targetPermissionName := firstNonEmpty(stringPtrValue(trust.TargetName), trust.TargetObjectID, "unknown target")
	targetPermission, ok := permissionsByPrincipal[targetPermissionID]
	if !ok && trust.BackingServicePrincipalID != nil && strings.TrimSpace(*trust.BackingServicePrincipalID) != "" {
		targetPermissionID = strings.TrimSpace(*trust.BackingServicePrincipalID)
		targetPermissionName = firstNonEmpty(stringPtrValue(trust.BackingServicePrincipalName), targetPermissionID, "unknown target")
		targetPermission, ok = permissionsByPrincipal[targetPermissionID]
	}

	escalationMechanism := strings.TrimSpace(stringPtrValue(trust.EscalationMechanism))
	usableIdentityResult := strings.TrimSpace(stringPtrValue(trust.UsableIdentityResult))
	if !ok || escalationMechanism == "" || usableIdentityResult == "" {
		return models.ChainPathRecord{}, nil, false
	}
	if !escalationPermissionAddsNetValue(foothold.CurrentPermission, &targetPermission) {
		return models.ChainPathRecord{}, nil, false
	}

	pathConcept := escalationTrustPathConcept(trust.TrustType)
	pathType := escalationPathTypeLabel(pathConcept)
	priority, urgency, semanticNextReview := escalationSemanticDecision(pathConcept, escalationPathResolutionConfirmed)
	nextReview := firstNonEmpty(stringPtrValue(trust.NextReview), semanticNextReview)
	strongerOutcome := escalationPermissionControlSummary(targetPermission)
	confidenceBoundary := strings.TrimSpace(fmt.Sprintf(
		"%s %s HO-Azure can also confirm the stronger target's Azure control. HO-Azure does not prove successful conversion of that control path into usable downstream identity access from this row alone.",
		escalationMechanism,
		usableIdentityResult,
	))
	whyCare := escalationTrustNote(trust, foothold.CurrentPermission, &targetPermission, targetPermissionName)
	visiblePath := escalationVisiblePath(trust.TrustType)
	sourceCommand := "role-trusts"
	sourceContext := firstNonEmpty(stringPtrValue(trust.SourceName), trust.SourceObjectID)
	confirmationBasis := firstNonEmpty(trust.Confidence, "confirmed")
	insertionPoint := escalationMechanism

	record := buildEscalationTrustRecordBase(
		foothold,
		fmt.Sprintf("escalation-path::%s::%s::%s", foothold.PrincipalID, pathConcept, targetPermissionID),
		trust.TrustType,
		sourceCommand,
		sourceContext,
		confirmationBasis,
		priority,
		urgency,
		visiblePath,
		insertionPoint,
		pathConcept,
		pathType,
		strongerOutcome,
		whyCare,
		confidenceBoundary,
		[]string{"current-foothold", "trust-edge"},
		targetPermissionID,
		targetPermissionName,
		nextReview,
		"HO-Azure does not prove successful conversion of the visible trust-control path into usable downstream identity access.",
		mergeRelatedIDs(foothold.RelatedIDs, trust.RelatedIDs, []string{targetPermissionID}),
	)
	return record, &targetPermission, true
}

func buildEscalationFederatedTakeoverRecord(
	foothold escalationFootholdContext,
	currentTrust models.RoleTrustSummary,
	federatedTrust models.RoleTrustSummary,
	permissionsByPrincipal map[string]models.PermissionRow,
) (models.ChainPathRecord, *models.PermissionRow, bool) {
	targetPermissionID := strings.TrimSpace(federatedTrust.TargetObjectID)
	targetPermission, ok := permissionsByPrincipal[targetPermissionID]
	if !ok || !escalationPermissionAddsNetValue(foothold.CurrentPermission, &targetPermission) {
		return models.ChainPathRecord{}, nil, false
	}

	appName := firstNonEmpty(stringPtrValue(currentTrust.TargetName), currentTrust.TargetObjectID, "unknown application")
	targetPermissionName := firstNonEmpty(stringPtrValue(federatedTrust.TargetName), federatedTrust.TargetObjectID, "unknown target")
	usableIdentityResult := strings.TrimSpace(stringPtrValue(federatedTrust.UsableIdentityResult))
	if usableIdentityResult == "" {
		return models.ChainPathRecord{}, nil, false
	}

	priority, urgency, semanticNextReview := escalationSemanticDecision("trust-expansion", escalationPathResolutionConfirmed)
	nextReview := firstNonEmpty(stringPtrValue(federatedTrust.NextReview), semanticNextReview)
	strongerOutcome := escalationPermissionControlSummary(targetPermission)
	confidenceBoundary := strings.TrimSpace(fmt.Sprintf(
		"The current foothold can control application '%s'. Application '%s' already has federated trust that can yield service principal '%s' access. %s HO-Azure can also confirm the stronger target's Azure control. HO-Azure does not prove that the current foothold already controls the visible federated subject or has already changed this federated trust from this row alone.",
		appName,
		appName,
		targetPermissionName,
		usableIdentityResult,
	))
	whyCare := escalationFederatedNote(appName, targetPermissionName, foothold.CurrentPermission, &targetPermission)
	if cutPoint := strings.TrimSpace(stringPtrValue(currentTrust.DefenderCutPoint)); cutPoint != "" {
		whyCare = whyCare + " " + cutPoint
	}
	insertionPoint := fmt.Sprintf(
		"Application '%s' already has federated trust that can yield service principal '%s' access.",
		appName,
		targetPermissionName,
	)
	sourceCommand := "role-trusts"
	sourceContext := firstNonEmpty(stringPtrValue(currentTrust.SourceName), currentTrust.SourceObjectID)
	confirmationBasis := firstNonEmpty(federatedTrust.Confidence, "confirmed")
	pathConcept := "trust-expansion"
	pathType := escalationPathTypeLabel(pathConcept)
	record := buildEscalationTrustRecordBase(
		foothold,
		fmt.Sprintf("escalation-path::%s::trust-expansion::%s::federated", foothold.PrincipalID, targetPermissionID),
		"federated-credential",
		sourceCommand,
		sourceContext,
		confirmationBasis,
		priority,
		urgency,
		"Current foothold -> app control -> existing federated trust -> higher-value identity",
		insertionPoint,
		pathConcept,
		pathType,
		strongerOutcome,
		whyCare,
		confidenceBoundary,
		[]string{"current-foothold", "trust-edge", "federated-trust"},
		targetPermissionID,
		targetPermissionName,
		nextReview,
		"HO-Azure does not prove that the current foothold already controls the visible federated subject or has already changed the federated trust.",
		mergeRelatedIDs(foothold.RelatedIDs, currentTrust.RelatedIDs, federatedTrust.RelatedIDs, []string{targetPermissionID}),
	)
	return record, &targetPermission, true
}

func buildEscalationTrustRecordBase(
	foothold escalationFootholdContext,
	chainID string,
	clueType string,
	sourceCommand string,
	sourceContext string,
	confirmationBasis string,
	priority string,
	urgency string,
	visiblePath string,
	insertionPoint string,
	pathConcept string,
	pathType string,
	strongerOutcome string,
	whyCare string,
	confidenceBoundary string,
	joinedSurfaceTypes []string,
	targetPermissionID string,
	targetPermissionName string,
	nextReview string,
	missingConfirmation string,
	relatedIDs []string,
) models.ChainPathRecord {
	return models.ChainPathRecord{
		ChainID:             chainID,
		AssetID:             foothold.PrincipalID,
		AssetName:           foothold.StartingFoothold,
		AssetKind:           foothold.PrincipalType,
		StartingFoothold:    models.StringPtr(foothold.StartingFoothold),
		SourceCommand:       &sourceCommand,
		SourceContext:       &sourceContext,
		ClueType:            clueType,
		ConfirmationBasis:   &confirmationBasis,
		Priority:            priority,
		Urgency:             models.StringPtr(urgency),
		VisiblePath:         visiblePath,
		InsertionPoint:      &insertionPoint,
		PathConcept:         models.StringPtr(pathConcept),
		PathType:            models.StringPtr(pathType),
		StrongerOutcome:     models.StringPtr(strongerOutcome),
		WhyCare:             models.StringPtr(whyCare),
		LikelyImpact:        models.StringPtr(strongerOutcome),
		ConfidenceBoundary:  models.StringPtr(confidenceBoundary),
		TargetService:       "identity-trust",
		TargetResolution:    escalationPathResolutionConfirmed,
		EvidenceCommands:    []string{"role-trusts", "permissions"},
		JoinedSurfaceTypes:  joinedSurfaceTypes,
		TargetCount:         1,
		TargetIDs:           []string{targetPermissionID},
		TargetNames:         []string{targetPermissionName},
		NextReview:          nextReview,
		Summary:             confidenceBoundary,
		MissingConfirmation: missingConfirmation,
		RelatedIDs:          relatedIDs,
	}
}

func escalationPathLess(
	left models.ChainPathRecord,
	right models.ChainPathRecord,
	permissionsByPrincipal map[string]models.PermissionRow,
) bool {
	if left.Priority != right.Priority {
		return prioritySortValue(left.Priority) < prioritySortValue(right.Priority)
	}
	leftCurrent, leftHasCurrent := permissionsByPrincipal[left.AssetID]
	rightCurrent, rightHasCurrent := permissionsByPrincipal[right.AssetID]
	leftTarget, leftHasTarget := escalationRecordTargetPermission(left, permissionsByPrincipal)
	rightTarget, rightHasTarget := escalationRecordTargetPermission(right, permissionsByPrincipal)
	leftUpgrade := escalationPermissionUpgradeScore(valueOrNilPermission(leftCurrent, leftHasCurrent), valueOrNilPermission(leftTarget, leftHasTarget))
	rightUpgrade := escalationPermissionUpgradeScore(valueOrNilPermission(rightCurrent, rightHasCurrent), valueOrNilPermission(rightTarget, rightHasTarget))
	if leftUpgrade != rightUpgrade {
		return leftUpgrade > rightUpgrade
	}
	leftNewScopes := len(escalationPermissionNewScopeIDs(valueOrNilPermission(leftCurrent, leftHasCurrent), valueOrNilPermission(leftTarget, leftHasTarget)))
	rightNewScopes := len(escalationPermissionNewScopeIDs(valueOrNilPermission(rightCurrent, rightHasCurrent), valueOrNilPermission(rightTarget, rightHasTarget)))
	if leftNewScopes != rightNewScopes {
		return leftNewScopes > rightNewScopes
	}
	leftEffort := escalationPathEffortRank(stringPtrValue(left.PathConcept), left.ClueType)
	rightEffort := escalationPathEffortRank(stringPtrValue(right.PathConcept), right.ClueType)
	if leftEffort != rightEffort {
		return leftEffort < rightEffort
	}
	leftResolution := escalationResolutionRank(left.TargetResolution)
	rightResolution := escalationResolutionRank(right.TargetResolution)
	if leftResolution != rightResolution {
		return leftResolution < rightResolution
	}
	if left.AssetName != right.AssetName {
		return left.AssetName < right.AssetName
	}
	return left.ChainID < right.ChainID
}

func escalationRecordTargetPermission(record models.ChainPathRecord, permissionsByPrincipal map[string]models.PermissionRow) (models.PermissionRow, bool) {
	for _, targetID := range record.TargetIDs {
		permission, ok := permissionsByPrincipal[targetID]
		if ok {
			return permission, true
		}
	}
	return models.PermissionRow{}, false
}

func valueOrNilPermission(row models.PermissionRow, ok bool) *models.PermissionRow {
	if !ok {
		return nil
	}
	rowCopy := row
	return &rowCopy
}

func escalationResolutionRank(resolution string) int {
	switch resolution {
	case escalationPathResolutionConfirmed:
		return 0
	default:
		return 9
	}
}

func escalationPathEffortRank(pathConcept string, clueType string) int {
	if pathConcept == "current-foothold-direct-control" {
		return 0
	}
	if pathConcept == "app-permission-reach" {
		return 1
	}
	switch clueType {
	case "federated-credential":
		return 2
	case "service-principal-owner":
		return 3
	case "app-owner":
		return 4
	default:
		return 9
	}
}

func escalationTrustPathConcept(trustType string) string {
	if trustType == "app-to-service-principal" {
		return "app-permission-reach"
	}
	return "trust-expansion"
}

func escalationVisiblePath(trustType string) string {
	switch trustType {
	case "service-principal-owner":
		return "Current foothold -> service principal takeover -> higher-value identity"
	case "app-owner":
		return "Current foothold -> app control -> higher-value identity"
	case "app-to-service-principal":
		return "Current foothold -> app permission -> higher-value identity"
	default:
		return "Current foothold -> trust edge -> higher-value identity"
	}
}

func escalationSemanticDecision(pathConcept string, targetResolution string) (string, string, string) {
	switch pathConcept {
	case "current-foothold-direct-control":
		return "high", "pivot-now", "Check rbac for the exact assignment evidence behind the current foothold."
	case "app-permission-reach":
		if targetResolution == escalationPathResolutionConfirmed {
			return "medium", "review-soon", "Review the exact application-permission grant and the stronger target behind this path."
		}
		return "low", "bookmark", "Confirm whether this application-permission target adds meaningful Azure control beyond the current foothold."
	case "trust-expansion":
		if targetResolution == escalationPathResolutionConfirmed {
			return "medium", "review-soon", "Check permissions for the stronger target behind this trust edge."
		}
		return "low", "bookmark", "Confirm whether the trust target also holds meaningful Azure control."
	default:
		return "low", "bookmark", "Review the visible escalation story before deeper follow-up."
	}
}

func escalationPathTypeLabel(pathConcept string) string {
	switch pathConcept {
	case "current-foothold-direct-control":
		return "current-foothold direct control"
	case "app-permission-reach":
		return "app-permission reach"
	case "trust-expansion":
		return "trust expansion"
	default:
		return firstNonEmpty(pathConcept, "escalation path")
	}
}

func escalationTrustNote(
	trust models.RoleTrustSummary,
	currentPermission *models.PermissionRow,
	targetPermission *models.PermissionRow,
	targetPermissionName string,
) string {
	targetName := firstNonEmpty(stringPtrValue(trust.TargetName), trust.TargetObjectID)
	backingServicePrincipalName := firstNonEmpty(stringPtrValue(trust.BackingServicePrincipalName), targetPermissionName)
	gainText := escalationPermissionGainText(currentPermission, targetPermission)

	switch trust.TrustType {
	case "app-owner":
		if targetName != "" && backingServicePrincipalName != "" {
			return fmt.Sprintf(
				"The current foothold can control application '%s', which backs service principal '%s'. %s HO-Azure is not proving that the current foothold has already turned application control into usable '%s' access.",
				targetName,
				backingServicePrincipalName,
				gainText,
				backingServicePrincipalName,
			)
		}
	case "service-principal-owner":
		if targetName != "" {
			return fmt.Sprintf(
				"The current foothold can take over service principal '%s'. %s HO-Azure is not proving that the current foothold has already added or used authentication material for service principal '%s'.",
				targetName,
				gainText,
				targetName,
			)
		}
	case "app-to-service-principal":
		if targetName != "" {
			return fmt.Sprintf(
				"The current foothold already has application-permission reach into service principal '%s'. %s HO-Azure is not proving that the current foothold has already exercised one exact downstream action through '%s'.",
				targetName,
				gainText,
				targetName,
			)
		}
	}

	return fmt.Sprintf(
		"The current foothold can reach '%s'. %s HO-Azure is not proving that the current foothold has already turned this trust edge into usable downstream identity access.",
		targetPermissionName,
		gainText,
	)
}

func escalationFederatedNote(
	appName string,
	targetPermissionName string,
	currentPermission *models.PermissionRow,
	targetPermission *models.PermissionRow,
) string {
	gainText := escalationPermissionGainText(currentPermission, targetPermission)
	return fmt.Sprintf(
		"The current foothold can control application '%s', and that application already has federated trust into service principal '%s'. %s HO-Azure is not proving that the current foothold already controls the visible federated subject or has already changed that federated trust to make '%s' usable.",
		appName,
		targetPermissionName,
		gainText,
		targetPermissionName,
	)
}

func escalationPermissionScopeText(permission models.PermissionRow) string {
	scopeCount := permission.ScopeCount
	if scopeCount == 0 {
		scopeCount = len(permission.ScopeIDs)
	}
	if scopeCount <= 1 {
		return "subscription-wide scope"
	}
	return fmt.Sprintf("%d visible scopes", scopeCount)
}

func escalationPermissionControlSummary(permission models.PermissionRow) string {
	roleText := strings.Join(permission.HighImpactRoles, ", ")
	if roleText == "" {
		roleText = "high-impact roles"
	}
	return fmt.Sprintf("%s across %s", roleText, escalationPermissionScopeText(permission))
}

func escalationPermissionAddsNetValue(currentPermission *models.PermissionRow, targetPermission *models.PermissionRow) bool {
	if targetPermission == nil {
		return false
	}
	if currentPermission == nil {
		return true
	}
	if len(escalationPermissionNewScopeIDs(currentPermission, targetPermission)) > 0 {
		return true
	}
	return escalationPermissionUpgradeScore(currentPermission, targetPermission) > 0
}

func escalationPermissionUpgradeScore(currentPermission *models.PermissionRow, targetPermission *models.PermissionRow) int {
	if targetPermission == nil {
		return 0
	}
	currentStrength := escalationPermissionRoleStrength(currentPermission)
	targetStrength := escalationPermissionRoleStrength(targetPermission)
	if targetStrength > currentStrength {
		return targetStrength - currentStrength
	}
	return 0
}

func escalationPermissionNewScopeIDs(currentPermission *models.PermissionRow, targetPermission *models.PermissionRow) []string {
	if targetPermission == nil {
		return nil
	}
	targetScopeIDs := append([]string{}, targetPermission.ScopeIDs...)
	if currentPermission == nil {
		return targetScopeIDs
	}

	newScopes := make([]string, 0, len(targetScopeIDs))
	for _, targetScopeID := range targetScopeIDs {
		applies := false
		for _, currentScopeID := range currentPermission.ScopeIDs {
			if escalationScopeAppliesToResource(currentScopeID, targetScopeID) {
				applies = true
				break
			}
		}
		if !applies {
			newScopes = append(newScopes, targetScopeID)
		}
	}
	return newScopes
}

func escalationPermissionRoleStrength(permission *models.PermissionRow) int {
	if permission == nil {
		return 0
	}
	strength := 0
	for _, role := range permission.HighImpactRoles {
		switch escalationNormalizeRoleName(role) {
		case "owner":
			if strength < 3 {
				strength = 3
			}
		case "contributor":
			if strength < 1 {
				strength = 1
			}
		}
	}
	return strength
}

func escalationPermissionCapabilityText(permission *models.PermissionRow) string {
	if permission == nil {
		return "meaningful Azure control"
	}
	roles := map[string]struct{}{}
	for _, role := range permission.HighImpactRoles {
		roles[escalationNormalizeRoleName(role)] = struct{}{}
	}
	if _, ok := roles["owner"]; ok {
		return "Owner-level Azure control, including role assignment"
	}
	if _, ok := roles["contributor"]; ok {
		return "write/change control"
	}
	return "meaningful Azure control"
}

func escalationPermissionGainText(currentPermission *models.PermissionRow, targetPermission *models.PermissionRow) string {
	if targetPermission == nil {
		return "That would reach meaningful Azure control on the visible scope."
	}
	targetScopeIDs := append([]string{}, targetPermission.ScopeIDs...)
	newScopeIDs := escalationPermissionNewScopeIDs(currentPermission, targetPermission)
	targetCapability := escalationPermissionCapabilityText(targetPermission)
	if len(newScopeIDs) > 0 {
		return fmt.Sprintf("That would add %s on %s.", targetCapability, escalationScopeListText(newScopeIDs))
	}
	currentCapability := escalationPermissionCapabilityText(currentPermission)
	if escalationPermissionRoleStrength(targetPermission) > escalationPermissionRoleStrength(currentPermission) {
		return fmt.Sprintf(
			"That would upgrade the current foothold from %s to %s on %s.",
			currentCapability,
			targetCapability,
			escalationScopeListText(targetScopeIDs),
		)
	}
	return fmt.Sprintf("That would reach %s on %s.", targetCapability, escalationScopeListText(targetScopeIDs))
}

func escalationScopeListText(scopeIDs []string) string {
	cleaned := make([]string, 0, len(scopeIDs))
	for _, scopeID := range scopeIDs {
		if strings.TrimSpace(scopeID) != "" {
			cleaned = append(cleaned, scopeID)
		}
	}
	if len(cleaned) == 0 {
		return "the visible scope"
	}
	if len(cleaned) > 3 {
		return fmt.Sprintf("%d visible scopes", len(cleaned))
	}

	resourceGroups := make([]string, 0, len(cleaned))
	for _, scopeID := range cleaned {
		if escalationArmScopeKind(scopeID) == "resource_group" {
			resourceGroups = append(resourceGroups, escalationArmScopeName(scopeID))
		}
	}
	if len(resourceGroups) == len(cleaned) {
		quoted := make([]string, 0, len(resourceGroups))
		for _, name := range resourceGroups {
			quoted = append(quoted, fmt.Sprintf("resource group '%s'", firstNonEmpty(name, "unknown")))
		}
		if len(quoted) == 1 {
			return quoted[0]
		}
		return strings.Join(quoted[:len(quoted)-1], ", ") + " and " + quoted[len(quoted)-1]
	}

	labels := make([]string, 0, len(cleaned))
	for _, scopeID := range cleaned {
		labels = append(labels, escalationScopeLabel(scopeID))
	}
	if len(labels) == 1 {
		return labels[0]
	}
	return strings.Join(labels[:len(labels)-1], ", ") + " and " + labels[len(labels)-1]
}

func escalationScopeLabel(scopeID string) string {
	switch escalationArmScopeKind(scopeID) {
	case "resource_group":
		return fmt.Sprintf("resource group '%s'", firstNonEmpty(escalationArmScopeName(scopeID), "unknown"))
	case "subscription":
		return "subscription scope"
	case "resource":
		return fmt.Sprintf("resource '%s'", firstNonEmpty(escalationArmScopeName(scopeID), "unknown resource"))
	default:
		return "visible scope"
	}
}

func escalationScopeAppliesToResource(currentScopeID string, targetScopeID string) bool {
	current := strings.TrimRight(strings.TrimSpace(currentScopeID), "/")
	target := strings.TrimRight(strings.TrimSpace(targetScopeID), "/")
	if current == "" || target == "" {
		return false
	}
	return current == target || strings.HasPrefix(target, current+"/")
}

func escalationArmScopeKind(scopeID string) string {
	scope := strings.Trim(strings.TrimSpace(scopeID), "/")
	if scope == "" {
		return ""
	}
	parts := strings.Split(scope, "/")
	if len(parts) >= 2 && parts[0] == "subscriptions" {
		if len(parts) == 2 {
			return "subscription"
		}
		if len(parts) >= 4 && parts[2] == "resourceGroups" {
			if len(parts) == 4 {
				return "resource_group"
			}
			return "resource"
		}
		return "resource"
	}
	return ""
}

func escalationArmScopeName(scopeID string) string {
	scope := strings.Trim(strings.TrimSpace(scopeID), "/")
	if scope == "" {
		return ""
	}
	parts := strings.Split(scope, "/")
	if len(parts) >= 4 && parts[0] == "subscriptions" && parts[2] == "resourceGroups" {
		return parts[3]
	}
	return parts[len(parts)-1]
}

func escalationNormalizeRoleName(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}
