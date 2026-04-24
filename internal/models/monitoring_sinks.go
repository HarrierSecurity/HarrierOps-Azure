package models

type MonitoringSinkReference struct {
	SourceCommand     string  `json:"source_command"`
	SourceResourceID  string  `json:"source_resource_id"`
	SourceName        string  `json:"source_name"`
	ReferenceName     *string `json:"reference_name,omitempty"`
	ReferenceType     string  `json:"reference_type"`
	DestinationDetail *string `json:"destination_detail,omitempty"`
}

type MonitoringSinkAsset struct {
	ID               string                    `json:"id"`
	Name             string                    `json:"name"`
	Kind             string                    `json:"kind"`
	ResourceType     string                    `json:"resource_type"`
	ResourceGroup    string                    `json:"resource_group"`
	Location         string                    `json:"location"`
	VisibilitySource string                    `json:"visibility_source"`
	SentinelEnabled  *bool                     `json:"sentinel_enabled,omitempty"`
	References       []MonitoringSinkReference `json:"references"`
	ReferenceCount   int                       `json:"reference_count"`
	Summary          string                    `json:"summary"`
	RelatedIDs       []string                  `json:"related_ids"`
}

type MonitoringSinksOutput struct {
	Sinks    []MonitoringSinkAsset  `json:"sinks"`
	Findings []Finding              `json:"findings"`
	Issues   []Issue                `json:"issues"`
	Metadata RuntimeCommandMetadata `json:"metadata"`
}
