package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func appCredentialsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.AppCredentials(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		rows := append([]models.AppCredentialSummary{}, facts.AppCredentials...)
		models.SortAppCredentialRows(rows)

		return models.AppCredentialsOutput{
			Metadata:       scopedMetadata(now, request, facts.TenantID, facts.SubscriptionID, "app-credentials"),
			AppCredentials: rows,
			Issues:         facts.Issues,
		}, nil
	}
}
