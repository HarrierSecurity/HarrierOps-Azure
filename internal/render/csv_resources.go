package render

import (
	"fmt"

	"harrierops-azure/internal/models"
)

func acrCSV(payload models.AcrOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Registries))
	for _, registry := range payload.Registries {
		rows = append(rows, []string{
			boolPtrString(registry.AdminUserEnabled),
			boolPtrString(registry.AnonymousPullEnabled),
			intPtrString(registry.BroadWebhookScopeCount),
			boolPtrString(registry.DataEndpointEnabled),
			intPtrString(registry.EnabledWebhookCount),
			registry.ID,
			valueOrEmpty(registry.Location),
			valueOrEmpty(registry.LoginServer),
			registry.Name,
			valueOrEmpty(registry.NetworkRuleBypassOptions),
			valueOrEmpty(registry.NetworkRuleDefaultAction),
			intPtrString(registry.PrivateEndpointConnectionCount),
			valueOrEmpty(registry.PublicNetworkAccess),
			valueOrEmpty(registry.QuarantinePolicyStatus),
			jsonStringSlice(registry.RelatedIDs),
			intPtrString(registry.ReplicationCount),
			jsonStringSlice(registry.ReplicationRegions),
			registry.ResourceGroup,
			intPtrString(registry.RetentionPolicyDays),
			valueOrEmpty(registry.RetentionPolicyStatus),
			valueOrEmpty(registry.SKUName),
			valueOrEmpty(registry.State),
			registry.Summary,
			valueOrEmpty(registry.TrustPolicyStatus),
			valueOrEmpty(registry.TrustPolicyType),
			jsonStringSlice(registry.WebhookActionTypes),
			intPtrString(registry.WebhookCount),
			valueOrEmpty(registry.WorkloadClientID),
			jsonStringSlice(registry.WorkloadIdentityIDs),
			valueOrEmpty(registry.WorkloadIdentityType),
			valueOrEmpty(registry.WorkloadPrincipalID),
		})
	}

	return encodeCSV([]string{
		"admin_user_enabled",
		"anonymous_pull_enabled",
		"broad_webhook_scope_count",
		"data_endpoint_enabled",
		"enabled_webhook_count",
		"id",
		"location",
		"login_server",
		"name",
		"network_rule_bypass_options",
		"network_rule_default_action",
		"private_endpoint_connection_count",
		"public_network_access",
		"quarantine_policy_status",
		"related_ids",
		"replication_count",
		"replication_regions",
		"resource_group",
		"retention_policy_days",
		"retention_policy_status",
		"sku_name",
		"state",
		"summary",
		"trust_policy_status",
		"trust_policy_type",
		"webhook_action_types",
		"webhook_count",
		"workload_client_id",
		"workload_identity_ids",
		"workload_identity_type",
		"workload_principal_id",
	}, rows)
}

func databasesCSV(payload models.DatabasesOutput) (string, error) {
	rows := make([][]string, 0, len(payload.DatabaseServers))
	for _, server := range payload.DatabaseServers {
		rows = append(rows, []string{
			intPtrString(server.DatabaseCount),
			valueOrEmpty(server.DelegatedSubnetResourceID),
			server.Engine,
			valueOrEmpty(server.FullyQualifiedDomainName),
			valueOrEmpty(server.HighAvailabilityMode),
			server.ID,
			valueOrEmpty(server.Location),
			valueOrEmpty(server.MinimalTLSVersion),
			server.Name,
			valueOrEmpty(server.PrivateDNSZoneResourceID),
			valueOrEmpty(server.PublicNetworkAccess),
			jsonStringSlice(server.RelatedIDs),
			server.ResourceGroup,
			valueOrEmpty(server.ServerVersion),
			valueOrEmpty(server.State),
			server.Summary,
			jsonStringSlice(server.UserDatabaseNames),
			valueOrEmpty(server.WorkloadClientID),
			jsonStringSlice(server.WorkloadIdentityIDs),
			valueOrEmpty(server.WorkloadIdentityType),
			valueOrEmpty(server.WorkloadPrincipalID),
		})
	}

	return encodeCSV([]string{
		"database_count",
		"delegated_subnet_resource_id",
		"engine",
		"fully_qualified_domain_name",
		"high_availability_mode",
		"id",
		"location",
		"minimal_tls_version",
		"name",
		"private_dns_zone_resource_id",
		"public_network_access",
		"related_ids",
		"resource_group",
		"server_version",
		"state",
		"summary",
		"user_database_names",
		"workload_client_id",
		"workload_identity_ids",
		"workload_identity_type",
		"workload_principal_id",
	}, rows)
}

