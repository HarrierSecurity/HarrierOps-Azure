package models

type AcrRegistryAsset struct {
	ID                             string   `json:"id"`
	Name                           string   `json:"name"`
	ResourceGroup                  string   `json:"resource_group"`
	Location                       *string  `json:"location"`
	State                          *string  `json:"state"`
	LoginServer                    *string  `json:"login_server"`
	SKUName                        *string  `json:"sku_name"`
	PublicNetworkAccess            *string  `json:"public_network_access"`
	NetworkRuleDefaultAction       *string  `json:"network_rule_default_action"`
	NetworkRuleBypassOptions       *string  `json:"network_rule_bypass_options"`
	AdminUserEnabled               *bool    `json:"admin_user_enabled"`
	AnonymousPullEnabled           *bool    `json:"anonymous_pull_enabled"`
	DataEndpointEnabled            *bool    `json:"data_endpoint_enabled"`
	PrivateEndpointConnectionCount *int     `json:"private_endpoint_connection_count"`
	WebhookCount                   *int     `json:"webhook_count"`
	EnabledWebhookCount            *int     `json:"enabled_webhook_count"`
	WebhookActionTypes             []string `json:"webhook_action_types"`
	BroadWebhookScopeCount         *int     `json:"broad_webhook_scope_count"`
	ReplicationCount               *int     `json:"replication_count"`
	ReplicationRegions             []string `json:"replication_regions"`
	QuarantinePolicyStatus         *string  `json:"quarantine_policy_status"`
	RetentionPolicyStatus          *string  `json:"retention_policy_status"`
	RetentionPolicyDays            *int     `json:"retention_policy_days"`
	TrustPolicyStatus              *string  `json:"trust_policy_status"`
	TrustPolicyType                *string  `json:"trust_policy_type"`
	WorkloadIdentityType           *string  `json:"workload_identity_type"`
	WorkloadPrincipalID            *string  `json:"workload_principal_id"`
	WorkloadClientID               *string  `json:"workload_client_id"`
	WorkloadIdentityIDs            []string `json:"workload_identity_ids"`
	Summary                        string   `json:"summary"`
	RelatedIDs                     []string `json:"related_ids"`
}

type AcrMetadata = RuntimeCommandMetadata

type AcrOutput struct {
	Registries []AcrRegistryAsset `json:"registries"`
	Findings   []Finding          `json:"findings"`
	Issues     []Issue            `json:"issues"`
	Metadata   AcrMetadata        `json:"metadata"`
}
