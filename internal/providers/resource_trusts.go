package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (provider AzureProvider) ResourceTrusts(ctx context.Context, tenant string, subscription string) (ResourceTrustsFacts, error) {
	return collectResourceTrusts(ctx, provider, tenant, subscription)
}

func (provider StaticProvider) ResourceTrusts(ctx context.Context, tenant string, subscription string) (ResourceTrustsFacts, error) {
	return collectResourceTrusts(ctx, provider, tenant, subscription)
}

func collectResourceTrusts(ctx context.Context, provider Provider, tenant string, subscription string) (ResourceTrustsFacts, error) {
	storageFacts, err := provider.Storage(ctx, tenant, subscription)
	if err != nil {
		return ResourceTrustsFacts{}, err
	}

	keyVaultFacts, err := provider.KeyVault(ctx, tenant, subscription)
	if err != nil {
		return ResourceTrustsFacts{}, err
	}

	identity, identityIssues := MergeArtifactIdentityFacts(storageFacts.ArtifactIdentityFacts, keyVaultFacts.ArtifactIdentityFacts)
	issues := append(append([]models.Issue{}, storageFacts.Issues...), keyVaultFacts.Issues...)
	issues = append(issues, identityIssues...)

	return ResourceTrustsFacts{
		ArtifactIdentityFacts: identity,
		TenantID:              firstNonEmpty(storageFacts.TenantID, keyVaultFacts.TenantID),
		SubscriptionID:        firstNonEmpty(storageFacts.SubscriptionID, keyVaultFacts.SubscriptionID),
		StorageAssets:         append([]models.StorageAsset{}, storageFacts.StorageAssets...),
		KeyVaults:             append([]models.KeyVaultAsset{}, keyVaultFacts.KeyVaults...),
		Issues:                issues,
	}, nil
}

func MergeArtifactIdentityFacts(values ...ArtifactIdentityFacts) (ArtifactIdentityFacts, []models.Issue) {
	var selected ArtifactIdentityFacts
	issues := []models.Issue{}
	for _, value := range values {
		if !artifactIdentityFactsPresent(value) {
			continue
		}
		if !artifactIdentityFactsPresent(selected) {
			selected = value
			continue
		}
		if !artifactIdentityFactsEqual(selected, value) {
			issues = append(issues, models.Issue{
				Kind:    "artifact_identity_mismatch",
				Message: "Source helper artifacts carry different identity context; resource trust provenance uses the first visible context and marks the mismatch.",
				Scope:   "resource-trusts",
				Context: map[string]string{
					"first_principal_id":  selected.CurrentPrincipal.ID,
					"second_principal_id": value.CurrentPrincipal.ID,
					"first_auth_mode":     selected.AuthMode,
					"second_auth_mode":    value.AuthMode,
					"first_token_source":  selected.TokenSource,
					"second_token_source": value.TokenSource,
				},
			})
		}
	}
	return selected, issues
}

func artifactIdentityFactsPresent(value ArtifactIdentityFacts) bool {
	return value.CurrentPrincipal.ID != "" || value.CurrentPrincipal.TenantID != "" || value.AuthMode != "" || value.TokenSource != ""
}

func artifactIdentityFactsEqual(left ArtifactIdentityFacts, right ArtifactIdentityFacts) bool {
	return left.CurrentPrincipal.ID == right.CurrentPrincipal.ID &&
		left.CurrentPrincipal.PrincipalType == right.CurrentPrincipal.PrincipalType &&
		left.CurrentPrincipal.TenantID == right.CurrentPrincipal.TenantID &&
		left.AuthMode == right.AuthMode &&
		left.TokenSource == right.TokenSource
}