func keyVaultCSV(payload models.KeyVaultOutput) (string, error) {
	rows := make([][]string, 0, len(payload.KeyVaults))
	for _, vault := range payload.KeyVaults {
		rows = append(rows, []string{
			fmt.Sprintf("%d", vault.AccessPolicyCount),
			fmt.Sprintf("%t", vault.EnableRBACAuthorization),
			vault.ID,
			valueOrEmpty(vault.Location),
			vault.Name,
			valueOrEmpty(vault.NetworkDefaultAction),
			fmt.Sprintf("%t", vault.PrivateEndpointEnabled),
			valueOrEmpty(vault.PublicNetworkAccess),
			fmt.Sprintf("%t", vault.PurgeProtectionEnabled),
			vault.ResourceGroup,
			valueOrEmpty(vault.SKUName),
			fmt.Sprintf("%t", vault.SoftDeleteEnabled),
			valueOrEmpty(vault.TenantID),
			valueOrEmpty(vault.VaultURI),
		})
	}

	return encodeCSV([]string{
		"access_policy_count",
		"enable_rbac_authorization",
		"id",
		"location",
		"name",
		"network_default_action",
		"private_endpoint_enabled",
		"public_network_access",
		"purge_protection_enabled",
		"resource_group",
		"sku_name",
		"soft_delete_enabled",
		"tenant_id",
		"vault_uri",
	}, rows)
}

func storageCSV(payload models.StorageOutput) (string, error) {
	rows := make([][]string, 0, len(payload.StorageAssets))
	for _, asset := range payload.StorageAssets {
		rows = append(rows, []string{
			boolPtrString(asset.AllowSharedKeyAccess),
			jsonStringSlice(asset.AnonymousAccessIndicators),
			intPtrString(asset.ContainerCount),
			valueOrEmpty(asset.DNSEndpointType),
			intPtrString(asset.FileShareCount),
			boolPtrString(asset.HTTPSTrafficOnlyEnabled),
			asset.ID,
			boolPtrString(asset.IsHNSEnabled),
			boolPtrString(asset.IsSFTPEnabled),
			valueOrEmpty(asset.Location),
			valueOrEmpty(asset.MinimumTLSVersion),
			asset.Name,
			valueOrEmpty(asset.NetworkDefaultAction),
			boolPtrString(asset.NFSV3Enabled),
			fmt.Sprintf("%t", asset.PrivateEndpointEnabled),
			fmt.Sprintf("%t", asset.PublicAccess),
			valueOrEmpty(asset.PublicNetworkAccess),
			intPtrString(asset.QueueCount),
			asset.ResourceGroup,
			intPtrString(asset.TableCount),
		})
	}

	return encodeCSV([]string{
		"allow_shared_key_access",
		"anonymous_access_indicators",
		"container_count",
		"dns_endpoint_type",
		"file_share_count",
		"https_traffic_only_enabled",
		"id",
		"is_hns_enabled",
		"is_sftp_enabled",
		"location",
		"minimum_tls_version",
		"name",
		"network_default_action",
		"nfs_v3_enabled",
		"private_endpoint_enabled",
		"public_access",
		"public_network_access",
		"queue_count",
		"resource_group",
		"table_count",
	}, rows)
}

