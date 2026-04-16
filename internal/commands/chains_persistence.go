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

type persistenceCurrentIdentityControl struct {
	RoleName string
	ScopeID  string
}

func buildPersistencePathOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	family contracts.FamilyContract,
) (models.ChainsOutput, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	appCredentialsFuture := runGroupedCommandOutput[models.AppCredentialsOutput](group, ctx, request, appCredentialsHandler(provider, now), "app-credentials")
	automationFuture := runGroupedCommandOutput[models.AutomationOutput](group, ctx, request, automationHandler(provider, now), "automation")
	permissionsFuture := runGroupedCommandOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions")
	rbacFuture := runGroupedCommandOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac")
	functionsFuture := runGroupedCommandOutput[models.FunctionsOutput](group, ctx, request, functionsHandler(provider, now), "functions")
	appServicesFuture := runGroupedCommandOutput[models.AppServicesOutput](group, ctx, request, appServicesHandler(provider, now), "app-services")
	managedIdentitiesFuture := runGroupedCommandOutput[models.ManagedIdentitiesOutput](group, ctx, request, managedIdentitiesHandler(provider, now), "managed-identities")
	roleTrustsFuture := runGroupedCommandOutput[models.RoleTrustsOutput](group, ctx, request, roleTrustsHandler(provider, now), "role-trusts")

	appCredentials, err := appCredentialsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	automation, err := automationFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	permissions, err := permissionsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	rbac, err := rbacFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	functions, err := functionsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	appServices, err := appServicesFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	managedIdentities, err := managedIdentitiesFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	roleTrusts, err := roleTrustsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}

	permissionsByPrincipal := map[string]models.PermissionRow{}
	currentIdentityPrincipals := map[string]struct{}{}
	for _, permission := range permissions.Permissions {
		if permission.PrincipalID == "" {
			continue
		}
		permissionsByPrincipal[permission.PrincipalID] = permission
		if permission.IsCurrentIdentity {
			currentIdentityPrincipals[permission.PrincipalID] = struct{}{}
		}
	}

	currentIdentityAssignments := make([]models.RoleAssignment, 0)
	for _, assignment := range rbac.RoleAssignments {
		if _, ok := currentIdentityPrincipals[assignment.PrincipalID]; ok {
			currentIdentityAssignments = append(currentIdentityAssignments, assignment)
		}
	}

	managedIdentityNames := map[string]string{}
	for _, identity := range managedIdentities.Identities {
		if identity.PrincipalID != nil && strings.TrimSpace(*identity.PrincipalID) != "" && strings.TrimSpace(identity.Name) != "" {
			managedIdentityNames[*identity.PrincipalID] = identity.Name
		}
	}

	paths := make([]models.ChainPathRecord, 0, len(appCredentials.AppCredentials)+len(automation.AutomationAccounts))
	for _, item := range appCredentials.AppCredentials {
		record, ok := buildAppCredentialPersistenceRecord(item, permissionsByPrincipal)
		if ok {
			paths = append(paths, record)
		}
	}
	for _, account := range automation.AutomationAccounts {
		record, ok := buildAutomationPersistenceRecord(account, currentIdentityAssignments, permissionsByPrincipal, managedIdentityNames)
		if ok {
			paths = append(paths, record)
		}
	}

	sort.SliceStable(paths, func(i int, j int) bool {
		left := paths[i]
		right := paths[j]
		if prioritySortValue(left.Priority) != prioritySortValue(right.Priority) {
			return prioritySortValue(left.Priority) < prioritySortValue(right.Priority)
		}
		if persistenceRowRank(stringPtrValue(left.PathType)) != persistenceRowRank(stringPtrValue(right.PathType)) {
			return persistenceRowRank(stringPtrValue(left.PathType)) < persistenceRowRank(stringPtrValue(right.PathType))
		}
		if left.AssetName != right.AssetName {
			return left.AssetName < right.AssetName
		}
		return stringPtrValue(left.PersistenceType) < stringPtrValue(right.PersistenceType)
	})

	issues := append([]models.Issue{}, appCredentials.Issues...)
	issues = append(issues, automation.Issues...)
	issues = append(issues, permissions.Issues...)
	issues = append(issues, rbac.Issues...)
	issues = append(issues, functions.Issues...)
	issues = append(issues, appServices.Issues...)
	issues = append(issues, managedIdentities.Issues...)
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

