package models

type AksClusterAsset struct {
	ID                        string   `json:"id"`
	Name                      string   `json:"name"`
	ResourceGroup             string   `json:"resource_group"`
	Location                  *string  `json:"location"`
	ProvisioningState         *string  `json:"provisioning_state"`
	KubernetesVersion         *string  `json:"kubernetes_version"`
	SKUTier                   *string  `json:"sku_tier"`
	NodeResourceGroup         *string  `json:"node_resource_group"`
	FQDN                      *string  `json:"fqdn"`
	PrivateFQDN               *string  `json:"private_fqdn"`
	PrivateClusterEnabled     *bool    `json:"private_cluster_enabled"`
	PublicFQDNEnabled         *bool    `json:"public_fqdn_enabled"`
	ClusterIdentityType       *string  `json:"cluster_identity_type"`
	ClusterPrincipalID        *string  `json:"cluster_principal_id"`
	ClusterClientID           *string  `json:"cluster_client_id"`
	ClusterIdentityIDs        []string `json:"cluster_identity_ids"`
	AADManaged                *bool    `json:"aad_managed"`
	AzureRBACEnabled          *bool    `json:"azure_rbac_enabled"`
	LocalAccountsDisabled     *bool    `json:"local_accounts_disabled"`
	NetworkPlugin             *string  `json:"network_plugin"`
	NetworkPolicy             *string  `json:"network_policy"`
	OutboundType              *string  `json:"outbound_type"`
	AgentPoolCount            *int     `json:"agent_pool_count"`
	OIDCIssuerEnabled         *bool    `json:"oidc_issuer_enabled"`
	OIDCIssuerURL             *string  `json:"oidc_issuer_url"`
	WorkloadIdentityEnabled   *bool    `json:"workload_identity_enabled"`
	AddonNames                []string `json:"addon_names"`
	WebAppRoutingEnabled      *bool    `json:"web_app_routing_enabled"`
	WebAppRoutingDNSZoneCount *int     `json:"web_app_routing_dns_zone_count"`
	Summary                   string   `json:"summary"`
	RelatedIDs                []string `json:"related_ids"`
}

type AksMetadata = RuntimeCommandMetadata

type AksOutput struct {
	AksClusters []AksClusterAsset `json:"aks_clusters"`
	Findings    []Finding         `json:"findings"`
	Issues      []Issue           `json:"issues"`
	Metadata    AksMetadata       `json:"metadata"`
}
