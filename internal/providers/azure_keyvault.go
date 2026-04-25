package providers

import (
	"context"
	"sort"

	"harrierops-azure/internal/models"
)

const armKeyVaultAPIVersion = "2024-11-01"

func (provider AzureProvider) KeyVault(ctx context.Context, tenant string, subscription string) (KeyVaultFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return KeyVaultFacts{}, err
	}

	vaults, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.KeyVault/vaults",
		armKeyVaultAPIVersion,
	)
	if err != nil {
		return KeyVaultFacts{
			ArtifactIdentityFacts: azureArtifactIdentityFacts(session),
			TenantID:              session.tenantID,
			SubscriptionID:        session.subscription.ID,
			KeyVaults:             []models.KeyVaultAsset{},
			Issues:                []models.Issue{issueFromError("keyvault", err)},
		}, nil
	}

	rows := make([]models.KeyVaultAsset, 0, len(vaults))
	for _, vault := range vaults {
		rows = append(rows, keyVaultSummary(vault))
	}
	sort.Slice(rows, func(i int, j int) bool {
		if rows[i].Name != rows[j].Name {
			return rows[i].Name < rows[j].Name
		}
		return rows[i].ID < rows[j].ID
	})

	return KeyVaultFacts{
		ArtifactIdentityFacts: azureArtifactIdentityFacts(session),
		TenantID:              session.tenantID,
		SubscriptionID:        session.subscription.ID,
		KeyVaults:             rows,
		Issues:                []models.Issue{},
	}, nil
}

func keyVaultSummary(vault map[string]any) models.KeyVaultAsset {
	vaultID := mapStringValue(vault, "id")
	properties := mapValue(vault, "properties")
	networkACLs := mapValue(properties, "networkAcls", "network_acls")
	privateEndpoints := listValue(properties, "privateEndpointConnections", "private_endpoint_connections")
	if len(privateEndpoints) == 0 {
		privateEndpoints = listValue(vault, "privateEndpointConnections", "private_endpoint_connections")
	}
	sku := mapValue(vault, "sku")

	return models.KeyVaultAsset{
		AccessPolicyCount:       len(listValue(properties, "accessPolicies", "access_policies")),
		EnableRBACAuthorization: mapBoolValue(properties, "enableRbacAuthorization", "enable_rbac_authorization"),
		ID:                      firstNonEmpty(vaultID, "/unknown/"+firstNonEmpty(mapStringValue(vault, "name"), "unknown")),
		Location:                stringPtr(mapStringValue(vault, "location")),
		Name:                    firstNonEmpty(mapStringValue(vault, "name"), resourceNameFromID(vaultID), "unknown"),
		NetworkDefaultAction:    stringPtr(mapStringValue(networkACLs, "defaultAction", "default_action")),
		PrivateEndpointEnabled:  len(privateEndpoints) > 0,
		PublicNetworkAccess:     stringPtr(mapStringValue(properties, "publicNetworkAccess", "public_network_access")),
		PurgeProtectionEnabled:  mapBoolValue(properties, "enablePurgeProtection", "enable_purge_protection"),
		ResourceGroup:           resourceGroupFromID(vaultID),
		SKUName:                 stringPtr(mapStringValue(sku, "name")),
		SoftDeleteEnabled:       mapBoolValue(properties, "enableSoftDelete", "enable_soft_delete"),
		TenantID:                stringPtr(mapStringValue(properties, "tenantId", "tenant_id")),
		VaultURI:                stringPtr(mapStringValue(properties, "vaultUri", "vault_uri")),
	}
}
