package commands

import (
	"context"
	"strings"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

type privescSourceProvider interface {
	PrivescFromSources(context.Context, providers.PermissionsFacts, providers.PrincipalsFacts, providers.ManagedIdentitiesFacts, providers.VMsFacts) (providers.PrivescFacts, error)
}

func privescHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, sessionArtifacts, err := privescFacts(ctx, request, provider, now)
		if err != nil {
			return nil, err
		}

		paths := append([]models.PrivescPathSummary{}, facts.Paths...)
		for idx := range paths {
			paths[idx].Target = privescArtifactTarget(paths[idx])
		}

		return models.PrivescOutput{
			Issues: facts.Issues,
			Metadata: withPrincipalsSessionArtifacts(models.PrincipalsMetadata{
				AuthMode:           nil,
				Command:            "privesc",
				DevOpsOrganization: models.StringPtr(request.DevOpsOrganization),
				GeneratedAt:        now().UTC().Format(time.RFC3339),
				SchemaVersion:      contracts.AzureFoxSchemaVersion,
				SubscriptionID:     models.StringPtr(facts.SubscriptionID),
				TenantID:           models.StringPtr(facts.TenantID),
				TokenSource:        nil,
			}, sessionArtifacts),
			Paths: paths,
		}, nil
	}
}

func privescFacts(ctx context.Context, request Request, provider providers.Provider, now func() time.Time) (providers.PrivescFacts, []models.SessionArtifact, error) {
	sourceProvider, ok := provider.(privescSourceProvider)
	if !ok {
		facts, err := provider.Privesc(ctx, request.Tenant, request.Subscription)
		return facts, nil, err
	}

	group := newCommandOutputGroup(4)
	expected := helperArtifactExpectedSessions(ctx, request, provider, now, "permissions", "principals", "managed-identities", "vms")
	permissionsFuture := runHelperOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions", expected)
	principalsFuture := runHelperOutput[models.PrincipalsOutput](group, ctx, request, principalsHandler(provider, now), "principals", expected)
	managedIdentitiesFuture := runHelperOutput[models.ManagedIdentitiesOutput](group, ctx, request, managedIdentitiesHandler(provider, now), "managed-identities", expected)
	vmsFuture := runHelperOutput[models.VmsOutput](group, ctx, request, vmsHandler(provider, now), "vms", expected)

	permissions, permissionsSource, err := permissionsFuture.waitWithSource()
	if err != nil {
		return providers.PrivescFacts{}, nil, err
	}
	principals, principalsSource, err := principalsFuture.waitWithSource()
	if err != nil {
		return providers.PrivescFacts{}, nil, err
	}
	managedIdentities, managedIdentitiesSource, err := managedIdentitiesFuture.waitWithSource()
	if err != nil {
		return providers.PrivescFacts{}, nil, err
	}
	vms, vmsSource, err := vmsFuture.waitWithSource()
	if err != nil {
		return providers.PrivescFacts{}, nil, err
	}

	facts, err := sourceProvider.PrivescFromSources(
		ctx,
		permissionsFactsFromOutput(permissions),
		principalsFactsFromOutput(principals),
		managedIdentitiesFactsFromOutput(managedIdentities),
		vmsFactsFromOutput(vms),
	)
	if err != nil {
		return providers.PrivescFacts{}, nil, err
	}

	sessionArtifacts := []models.SessionArtifact{}
	if permissionsSource != nil {
		sessionArtifacts = append(sessionArtifacts, *permissionsSource)
	}
	if principalsSource != nil {
		sessionArtifacts = append(sessionArtifacts, *principalsSource)
	}
	if managedIdentitiesSource != nil {
		sessionArtifacts = append(sessionArtifacts, *managedIdentitiesSource)
	}
	if vmsSource != nil {
		sessionArtifacts = append(sessionArtifacts, *vmsSource)
	}
	return facts, sessionArtifacts, nil
}

