package providers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

const armDataCollectionRulesAPIVersion = "2024-03-11"

func (provider AzureProvider) DCR(ctx context.Context, tenant string, subscription string) (DCRFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return DCRFacts{}, err
	}

	rulePath := "/subscriptions/" + session.subscription.ID + "/providers/Microsoft.Insights/dataCollectionRules"
	rules, err := armListObjects(ctx, session.credential, rulePath, armDataCollectionRulesAPIVersion)
	if err != nil {
		return DCRFacts{}, err
	}

	assets := []models.DCRAsset{}
	issues := []models.Issue{}
	for _, rule := range rules {
		associations, associationIssues := collectDCRAssociations(ctx, session, rule)
		issues = append(issues, associationIssues...)
		assets = append(assets, dcrAssetFromMap(rule, associations))
	}

	return DCRFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		DCRs:           assets,
		Issues:         issues,
	}, nil
}

func collectDCRAssociations(ctx context.Context, session azureSession, rule map[string]any) ([]models.DCRAssociation, []models.Issue) {
	ruleID := mapStringValue(rule, "id")
	if ruleID == "" {
		return nil, nil
	}
	path := strings.TrimRight(ruleID, "/") + "/associations"
	rows, err := armListObjects(ctx, session.credential, path, armDataCollectionRulesAPIVersion)
	if err != nil {
		return nil, []models.Issue{issueFromError("dcr.associations["+ruleID+"]", err)}
	}
	associations := make([]models.DCRAssociation, 0, len(rows))
	for _, row := range rows {
		associations = append(associations, dcrAssociationFromMap(row, ruleID))
	}
	sort.SliceStable(associations, func(i, j int) bool {
		if associations[i].TargetID != associations[j].TargetID {
			return associations[i].TargetID < associations[j].TargetID
		}
		return associations[i].Name < associations[j].Name
	})
	return associations, nil
}

func dcrAssetFromMap(rule map[string]any, associations []models.DCRAssociation) models.DCRAsset {
	properties := mapValue(rule, "properties")
	id := mapStringValue(rule, "id")
	dataSources := dcrDataSources(properties)
	dataFlows := dcrDataFlows(properties)
	destinations := dcrDestinations(properties)
	streams := dcrStreams(dataSources, dataFlows)
	highSignalStreams := dcrHighSignalStreams(streams)
	destinationTypes := dcrDestinationTypes(destinations)
	dataSourceTypes := dcrDataSourceTypes(dataSources)
	relatedIDs := dcrRelatedIDs(id, properties, destinations, associations)

	asset := models.DCRAsset{
		ID:                       id,
		Name:                     firstNonEmpty(mapStringValue(rule, "name"), resourceNameFromID(id), "unknown"),
		ResourceGroup:            resourceGroupFromID(id),
		Location:                 mapStringValue(rule, "location"),
		Kind:                     stringPtr(mapStringValue(rule, "kind")),
		Description:              stringPtr(mapStringValue(properties, "description")),
		DataCollectionEndpointID: stringPtr(mapStringValue(properties, "dataCollectionEndpointId")),
		DataSources:              dataSources,
		DataFlows:                dataFlows,
		Destinations:             destinations,
		Associations:             associations,
		DataSourceTypes:          dataSourceTypes,
		Streams:                  streams,
		HighSignalStreams:        highSignalStreams,
		DestinationTypes:         destinationTypes,
		TransformationCount:      dcrTransformationCount(dataSources, dataFlows),
		AssociationCount:         len(associations),
		RelatedIDs:               relatedIDs,
	}
	asset.Summary = dcrSummary(asset)
	return asset
}

func dcrDataSources(properties map[string]any) []models.DCRDataSource {
	rawSources := mapValue(properties, "dataSources")
	sources := []models.DCRDataSource{}
	for sourceType, value := range rawSources {
		for _, item := range listValue(map[string]any{"value": value}, "value") {
			mapped, ok := item.(map[string]any)
			if !ok {
				continue
			}
			sources = append(sources, models.DCRDataSource{
				Name:                    firstNonEmpty(mapStringValue(mapped, "name"), sourceType),
				Type:                    sourceType,
				Streams:                 sortedUniqueStrings(dcrStringList(mapped, "streams", "streamDeclarations")),
				TransformKqlPresent:     strings.TrimSpace(mapStringValue(mapped, "transformKql")) != "",
				TransformKqlFingerprint: dcrTransformFingerprint(mapStringValue(mapped, "transformKql")),
				TransformKqlLength:      dcrTransformLength(mapStringValue(mapped, "transformKql")),
			})
		}
	}
	sort.SliceStable(sources, func(i, j int) bool {
		if sources[i].Type != sources[j].Type {
			return sources[i].Type < sources[j].Type
		}
		return sources[i].Name < sources[j].Name
	})
	return sources
}

