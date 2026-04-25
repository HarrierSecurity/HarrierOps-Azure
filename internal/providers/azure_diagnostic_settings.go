package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

const (
	armResourcesAPIVersion          = "2021-04-01"
	armDiagnosticSettingsAPIVersion = "2021-05-01-preview"
)

func (provider AzureProvider) DiagnosticSettings(ctx context.Context, tenant string, subscription string) (DiagnosticSettingsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return DiagnosticSettingsFacts{}, err
	}

	resourcePath := "/subscriptions/" + session.subscription.ID + "/resources"
	resources, err := armListObjects(ctx, session.credential, resourcePath, armResourcesAPIVersion)
	if err != nil {
		return DiagnosticSettingsFacts{}, err
	}

	sources := []models.DiagnosticSettingsSource{
		diagnosticSettingsSubscriptionSource(session.subscription.ID),
	}
	for _, resource := range resources {
		id := mapStringValue(resource, "id")
		if id == "" {
			continue
		}
		sources = append(sources, models.DiagnosticSettingsSource{
			ID:            id,
			Name:          firstNonEmpty(mapStringValue(resource, "name"), resourceNameFromID(id), "unknown"),
			Type:          mapStringValue(resource, "type"),
			ResourceGroup: resourceGroupFromID(id),
			Location:      mapStringValue(resource, "location"),
		})
	}

	issues := []models.Issue{}
	for index := range sources {
		settings, settingIssues := collectDiagnosticSettings(ctx, session, sources[index].ID)
		issues = append(issues, settingIssues...)
		categories, categoryIssues := collectDiagnosticSettingsCategories(ctx, session, sources[index].ID)
		issues = append(issues, categoryIssues...)
		sources[index] = diagnosticSettingsHydrateSource(sources[index], settings, categories)
	}

	sort.SliceStable(sources, func(i, j int) bool {
		if diagnosticSettingsSourceRank(sources[i]) != diagnosticSettingsSourceRank(sources[j]) {
			return diagnosticSettingsSourceRank(sources[i]) > diagnosticSettingsSourceRank(sources[j])
		}
		if sources[i].Type != sources[j].Type {
			return sources[i].Type < sources[j].Type
		}
		return sources[i].Name < sources[j].Name
	})

	return DiagnosticSettingsFacts{
		ArtifactIdentityFacts: azureArtifactIdentityFacts(session),
		TenantID:              session.tenantID,
		SubscriptionID:        session.subscription.ID,
		Sources:               sources,
		Issues:                issues,
	}, nil
}

func diagnosticSettingsSubscriptionSource(subscriptionID string) models.DiagnosticSettingsSource {
	id := "/subscriptions/" + subscriptionID
	return models.DiagnosticSettingsSource{
		ID:       id,
		Name:     "subscription",
		Type:     "Microsoft.Resources/subscriptions",
		Location: "global",
	}
}

func collectDiagnosticSettings(ctx context.Context, session azureSession, sourceID string) ([]models.DiagnosticSettingAsset, []models.Issue) {
	path := strings.TrimRight(sourceID, "/") + "/providers/Microsoft.Insights/diagnosticSettings"
	rows, err := armListObjects(ctx, session.credential, path, armDiagnosticSettingsAPIVersion)
	if err != nil {
		return nil, []models.Issue{issueFromError("diagnostic-settings["+sourceID+"]", err)}
	}
	settings := make([]models.DiagnosticSettingAsset, 0, len(rows))
	for _, row := range rows {
		settings = append(settings, diagnosticSettingFromMap(row, sourceID))
	}
	sort.SliceStable(settings, func(i, j int) bool {
		return settings[i].Name < settings[j].Name
	})
	return settings, nil
}

