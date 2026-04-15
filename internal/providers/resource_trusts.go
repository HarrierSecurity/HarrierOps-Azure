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

	return ResourceTrustsFacts{
		TenantID:       firstNonEmpty(storageFacts.TenantID, keyVaultFacts.TenantID),
		SubscriptionID: firstNonEmpty(storageFacts.SubscriptionID, keyVaultFacts.SubscriptionID),
		StorageAssets:  append([]models.StorageAsset{}, storageFacts.StorageAssets...),
		KeyVaults:      append([]models.KeyVaultAsset{}, keyVaultFacts.KeyVaults...),
		Issues:         append(append([]models.Issue{}, storageFacts.Issues...), keyVaultFacts.Issues...),
	}, nil
}
