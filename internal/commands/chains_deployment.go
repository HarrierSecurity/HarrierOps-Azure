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

type deploymentTargetSpec struct {
	Command string
	Label   string
	Service string
}

type deploymentTarget struct {
	ID         string
	Name       string
	Location   *string
	Summary    string
	Providers  []string
	RelatedIDs []string
}

var (
	deploymentTargetSpecs = map[string]deploymentTargetSpec{
		"aks":             {Command: "aks", Label: "AKS cluster", Service: "aks"},
		"app-services":    {Command: "app-services", Label: "App Service", Service: "app-service"},
		"functions":       {Command: "functions", Label: "Function App", Service: "function-app"},
		"arm-deployments": {Command: "arm-deployments", Label: "ARM deployment", Service: "arm-deployment"},
	}
	deploymentCanonicalConfirmationBases = map[string]struct{}{
		"asset-id-match":            {},
		"resource-id-match":         {},
		"principal-id-match":        {},
		"managed-identity-id-match": {},
		"normalized-uri-match":      {},
		"parsed-config-target":      {},
	}
	devopsDeploymentHintMap = map[string]string{
		"aks/kubernetes":      "aks",
		"app service":         "app-services",
		"azure functions":     "functions",
		"function app":        "functions",
		"functions":           "functions",
		"arm/bicep/terraform": "arm-deployments",
	}
	automationTargetFamilies = []string{"app-services", "functions", "aks"}
	automationTargetTokens   = map[string][]string{
		"app-services": {"app", "api", "site", "web"},
		"functions":    {"func", "function", "functions", "webjob"},
		"aks":          {"aks", "cluster", "k8s", "kube", "kubernetes"},
	}
	automationStopwords = map[string]struct{}{
		"account": {}, "apply": {}, "baseline": {}, "config": {}, "configure": {}, "maintenance": {},
		"nightly": {}, "reapply": {}, "reconcile": {}, "redeploy": {}, "rotate": {}, "runbook": {},
		"sync": {},
	}
	deploymentJoinQualityOrder = map[string]int{
		"named match":         0,
		"narrowed candidates": 1,
		"visibility blocked":  2,
	}
)

func buildDeploymentPathOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	family contracts.FamilyContract,
) (models.ChainsOutput, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	devopsFuture := runGroupedCommandOutput[models.DevopsOutput](group, ctx, request, devopsHandler(provider, now), "devops")
	automationFuture := runGroupedCommandOutput[models.AutomationOutput](group, ctx, request, automationHandler(provider, now), "automation")
	permissionsFuture := runGroupedCommandOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions")
	rbacFuture := runGroupedCommandOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac")
	roleTrustsFuture := runGroupedCommandOutput[models.RoleTrustsOutput](group, ctx, request, roleTrustsHandler(provider, now), "role-trusts")
	keyvaultFuture := runGroupedCommandOutput[models.KeyVaultOutput](group, ctx, request, keyVaultHandler(provider, now), "keyvault")
	armDeploymentsFuture := runGroupedCommandOutput[models.ArmDeploymentsOutput](group, ctx, request, armDeploymentsHandler(provider, now), "arm-deployments")
	aksFuture := runGroupedCommandOutput[models.AksOutput](group, ctx, request, aksHandler(provider, now), "aks")
	functionsFuture := runGroupedCommandOutput[models.FunctionsOutput](group, ctx, request, functionsHandler(provider, now), "functions")
	appServicesFuture := runGroupedCommandOutput[models.AppServicesOutput](group, ctx, request, appServicesHandler(provider, now), "app-services")

	devops, err := devopsFuture.wait()
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
	roleTrusts, err := roleTrustsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	keyvault, err := keyvaultFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	armDeployments, err := armDeploymentsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	aks, err := aksFuture.wait()
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

	targetsByFamily := map[string][]deploymentTarget{
		"aks":             deploymentAksTargets(aks),
		"app-services":    deploymentAppServiceTargets(appServices),
		"functions":       deploymentFunctionTargets(functions),
		"arm-deployments": deploymentArmTargets(armDeployments),
	}
	targetVisibilityNotes := map[string]*string{
		"aks":             targetVisibilityNote("AKS", aks.Issues),
		"app-services":    targetVisibilityNote("App Service", appServices.Issues),
		"functions":       targetVisibilityNote("Function App", functions.Issues),
		"arm-deployments": targetVisibilityNote("ARM deployment", armDeployments.Issues),
	}
	targetVisibilityIssues := map[string]*string{
		"aks":             targetVisibilityIssue(aks.Issues),
		"app-services":    targetVisibilityIssue(appServices.Issues),
		"functions":       targetVisibilityIssue(functions.Issues),
		"arm-deployments": targetVisibilityIssue(armDeployments.Issues),
	}

	permissionsByPrincipal := map[string]models.PermissionRow{}
	currentIdentityPrincipals := map[string]struct{}{}
	for _, permission := range permissions.Permissions {
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

	trustsBySource := map[string][]models.RoleTrustSummary{}
	trustsByTarget := map[string][]models.RoleTrustSummary{}
	for _, trust := range roleTrusts.Trusts {
		if trust.SourceObjectID != "" {
			trustsBySource[trust.SourceObjectID] = append(trustsBySource[trust.SourceObjectID], trust)
		}
		if trust.TargetObjectID != "" {
			trustsByTarget[trust.TargetObjectID] = append(trustsByTarget[trust.TargetObjectID], trust)
		}
	}

	keyVaultsByName := map[string]models.KeyVaultAsset{}
	for _, vault := range keyvault.KeyVaults {
		keyVaultsByName[strings.ToLower(vault.Name)] = vault
	}

	supportingDeployments := deploymentSupportingDeployments(armDeployments)

	paths := make([]models.ChainPathRecord, 0)
	for _, pipeline := range devops.Pipelines {
		for _, familyName := range devopsTargetFamilies(pipeline) {
			record, ok := buildDevopsDeploymentRecord(
				pipeline,
				familyName,
				targetsByFamily,
				targetVisibilityNotes[familyName],
				targetVisibilityIssues[familyName],
				permissionsByPrincipal,
				trustsBySource,
				trustsByTarget,
				keyVaultsByName,
				supportingDeployments[familyName],
			)
			if ok {
				paths = append(paths, record)
			}
		}
	}

	for _, account := range automation.AutomationAccounts {
		record, ok := buildAutomationDeploymentRecord(
			account,
			targetsByFamily,
			targetVisibilityNotes,
			targetVisibilityIssues,
			permissionsByPrincipal,
			currentIdentityAssignments,
			trustsBySource,
			supportingDeployments,
		)
		if ok {
			paths = append(paths, record)
		}
	}

	sort.SliceStable(paths, func(i int, j int) bool {
		left := paths[i]
		right := paths[j]
		if left.Priority != right.Priority {
			return prioritySortValue(left.Priority) < prioritySortValue(right.Priority)
		}
		if left.TargetResolution != right.TargetResolution {
			return deploymentJoinQualityOrder[left.TargetResolution] < deploymentJoinQualityOrder[right.TargetResolution]
		}
		if left.AssetName != right.AssetName {
			return left.AssetName < right.AssetName
		}
		return left.TargetService < right.TargetService
	})

	issues := append([]models.Issue{}, devops.Issues...)
	issues = append(issues, automation.Issues...)
	issues = append(issues, permissions.Issues...)
	issues = append(issues, rbac.Issues...)
	issues = append(issues, roleTrusts.Issues...)
	issues = append(issues, keyvault.Issues...)
	issues = append(issues, armDeployments.Issues...)
	issues = append(issues, aks.Issues...)
	issues = append(issues, functions.Issues...)
	issues = append(issues, appServices.Issues...)

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

func deploymentAksTargets(output models.AksOutput) []deploymentTarget {
	targets := make([]deploymentTarget, 0, len(output.AksClusters))
	for _, item := range output.AksClusters {
		targets = append(targets, deploymentTarget{ID: item.ID, Name: item.Name, Location: item.Location, Summary: item.Summary, RelatedIDs: item.RelatedIDs})
	}
	return targets
}

func deploymentAppServiceTargets(output models.AppServicesOutput) []deploymentTarget {
	targets := make([]deploymentTarget, 0, len(output.AppServices))
	for _, item := range output.AppServices {
		location := models.StringPtr(item.Location)
		targets = append(targets, deploymentTarget{ID: item.ID, Name: item.Name, Location: location, Summary: item.Summary, RelatedIDs: item.RelatedIDs})
	}
	return targets
}

func deploymentFunctionTargets(output models.FunctionsOutput) []deploymentTarget {
	targets := make([]deploymentTarget, 0, len(output.FunctionApps))
	for _, item := range output.FunctionApps {
		location := models.StringPtr(item.Location)
		targets = append(targets, deploymentTarget{ID: item.ID, Name: item.Name, Location: location, Summary: item.Summary, RelatedIDs: item.RelatedIDs})
	}
	return targets
}

func deploymentArmTargets(output models.ArmDeploymentsOutput) []deploymentTarget {
	targets := make([]deploymentTarget, 0, len(output.Deployments))
	for _, item := range output.Deployments {
		targets = append(targets, deploymentTarget{ID: item.ID, Name: item.Name, Summary: item.Summary, Providers: append([]string{}, item.Providers...), RelatedIDs: item.RelatedIDs})
	}
	return targets
}

func deploymentSupportingDeployments(output models.ArmDeploymentsOutput) map[string][]models.ArmDeploymentSummary {
	byFamily := map[string][]models.ArmDeploymentSummary{
		"aks":             {},
		"app-services":    {},
		"functions":       {},
		"arm-deployments": append([]models.ArmDeploymentSummary{}, output.Deployments...),
	}
	for _, item := range output.Deployments {
		for _, providerName := range item.Providers {
			switch strings.ToLower(strings.TrimSpace(providerName)) {
			case "microsoft.web":
				byFamily["app-services"] = append(byFamily["app-services"], item)
				byFamily["functions"] = append(byFamily["functions"], item)
			case "microsoft.containerservice":
				byFamily["aks"] = append(byFamily["aks"], item)
			}
		}
	}
	return byFamily
}

func devopsTargetFamilies(pipeline models.DevopsPipelineAsset) []string {
	seen := map[string]struct{}{}
	families := make([]string, 0)
	for _, clue := range pipeline.TargetClues {
		key := strings.ToLower(strings.TrimSpace(clue))
		if index := strings.Index(key, ":"); index >= 0 {
			key = strings.TrimSpace(key[:index])
		}
		familyName, ok := devopsDeploymentHintMap[key]
		if !ok {
			continue
		}
		if _, exists := seen[familyName]; exists {
			continue
		}
		seen[familyName] = struct{}{}
		families = append(families, familyName)
	}
	sort.Strings(families)
	return families
}

func buildDevopsDeploymentRecord(
	pipeline models.DevopsPipelineAsset,
	familyName string,
	targetsByFamily map[string][]deploymentTarget,
	targetVisibilityNote *string,
	targetVisibilityIssue *string,
	permissionsByPrincipal map[string]models.PermissionRow,
	trustsBySource map[string][]models.RoleTrustSummary,
	trustsByTarget map[string][]models.RoleTrustSummary,
	keyVaultsByName map[string]models.KeyVaultAsset,
	supporting []models.ArmDeploymentSummary,
) (models.ChainPathRecord, bool) {
	spec := deploymentTargetSpecs[familyName]
	exactTargets, confirmationBasis := devopsExactTargets(pipeline, familyName, targetsByFamily[familyName])
	selectedTargets := exactTargets
	if len(selectedTargets) == 0 && targetVisibilityIssue == nil {
		selectedTargets = append([]deploymentTarget{}, targetsByFamily[familyName]...)
	}
	targetResolution := deploymentTargetResolution(selectedTargets, confirmationBasis, targetVisibilityIssue, pipeline.MissingTargetMapping)
	if targetResolution == "" {
		return models.ChainPathRecord{}, false
	}

	permission, permissionOK := deploymentJoinedPermission(pipeline.AzureServiceConnectionPrincipalIDs, permissionsByPrincipal)
	actionability, actionabilityState, priority, urgency := devopsActionability(pipeline)
	insertPoint, insertPointDisplay := devopsInsertionPoint(pipeline, actionability)
	targetNames := deploymentTargetNames(selectedTargets, targetResolution, targetVisibilityIssue)
	targetIDs := deploymentTargetIDs(selectedTargets, targetResolution, targetVisibilityIssue)
	likelyImpact := deploymentLikelyImpact(spec, targetNames, targetResolution, pipeline.MissingTargetMapping)
	executionIdentityName := devopsExecutionIdentityName(pipeline, permission, permissionOK, trustsBySource, trustsByTarget)
	confidenceBoundary, missingConfirmation := devopsConfidenceBoundary(spec, pipeline, executionIdentityName, targetResolution)
	nextReview := devopsNextReview(spec, pipeline, targetNames, actionability, permissionOK, supporting)
	whyCare := devopsWhyCare(
		pipeline,
		permission,
		permissionOK,
		executionIdentityName,
		trustsBySource,
		trustsByTarget,
		keyVaultsByName,
		selectedTargets,
		spec,
		familyName,
		targetResolution,
	)
	note := whyCare
	summary := whyCare
	if sentence := deploymentLikelyImpactSentence(spec, targetNames, targetResolution); sentence != "" {
		summary += " " + sentence
	}
	if confidenceBoundary != "" {
		summary += " " + confidenceBoundary
	}
	if len(supporting) > 0 && familyName != "arm-deployments" {
		summary += " " + deploymentSupportingSentence(supporting, spec)
	}

	sourceValue := deploymentDisplaySource(pipeline.Name)
	record := models.ChainPathRecord{
		ChainID:             "deployment-path::" + pipeline.ID + "::" + spec.Service,
		AssetID:             pipeline.ID,
		AssetName:           pipeline.Name,
		AssetKind:           "DevOpsPipeline",
		SourceCommand:       models.StringPtr("devops"),
		SourceContext:       models.StringPtr(pipeline.ProjectName),
		Source:              &sourceValue,
		ClueType:            "controllable-change-path",
		ConfirmationBasis:   stringPtrIf(confirmationBasis),
		Priority:            priority,
		Urgency:             models.StringPtr(urgency),
		Actionability:       models.StringPtr(actionability),
		ActionabilityState:  models.StringPtr(actionabilityState),
		VisiblePath:         "Controllable Azure pipeline -> likely " + spec.Label,
		InsertionPoint:      models.StringPtr(insertPoint),
		InsertionPointLabel: models.StringPtr(insertPointDisplay),
		PathConcept:         models.StringPtr("controllable-change-path"),
		PrimarySurface:      stringPtrIf(pipeline.PrimaryInjectionSurface),
		PrimaryInputRef:     stringPtrIf(pipeline.PrimaryTrustedInputRef),
		LikelyImpact:        stringPtrIf(likelyImpact),
		LikelyAzureImpact:   stringPtrIf(likelyImpact),
		ConfidenceBoundary:  stringPtrIf(confidenceBoundary),
		WhatsMissing:        stringPtrIf(confidenceBoundary),
		MissingConfirmation: missingConfirmation,
		NextReview:          nextReview,
		Note:                stringPtrIf(note),
		WhyCare:             stringPtrIf(whyCare),
		Summary:             summary,
		TargetService:       spec.Service,
		TargetResolution:    targetResolution,
		EvidenceCommands:    devopsEvidenceCommands(familyName, pipeline, permissionOK, supporting),
		JoinedSurfaceTypes:  devopsJoinedSurfaceTypes(pipeline, permissionOK, familyName, supporting),
		TargetCount:         len(targetIDs),
		TargetIDs:           targetIDs,
		TargetNames:         targetNames,
		TargetVisibility:    targetVisibilityIssue,
		RelatedIDs:          mergeRelatedIDs(pipeline.RelatedIDs, targetIDs, deploymentSupportingIDs(supporting)),
	}
	if targetVisibilityNote != nil && record.Summary != "" && targetResolution != "named match" && targetResolution != "narrowed candidates" {
		record.Summary += " " + *targetVisibilityNote
	}
	return record, true
}

func buildAutomationDeploymentRecord(
	account models.AutomationAccountAsset,
	targetsByFamily map[string][]deploymentTarget,
	targetVisibilityNotes map[string]*string,
	targetVisibilityIssues map[string]*string,
	permissionsByPrincipal map[string]models.PermissionRow,
	currentIdentityAssignments []models.RoleAssignment,
	trustsBySource map[string][]models.RoleTrustSummary,
	supporting map[string][]models.ArmDeploymentSummary,
) (models.ChainPathRecord, bool) {
	posture := automationPosture(account)
	if posture == "" {
		return models.ChainPathRecord{}, false
	}

	targetFamily, selectedTargets, confirmationBasis := automationTargetSelection(account, targetsByFamily, supporting)
	if targetFamily == "" {
		targetFamily = "arm-deployments"
		selectedTargets = nil
		confirmationBasis = ""
	}
	spec := deploymentTargetSpecs[targetFamily]
	targetVisibilityIssue := targetVisibilityIssues[targetFamily]
	if posture == "support-only" && targetVisibilityIssue == nil {
		targetVisibilityIssue = models.StringPtr("current automation surface does not name downstream Azure targets")
	}
	targetResolution := deploymentTargetResolution(selectedTargets, confirmationBasis, targetVisibilityIssue, account.MissingTargetMapping)
	if posture == "support-only" {
		targetResolution = "visibility blocked"
	}
	if targetResolution == "" {
		return models.ChainPathRecord{}, false
	}

	permission, permissionOK := deploymentAutomationPermission(account, permissionsByPrincipal)
	actionability, actionabilityState, priority, urgency := automationActionability(account, posture, currentIdentityAssignments)
	insertPoint, insertPointDisplay := automationInsertionPoint(account, currentIdentityAssignments)
	targetNames := deploymentTargetNames(selectedTargets, targetResolution, targetVisibilityIssue)
	targetIDs := deploymentTargetIDs(selectedTargets, targetResolution, targetVisibilityIssue)
	likelyImpact := deploymentLikelyImpact(spec, targetNames, targetResolution, posture == "support-only")
	confidenceBoundary, missingConfirmation := automationConfidenceBoundary(spec, posture, permissionOK, targetResolution)
	nextReview := automationNextReview(spec, posture, targetNames, permissionOK, supporting[targetFamily])
	whyCare := automationWhyCare(account, posture, permission, permissionOK, trustsBySource, selectedTargets, spec, targetResolution)
	note := whyCare
	summary := whyCare
	if posture == "support-only" {
		summary += " HO-Azure has not yet mapped the downstream Azure footprint cleanly, so current visible ARM deployment clues only narrow the downstream story right now."
	}
	if sentence := deploymentLikelyImpactSentence(spec, targetNames, targetResolution); sentence != "" {
		summary += " " + sentence
	}
	if confidenceBoundary != "" {
		summary += " " + confidenceBoundary
	}
	if len(supporting[targetFamily]) > 0 && targetFamily != "arm-deployments" {
		summary += " " + deploymentSupportingSentence(supporting[targetFamily], spec)
	}

	sourceValue := deploymentDisplaySource(account.Name)
	record := models.ChainPathRecord{
		ChainID:             "deployment-path::" + account.ID + "::" + spec.Service,
		AssetID:             account.ID,
		AssetName:           account.Name,
		AssetKind:           "AutomationAccount",
		Location:            account.Location,
		SourceCommand:       models.StringPtr("automation"),
		SourceContext:       account.IdentityType,
		Source:              &sourceValue,
		ClueType:            deploymentAutomationClueType(posture),
		ConfirmationBasis:   stringPtrIf(confirmationBasis),
		Priority:            priority,
		Urgency:             models.StringPtr(urgency),
		Actionability:       models.StringPtr(actionability),
		ActionabilityState:  models.StringPtr(actionabilityState),
		VisiblePath:         deploymentAutomationVisiblePath(posture, spec),
		InsertionPoint:      models.StringPtr(insertPoint),
		InsertionPointLabel: models.StringPtr(insertPointDisplay),
		PathConcept:         models.StringPtr(deploymentAutomationClueType(posture)),
		LikelyImpact:        stringPtrIf(likelyImpact),
		LikelyAzureImpact:   stringPtrIf(likelyImpact),
		ConfidenceBoundary:  stringPtrIf(confidenceBoundary),
		WhatsMissing:        stringPtrIf(confidenceBoundary),
		MissingConfirmation: missingConfirmation,
		NextReview:          nextReview,
		Note:                stringPtrIf(note),
		WhyCare:             stringPtrIf(whyCare),
		Summary:             summary,
		TargetService:       spec.Service,
		TargetResolution:    targetResolution,
		EvidenceCommands:    automationEvidenceCommands(targetFamily, permissionOK, supporting[targetFamily]),
		JoinedSurfaceTypes:  automationJoinedSurfaceTypes(account, posture, targetFamily),
		TargetCount:         len(targetIDs),
		TargetIDs:           targetIDs,
		TargetNames:         targetNames,
		TargetVisibility:    targetVisibilityIssue,
		RelatedIDs:          mergeRelatedIDs(account.RelatedIDs, targetIDs, deploymentSupportingIDs(supporting[targetFamily])),
	}
	if targetVisibilityNotes[targetFamily] != nil && record.Summary != "" && targetResolution != "named match" && targetResolution != "narrowed candidates" {
		record.Summary += " " + *targetVisibilityNotes[targetFamily]
	}
	return record, true
}

func devopsExactTargets(pipeline models.DevopsPipelineAsset, familyName string, targets []deploymentTarget) ([]deploymentTarget, string) {
	exactNames := make([]string, 0)
	prefixes := []string{}
	switch familyName {
	case "app-services":
		prefixes = []string{"app service:"}
	case "functions":
		prefixes = []string{"function app:", "azure functions:"}
	case "aks":
		prefixes = []string{"aks:", "aks/kubernetes:"}
	}
	for _, clue := range pipeline.TargetClues {
		normalized := strings.ToLower(strings.TrimSpace(clue))
		for _, prefix := range prefixes {
			if strings.HasPrefix(normalized, prefix) {
				exactNames = append(exactNames, strings.TrimSpace(clue[len(prefix):]))
			}
		}
	}
	if len(exactNames) == 0 {
		return nil, ""
	}
	selected := make([]deploymentTarget, 0)
	for _, target := range targets {
		for _, exactName := range exactNames {
			if strings.EqualFold(target.Name, exactName) {
				selected = append(selected, target)
			}
		}
	}
	if len(selected) == 0 {
		return nil, ""
	}
	return dedupeDeploymentTargets(selected), "parsed-config-target"
}

func deploymentTargetResolution(selected []deploymentTarget, confirmationBasis string, visibilityIssue *string, missingTargetMapping bool) string {
	if len(selected) == 1 {
		if _, ok := deploymentCanonicalConfirmationBases[confirmationBasis]; ok {
			return "named match"
		}
		return "narrowed candidates"
	}
	if len(selected) > 1 {
		return "narrowed candidates"
	}
	if visibilityIssue != nil || missingTargetMapping {
		return "visibility blocked"
	}
	return ""
}

func deploymentTargetNames(selected []deploymentTarget, targetResolution string, visibilityIssue *string) []string {
	if visibilityIssue != nil && targetResolution == "visibility blocked" {
		return []string{}
	}
	names := make([]string, 0, len(selected))
	for _, target := range selected {
		if target.Name != "" {
			names = append(names, target.Name)
		}
	}
	return names
}

func deploymentTargetIDs(selected []deploymentTarget, targetResolution string, visibilityIssue *string) []string {
	if visibilityIssue != nil && targetResolution == "visibility blocked" {
		return []string{}
	}
	ids := make([]string, 0, len(selected))
	for _, target := range selected {
		if target.ID != "" {
			ids = append(ids, target.ID)
		}
	}
	return ids
}

func deploymentJoinedPermission(principalIDs []string, permissionsByPrincipal map[string]models.PermissionRow) (models.PermissionRow, bool) {
	for _, principalID := range principalIDs {
		if permission, ok := permissionsByPrincipal[principalID]; ok {
			return permission, true
		}
	}
	return models.PermissionRow{}, false
}

func deploymentAutomationPermission(account models.AutomationAccountAsset, permissionsByPrincipal map[string]models.PermissionRow) (models.PermissionRow, bool) {
	if account.PrincipalID != nil {
		if permission, ok := permissionsByPrincipal[*account.PrincipalID]; ok {
			return permission, true
		}
	}
	if account.ClientID != nil {
		if permission, ok := permissionsByPrincipal[*account.ClientID]; ok {
			return permission, true
		}
	}
	return models.PermissionRow{}, false
}

func devopsActionability(pipeline models.DevopsPipelineAsset) (string, string, string, string) {
	if pipeline.CurrentOperatorCanContributeSource != nil && *pipeline.CurrentOperatorCanContributeSource {
		return "currently actionable", "currently actionable", "high", "pivot-now"
	}
	if pipeline.CurrentOperatorCanEdit != nil && *pipeline.CurrentOperatorCanEdit {
		return "currently actionable", "currently actionable", "high", "pivot-now"
	}
	if pipeline.CurrentOperatorCanQueue != nil && *pipeline.CurrentOperatorCanQueue {
		return "conditionally actionable", "conditionally actionable", "medium", "review-soon"
	}
	return "grounded, insertion unproven", "consequence-grounded but insertion point unproven", "medium", "review-soon"
}

func devopsInsertionPoint(pipeline models.DevopsPipelineAsset, actionability string) (string, string) {
	ref := deploymentTrustedInputLabel(pipeline.PrimaryTrustedInputRef)
	if actionability == "currently actionable" {
		surface := pipeline.PrimaryInjectionSurface
		if surface == "" {
			surface = "trusted input"
		}
		if containsString(pipeline.CurrentOperatorInjectionSurfaceTypes, "pull-request") {
			return fmt.Sprintf("Poison %s through %s, pull-request.", ref, surface), fmt.Sprintf("Poison %s\nthrough %s,\npull-request.", ref, surface)
		}
		return fmt.Sprintf("Poison %s through %s.", ref, surface), fmt.Sprintf("Poison %s\nthrough %s.", ref, surface)
	}
	if actionability == "conditionally actionable" {
		return fmt.Sprintf("Queue this pipeline now; %s is only readable.", ref), fmt.Sprintf("Queue this pipeline now;\n%s is only readable.", ref)
	}
	if strings.HasPrefix(pipeline.PrimaryTrustedInputRef, "pipeline-artifact:") {
		return fmt.Sprintf("Current scope only shows that %s is referenced here.", ref), fmt.Sprintf("Current scope only shows that %s is referenced here.", ref)
	}
	return fmt.Sprintf("Current scope only shows the external reference %s.", ref), fmt.Sprintf("Current scope only shows the external reference %s.", ref)
}

func deploymentTrustedInputLabel(ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "trusted input"
	}
	if strings.HasPrefix(ref, "repository:") {
		return "repository " + strings.TrimPrefix(ref, "repository:")
	}
	if strings.HasPrefix(ref, "pipeline-artifact:") {
		return "pipeline artifact " + strings.TrimPrefix(ref, "pipeline-artifact:")
	}
	return ref
}

func deploymentLikelyImpact(spec deploymentTargetSpec, targetNames []string, targetResolution string, missingTargetMapping bool) string {
	if targetResolution == "named match" && len(targetNames) == 1 {
		return "exact " + strings.ToLower(spec.Label) + ": " + targetNames[0]
	}
	if targetResolution == "narrowed candidates" && len(targetNames) > 0 {
		return fmt.Sprintf("%d visible %s candidate(s): %s", len(targetNames), strings.ToLower(spec.Label), strings.Join(targetNames, ", "))
	}
	if targetResolution == "visibility blocked" || missingTargetMapping {
		return "Azure footprint not yet mapped; visible " + strings.ToLower(spec.Label) + " clues only"
	}
	return ""
}

func deploymentLikelyImpactSentence(spec deploymentTargetSpec, targetNames []string, targetResolution string) string {
	if len(targetNames) == 0 {
		return ""
	}
	switch targetResolution {
	case "named match":
		return fmt.Sprintf("The likeliest downstream Azure footprint is the exact %s target %s.", spec.Label, targetNames[0])
	case "narrowed candidates":
		return fmt.Sprintf("The likeliest downstream Azure footprint is narrowed to %d visible %s candidate(s): %s.", len(targetNames), spec.Label, strings.Join(targetNames, ", "))
	default:
		return ""
	}
}

func devopsConfidenceBoundary(spec deploymentTargetSpec, pipeline models.DevopsPipelineAsset, executionIdentityName string, targetResolution string) (string, string) {
	if (pipeline.CurrentOperatorCanContributeSource != nil && *pipeline.CurrentOperatorCanContributeSource) || (pipeline.CurrentOperatorCanEdit != nil && *pipeline.CurrentOperatorCanEdit) {
		if targetResolution == "named match" {
			identityName := "the Azure identity tied to this pipeline"
			if strings.TrimSpace(executionIdentityName) != "" {
				identityName = executionIdentityName
			}
			text := fmt.Sprintf("Current evidence shows you can poison this trusted input so it runs as Azure identity '%s' against the exact %s target, but not a separate direct sign-in as Azure identity '%s'.", identityName, spec.Label, identityName)
			return text, fmt.Sprintf("Current evidence already confirms source-side poisoning and the exact %s target; what is still missing is a separate direct sign-in as Azure identity '%s'.", spec.Label, identityName)
		}
		text := fmt.Sprintf("This row proves current-credential source poisoning and run-path control, but not the exact %s target.", spec.Label)
		return text, fmt.Sprintf("Missing exact %s mapping; current evidence already confirms a writable trusted input or current-credential definition-edit path on the source side.", spec.Label)
	}
	if targetResolution == "named match" {
		identityName := "the Azure identity tied to this pipeline"
		if strings.TrimSpace(executionIdentityName) != "" {
			identityName = executionIdentityName
		}
		text := fmt.Sprintf("Current evidence shows you can poison this trusted input so it runs as Azure identity '%s' against the exact %s target, but not a separate direct sign-in as Azure identity '%s'.", identityName, spec.Label, identityName)
		return text, fmt.Sprintf("Current evidence names the exact %s target, but does not confirm a writable trusted input or current-credential definition-edit path on the source side.", spec.Label)
	}
	if pipeline.CurrentOperatorCanQueue != nil && *pipeline.CurrentOperatorCanQueue {
		text := fmt.Sprintf("This row proves current-credential run-path control, but not a writable source or the exact %s target.", spec.Label)
		return text, fmt.Sprintf("Missing exact %s mapping and source-side poisoning proof; current evidence does not confirm writable trusted input or current-credential definition-edit path.", spec.Label)
	}
	text := fmt.Sprintf("This row narrows the likely %s targets. Current evidence does not show that current credentials can run this path.", spec.Label)
	return text, fmt.Sprintf("Missing exact %s mapping and source-side poisoning proof; current evidence does not confirm writable trusted input or current-credential definition-edit path.", spec.Label)
}

func devopsNextReview(spec deploymentTargetSpec, pipeline models.DevopsPipelineAsset, targetNames []string, actionability string, permissionOK bool, supporting []models.ArmDeploymentSummary) string {
	if actionability == "currently actionable" {
		text := "Current credentials can already poison a trusted input"
		if len(targetNames) == 1 {
			text += fmt.Sprintf("; HO-Azure already named the exact %s target %s", spec.Label, strings.Join(targetNames, ", "))
		} else if len(targetNames) > 1 {
			text += fmt.Sprintf("; HO-Azure already narrowed the visible %s candidates to %s", spec.Label, strings.Join(targetNames, ", "))
		}
		if len(supporting) > 0 {
			text += "; supporting ARM deployment history includes " + deploymentNames(supporting)
		}
		return text + "."
	}
	if actionability == "conditionally actionable" {
		return fmt.Sprintf("Current credentials can already start this path, but current evidence does not show a writable trusted input; Current scope shows %s as readable, not writable; HO-Azure already narrowed the visible %s candidates to %s; current evidence does not identify which target this path changes.", deploymentTrustedInputLabel(pipeline.PrimaryTrustedInputRef), spec.Label, strings.Join(targetNames, ", "))
	}
	if len(supporting) > 0 && spec.Command == "app-services" {
		return fmt.Sprintf("Current evidence does not show current-credential control of the backing identity or Azure-linked connection; Current scope only shows that %s is referenced here; HO-Azure already narrowed the visible %s candidates to %s; current evidence does not identify which target this path changes; supporting ARM deployment history includes %s.", deploymentTrustedInputLabel(pipeline.PrimaryTrustedInputRef), spec.Label, strings.Join(targetNames, ", "), deploymentNames(supporting))
	}
	return fmt.Sprintf("Current evidence does not show current-credential control of the backing identity or Azure-linked connection; Current scope only shows the external reference %s; HO-Azure already narrowed the visible %s candidates to %s; current evidence does not identify which target this path changes.", deploymentTrustedInputLabel(pipeline.PrimaryTrustedInputRef), spec.Label, strings.Join(targetNames, ", "))
}

func devopsWhyCare(
	pipeline models.DevopsPipelineAsset,
	permission models.PermissionRow,
	permissionOK bool,
	executionIdentityName string,
	trustsBySource map[string][]models.RoleTrustSummary,
	trustsByTarget map[string][]models.RoleTrustSummary,
	keyVaultsByName map[string]models.KeyVaultAsset,
	selectedTargets []deploymentTarget,
	spec deploymentTargetSpec,
	familyName string,
	targetResolution string,
) string {
	ref := deploymentTrustedInputLabel(pipeline.PrimaryTrustedInputRef)
	executionIdentity := "the Azure identity tied to this pipeline"
	if strings.TrimSpace(executionIdentityName) != "" {
		executionIdentity = "Azure identity '" + executionIdentityName + "'"
	}
	intro := fmt.Sprintf("This path trusts %s.", ref)
	if pipeline.CurrentOperatorCanContributeSource != nil && *pipeline.CurrentOperatorCanContributeSource {
		intro += " Current credentials can already modify that trusted input."
		intro += " The resulting run would use " + executionIdentity + " when it makes changes in Azure."
	} else {
		intro += " If that trusted input is changed upstream, the resulting run would use " + executionIdentity + " when it makes changes in Azure."
	}
	intro += " HO-Azure already ties this path to " + deploymentConsequenceNarrative(pipeline.ConsequenceTypes) + "."
	if extra := devopsSupportNarrative(pipeline); extra != "" {
		intro += " Additional visible deployment support around this path includes " + extra + "."
	}
	if len(pipeline.KeyVaultNames) > 0 {
		for _, name := range pipeline.KeyVaultNames {
			if _, ok := keyVaultsByName[strings.ToLower(name)]; !ok {
				intro += fmt.Sprintf(" HO-Azure has not matched named Key Vault support (%s) to visible vault inventory.", name)
				break
			}
		}
	}
	if permissionOK {
		intro += fmt.Sprintf(" This pipeline runs as Azure identity '%s', which already has %s across %s.", permission.DisplayName, firstRoleName(permission), deploymentScopePhrase(permission))
	}
	if trustText := devopsTrustNarrative(pipeline, permission, permissionOK, executionIdentityName, trustsBySource, trustsByTarget); trustText != "" {
		intro += " " + trustText
	}
	if targetNarrative := deploymentTargetNarrative(selectedTargets, spec, familyName, targetResolution == "named match"); targetNarrative != "" {
		intro += " " + targetNarrative
	}
	return intro
}

func devopsEvidenceCommands(familyName string, pipeline models.DevopsPipelineAsset, permissionOK bool, supporting []models.ArmDeploymentSummary) []string {
	commands := []string{"devops", "permissions"}
	if familyName != "arm-deployments" || len(pipeline.AzureServiceConnectionClientIDs) > 0 || len(pipeline.AzureServiceConnectionPrincipalIDs) > 0 {
		commands = append(commands, "role-trusts")
	}
	if containsString(pipeline.SecretSupportTypes, "keyvault-backed-inputs") || len(pipeline.KeyVaultNames) > 0 {
		commands = append(commands, "keyvault")
	}
	commands = append(commands, deploymentTargetSpecs[familyName].Command)
	if len(supporting) > 0 && familyName != "arm-deployments" {
		commands = append(commands, "arm-deployments")
	}
	return dedupeStrings(commands)
}

func devopsJoinedSurfaceTypes(pipeline models.DevopsPipelineAsset, permissionOK bool, familyName string, supporting []models.ArmDeploymentSummary) []string {
	types := []string{"devops"}
	if len(pipeline.AzureServiceConnectionNames) > 0 {
		types = append(types, "azure-service-connection")
	}
	if pipeline.RepositoryName != "" {
		types = append(types, "repo-backed-definition")
	}
	for _, trigger := range pipeline.TriggerTypes {
		switch strings.ToLower(trigger) {
		case "continuousintegration":
			types = append(types, "auto-trigger")
		case "pullrequest":
			types = append(types, "pull-request-trigger")
		case "schedule":
			types = append(types, "scheduled-trigger")
		}
	}
	if permissionOK {
		types = append(types, "permission-summary")
	}
	if familyName != "arm-deployments" {
		types = append(types, "target-family-clue")
	}
	if len(supporting) > 0 && familyName != "arm-deployments" {
		types = append(types, "provider-family-match")
	}
	return dedupeStrings(types)
}

func automationPosture(account models.AutomationAccountAsset) string {
	hasExecution := intPtrValue(account.PublishedRunbookCount) > 0 || intPtrValue(account.ScheduleCount) > 0 || intPtrValue(account.JobScheduleCount) > 0 || intPtrValue(account.WebhookCount) > 0 || intPtrValue(account.HybridWorkerGroupCount) > 0
	hasIdentity := account.IdentityType != nil && strings.TrimSpace(*account.IdentityType) != ""
	hasSecretSupport := len(account.SecretSupportTypes) > 0
	if hasIdentity && hasExecution {
		return "execution-hub"
	}
	if hasSecretSupport {
		return "support-only"
	}
	return ""
}

func automationTargetSelection(account models.AutomationAccountAsset, targetsByFamily map[string][]deploymentTarget, supporting map[string][]models.ArmDeploymentSummary) (string, []deploymentTarget, string) {
	bestFamily := ""
	bestTargets := []deploymentTarget{}
	bestSortKey := []int(nil)
	bestConfirmation := ""

	for familyIndex, familyName := range automationTargetFamilies {
		exactTargets, narrowedTargets, confirmationBasis := automationTargetMatches(account, targetsByFamily[familyName], supporting[familyName])
		selectedTargets := exactTargets
		if len(selectedTargets) == 0 {
			selectedTargets = narrowedTargets
		}
		if len(selectedTargets) == 0 {
			continue
		}

		sortKey := []int{1, len(selectedTargets), 1, familyIndex}
		if len(exactTargets) > 0 {
			sortKey[0] = 0
		}
		if len(supporting[familyName]) > 0 {
			sortKey[2] = 0
		}
		if bestSortKey != nil && !lessIntTuple(sortKey, bestSortKey) {
			continue
		}

		bestSortKey = sortKey
		bestFamily = familyName
		bestTargets = selectedTargets
		bestConfirmation = confirmationBasis
	}

	if bestFamily != "" {
		return bestFamily, bestTargets, bestConfirmation
	}
	if account.MissingTargetMapping {
		return "arm-deployments", nil, ""
	}
	return "", nil, ""
}

func automationTargetMatches(account models.AutomationAccountAsset, candidates []deploymentTarget, supporting []models.ArmDeploymentSummary) ([]deploymentTarget, []deploymentTarget, string) {
	evidenceGroups := automationTargetEvidenceGroups(account)
	runbookNames := automationRunbookNames(account)
	if len(evidenceGroups) == 0 && len(runbookNames) == 0 {
		return nil, nil, ""
	}

	bestNameOnlyTargets := []deploymentTarget{}
	bestNameOnlyRank := 99
	for rank, group := range evidenceGroups {
		matchedTargets := automationExactNameMatches(candidates, group)
		if len(matchedTargets) == 0 {
			continue
		}
		if len(supporting) > 0 && len(matchedTargets) == 1 {
			return matchedTargets, matchedTargets, "same-workload-corroborated"
		}
		if len(bestNameOnlyTargets) == 0 || len(matchedTargets) < len(bestNameOnlyTargets) || (len(matchedTargets) == len(bestNameOnlyTargets) && rank < bestNameOnlyRank) {
			bestNameOnlyTargets = matchedTargets
			bestNameOnlyRank = rank
		}
	}
	if len(bestNameOnlyTargets) > 0 {
		return nil, bestNameOnlyTargets, "name-only-inference"
	}

	overlapTokens := automationOverlapSignalTokens(account, runbookNames)
	if len(overlapTokens) > 0 {
		narrowedTargets := make([]deploymentTarget, 0)
		for _, candidate := range candidates {
			if len(intersectStrings(automationNameTokens(candidate.Name), overlapTokens)) > 0 {
				narrowedTargets = append(narrowedTargets, candidate)
			}
		}
		if len(narrowedTargets) > 0 && len(supporting) > 0 {
			return nil, dedupeDeploymentTargets(narrowedTargets), "same-workload-corroborated"
		}
	}

	return nil, nil, ""
}

func automationTargetEvidenceGroups(account models.AutomationAccountAsset) [][]string {
	groups := make([][]string, 0)
	if names := automationActiveTriggerNames(account); len(names) > 0 {
		groups = append(groups, names)
	}
	if primary := stringPtrValue(account.PrimaryRunbookName); primary != "" {
		groups = append(groups, []string{primary})
	}
	if names := automationActiveModeRunbookNames(account); len(names) > 0 {
		groups = append(groups, names)
	}
	if names := automationRunbookNames(account); len(names) > 0 {
		groups = append(groups, names)
	}

	seen := map[string]struct{}{}
	out := make([][]string, 0, len(groups))
	for _, group := range groups {
		normalized := make([]string, 0, len(group))
		for _, name := range group {
			key := normalizeTargetName(name)
			if key != "" {
				normalized = append(normalized, key)
			}
		}
		sort.Strings(normalized)
		groupKey := strings.Join(normalized, "|")
		if groupKey == "" {
			continue
		}
		if _, ok := seen[groupKey]; ok {
			continue
		}
		seen[groupKey] = struct{}{}
		out = append(out, group)
	}
	return out
}

func automationExactNameMatches(candidates []deploymentTarget, names []string) []deploymentTarget {
	normalizedNames := map[string]struct{}{}
	for _, name := range names {
		key := normalizeTargetName(name)
		if key != "" {
			normalizedNames[key] = struct{}{}
		}
	}
	matched := make([]deploymentTarget, 0)
	for _, candidate := range candidates {
		if _, ok := normalizedNames[normalizeTargetName(candidate.Name)]; ok {
			matched = append(matched, candidate)
		}
	}
	return dedupeDeploymentTargets(matched)
}

func automationActiveModeRunbookNames(account models.AutomationAccountAsset) []string {
	names := []string{}
	if primary := stringPtrValue(account.PrimaryRunbookName); primary != "" {
		names = append(names, primary)
	}
	switch strings.ToLower(stringPtrValue(account.PrimaryStartMode)) {
	case "webhook":
		names = appendUniqueStrings(names, account.WebhookRunbookNames...)
	case "schedule":
		names = appendUniqueStrings(names, account.ScheduleRunbookNames...)
	case "manual-only", "published-runbook":
		names = appendUniqueStrings(names, account.PublishedRunbookNames...)
	}
	return names
}

func automationActiveTriggerNames(account models.AutomationAccountAsset) []string {
	mode := strings.ToLower(stringPtrValue(account.PrimaryStartMode))
	names := []string{}
	for _, value := range account.TriggerJoinIDs {
		text := strings.TrimSpace(value)
		lowered := strings.ToLower(text)
		switch mode {
		case "webhook":
			if strings.HasPrefix(lowered, "automation-webhook:") {
				names = appendUniqueStrings(names, strings.TrimSpace(text[len("automation-webhook:"):]))
			}
		case "schedule":
			if strings.HasPrefix(lowered, "automation-job-schedule:") {
				names = appendUniqueStrings(names, strings.TrimSpace(text[len("automation-job-schedule:"):]))
			}
		}
	}
	return names
}

func automationOverlapSignalTokens(account models.AutomationAccountAsset, runbookNames []string) []string {
	if tokens := automationRunbookTokens(automationActiveTriggerNames(account)); len(tokens) > 0 {
		return tokens
	}
	if tokens := automationRunbookTokens(automationActiveModeRunbookNames(account)); len(tokens) > 0 {
		return tokens
	}
	if stringPtrValue(account.PrimaryRunbookName) != "" || stringPtrValue(account.PrimaryStartMode) != "" {
		return nil
	}
	return automationRunbookTokens(runbookNames)
}

func automationRunbookNames(account models.AutomationAccountAsset) []string {
	names := []string{}
	if primary := stringPtrValue(account.PrimaryRunbookName); primary != "" {
		names = append(names, primary)
	}
	names = appendUniqueStrings(names, account.PublishedRunbookNames...)
	names = appendUniqueStrings(names, account.ScheduleRunbookNames...)
	names = appendUniqueStrings(names, account.WebhookRunbookNames...)
	return names
}

func automationRunbookTokens(names []string) []string {
	tokens := []string{}
	for _, name := range names {
		tokens = append(tokens, automationNameTokens(name)...)
	}
	return dedupeStrings(tokens)
}

func automationActionability(account models.AutomationAccountAsset, posture string, currentIdentityAssignments []models.RoleAssignment) (string, string, string, string) {
	if posture == "support-only" {
		return "support-only", "support-only", "low", "bookmark"
	}
	if automationCurrentOperatorCanEdit(currentIdentityAssignments) {
		return "currently actionable", "currently actionable", "high", "pivot-now"
	}
	return "conditionally actionable", "conditionally actionable", "medium", "review-soon"
}

func automationCurrentOperatorCanEdit(assignments []models.RoleAssignment) bool {
	for _, assignment := range assignments {
		name := strings.ToLower(strings.TrimSpace(assignment.RoleName))
		if name == "owner" || name == "contributor" || name == "automation contributor" || name == "automation operator" {
			return true
		}
	}
	return false
}

func automationInsertionPoint(account models.AutomationAccountAsset, currentIdentityAssignments []models.RoleAssignment) (string, string) {
	startMode := strings.ToLower(stringPtrValue(account.PrimaryStartMode))
	runbook := stringPtrValue(account.PrimaryRunbookName)
	parts := []string{}
	switch startMode {
	case "webhook":
		parts = append(parts, fmt.Sprintf("webhook path can start runbook %s under automation identity %s", runbook, stringPtrValue(account.IdentityType)))
	case "schedule", "job-schedule":
		parts = append(parts, fmt.Sprintf("schedule path can start runbook %s", runbook))
	default:
		parts = append(parts, "published runbook path can start automation execution")
	}
	if intPtrValue(account.HybridWorkerGroupCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d Hybrid Worker reach point(s)", intPtrValue(account.HybridWorkerGroupCount)))
	}
	if automationCurrentOperatorCanEdit(currentIdentityAssignments) {
		scopePhrase := "subscription scope"
		for _, assignment := range currentIdentityAssignments {
			if strings.EqualFold(assignment.RoleName, "Owner") || strings.EqualFold(assignment.RoleName, "Contributor") {
				scopePhrase = automationScopeLabel(assignment.ScopeID)
				break
			}
		}
		parts = append(parts, fmt.Sprintf("current role assignment Owner at %s can edit runbook %s or its %s-backed execution boundary", scopePhrase, runbook, startMode))
	}
	if startMode == "webhook" {
		parts = append(parts, "current scope does not expose the webhook URI value")
	}
	text := strings.Join(parts, "; ") + "."
	return text, strings.Join(parts, ";\n") + "."
}

func automationConfidenceBoundary(spec deploymentTargetSpec, posture string, permissionOK bool, targetResolution string) (string, string) {
	if posture == "support-only" {
		text := "This row proves source-side control, but HO-Azure has not yet mapped the downstream Azure footprint beyond ARM deployment evidence."
		return text, "Missing exact target mapping and a separate execution foothold; current evidence only shows secret-backed support around a live Azure change path."
	}
	text := fmt.Sprintf("This row proves source-side control, but not the exact %s target.", spec.Label)
	return text, fmt.Sprintf("Missing exact %s mapping; current RBAC evidence already shows edit-capable automation control, but HO-Azure has not yet mapped which Azure target that control reaches.", spec.Label)
}

func automationNextReview(spec deploymentTargetSpec, posture string, targetNames []string, permissionOK bool, supporting []models.ArmDeploymentSummary) string {
	if posture == "support-only" {
		return fmt.Sprintf("Current evidence does not yet show a current-credential start or control path for this secret-backed support; Current evidence does not yet map what runbook %s changes on the Azure side.", "Lab-Maintenance")
	}
	text := fmt.Sprintf("Current RBAC evidence already shows edit-capable automation control here; HO-Azure already narrowed the visible %s candidates to %s; current evidence does not identify which target this path changes", spec.Label, strings.Join(targetNames, ", "))
	if len(supporting) > 0 {
		text += "; supporting ARM deployment history includes " + deploymentNames(supporting)
	}
	return text + "."
}

func automationWhyCare(account models.AutomationAccountAsset, posture string, permission models.PermissionRow, permissionOK bool, trustsBySource map[string][]models.RoleTrustSummary, selectedTargets []deploymentTarget, spec deploymentTargetSpec, targetResolution string) string {
	if posture == "support-only" {
		text := fmt.Sprintf("Automation account '%s' does not on its own show a current-credential Azure change path, but it concentrates %s around reusable automation.", account.Name, deploymentSupportPhrase(automationSupportParts(account)))
		if clause := automationCurrentOperatorClause(account); clause != "" {
			text += " " + clause + "."
		}
		return text
	}

	parts := []string{}
	startMode := strings.ToLower(stringPtrValue(account.PrimaryStartMode))
	runbook := stringPtrValue(account.PrimaryRunbookName)
	if startMode == "webhook" {
		parts = append(parts, fmt.Sprintf("webhook path can start runbook %s under automation identity %s", runbook, stringPtrValue(account.IdentityType)))
	} else {
		parts = append(parts, fmt.Sprintf("%s path can start runbook %s", startMode, runbook))
	}
	parts = append(parts, fmt.Sprintf("%d published runbook(s)", intPtrValue(account.PublishedRunbookCount)))
	if intPtrValue(account.ScheduleCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d schedule-backed run path(s)", intPtrValue(account.ScheduleCount)))
	}
	if intPtrValue(account.HybridWorkerGroupCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d Hybrid Worker reach point(s)", intPtrValue(account.HybridWorkerGroupCount)))
	}
	text := fmt.Sprintf("Automation account '%s' combines %s.", account.Name, strings.Join(parts, ", "))
	text += " HO-Azure already ties this path to configuration change reach, recurring Azure execution, and secret-backed deployment support."
	if extra := automationSupportNarrative(account); extra != "" {
		text += " Additional visible deployment support around this account includes " + extra + "."
	}
	if clause := automationCurrentOperatorClause(account); clause != "" {
		text += " " + clause + "."
	}
	if permissionOK {
		text += fmt.Sprintf(" Azure identity '%s' already has %s across %s.", permission.DisplayName, firstRoleName(permission), deploymentScopePhrase(permission))
	}
	if account.PrincipalID != nil {
		if trusts := trustsBySource[*account.PrincipalID]; len(trusts) > 0 {
			trust := trusts[0]
			targetName := stringPtrValue(trust.TargetName)
			if targetName != "" {
				text += fmt.Sprintf(" HO-Azure also sees a separate identity-control path into Azure identity '%s' through service principal '%s'.", targetName, account.Name)
			}
		}
	}
	if targetNarrative := deploymentTargetNarrative(selectedTargets, spec, spec.Command, targetResolution == "named match"); targetNarrative != "" {
		text += " " + targetNarrative
	}
	return text
}

func automationEvidenceCommands(targetFamily string, permissionOK bool, supporting []models.ArmDeploymentSummary) []string {
	commands := []string{"automation", "permissions", "rbac", deploymentTargetSpecs[targetFamily].Command}
	if permissionOK {
		commands = append(commands, "role-trusts")
	}
	if len(supporting) > 0 && targetFamily != "arm-deployments" {
		commands = append(commands, "arm-deployments")
	}
	return dedupeStrings(commands)
}

func automationJoinedSurfaceTypes(account models.AutomationAccountAsset, posture string, targetFamily string) []string {
	types := []string{"automation"}
	if account.IdentityType != nil && *account.IdentityType != "" {
		types = append(types, "managed-identity")
	}
	if intPtrValue(account.PublishedRunbookCount) > 0 {
		types = append(types, "published-runbooks")
	}
	if intPtrValue(account.ScheduleCount) > 0 || intPtrValue(account.JobScheduleCount) > 0 {
		types = append(types, "scheduled-start")
	}
	if intPtrValue(account.WebhookCount) > 0 {
		types = append(types, "webhook-start")
	}
	if intPtrValue(account.HybridWorkerGroupCount) > 0 {
		types = append(types, "hybrid-worker-reach")
	}
	if targetFamily != "arm-deployments" {
		types = append(types, "provider-family-match")
	}
	return dedupeStrings(types)
}

func deploymentConsequenceNarrative(consequenceTypes []string) string {
	order := []string{
		"modify-infra",
		"infrastructure-deployment",
		"redeploy-workload",
		"reintroduce-config",
		"run-recurring-execution",
		"consume-secret-backed-deployment-material",
	}
	labels := map[string]string{
		"modify-infra":                              "infrastructure deployment reach",
		"infrastructure-deployment":                 "infrastructure deployment reach",
		"redeploy-workload":                         "workload deployment reach",
		"reintroduce-config":                        "configuration change reach",
		"run-recurring-execution":                   "recurring Azure execution",
		"consume-secret-backed-deployment-material": "secret-backed deployment support",
	}
	seen := map[string]struct{}{}
	parts := []string{}
	for _, value := range order {
		if !containsString(consequenceTypes, value) {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		parts = append(parts, labels[value])
	}
	for _, value := range consequenceTypes {
		switch value {
		case "modify-infra", "infrastructure-deployment", "redeploy-workload", "reintroduce-config", "run-recurring-execution", "consume-secret-backed-deployment-material":
			continue
		default:
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				parts = append(parts, trimmed)
			}
		}
	}
	return naturalJoin(dedupeStrings(parts))
}

func devopsSupportNarrative(pipeline models.DevopsPipelineAsset) string {
	return deploymentSupportPhrase(devopsSupportParts(pipeline))
}

func devopsSupportParts(pipeline models.DevopsPipelineAsset) []string {
	parts := []string{}
	for _, value := range pipeline.SecretSupportTypes {
		switch value {
		case "secret-variables":
			parts = append(parts, "secret variables")
		case "signing-keys":
			parts = append(parts, "signing keys")
		case "registry-creds":
			parts = append(parts, "registry credentials")
		case "deployment-creds":
			parts = append(parts, "deployment credentials")
		case "keyvault-backed-inputs":
			parts = append(parts, "Key Vault-backed inputs")
		case "publish-profiles":
			parts = append(parts, "publish profiles")
		}
	}
	return dedupeStrings(parts)
}

func automationSupportNarrative(account models.AutomationAccountAsset) string {
	return deploymentSupportPhrase(automationSupportParts(account))
}

func automationSupportParts(account models.AutomationAccountAsset) []string {
	parts := []string{}
	for _, value := range account.SecretSupportTypes {
		switch value {
		case "credentials":
			parts = append(parts, "credentials")
		case "certificates":
			parts = append(parts, "certificates")
		case "connections":
			parts = append(parts, "connections")
		case "encrypted-variables":
			parts = append(parts, "encrypted variables")
		}
	}
	return dedupeStrings(parts)
}

func firstTrustNarrative(permission models.PermissionRow, trustsByTarget map[string][]models.RoleTrustSummary) string {
	if permission.PrincipalID == "" {
		return ""
	}
	trusts := trustsByTarget[permission.PrincipalID]
	if len(trusts) == 0 {
		return ""
	}
	trust := trusts[0]
	sourceName := stringPtrValue(trust.SourceName)
	switch trust.TrustType {
	case "app-owner", "federated-credential":
		if sourceName != "" {
			return fmt.Sprintf("HO-Azure also sees a separate app trust path into that same Azure identity through app '%s'.", sourceName)
		}
	case "service-principal-owner":
		if sourceName != "" {
			return fmt.Sprintf("HO-Azure also sees a separate identity-control path into Azure identity '%s' through service principal '%s'.", stringPtrValue(trust.TargetName), sourceName)
		}
	}
	return ""
}

func devopsIdentityRefs(pipeline models.DevopsPipelineAsset) []string {
	refs := append([]string{}, pipeline.AzureServiceConnectionPrincipalIDs...)
	refs = append(refs, pipeline.AzureServiceConnectionClientIDs...)
	return dedupeStrings(refs)
}

func devopsExecutionIdentityName(
	pipeline models.DevopsPipelineAsset,
	permission models.PermissionRow,
	permissionOK bool,
	trustsBySource map[string][]models.RoleTrustSummary,
	trustsByTarget map[string][]models.RoleTrustSummary,
) string {
	if permissionOK && strings.TrimSpace(permission.DisplayName) != "" {
		return permission.DisplayName
	}
	for _, ref := range devopsIdentityRefs(pipeline) {
		for _, trust := range trustsByTarget[ref] {
			if name := stringPtrValue(trust.TargetName); name != "" {
				return name
			}
		}
		for _, trust := range trustsBySource[ref] {
			if name := stringPtrValue(trust.SourceName); name != "" {
				return name
			}
		}
	}
	return ""
}

func devopsTrustNarrative(
	pipeline models.DevopsPipelineAsset,
	permission models.PermissionRow,
	permissionOK bool,
	executionIdentityName string,
	trustsBySource map[string][]models.RoleTrustSummary,
	trustsByTarget map[string][]models.RoleTrustSummary,
) string {
	if permissionOK {
		if text := firstTrustNarrative(permission, trustsByTarget); text != "" {
			return text
		}
	}
	for _, ref := range devopsIdentityRefs(pipeline) {
		for _, trust := range trustsByTarget[ref] {
			if text := devopsTrustNarrativeFromTrust(trust, executionIdentityName); text != "" {
				return text
			}
		}
	}
	for _, ref := range devopsIdentityRefs(pipeline) {
		for _, trust := range trustsBySource[ref] {
			if text := devopsTrustNarrativeFromTrust(trust, executionIdentityName); text != "" {
				return text
			}
		}
	}
	return ""
}

func devopsTrustNarrativeFromTrust(trust models.RoleTrustSummary, executionIdentityName string) string {
	targetName := stringPtrValue(trust.TargetName)
	sourceName := stringPtrValue(trust.SourceName)
	targetIdentityText := "Azure identity '" + targetName + "'"
	if targetName != "" && strings.EqualFold(targetName, executionIdentityName) {
		targetIdentityText = "that same Azure identity"
	}
	switch trust.TrustType {
	case "service-principal-owner":
		if sourceName != "" && targetName != "" {
			return fmt.Sprintf("HO-Azure also sees a separate identity-control path into %s through service principal '%s'.", targetIdentityText, sourceName)
		}
	case "app-owner":
		if sourceName != "" && targetName != "" {
			return fmt.Sprintf("HO-Azure also sees a separate app control path through '%s' into app '%s'.", sourceName, targetName)
		}
	case "federated-credential":
		if sourceName != "" && targetName != "" {
			return fmt.Sprintf("HO-Azure also sees a separate app trust path into %s through app '%s'.", targetIdentityText, sourceName)
		}
	}
	return ""
}

func deploymentTargetNarrative(targets []deploymentTarget, spec deploymentTargetSpec, familyName string, exactMatch bool) string {
	if len(targets) == 0 {
		return ""
	}
	if exactMatch && len(targets) == 1 && familyName != "arm-deployments" {
		return "Visible target-side record: " + targets[0].Summary
	}
	switch familyName {
	case "app-services":
		publicCount := 0
		identityCount := 0
		weakTLSCount := 0
		for _, target := range targets {
			lower := strings.ToLower(target.Summary)
			if strings.Contains(lower, "public network access enabled") {
				publicCount++
			}
			if strings.Contains(lower, "managed identity") {
				identityCount++
			}
			if strings.Contains(lower, "https-only disabled") || strings.Contains(lower, "tls 1.0") {
				weakTLSCount++
			}
		}
		text := fmt.Sprintf("Visible %s evidence keeps %d candidate(s) in play", spec.Label, len(targets))
		clauses := []string{}
		if publicCount > 0 {
			clauses = append(clauses, fmt.Sprintf("%d keep public network access enabled", publicCount))
		}
		if identityCount > 0 {
			clauses = append(clauses, fmt.Sprintf("%d carry managed identity", identityCount))
		}
		if len(clauses) > 0 {
			text += "; " + strings.Join(clauses, " and ") + "."
		} else {
			text += "."
		}
		if weakTLSCount > 0 {
			text += fmt.Sprintf(" %d still show weaker HTTPS or TLS posture.", weakTLSCount)
		}
		return text
	case "aks":
		privateCount := 0
		wiCount := 0
		spCount := 0
		for _, target := range targets {
			lower := strings.ToLower(target.Summary)
			if strings.Contains(lower, "private api endpoint") {
				privateCount++
			}
			if strings.Contains(lower, "workload identity enabled") {
				wiCount++
			}
			if strings.Contains(lower, "service principal") {
				spCount++
			}
		}
		clauses := []string{}
		if privateCount > 0 {
			clauses = append(clauses, fmt.Sprintf("%d keep private API endpoints", privateCount))
		}
		if wiCount > 0 {
			clauses = append(clauses, fmt.Sprintf("%d keep workload identity enabled", wiCount))
		}
		if spCount > 0 {
			clauses = append(clauses, fmt.Sprintf("%d still use service principal auth", spCount))
		}
		return fmt.Sprintf("Visible AKS evidence keeps %d candidate(s) in play; %s.", len(targets), naturalJoin(clauses))
	case "arm-deployments":
		rgCount := 0
		subCount := 0
		providers := []string{}
		for _, target := range targets {
			if strings.Contains(target.ID, "/resourceGroups/") {
				rgCount++
			} else {
				subCount++
			}
			providers = append(providers, target.Providers...)
		}
		providers = dedupeStrings(providers)
		sort.Strings(providers)
		return fmt.Sprintf("Visible ARM deployment history keeps %d resource group deployment(s), %d subscription deployment(s) in play across providers %s.", rgCount, subCount, strings.Join(providers, ", "))
	default:
		if len(targets) == 1 {
			return "Visible target-side record: " + targets[0].Summary
		}
	}
	return ""
}

func deploymentSupportingSentence(supporting []models.ArmDeploymentSummary, spec deploymentTargetSpec) string {
	return fmt.Sprintf("Supporting ARM deployment history for the same target family includes %s, which supports the likely Azure footprint without proving the exact target.", deploymentNames(supporting))
}

func deploymentNames(items []models.ArmDeploymentSummary) string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if item.Name != "" {
			names = append(names, item.Name)
		}
	}
	return strings.Join(names, ", ")
}

