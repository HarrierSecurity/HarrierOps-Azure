package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func keyVaultHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.KeyVault(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		vaults := sortedByLess(facts.KeyVaults, keyVaultLess)

		return models.KeyVaultOutput{
			Findings:  keyVaultFindings(vaults),
			Issues:    facts.Issues,
			KeyVaults: vaults,
			Metadata:  commandMetadata("keyvault", now, request, facts.TenantID, facts.SubscriptionID, ""),
		}, nil
	}
}

func keyVaultLess(left models.KeyVaultAsset, right models.KeyVaultAsset) bool {
	leftExposure, leftPurge, leftAuth := keyVaultPriorityRank(left)
	rightExposure, rightPurge, rightAuth := keyVaultPriorityRank(right)

	if leftExposure != rightExposure {
		return leftExposure < rightExposure
	}
	if leftPurge != rightPurge {
		return leftPurge < rightPurge
	}
	if leftAuth != rightAuth {
		return leftAuth < rightAuth
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func keyVaultPriorityRank(item models.KeyVaultAsset) (int, int, int) {
	publicEnabled := normalizedLower(keyVaultStringValue(item.PublicNetworkAccess)) == "enabled"
	defaultAction := normalizedLower(keyVaultStringValue(item.NetworkDefaultAction))

	exposureRank := 4
	switch {
	case publicEnabled && defaultAction == "" && !item.PrivateEndpointEnabled:
		exposureRank = 0
	case publicEnabled && defaultAction == "allow":
		exposureRank = 1
	case publicEnabled && !item.PrivateEndpointEnabled:
		exposureRank = 2
	case publicEnabled:
		exposureRank = 3
	}

	purgeRank := 1
	if !item.PurgeProtectionEnabled {
		purgeRank = 0
	}
	authRank := 1
	if !item.EnableRBACAuthorization {
		authRank = 0
	}
	return exposureRank, purgeRank, authRank
}

func keyVaultFindings(vaults []models.KeyVaultAsset) []models.KeyVaultFinding {
	findings := []models.KeyVaultFinding{}

	for _, vault := range vaults {
		publicNetworkAccess := normalizedLower(keyVaultStringValue(vault.PublicNetworkAccess))
		networkDefaultAction := normalizedLower(keyVaultStringValue(vault.NetworkDefaultAction))
		implicitOpenACL := publicNetworkAccess == "enabled" && networkDefaultAction == ""

		if publicNetworkAccess == "enabled" {
			switch {
			case (networkDefaultAction == "allow" || implicitOpenACL) && !vault.PrivateEndpointEnabled:
				description := "Key Vault '" + vault.Name + "' has public network access enabled, default network action Allow, and no private endpoint visible. Review whether that secret-management surface is intentionally internet reachable."
				if implicitOpenACL {
					description = "Key Vault '" + vault.Name + "' has public network access enabled, Azure omitted the network ACL object, and no private endpoint is visible. Azure can return that shape for a fully open vault. Review whether that secret-management surface is intentionally internet reachable."
				}
				findings = append(findings, models.KeyVaultFinding{
					Description: description,
					ID:          "keyvault-public-network-open-" + vault.ID,
					RelatedIDs:  []string{vault.ID},
					Severity:    "high",
					Title:       "Key Vault is broadly reachable on the public network",
				})
			case !vault.PrivateEndpointEnabled:
				findings = append(findings, models.KeyVaultFinding{
					Description: "Key Vault '" + vault.Name + "' keeps public network access enabled with default network action '" + keyVaultDefaultActionText(vault.NetworkDefaultAction) + "' and no private endpoint visible. Review whether that public path is still intended.",
					ID:          "keyvault-public-network-enabled-" + vault.ID,
					RelatedIDs:  []string{vault.ID},
					Severity:    "medium",
					Title:       "Key Vault remains reachable through a public network path",
				})
			default:
				findings = append(findings, models.KeyVaultFinding{
					Description: "Key Vault '" + vault.Name + "' has public network access enabled while a private endpoint is also present. Validate whether the public path is still required.",
					ID:          "keyvault-public-network-with-private-endpoint-" + vault.ID,
					RelatedIDs:  []string{vault.ID},
					Severity:    "low",
					Title:       "Key Vault keeps a public network path alongside Private Link",
				})
			}
		}

		if !vault.PurgeProtectionEnabled {
			findings = append(findings, models.KeyVaultFinding{
				Description: "Key Vault '" + vault.Name + "' does not have purge protection enabled. Validate whether destructive recovery protections are intentionally absent.",
				ID:          "keyvault-purge-protection-disabled-" + vault.ID,
				RelatedIDs:  []string{vault.ID},
				Severity:    "medium",
				Title:       "Key Vault purge protection is disabled",
			})
		}
	}

	return findings
}

func keyVaultStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func keyVaultDefaultActionText(value *string) string {
	if value == nil || *value == "" {
		return "unknown"
	}
	return *value
}