func snapshotsDisksCSV(payload models.SnapshotsDisksOutput) (string, error) {
	rows := make([][]string, 0, len(payload.SnapshotDiskAssets))
	for _, asset := range payload.SnapshotDiskAssets {
		rows = append(rows, []string{
			asset.ID,
			asset.Name,
			asset.AssetKind,
			asset.ResourceGroup,
			valueOrEmpty(asset.Location),
			valueOrEmpty(asset.DiskRole),
			asset.AttachmentState,
			valueOrEmpty(asset.AttachedToID),
			valueOrEmpty(asset.AttachedToName),
			valueOrEmpty(asset.SourceResourceID),
			valueOrEmpty(asset.SourceResourceName),
			valueOrEmpty(asset.SourceResourceKind),
			valueOrEmpty(asset.OSType),
			intPtrString(asset.SizeGB),
			valueOrEmpty(asset.TimeCreated),
			boolPtrString(asset.Incremental),
			valueOrEmpty(asset.NetworkAccessPolicy),
			valueOrEmpty(asset.PublicNetworkAccess),
			valueOrEmpty(asset.DiskAccessID),
			intPtrString(asset.MaxShares),
			valueOrEmpty(asset.EncryptionType),
			valueOrEmpty(asset.DiskEncryptionSetID),
			asset.Summary,
			jsonStringSlice(asset.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"name",
		"asset_kind",
		"resource_group",
		"location",
		"disk_role",
		"attachment_state",
		"attached_to_id",
		"attached_to_name",
		"source_resource_id",
		"source_resource_name",
		"source_resource_kind",
		"os_type",
		"size_gb",
		"time_created",
		"incremental",
		"network_access_policy",
		"public_network_access",
		"disk_access_id",
		"max_shares",
		"encryption_type",
		"disk_encryption_set_id",
		"summary",
		"related_ids",
	}, rows)
}

func applicationGatewayCSV(payload models.ApplicationGatewayOutput) (string, error) {
	rows := make([][]string, 0, len(payload.ApplicationGateways))
	for _, gateway := range payload.ApplicationGateways {
		rows = append(rows, []string{
			intString(gateway.BackendPoolCount),
			intString(gateway.BackendTargetCount),
			valueOrEmpty(gateway.FirewallPolicyID),
			gateway.ID,
			intString(gateway.ListenerCount),
			valueOrEmpty(gateway.Location),
			gateway.Name,
			intString(gateway.PrivateFrontendCount),
			jsonStringSlice(gateway.PrivateFrontendIPs),
			intString(gateway.PublicFrontendCount),
			jsonStringSlice(gateway.PublicIPAddressIDs),
			jsonStringSlice(gateway.PublicIPAddresses),
			intString(gateway.RedirectConfigurationCount),
			jsonStringSlice(gateway.RelatedIDs),
			intString(gateway.RequestRoutingRuleCount),
			gateway.ResourceGroup,
			intString(gateway.RewriteRuleSetCount),
			valueOrEmpty(gateway.SKUName),
			valueOrEmpty(gateway.SKUTier),
			valueOrEmpty(gateway.State),
			jsonStringSlice(gateway.SubnetIDs),
			gateway.Summary,
			intString(gateway.URLPathMapCount),
			boolPtrString(gateway.WAFEnabled),
			valueOrEmpty(gateway.WAFMode),
		})
	}

	return encodeCSV([]string{
		"backend_pool_count",
		"backend_target_count",
		"firewall_policy_id",
		"id",
		"listener_count",
		"location",
		"name",
		"private_frontend_count",
		"private_frontend_ips",
		"public_frontend_count",
		"public_ip_address_ids",
		"public_ip_addresses",
		"redirect_configuration_count",
		"related_ids",
		"request_routing_rule_count",
		"resource_group",
		"rewrite_rule_set_count",
		"sku_name",
		"sku_tier",
		"state",
		"subnet_ids",
		"summary",
		"url_path_map_count",
		"waf_enabled",
		"waf_mode",
	}, rows)
}

func dnsCSV(payload models.DnsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.DNSZones))
	for _, zone := range payload.DNSZones {
		rows = append(rows, []string{
			zone.ID,
			intPtrString(zone.LinkedVirtualNetworkCount),
			valueOrEmpty(zone.Location),
			intPtrString(zone.MaxRecordSetCount),
			zone.Name,
			jsonStringSlice(zone.NameServers),
			intPtrString(zone.PrivateEndpointReferenceCount),
			intPtrString(zone.RecordSetCount),
			intPtrString(zone.RegistrationVirtualNetworkCount),
			jsonStringSlice(zone.RelatedIDs),
			zone.ResourceGroup,
			zone.Summary,
			zone.ZoneKind,
		})
	}

	return encodeCSV([]string{
		"id",
		"linked_virtual_network_count",
		"location",
		"max_record_set_count",
		"name",
		"name_servers",
		"private_endpoint_reference_count",
		"record_set_count",
		"registration_virtual_network_count",
		"related_ids",
		"resource_group",
		"summary",
		"zone_kind",
	}, rows)
}