func deploymentSupportingIDs(items []models.ArmDeploymentSummary) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item.ID != "" {
			ids = append(ids, item.ID)
		}
	}
	return ids
}

func deploymentPermissionName(permissionOK bool, pipeline models.DevopsPipelineAsset, name string) string {
	if permissionOK {
		return name
	}
	return "the Azure identity tied to this pipeline"
}

func firstRoleName(permission models.PermissionRow) string {
	if len(permission.HighImpactRoles) > 0 {
		return permission.HighImpactRoles[0]
	}
	if len(permission.AllRoleNames) > 0 {
		return permission.AllRoleNames[0]
	}
	return "Azure control"
}

func deploymentScopePhrase(permission models.PermissionRow) string {
	if len(permission.ScopeIDs) >= 1 {
		if len(permission.ScopeIDs) == 1 {
			return "subscription-wide scope"
		}
		return fmt.Sprintf("%d visible scopes", len(permission.ScopeIDs))
	}
	return "visible scope"
}

func deploymentShortScope(scopeID string) string {
	if strings.Contains(scopeID, "/subscriptions/") && !strings.Contains(scopeID, "/resourceGroups/") {
		return "subscription-wide scope"
	}
	if index := strings.Index(scopeID, "/resourceGroups/"); index >= 0 {
		return "subscription scope"
	}
	return "visible scope"
}

