package commands

import (
	"context"
	"sort"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func networkEffectiveHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.NetworkEffective(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		effectiveExposures := append([]models.NetworkEffectiveSummary{}, facts.EffectiveExposures...)
		sort.SliceStable(effectiveExposures, func(i int, j int) bool {
			left := effectiveExposures[i]
			right := effectiveExposures[j]

			if exposurePriorityRank(left.EffectiveExposure) != exposurePriorityRank(right.EffectiveExposure) {
				return exposurePriorityRank(left.EffectiveExposure) < exposurePriorityRank(right.EffectiveExposure)
			}
			if left.AssetName != right.AssetName {
				return left.AssetName < right.AssetName
			}
			return left.Endpoint < right.Endpoint
		})

		return models.NetworkEffectiveOutput{
			EffectiveExposures: effectiveExposures,
			Findings:           []models.Finding{},
			Issues:             facts.Issues,
			Metadata:           scopedMetadata(now, request, facts.TenantID, facts.SubscriptionID, "network-effective"),
		}, nil
	}
}
