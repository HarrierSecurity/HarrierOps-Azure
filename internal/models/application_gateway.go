package models

type ApplicationGatewayAsset struct {
	BackendPoolCount           int      `json:"backend_pool_count"`
	BackendTargetCount         int      `json:"backend_target_count"`
	FirewallPolicyID           *string  `json:"firewall_policy_id"`
	ID                         string   `json:"id"`
	ListenerCount              int      `json:"listener_count"`
	Location                   *string  `json:"location"`
	Name                       string   `json:"name"`
	PrivateFrontendCount       int      `json:"private_frontend_count"`
	PrivateFrontendIPs         []string `json:"private_frontend_ips"`
	PublicFrontendCount        int      `json:"public_frontend_count"`
	PublicIPAddressIDs         []string `json:"public_ip_address_ids"`
	PublicIPAddresses          []string `json:"public_ip_addresses"`
	RedirectConfigurationCount int      `json:"redirect_configuration_count"`
	RelatedIDs                 []string `json:"related_ids"`
	RequestRoutingRuleCount    int      `json:"request_routing_rule_count"`
	ResourceGroup              string   `json:"resource_group"`
	RewriteRuleSetCount        int      `json:"rewrite_rule_set_count"`
	SKUName                    *string  `json:"sku_name"`
	SKUTier                    *string  `json:"sku_tier"`
	State                      *string  `json:"state"`
	SubnetIDs                  []string `json:"subnet_ids"`
	Summary                    string   `json:"summary"`
	URLPathMapCount            int      `json:"url_path_map_count"`
	WAFEnabled                 *bool    `json:"waf_enabled"`
	WAFMode                    *string  `json:"waf_mode"`
}

type ApplicationGatewayMetadata = RuntimeCommandMetadata

type ApplicationGatewayOutput struct {
	ApplicationGateways []ApplicationGatewayAsset  `json:"application_gateways"`
	Findings            []Finding                  `json:"findings"`
	Issues              []Issue                    `json:"issues"`
	Metadata            ApplicationGatewayMetadata `json:"metadata"`
}
