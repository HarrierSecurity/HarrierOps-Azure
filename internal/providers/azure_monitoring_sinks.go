package providers

import (
	"context"
	"strings"

	"harrierops-azure/internal/models"
)

func (provider AzureProvider) MonitoringSinks(ctx context.Context, tenant string, subscription string) (MonitoringSinksFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return MonitoringSinksFacts{}, err
	}

	resources, err := armListObjects(ctx, session.credential, "/subscriptions/"+session.subscription.ID+"/resources", armResourcesAPIVersion)
	if err != nil {
		return MonitoringSinksFacts{}, err
	}

	sinks := []models.MonitoringSinkAsset{}
	sentinelWorkspaces := monitoringSinksSentinelWorkspaceNames(resources)
	for _, resource := range resources {
		sink, ok := monitoringSinkFromResource(resource, sentinelWorkspaces)
		if !ok {
			continue
		}
		sinks = append(sinks, sink)
	}

	issues := []models.Issue{}
	dcrFacts, err := provider.DCR(ctx, tenant, subscription)
	if err != nil {
		issues = append(issues, issueFromError("monitoring-sinks.dcr", err))
	} else {
		issues = append(issues, dcrFacts.Issues...)
		monitoringSinksEnsureDCRDestinations(&sinks, dcrFacts.DCRs)
		monitoringSinksAttachDCRReferences(sinks, dcrFacts.DCRs)
	}

	diagnosticFacts, err := provider.DiagnosticSettings(ctx, tenant, subscription)
	if err != nil {
		issues = append(issues, issueFromError("monitoring-sinks.diagnostic-settings", err))
	} else {
		issues = append(issues, diagnosticFacts.Issues...)
		monitoringSinksEnsureDiagnosticDestinations(&sinks, diagnosticFacts.Sources)
		monitoringSinksAttachDiagnosticReferences(sinks, diagnosticFacts.Sources)
	}

	sinks = monitoringSinksFinalize(sinks)

	return MonitoringSinksFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		Sinks:          sinks,
		Issues:         issues,
	}, nil
}

func monitoringSinkFromResource(resource map[string]any, sentinelWorkspaces map[string]bool) (models.MonitoringSinkAsset, bool) {
	id := mapStringValue(resource, "id")
	resourceType := mapStringValue(resource, "type")
	name := firstNonEmpty(mapStringValue(resource, "name"), resourceNameFromID(id), "unknown")
	kind := monitoringSinkKind(resourceType)
	if kind == "" {
		return models.MonitoringSinkAsset{}, false
	}
	sink := models.MonitoringSinkAsset{
		ID:               id,
		Name:             name,
		Kind:             kind,
		ResourceType:     resourceType,
		ResourceGroup:    resourceGroupFromID(id),
		Location:         mapStringValue(resource, "location"),
		VisibilitySource: "resource inventory",
		RelatedIDs:       []string{id},
	}
	if kind == "logAnalytics" {
		enabled := sentinelWorkspaces[strings.ToLower(name)]
		sink.SentinelEnabled = boolPtr(enabled)
		if enabled {
			sink.Kind = "sentinel"
		}
	}
	return sink, true
}

func monitoringSinkKind(resourceType string) string {
	switch strings.ToLower(strings.TrimSpace(resourceType)) {
	case "microsoft.operationalinsights/workspaces":
		return "logAnalytics"
	case "microsoft.eventhub/namespaces", "microsoft.eventhub/namespaces/authorizationrules":
		return "eventHubs"
	case "microsoft.storage/storageaccounts":
		return "storage"
	default:
		return ""
	}
}

func monitoringSinksSentinelWorkspaceNames(resources []map[string]any) map[string]bool {
	values := map[string]bool{}
	for _, resource := range resources {
		if !strings.EqualFold(mapStringValue(resource, "type"), "Microsoft.OperationsManagement/solutions") {
			continue
		}
		name := mapStringValue(resource, "name")
		normalized := strings.ToLower(name)
		if !strings.Contains(normalized, "securityinsights") {
			continue
		}
		if start := strings.Index(name, "("); start >= 0 {
			if end := strings.Index(name[start+1:], ")"); end >= 0 {
				values[strings.ToLower(name[start+1:start+1+end])] = true
			}
		}
		properties := mapValue(resource, "properties")
		workspaceID := firstNonEmpty(mapStringValue(properties, "workspaceResourceId"), mapStringValue(properties, "workspaceResourceID"))
		if workspaceName := resourceNameFromID(workspaceID); workspaceName != "" {
			values[strings.ToLower(workspaceName)] = true
		}
	}
	return values
}

func monitoringSinksEnsureDCRDestinations(sinks *[]models.MonitoringSinkAsset, dcrs []models.DCRAsset) {
	for _, dcr := range dcrs {
		for _, destination := range dcr.Destinations {
			monitoringSinksEnsureDestination(sinks, destination.Type, stringPtrValue(destination.ResourceID), stringPtrValue(destination.Detail))
		}
	}
}

