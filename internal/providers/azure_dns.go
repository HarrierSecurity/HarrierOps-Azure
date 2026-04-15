package providers

import (
	"context"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"harrierops-azure/internal/models"
)

const (
	armSubscriptionResourcesAPIVersion = "2021-04-01"
	armPublicDNSAPIVersion             = "2018-05-01"
	armPrivateDNSAPIVersion            = "2020-06-01"
	armNetworkAPIVersion               = "2024-05-01"
)

func (provider AzureProvider) DNS(ctx context.Context, tenant string, subscription string) (DNSFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return DNSFacts{}, err
	}

	resources, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/resources",
		armSubscriptionResourcesAPIVersion,
	)
	dnsZones := []models.DnsZoneAsset{}
	issues := []models.Issue{}
	if err != nil {
		issues = append(issues, issueFromError("dns.resources", err))
		return DNSFacts{
			TenantID:       session.tenantID,
			SubscriptionID: session.subscription.ID,
			DNSZones:       dnsZones,
			Issues:         issues,
		}, nil
	}

	for _, resource := range resources {
		resourceType := strings.ToLower(mapStringValue(resource, "type"))
		switch resourceType {
		case "microsoft.network/dnszones":
			hydrated, zoneIssues := dnsHydrateResource(ctx, session.credential, resource, armPublicDNSAPIVersion)
			issues = append(issues, zoneIssues...)
			dnsZones = append(dnsZones, dnsZoneSummary(hydrated, "public"))
		case "microsoft.network/privatednszones":
			hydrated, zoneIssues := dnsHydrateResource(ctx, session.credential, resource, armPrivateDNSAPIVersion)
			issues = append(issues, zoneIssues...)
			dnsZones = append(dnsZones, dnsZoneSummary(hydrated, "private"))
		}
	}

	privateZoneReferences, referenceIssues := dnsPrivateZoneReferences(ctx, session)
	issues = append(issues, referenceIssues...)

	for index, zone := range dnsZones {
		if zone.ZoneKind != "private" {
			continue
		}
		joinKey := strings.ToLower(strings.TrimSpace(zone.ID))
		privateEndpointIDs := dedupeStrings(privateZoneReferences[joinKey])
		sort.Strings(privateEndpointIDs)
		count := len(privateEndpointIDs)
		dnsZones[index].PrivateEndpointReferenceCount = &count
		dnsZones[index].Summary = dnsZoneOperatorSummary(
			zone.Name,
			zone.ZoneKind,
			zone.RecordSetCount,
			len(zone.NameServers),
			zone.LinkedVirtualNetworkCount,
			zone.RegistrationVirtualNetworkCount,
			dnsZones[index].PrivateEndpointReferenceCount,
		)
		dnsZones[index].RelatedIDs = dedupeStrings(append(zone.RelatedIDs, privateEndpointIDs...))
	}

	return DNSFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		DNSZones:       dnsZones,
		Issues:         issues,
	}, nil
}

func dnsHydrateResource(
	ctx context.Context,
	credential azcore.TokenCredential,
	resource map[string]any,
	apiVersion string,
) (map[string]any, []models.Issue) {
	resourceID := mapStringValue(resource, "id")
	resourceType := strings.ToLower(mapStringValue(resource, "type"))
	if resourceID == "" || apiVersion == "" || !dnsResourceNeedsHydration(resource, resourceType) {
		return resource, nil
	}
	fullResource, err := armGetObject(ctx, credential, resourceID, apiVersion)
	if err != nil {
		return resource, []models.Issue{issueFromError("dns.resource["+resourceType+"/"+resourceID+"]", err)}
	}
	return fullResource, nil
}

func dnsPrivateZoneReferences(ctx context.Context, session azureSession) (map[string][]string, []models.Issue) {
	references := map[string][]string{}
	issues := []models.Issue{}

	privateEndpoints, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.Network/privateEndpoints",
		armNetworkAPIVersion,
	)
	if err != nil {
		issues = append(issues, issueFromError("dns.private_endpoints", err))
		return references, issues
	}

	for _, privateEndpoint := range privateEndpoints {
		privateEndpointID := mapStringValue(privateEndpoint, "id")
		resourceGroup, privateEndpointName := resourceGroupAndNameFromID(privateEndpointID)
		if privateEndpointID == "" || resourceGroup == "" || privateEndpointName == "" {
			continue
		}

		zoneGroups, err := armListObjects(
			ctx,
			session.credential,
			privateEndpointID+"/privateDnsZoneGroups",
			armNetworkAPIVersion,
		)
		if err != nil {
			issues = append(issues, issueFromError("dns.private_dns_zone_groups["+resourceGroup+"/"+privateEndpointName+"]", err))
			continue
		}

		for _, zoneGroup := range zoneGroups {
			for _, zoneConfig := range listValue(mapValue(zoneGroup, "properties"), "privateDnsZoneConfigs", "private_dns_zone_configs") {
				zoneID := firstNonEmpty(
					mapStringValue(zoneConfig, "privateDnsZoneId", "private_dns_zone_id"),
					mapStringValue(mapValue(zoneConfig, "properties"), "privateDnsZoneId", "private_dns_zone_id"),
				)
				zoneID = strings.ToLower(strings.TrimSpace(zoneID))
				if zoneID == "" {
					continue
				}
				references[zoneID] = append(references[zoneID], privateEndpointID)
			}
		}
	}

	for zoneID, ids := range references {
		references[zoneID] = dedupeStrings(ids)
	}
	return references, issues
}

