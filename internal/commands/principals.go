package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

type principalsSourceProvider interface {
	PrincipalsFromSources(context.Context, string, string, providers.RBACFacts, providers.WhoAmIFacts, providers.ManagedIdentitiesFacts) (providers.PrincipalsFacts, error)
}

func principalsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, sessionArtifacts, err := principalsFacts(ctx, request, provider, now)
		if err != nil {
			return nil, err
		}

		subscriptionID := request.Subscription
		if subscriptionID == "" {
			subscriptionID = facts.SubscriptionID
		}

		return models.PrincipalsOutput{
			Issues: facts.Issues,
			Metadata: withPrincipalsSessionArtifacts(withPrincipalsArtifactContext(models.PrincipalsMetadata{
				AuthMode:           models.StringPtr(facts.AuthMode),
				Command:            "principals",
				DevOpsOrganization: models.StringPtr(request.DevOpsOrganization),
				GeneratedAt:        now().UTC().Format(time.RFC3339),
				SchemaVersion:      contracts.AzureFoxSchemaVersion,
				SubscriptionID:     models.StringPtr(subscriptionID),
				TenantID:           models.StringPtr(facts.TenantID),
				TokenSource:        models.StringPtr(facts.TokenSource),
			}, request, facts.CurrentPrincipal, facts.AuthMode, facts.TokenSource), sessionArtifacts),
			Principals: append([]models.PrincipalSummary{}, facts.Principals...),
		}, nil
	}
}

func principalsFacts(ctx context.Context, request Request, provider providers.Provider, now func() time.Time) (providers.PrincipalsFacts, []models.SessionArtifact, error) {
	sourceProvider, ok := provider.(principalsSourceProvider)
	if !ok {
		facts, err := provider.Principals(ctx, request.Tenant, request.Subscription)
		return facts, nil, err
	}

	group := newCommandOutputGroup(3)
	expected := helperArtifactExpectedSessions(ctx, request, provider, now, "rbac", "whoami", "managed-identities")
	rbacFuture := runHelperOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac", expected)
	whoamiFuture := runHelperOutput[models.WhoAmIOutput](group, ctx, request, whoAmIHandler(provider, now), "whoami", expected)
	managedIdentityFuture := runHelperOutput[models.ManagedIdentitiesOutput](group, ctx, request, managedIdentitiesHandler(provider, now), "managed-identities", expected)

	rbac, rbacSource, err := rbacFuture.waitWithSource()
	if err != nil {
		return providers.PrincipalsFacts{}, nil, err
	}
	whoami, whoamiSource, err := whoamiFuture.waitWithSource()
	if err != nil {
		return providers.PrincipalsFacts{}, nil, err
	}
	managedIdentities, managedIdentitiesSource, err := managedIdentityFuture.waitWithSource()
	if err != nil {
		return providers.PrincipalsFacts{}, nil, err
	}

	rbacFacts := rbacFactsFromOutput(rbac)
	whoamiFacts := whoAmIFactsFromOutput(whoami)
	managedIdentityFacts := managedIdentitiesFactsFromOutput(managedIdentities)
	tenantID := firstNonEmpty(rbacFacts.TenantID, whoamiFacts.TenantID, managedIdentityFacts.TenantID)
	subscriptionID := firstNonEmpty(rbacFacts.SubscriptionID, whoamiFacts.Subscription.ID, managedIdentityFacts.SubscriptionID)
	facts, err := sourceProvider.PrincipalsFromSources(
		ctx,
		tenantID,
		subscriptionID,
		rbacFacts,
		whoamiFacts,
		managedIdentityFacts,
	)
	if err != nil {
		return providers.PrincipalsFacts{}, nil, err
	}

	sessionArtifacts := []models.SessionArtifact{}
	if rbacSource != nil {
		sessionArtifacts = append(sessionArtifacts, *rbacSource)
	}
	if whoamiSource != nil {
		sessionArtifacts = append(sessionArtifacts, *whoamiSource)
	}
	if managedIdentitiesSource != nil {
		sessionArtifacts = append(sessionArtifacts, *managedIdentitiesSource)
	}
	return facts, sessionArtifacts, nil
}

func whoAmIFactsFromOutput(output models.WhoAmIOutput) providers.WhoAmIFacts {
	return providers.WhoAmIFacts{
		TenantID:        output.TenantID,
		Subscription:    output.Subscription,
		Principal:       output.Principal,
		EffectiveScopes: append([]models.ScopeRef{}, output.EffectiveScopes...),
		TokenSource:     stringPtrValue(output.Metadata.TokenSource),
		AuthMode:        stringPtrValue(output.Metadata.AuthMode),
		Issues:          append([]models.Issue{}, output.Issues...),
	}
}

func managedIdentitiesFactsFromOutput(output models.ManagedIdentitiesOutput) providers.ManagedIdentitiesFacts {
	identity := artifactIdentityFactsFromScopedMetadata(output.Metadata)
	return providers.ManagedIdentitiesFacts{
		ArtifactIdentityFacts: identity,
		TenantID:              stringPtrValue(output.Metadata.TenantID),
		SubscriptionID:        stringPtrValue(output.Metadata.SubscriptionID),
		Identities:            append([]models.ManagedIdentity{}, output.Identities...),
		RoleAssignments:       append([]models.ManagedIdentityRoleAssignment{}, output.RoleAssignments...),
		Findings:              append([]models.ManagedIdentityFinding{}, output.Findings...),
		Issues:                append([]models.Issue{}, output.Issues...),
	}
}

func artifactIdentityFactsFromScopedMetadata(metadata models.ScopedCommandMetadata) providers.ArtifactIdentityFacts {
	return artifactIdentityFactsFromContext(metadata.ArtifactContext, metadata.AuthMode, metadata.TokenSource)
}

func withPrincipalsSessionArtifacts(metadata models.PrincipalsMetadata, artifacts []models.SessionArtifact) models.PrincipalsMetadata {
	if len(artifacts) == 0 {
		return metadata
	}
	metadata.SessionArtifacts = append([]models.SessionArtifact{}, artifacts...)
	return metadata
}