func monitoringSinksEnsureDiagnosticDestinations(sinks *[]models.MonitoringSinkAsset, sources []models.DiagnosticSettingsSource) {
	for _, source := range sources {
		for _, setting := range source.DiagnosticSettings {
			for _, destination := range setting.Destinations {
				monitoringSinksEnsureDestination(sinks, destination.Type, stringPtrValue(destination.ResourceID), stringPtrValue(destination.Detail))
			}
		}
	}
}

func monitoringSinksEnsureDestination(sinks *[]models.MonitoringSinkAsset, destinationType string, resourceID string, detail string) {
	key := monitoringSinkDestinationKey(destinationType, resourceID, detail)
	if key == "" || monitoringSinksFindIndex(*sinks, key) >= 0 {
		return
	}
	*sinks = append(*sinks, models.MonitoringSinkAsset{
		ID:               key,
		Name:             firstNonEmpty(resourceNameFromID(resourceID), detail, destinationType),
		Kind:             monitoringSinkKindFromDestination(destinationType),
		ResourceType:     monitoringSinkResourceType(destinationType, resourceID),
		ResourceGroup:    resourceGroupFromID(resourceID),
		VisibilitySource: "declared destination",
		RelatedIDs:       []string{key},
	})
}

func monitoringSinksAttachDCRReferences(sinks []models.MonitoringSinkAsset, dcrs []models.DCRAsset) {
	for _, dcr := range dcrs {
		for _, destination := range dcr.Destinations {
			key := monitoringSinkDestinationKey(destination.Type, stringPtrValue(destination.ResourceID), stringPtrValue(destination.Detail))
			index := monitoringSinksFindIndex(sinks, key)
			if index < 0 {
				continue
			}
			sinks[index].References = append(sinks[index].References, models.MonitoringSinkReference{
				SourceCommand:     "dcr",
				SourceResourceID:  dcr.ID,
				SourceName:        dcr.Name,
				ReferenceName:     stringPtr(destination.Name),
				ReferenceType:     destination.Type,
				DestinationDetail: destination.Detail,
			})
		}
	}
}

func monitoringSinksAttachDiagnosticReferences(sinks []models.MonitoringSinkAsset, sources []models.DiagnosticSettingsSource) {
	for _, source := range sources {
		for _, setting := range source.DiagnosticSettings {
			for _, destination := range setting.Destinations {
				key := monitoringSinkDestinationKey(destination.Type, stringPtrValue(destination.ResourceID), stringPtrValue(destination.Detail))
				index := monitoringSinksFindIndex(sinks, key)
				if index < 0 {
					continue
				}
				sinks[index].References = append(sinks[index].References, models.MonitoringSinkReference{
					SourceCommand:     "diagnostic-settings",
					SourceResourceID:  source.ID,
					SourceName:        source.Name,
					ReferenceName:     stringPtr(setting.Name),
					ReferenceType:     destination.Type,
					DestinationDetail: destination.Detail,
				})
			}
		}
	}
}

func monitoringSinksFindIndex(sinks []models.MonitoringSinkAsset, key string) int {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if normalized == "" {
		return -1
	}
	for index, sink := range sinks {
		if strings.ToLower(strings.TrimSpace(sink.ID)) == normalized {
			return index
		}
	}
	return -1
}

func monitoringSinkDestinationKey(destinationType string, resourceID string, detail string) string {
	resourceID = strings.TrimSpace(resourceID)
	if resourceID != "" {
		return resourceID
	}
	if strings.TrimSpace(detail) == "" {
		return ""
	}
	return "declared:" + monitoringSinkKindFromDestination(destinationType) + ":" + strings.TrimSpace(detail)
}

func monitoringSinkKindFromDestination(destinationType string) string {
	switch strings.ToLower(strings.TrimSpace(destinationType)) {
	case "loganalytics":
		return "logAnalytics"
	case "eventhubs":
		return "eventHubs"
	case "storage":
		return "storage"
	case "marketplacepartner":
		return "marketplacePartner"
	default:
		return firstNonEmpty(destinationType, "unknown")
	}
}

func monitoringSinkResourceType(destinationType string, resourceID string) string {
	if resourceID != "" {
		return resourceTypeFromID(resourceID)
	}
	switch monitoringSinkKindFromDestination(destinationType) {
	case "logAnalytics":
		return "Microsoft.OperationalInsights/workspaces"
	case "eventHubs":
		return "Microsoft.EventHub/namespaces/authorizationRules"
	case "storage":
		return "Microsoft.Storage/storageAccounts"
	default:
		return "declared destination"
	}
}

func resourceTypeFromID(resourceID string) string {
	normalized := strings.Trim(resourceID, "/")
	parts := strings.Split(normalized, "/")
	for i := 0; i < len(parts)-2; i++ {
		if strings.EqualFold(parts[i], "providers") {
			return parts[i+1] + "/" + parts[i+2]
		}
	}
	return ""
}

func monitoringSinkReferenceIDs(references []models.MonitoringSinkReference) []string {
	values := []string{}
	for _, reference := range references {
		values = append(values, reference.SourceResourceID)
	}
	return values
}
