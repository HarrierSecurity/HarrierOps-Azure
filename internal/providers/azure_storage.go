package providers

import (
	"context"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

const armStorageAPIVersion = "2025-06-01"

func (provider AzureProvider) Storage(ctx context.Context, tenant string, subscription string) (StorageFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return StorageFacts{}, err
	}

	accounts, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.Storage/storageAccounts",
		armStorageAPIVersion,
	)
	if err != nil {
		return StorageFacts{
			ArtifactIdentityFacts: azureArtifactIdentityFacts(session),
			TenantID:              session.tenantID,
			SubscriptionID:        session.subscription.ID,
			StorageAssets:         []models.StorageAsset{},
			Issues:                []models.Issue{issueFromError("storage", err)},
		}, nil
	}

	assets := make([]models.StorageAsset, 0, len(accounts))
	issues := []models.Issue{}
	for _, account := range accounts {
		assets = append(assets, storageSummary(ctx, session, account, &issues))
	}
	sort.Slice(assets, func(i int, j int) bool {
		if assets[i].Name != assets[j].Name {
			return assets[i].Name < assets[j].Name
		}
		return assets[i].ID < assets[j].ID
	})

	return StorageFacts{
		ArtifactIdentityFacts: azureArtifactIdentityFacts(session),
		TenantID:              session.tenantID,
		SubscriptionID:        session.subscription.ID,
		StorageAssets:         assets,
		Issues:                issues,
	}, nil
}

func storageSummary(ctx context.Context, session azureSession, account map[string]any, issues *[]models.Issue) models.StorageAsset {
	accountID := mapStringValue(account, "id")
	accountName := firstNonEmpty(mapStringValue(account, "name"), resourceNameFromID(accountID), "unknown")
	resourceGroup := resourceGroupFromID(accountID)
	properties := mapValue(account, "properties")
	networkRuleSet := mapValue(properties, "networkAcls", "network_acls", "networkRuleSet", "network_rule_set")
	privateEndpoints := listValue(properties, "privateEndpointConnections", "private_endpoint_connections")

	containerCount := storageChildCount(ctx, session, resourceGroup, accountName, accountID+"/blobServices/default/containers", "blob_containers", issues)
	fileShareCount := storageChildCount(ctx, session, resourceGroup, accountName, accountID+"/fileServices/default/shares", "file_shares", issues)
	queueCount := storageChildCount(ctx, session, resourceGroup, accountName, accountID+"/queueServices/default/queues", "queue", issues)
	tableCount := storageChildCount(ctx, session, resourceGroup, accountName, accountID+"/tableServices/default/tables", "table", issues)

	publicAccess := mapBoolValue(properties, "allowBlobPublicAccess", "allow_blob_public_access")
	networkDefaultAction := stringPtr(mapStringValue(networkRuleSet, "defaultAction", "default_action"))
	indicators := []string{}
	if publicAccess {
		indicators = append(indicators, "allow_blob_public_access=true")
	}
	if strings.EqualFold(stringPtrValue(networkDefaultAction), "allow") {
		indicators = append(indicators, "network_default_action=Allow")
	}

	return models.StorageAsset{
		AllowSharedKeyAccess:      optionalBoolPtr(properties, "allowSharedKeyAccess", "allow_shared_key_access"),
		AnonymousAccessIndicators: indicators,
		ContainerCount:            containerCount,
		DNSEndpointType:           stringPtr(mapStringValue(properties, "dnsEndpointType", "dns_endpoint_type")),
		FileShareCount:            fileShareCount,
		HTTPSTrafficOnlyEnabled: optionalBoolPtr(
			properties,
			"supportsHttpsTrafficOnly",
			"supports_https_traffic_only",
			"enableHttpsTrafficOnly",
			"enable_https_traffic_only",
		),
		ID:                     firstNonEmpty(accountID, "/unknown/"+accountName),
		IsHNSEnabled:           optionalBoolPtr(properties, "isHnsEnabled", "is_hns_enabled"),
		IsSFTPEnabled:          optionalBoolPtr(properties, "isSftpEnabled", "is_sftp_enabled"),
		Location:               stringPtr(mapStringValue(account, "location")),
		MinimumTLSVersion:      stringPtr(mapStringValue(properties, "minimumTlsVersion", "minimum_tls_version")),
		Name:                   accountName,
		NetworkDefaultAction:   networkDefaultAction,
		NFSV3Enabled:           optionalBoolPtr(properties, "isNfsV3Enabled", "is_nfs_v3_enabled", "enableNfsV3", "enable_nfs_v3"),
		PrivateEndpointEnabled: len(privateEndpoints) > 0,
		PublicAccess:           publicAccess,
		PublicNetworkAccess:    stringPtr(mapStringValue(properties, "publicNetworkAccess", "public_network_access")),
		QueueCount:             queueCount,
		ResourceGroup:          resourceGroup,
		TableCount:             tableCount,
	}
}

func storageChildCount(
	ctx context.Context,
	session azureSession,
	resourceGroup string,
	accountName string,
	path string,
	suffix string,
	issues *[]models.Issue,
) *int {
	if resourceGroup == "" {
		return nil
	}
	children, err := armListObjects(ctx, session.credential, path, armStorageAPIVersion)
	if err != nil {
		*issues = append(*issues, issueFromError("storage["+resourceGroup+"/"+accountName+"]."+suffix, err))
		return nil
	}
	count := len(children)
	return &count
}
