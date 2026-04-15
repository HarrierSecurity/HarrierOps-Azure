package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func workloadsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Workloads(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		workloads := append([]models.WorkloadSummary{}, facts.Workloads...)

		return models.WorkloadsOutput{
			Metadata:  scopedMetadata(now, request, facts.TenantID, facts.SubscriptionID, "workloads"),
			Workloads: workloads,
			Findings:  []models.Finding{},
			Issues:    facts.Issues,
		}, nil
	}
}
