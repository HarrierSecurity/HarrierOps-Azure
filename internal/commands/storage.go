package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func storageHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Storage(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		assets := sortedByLess(facts.StorageAssets, storageLess)

		return models.StorageOutput{
			Findings:      storageFindings(assets),
			Issues:        facts.Issues,
			Metadata:      withArtifactContext(commandMetadata("storage", now, request, facts.TenantID, facts.SubscriptionID, facts.TokenSource), request, facts.CurrentPrincipal, facts.AuthMode, facts.TokenSource),
			StorageAssets: assets,
		}, nil
	}
}

func storageLess(left models.StorageAsset, right models.StorageAsset) bool {
	leftPublicAccess, leftNetwork, leftSharedKey, leftTLS, leftHTTPS, leftPrivateEndpoint := storagePriorityRank(left)
	rightPublicAccess, rightNetwork, rightSharedKey, rightTLS, rightHTTPS, rightPrivateEndpoint := storagePriorityRank(right)

	if leftPublicAccess != rightPublicAccess {
		return leftPublicAccess < rightPublicAccess
	}
	if leftNetwork != rightNetwork {
		return leftNetwork < rightNetwork
	}
	if leftSharedKey != rightSharedKey {
		return leftSharedKey < rightSharedKey
	}
	if leftTLS != rightTLS {
		return leftTLS < rightTLS
	}
	if leftHTTPS != rightHTTPS {
		return leftHTTPS < rightHTTPS
	}
	if leftPrivateEndpoint != rightPrivateEndpoint {
		return leftPrivateEndpoint < rightPrivateEndpoint
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func storagePriorityRank(item models.StorageAsset) (int, int, int, int, int, int) {
	publicAccessRank := 1
	if item.PublicAccess {
		publicAccessRank = 0
	}

	publicNetworkEnabled := normalizedLower(storageStringValue(item.PublicNetworkAccess)) == "enabled"
	networkDefaultAction := normalizedLower(storageStringValue(item.NetworkDefaultAction))
	networkRank := 3
	switch {
	case publicNetworkEnabled && networkDefaultAction == "allow":
		networkRank = 0
	case publicNetworkEnabled:
		networkRank = 1
	case networkDefaultAction == "allow":
		networkRank = 2
	}

	sharedKeyRank := 1
	if item.AllowSharedKeyAccess != nil && *item.AllowSharedKeyAccess {
		sharedKeyRank = 0
	}

	tlsRank := 1
	switch normalizedLower(storageStringValue(item.MinimumTLSVersion)) {
	case "tls1_0":
		tlsRank = 0
	case "tls1_1":
		tlsRank = 1
	case "tls1_2":
		tlsRank = 2
	case "tls1_3":
		tlsRank = 3
	}

	httpsRank := 1
	if item.HTTPSTrafficOnlyEnabled != nil && !*item.HTTPSTrafficOnlyEnabled {
		httpsRank = 0
	}

	privateEndpointRank := 1
	if !item.PrivateEndpointEnabled {
		privateEndpointRank = 0
	}

	return publicAccessRank, networkRank, sharedKeyRank, tlsRank, httpsRank, privateEndpointRank
}

func storageFindings(assets []models.StorageAsset) []models.StorageFinding {
	findings := []models.StorageFinding{}
	for _, asset := range assets {
		if asset.PublicAccess {
			findings = append(findings, models.StorageFinding{
				Description: "Storage account '" + asset.Name + "' has blob public access enabled. Validate anonymous access and exposed data paths.",
				ID:          "storage-public-" + asset.ID,
				RelatedIDs:  []string{asset.ID},
				Severity:    "high",
				Title:       "Storage account allows public blob access",
			})
		}
		if normalizedLower(storageStringValue(asset.NetworkDefaultAction)) == "allow" {
			findings = append(findings, models.StorageFinding{
				Description: "Storage account '" + asset.Name + "' default firewall action is Allow. Review allowed network sources and private endpoint posture.",
				ID:          "storage-firewall-open-" + asset.ID,
				RelatedIDs:  []string{asset.ID},
				Severity:    "medium",
				Title:       "Storage account network default action is Allow",
			})
		}
	}
	return findings
}

func storageStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
