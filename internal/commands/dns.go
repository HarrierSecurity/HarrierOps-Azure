package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func dnsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.DNS(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		zones := sortedByLess(facts.DNSZones, dnsZoneLess)

		return models.DnsOutput{
			DNSZones: zones,
			Findings: []models.Finding{},
			Issues:   facts.Issues,
			Metadata: runtimeCommandMetadata("dns", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func dnsZoneLess(left models.DnsZoneAsset, right models.DnsZoneAsset) bool {
	if (left.ZoneKind != "public") != (right.ZoneKind != "public") {
		return left.ZoneKind == "public"
	}
	if len(left.NameServers) != len(right.NameServers) {
		return len(left.NameServers) > len(right.NameServers)
	}
	if dnsIntValue(left.PrivateEndpointReferenceCount) != dnsIntValue(right.PrivateEndpointReferenceCount) {
		return dnsIntValue(left.PrivateEndpointReferenceCount) > dnsIntValue(right.PrivateEndpointReferenceCount)
	}
	if dnsIntValue(left.LinkedVirtualNetworkCount) != dnsIntValue(right.LinkedVirtualNetworkCount) {
		return dnsIntValue(left.LinkedVirtualNetworkCount) > dnsIntValue(right.LinkedVirtualNetworkCount)
	}
	if dnsIntValue(left.RegistrationVirtualNetworkCount) != dnsIntValue(right.RegistrationVirtualNetworkCount) {
		return dnsIntValue(left.RegistrationVirtualNetworkCount) > dnsIntValue(right.RegistrationVirtualNetworkCount)
	}
	if dnsIntValue(left.RecordSetCount) != dnsIntValue(right.RecordSetCount) {
		return dnsIntValue(left.RecordSetCount) > dnsIntValue(right.RecordSetCount)
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func dnsIntValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
