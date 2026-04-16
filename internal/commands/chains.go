package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

const (
	chainsCurrentBehavior = "Family overview and grouped runner. Use `ho-azure chains` or `ho-azure chains help` to list families, then `ho-azure chains <family>` to run an implemented family."
	chainsCommandState    = contracts.StatusImplemented
	credentialPathState   = "extraction-only"
	chainsFanoutLimit     = 4
)

var (
	chainsInputModes            = []string{"live", "artifacts"}
	chainsPreferredArtifactMode = []string{"loot", "json"}
	credentialJoinQualityOrder  = map[string]int{
		"named match":              0,
		"narrowed candidates":      1,
		"tenant-wide candidates":   2,
		"visibility blocked":       3,
		"service hint only":        4,
		"named target not visible": 5,
	}
)

type credentialPathTargetView struct {
	Service         string
	Candidates      []credentialPathTarget
	VisibilityNote  *string
	VisibilityIssue *string
}

type credentialPathTarget struct {
	ID       string
	Name     string
	Location *string
	VaultURI *string
}

type chainsFamilyBuilder func(context.Context, providers.Provider, func() time.Time, Request, contracts.FamilyContract) (models.ChainsOutput, error)

var chainsFamilyBuilders = map[string]chainsFamilyBuilder{
	"credential-path": buildCredentialPathOutput,
	"deployment-path": buildDeploymentPathOutput,
	"escalation-path": buildEscalationPathOutput,
	"compute-control": buildComputeControlOutput,
}

func chainsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		family := strings.TrimSpace(request.ChainFamily)
		if family == "" {
			return buildChainsOverview(now, request, nil), nil
		}

		contract, ok := contracts.Family(family)
		if !ok {
			return nil, fmt.Errorf("unknown chain family %q", family)
		}
		builder, ok := chainsFamilyBuilders[family]
		if !ok {
			return nil, fmt.Errorf("chain family %q is not implemented yet; scaffold contract is in place for migration", family)
		}
		return builder(ctx, provider, now, request, contract)
	}
}

func buildChainsOverview(now func() time.Time, request Request, selectedFamily *string) models.ChainsOverviewOutput {
	families := make([]models.ChainFamilyDescriptor, 0, len(contracts.FamilyNames()))
	for _, name := range contracts.FamilyNames() {
		family, _ := contracts.Family(name)
		sources := make([]models.ChainSourceDescriptor, 0, len(family.SourceCommandMinimum))
		for _, source := range family.SourceCommandMinimum {
			sources = append(sources, models.ChainSourceDescriptor{
				Command:       source.Command,
				MinimumFields: append([]string{}, source.MinimumFields...),
				Rationale:     source.Rationale,
			})
		}
		families = append(families, models.ChainFamilyDescriptor{
			Family:              family.Name,
			State:               family.Status,
			Meaning:             family.Meaning,
			Summary:             family.Summary,
			AllowedClaim:        family.AllowedClaim,
			CurrentGap:          family.CurrentGap,
			BestCurrentExamples: append([]string{}, family.BestCurrentExamples...),
			SourceCommands:      sources,
		})
	}

	return models.ChainsOverviewOutput{
		Metadata:               scopedMetadata(now, request, request.Tenant, request.Subscription, "chains"),
		GroupedCommandName:     "chains",
		CommandState:           chainsCommandState,
		CurrentBehavior:        chainsCurrentBehavior,
		PlannedInputModes:      append([]string{}, chainsInputModes...),
		PreferredArtifactOrder: append([]string{}, chainsPreferredArtifactMode...),
		SelectedFamily:         selectedFamily,
		Families:               families,
		Issues:                 []models.Issue{},
	}
}

func buildCredentialPathOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	family contracts.FamilyContract,
) (models.ChainsOutput, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	envVarsFuture := runGroupedCommandOutput[models.EnvVarsOutput](group, ctx, request, envVarsHandler(provider, now), "env-vars")
	tokenSurfacesFuture := runGroupedCommandOutput[models.TokensCredentialsOutput](group, ctx, request, tokensCredentialsHandler(provider, now), "tokens-credentials")
	databasesFuture := runGroupedCommandOutput[models.DatabasesOutput](group, ctx, request, databasesHandler(provider, now), "databases")
	storageFuture := runGroupedCommandOutput[models.StorageOutput](group, ctx, request, storageHandler(provider, now), "storage")
	keyvaultsFuture := runGroupedCommandOutput[models.KeyVaultOutput](group, ctx, request, keyVaultHandler(provider, now), "keyvault")

	envVars, err := envVarsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	tokenSurfaces, err := tokenSurfacesFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	databases, err := databasesFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	storage, err := storageFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}
	keyvaults, err := keyvaultsFuture.wait()
	if err != nil {
		return models.ChainsOutput{}, err
	}

	targetViews := map[string]credentialPathTargetView{
		"database": buildDatabaseTargetView(databases),
		"storage":  buildStorageTargetView(storage),
		"keyvault": buildKeyVaultTargetView(keyvaults),
	}

	tokenSettingIndex := map[string][]models.TokenCredentialSurfaceSummary{}
	tokenTargetIndex := map[string][]models.TokenCredentialSurfaceSummary{}
	for _, surface := range tokenSurfaces.Surfaces {
		signal := parseOperatorSignal(surface.OperatorSignal)
		if setting := strings.TrimSpace(signal["setting"]); setting != "" {
			key := surface.AssetID + "::" + strings.ToLower(setting)
			tokenSettingIndex[key] = append(tokenSettingIndex[key], surface)
		}
		if target := strings.TrimSpace(signal["target"]); target != "" {
			key := surface.AssetID + "::" + normalizeReferenceTarget(target)
			tokenTargetIndex[key] = append(tokenTargetIndex[key], surface)
		}
	}

	paths := make([]models.ChainPathRecord, 0)
	for _, envVar := range envVars.EnvVars {
		joinedSurfaces := append([]models.TokenCredentialSurfaceSummary{}, tokenSettingIndex[envVar.AssetID+"::"+strings.ToLower(envVar.SettingName)]...)
		if envVar.ValueType == "keyvault-ref" && envVar.ReferenceTarget != nil && *envVar.ReferenceTarget != "" {
			joinedSurfaces = append(joinedSurfaces, tokenTargetIndex[envVar.AssetID+"::"+normalizeReferenceTarget(*envVar.ReferenceTarget)]...)
		}

		if envVar.ValueType == "keyvault-ref" && envVar.ReferenceTarget != nil && *envVar.ReferenceTarget != "" {
			record := buildKeyVaultCredentialRecord(envVar, joinedSurfaces, targetViews["keyvault"])
			if record != nil {
				paths = append(paths, *record)
			}
			continue
		}

		if !credentialLikeEnvVar(envVar, joinedSurfaces) {
			continue
		}

		for _, service := range envVar.TargetServices {
			targetService := string(service)
			if targetService == "" {
				continue
			}
			paths = append(paths, buildCandidateCredentialRecord(envVar, joinedSurfaces, targetViews[targetService], targetService))
		}
	}

	sort.SliceStable(paths, func(i int, j int) bool {
		left := paths[i]
		right := paths[j]
		if left.Priority != right.Priority {
			return prioritySortValue(left.Priority) < prioritySortValue(right.Priority)
		}
		if left.TargetResolution != right.TargetResolution {
			return credentialJoinQualityOrder[left.TargetResolution] < credentialJoinQualityOrder[right.TargetResolution]
		}
		if left.AssetName != right.AssetName {
			return left.AssetName < right.AssetName
		}
		leftSetting := stringPtrValue(left.SettingName)
		rightSetting := stringPtrValue(right.SettingName)
		if leftSetting != rightSetting {
			return leftSetting < rightSetting
		}
		return left.TargetService < right.TargetService
	})

	issues := append([]models.Issue{}, envVars.Issues...)
	issues = append(issues, tokenSurfaces.Issues...)
	issues = append(issues, databases.Issues...)
	issues = append(issues, storage.Issues...)
	issues = append(issues, keyvaults.Issues...)

	return models.ChainsOutput{
		Metadata:                scopedMetadata(now, request, request.Tenant, request.Subscription, "chains"),
		GroupedCommandName:      "chains",
		Family:                  family.Name,
		InputMode:               "live",
		CommandState:            credentialPathState,
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

func runCommandOutput[T any](ctx context.Context, request Request, handler Handler, name string) (T, error) {
	var zero T
	payload, err := handler(ctx, request)
	if err != nil {
		return zero, err
	}
	out, ok := payload.(T)
	if !ok {
		return zero, fmt.Errorf("unexpected payload type for %s: %T", name, payload)
	}
	return out, nil
}

type asyncCommandOutput[T any] struct {
	result chan asyncCommandResult[T]
}

type asyncCommandResult[T any] struct {
	value T
	err   error
}

type commandOutputGroup struct {
	limiter chan struct{}
}

func newCommandOutputGroup(limit int) commandOutputGroup {
	if limit < 1 {
		limit = 1
	}
	return commandOutputGroup{limiter: make(chan struct{}, limit)}
}

func runGroupedCommandOutput[T any](group commandOutputGroup, ctx context.Context, request Request, handler Handler, name string) asyncCommandOutput[T] {
	result := make(chan asyncCommandResult[T], 1)
	go func() {
		group.limiter <- struct{}{}
		defer func() {
			<-group.limiter
		}()
		value, err := runCommandOutput[T](ctx, request, handler, name)
		result <- asyncCommandResult[T]{value: value, err: err}
	}()
	return asyncCommandOutput[T]{result: result}
}

func (future asyncCommandOutput[T]) wait() (T, error) {
	result := <-future.result
	return result.value, result.err
}

func buildDatabaseTargetView(output models.DatabasesOutput) credentialPathTargetView {
	return buildCredentialTargetView(
		"database",
		"database",
		output.Issues,
		buildCredentialTargets(output.DatabaseServers, func(server models.DatabaseServerAsset) credentialPathTarget {
			return credentialPathTarget{
				ID:       server.ID,
				Name:     server.Name,
				Location: server.Location,
			}
		}),
	)
}

func buildStorageTargetView(output models.StorageOutput) credentialPathTargetView {
	return buildCredentialTargetView(
		"storage",
		"storage",
		output.Issues,
		buildCredentialTargets(output.StorageAssets, func(account models.StorageAsset) credentialPathTarget {
			return credentialPathTarget{
				ID:       account.ID,
				Name:     account.Name,
				Location: account.Location,
			}
		}),
	)
}

func buildCredentialTargets[T any](items []T, mapTarget func(T) credentialPathTarget) []credentialPathTarget {
	candidates := make([]credentialPathTarget, 0, len(items))
	for _, item := range items {
		candidates = append(candidates, mapTarget(item))
	}
	return candidates
}

func buildCredentialTargetView(service string, visibilityLabel string, issues []models.Issue, candidates []credentialPathTarget) credentialPathTargetView {
	return credentialPathTargetView{
		Service:         service,
		Candidates:      candidates,
		VisibilityNote:  targetVisibilityNote(visibilityLabel, issues),
		VisibilityIssue: targetVisibilityIssue(issues),
	}
}

func buildKeyVaultTargetView(output models.KeyVaultOutput) credentialPathTargetView {
	return buildCredentialTargetView(
		"keyvault",
		"Key Vault",
		output.Issues,
		buildCredentialTargets(output.KeyVaults, func(vault models.KeyVaultAsset) credentialPathTarget {
			return credentialPathTarget{
				ID:       vault.ID,
				Name:     vault.Name,
				Location: vault.Location,
				VaultURI: vault.VaultURI,
			}
		}),
	)
}

func buildKeyVaultCredentialRecord(
	envVar models.EnvVarSummary,
	joinedSurfaces []models.TokenCredentialSurfaceSummary,
	targetView credentialPathTargetView,
) *models.ChainPathRecord {
	if envVar.ReferenceTarget == nil || *envVar.ReferenceTarget == "" {
		return nil
	}

	targetHost := referenceHost(*envVar.ReferenceTarget)
	targetIDs := []string{}
	targetNames := []string{}
	for _, candidate := range targetView.Candidates {
		if targetHost == referenceHost(stringPtrValue(candidate.VaultURI)) {
			if candidate.ID != "" {
				targetIDs = append(targetIDs, candidate.ID)
			}
			if candidate.Name != "" {
				targetNames = append(targetNames, candidate.Name)
			}
		}
	}

	summary := fmt.Sprintf(
		"%s '%s' maps setting '%s' to the named Key Vault reference '%s'.",
		envVar.AssetKind,
		envVar.AssetName,
		envVar.SettingName,
		*envVar.ReferenceTarget,
	)
	targetResolution := "named target not visible"
	if len(targetNames) > 0 {
		targetResolution = "named match"
		summary = fmt.Sprintf(
			"%s HO-Azure can join that reference to visible Key Vault inventory: %s.",
			summary,
			strings.Join(targetNames, ", "),
		)
	} else if targetView.VisibilityNote != nil && *targetView.VisibilityNote != "" {
		summary = fmt.Sprintf("%s %s", summary, *targetView.VisibilityNote)
	}

	settingName := envVar.SettingName
	record := models.ChainPathRecord{
		ChainID:             credentialPathID(envVar.AssetID, envVar.SettingName, "keyvault"),
		AssetID:             envVar.AssetID,
		AssetName:           envVar.AssetName,
		AssetKind:           envVar.AssetKind,
		Location:            models.StringPtr(envVar.Location),
		SettingName:         &settingName,
		ClueType:            "keyvault-reference",
		Priority:            "high",
		Urgency:             models.StringPtr("review-soon"),
		VisiblePath:         "Key Vault-backed setting -> named vault",
		ConfidenceBoundary:  models.StringPtr("Your current identity can read this secret."),
		TargetService:       "keyvault",
		TargetResolution:    targetResolution,
		EvidenceCommands:    []string{"env-vars", "tokens-credentials", "keyvault"},
		JoinedSurfaceTypes:  joinedSurfaceTypes(joinedSurfaces, "keyvault-reference"),
		TargetCount:         len(targetIDs),
		TargetIDs:           targetIDs,
		TargetNames:         targetNames,
		TargetVisibility:    nil,
		NextReview:          "Check vault access path and referenced secret use.",
		Summary:             summary,
		MissingConfirmation: "The named Key Vault dependency is visible, but current artifacts do not confirm secret read access, secret values, or successful downstream use.",
		RelatedIDs:          mergeRelatedIDs(envVar.RelatedIDs, relatedIDsFromSurfaces(joinedSurfaces), targetIDs),
	}
	return &record
}

func buildCandidateCredentialRecord(
	envVar models.EnvVarSummary,
	joinedSurfaces []models.TokenCredentialSurfaceSummary,
	targetView credentialPathTargetView,
	targetService string,
) models.ChainPathRecord {
	targetResolution := "service hint only"
	targetIDs := []string{}
	targetNames := []string{}
	if targetView.VisibilityIssue != nil {
		targetResolution = "visibility blocked"
	} else {
		locationMatches := make([]credentialPathTarget, 0)
		for _, candidate := range targetView.Candidates {
			if envVar.Location != "" && candidate.Location != nil && *candidate.Location == envVar.Location {
				locationMatches = append(locationMatches, candidate)
			}
		}
		selected := locationMatches
		if len(selected) > 0 {
			targetResolution = "narrowed candidates"
		} else if len(targetView.Candidates) > 0 {
			selected = targetView.Candidates
			targetResolution = "tenant-wide candidates"
		}
		for _, candidate := range selected {
			if candidate.ID != "" {
				targetIDs = append(targetIDs, candidate.ID)
			}
			if candidate.Name != "" {
				targetNames = append(targetNames, candidate.Name)
			}
		}
	}

	priority := "low"
	if targetResolution == "narrowed candidates" && len(targetIDs) == 1 && targetService == "database" {
		priority = "medium"
	}
	urgency := credentialUrgency(priority)
	nextReview := fmt.Sprintf("Review the visible %s candidates next.", targetService)
	if targetService == "database" {
		nextReview = "Current env-vars and token surfaces do not name the exact database target."
	} else if targetService == "storage" {
		nextReview = "Current env-vars and token surfaces do not name the exact storage target."
	}

	targetLabel := targetService
	if targetService == "database" {
		targetLabel = "database"
	}
	summary := fmt.Sprintf(
		"%s '%s' exposes credential-like setting '%s', and the visible naming suggests a %s path.",
		envVar.AssetKind,
		envVar.AssetName,
		envVar.SettingName,
		targetLabel,
	)
	switch targetResolution {
	case "narrowed candidates":
		summary = fmt.Sprintf(
			"%s HO-Azure cannot name the exact %s from the setting alone, but it can narrow the next review set to %d visible %s candidate(s) in the same Azure location: %s.",
			summary,
			targetLabel,
			len(targetNames),
			targetLabel,
			strings.Join(targetNames, ", "),
		)
	case "tenant-wide candidates":
		summary = fmt.Sprintf(
			"%s HO-Azure cannot narrow the exact target beyond the tenant-visible %s candidate set: %s.",
			summary,
			targetLabel,
			strings.Join(targetNames, ", "),
		)
	case "visibility blocked":
		summary = fmt.Sprintf(
			"%s HO-Azure cannot tell which %s it reaches because current credentials do not show enough target-side visibility.",
			summary,
			targetLabel,
		)
	default:
		summary = fmt.Sprintf(
			"%s HO-Azure has not yet narrowed it to a specific %s asset.",
			summary,
			targetLabel,
		)
	}
	if targetView.VisibilityNote != nil && *targetView.VisibilityNote != "" && targetResolution != "narrowed candidates" {
		summary = fmt.Sprintf("%s %s", summary, *targetView.VisibilityNote)
	}

	missing := fmt.Sprintf("Current evidence suggests a %s path, but no concrete downstream target is visible from current inventory and HO-Azure has not proved a working credential.", targetLabel)
	if targetResolution == "narrowed candidates" {
		missing = fmt.Sprintf("The current artifacts do not show a direct %s hostname, connection string value, or confirmed successful credential use from this workload.", targetLabel)
	}
	if targetResolution == "tenant-wide candidates" {
		missing = fmt.Sprintf("Current evidence only narrows this to a broad visible %s set and does not name the exact downstream target. HO-Azure also has not proved a working credential there.", targetLabel)
	}
	if targetResolution == "visibility blocked" && targetView.VisibilityIssue != nil {
		missing = fmt.Sprintf("Current scope does not confirm which %s target this setting reaches. HO-Azure also has not proved a working credential there.", targetLabel)
	}

	settingName := envVar.SettingName
	record := models.ChainPathRecord{
		ChainID:             credentialPathID(envVar.AssetID, envVar.SettingName, targetService),
		AssetID:             envVar.AssetID,
		AssetName:           envVar.AssetName,
		AssetKind:           envVar.AssetKind,
		Location:            models.StringPtr(envVar.Location),
		SettingName:         &settingName,
		ClueType:            "plain-text-secret",
		Priority:            priority,
		Urgency:             models.StringPtr(urgency),
		VisiblePath:         fmt.Sprintf("Credential-like setting -> likely %s path", targetService),
		ConfidenceBoundary:  models.StringPtr(credentialConfidenceBoundary(targetLabel, targetResolution, len(targetNames))),
		TargetService:       targetService,
		TargetResolution:    targetResolution,
		EvidenceCommands:    []string{"env-vars", "tokens-credentials", pluralCredentialTargetCommand(targetService)},
		JoinedSurfaceTypes:  joinedSurfaceTypes(joinedSurfaces, "plain-text-secret"),
		TargetCount:         len(targetIDs),
		TargetIDs:           targetIDs,
		TargetNames:         targetNames,
		TargetVisibility:    targetView.VisibilityIssue,
		NextReview:          nextReview,
		Summary:             summary,
		MissingConfirmation: missing,
		RelatedIDs:          mergeRelatedIDs(envVar.RelatedIDs, relatedIDsFromSurfaces(joinedSurfaces), targetIDs),
	}
	return record
}

func credentialUrgency(priority string) string {
	if priority == "low" {
		return "bookmark"
	}
	return "review-soon"
}

func credentialConfidenceBoundary(targetLabel string, targetResolution string, targetCount int) string {
	switch targetResolution {
	case "named match":
		return "Your current identity can read this secret."
	case "narrowed candidates":
		candidateLabel := targetLabel
		if targetCount == 1 {
			return fmt.Sprintf("HO-Azure narrowed this to %d visible %s candidate, but the loaded evidence does not name the exact target, so this setting is not confirmed to reach it.", targetCount, candidateLabel)
		}
		return fmt.Sprintf("HO-Azure narrowed this to %d visible %s candidates, but the loaded evidence does not name the exact target, so this setting is not confirmed to reach it.", targetCount, candidateLabel)
	case "tenant-wide candidates":
		return fmt.Sprintf("HO-Azure can only narrow this to a broad visible %s set, so this setting is not confirmed to reach a named target.", targetLabel)
	case "visibility blocked":
		return fmt.Sprintf("HO-Azure cannot confirm the downstream %s target from current visibility.", targetLabel)
	case "named target not visible":
		return "This app names a Key Vault HO-Azure cannot see in current inventory."
	default:
		return fmt.Sprintf("HO-Azure sees a likely %s path, but no target inventory is visible yet.", targetLabel)
	}
}

func credentialLikeEnvVar(envVar models.EnvVarSummary, joinedSurfaces []models.TokenCredentialSurfaceSummary) bool {
	if envVar.LooksSensitive && envVar.ValueType == "plain-text" {
		return true
	}
	for _, surface := range joinedSurfaces {
		if surface.SurfaceType == models.TokenCredentialSurfacePlainTextSecret {
			return true
		}
	}
	return false
}

func targetVisibilityNote(targetLabel string, issues []models.Issue) *string {
	for _, issue := range issues {
		if issue.Kind == "permission_denied" || issue.Kind == "partial_collection" {
			text := fmt.Sprintf("Current scope may not show full %s visibility, so this target picture may be incomplete.", targetLabel)
			return &text
		}
	}
	return nil
}

func targetVisibilityIssue(issues []models.Issue) *string {
	for _, issue := range issues {
		if issue.Kind == "permission_denied" || issue.Kind == "partial_collection" {
			text := issue.Kind + ": " + issue.Message
			return &text
		}
	}
	return nil
}

func parseOperatorSignal(value string) map[string]string {
	signal := map[string]string{}
	for _, part := range strings.Split(value, ";") {
		key, raw, found := strings.Cut(strings.TrimSpace(part), "=")
		if !found {
			continue
		}
		signal[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(raw)
	}
	return signal
}

func normalizeReferenceTarget(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	normalized = strings.TrimPrefix(normalized, "https://")
	return strings.Trim(normalized, "/")
}

func referenceHost(value string) string {
	normalized := normalizeReferenceTarget(value)
	if host, _, found := strings.Cut(normalized, "/"); found {
		return host
	}
	return normalized
}

func credentialPathID(assetID string, settingName string, targetService string) string {
	return fmt.Sprintf("credential-path::%s::%s::%s", assetID, strings.ReplaceAll(strings.ToLower(settingName), "_", "-"), targetService)
}

func pluralCredentialTargetCommand(targetService string) string {
	if targetService == "database" {
		return "databases"
	}
	return targetService
}

func joinedSurfaceTypes(joinedSurfaces []models.TokenCredentialSurfaceSummary, fallback string) []string {
	seen := map[string]struct{}{}
	types := []string{}
	for _, surface := range joinedSurfaces {
		value := string(surface.SurfaceType)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		types = append(types, value)
	}
	if len(types) == 0 {
		return []string{fallback}
	}
	sort.Strings(types)
	return types
}

func mergeRelatedIDs(groups ...[]string) []string {
	seen := map[string]struct{}{}
	merged := []string{}
	for _, group := range groups {
		for _, value := range group {
			if strings.TrimSpace(value) == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			merged = append(merged, value)
		}
	}
	return merged
}

func relatedIDsFromSurfaces(surfaces []models.TokenCredentialSurfaceSummary) []string {
	merged := []string{}
	for _, surface := range surfaces {
		merged = append(merged, surface.RelatedIDs...)
	}
	return merged
}

func prioritySortValue(priority string) int {
	switch priority {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	default:
		return 9
	}
}

func firstNonEmptyPtr(values ...*string) *string {
	for _, value := range values {
		if value != nil && strings.TrimSpace(*value) != "" {
			return value
		}
	}
	return nil
}

func chainsSourceArtifacts(commands []string) []models.ChainSourceArtifact {
	artifacts := make([]models.ChainSourceArtifact, 0, len(commands))
	base := filepath.Join(os.TempDir(), "ho-azure-chains-inputs", "loot")
	for _, command := range commands {
		artifacts = append(artifacts, models.ChainSourceArtifact{
			Command:      command,
			ArtifactType: "loot",
			Path:         filepath.Join(base, command+".json"),
		})
	}
	return artifacts
}