func collectDiagnosticSettingsCategories(ctx context.Context, session azureSession, sourceID string) ([]models.DiagnosticSettingsCategory, []models.Issue) {
	path := strings.TrimRight(sourceID, "/") + "/providers/Microsoft.Insights/diagnosticSettingsCategories"
	rows, err := armListObjects(ctx, session.credential, path, armDiagnosticSettingsAPIVersion)
	if err != nil {
		return nil, []models.Issue{issueFromError("diagnostic-settings-categories["+sourceID+"]", err)}
	}
	categories := make([]models.DiagnosticSettingsCategory, 0, len(rows))
	for _, row := range rows {
		category := diagnosticSettingsSupportedCategoryFromMap(row)
		if category.Name == "" {
			continue
		}
		categories = append(categories, category)
	}
	sort.SliceStable(categories, func(i, j int) bool {
		if categories[i].Type != categories[j].Type {
			return categories[i].Type < categories[j].Type
		}
		return categories[i].Name < categories[j].Name
	})
	return categories, nil
}

func diagnosticSettingsSupportedCategoryFromMap(row map[string]any) models.DiagnosticSettingsCategory {
	properties := mapValue(row, "properties")
	name := firstNonEmpty(
		mapStringValue(row, "name"),
		mapStringValue(properties, "category"),
		resourceNameFromID(mapStringValue(row, "id")),
	)
	categoryType := firstNonEmpty(mapStringValue(properties, "categoryType", "category_type"), "log")
	return models.DiagnosticSettingsCategory{
		Name:    name,
		Type:    categoryType,
		Enabled: true,
	}
}

func diagnosticSettingFromMap(row map[string]any, sourceID string) models.DiagnosticSettingAsset {
	properties := mapValue(row, "properties")
	id := mapStringValue(row, "id")
	logs := diagnosticSettingsCategories(properties, "logs", "log")
	metrics := diagnosticSettingsCategories(properties, "metrics", "metric")
	destinations := diagnosticSettingsDestinations(properties)
	setting := models.DiagnosticSettingAsset{
		ID:                   id,
		Name:                 firstNonEmpty(mapStringValue(row, "name"), resourceNameFromID(id), "unknown"),
		SourceResourceID:     sourceID,
		Destinations:         destinations,
		Logs:                 logs,
		Metrics:              metrics,
		EnabledCategories:    diagnosticSettingsEnabledCategories(logs, metrics),
		DisabledCategories:   diagnosticSettingsDisabledCategories(logs, metrics),
		CategoryGroups:       diagnosticSettingsCategoryGroups(logs),
		HighSignalCategories: diagnosticSettingsHighSignalCategories(logs, metrics),
		DestinationTypes:     diagnosticSettingsDestinationTypes(destinations),
	}
	setting.RelatedIDs = diagnosticSettingsRelatedIDs(sourceID, setting)
	setting.Summary = diagnosticSettingSummary(setting)
	return setting
}

func diagnosticSettingsCategories(properties map[string]any, key string, categoryType string) []models.DiagnosticSettingsCategory {
	categories := []models.DiagnosticSettingsCategory{}
	for _, item := range listValue(properties, key) {
		mapped, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := firstNonEmpty(mapStringValue(mapped, "category"), mapStringValue(mapped, "categoryGroup"), "unknown")
		if name == "unknown" {
			continue
		}
		categories = append(categories, models.DiagnosticSettingsCategory{
			Name:    name,
			Type:    categoryType,
			Enabled: diagnosticSettingsBoolValue(mapped["enabled"]),
		})
	}
	sort.SliceStable(categories, func(i, j int) bool {
		if categories[i].Type != categories[j].Type {
			return categories[i].Type < categories[j].Type
		}
		return categories[i].Name < categories[j].Name
	})
	return categories
}

func diagnosticSettingsDestinations(properties map[string]any) []models.DiagnosticSettingsDestination {
	candidates := []models.DiagnosticSettingsDestination{
		{Type: "logAnalytics", ResourceID: stringPtr(mapStringValue(properties, "workspaceId")), Detail: stringPtr(mapStringValue(properties, "logAnalyticsDestinationType"))},
		{Type: "storage", ResourceID: stringPtr(mapStringValue(properties, "storageAccountId"))},
		{Type: "eventHubs", ResourceID: stringPtr(mapStringValue(properties, "eventHubAuthorizationRuleId")), Detail: stringPtr(mapStringValue(properties, "eventHubName"))},
		{Type: "marketplacePartner", ResourceID: stringPtr(mapStringValue(properties, "marketplacePartnerId"))},
	}
	destinations := []models.DiagnosticSettingsDestination{}
	for _, destination := range candidates {
		if stringPtrValue(destination.ResourceID) == "" && stringPtrValue(destination.Detail) == "" {
			continue
		}
		destinations = append(destinations, destination)
	}
	return destinations
}