func buildAppCredentialPersistenceRecord(
	item models.AppCredentialSummary,
	permissionsByPrincipal map[string]models.PermissionRow,
) (models.ChainPathRecord, bool) {
	rowClass, classification, ok := persistenceAppCredentialRowShape(item.RowClass)
	if !ok {
		return models.ChainPathRecord{}, false
	}

	persistenceType := persistenceAppCredentialType(item)
	surface := item.TargetObjectType
	whatPersists := persistenceAppCredentialWhatPersists(item)
	footholdAnchor := persistenceAppCredentialAnchor(item)
	durability := persistenceAppCredentialDurability(rowClass)
	missingProof := persistenceAppCredentialMissingProof(item, rowClass)
	currentEvidence := strings.TrimSpace(item.CurrentEvidence)
	operatorActionability := strings.TrimSpace(item.OperatorActionability)
	recommendedFixFocus := strings.TrimSpace(item.RecommendedFixFocus)
	roleContext := strings.TrimSpace(item.RoleContext)
	noteText := joinSentences(currentEvidence, roleContext, operatorActionability)
	summary := joinSentences(noteText, missingProof)

	permission, permissionOK := persistenceAppCredentialPermission(item, permissionsByPrincipal)
	priority, urgency := persistencePriorityForIdentityRow(rowClass, permissionOK)
	nextReview := persistenceAppCredentialNextReview(item, rowClass)
	targetIDs, targetNames := persistenceAppCredentialTargets(item)
	sourceContext := item.ControlPath

	return models.ChainPathRecord{
		ChainID:                        fmt.Sprintf("persistence-path::identity::%s::%s", item.TargetObjectID, strings.ReplaceAll(rowClass, "_", "-")),
		AssetID:                        item.TargetObjectID,
		AssetName:                      item.TargetObjectName,
		AssetKind:                      item.TargetObjectType,
		Surface:                        stringPtrIf(surface),
		PersistenceType:                stringPtrIf(persistenceType),
		Classification:                 stringPtrIf(classification),
		Durability:                     stringPtrIf(durability),
		WhatPersists:                   stringPtrIf(whatPersists),
		FootholdAnchor:                 stringPtrIf(footholdAnchor),
		SurvivesHostRebuild:            persistenceBool(true),
		SurvivesOriginalAccountCleanup: persistenceBool(true),
		CurrentEvidence:                stringPtrIf(currentEvidence),
		MissingProof:                   stringPtrIf(missingProof),
		OperatorActionability:          stringPtrIf(operatorActionability),
		RecommendedFixFocus:            stringPtrIf(recommendedFixFocus),
		SourceCommand:                  models.StringPtr("app-credentials"),
		SourceContext:                  stringPtrIf(sourceContext),
		Source:                         stringPtrIf(item.TargetObjectName),
		ClueType:                       "identity-shadowing",
		ConfirmationBasis:              stringPtrIf(item.ControlPath),
		Priority:                       priority,
		Urgency:                        stringPtrIf(urgency),
		VisiblePath:                    persistenceVisiblePath(rowClass),
		PathConcept:                    stringPtrIf(classification),
		PathType:                       stringPtrIf(rowClass),
		ConfidenceBoundary:             stringPtrIf(missingProof),
		MissingConfirmation:            missingProof,
		NextReview:                     nextReview,
		Note:                           stringPtrIf(noteText),
		WhyCare:                        stringPtrIf(noteText),
		Summary:                        summary,
		TargetService:                  "identity-persistence",
		TargetResolution:               "named match",
		EvidenceCommands:               persistenceEvidenceCommands("app-credentials", permissionOK),
		JoinedSurfaceTypes:             persistenceAppCredentialJoinedSurfaces(item, permissionOK),
		TargetCount:                    len(targetIDs),
		TargetIDs:                      targetIDs,
		TargetNames:                    targetNames,
		RelatedIDs:                     mergeRelatedIDs(item.RelatedIDs, targetIDs),
		StrongerOutcome:                stringPtrIf(persistenceOutcomeText(rowClass, permission)),
	}, true
}