func dcrDataFlows(properties map[string]any) []models.DCRDataFlow {
	flows := []models.DCRDataFlow{}
	for _, item := range listValue(properties, "dataFlows") {
		mapped, ok := item.(map[string]any)
		if !ok {
			continue
		}
		transformKql := mapStringValue(mapped, "transformKql")
		flows = append(flows, models.DCRDataFlow{
			Streams:                 sortedUniqueStrings(dcrStringList(mapped, "streams")),
			Destinations:            sortedUniqueStrings(dcrStringList(mapped, "destinations")),
			OutputStream:            stringPtr(mapStringValue(mapped, "outputStream")),
			BuiltInTransform:        stringPtr(mapStringValue(mapped, "builtInTransform")),
			TransformKqlPresent:     strings.TrimSpace(transformKql) != "",
			TransformKqlFingerprint: dcrTransformFingerprint(transformKql),
			TransformKqlLength:      dcrTransformLength(transformKql),
		})
	}
	sort.SliceStable(flows, func(i, j int) bool {
		left := strings.Join(flows[i].Streams, ",") + "|" + strings.Join(flows[i].Destinations, ",")
		right := strings.Join(flows[j].Streams, ",") + "|" + strings.Join(flows[j].Destinations, ",")
		return left < right
	})
	return flows
}

func dcrDestinations(properties map[string]any) []models.DCRDestination {
	rawDestinations := mapValue(properties, "destinations")
	destinations := []models.DCRDestination{}
	for destinationType, value := range rawDestinations {
		for _, item := range dcrDestinationItems(value) {
			mapped, ok := item.(map[string]any)
			if !ok {
				continue
			}
			destinations = append(destinations, models.DCRDestination{
				Name:       firstNonEmpty(mapStringValue(mapped, "name"), destinationType),
				Type:       destinationType,
				ResourceID: stringPtr(dcrDestinationResourceID(mapped)),
				Detail:     stringPtr(dcrDestinationDetail(mapped)),
			})
		}
	}
	sort.SliceStable(destinations, func(i, j int) bool {
		if destinations[i].Type != destinations[j].Type {
			return destinations[i].Type < destinations[j].Type
		}
		return destinations[i].Name < destinations[j].Name
	})
	return destinations
}

func dcrDestinationItems(value any) []any {
	if values, ok := value.([]any); ok {
		return values
	}
	if mapped, ok := value.(map[string]any); ok {
		if nested := listValue(mapped, "value"); len(nested) > 0 {
			return nested
		}
		return []any{mapped}
	}
	return nil
}

func dcrAssociationFromMap(row map[string]any, fallbackRuleID string) models.DCRAssociation {
	properties := mapValue(row, "properties")
	id := mapStringValue(row, "id")
	ruleID := firstNonEmpty(mapStringValue(properties, "dataCollectionRuleId"), fallbackRuleID)
	return models.DCRAssociation{
		ID:                       id,
		Name:                     firstNonEmpty(mapStringValue(row, "name"), resourceNameFromID(id), "unknown"),
		TargetID:                 firstNonEmpty(mapStringValue(properties, "targetResourceId"), dcrAssociationTargetFromID(id)),
		DataCollectionRuleID:     stringPtr(ruleID),
		DataCollectionEndpointID: stringPtr(mapStringValue(properties, "dataCollectionEndpointId")),
		Description:              stringPtr(mapStringValue(properties, "description")),
	}
}

func dcrAssociationTargetFromID(id string) string {
	needle := "/providers/Microsoft.Insights/dataCollectionRuleAssociations/"
	index := strings.Index(strings.ToLower(id), strings.ToLower(needle))
	if index < 0 {
		return ""
	}
	return strings.TrimRight(id[:index], "/")
}

func dcrStringList(input map[string]any, keys ...string) []string {
	values := []string{}
	for _, key := range keys {
		for _, item := range listValue(input, key) {
			if value := strings.TrimSpace(stringValue(item)); value != "" {
				values = append(values, value)
			}
		}
	}
	return values
}

func dcrDestinationResourceID(destination map[string]any) string {
	return firstNonEmpty(
		mapStringValue(destination, "workspaceResourceId"),
		mapStringValue(destination, "resourceId"),
		mapStringValue(destination, "eventHubResourceId"),
		mapStringValue(destination, "storageAccountResourceId"),
		mapStringValue(destination, "accountResourceId"),
	)
}

