package models

type VmssAsset struct {
	ApplicationGatewayBackendPoolCount int      `json:"application_gateway_backend_pool_count"`
	ClientID                           *string  `json:"client_id"`
	ID                                 string   `json:"id"`
	IdentityIDs                        []string `json:"identity_ids"`
	IdentityType                       *string  `json:"identity_type"`
	InboundNATPoolCount                int      `json:"inbound_nat_pool_count"`
	InstanceCount                      *int     `json:"instance_count"`
	LoadBalancerBackendPoolCount       int      `json:"load_balancer_backend_pool_count"`
	Location                           string   `json:"location"`
	Name                               string   `json:"name"`
	NICConfigurationCount              int      `json:"nic_configuration_count"`
	OrchestrationMode                  *string  `json:"orchestration_mode"`
	Overprovision                      *bool    `json:"overprovision"`
	PrincipalID                        *string  `json:"principal_id"`
	PublicIPConfigurationCount         int      `json:"public_ip_configuration_count"`
	RelatedIDs                         []string `json:"related_ids"`
	ResourceGroup                      string   `json:"resource_group"`
	SinglePlacementGroup               *bool    `json:"single_placement_group"`
	SKUName                            *string  `json:"sku_name"`
	SubnetIDs                          []string `json:"subnet_ids"`
	Summary                            string   `json:"summary"`
	UpgradeMode                        *string  `json:"upgrade_mode"`
	ZoneBalance                        *bool    `json:"zone_balance"`
	Zones                              []string `json:"zones"`
}

type VmssOutput struct {
	Findings   []Finding   `json:"findings"`
	Issues     []Issue     `json:"issues"`
	Metadata   Metadata    `json:"metadata"`
	VmssAssets []VmssAsset `json:"vmss_assets"`
}
