package models

type NetworkCommandMetadata struct {
	Command        string  `json:"command"`
	GeneratedAt    string  `json:"generated_at"`
	SchemaVersion  string  `json:"schema_version"`
	SubscriptionID *string `json:"subscription_id"`
	TenantID       *string `json:"tenant_id"`
	TokenSource    *string `json:"token_source"`
}

type NetworkPortSummary struct {
	AllowSourceSummary string   `json:"allow_source_summary"`
	AssetID            string   `json:"asset_id"`
	AssetName          string   `json:"asset_name"`
	Endpoint           string   `json:"endpoint"`
	ExposureConfidence string   `json:"exposure_confidence"`
	Port               string   `json:"port"`
	Protocol           string   `json:"protocol"`
	RelatedIDs         []string `json:"related_ids"`
	Summary            string   `json:"summary"`
}

type NetworkPortsOutput struct {
	Findings     []Finding              `json:"findings"`
	Issues       []Issue                `json:"issues"`
	Metadata     NetworkCommandMetadata `json:"metadata"`
	NetworkPorts []NetworkPortSummary   `json:"network_ports"`
}
