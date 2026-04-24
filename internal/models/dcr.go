package models

type DCRDataSource struct {
	Name                    string   `json:"name"`
	Type                    string   `json:"type"`
	Streams                 []string `json:"streams"`
	TransformKqlPresent     bool     `json:"transform_kql_present"`
	TransformKqlFingerprint *string  `json:"transform_kql_fingerprint,omitempty"`
	TransformKqlLength      *int     `json:"transform_kql_length,omitempty"`
}

type DCRDataFlow struct {
	Streams                 []string `json:"streams"`
	Destinations            []string `json:"destinations"`
	OutputStream            *string  `json:"output_stream,omitempty"`
	BuiltInTransform        *string  `json:"built_in_transform,omitempty"`
	TransformKqlPresent     bool     `json:"transform_kql_present"`
	TransformKqlFingerprint *string  `json:"transform_kql_fingerprint,omitempty"`
	TransformKqlLength      *int     `json:"transform_kql_length,omitempty"`
}

type DCRDestination struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	ResourceID *string `json:"resource_id,omitempty"`
	Detail     *string `json:"detail,omitempty"`
}

type DCRAssociation struct {
	ID                       string  `json:"id"`
	Name                     string  `json:"name"`
	TargetID                 string  `json:"target_id"`
	DataCollectionRuleID     *string `json:"data_collection_rule_id,omitempty"`
	DataCollectionEndpointID *string `json:"data_collection_endpoint_id,omitempty"`
	Description              *string `json:"description,omitempty"`
}

type DCRAsset struct {
	ID                       string           `json:"id"`
	Name                     string           `json:"name"`
	ResourceGroup            string           `json:"resource_group"`
	Location                 string           `json:"location"`
	Kind                     *string          `json:"kind,omitempty"`
	Description              *string          `json:"description,omitempty"`
	DataCollectionEndpointID *string          `json:"data_collection_endpoint_id,omitempty"`
	DataSources              []DCRDataSource  `json:"data_sources"`
	DataFlows                []DCRDataFlow    `json:"data_flows"`
	Destinations             []DCRDestination `json:"destinations"`
	Associations             []DCRAssociation `json:"associations"`
	DataSourceTypes          []string         `json:"data_source_types"`
	Streams                  []string         `json:"streams"`
	HighSignalStreams        []string         `json:"high_signal_streams"`
	DestinationTypes         []string         `json:"destination_types"`
	TransformationCount      int              `json:"transformation_count"`
	AssociationCount         int              `json:"association_count"`
	RelatedIDs               []string         `json:"related_ids"`
	Summary                  string           `json:"summary"`
}

type DCRMetadata = RuntimeCommandMetadata

type DCROutput struct {
	DCRs     []DCRAsset  `json:"dcrs"`
	Findings []Finding   `json:"findings"`
	Issues   []Issue     `json:"issues"`
	Metadata DCRMetadata `json:"metadata"`
}
