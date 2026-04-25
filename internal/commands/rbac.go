package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func rbacHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.RBAC(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		subscriptionID := facts.SubscriptionID
		if len(facts.Scopes) > 0 {
			subscriptionID = facts.Scopes[0].ID
		}
		if subscriptionID == "" && request.Subscription != "" {
			subscriptionID = request.Subscription
		}

		return models.RbacOutput{
			Issues: facts.Issues,
			Metadata: withArtifactContext(
				commandMetadata("rbac", now, request, facts.TenantID, subscriptionIDForMetadata(subscriptionID), facts.TokenSource),
				request,
				facts.CurrentPrincipal,
				facts.AuthMode,
				facts.TokenSource,
			),
			Principals:      facts.Principals,
			RoleAssignments: facts.RoleAssignments,
			Scopes:          facts.Scopes,
		}, nil
	}
}

func subscriptionIDForMetadata(scopeOrSubscriptionID string) string {
	const prefix = "/subscriptions/"
	if len(scopeOrSubscriptionID) > len(prefix) && scopeOrSubscriptionID[:len(prefix)] == prefix {
		return scopeOrSubscriptionID[len(prefix):]
	}
	return scopeOrSubscriptionID
}
