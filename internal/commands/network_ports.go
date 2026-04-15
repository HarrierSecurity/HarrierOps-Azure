package commands

import (
	"context"
	"sort"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func networkPortsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.NetworkPorts(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		networkPorts := append([]models.NetworkPortSummary{}, facts.NetworkPorts...)
		sort.SliceStable(networkPorts, func(i int, j int) bool {
			left := networkPorts[i]
			right := networkPorts[j]

			if exposurePriorityRank(left.ExposureConfidence) != exposurePriorityRank(right.ExposureConfidence) {
				return exposurePriorityRank(left.ExposureConfidence) < exposurePriorityRank(right.ExposureConfidence)
			}
			if left.AssetName != right.AssetName {
				return left.AssetName < right.AssetName
			}
			if left.Endpoint != right.Endpoint {
				return left.Endpoint < right.Endpoint
			}
			return left.Port < right.Port
		})

		return models.NetworkPortsOutput{
			Findings:     []models.Finding{},
			Issues:       facts.Issues,
			Metadata:     networkMetadata(now, facts.TenantID, facts.SubscriptionID, "network-ports"),
			NetworkPorts: networkPorts,
		}, nil
	}
}