func automationScopeLabel(scopeID string) string {
	if strings.Contains(scopeID, "/subscriptions/") && !strings.Contains(scopeID, "/resourceGroups/") {
		return "subscription scope"
	}
	if strings.Contains(scopeID, "/resourceGroups/") {
		return "resource group " + armScopeName(scopeID)
	}
	return "a parent scope of this automation account"
}

func armScopeName(scopeID string) string {
	parts := strings.Split(scopeID, "/")
	for index := 0; index < len(parts)-1; index++ {
		if strings.EqualFold(parts[index], "resourceGroups") {
			return parts[index+1]
		}
	}
	return "unknown"
}

func deploymentDisplaySource(name string) string {
	if len(name) > 18 && strings.Contains(name, "-") {
		return strings.ReplaceAll(name, "-", "-\n")
	}
	return name
}

func deploymentAutomationClueType(posture string) string {
	if posture == "support-only" {
		return "secret-escalation-support"
	}
	return "execution-hub"
}

func deploymentAutomationVisiblePath(posture string, spec deploymentTargetSpec) string {
	if posture == "support-only" {
		return "Secret-backed automation support -> likely " + spec.Label
	}
	return "Managed-identity execution hub -> likely " + spec.Label
}

func automationSignalTokens(account models.AutomationAccountAsset) []string {
	all := []string{account.Name}
	all = append(all, account.PublishedRunbookNames...)
	all = append(all, account.ScheduleRunbookNames...)
	all = append(all, account.WebhookRunbookNames...)
	all = append(all, account.TriggerJoinIDs...)
	tokens := []string{}
	for _, value := range all {
		tokens = append(tokens, automationNameTokens(value)...)
	}
	return dedupeStrings(tokens)
}