func diagnosticSettingsHydrateSource(source models.DiagnosticSettingsSource, settings []models.DiagnosticSettingAsset, supported []models.DiagnosticSettingsCategory) models.DiagnosticSettingsSource {
	enabled := []string{}
	disabled := []string{}
	supportedNames := []string{}
	groups := []string{}
	highSignal := []string{}
	destinationTypes := []string{}
	relatedIDs := []string{source.ID}
	for _, setting := range settings {
		enabled = append(enabled, setting.EnabledCategories...)
		disabled = append(disabled, setting.DisabledCategories...)
		groups = append(groups, setting.CategoryGroups...)
		highSignal = append(highSignal, setting.HighSignalCategories...)
		destinationTypes = append(destinationTypes, setting.DestinationTypes...)
		relatedIDs = append(relatedIDs, setting.RelatedIDs...)
	}
	for _, category := range supported {
		supportedNames = append(supportedNames, category.Name)
		if diagnosticSettingsCategoryLooksHighSignal(category.Name) {
			highSignal = append(highSignal, category.Name)
		}
	}
	source.DiagnosticSettings = settings
	source.DiagnosticSettingCount = len(settings)
	source.EnabledCategories = sortedUniqueStrings(enabled)
	source.DisabledCategories = sortedUniqueStrings(disabled)
	source.SupportedCategories = sortedUniqueStrings(supportedNames)
	source.NotExportedSupported = diagnosticSettingsNotExportedSupported(supported, source.EnabledCategories)
	source.SupportedCategoryCatalog = len(supported) > 0
	source.CategoryGroups = sortedUniqueStrings(groups)
	source.HighSignalCategories = sortedUniqueStrings(highSignal)
	source.DestinationTypes = sortedUniqueStrings(destinationTypes)
	source.HasDiagnosticSettings = len(settings) > 0
	source.HasPartialLogPosture = len(source.DisabledCategories) > 0 || len(source.NotExportedSupported) > 0 || (len(source.EnabledCategories) > 0 && len(source.DestinationTypes) > 0)
	source.HasHighSignalLogPosture = len(source.HighSignalCategories) > 0 || diagnosticSettingsSourceLooksHighSignal(source)
	source.HasNonWorkspaceDestination = diagnosticSettingsHasNonWorkspaceDestination(source.DestinationTypes)
	source.RelatedIDs = sortedUniqueStrings(relatedIDs)
	source.Summary = diagnosticSettingsSourceSummary(source)
	return source
}

func diagnosticSettingsNotExportedSupported(supported []models.DiagnosticSettingsCategory, enabled []string) []string {
	enabledSet := map[string]bool{}
	for _, category := range enabled {
		enabledSet[strings.ToLower(category)] = true
	}
	values := []string{}
	for _, category := range supported {
		if enabledSet["alllogs"] && !strings.EqualFold(category.Type, "metric") {
			continue
		}
		if !enabledSet[strings.ToLower(category.Name)] {
			values = append(values, category.Name)
		}
	}
	return sortedUniqueStrings(values)
}

func diagnosticSettingsEnabledCategories(groups ...[]models.DiagnosticSettingsCategory) []string {
	values := []string{}
	for _, group := range groups {
		for _, category := range group {
			if category.Enabled {
				values = append(values, category.Name)
			}
		}
	}
	return sortedUniqueStrings(values)
}

