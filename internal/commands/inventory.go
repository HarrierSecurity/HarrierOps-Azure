package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func inventoryHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Inventory(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		return models.InventoryOutput{
			Issues:             facts.Issues,
			Metadata:           commandMetadata("inventory", now, request, facts.TenantID, facts.Subscription.ID, ""),
			ResourceCount:      facts.ResourceCount,
			ResourceGroupCount: facts.ResourceGroupCount,
			Subscription:       facts.Subscription,
			TopResourceTypes:   facts.TopResourceTypes,
		}, nil
	}
}
