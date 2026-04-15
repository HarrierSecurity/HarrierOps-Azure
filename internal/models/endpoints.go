package models

type EndpointSummary struct {
	Endpoint        string   `json:"endpoint"`
	EndpointType    string   `json:"endpoint_type"`
	ExposureFamily  string   `json:"exposure_family"`
	IngressPath     string   `json:"ingress_path"`
	RelatedIDs      []string `json:"related_ids"`
	SourceAssetID   string   `json:"source_asset_id"`
	SourceAssetKind string   `json:"source_asset_kind"`
	SourceAssetName string   `json:"source_asset_name"`
	Summary         string   `json:"summary"`
}

type EndpointsOutput struct {
	Endpoints []EndpointSummary     `json:"endpoints"`
	Findings  []Finding             `json:"findings"`
	Issues    []Issue               `json:"issues"`
	Metadata  ScopedCommandMetadata `json:"metadata"`
}