func diagnosticSettingsDisabledCategories(groups ...[]models.DiagnosticSettingsCategory) []string {
	values := []string{}
	for _, group := range groups {
		for _, category := range group {
			if !category.Enabled {
				values = append(values, category.Name)
			}
		}
	}
	return sortedUniqueStrings(values)
}

func diagnosticSettingsCategoryGroups(logs []models.DiagnosticSettingsCategory) []string {
	values := []string{}
	for _, category := range logs {
		if category.Enabled && (strings.EqualFold(category.Name, "allLogs") || strings.Contains(strings.ToLower(category.Name), "audit")) {
			values = append(values, category.Name)
		}
	}
	return sortedUniqueStrings(values)
}

func diagnosticSettingsHighSignalCategories(groups ...[]models.DiagnosticSettingsCategory) []string {
	values := []string{}
	for _, group := range groups {
		for _, category := range group {
			if diagnosticSettingsCategoryLooksHighSignal(category.Name) {
				values = append(values, category.Name)
			}
		}
	}
	return sortedUniqueStrings(values)
}

func diagnosticSettingsDestinationTypes(destinations []models.DiagnosticSettingsDestination) []string {
	values := []string{}
	for _, destination := range destinations {
		values = append(values, destination.Type)
	}
	return sortedUniqueStrings(values)
}

func diagnosticSettingsBoolValue(value any) bool {
	boolValue, ok := value.(bool)
	return ok && boolValue
}

func diagnosticSettingsRelatedIDs(sourceID string, setting models.DiagnosticSettingAsset) []string {
	values := []string{sourceID, setting.ID}
	for _, destination := range setting.Destinations {
		values = append(values, stringPtrValue(destination.ResourceID))
	}
	return sortedUniqueStrings(values)
}

func diagnosticSettingsHasNonWorkspaceDestination(types []string) bool {
	for _, destinationType := range types {
		if destinationType != "logAnalytics" {
			return true
		}
	}
	return false
}

func diagnosticSettingsSourceLooksHighSignal(source models.DiagnosticSettingsSource) bool {
	normalized := strings.ToLower(source.Type)
	return strings.Contains(normalized, "keyvault") ||
		strings.Contains(normalized, "storage") ||
		strings.Contains(normalized, "sql") ||
		strings.Contains(normalized, "web/sites") ||
		strings.Contains(normalized, "container")
}

func diagnosticSettingsCategoryLooksHighSignal(category string) bool {
	normalized := strings.ToLower(category)
	return strings.Contains(normalized, "audit") ||
		strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "key") ||
		strings.Contains(normalized, "auth") ||
		strings.Contains(normalized, "signin") ||
		strings.Contains(normalized, "firewall") ||
		strings.Contains(normalized, "request")
}

func diagnosticSettingsSourceRank(source models.DiagnosticSettingsSource) int {
	rank := 0
	if source.HasNonWorkspaceDestination {
		rank += 2
	}
	if len(source.DisabledCategories) > 0 {
		rank += 2
	}
	if source.HasHighSignalLogPosture {
		rank += 2
	}
	if !source.HasDiagnosticSettings && diagnosticSettingsSourceLooksHighSignal(source) {
		rank++
	}
	return rank
}

func diagnosticSettingSummary(setting models.DiagnosticSettingAsset) string {
	return fmt.Sprintf("diagnostic setting %q exports %d enabled categor(ies), has %d disabled categor(ies), and routes to %s.", setting.Name, len(setting.EnabledCategories), len(setting.DisabledCategories), strings.Join(setting.DestinationTypes, ", "))
}

func diagnosticSettingsSourceSummary(source models.DiagnosticSettingsSource) string {
	if len(source.DiagnosticSettings) == 0 {
		return fmt.Sprintf("%s %q has no visible diagnostic settings.", source.Type, source.Name)
	}
	return fmt.Sprintf("%s %q has %d diagnostic setting(s), %d enabled categor(ies), %d disabled categor(ies), and destinations: %s.", source.Type, source.Name, len(source.DiagnosticSettings), len(source.EnabledCategories), len(source.DisabledCategories), strings.Join(source.DestinationTypes, ", "))
}
