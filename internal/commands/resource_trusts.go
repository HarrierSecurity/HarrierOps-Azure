package commands

import (
	"context"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func resourceTrustsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.ResourceTrusts(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		return models.ResourceTrustsOutput{
			Findings:       resourceTrustFindings(facts.StorageAssets, facts.KeyVaults),
			Issues:         facts.Issues,
			Metadata:       commandMetadata("resource-trusts", now, request, facts.TenantID, facts.SubscriptionID, ""),
			ResourceTrusts: sortedByLess(composeResourceTrusts(facts.StorageAssets, facts.KeyVaults), resourceTrustLess),
		}, nil
	}
}

func composeResourceTrusts(storageAssets []models.StorageAsset, keyVaults []models.KeyVaultAsset) []models.ResourceTrustSummary {
	trusts := append([]models.ResourceTrustSummary{}, resourceTrustsFromStorage(storageAssets)...)
	trusts = append(trusts, resourceTrustsFromKeyVault(keyVaults)...)
	return trusts
}

func resourceTrustsFromStorage(storageAssets []models.StorageAsset) []models.ResourceTrustSummary {
	trusts := []models.ResourceTrustSummary{}

	for _, asset := range storageAssets {
		if asset.ID == "" {
			continue
		}

		if asset.PublicAccess {
			trusts = append(trusts, models.ResourceTrustSummary{
				Confidence:   "confirmed",
				Exposure:     "high",
				RelatedIDs:   []string{asset.ID},
				ResourceID:   asset.ID,
				ResourceName: asset.Name,
				ResourceType: "StorageAccount",
				Summary:      "Storage account '" + resourceTrustName(asset.Name, asset.ID) + "' permits public blob access from the public network.",
				Target:       "public-network",
				TrustType:    "anonymous-blob-access",
			})
		}

		if normalizedLower(storageStringValue(asset.NetworkDefaultAction)) == "allow" {
			trusts = append(trusts, models.ResourceTrustSummary{
				Confidence:   "confirmed",
				Exposure:     "medium",
				RelatedIDs:   []string{asset.ID},
				ResourceID:   asset.ID,
				ResourceName: asset.Name,
				ResourceType: "StorageAccount",
				Summary:      "Storage account '" + resourceTrustName(asset.Name, asset.ID) + "' accepts public network traffic by default.",
				Target:       "public-network",
				TrustType:    "public-network-default",
			})
		}

		if asset.PrivateEndpointEnabled {
			trusts = append(trusts, models.ResourceTrustSummary{
				Confidence:   "confirmed",
				Exposure:     "restricted",
				RelatedIDs:   []string{asset.ID},
				ResourceID:   asset.ID,
				ResourceName: asset.Name,
				ResourceType: "StorageAccount",
				Summary:      "Storage account '" + resourceTrustName(asset.Name, asset.ID) + "' exposes a private endpoint path through Azure Private Link.",
				Target:       "private-link",
				TrustType:    "private-endpoint",
			})
		}
	}

	return trusts
}

func resourceTrustsFromKeyVault(keyVaults []models.KeyVaultAsset) []models.ResourceTrustSummary {
	trusts := []models.ResourceTrustSummary{}

	for _, vault := range keyVaults {
		if vault.ID == "" {
			continue
		}

		publicNetworkAccess := normalizedLower(keyVaultStringValue(vault.PublicNetworkAccess))
		networkDefaultAction := normalizedLower(keyVaultStringValue(vault.NetworkDefaultAction))
		if publicNetworkAccess == "enabled" {
			exposure := "medium"
			if networkDefaultAction == "allow" || networkDefaultAction == "" {
				exposure = "high"
			}
			trusts = append(trusts, models.ResourceTrustSummary{
				Confidence:   "confirmed",
				Exposure:     exposure,
				RelatedIDs:   []string{vault.ID},
				ResourceID:   vault.ID,
				ResourceName: vault.Name,
				ResourceType: "KeyVault",
				Summary:      "Key Vault '" + resourceTrustName(vault.Name, vault.ID) + "' remains reachable through a public network path.",
				Target:       "public-network",
				TrustType:    "public-network",
			})
		}

		if vault.PrivateEndpointEnabled {
			trusts = append(trusts, models.ResourceTrustSummary{
				Confidence:   "confirmed",
				Exposure:     "restricted",
				RelatedIDs:   []string{vault.ID},
				ResourceID:   vault.ID,
				ResourceName: vault.Name,
				ResourceType: "KeyVault",
				Summary:      "Key Vault '" + resourceTrustName(vault.Name, vault.ID) + "' exposes a private endpoint path through Azure Private Link.",
				Target:       "private-link",
				TrustType:    "private-endpoint",
			})
		}
	}

	return trusts
}

func resourceTrustFindings(storageAssets []models.StorageAsset, keyVaults []models.KeyVaultAsset) []models.ResourceTrustFinding {
	findings := make([]models.ResourceTrustFinding, 0, len(storageAssets)+len(keyVaults))
	for _, finding := range storageFindings(storageAssets) {
		findings = append(findings, models.ResourceTrustFinding{
			Description: finding.Description,
			ID:          finding.ID,
			RelatedIDs:  append([]string{}, finding.RelatedIDs...),
			Severity:    finding.Severity,
			Title:       finding.Title,
		})
	}
	for _, finding := range keyVaultFindings(keyVaults) {
		if strings.HasPrefix(finding.ID, "keyvault-purge-protection-disabled-") {
			continue
		}
		findings = append(findings, models.ResourceTrustFinding{
			Description: finding.Description,
			ID:          finding.ID,
			RelatedIDs:  append([]string{}, finding.RelatedIDs...),
			Severity:    finding.Severity,
			Title:       finding.Title,
		})
	}
	return findings
}

func resourceTrustLess(left models.ResourceTrustSummary, right models.ResourceTrustSummary) bool {
	leftHigh := resourceTrustHighRank(left.Exposure)
	rightHigh := resourceTrustHighRank(right.Exposure)
	if leftHigh != rightHigh {
		return leftHigh < rightHigh
	}
	if left.ResourceType != right.ResourceType {
		return left.ResourceType < right.ResourceType
	}
	leftName := resourceTrustName(left.ResourceName, left.ResourceID)
	rightName := resourceTrustName(right.ResourceName, right.ResourceID)
	if leftName != rightName {
		return leftName < rightName
	}
	if left.TrustType != right.TrustType {
		return left.TrustType < right.TrustType
	}
	return left.ResourceID < right.ResourceID
}

func resourceTrustHighRank(exposure string) int {
	if normalizedLower(exposure) == "high" {
		return 0
	}
	return 1
}

func resourceTrustName(name string, resourceID string) string {
	if name != "" {
		return name
	}
	return resourceID
}
