package commands

import (
	"context"
	"sort"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func endpointsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Endpoints(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		endpoints := append([]models.EndpointSummary{}, facts.Endpoints...)
		sort.SliceStable(endpoints, func(i int, j int) bool {
			return endpointLess(endpoints[i], endpoints[j])
		})

		return models.EndpointsOutput{
			Endpoints: endpoints,
			Findings:  []models.Finding{},
			Issues:    facts.Issues,
			Metadata:  scopedMetadata(now, request, facts.TenantID, facts.SubscriptionID, "endpoints"),
		}, nil
	}
}

func endpointLess(left models.EndpointSummary, right models.EndpointSummary) bool {
	if endpointTypeRank(left.EndpointType) != endpointTypeRank(right.EndpointType) {
		return endpointTypeRank(left.EndpointType) < endpointTypeRank(right.EndpointType)
	}
	if left.SourceAssetName != right.SourceAssetName {
		return left.SourceAssetName < right.SourceAssetName
	}
	return left.Endpoint < right.Endpoint
}

func endpointTypeRank(value string) int {
	if normalizedLower(value) == "ip" {
		return 0
	}
	return 1
}
