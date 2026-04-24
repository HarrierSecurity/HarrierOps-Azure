package models

type RelayHybridConnectionAsset struct {
	ID                          string   `json:"id"`
	Name                        string   `json:"hybrid_connection"`
	RequiresClientAuthorization *bool    `json:"requires_client_authorization,omitempty"`
	UserMetadata                *string  `json:"user_metadata,omitempty"`
	ListenerCount               *int     `json:"listener_count,omitempty"`
	AppServiceAttachments       []string `json:"app_service_attachments"`
	Summary                     string   `json:"summary"`
	RelatedIDs                  []string `json:"related_ids"`
}

type RelayNamespaceAsset struct {
	ID                     string                       `json:"id"`
	Name                   string                       `json:"namespace"`
	ResourceGroup          string                       `json:"resource_group"`
	Location               *string                      `json:"location,omitempty"`
	SKUName                *string                      `json:"sku_name,omitempty"`
	ProvisioningState      *string                      `json:"provisioning_state,omitempty"`
	ServiceBusEndpoint     *string                      `json:"service_bus_endpoint,omitempty"`
	MetricID               *string                      `json:"metric_id,omitempty"`
	HybridConnectionCount  *int                         `json:"hybrid_connection_count,omitempty"`
	AuthorizationRuleCount *int                         `json:"authorization_rule_count,omitempty"`
	HybridConnections      []RelayHybridConnectionAsset `json:"hybrid_connections"`
	Summary                string                       `json:"summary"`
	RelatedIDs             []string                     `json:"related_ids"`
}

type RelayMetadata = RuntimeCommandMetadata

type RelayOutput struct {
	Findings   []Finding             `json:"findings"`
	Issues     []Issue               `json:"issues"`
	Metadata   RelayMetadata         `json:"metadata"`
	Namespaces []RelayNamespaceAsset `json:"namespaces"`
}
