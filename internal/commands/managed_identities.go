package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func managedIdentitiesHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.ManagedIdentities(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		return models.ManagedIdentitiesOutput{
			Metadata:        scopedMetadata(now, request, facts.TenantID, facts.SubscriptionID, "managed-identities"),
			Identities:      facts.Identities,
			RoleAssignments: facts.RoleAssignments,
			Findings:        facts.Findings,
			Issues:          facts.Issues,
		}, nil
	}
}