func buildAutomationPersistenceRecord(
	account models.AutomationAccountAsset,
	currentIdentityAssignments []models.RoleAssignment,
	permissionsByPrincipal map[string]models.PermissionRow,
	managedIdentityNames map[string]string,
) (models.ChainPathRecord, bool) {
	control, ok := persistenceAutomationControl(account.ID, currentIdentityAssignments)
	if !ok {
		return models.ChainPathRecord{}, false
	}

	rowClass, ok := persistenceAutomationRowClass(account)
	if !ok {
		return models.ChainPathRecord{}, false
	}

	persistenceType := persistenceAutomationType(account)
	durability := persistenceAutomationDurability(rowClass)
	whatPersists := persistenceAutomationWhatPersists(account, rowClass)
	footholdAnchor := persistenceAutomationAnchor(account)
	currentEvidence := persistenceAutomationCurrentEvidence(account, control, rowClass)
	missingProof := persistenceAutomationMissingProof(account, rowClass)
	operatorActionability := persistenceAutomationOperatorActionability(account, rowClass)
	recommendedFixFocus := persistenceAutomationFixFocus(account)
	permission, permissionOK := persistenceAutomationPermission(account, permissionsByPrincipal)
	roleContext := persistenceAutomationRoleContext(permission, permissionOK, managedIdentityNames)
	noteText := joinSentences(currentEvidence, roleContext, operatorActionability)
	summary := joinSentences(noteText, missingProof)
	priority, urgency := persistencePriorityForAutomationRow(rowClass, permissionOK)
	nextReview := persistenceAutomationNextReview(account, rowClass)
	targetIDs, targetNames := persistenceAutomationTargets(account, permission, permissionOK, managedIdentityNames)

	return models.ChainPathRecord{
		ChainID:                        fmt.Sprintf("persistence-path::automation::%s::%s", account.ID, strings.ReplaceAll(rowClass, "_", "-")),
		AssetID:                        account.ID,
		AssetName:                      account.Name,
		AssetKind:                      "AutomationAccount",
		Location:                       account.Location,
		Surface:                        models.StringPtr("Automation Account"),
		PersistenceType:                stringPtrIf(persistenceType),
		Classification:                 models.StringPtr("event_driven_persistence"),
		Durability:                     stringPtrIf(durability),
		WhatPersists:                   stringPtrIf(whatPersists),
		FootholdAnchor:                 stringPtrIf(footholdAnchor),
		SurvivesHostRebuild:            persistenceBool(true),
		SurvivesOriginalAccountCleanup: persistenceBool(true),
		CurrentEvidence:                stringPtrIf(currentEvidence),
		MissingProof:                   stringPtrIf(missingProof),
		OperatorActionability:          stringPtrIf(operatorActionability),
		RecommendedFixFocus:            stringPtrIf(recommendedFixFocus),
		SourceCommand:                  models.StringPtr("automation"),
		SourceContext:                  stringPtrIf(control.RoleName),
		Source:                         stringPtrIf(account.Name),
		ClueType:                       "automation-persistence",
		ConfirmationBasis:              stringPtrIf(control.RoleName),
		Priority:                       priority,
		Urgency:                        stringPtrIf(urgency),
		VisiblePath:                    persistenceVisiblePath(rowClass),
		PathConcept:                    models.StringPtr("event_driven_persistence"),
		PathType:                       stringPtrIf(rowClass),
		ConfidenceBoundary:             stringPtrIf(missingProof),
		MissingConfirmation:            missingProof,
		NextReview:                     nextReview,
		Note:                           stringPtrIf(noteText),
		WhyCare:                        stringPtrIf(noteText),
		Summary:                        summary,
		TargetService:                  "automation-persistence",
		TargetResolution:               "named match",
		EvidenceCommands:               persistenceEvidenceCommands("automation", permissionOK),
		JoinedSurfaceTypes:             persistenceAutomationJoinedSurfaces(account, rowClass, permissionOK),
		TargetCount:                    len(targetIDs),
		TargetIDs:                      targetIDs,
		TargetNames:                    targetNames,
		RelatedIDs:                     mergeRelatedIDs(account.RelatedIDs, targetIDs),
		StrongerOutcome:                stringPtrIf(persistenceOutcomeText(rowClass, permission)),
	}, true
}

