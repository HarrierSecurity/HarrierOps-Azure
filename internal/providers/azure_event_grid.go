package providers

import (
	"context"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

const armEventGridAPIVersion = "2025-02-15"

func (provider AzureProvider) EventGrid(ctx context.Context, tenant string, subscription string) (EventGridFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return EventGridFacts{}, err
	}

	subscriptionPath := "/subscriptions/" + session.subscription.ID + "/providers/Microsoft.EventGrid/eventSubscriptions"
	items, listErr := armListObjects(
		ctx,
		session.credential,
		subscriptionPath,
		armEventGridAPIVersion,
	)

	routes := []models.EventGridRouteAsset{}
	issues := []models.Issue{}
	if listErr != nil {
		issues = append(issues, issueFromError("event-grid.event-subscriptions", listErr))
	}
	scopedItems, scopedIssues := eventGridTopicTypeItems(ctx, session)
	items = append(items, scopedItems...)
	issues = append(issues, scopedIssues...)

	seen := map[string]struct{}{}
	for _, item := range items {
		route := eventGridRouteAsset(item)
		if _, exists := seen[route.ID]; exists {
			continue
		}
		seen[route.ID] = struct{}{}
		routes = append(routes, route)
	}

	return EventGridFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		Routes:         routes,
		Issues:         issues,
	}, nil
}

func eventGridRouteAsset(item map[string]any) models.EventGridRouteAsset {
	routeID := mapStringValue(item, "id")
	properties := mapValue(item, "properties")
	destinationContainer := mapValue(properties, "deliveryWithResourceIdentity")
	destination := mapValue(properties, "destination")
	if len(destination) == 0 {
		destination = mapValue(destinationContainer, "destination")
	}

	destinationType := firstNonEmpty(
		mapStringValue(destination, "endpointType", "endpoint_type"),
		mapStringValue(destination, "endpointtype"),
		"unknown",
	)
	destinationTargetID := eventGridDestinationTargetID(destination)
	identityType, identityID := eventGridIdentityContext(destinationContainer, properties)
	sourceID := eventGridSourceID(routeID)
	sourceType := eventGridSourceType(sourceID)
	externalDelivery := strings.EqualFold(destinationType, "WebHook")
	classification := eventGridClassification(destinationType)

	return models.EventGridRouteAsset{
		ID:                  firstNonEmpty(routeID, "/unknown/event-grid-route"),
		Name:                firstNonEmpty(mapStringValue(item, "name"), resourceNameFromID(routeID), "unknown"),
		DestinationType:     destinationType,
		Classification:      classification,
		SourceID:            sourceID,
		SourceType:          sourceType,
		DestinationTargetID: destinationTargetID,
		ExternalDelivery:    externalDelivery,
		ProvisioningState:   stringPtr(mapStringValue(properties, "provisioningState", "provisioning_state")),
		IdentityType:        identityType,
		IdentityID:          identityID,
		EventDeliverySchema: stringPtr(mapStringValue(properties, "eventDeliverySchema", "event_delivery_schema")),
		IncludedEventTypes:  eventGridIncludedEventTypes(properties),
		Summary: eventGridOperatorSummary(
			sourceType,
			destinationType,
			classification,
			externalDelivery,
			destinationTargetID,
			identityType,
		),
		RelatedIDs: eventGridRelatedIDs(sourceID, destinationTargetID, identityID),
	}
}

func eventGridTopicTypeItems(ctx context.Context, session azureSession) ([]map[string]any, []models.Issue) {
	topicTypes, err := armListObjects(
		ctx,
		session.credential,
		"/providers/Microsoft.EventGrid/topicTypes",
		armEventGridAPIVersion,
	)
	if err != nil {
		return []map[string]any{}, []models.Issue{issueFromError("event-grid.topic-types", err)}
	}

	locations, locationIssues := eventGridSubscriptionLocations(ctx, session)
	scopes := eventGridEnumerationScopes(session.subscription.ID, topicTypes, locations)
	items, scopeIssues := eventGridItemsFromScopes(ctx, scopes, func(ctx context.Context, path string) ([]map[string]any, error) {
		return armListObjects(ctx, session.credential, path, armEventGridAPIVersion)
	})

	return items, append(locationIssues, scopeIssues...)
}

