package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func principalsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Principals(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		subscriptionID := request.Subscription
		if subscriptionID == "" {
			subscriptionID = facts.SubscriptionID
		}

		return models.PrincipalsOutput{
			Issues: facts.Issues,
			Metadata: models.PrincipalsMetadata{
				AuthMode:           nil,
				Command:            "principals",
				DevOpsOrganization: models.StringPtr(request.DevOpsOrganization),
				GeneratedAt:        now().UTC().Format(time.RFC3339),
				SchemaVersion:      contracts.AzureFoxSchemaVersion,
				SubscriptionID:     models.StringPtr(subscriptionID),
				TenantID:           models.StringPtr(facts.TenantID),
				TokenSource:        nil,
			},
			Principals: append([]models.PrincipalSummary{}, facts.Principals...),
		}, nil
	}
}