func aksCSV(payload models.AksOutput) (string, error) {
	rows := make([][]string, 0, len(payload.AksClusters))
	for _, cluster := range payload.AksClusters {
		rows = append(rows, []string{
			boolPtrString(cluster.AADManaged),
			jsonStringSlice(cluster.AddonNames),
			intPtrString(cluster.AgentPoolCount),
			boolPtrString(cluster.AzureRBACEnabled),
			valueOrEmpty(cluster.ClusterClientID),
			jsonStringSlice(cluster.ClusterIdentityIDs),
			valueOrEmpty(cluster.ClusterIdentityType),
			valueOrEmpty(cluster.ClusterPrincipalID),
			valueOrEmpty(cluster.FQDN),
			cluster.ID,
			valueOrEmpty(cluster.KubernetesVersion),
			boolPtrString(cluster.LocalAccountsDisabled),
			valueOrEmpty(cluster.Location),
			cluster.Name,
			valueOrEmpty(cluster.NetworkPlugin),
			valueOrEmpty(cluster.NetworkPolicy),
			valueOrEmpty(cluster.NodeResourceGroup),
			boolPtrString(cluster.OIDCIssuerEnabled),
			valueOrEmpty(cluster.OIDCIssuerURL),
			valueOrEmpty(cluster.OutboundType),
			boolPtrString(cluster.PrivateClusterEnabled),
			valueOrEmpty(cluster.PrivateFQDN),
			valueOrEmpty(cluster.ProvisioningState),
			boolPtrString(cluster.PublicFQDNEnabled),
			jsonStringSlice(cluster.RelatedIDs),
			cluster.ResourceGroup,
			valueOrEmpty(cluster.SKUTier),
			cluster.Summary,
			boolPtrString(cluster.WebAppRoutingEnabled),
			intPtrString(cluster.WebAppRoutingDNSZoneCount),
			boolPtrString(cluster.WorkloadIdentityEnabled),
		})
	}

	return encodeCSV([]string{
		"aad_managed",
		"addon_names",
		"agent_pool_count",
		"azure_rbac_enabled",
		"cluster_client_id",
		"cluster_identity_ids",
		"cluster_identity_type",
		"cluster_principal_id",
		"fqdn",
		"id",
		"kubernetes_version",
		"local_accounts_disabled",
		"location",
		"name",
		"network_plugin",
		"network_policy",
		"node_resource_group",
		"oidc_issuer_enabled",
		"oidc_issuer_url",
		"outbound_type",
		"private_cluster_enabled",
		"private_fqdn",
		"provisioning_state",
		"public_fqdn_enabled",
		"related_ids",
		"resource_group",
		"sku_tier",
		"summary",
		"web_app_routing_enabled",
		"web_app_routing_dns_zone_count",
		"workload_identity_enabled",
	}, rows)
}

func apiMgmtCSV(payload models.ApiMgmtOutput) (string, error) {
	rows := make([][]string, 0, len(payload.ApiManagementServices))
	for _, service := range payload.ApiManagementServices {
		rows = append(rows, []string{
			intPtrString(service.ActiveSubscriptionCount),
			intPtrString(service.APICount),
			intPtrString(service.APISubscriptionRequiredCount),
			jsonStringSlice(service.BackendHostnames),
			intPtrString(service.BackendCount),
			valueOrEmpty(service.DeveloperPortalStatus),
			boolPtrString(service.GatewayEnabled),
			jsonStringSlice(service.GatewayHostnames),
			service.ID,
			valueOrEmpty(service.LegacyPortalStatus),
			valueOrEmpty(service.Location),
			jsonStringSlice(service.ManagementHostnames),
			service.Name,
			intPtrString(service.NamedValueCount),
			intPtrString(service.NamedValueKeyVaultCount),
			intPtrString(service.NamedValueSecretCount),
			jsonStringSlice(service.PortalHostnames),
			valueOrEmpty(service.PublicIPAddressID),
			jsonStringSlice(service.PrivateIPAddresses),
			jsonStringSlice(service.PublicIPAddresses),
			valueOrEmpty(service.PublicNetworkAccess),
			jsonStringSlice(service.RelatedIDs),
			service.ResourceGroup,
			intPtrString(service.SKUCapacity),
			valueOrEmpty(service.SKUName),
			valueOrEmpty(service.State),
			intPtrString(service.SubscriptionCount),
			service.Summary,
			valueOrEmpty(service.VirtualNetworkType),
			valueOrEmpty(service.WorkloadClientID),
			jsonStringSlice(service.WorkloadIdentityIDs),
			valueOrEmpty(service.WorkloadIdentityType),
			valueOrEmpty(service.WorkloadPrincipalID),
		})
	}

	return encodeCSV([]string{
		"active_subscription_count",
		"api_count",
		"api_subscription_required_count",
		"backend_hostnames",
		"backend_count",
		"developer_portal_status",
		"gateway_enabled",
		"gateway_hostnames",
		"id",
		"legacy_portal_status",
		"location",
		"management_hostnames",
		"name",
		"named_value_count",
		"named_value_key_vault_count",
		"named_value_secret_count",
		"portal_hostnames",
		"public_ip_address_id",
		"private_ip_addresses",
		"public_ip_addresses",
		"public_network_access",
		"related_ids",
		"resource_group",
		"sku_capacity",
		"sku_name",
		"state",
		"subscription_count",
		"summary",
		"virtual_network_type",
		"workload_client_id",
		"workload_identity_ids",
		"workload_identity_type",
		"workload_principal_id",
	}, rows)
}
