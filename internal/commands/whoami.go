package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func whoAmIHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.WhoAmI(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		return models.WhoAmIOutput{
			EffectiveScopes: facts.EffectiveScopes,
			Issues:          facts.Issues,
			Metadata:        whoAmIMetadata(now, request, facts.TenantID, facts.Subscription.ID, facts.TokenSource, facts.AuthMode),
			Principal:       facts.Principal,
			Subscription:    facts.Subscription,
			TenantID:        facts.TenantID,
		}, nil
	}
}