func eventGridSubscriptionLocations(ctx context.Context, session azureSession) ([]string, []models.Issue) {
	locations := []string{}
	resourceGroupPager := session.clientFactory.NewResourceGroupsClient().NewListPager(nil)
	for resourceGroupPager.More() {
		page, err := resourceGroupPager.NextPage(ctx)
		if err != nil {
			return dedupeStrings(locations), []models.Issue{issueFromError("event-grid.resource-groups", err)}
		}
		for _, group := range page.Value {
			location := eventGridCanonicalLocation(stringValue(group.Location))
			if location == "" {
				continue
			}
			locations = append(locations, location)
		}
	}

	return dedupeStrings(locations), []models.Issue{}
}

func eventGridEnumerationScopes(subscriptionID string, topicTypes []map[string]any, locations []string) []eventGridEnumerationScope {
	scopes := []eventGridEnumerationScope{}
	for _, topicType := range topicTypes {
		name := strings.TrimSpace(mapStringValue(topicType, "name"))
		if name == "" {
			continue
		}
		properties := mapValue(topicType, "properties")
		if eventGridTopicTypeSupportsGlobalEnumeration(properties) {
			scopes = append(scopes, eventGridEnumerationScope{
				path:       "/subscriptions/" + subscriptionID + "/providers/Microsoft.EventGrid/topicTypes/" + name + "/eventSubscriptions",
				issueScope: "event-grid.topic-type[" + name + "]",
			})
		}

		if !eventGridTopicTypeSupportsRegionalEnumeration(properties) {
			continue
		}

		for _, location := range eventGridTopicTypeLocations(properties, locations) {
			scopes = append(scopes, eventGridEnumerationScope{
				path:       "/subscriptions/" + subscriptionID + "/providers/Microsoft.EventGrid/locations/" + location + "/topicTypes/" + name + "/eventSubscriptions",
				issueScope: "event-grid.topic-type[" + name + "@" + location + "]",
			})
		}
	}
	return scopes
}

func eventGridTopicTypeSupportsGlobalEnumeration(properties map[string]any) bool {
	if strings.EqualFold(mapStringValue(properties, "resourceRegionType", "resource_region_type"), "GlobalResource") {
		return true
	}
	if value := optionalBoolPtr(properties, "areRegionalAndGlobalSourcesSupported", "are_regional_and_global_sources_supported"); value != nil && *value {
		return true
	}
	for _, value := range listValue(properties, "supportedScopesForSource", "supported_scopes_for_source") {
		if strings.EqualFold(strings.TrimSpace(stringValue(value)), "AzureSubscription") {
			return true
		}
	}
	return false
}

func eventGridTopicTypeSupportsRegionalEnumeration(properties map[string]any) bool {
	if strings.EqualFold(mapStringValue(properties, "resourceRegionType", "resource_region_type"), "RegionalResource") {
		return true
	}
	for _, value := range listValue(properties, "supportedScopesForSource", "supported_scopes_for_source") {
		if strings.EqualFold(strings.TrimSpace(stringValue(value)), "Resource") ||
			strings.EqualFold(strings.TrimSpace(stringValue(value)), "ResourceGroup") {
			return true
		}
	}
	return false
}

func eventGridTopicTypeLocations(properties map[string]any, subscriptionLocations []string) []string {
	supported := []string{}
	for _, value := range listValue(properties, "supportedLocations", "supported_locations") {
		location := eventGridCanonicalLocation(stringValue(value))
		if location == "" {
			continue
		}
		supported = append(supported, location)
	}
	supported = dedupeStrings(supported)
	if len(supported) == 0 {
		return dedupeStrings(subscriptionLocations)
	}
	if len(subscriptionLocations) == 0 {
		return supported
	}

	available := map[string]struct{}{}
	for _, location := range supported {
		available[location] = struct{}{}
	}

	out := []string{}
	for _, location := range dedupeStrings(subscriptionLocations) {
		if _, exists := available[location]; exists {
			out = append(out, location)
		}
	}
	return out
}

func eventGridCanonicalLocation(location string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(location)), " ", "")
}

type eventGridEnumerationScope struct {
	path       string
	issueScope string
}

func eventGridItemsFromScopes(
	ctx context.Context,
	scopes []eventGridEnumerationScope,
	listFn func(context.Context, string) ([]map[string]any, error),
) ([]map[string]any, []models.Issue) {
	items := []map[string]any{}
	issues := []models.Issue{}
	seen := map[string]struct{}{}

	for _, scope := range scopes {
		rows, err := listFn(ctx, scope.path)
		if err != nil {
			if eventGridIgnoreEnumerationError(err) {
				continue
			}
			issues = append(issues, issueFromError(scope.issueScope, err))
			continue
		}
		for _, row := range rows {
			routeID := mapStringValue(row, "id")
			if routeID == "" {
				items = append(items, row)
				continue
			}
			if _, exists := seen[routeID]; exists {
				continue
			}
			seen[routeID] = struct{}{}
			items = append(items, row)
		}
	}

	return items, issues
}

