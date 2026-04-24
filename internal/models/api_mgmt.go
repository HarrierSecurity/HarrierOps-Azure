package models

type ApiMgmtServiceAsset struct {
	ID                           string   `json:"id"`
	Name                         string   `json:"name"`
	ResourceGroup                string   `json:"resource_group"`
	Location                     *string  `json:"location"`
	State                        *string  `json:"state"`
	SKUName                      *string  `json:"sku_name"`
	SKUCapacity                  *int     `json:"sku_capacity"`
	PublicNetworkAccess          *string  `json:"public_network_access"`
	VirtualNetworkType           *string  `json:"virtual_network_type"`
	PublicIPAddressID            *string  `json:"public_ip_address_id"`
	PublicIPAddresses            []string `json:"public_ip_addresses"`
	PrivateIPAddresses           []string `json:"private_ip_addresses"`
	GatewayHostnames             []string `json:"gateway_hostnames"`
	ManagementHostnames          []string `json:"management_hostnames"`
	PortalHostnames              []string `json:"portal_hostnames"`
	WorkloadIdentityType         *string  `json:"workload_identity_type"`
	WorkloadPrincipalID          *string  `json:"workload_principal_id"`
	WorkloadClientID             *string  `json:"workload_client_id"`
	WorkloadIdentityIDs          []string `json:"workload_identity_ids"`
	GatewayEnabled               *bool    `json:"gateway_enabled"`
	DeveloperPortalStatus        *string  `json:"developer_portal_status"`
	LegacyPortalStatus           *string  `json:"legacy_portal_status"`
	APICount                     *int     `json:"api_count"`
	APISubscriptionRequiredCount *int     `json:"api_subscription_required_count"`
	SubscriptionCount            *int     `json:"subscription_count"`
	ActiveSubscriptionCount      *int     `json:"active_subscription_count"`
	BackendCount                 *int     `json:"backend_count"`
	BackendHostnames             []string `json:"backend_hostnames"`
	PolicyCount                  *int     `json:"policy_count"`
	PolicyControlTypes           []string `json:"policy_control_types"`
	NamedValueCount              *int     `json:"named_value_count"`
	NamedValueSecretCount        *int     `json:"named_value_secret_count"`
	NamedValueKeyVaultCount      *int     `json:"named_value_key_vault_count"`
	Summary                      string   `json:"summary"`
	RelatedIDs                   []string `json:"related_ids"`
}

type ApiMgmtMetadata = RuntimeCommandMetadata

type ApiMgmtOutput struct {
	ApiManagementServices []ApiMgmtServiceAsset `json:"api_management_services"`
	Findings              []Finding             `json:"findings"`
	Issues                []Issue               `json:"issues"`
	Metadata              ApiMgmtMetadata       `json:"metadata"`
}
