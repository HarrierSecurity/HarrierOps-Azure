package commands

import (
	"context"
	"sort"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func appInsightsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.AppInsights(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}
		components := append([]models.AppInsightsComponent{}, facts.Components...)
		targets := append([]models.AppInsightsAppTarget{}, facts.Targets...)
		sort.SliceStable(targets, func(i, j int) bool {
			if appInsightsCommandRank(targets[i]) != appInsightsCommandRank(targets[j]) {
				return appInsightsCommandRank(targets[i]) > appInsightsCommandRank(targets[j])
			}
			return targets[i].Name < targets[j].Name
		})
		return models.AppInsightsOutput{
			Components: components,
			Targets:    targets,
			Findings:   []models.Finding{},
			Issues:     facts.Issues,
			Metadata:   runtimeCommandMetadata("appinsights", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func appInsightsCommandRank(target models.AppInsightsAppTarget) int {
	return len(target.FilteringClues)*3 + len(target.SamplingClues)*2 + len(target.LoggingLevelClues) + len(target.InstrumentationClues)
}