func dcrDestinationDetail(destination map[string]any) string {
	return firstNonEmpty(
		mapStringValue(destination, "eventHubName"),
		mapStringValue(destination, "containerName"),
		mapStringValue(destination, "databaseName"),
		mapStringValue(destination, "stream"),
		mapStringValue(destination, "name"),
	)
}

func dcrTransformFingerprint(query string) *string {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(query))
	fingerprint := hex.EncodeToString(sum[:])[:12]
	return &fingerprint
}

func dcrTransformLength(query string) *int {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	length := len(query)
	return &length
}

func dcrTransformationCount(dataSources []models.DCRDataSource, dataFlows []models.DCRDataFlow) int {
	count := 0
	for _, source := range dataSources {
		if source.TransformKqlPresent {
			count++
		}
	}
	for _, flow := range dataFlows {
		if flow.TransformKqlPresent || stringPtrValue(flow.BuiltInTransform) != "" {
			count++
		}
	}
	return count
}

func dcrDataSourceTypes(dataSources []models.DCRDataSource) []string {
	values := []string{}
	for _, source := range dataSources {
		values = append(values, source.Type)
	}
	return sortedUniqueStrings(values)
}

func dcrDestinationTypes(destinations []models.DCRDestination) []string {
	values := []string{}
	for _, destination := range destinations {
		values = append(values, destination.Type)
	}
	return sortedUniqueStrings(values)
}

func dcrStreams(dataSources []models.DCRDataSource, dataFlows []models.DCRDataFlow) []string {
	values := []string{}
	for _, source := range dataSources {
		values = append(values, source.Streams...)
	}
	for _, flow := range dataFlows {
		values = append(values, flow.Streams...)
		if stringPtrValue(flow.OutputStream) != "" {
			values = append(values, stringPtrValue(flow.OutputStream))
		}
	}
	return sortedUniqueStrings(values)
}

func dcrHighSignalStreams(streams []string) []string {
	values := []string{}
	for _, stream := range streams {
		if dcrStreamSignalRank(stream) > 0 {
			values = append(values, stream)
		}
	}
	sort.SliceStable(values, func(i, j int) bool {
		leftRank := dcrStreamSignalRank(values[i])
		rightRank := dcrStreamSignalRank(values[j])
		if leftRank != rightRank {
			return leftRank > rightRank
		}
		return values[i] < values[j]
	})
	return values
}

func dcrStreamSignalRank(stream string) int {
	normalized := strings.ToLower(stream)
	switch {
	case strings.Contains(normalized, "security") || strings.Contains(normalized, "audit") || strings.Contains(normalized, "signin") || strings.Contains(normalized, "auth"):
		return 5
	case strings.Contains(normalized, "windowsevent") || strings.Contains(normalized, "event"):
		return 4
	case strings.Contains(normalized, "syslog"):
		return 3
	case strings.Contains(normalized, "process") || strings.Contains(normalized, "command"):
		return 2
	case strings.Contains(normalized, "keyvault") || strings.Contains(normalized, "secret"):
		return 2
	default:
		return 0
	}
}

func dcrRelatedIDs(ruleID string, properties map[string]any, destinations []models.DCRDestination, associations []models.DCRAssociation) []string {
	values := []string{ruleID, mapStringValue(properties, "dataCollectionEndpointId")}
	for _, destination := range destinations {
		values = append(values, stringPtrValue(destination.ResourceID))
	}
	for _, association := range associations {
		values = append(values, association.ID, association.TargetID, stringPtrValue(association.DataCollectionRuleID), stringPtrValue(association.DataCollectionEndpointID))
	}
	return sortedUniqueStrings(values)
}

func dcrSummary(asset models.DCRAsset) string {
	parts := []string{
		fmt.Sprintf("DCR %q has %d data source(s), %d data flow(s), %d destination(s), and %d association(s)", asset.Name, len(asset.DataSources), len(asset.DataFlows), len(asset.Destinations), asset.AssociationCount),
	}
	if asset.TransformationCount > 0 {
		parts = append(parts, fmt.Sprintf("%d transformation clue(s) present", asset.TransformationCount))
	}
	if len(asset.HighSignalStreams) > 0 {
		parts = append(parts, "high-signal streams: "+strings.Join(asset.HighSignalStreams, ", "))
	}
	if len(asset.DestinationTypes) > 0 {
		parts = append(parts, "destinations: "+strings.Join(asset.DestinationTypes, ", "))
	}
	return strings.Join(parts, "; ") + "."
}