func dnsZoneSummary(resource map[string]any, zoneKind string) models.DnsZoneAsset {
	properties := mapValue(resource, "properties")
	nameServers := dedupeStrings(dnsStringList(properties, "nameServers", "name_servers"))

	return models.DnsZoneAsset{
		ID:                              firstNonEmpty(mapStringValue(resource, "id"), "/unknown/"+firstNonEmpty(mapStringValue(resource, "name"), "unknown")),
		Name:                            firstNonEmpty(mapStringValue(resource, "name"), "unknown"),
		ResourceGroup:                   resourceGroupFromID(mapStringValue(resource, "id")),
		Location:                        stringPtr(mapStringValue(resource, "location")),
		ZoneKind:                        zoneKind,
		RecordSetCount:                  dnsOptionalInt(properties, "numberOfRecordSets", "number_of_record_sets"),
		MaxRecordSetCount:               dnsOptionalInt(properties, "maxNumberOfRecordSets", "max_number_of_record_sets"),
		NameServers:                     nameServers,
		LinkedVirtualNetworkCount:       dnsOptionalInt(properties, "numberOfVirtualNetworkLinks", "number_of_virtual_network_links"),
		RegistrationVirtualNetworkCount: dnsOptionalInt(properties, "numberOfVirtualNetworkLinksWithRegistration", "number_of_virtual_network_links_with_registration"),
		PrivateEndpointReferenceCount:   nil,
		Summary: dnsZoneOperatorSummary(
			firstNonEmpty(mapStringValue(resource, "name"), "unknown"),
			zoneKind,
			dnsOptionalInt(properties, "numberOfRecordSets", "number_of_record_sets"),
			len(nameServers),
			dnsOptionalInt(properties, "numberOfVirtualNetworkLinks", "number_of_virtual_network_links"),
			dnsOptionalInt(properties, "numberOfVirtualNetworkLinksWithRegistration", "number_of_virtual_network_links_with_registration"),
			nil,
		),
		RelatedIDs: dedupeStrings([]string{mapStringValue(resource, "id")}),
	}
}

func dnsZoneOperatorSummary(
	zoneName string,
	zoneKind string,
	recordSetCount *int,
	nameServerCount int,
	linkedVirtualNetworkCount *int,
	registrationVirtualNetworkCount *int,
	privateEndpointReferenceCount *int,
) string {
	inventoryPhrase := "does not expose a readable record-set total from the current read path"
	if recordSetCount != nil {
		inventoryPhrase = "shows " + dnsIntText(*recordSetCount) + " visible record set(s)"
	}

	if zoneKind == "public" {
		namespacePhrase := "does not expose readable name server delegation details"
		if nameServerCount > 0 {
			namespacePhrase = "delegates authority through " + dnsIntText(nameServerCount) + " visible Azure name server(s)"
		}
		return "Public DNS zone '" + zoneName + "' " + inventoryPhrase + " and " + namespacePhrase + "."
	}

	parts := []string{}
	if linkedVirtualNetworkCount != nil {
		parts = append(parts, dnsIntText(*linkedVirtualNetworkCount)+" virtual network link(s)")
	}
	if registrationVirtualNetworkCount != nil {
		parts = append(parts, dnsIntText(*registrationVirtualNetworkCount)+" registration-enabled link(s)")
	}
	if privateEndpointReferenceCount != nil {
		parts = append(parts, dnsIntText(*privateEndpointReferenceCount)+" visible private endpoint reference(s)")
	}
	namespacePhrase := "does not expose readable virtual network link counts"
	if len(parts) > 0 {
		namespacePhrase = "tracks " + strings.Join(parts, ", ")
	}
	return "Private DNS zone '" + zoneName + "' " + inventoryPhrase + " and " + namespacePhrase + "."
}

func dnsResourceNeedsHydration(resource map[string]any, resourceType string) bool {
	properties := mapValue(resource, "properties")
	recordSetCount := dnsOptionalInt(properties, "numberOfRecordSets", "number_of_record_sets")
	if recordSetCount == nil {
		return true
	}
	if resourceType == "microsoft.network/dnszones" {
		return len(dnsStringList(properties, "nameServers", "name_servers")) == 0
	}
	if resourceType == "microsoft.network/privatednszones" {
		return dnsOptionalInt(properties, "numberOfVirtualNetworkLinks", "number_of_virtual_network_links") == nil ||
			dnsOptionalInt(properties, "numberOfVirtualNetworkLinksWithRegistration", "number_of_virtual_network_links_with_registration") == nil
	}
	return false
}

func dnsStringList(input map[string]any, keys ...string) []string {
	values := []string{}
	for _, raw := range listValue(input, keys...) {
		text := stringValue(raw)
		if strings.TrimSpace(text) != "" {
			values = append(values, text)
		}
	}
	values = dedupeStrings(values)
	sort.Strings(values)
	return values
}

func dnsOptionalInt(input map[string]any, keys ...string) *int {
	for _, key := range keys {
		if _, exists := input[key]; exists {
			value := mapIntValue(input, key)
			return &value
		}
	}
	return nil
}

func dnsIntText(value int) string {
	return stringValue(value)
}