func persistenceAppCredentialRowShape(sourceRowClass string) (string, string, bool) {
	switch sourceRowClass {
	case "existing_credential", "federated_trust_present":
		return "existing_persistence", "durable_persistence", true
	case "directly_addable", "directly_addable_federated_trust":
		return "directly_establishable", "durable_persistence", true
	case "control_context_only":
		return "enabler_only", "enabler_only", true
	default:
		return "", "", false
	}
}

func persistenceAppCredentialType(item models.AppCredentialSummary) string {
	switch strings.TrimSpace(stringPtrValue(item.CredentialType)) {
	case "federated":
		return "application federated trust"
	case "password":
		return "service principal password credential"
	case "key":
		return item.TargetObjectType + " key credential"
	case "password-or-key":
		if strings.EqualFold(item.TargetObjectType, "ServicePrincipal") {
			return "service principal credential"
		}
		return "application credential"
	default:
		if strings.EqualFold(item.TargetObjectType, "ServicePrincipal") {
			return "service principal auth surface"
		}
		return "application auth surface"
	}
}

func persistenceAppCredentialWhatPersists(item models.AppCredentialSummary) string {
	switch strings.TrimSpace(stringPtrValue(item.CredentialType)) {
	case "federated":
		return "An Azure-facing application trust that will keep accepting the external issuer and subject pattern later"
	case "password", "key", "password-or-key":
		return "Azure-accepted authentication material on the application or service principal"
	default:
		return "A cloud-side identity object that can preserve Azure access later"
	}
}

func persistenceAppCredentialDurability(rowClass string) string {
	switch rowClass {
	case "existing_persistence":
		return "durable identity object already exists in Entra"
	case "directly_establishable":
		return "current identity can add the durable identity object now"
	default:
		return "identity control context is visible, but the durable object is not"
	}
}

func persistenceAppCredentialAnchor(item models.AppCredentialSummary) string {
	target := fmt.Sprintf("%s '%s'", item.TargetObjectType, item.TargetObjectName)
	backing := strings.TrimSpace(stringPtrValue(item.BackingServicePrincipalName))
	if backing == "" || strings.EqualFold(item.TargetObjectType, "ServicePrincipal") {
		return target
	}
	return fmt.Sprintf("%s -> service principal '%s'", target, backing)
}

func persistenceAppCredentialMissingProof(item models.AppCredentialSummary, rowClass string) string {
	if text := strings.TrimSpace(item.MissingProof); text != "" {
		return text
	}
	switch rowClass {
	case "existing_persistence":
		return "This row shows that durable identity material or trust already exists, not that the current identity can change it."
	case "directly_establishable":
		return "This row shows a direct path to add durable identity material, not that a new credential or federated trust has already been created."
	default:
		return "Current evidence shows identity control context only, and no visible path to a durable credential or federated-trust object is confirmed here."
	}
}

