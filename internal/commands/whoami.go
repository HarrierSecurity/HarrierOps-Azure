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
			Metadata: models.WhoAmIMetadata{
				AuthMode: models.StringPtr(facts.AuthMode),
				Metadata: withArtifactContext(
					commandMetadata("whoami", now, request, facts.TenantID, facts.Subscription.ID, facts.TokenSource),
					request,
					facts.Principal,
					facts.AuthMode,
					facts.TokenSource,
				),
			},
			Principal:    facts.Principal,
			Subscription: facts.Subscription,
			TenantID:     facts.TenantID,
		}, nil
	}
}
