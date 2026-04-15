package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func privescHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Privesc(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		return models.PrivescOutput{
			Issues: facts.Issues,
			Metadata: models.PrincipalsMetadata{
				AuthMode:           nil,
				Command:            "privesc",
				DevOpsOrganization: models.StringPtr(request.DevOpsOrganization),
				GeneratedAt:        now().UTC().Format(time.RFC3339),
				SchemaVersion:      contracts.AzureFoxSchemaVersion,
				SubscriptionID:     models.StringPtr(facts.SubscriptionID),
				TenantID:           models.StringPtr(facts.TenantID),
				TokenSource:        nil,
			},
			Paths: append([]models.PrivescPathSummary{}, facts.Paths...),
		}, nil
	}
}