func automationNameTokens(value string) []string {
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	})
	tokens := []string{}
	for _, field := range fields {
		if len(field) < 2 {
			continue
		}
		if _, stopword := automationStopwords[field]; stopword {
			continue
		}
		tokens = append(tokens, field)
	}
	return tokens
}

func overlapCount(left []string, right []string) int {
	set := map[string]struct{}{}
	for _, value := range left {
		set[value] = struct{}{}
	}
	count := 0
	for _, value := range right {
		if _, ok := set[value]; ok {
			count++
		}
	}
	return count
}

func intersectStrings(left []string, right []string) []string {
	set := map[string]struct{}{}
	for _, value := range right {
		set[value] = struct{}{}
	}
	out := []string{}
	for _, value := range left {
		if _, ok := set[value]; ok {
			out = append(out, value)
		}
	}
	return dedupeStrings(out)
}

func appendUniqueStrings(base []string, values ...string) []string {
	seen := map[string]struct{}{}
	for _, value := range base {
		seen[value] = struct{}{}
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		base = append(base, value)
	}
	return base
}

func normalizeTargetName(value string) string {
	return strings.NewReplacer(" ", "", "_", "", "/", "", "\\", "").Replace(strings.ToLower(strings.TrimSpace(value)))
}

func lessIntTuple(left []int, right []int) bool {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	for index := 0; index < limit; index++ {
		if left[index] == right[index] {
			continue
		}
		return left[index] < right[index]
	}
	return len(left) < len(right)
}