func eventGridIgnoreEnumerationError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, ": 404") ||
		strings.Contains(message, "resourcenotfound") ||
		strings.Contains(message, "invalidresourcetype") ||
		strings.Contains(message, "noregisteredproviderfound")
}

func eventGridDestinationTargetID(destination map[string]any) *string {
	properties := mapValue(destination, "properties")
	for _, key := range []string{"resourceId", "resource_id", "userAssignedIdentity", "user_assigned_identity"} {
		if value := mapStringValue(properties, key); value != "" {
			return models.StringPtr(value)
		}
	}
	if value := mapStringValue(destination, "resourceId", "resource_id"); value != "" {
		return models.StringPtr(value)
	}
	return nil
}

func eventGridIdentityContext(deliveryWithIdentity map[string]any, properties map[string]any) (*string, *string) {
	identity := mapValue(deliveryWithIdentity, "identity")
	if len(identity) == 0 {
		identity = mapValue(properties, "identity")
	}
	if len(identity) == 0 {
		return nil, nil
	}

	identityType := stringPtr(mapStringValue(identity, "type"))
	identityID := stringPtr(firstNonEmpty(
		mapStringValue(identity, "userAssignedIdentity", "user_assigned_identity"),
		mapStringValue(identity, "userAssignedIdentityResourceId", "user_assigned_identity_resource_id"),
	))
	return identityType, identityID
}

func eventGridIncludedEventTypes(properties map[string]any) []string {
	filter := mapValue(properties, "filter")
	values := listValue(filter, "includedEventTypes", "included_event_types")
	if len(values) == 0 {
		return []string{"All"}
	}

	out := make([]string, 0, len(values))
	for _, value := range values {
		text := strings.TrimSpace(stringValue(value))
		if text == "" {
			continue
		}
		out = append(out, text)
	}
	sort.Strings(out)
	return dedupeStrings(out)
}

func eventGridSourceID(routeID string) string {
	const suffix = "/providers/Microsoft.EventGrid/eventSubscriptions/"
	index := strings.LastIndex(strings.ToLower(routeID), strings.ToLower(suffix))
	if index < 0 {
		return ""
	}
	return routeID[:index]
}

func eventGridSourceType(sourceID string) string {
	parts := armIDParts(sourceID)
	if len(parts) == 0 {
		return ""
	}

	switch {
	case len(parts) == 2 && strings.EqualFold(parts[0], "subscriptions"):
		return "subscription"
	case len(parts) == 4 && strings.EqualFold(parts[2], "resourceGroups"):
		return "resource-group"
	}

	for index := 0; index < len(parts)-2; index++ {
		if strings.EqualFold(parts[index], "providers") {
			return parts[index+1] + "/" + parts[index+2]
		}
	}
	return ""
}

func eventGridClassification(destinationType string) string {
	switch {
	case strings.EqualFold(destinationType, "AzureFunction"), strings.EqualFold(destinationType, "HybridConnection"):
		return "execution-capable"
	case strings.EqualFold(destinationType, "WebHook"):
		return "external-callback"
	default:
		return "supporting-context"
	}
}

func eventGridOperatorSummary(
	sourceType string,
	destinationType string,
	classification string,
	externalDelivery bool,
	destinationTargetID *string,
	identityType *string,
) string {
	parts := []string{}
	switch classification {
	case "execution-capable":
		parts = append(parts, "Visible Event Grid routing terminates in a destination type that can plausibly execute or trigger code or workflow behavior.")
	case "external-callback":
		parts = append(parts, "Visible Event Grid routing terminates in a webhook-style destination that crosses the normal Azure resource boundary.")
	default:
		parts = append(parts, "Visible Event Grid routing provides trigger plumbing context, but the current read path does not yet show a directly execution-capable destination.")
	}
	if sourceType != "" {
		parts = append(parts, "Source type is "+sourceType+".")
	}
	if destinationTargetID != nil && *destinationTargetID != "" && !externalDelivery {
		parts = append(parts, "Destination target resolves to "+resourceNameFromID(*destinationTargetID)+".")
	}
	if identityType != nil && *identityType != "" {
		parts = append(parts, "Delivery uses managed identity context ("+*identityType+").")
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func eventGridRelatedIDs(sourceID string, destinationTargetID *string, identityID *string) []string {
	values := []string{sourceID}
	if destinationTargetID != nil {
		values = append(values, *destinationTargetID)
	}
	if identityID != nil {
		values = append(values, *identityID)
	}
	return dedupeStrings(values)
}
