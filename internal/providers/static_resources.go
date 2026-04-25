package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) Acr(_ context.Context, tenant string, subscription string) (AcrFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	eastus := "eastus"
	eastus2 := "eastus2"
	succeeded := "Succeeded"
	enabled := "Enabled"
	disabled := "Disabled"
	enabledLower := "enabled"
	disabledLower := "disabled"
	allow := "Allow"
	deny := "Deny"
	azureServices := "AzureServices"
	none := "None"
	standard := "Standard"
	premium := "Premium"
	systemAssigned := "SystemAssigned"
	notary := "notary"
	trueValue := true
	falseValue := false
	zero := 0
	one := 1
	two := 2
	thirty := 30

	publicRegistryID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.ContainerRegistry/registries/acr-public-legacy"
	opsRegistryID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.ContainerRegistry/registries/acr-ops-01"
	opsPrincipalID := "99990000-0000-0000-0000-000000000031"
	opsClientID := "99990000-0000-0000-0000-000000000032"

	return AcrFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Registries: []models.AcrRegistryAsset{
			{
				ID:                             publicRegistryID,
				Name:                           "acr-public-legacy",
				ResourceGroup:                  "rg-containers",
				Location:                       &eastus,
				State:                          &succeeded,
				LoginServer:                    models.StringPtr("acr-public-legacy.azurecr.io"),
				SKUName:                        &standard,
				PublicNetworkAccess:            &enabled,
				NetworkRuleDefaultAction:       &allow,
				NetworkRuleBypassOptions:       &azureServices,
				AdminUserEnabled:               &trueValue,
				AnonymousPullEnabled:           &trueValue,
				DataEndpointEnabled:            &trueValue,
				PrivateEndpointConnectionCount: &zero,
				WebhookCount:                   &two,
				EnabledWebhookCount:            &one,
				WebhookActionTypes:             []string{"delete", "push"},
				BroadWebhookScopeCount:         &one,
				ReplicationCount:               &zero,
				ReplicationRegions:             []string{},
				QuarantinePolicyStatus:         &disabledLower,
				RetentionPolicyStatus:          &disabledLower,
				RetentionPolicyDays:            nil,
				TrustPolicyStatus:              &disabledLower,
				TrustPolicyType:                nil,
				WorkloadIdentityType:           nil,
				WorkloadPrincipalID:            nil,
				WorkloadClientID:               nil,
				WorkloadIdentityIDs:            []string{},
				Summary: acrOperatorSummary(
					"acr-public-legacy",
					models.StringPtr("acr-public-legacy.azurecr.io"),
					nil,
					&enabled,
					&allow,
					&azureServices,
					&trueValue,
					&trueValue,
					&trueValue,
					&zero,
					&standard,
					&two,
					&one,
					[]string{"delete", "push"},
					&one,
					&zero,
					[]string{},
					&disabledLower,
					&disabledLower,
					nil,
					&disabledLower,
					nil,
				),
				RelatedIDs: []string{
					publicRegistryID,
					publicRegistryID + "/webhooks/push-all",
					publicRegistryID + "/webhooks/delete-all",
				},
			},
			{
				ID:                             opsRegistryID,
				Name:                           "acr-ops-01",
				ResourceGroup:                  "rg-containers",
				Location:                       &eastus2,
				State:                          &succeeded,
				LoginServer:                    models.StringPtr("acr-ops-01.azurecr.io"),
				SKUName:                        &premium,
				PublicNetworkAccess:            &disabled,
				NetworkRuleDefaultAction:       &deny,
				NetworkRuleBypassOptions:       &none,
				AdminUserEnabled:               &falseValue,
				AnonymousPullEnabled:           &falseValue,
				DataEndpointEnabled:            &falseValue,
				PrivateEndpointConnectionCount: &one,
				WebhookCount:                   &one,
				EnabledWebhookCount:            &one,
				WebhookActionTypes:             []string{"push"},
				BroadWebhookScopeCount:         &zero,
				ReplicationCount:               &two,
				ReplicationRegions:             []string{"northeurope", "westus2"},
				QuarantinePolicyStatus:         &enabledLower,
				RetentionPolicyStatus:          &enabledLower,
				RetentionPolicyDays:            &thirty,
				TrustPolicyStatus:              &enabledLower,
				TrustPolicyType:                &notary,
				WorkloadIdentityType:           &systemAssigned,
				WorkloadPrincipalID:            &opsPrincipalID,
				WorkloadClientID:               &opsClientID,
				WorkloadIdentityIDs:            []string{},
				Summary: acrOperatorSummary(
					"acr-ops-01",
					models.StringPtr("acr-ops-01.azurecr.io"),
					&systemAssigned,
					&disabled,
					&deny,
					&none,
					&falseValue,
					&falseValue,
					&falseValue,
					&one,
					&premium,
					&one,
					&one,
					[]string{"push"},
					&zero,
					&two,
					[]string{"northeurope", "westus2"},
					&enabledLower,
					&enabledLower,
					&thirty,
					&enabledLower,
					&notary,
				),
				RelatedIDs: []string{
					opsRegistryID,
					opsPrincipalID,
					opsRegistryID + "/webhooks/push-internal",
					opsRegistryID + "/replications/westus2",
					opsRegistryID + "/replications/northeurope",
				},
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) Databases(_ context.Context, tenant string, subscription string) (DatabasesFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	eastus := "eastus"
	eastus2 := "eastus2"
	centralus := "centralus"
	ready := "Ready"
	sqlVersion := "12.0"
	postgresVersion := "16"
	mysqlVersion := "8.0.21"
	tls12 := "1.2"
	enabled := "Enabled"
	disabled := "Disabled"
	systemAssigned := "SystemAssigned"
	haDisabled := "disabled"
	haZoneRedundant := "zone-redundant"

	sqlPublicID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.Sql/servers/sql-public-legacy"
	pgPublicID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.DBforPostgreSQL/flexibleServers/pg-public-legacy"
	sqlOpsID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.Sql/servers/sql-ops-01"
	mysqlOpsID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.DBforMySQL/flexibleServers/mysql-ops-01"
	mysqlSubnetID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.Network/virtualNetworks/vnet-data/subnets/mysql-flex"
	mysqlPrivateDNSID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.Network/privateDnsZones/mysql.database.azure.com"
	sqlOpsPrincipalID := "99990000-0000-0000-0000-000000000041"
	sqlOpsClientID := "99990000-0000-0000-0000-000000000042"
	mysqlOpsPrincipalID := "99990000-0000-0000-0000-000000000051"
	mysqlOpsClientID := "99990000-0000-0000-0000-000000000052"

	two := 2
	one := 1

	return DatabasesFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		DatabaseServers: []models.DatabaseServerAsset{
			{
				DatabaseCount:             &two,
				DelegatedSubnetResourceID: nil,
				Engine:                    "AzureSql",
				FullyQualifiedDomainName:  models.StringPtr("sql-public-legacy.database.windows.net"),
				HighAvailabilityMode:      nil,
				ID:                        sqlPublicID,
				Location:                  &eastus,
				MinimalTLSVersion:         &tls12,
				Name:                      "sql-public-legacy",
				PrivateDNSZoneResourceID:  nil,
				PublicNetworkAccess:       &enabled,
				RelatedIDs:                []string{sqlPublicID},
				ResourceGroup:             "rg-data",
				ServerVersion:             &sqlVersion,
				State:                     &ready,
				Summary: databaseServerOperatorSummary(
					"AzureSql",
					"sql-public-legacy",
					models.StringPtr("sql-public-legacy.database.windows.net"),
					nil,
					&enabled,
					&tls12,
					&sqlVersion,
					nil,
					nil,
					nil,
					&two,
					[]string{"orders", "reporting"},
				),
				UserDatabaseNames:    []string{"orders", "reporting"},
				WorkloadClientID:     nil,
				WorkloadIdentityIDs:  []string{},
				WorkloadIdentityType: nil,
				WorkloadPrincipalID:  nil,
			},
			{
				DatabaseCount:             &two,
				DelegatedSubnetResourceID: nil,
				Engine:                    "PostgreSqlFlexible",
				FullyQualifiedDomainName:  models.StringPtr("pg-public-legacy.postgres.database.azure.com"),
				HighAvailabilityMode:      &haDisabled,
				ID:                        pgPublicID,
				Location:                  &eastus2,
				MinimalTLSVersion:         nil,
				Name:                      "pg-public-legacy",
				PrivateDNSZoneResourceID:  nil,
				PublicNetworkAccess:       &enabled,
				RelatedIDs:                []string{pgPublicID},
				ResourceGroup:             "rg-data",
				ServerVersion:             &postgresVersion,
				State:                     &ready,
				Summary: databaseServerOperatorSummary(
					"PostgreSqlFlexible",
					"pg-public-legacy",
					models.StringPtr("pg-public-legacy.postgres.database.azure.com"),
					nil,
					&enabled,
					nil,
					&postgresVersion,
					&haDisabled,
					nil,
					nil,
					&two,
					[]string{"app", "orders"},
				),
				UserDatabaseNames:    []string{"app", "orders"},
				WorkloadClientID:     nil,
				WorkloadIdentityIDs:  []string{},
				WorkloadIdentityType: nil,
				WorkloadPrincipalID:  nil,
			},
			{
				DatabaseCount:             &one,
				DelegatedSubnetResourceID: nil,
				Engine:                    "AzureSql",
				FullyQualifiedDomainName:  models.StringPtr("sql-ops-01.database.windows.net"),
				HighAvailabilityMode:      nil,
				ID:                        sqlOpsID,
				Location:                  &centralus,
				MinimalTLSVersion:         &tls12,
				Name:                      "sql-ops-01",
				PrivateDNSZoneResourceID:  nil,
				PublicNetworkAccess:       &disabled,
				RelatedIDs:                []string{sqlOpsID, sqlOpsPrincipalID},
				ResourceGroup:             "rg-data",
				ServerVersion:             &sqlVersion,
				State:                     &ready,
				Summary: databaseServerOperatorSummary(
					"AzureSql",
					"sql-ops-01",
					models.StringPtr("sql-ops-01.database.windows.net"),
					&systemAssigned,
					&disabled,
					&tls12,
					&sqlVersion,
					nil,
					nil,
					nil,
					&one,
					[]string{"appdb"},
				),
				UserDatabaseNames:    []string{"appdb"},
				WorkloadClientID:     &sqlOpsClientID,
				WorkloadIdentityIDs:  []string{},
				WorkloadIdentityType: &systemAssigned,
				WorkloadPrincipalID:  &sqlOpsPrincipalID,
			},
			{
				DatabaseCount:             &one,
				DelegatedSubnetResourceID: &mysqlSubnetID,
				Engine:                    "MySqlFlexible",
				FullyQualifiedDomainName:  models.StringPtr("mysql-ops-01.mysql.database.azure.com"),
				HighAvailabilityMode:      &haZoneRedundant,
				ID:                        mysqlOpsID,
				Location:                  &centralus,
				MinimalTLSVersion:         nil,
				Name:                      "mysql-ops-01",
				PrivateDNSZoneResourceID:  &mysqlPrivateDNSID,
				PublicNetworkAccess:       &disabled,
				RelatedIDs:                []string{mysqlOpsID, mysqlOpsPrincipalID},
				ResourceGroup:             "rg-data",
				ServerVersion:             &mysqlVersion,
				State:                     &ready,
				Summary: databaseServerOperatorSummary(
					"MySqlFlexible",
					"mysql-ops-01",
					models.StringPtr("mysql-ops-01.mysql.database.azure.com"),
					&systemAssigned,
					&disabled,
					nil,
					&mysqlVersion,
					&haZoneRedundant,
					&mysqlSubnetID,
					&mysqlPrivateDNSID,
					&one,
					[]string{"inventory"},
				),
				UserDatabaseNames:    []string{"inventory"},
				WorkloadClientID:     &mysqlOpsClientID,
				WorkloadIdentityIDs:  []string{},
				WorkloadIdentityType: &systemAssigned,
				WorkloadPrincipalID:  &mysqlOpsPrincipalID,
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) KeyVault(_ context.Context, tenant string, subscription string) (KeyVaultFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	eastus := "eastus"
	enabled := "Enabled"
	disabled := "Disabled"
	deny := "Deny"
	standard := "standard"
	premium := "premium"

	openVaultID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-secrets/providers/Microsoft.KeyVault/vaults/kvlabopen01"
	denyVaultID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-secrets/providers/Microsoft.KeyVault/vaults/kvlabdeny01"
	hybridVaultID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-secrets/providers/Microsoft.KeyVault/vaults/kvlabhybrid01"
	privateVaultID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-secrets/providers/Microsoft.KeyVault/vaults/kvlabpriv01"

	return KeyVaultFacts{
		ArtifactIdentityFacts: staticArtifactIdentityFacts(session),
		TenantID:              session.TenantID,
		SubscriptionID:        subscriptionID,
		KeyVaults: []models.KeyVaultAsset{
			{
				AccessPolicyCount:       2,
				EnableRBACAuthorization: false,
				ID:                      openVaultID,
				Location:                &eastus,
				Name:                    "kvlabopen01",
				NetworkDefaultAction:    nil,
				PrivateEndpointEnabled:  false,
				PublicNetworkAccess:     &enabled,
				PurgeProtectionEnabled:  false,
				ResourceGroup:           "rg-secrets",
				SKUName:                 &standard,
				SoftDeleteEnabled:       true,
				TenantID:                models.StringPtr(session.TenantID),
				VaultURI:                models.StringPtr("https://kvlabopen01.vault.azure.net/"),
			},
			{
				AccessPolicyCount:       1,
				EnableRBACAuthorization: true,
				ID:                      denyVaultID,
				Location:                &eastus,
				Name:                    "kvlabdeny01",
				NetworkDefaultAction:    &deny,
				PrivateEndpointEnabled:  false,
				PublicNetworkAccess:     &enabled,
				PurgeProtectionEnabled:  true,
				ResourceGroup:           "rg-secrets",
				SKUName:                 &standard,
				SoftDeleteEnabled:       true,
				TenantID:                models.StringPtr(session.TenantID),
				VaultURI:                models.StringPtr("https://kvlabdeny01.vault.azure.net/"),
			},
			{
				AccessPolicyCount:       0,
				EnableRBACAuthorization: true,
				ID:                      hybridVaultID,
				Location:                &eastus,
				Name:                    "kvlabhybrid01",
				NetworkDefaultAction:    &deny,
				PrivateEndpointEnabled:  true,
				PublicNetworkAccess:     &enabled,
				PurgeProtectionEnabled:  true,
				ResourceGroup:           "rg-secrets",
				SKUName:                 &premium,
				SoftDeleteEnabled:       true,
				TenantID:                models.StringPtr(session.TenantID),
				VaultURI:                models.StringPtr("https://kvlabhybrid01.vault.azure.net/"),
			},
			{
				AccessPolicyCount:       0,
				EnableRBACAuthorization: true,
				ID:                      privateVaultID,
				Location:                &eastus,
				Name:                    "kvlabpriv01",
				NetworkDefaultAction:    &deny,
				PrivateEndpointEnabled:  true,
				PublicNetworkAccess:     &disabled,
				PurgeProtectionEnabled:  true,
				ResourceGroup:           "rg-secrets",
				SKUName:                 &premium,
				SoftDeleteEnabled:       true,
				TenantID:                models.StringPtr(session.TenantID),
				VaultURI:                models.StringPtr("https://kvlabpriv01.vault.azure.net/"),
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) ApplicationGateway(_ context.Context, tenant string, subscription string) (ApplicationGatewayFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	eastus := "eastus"
	eastus2 := "eastus2"
	centralus := "centralus"
	running := "Running"
	standardV2 := "Standard_v2"
	wafV2 := "WAF_v2"
	detection := "Detection"
	prevention := "Prevention"
	trueValue := true
	falseValue := false

	sharedID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-edge/providers/Microsoft.Network/applicationGateways/agw-shared-edge-01"
	customerID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-edge/providers/Microsoft.Network/applicationGateways/agw-customer-edge-02"
	internalID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-payments/providers/Microsoft.Network/applicationGateways/agw-internal-payments"
	sharedPublicIPID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-edge/providers/Microsoft.Network/publicIPAddresses/pip-agw-shared-edge-01"
	customerPublicIPID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-edge/providers/Microsoft.Network/publicIPAddresses/pip-agw-customer-edge-02"
	sharedSubnetID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-edge/providers/Microsoft.Network/virtualNetworks/vnet-edge/subnets/appgw-shared"
	customerSubnetID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-edge/providers/Microsoft.Network/virtualNetworks/vnet-edge/subnets/appgw-customer"
	internalSubnetID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-payments/providers/Microsoft.Network/virtualNetworks/vnet-payments/subnets/appgw-internal"
	customerPolicyID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-edge/providers/Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/agw-customer-edge-02-policy"
	internalPolicyID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-payments/providers/Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies/agw-internal-payments-policy"

	return ApplicationGatewayFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		ApplicationGateways: []models.ApplicationGatewayAsset{
			{
				BackendPoolCount:           3,
				BackendTargetCount:         5,
				FirewallPolicyID:           nil,
				ID:                         sharedID,
				ListenerCount:              4,
				Location:                   &eastus,
				Name:                       "agw-shared-edge-01",
				PrivateFrontendCount:       0,
				PrivateFrontendIPs:         []string{},
				PublicFrontendCount:        1,
				PublicIPAddressIDs:         []string{sharedPublicIPID},
				PublicIPAddresses:          []string{"20.30.40.50"},
				RedirectConfigurationCount: 1,
				RelatedIDs:                 []string{sharedID, sharedPublicIPID, sharedSubnetID},
				RequestRoutingRuleCount:    4,
				ResourceGroup:              "rg-edge",
				RewriteRuleSetCount:        0,
				SKUName:                    &standardV2,
				SKUTier:                    &standardV2,
				State:                      &running,
				SubnetIDs:                  []string{sharedSubnetID},
				Summary: applicationGatewayOperatorSummary(
					"agw-shared-edge-01",
					1,
					0,
					[]string{"20.30.40.50"},
					4,
					4,
					3,
					5,
					&falseValue,
					nil,
					nil,
				),
				URLPathMapCount: 1,
				WAFEnabled:      &falseValue,
				WAFMode:         nil,
			},
			{
				BackendPoolCount:           2,
				BackendTargetCount:         4,
				FirewallPolicyID:           &customerPolicyID,
				ID:                         customerID,
				ListenerCount:              3,
				Location:                   &eastus2,
				Name:                       "agw-customer-edge-02",
				PrivateFrontendCount:       1,
				PrivateFrontendIPs:         []string{"10.20.0.10"},
				PublicFrontendCount:        1,
				PublicIPAddressIDs:         []string{customerPublicIPID},
				PublicIPAddresses:          []string{"20.30.40.60"},
				RedirectConfigurationCount: 0,
				RelatedIDs:                 []string{customerID, customerPublicIPID, customerSubnetID, customerPolicyID},
				RequestRoutingRuleCount:    3,
				ResourceGroup:              "rg-edge",
				RewriteRuleSetCount:        1,
				SKUName:                    &wafV2,
				SKUTier:                    &wafV2,
				State:                      &running,
				SubnetIDs:                  []string{customerSubnetID},
				Summary: applicationGatewayOperatorSummary(
					"agw-customer-edge-02",
					1,
					1,
					[]string{"20.30.40.60"},
					3,
					3,
					2,
					4,
					&trueValue,
					&detection,
					&customerPolicyID,
				),
				URLPathMapCount: 1,
				WAFEnabled:      &trueValue,
				WAFMode:         &detection,
			},
			{
				BackendPoolCount:           2,
				BackendTargetCount:         2,
				FirewallPolicyID:           &internalPolicyID,
				ID:                         internalID,
				ListenerCount:              2,
				Location:                   &centralus,
				Name:                       "agw-internal-payments",
				PrivateFrontendCount:       1,
				PrivateFrontendIPs:         []string{"10.30.0.15"},
				PublicFrontendCount:        0,
				PublicIPAddressIDs:         []string{},
				PublicIPAddresses:          []string{},
				RedirectConfigurationCount: 0,
				RelatedIDs:                 []string{internalID, internalSubnetID, internalPolicyID},
				RequestRoutingRuleCount:    2,
				ResourceGroup:              "rg-payments",
				RewriteRuleSetCount:        0,
				SKUName:                    &wafV2,
				SKUTier:                    &wafV2,
				State:                      &running,
				SubnetIDs:                  []string{internalSubnetID},
				Summary: applicationGatewayOperatorSummary(
					"agw-internal-payments",
					0,
					1,
					[]string{},
					2,
					2,
					2,
					2,
					&trueValue,
					&prevention,
					&internalPolicyID,
				),
				URLPathMapCount: 0,
				WAFEnabled:      &trueValue,
				WAFMode:         &prevention,
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) Storage(_ context.Context, tenant string, subscription string) (StorageFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	eastus := "eastus"
	enabled := "Enabled"
	disabled := "Disabled"
	allow := "Allow"
	deny := "Deny"
	tls10 := "TLS1_0"
	tls12 := "TLS1_2"
	standard := "Standard"
	trueValue := true
	falseValue := false
	zero := 0
	one := 1
	two := 2
	three := 3

	publicID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.Storage/storageAccounts/stlabpub01"
	privateID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.Storage/storageAccounts/stlabpriv01"

	return StorageFacts{
		ArtifactIdentityFacts: staticArtifactIdentityFacts(session),
		TenantID:              session.TenantID,
		SubscriptionID:        subscriptionID,
		StorageAssets: []models.StorageAsset{
			{
				AllowSharedKeyAccess:      &trueValue,
				AnonymousAccessIndicators: []string{"allow_blob_public_access=true", "network_default_action=Allow"},
				ContainerCount:            &three,
				DNSEndpointType:           &standard,
				FileShareCount:            &one,
				HTTPSTrafficOnlyEnabled:   &falseValue,
				ID:                        publicID,
				IsHNSEnabled:              &falseValue,
				IsSFTPEnabled:             &falseValue,
				Location:                  &eastus,
				MinimumTLSVersion:         &tls10,
				Name:                      "stlabpub01",
				NetworkDefaultAction:      &allow,
				NFSV3Enabled:              &falseValue,
				PrivateEndpointEnabled:    false,
				PublicAccess:              true,
				PublicNetworkAccess:       &enabled,
				QueueCount:                &two,
				ResourceGroup:             "rg-data",
				TableCount:                &zero,
			},
			{
				AllowSharedKeyAccess:      &falseValue,
				AnonymousAccessIndicators: []string{},
				ContainerCount:            &two,
				DNSEndpointType:           &standard,
				FileShareCount:            &zero,
				HTTPSTrafficOnlyEnabled:   &trueValue,
				ID:                        privateID,
				IsHNSEnabled:              &trueValue,
				IsSFTPEnabled:             &trueValue,
				Location:                  &eastus,
				MinimumTLSVersion:         &tls12,
				Name:                      "stlabpriv01",
				NetworkDefaultAction:      &deny,
				NFSV3Enabled:              &falseValue,
				PrivateEndpointEnabled:    true,
				PublicAccess:              false,
				PublicNetworkAccess:       &disabled,
				QueueCount:                &zero,
				ResourceGroup:             "rg-data",
				TableCount:                &one,
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) SnapshotsDisks(_ context.Context, tenant string, subscription string) (SnapshotsDisksFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return SnapshotsDisksFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		SnapshotDiskAssets: []models.SnapshotDiskAsset{
			{
				ID:                  "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/disks/vm-web-01-os",
				Name:                "vm-web-01-os",
				AssetKind:           "disk",
				ResourceGroup:       "rg-workload",
				Location:            models.StringPtr("eastus"),
				DiskRole:            models.StringPtr("os-disk"),
				AttachmentState:     "attached",
				AttachedToID:        models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01"),
				AttachedToName:      models.StringPtr("vm-web-01"),
				SourceResourceID:    nil,
				SourceResourceName:  nil,
				SourceResourceKind:  nil,
				OSType:              models.StringPtr("Linux"),
				SizeGB:              intPtr(128),
				TimeCreated:         models.StringPtr("2026-03-28T15:10:00+00:00"),
				Incremental:         nil,
				NetworkAccessPolicy: models.StringPtr("AllowPrivate"),
				PublicNetworkAccess: models.StringPtr("Disabled"),
				DiskAccessID:        nil,
				MaxShares:           intPtr(1),
				EncryptionType:      models.StringPtr("EncryptionAtRestWithCustomerKey"),
				DiskEncryptionSetID: models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-sec/providers/Microsoft.Compute/diskEncryptionSets/des-prod"),
				Summary:             "Attached os-disk for vm-web-01; public network Disabled, network access AllowPrivate; encryption posture: EncryptionAtRestWithCustomerKey, disk encryption set linked.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-sec/providers/Microsoft.Compute/diskEncryptionSets/des-prod",
				},
			},
			{
				ID:                  "/subscriptions/" + subscriptionID + "/resourceGroups/rg-legacy/providers/Microsoft.Compute/disks/data-detached-legacy",
				Name:                "data-detached-legacy",
				AssetKind:           "disk",
				ResourceGroup:       "rg-legacy",
				Location:            models.StringPtr("eastus2"),
				DiskRole:            nil,
				AttachmentState:     "detached",
				AttachedToID:        nil,
				AttachedToName:      nil,
				SourceResourceID:    nil,
				SourceResourceName:  nil,
				SourceResourceKind:  nil,
				OSType:              nil,
				SizeGB:              intPtr(512),
				TimeCreated:         models.StringPtr("2026-03-20T09:22:00+00:00"),
				Incremental:         nil,
				NetworkAccessPolicy: models.StringPtr("AllowAll"),
				PublicNetworkAccess: models.StringPtr("Enabled"),
				DiskAccessID:        models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-legacy/providers/Microsoft.Compute/diskAccesses/legacy-diskaccess"),
				MaxShares:           intPtr(3),
				EncryptionType:      models.StringPtr("EncryptionAtRestWithPlatformKey"),
				DiskEncryptionSetID: nil,
				Summary:             "Detached managed disk; public network Enabled, network access AllowAll, max shares 3, disk access resource visible; encryption posture: EncryptionAtRestWithPlatformKey.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-legacy/providers/Microsoft.Compute/diskAccesses/legacy-diskaccess",
				},
			},
			{
				ID:                  "/subscriptions/" + subscriptionID + "/resourceGroups/rg-legacy/providers/Microsoft.Compute/snapshots/data-detached-legacy-snap",
				Name:                "data-detached-legacy-snap",
				AssetKind:           "snapshot",
				ResourceGroup:       "rg-legacy",
				Location:            models.StringPtr("eastus2"),
				DiskRole:            nil,
				AttachmentState:     "snapshot",
				AttachedToID:        nil,
				AttachedToName:      nil,
				SourceResourceID:    models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-legacy/providers/Microsoft.Compute/disks/data-detached-legacy"),
				SourceResourceName:  models.StringPtr("data-detached-legacy"),
				SourceResourceKind:  models.StringPtr("disk"),
				OSType:              nil,
				SizeGB:              intPtr(512),
				TimeCreated:         models.StringPtr("2026-04-01T03:15:00+00:00"),
				Incremental:         boolPtr(true),
				NetworkAccessPolicy: models.StringPtr("AllowAll"),
				PublicNetworkAccess: models.StringPtr("Enabled"),
				DiskAccessID:        models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-legacy/providers/Microsoft.Compute/diskAccesses/legacy-diskaccess"),
				MaxShares:           nil,
				EncryptionType:      models.StringPtr("EncryptionAtRestWithPlatformKey"),
				DiskEncryptionSetID: nil,
				Summary:             "Snapshot of data-detached-legacy; incremental copy path visible; public network Enabled, network access AllowAll, disk access resource visible; encryption posture: EncryptionAtRestWithPlatformKey.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-legacy/providers/Microsoft.Compute/disks/data-detached-legacy",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-legacy/providers/Microsoft.Compute/diskAccesses/legacy-diskaccess",
				},
			},
			{
				ID:                  "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/snapshots/vm-web-01-os-snap",
				Name:                "vm-web-01-os-snap",
				AssetKind:           "snapshot",
				ResourceGroup:       "rg-workload",
				Location:            models.StringPtr("eastus"),
				DiskRole:            nil,
				AttachmentState:     "snapshot",
				AttachedToID:        nil,
				AttachedToName:      nil,
				SourceResourceID:    models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/disks/vm-web-01-os"),
				SourceResourceName:  models.StringPtr("vm-web-01-os"),
				SourceResourceKind:  models.StringPtr("disk"),
				OSType:              models.StringPtr("Linux"),
				SizeGB:              intPtr(128),
				TimeCreated:         models.StringPtr("2026-04-02T06:40:00+00:00"),
				Incremental:         boolPtr(true),
				NetworkAccessPolicy: models.StringPtr("AllowPrivate"),
				PublicNetworkAccess: models.StringPtr("Disabled"),
				DiskAccessID:        nil,
				MaxShares:           nil,
				EncryptionType:      models.StringPtr("EncryptionAtRestWithCustomerKey"),
				DiskEncryptionSetID: models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-sec/providers/Microsoft.Compute/diskEncryptionSets/des-prod"),
				Summary:             "Snapshot of vm-web-01-os; incremental copy path visible; public network Disabled, network access AllowPrivate; encryption posture: EncryptionAtRestWithCustomerKey, disk encryption set linked.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/disks/vm-web-01-os",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-sec/providers/Microsoft.Compute/diskEncryptionSets/des-prod",
				},
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) DNS(_ context.Context, tenant string, subscription string) (DNSFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	global := "global"
	nine := 9
	four := 4
	six := 6
	two := 2
	one := 1
	tenThousand := 10000
	twentyFiveThousand := 25000

	publicZoneID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-network/providers/Microsoft.Network/dnszones/corp.example.com"
	partnerZoneID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-edge/providers/Microsoft.Network/dnszones/partner.example.net"
	privateZoneID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-network/providers/Microsoft.Network/privateDnsZones/privatelink.database.windows.net"
	privateEndpointPrimary := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.Network/privateEndpoints/pe-sql-primary"
	privateEndpointReplica := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.Network/privateEndpoints/pe-sql-replica"

	return DNSFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		DNSZones: []models.DnsZoneAsset{
			{
				ID:                              publicZoneID,
				Name:                            "corp.example.com",
				ResourceGroup:                   "rg-network",
				Location:                        &global,
				ZoneKind:                        "public",
				RecordSetCount:                  &nine,
				MaxRecordSetCount:               &tenThousand,
				NameServers:                     []string{"ns1-01.azure-dns.com", "ns2-01.azure-dns.net", "ns3-01.azure-dns.org", "ns4-01.azure-dns.info"},
				LinkedVirtualNetworkCount:       nil,
				RegistrationVirtualNetworkCount: nil,
				PrivateEndpointReferenceCount:   nil,
				Summary: dnsZoneOperatorSummary(
					"corp.example.com",
					"public",
					&nine,
					4,
					nil,
					nil,
					nil,
				),
				RelatedIDs: []string{publicZoneID},
			},
			{
				ID:                              partnerZoneID,
				Name:                            "partner.example.net",
				ResourceGroup:                   "rg-edge",
				Location:                        &global,
				ZoneKind:                        "public",
				RecordSetCount:                  &four,
				MaxRecordSetCount:               &tenThousand,
				NameServers:                     []string{"ns1-08.azure-dns.com", "ns2-08.azure-dns.net", "ns3-08.azure-dns.org", "ns4-08.azure-dns.info"},
				LinkedVirtualNetworkCount:       nil,
				RegistrationVirtualNetworkCount: nil,
				PrivateEndpointReferenceCount:   nil,
				Summary: dnsZoneOperatorSummary(
					"partner.example.net",
					"public",
					&four,
					4,
					nil,
					nil,
					nil,
				),
				RelatedIDs: []string{partnerZoneID},
			},
			{
				ID:                              privateZoneID,
				Name:                            "privatelink.database.windows.net",
				ResourceGroup:                   "rg-network",
				Location:                        &global,
				ZoneKind:                        "private",
				RecordSetCount:                  &six,
				MaxRecordSetCount:               &twentyFiveThousand,
				NameServers:                     []string{},
				LinkedVirtualNetworkCount:       &two,
				RegistrationVirtualNetworkCount: &one,
				PrivateEndpointReferenceCount:   &two,
				Summary: dnsZoneOperatorSummary(
					"privatelink.database.windows.net",
					"private",
					&six,
					0,
					&two,
					&one,
					&two,
				),
				RelatedIDs: []string{privateZoneID, privateEndpointPrimary, privateEndpointReplica},
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) AKS(_ context.Context, tenant string, subscription string) (AksFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	servicePrincipal := "ServicePrincipal"
	systemAssigned := "SystemAssigned"
	freeTier := "Free"
	standardTier := "Standard"
	eastus := "eastus"
	succeeded := "Succeeded"
	kubernetesLegacy := "1.27.9"
	kubernetesOps := "1.29.4"
	nodeRGPublic := "MC_rg-workload_aks-public-legacy_eastus"
	nodeRGOps := "MC_rg-workload_aks-ops-01_eastus"
	publicFQDN := "aks-public-legacy-ef567890.hcp.eastus.azmk8s.io"
	privateFQDN := "aks-ops-01-abcd1234.privatelink.eastus.azmk8s.io"
	opsFQDN := "aks-ops-01-abcd1234.hcp.eastus.azmk8s.io"
	kubenet := "kubenet"
	azure := "azure"
	calico := "calico"
	loadBalancer := "loadBalancer"
	legacyClientID := "99990000-0000-0000-0000-000000000021"
	opsClientID := "99990000-0000-0000-0000-000000000012"
	opsPrincipalID := "99990000-0000-0000-0000-000000000011"
	oidcIssuerURL := "https://oidc.prod-aks.azure.example/aks-ops-01"
	falseValue := false
	trueValue := true
	one := 1
	two := 2
	zero := 0

	publicClusterID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.ContainerService/managedClusters/aks-public-legacy"
	opsClusterID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.ContainerService/managedClusters/aks-ops-01"

	return AksFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		AksClusters: []models.AksClusterAsset{
			{
				ID:                        publicClusterID,
				Name:                      "aks-public-legacy",
				ResourceGroup:             "rg-workload",
				Location:                  &eastus,
				ProvisioningState:         &succeeded,
				KubernetesVersion:         &kubernetesLegacy,
				SKUTier:                   &freeTier,
				NodeResourceGroup:         &nodeRGPublic,
				FQDN:                      &publicFQDN,
				PrivateFQDN:               nil,
				PrivateClusterEnabled:     &falseValue,
				PublicFQDNEnabled:         nil,
				ClusterIdentityType:       &servicePrincipal,
				ClusterPrincipalID:        nil,
				ClusterClientID:           &legacyClientID,
				ClusterIdentityIDs:        []string{},
				AADManaged:                &falseValue,
				AzureRBACEnabled:          &falseValue,
				LocalAccountsDisabled:     &falseValue,
				NetworkPlugin:             &kubenet,
				NetworkPolicy:             nil,
				OutboundType:              &loadBalancer,
				AgentPoolCount:            &one,
				OIDCIssuerEnabled:         &falseValue,
				OIDCIssuerURL:             nil,
				WorkloadIdentityEnabled:   &falseValue,
				AddonNames:                []string{},
				WebAppRoutingEnabled:      &falseValue,
				WebAppRoutingDNSZoneCount: &zero,
				Summary: aksOperatorSummary(
					"aks-public-legacy",
					&kubernetesLegacy,
					&publicFQDN,
					nil,
					&falseValue,
					nil,
					&servicePrincipal,
					&legacyClientID,
					&falseValue,
					&falseValue,
					&falseValue,
					&kubenet,
					nil,
					&loadBalancer,
					&one,
					&falseValue,
					nil,
					&falseValue,
					[]string{},
					&falseValue,
					&zero,
				),
				RelatedIDs: []string{publicClusterID},
			},
			{
				ID:                        opsClusterID,
				Name:                      "aks-ops-01",
				ResourceGroup:             "rg-workload",
				Location:                  &eastus,
				ProvisioningState:         &succeeded,
				KubernetesVersion:         &kubernetesOps,
				SKUTier:                   &standardTier,
				NodeResourceGroup:         &nodeRGOps,
				FQDN:                      &opsFQDN,
				PrivateFQDN:               &privateFQDN,
				PrivateClusterEnabled:     &trueValue,
				PublicFQDNEnabled:         &falseValue,
				ClusterIdentityType:       &systemAssigned,
				ClusterPrincipalID:        &opsPrincipalID,
				ClusterClientID:           &opsClientID,
				ClusterIdentityIDs:        []string{},
				AADManaged:                &trueValue,
				AzureRBACEnabled:          &trueValue,
				LocalAccountsDisabled:     &trueValue,
				NetworkPlugin:             &azure,
				NetworkPolicy:             &calico,
				OutboundType:              &loadBalancer,
				AgentPoolCount:            &two,
				OIDCIssuerEnabled:         &trueValue,
				OIDCIssuerURL:             &oidcIssuerURL,
				WorkloadIdentityEnabled:   &trueValue,
				AddonNames:                []string{"azureKeyvaultSecretsProvider"},
				WebAppRoutingEnabled:      &trueValue,
				WebAppRoutingDNSZoneCount: &one,
				Summary: aksOperatorSummary(
					"aks-ops-01",
					&kubernetesOps,
					&opsFQDN,
					&privateFQDN,
					&trueValue,
					&falseValue,
					&systemAssigned,
					&opsClientID,
					&trueValue,
					&trueValue,
					&trueValue,
					&azure,
					&calico,
					&loadBalancer,
					&two,
					&trueValue,
					&oidcIssuerURL,
					&trueValue,
					[]string{"azureKeyvaultSecretsProvider"},
					&trueValue,
					&one,
				),
				RelatedIDs: []string{
					opsClusterID,
					opsPrincipalID,
				},
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) ApiMgmt(_ context.Context, tenant string, subscription string) (ApiMgmtFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	eastus := "eastus"
	succeeded := "Succeeded"
	developer := "Developer"
	enabled := "Enabled"
	external := "External"
	systemAssigned := "SystemAssigned"
	disabled := "Disabled"
	trueValue := true
	one := 1
	two := 2
	three := 3

	serviceID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.ApiManagement/service/apim-edge-01"
	publicIPID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Network/publicIPAddresses/pip-apim-edge-01"
	workloadPrincipalID := "99990000-0000-0000-0000-000000000001"
	workloadClientID := "99990000-0000-0000-0000-000000000002"

	return ApiMgmtFacts{
		ArtifactIdentityFacts: staticArtifactIdentityFacts(session),
		TenantID:              session.TenantID,
		SubscriptionID:        subscriptionID,
		ApiManagementServices: []models.ApiMgmtServiceAsset{
			{
				ID:                           serviceID,
				Name:                         "apim-edge-01",
				ResourceGroup:                "rg-apps",
				Location:                     &eastus,
				State:                        &succeeded,
				SKUName:                      &developer,
				SKUCapacity:                  &one,
				PublicNetworkAccess:          &enabled,
				VirtualNetworkType:           &external,
				PublicIPAddressID:            &publicIPID,
				PublicIPAddresses:            []string{"52.170.20.30"},
				PrivateIPAddresses:           []string{},
				GatewayHostnames:             []string{"apim-edge-01.azure-api.net", "api.contoso.com"},
				ManagementHostnames:          []string{"apim-edge-01.management.azure-api.net"},
				PortalHostnames:              []string{"portal.apim-edge-01.contoso.com"},
				WorkloadIdentityType:         &systemAssigned,
				WorkloadPrincipalID:          &workloadPrincipalID,
				WorkloadClientID:             &workloadClientID,
				WorkloadIdentityIDs:          []string{},
				GatewayEnabled:               &trueValue,
				DeveloperPortalStatus:        &enabled,
				LegacyPortalStatus:           &disabled,
				APICount:                     &two,
				APISubscriptionRequiredCount: &one,
				SubscriptionCount:            &three,
				ActiveSubscriptionCount:      &two,
				BackendCount:                 &one,
				BackendHostnames:             []string{"orders-internal.contoso.local"},
				PolicyCount:                  &two,
				PolicyControlTypes:           []string{"backend-routing", "conditional-routing", "header-auth", "request-rewrite"},
				NamedValueCount:              &two,
				NamedValueSecretCount:        &one,
				NamedValueKeyVaultCount:      &one,
				Summary: apiMgmtOperatorSummary(
					"apim-edge-01",
					[]string{"apim-edge-01.azure-api.net", "api.contoso.com"},
					[]string{"apim-edge-01.management.azure-api.net"},
					[]string{"portal.apim-edge-01.contoso.com"},
					&enabled,
					&external,
					&developer,
					&systemAssigned,
					&two,
					&one,
					&three,
					&two,
					&one,
					[]string{"orders-internal.contoso.local"},
					&two,
					[]string{"backend-routing", "conditional-routing", "header-auth", "request-rewrite"},
					&two,
					&one,
					&one,
					&trueValue,
					&enabled,
				),
				RelatedIDs: []string{
					serviceID,
					workloadPrincipalID,
					publicIPID,
				},
			},
		},
		Issues: []models.Issue{},
	}, nil
}
