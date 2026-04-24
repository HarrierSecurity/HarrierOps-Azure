package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func monitoringSinksHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.MonitoringSinks(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}
		return models.MonitoringSinksOutput{
			Sinks:    sortedByLess(facts.Sinks, monitoringSinkLess),
			Findings: []models.Finding{},
			Issues:   facts.Issues,
			Metadata: runtimeCommandMetadata("monitoring-sinks", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func monitoringSinkLess(left models.MonitoringSinkAsset, right models.MonitoringSinkAsset) bool {
	if left.ReferenceCount != right.ReferenceCount {
		return left.ReferenceCount > right.ReferenceCount
	}
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	return left.Name < right.Name
}
