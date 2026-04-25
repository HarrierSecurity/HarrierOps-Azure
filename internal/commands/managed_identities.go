package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

type managedIdentitiesSourceProvider interface {
	ManagedIdentitiesFromSources(context.Context, string, string, *providers.RBACFacts) (providers.ManagedIdentitiesFacts, error)
}

func managedIdentitiesHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, sessionArtifacts, err := managedIdentitiesFacts(ctx, request, provider, now)
		if err != nil {
			return nil, err
		}

		return models.ManagedIdentitiesOutput{
			Metadata:        withSessionArtifacts(withScopedArtifactContext(scopedMetadata(now, request, facts.TenantID, facts.SubscriptionID, "managed-identities"), request, facts.CurrentPrincipal, facts.AuthMode, facts.TokenSource), sessionArtifacts),
			Identities:      facts.Identities,
			RoleAssignments: facts.RoleAssignments,
			Findings:        facts.Findings,
			Issues:          facts.Issues,
		}, nil
	}
}

func managedIdentitiesFacts(ctx context.Context, request Request, provider providers.Provider, now func() time.Time) (providers.ManagedIdentitiesFacts, []models.SessionArtifact, error) {
	sourceProvider, ok := provider.(managedIdentitiesSourceProvider)
	if !ok {
		facts, err := provider.ManagedIdentities(ctx, request.Tenant, request.Subscription)
		return facts, nil, err
	}

	group := newCommandOutputGroup(1)
	expected := helperArtifactExpectedSessions(ctx, request, provider, now, "rbac")
	rbacFuture := runHelperOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac", expected)

	rbac, rbacSource, err := rbacFuture.waitWithSource()
	if err != nil {
		return providers.ManagedIdentitiesFacts{}, nil, err
	}
	rbacFacts := rbacFactsFromOutput(rbac)
	facts, err := sourceProvider.ManagedIdentitiesFromSources(ctx, request.Tenant, request.Subscription, &rbacFacts)
	if err != nil {
		return providers.ManagedIdentitiesFacts{}, nil, err
	}
	return facts, appendSessionArtifact(nil, rbacSource), nil
}

func rbacFactsFromOutput(output models.RbacOutput) providers.RBACFacts {
	identity := artifactIdentityFactsFromMetadata(output.Metadata)
	return providers.RBACFacts{
		TenantID:         stringPtrValue(output.Metadata.TenantID),
		SubscriptionID:   stringPtrValue(output.Metadata.SubscriptionID),
		CurrentPrincipal: identity.CurrentPrincipal,
		TokenSource:      identity.TokenSource,
		AuthMode:         identity.AuthMode,
		Principals:       append([]models.Principal{}, output.Principals...),
		Scopes:           append([]models.ScopeRef{}, output.Scopes...),
		RoleAssignments:  append([]models.RoleAssignment{}, output.RoleAssignments...),
		Issues:           append([]models.Issue{}, output.Issues...),
	}
}
