package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

func (provider AzureProvider) AppInsights(ctx context.Context, tenant string, subscription string) (AppInsightsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return AppInsightsFacts{}, err
	}

	resources, err := armListObjects(ctx, session.credential, "/subscriptions/"+session.subscription.ID+"/resources", armResourcesAPIVersion)
	if err != nil {
		return AppInsightsFacts{}, err
	}
	components := []models.AppInsightsComponent{}
	for _, resource := range resources {
		if !strings.EqualFold(mapStringValue(resource, "type"), "Microsoft.Insights/components") {
			continue
		}
		components = append(components, appInsightsComponentFromResource(resource))
	}

	webAppsState, err := provider.webAppsState(session)
	if err != nil {
		return AppInsightsFacts{}, fmt.Errorf("build web apps client: %w", err)
	}
	targets := []models.AppInsightsAppTarget{}
	issues := []models.Issue{}
	apps, listErr := webAppsState.list(ctx)
	if listErr != nil {
		issues = append(issues, issueFromError("appinsights.web_apps", listErr))
	}
	for _, app := range apps {
		if app.assetKind == "" || app.resourceGroup == "" || app.name == "" {
			continue
		}
		settingsMap, err := webAppsState.settingsMap(ctx, app)
		if err != nil {
			issues = append(issues, issueFromError("appinsights["+app.resourceGroup+"/"+app.name+"].app_settings", err))
			continue
		}
		target := appInsightsTargetFromSettings(app.appMap, app.assetKind, mapValue(settingsMap, "properties"))
		if len(target.InstrumentationClues) == 0 && len(target.SamplingClues) == 0 && len(target.FilteringClues) == 0 && len(target.LoggingLevelClues) == 0 {
			continue
		}
		targets = append(targets, target)
	}

	sort.SliceStable(components, func(i, j int) bool {
		return components[i].Name < components[j].Name
	})
	sort.SliceStable(targets, func(i, j int) bool {
		if appInsightsTargetRank(targets[i]) != appInsightsTargetRank(targets[j]) {
			return appInsightsTargetRank(targets[i]) > appInsightsTargetRank(targets[j])
		}
		return targets[i].Name < targets[j].Name
	})

	return AppInsightsFacts{
		ArtifactIdentityFacts: azureArtifactIdentityFacts(session),
		TenantID:              session.tenantID,
		SubscriptionID:        session.subscription.ID,
		Components:            components,
		Targets:               targets,
		Issues:                issues,
	}, nil
}

func appInsightsComponentFromResource(resource map[string]any) models.AppInsightsComponent {
	id := mapStringValue(resource, "id")
	properties := mapValue(resource, "properties")
	component := models.AppInsightsComponent{
		ID:                  id,
		Name:                firstNonEmpty(mapStringValue(resource, "name"), resourceNameFromID(id), "unknown"),
		ResourceGroup:       resourceGroupFromID(id),
		Location:            mapStringValue(resource, "location"),
		Kind:                stringPtr(mapStringValue(resource, "kind")),
		ApplicationType:     stringPtr(mapStringValue(properties, "Application_Type", "applicationType", "application_type")),
		WorkspaceResourceID: stringPtr(mapStringValue(properties, "WorkspaceResourceId", "workspaceResourceId", "workspace_resource_id")),
		IngestionMode:       stringPtr(mapStringValue(properties, "IngestionMode", "ingestionMode", "ingestion_mode")),
		RelatedIDs:          []string{id},
	}
	component.Summary = fmt.Sprintf("Application Insights component %q is visible in %s.", component.Name, firstNonEmpty(component.Location, "unknown location"))
	return component
}

func appInsightsTargetFromSettings(app map[string]any, assetKind string, settings map[string]any) models.AppInsightsAppTarget {
	appID := mapStringValue(app, "id")
	target := models.AppInsightsAppTarget{
		ID:            appID,
		Name:          firstNonEmpty(mapStringValue(app, "name"), resourceNameFromID(appID), "unknown"),
		Kind:          assetKind,
		ResourceGroup: resourceGroupFromID(appID),
		Location:      mapStringValue(app, "location"),
		RelatedIDs:    []string{appID},
	}
	for settingName, settingValue := range settings {
		class := appInsightsSettingClass(settingName)
		clue := appInsightsSettingClue(settingName, settingValue, class)
		switch class {
		case "instrumentation":
			target.InstrumentationClues = append(target.InstrumentationClues, clue)
		case "sampling":
			target.SamplingClues = append(target.SamplingClues, clue)
		case "filtering":
			target.FilteringClues = append(target.FilteringClues, clue)
		case "logging-level":
			target.LoggingLevelClues = append(target.LoggingLevelClues, clue)
		}
	}
	target.InstrumentationClues = sortedUniqueStrings(target.InstrumentationClues)
	target.SamplingClues = sortedUniqueStrings(target.SamplingClues)
	target.FilteringClues = sortedUniqueStrings(target.FilteringClues)
	target.LoggingLevelClues = sortedUniqueStrings(target.LoggingLevelClues)
	target.VisibleTelemetryTypes = appInsightsTelemetryTypes(target)
	target.Summary = appInsightsTargetSummary(target)
	return target
}