func deploymentSupportPhrase(parts []string) string {
	parts = dedupeStrings(parts)
	return strings.Join(parts, " and ")
}

func automationCurrentOperatorClause(account models.AutomationAccountAsset) string {
	startMode := strings.ToLower(stringPtrValue(account.PrimaryStartMode))
	runbook := stringPtrValue(account.PrimaryRunbookName)
	if runbook == "" {
		return ""
	}
	switch startMode {
	case "webhook":
		return fmt.Sprintf("Current role assignment Owner at subscription scope can edit runbook %s or its webhook-backed execution boundary; current scope does not expose the webhook URI value", runbook)
	case "schedule", "job-schedule":
		return fmt.Sprintf("Current role assignment Owner at subscription scope can edit runbook %s or its schedule-backed execution boundary", runbook)
	default:
		return fmt.Sprintf("Current role assignment Owner at subscription scope can edit published runbook %s", runbook)
	}
}

func dedupeDeploymentTargets(values []deploymentTarget) []deploymentTarget {
	seen := map[string]struct{}{}
	out := make([]deploymentTarget, 0, len(values))
	for _, value := range values {
		key := value.ID + "::" + value.Name
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return out
}

func stringPtrIf(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func naturalJoin(values []string) string {
	values = dedupeStrings(values)
	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	case 2:
		return values[0] + " and " + values[1]
	default:
		return strings.Join(values[:len(values)-1], ", ") + ", and " + values[len(values)-1]
	}
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), needle) {
			return true
		}
	}
	return false
}