func persistenceAppCredentialPermission(item models.AppCredentialSummary, permissionsByPrincipal map[string]models.PermissionRow) (models.PermissionRow, bool) {
	if item.BackingServicePrincipalID != nil {
		if permission, ok := permissionsByPrincipal[*item.BackingServicePrincipalID]; ok {
			return permission, permission.Privileged
		}
	}
	if strings.EqualFold(item.TargetObjectType, "ServicePrincipal") {
		if permission, ok := permissionsByPrincipal[item.TargetObjectID]; ok {
			return permission, permission.Privileged
		}
	}
	return models.PermissionRow{}, false
}

func persistenceAppCredentialTargets(item models.AppCredentialSummary) ([]string, []string) {
	targetIDs := []string{item.TargetObjectID}
	targetNames := []string{item.TargetObjectName}
	if item.BackingServicePrincipalID != nil && strings.TrimSpace(*item.BackingServicePrincipalID) != "" && *item.BackingServicePrincipalID != item.TargetObjectID {
		targetIDs = append(targetIDs, *item.BackingServicePrincipalID)
	}
	if item.BackingServicePrincipalName != nil && strings.TrimSpace(*item.BackingServicePrincipalName) != "" && !strings.EqualFold(*item.BackingServicePrincipalName, item.TargetObjectName) {
		targetNames = append(targetNames, *item.BackingServicePrincipalName)
	}
	return dedupeStrings(targetIDs), dedupeStrings(targetNames)
}

func persistenceAppCredentialJoinedSurfaces(item models.AppCredentialSummary, permissionOK bool) []string {
	surfaces := []string{"identity-surface"}
	switch strings.TrimSpace(stringPtrValue(item.CredentialType)) {
	case "federated":
		surfaces = append(surfaces, "federated-trust")
	default:
		surfaces = append(surfaces, "credential-material")
	}
	if permissionOK {
		surfaces = append(surfaces, "role-bearing-context")
	}
	return dedupeStrings(surfaces)
}

func persistencePriorityForIdentityRow(rowClass string, permissionOK bool) (string, string) {
	switch rowClass {
	case "existing_persistence":
		if permissionOK {
			return "high", "pivot-now"
		}
		return "medium", "review-soon"
	case "directly_establishable":
		if permissionOK {
			return "high", "pivot-now"
		}
		return "medium", "review-soon"
	default:
		return "low", "bookmark"
	}
}

func persistenceAppCredentialNextReview(item models.AppCredentialSummary, rowClass string) string {
	switch rowClass {
	case "existing_persistence":
		return fmt.Sprintf("Review who can modify %s '%s' and whether this durable credential or trust is still required.", item.TargetObjectType, item.TargetObjectName)
	case "directly_establishable":
		return fmt.Sprintf("Remove the visible owner or credential-mutation path that lets the current identity control %s '%s'.", item.TargetObjectType, item.TargetObjectName)
	default:
		return fmt.Sprintf("Keep this row weak unless current evidence later shows a visible credential or federated-trust mutation path on %s '%s'.", item.TargetObjectType, item.TargetObjectName)
	}
}

func persistenceAutomationControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	bestRank := 99
	best := persistenceCurrentIdentityControl{}
	for _, assignment := range assignments {
		role := strings.ToLower(strings.TrimSpace(assignment.RoleName))
		if role != "owner" && role != "contributor" && role != "automation contributor" {
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

func persistenceScopeRank(scopeID string, resourceID string) (int, bool) {
	scopeID = strings.TrimSpace(scopeID)
	resourceID = strings.TrimSpace(resourceID)
	if scopeID == "" || resourceID == "" {
		return 0, false
	}
	scopeLower := strings.ToLower(strings.TrimRight(scopeID, "/"))
	resourceLower := strings.ToLower(strings.TrimRight(resourceID, "/"))
	if scopeLower == resourceLower {
		return 0, true
	}
	if !strings.HasPrefix(resourceLower, scopeLower+"/") {
		return 0, false
	}
	if strings.Contains(scopeLower, "/resourcegroups/") {
		return 1, true
	}
	if strings.Contains(scopeLower, "/subscriptions/") {
		return 2, true
	}
	return 3, true
}

func persistenceScopeLabel(scopeID string) string {
	if strings.Contains(scopeID, "/subscriptions/") && !strings.Contains(scopeID, "/resourceGroups/") {
		return "subscription scope"
	}
	if strings.Contains(scopeID, "/resourceGroups/") {
		return "resource group " + armScopeName(scopeID)
	}
	return "a parent scope of this resource"
}

func persistenceAutomationRowClass(account models.AutomationAccountAsset) (string, bool) {
	if intPtrValue(account.PublishedRunbookCount) <= 0 {
		return "", false
	}
	if intPtrValue(account.WebhookCount) > 0 || intPtrValue(account.JobScheduleCount) > 0 || intPtrValue(account.ScheduleCount) > 0 {
		return "existing_persistence", true
	}
	if len(account.PublishedRunbookNames) > 0 {
		return "near_complete_setup", true
	}
	return "", false
}

func persistenceAutomationType(account models.AutomationAccountAsset) string {
	parts := []string{}
	if intPtrValue(account.WebhookCount) > 0 {
		parts = append(parts, "Automation webhook")
	}
	if intPtrValue(account.JobScheduleCount) > 0 || intPtrValue(account.ScheduleCount) > 0 {
		parts = append(parts, "Automation schedule")
	}
	if len(parts) == 0 {
		return "Automation runbook surface"
	}
	return strings.Join(parts, " + ")
}

func persistenceAutomationDurability(rowClass string) string {
	switch rowClass {
	case "existing_persistence":
		return "durable Automation trigger object already exists"
	default:
		return "published runbooks are visible, but one durable trigger object is still missing"
	}
}

func persistenceAutomationWhatPersists(account models.AutomationAccountAsset, rowClass string) string {
	if rowClass == "near_complete_setup" {
		return "The Automation control plane and published runbooks are one durable schedule or webhook away from cloud-side re-entry"
	}
	if intPtrValue(account.WebhookCount) > 0 && (intPtrValue(account.JobScheduleCount) > 0 || intPtrValue(account.ScheduleCount) > 0) {
		return "Azure keeps webhook and schedule objects that can re-run published runbooks later"
	}
	if intPtrValue(account.WebhookCount) > 0 {
		return "Azure keeps a webhook object that can re-run published runbooks later"
	}
	return "Azure keeps a schedule object that can re-run published runbooks later"
}

func persistenceAutomationAnchor(account models.AutomationAccountAsset) string {
	if runbook := strings.TrimSpace(stringPtrValue(account.PrimaryRunbookName)); runbook != "" {
		return fmt.Sprintf("Automation Account '%s' (runbook '%s')", account.Name, runbook)
	}
	return fmt.Sprintf("Automation Account '%s'", account.Name)
}

func persistenceAutomationCurrentEvidence(account models.AutomationAccountAsset, control persistenceCurrentIdentityControl, rowClass string) string {
	triggerParts := []string{}
	if intPtrValue(account.WebhookCount) > 0 {
		triggerParts = append(triggerParts, fmt.Sprintf("%d webhook(s)", intPtrValue(account.WebhookCount)))
	}
	if intPtrValue(account.JobScheduleCount) > 0 {
		triggerParts = append(triggerParts, fmt.Sprintf("%d job schedule(s)", intPtrValue(account.JobScheduleCount)))
	} else if intPtrValue(account.ScheduleCount) > 0 {
		triggerParts = append(triggerParts, fmt.Sprintf("%d schedule(s)", intPtrValue(account.ScheduleCount)))
	}
	runbookPhrase := fmt.Sprintf("%d published runbook(s)", intPtrValue(account.PublishedRunbookCount))
	if primary := strings.TrimSpace(stringPtrValue(account.PrimaryRunbookName)); primary != "" {
		runbookPhrase += fmt.Sprintf(" including '%s'", primary)
	}
	if rowClass == "near_complete_setup" {
		return fmt.Sprintf("Current identity has %s over Automation Account '%s', and the account already exposes %s but no visible durable schedule or webhook is confirmed here.", control.RoleName, account.Name, runbookPhrase)
	}
	return fmt.Sprintf("Current identity has %s over Automation Account '%s', and the account already exposes %s backed by %s.", control.RoleName, account.Name, runbookPhrase, naturalJoin(triggerParts))
}

func persistenceAutomationMissingProof(account models.AutomationAccountAsset, rowClass string) string {
	parts := []string{}
	if rowClass == "near_complete_setup" {
		parts = append(parts, "Current evidence shows edit control and published runbooks, but no visible durable schedule or webhook closes the persistence path yet.")
	} else {
		parts = append(parts, "This row shows durable trigger-backed re-entry from control-plane state, not whether the current runbook content is malicious.")
	}
	if account.MissingTargetMapping {
		parts = append(parts, "The current environment does not show which workflow, app, or credential target this Automation path reaches directly.")
	}
	return joinSentences(parts...)
}

func persistenceAutomationOperatorActionability(account models.AutomationAccountAsset, rowClass string) string {
	if rowClass == "near_complete_setup" {
		return "Treat this as one durable trigger away from cloud-side persistence and review who can add or edit schedules, webhooks, and published runbooks on this Automation Account."
	}
	return "Treat this as cloud-side re-entry and review who can edit the Automation Account, its published runbooks, and its durable triggers."
}

func persistenceAutomationFixFocus(account models.AutomationAccountAsset) string {
	if intPtrValue(account.WebhookCount) > 0 || intPtrValue(account.JobScheduleCount) > 0 || intPtrValue(account.ScheduleCount) > 0 {
		return "Remove unneeded schedules or webhooks, reduce write access to the Automation Account, and rotate or trim stored Automation secrets and connections."
	}
	return "Reduce write access to the Automation Account and prevent durable schedules or webhooks from being added without review."
}

func persistenceAutomationPermission(
	account models.AutomationAccountAsset,
	permissionsByPrincipal map[string]models.PermissionRow,
) (models.PermissionRow, bool) {
	if account.PrincipalID != nil {
		if permission, ok := permissionsByPrincipal[*account.PrincipalID]; ok {
			return permission, permission.Privileged
		}
	}
	if account.ClientID != nil {
		if permission, ok := permissionsByPrincipal[*account.ClientID]; ok {
			return permission, permission.Privileged
		}
	}
	return models.PermissionRow{}, false
}

func persistenceAutomationRoleContext(permission models.PermissionRow, permissionOK bool, managedIdentityNames map[string]string) string {
	if !permissionOK {
		return ""
	}
	name := strings.TrimSpace(permission.DisplayName)
	if managedName, ok := managedIdentityNames[permission.PrincipalID]; ok && strings.TrimSpace(managedName) != "" {
		name = managedName
	}
	roleText := naturalJoin(permission.HighImpactRoles)
	if roleText == "" {
		roleText = naturalJoin(permission.AllRoleNames)
	}
	if roleText == "" {
		roleText = "high-impact Azure roles"
	}
	scopeLabel := "1 visible scope"
	if permission.ScopeCount > 1 {
		scopeLabel = fmt.Sprintf("%d visible scopes", permission.ScopeCount)
	}
	return fmt.Sprintf("The automation identity '%s' already holds %s across %s.", name, roleText, scopeLabel)
}

func persistencePriorityForAutomationRow(rowClass string, permissionOK bool) (string, string) {
	if rowClass == "existing_persistence" && permissionOK {
		return "high", "pivot-now"
	}
	if rowClass == "existing_persistence" {
		return "medium", "review-soon"
	}
	if permissionOK {
		return "medium", "review-soon"
	}
	return "low", "bookmark"
}

func persistenceAutomationNextReview(account models.AutomationAccountAsset, rowClass string) string {
	if rowClass == "near_complete_setup" {
		return fmt.Sprintf("Current identity can already edit Automation Account '%s'; keep this row weak until a durable schedule or webhook is visible.", account.Name)
	}
	if account.MissingTargetMapping {
		return fmt.Sprintf("Review the runbooks, schedules, and webhooks on Automation Account '%s'; current evidence does not show which downstream app or workflow this path reaches directly.", account.Name)
	}
	return fmt.Sprintf("Review the runbooks, schedules, and webhooks on Automation Account '%s' and tighten who can edit them.", account.Name)
}

func persistenceAutomationTargets(
	account models.AutomationAccountAsset,
	permission models.PermissionRow,
	permissionOK bool,
	managedIdentityNames map[string]string,
) ([]string, []string) {
	targetIDs := []string{account.ID}
	targetNames := []string{account.Name}
	if account.PrincipalID != nil && strings.TrimSpace(*account.PrincipalID) != "" {
		targetIDs = append(targetIDs, *account.PrincipalID)
		identityName := strings.TrimSpace(permission.DisplayName)
		if managedName, ok := managedIdentityNames[*account.PrincipalID]; ok && strings.TrimSpace(managedName) != "" {
			identityName = managedName
		}
		if permissionOK && identityName != "" {
			targetNames = append(targetNames, identityName)
		}
	}
	return dedupeStrings(targetIDs), dedupeStrings(targetNames)
}

func persistenceAutomationJoinedSurfaces(account models.AutomationAccountAsset, rowClass string, permissionOK bool) []string {
	surfaces := []string{"automation-account", "published-runbook"}
	if intPtrValue(account.WebhookCount) > 0 {
		surfaces = append(surfaces, "webhook")
	}
	if intPtrValue(account.JobScheduleCount) > 0 || intPtrValue(account.ScheduleCount) > 0 {
		surfaces = append(surfaces, "schedule")
	}
	if permissionOK {
		surfaces = append(surfaces, "role-bearing-context")
	}
	if rowClass == "near_complete_setup" {
		surfaces = append(surfaces, "trigger-gap")
	}
	return dedupeStrings(surfaces)
}

func persistenceOutcomeText(rowClass string, permission models.PermissionRow) string {
	switch rowClass {
	case "existing_persistence":
		if permission.Privileged {
			return fmt.Sprintf("Durable cloud-side persistence already exists and preserves access into a principal that already holds %s.", naturalJoin(permission.HighImpactRoles))
		}
		return "Durable cloud-side persistence already exists."
	case "directly_establishable":
		if permission.Privileged {
			return fmt.Sprintf("Current identity can directly establish durable cloud persistence that would preserve access into a principal that already holds %s.", naturalJoin(permission.HighImpactRoles))
		}
		return "Current identity can directly establish durable cloud persistence."
	case "near_complete_setup":
		return "Current identity is one durable object away from cloud-side persistence."
	default:
		return "Current evidence shows a persistence enabler, not durable persistence by itself."
	}
}

func persistenceEvidenceCommands(primary string, includePermissions bool) []string {
	commands := []string{primary}
	if includePermissions {
		commands = append(commands, "permissions")
	}
	if primary == "automation" {
		commands = append(commands, "rbac")
	}
	return commands
}

func persistenceVisiblePath(rowClass string) string {
	switch rowClass {
	case "existing_persistence":
		return "durable cloud-side object already exists"
	case "directly_establishable":
		return "current identity can add durable object now"
	case "near_complete_setup":
		return "service is one durable object away"
	default:
		return "supporting persistence context only"
	}
}

func persistenceRowRank(rowClass string) int {
	switch rowClass {
	case "existing_persistence":
		return 0
	case "directly_establishable":
		return 1
	case "near_complete_setup":
		return 2
	case "enabler_only":
		return 3
	default:
		return 9
	}
}

func persistenceBool(value bool) *bool {
	return &value
}

func joinSentences(parts ...string) string {
	clean := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		clean = append(clean, part)
	}
	return strings.Join(clean, " ")
}