func appInsightsSettingClue(name string, value any, class string) string {
	name = strings.TrimSpace(name)
	if class != "sampling" && class != "logging-level" {
		return name
	}
	safeValue := appInsightsSafeSettingValue(name, value)
	if safeValue == "" {
		return name
	}
	return name + "=" + safeValue
}

func appInsightsSafeSettingValue(name string, value any) string {
	if looksSensitiveSettingName(name) {
		return ""
	}
	text := strings.TrimSpace(stringValue(value))
	if text == "" || appInsightsValueLooksSecret(text) {
		return ""
	}
	if len(text) > 80 {
		text = text[:77] + "..."
	}
	return text
}

func appInsightsValueLooksSecret(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if strings.HasPrefix(normalized, "@microsoft.keyvault(") {
		return true
	}
	for _, token := range []string{
		"accountkey=",
		"clientsecret=",
		"instrumentationkey=",
		"sharedaccesskey=",
		"sig=",
		"password=",
		"secret=",
	} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func appInsightsSettingClass(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.Contains(normalized, "sampling"):
		return "sampling"
	case strings.Contains(normalized, "telemetryprocessor") ||
		(strings.Contains(normalized, "filter") || strings.Contains(normalized, "processor")) &&
			(strings.Contains(normalized, "applicationinsights") || strings.Contains(normalized, "appinsights") || strings.Contains(normalized, "telemetry")):
		return "filtering"
	case strings.Contains(normalized, "loglevel") && (strings.Contains(normalized, "applicationinsights") || strings.Contains(normalized, "logging")):
		return "logging-level"
	case strings.Contains(normalized, "applicationinsights") || strings.Contains(normalized, "appinsights") || strings.Contains(normalized, "instrumentationkey"):
		return "instrumentation"
	default:
		return ""
	}
}

func appInsightsTelemetryTypes(target models.AppInsightsAppTarget) []string {
	values := []string{}
	for _, clue := range append(append([]string{}, target.SamplingClues...), append(target.FilteringClues, target.LoggingLevelClues...)...) {
		normalized := strings.ToLower(clue)
		switch {
		case strings.Contains(normalized, "request"):
			values = append(values, "requests")
		case strings.Contains(normalized, "depend"):
			values = append(values, "dependencies")
		case strings.Contains(normalized, "exception") || strings.Contains(normalized, "error"):
			values = append(values, "exceptions")
		case strings.Contains(normalized, "trace") || strings.Contains(normalized, "log"):
			values = append(values, "traces")
		}
	}
	return sortedUniqueStrings(values)
}

func appInsightsTargetRank(target models.AppInsightsAppTarget) int {
	return len(target.FilteringClues)*3 + len(target.SamplingClues)*2 + len(target.LoggingLevelClues) + len(target.InstrumentationClues)
}

func appInsightsTargetSummary(target models.AppInsightsAppTarget) string {
	parts := []string{}
	if len(target.InstrumentationClues) > 0 {
		parts = append(parts, fmt.Sprintf("%d instrumentation clue(s)", len(target.InstrumentationClues)))
	}
	if len(target.SamplingClues) > 0 {
		parts = append(parts, fmt.Sprintf("%d sampling clue(s)", len(target.SamplingClues)))
	}
	if len(target.FilteringClues) > 0 {
		parts = append(parts, fmt.Sprintf("%d filtering clue(s)", len(target.FilteringClues)))
	}
	if len(target.LoggingLevelClues) > 0 {
		parts = append(parts, fmt.Sprintf("%d logging-level clue(s)", len(target.LoggingLevelClues)))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("%s %q has no visible Application Insights posture clues.", target.Kind, target.Name)
	}
	return fmt.Sprintf("%s %q has %s visible from app settings.", target.Kind, target.Name, strings.Join(parts, ", "))
}
