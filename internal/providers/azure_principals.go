package providers

import "context"

func (provider AzureProvider) Principals(ctx context.Context, tenant string, subscription string) (PrincipalsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return PrincipalsFacts{}, err
	}

	rbacFacts := provider.collectRBACFacts(ctx, session)
	managedIdentityFacts := provider.collectManagedIdentityFacts(ctx, tenant, subscription, session, rbacFacts)
	whoamiFacts, err := provider.WhoAmI(ctx, tenant, subscription)
	if err != nil {
		return PrincipalsFacts{}, err
	}

	return principalsFactsFromSources(session.tenantID, session.subscription.ID, rbacFacts, whoamiFacts, managedIdentityFacts), nil
}
