package models

type KeyVaultAsset struct {
	AccessPolicyCount       int     `json:"access_policy_count"`
	EnableRBACAuthorization bool    `json:"enable_rbac_authorization"`
	ID                      string  `json:"id"`
	Location                *string `json:"location"`
	Name                    string  `json:"name"`
	NetworkDefaultAction    *string `json:"network_default_action"`
	PrivateEndpointEnabled  bool    `json:"private_endpoint_enabled"`
	PublicNetworkAccess     *string `json:"public_network_access"`
	PurgeProtectionEnabled  bool    `json:"purge_protection_enabled"`
	ResourceGroup           string  `json:"resource_group"`
	SKUName                 *string `json:"sku_name"`
	SoftDeleteEnabled       bool    `json:"soft_delete_enabled"`
	TenantID                *string `json:"tenant_id"`
	VaultURI                *string `json:"vault_uri"`
}

type KeyVaultFinding struct {
	Description string   `json:"description"`
	ID          string   `json:"id"`
	RelatedIDs  []string `json:"related_ids,omitempty"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
}

type KeyVaultMetadata = Metadata

type KeyVaultOutput struct {
	Findings  []KeyVaultFinding `json:"findings"`
	Issues    []Issue           `json:"issues"`
	KeyVaults []KeyVaultAsset   `json:"key_vaults"`
	Metadata  KeyVaultMetadata  `json:"metadata"`
}
