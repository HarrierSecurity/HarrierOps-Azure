package commands

import (
	"context"
	"strings"
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

		paths := append([]models.PrivescPathSummary{}, facts.Paths...)
		for idx := range paths {
			paths[idx].Target = privescArtifactTarget(paths[idx])
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
			Paths: paths,
		}, nil
	}
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