func permissionsFactsFromOutput(output models.PermissionsOutput) providers.PermissionsFacts {
	identity := artifactIdentityFactsFromScopedMetadata(output.Metadata)
	return providers.PermissionsFacts{
		TenantID:         stringPtrValue(output.Metadata.TenantID),
		SubscriptionID:   stringPtrValue(output.Metadata.SubscriptionID),
		CurrentPrincipal: identity.CurrentPrincipal,
		TokenSource:      identity.TokenSource,
		AuthMode:         identity.AuthMode,
		Permissions:      permissionFactsFromRows(output.Permissions),
		Principals:       permissionPrincipalFactsFromRows(output.Permissions),
		Issues:           append([]models.Issue{}, output.Issues...),
	}
}

func principalsFactsFromOutput(output models.PrincipalsOutput) providers.PrincipalsFacts {
	identity := artifactIdentityFactsFromPrincipalsMetadata(output.Metadata)
	return providers.PrincipalsFacts{
		TenantID:         stringPtrValue(output.Metadata.TenantID),
		SubscriptionID:   stringPtrValue(output.Metadata.SubscriptionID),
		CurrentPrincipal: identity.CurrentPrincipal,
		TokenSource:      identity.TokenSource,
		AuthMode:         identity.AuthMode,
		Principals:       append([]models.PrincipalSummary{}, output.Principals...),
		Issues:           append([]models.Issue{}, output.Issues...),
	}
}

func vmsFactsFromOutput(output models.VmsOutput) providers.VMsFacts {
	identity := artifactIdentityFactsFromMetadata(output.Metadata)
	return providers.VMsFacts{
		ArtifactIdentityFacts: identity,
		TenantID:              stringPtrValue(output.Metadata.TenantID),
		SubscriptionID:        stringPtrValue(output.Metadata.SubscriptionID),
		VMAssets:              append([]models.VmAsset{}, output.VMAssets...),
		Issues:                append([]models.Issue{}, output.Issues...),
	}
}

func permissionFactsFromRows(rows []models.PermissionRow) []providers.PermissionFact {
	facts := make([]providers.PermissionFact, 0, len(rows))
	for _, row := range rows {
		facts = append(facts, providers.PermissionFact{
			PrincipalID:         row.PrincipalID,
			DisplayName:         row.DisplayName,
			PrincipalType:       row.PrincipalType,
			HighImpactRoles:     append([]string{}, row.HighImpactRoles...),
			AllRoleNames:        append([]string{}, row.AllRoleNames...),
			RoleAssignmentCount: row.RoleAssignmentCount,
			ScopeCount:          row.ScopeCount,
			ScopeIDs:            append([]string{}, row.ScopeIDs...),
			Privileged:          row.Privileged,
			IsCurrentIdentity:   row.IsCurrentIdentity,
		})
	}
	return facts
}

func permissionPrincipalFactsFromRows(rows []models.PermissionRow) []providers.PermissionPrincipalFact {
	facts := make([]providers.PermissionPrincipalFact, 0, len(rows))
	for _, row := range rows {
		facts = append(facts, providers.PermissionPrincipalFact{
			ID: row.PrincipalID,
		})
	}
	return facts
}

func artifactIdentityFactsFromPrincipalsMetadata(metadata models.PrincipalsMetadata) providers.ArtifactIdentityFacts {
	return artifactIdentityFactsFromContext(metadata.ArtifactContext, metadata.AuthMode, metadata.TokenSource)
}

func privescArtifactTarget(path models.PrivescPathSummary) string {
	if path.CurrentIdentity {
		return "current foothold (" + privescArtifactPrincipalType(path.PrincipalType) + ")"
	}
	principal := strings.TrimSpace(path.Principal)
	asset := strings.TrimSpace(optionalString(path.Asset))
	principalType := privescArtifactPrincipalType(path.PrincipalType)
	if principal != "" && asset != "" {
		return principalType + " " + principal + " via " + asset
	}
	if principal != "" {
		return principalType + " " + principal
	}
	if asset != "" {
		return asset
	}
	return "-"
}

func privescArtifactPrincipalType(principalType string) string {
	switch strings.TrimSpace(principalType) {
	case "ManagedIdentity":
		return "managed identity"
	case "ServicePrincipal":
		return "service principal"
	case "User":
		return "user"
	default:
		normalized := strings.TrimSpace(principalType)
		if normalized == "" {
			return "principal"
		}
		return strings.ToLower(normalized)
	}
}
