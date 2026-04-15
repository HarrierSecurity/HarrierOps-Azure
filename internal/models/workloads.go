package models

type WorkloadSummary struct {
	AssetID             string   `json:"asset_id"`
	AssetKind           string   `json:"asset_kind"`
	AssetName           string   `json:"asset_name"`
	Endpoints           []string `json:"endpoints"`
	ExposureFamilies    []string `json:"exposure_families"`
	IdentityClientID    *string  `json:"identity_client_id"`
	IdentityIDs         []string `json:"identity_ids"`
	IdentityPrincipalID *string  `json:"identity_principal_id"`
	IdentityType        *string  `json:"identity_type"`
	IngressPaths        []string `json:"ingress_paths"`
	Location            string   `json:"location"`
	RelatedIDs          []string `json:"related_ids"`
	ResourceGroup       string   `json:"resource_group"`
	Summary             string   `json:"summary"`
}

type WorkloadsOutput struct {
	Metadata  ScopedCommandMetadata `json:"metadata"`
	Workloads []WorkloadSummary     `json:"workloads"`
	Findings  []Finding             `json:"findings"`
	Issues    []Issue               `json:"issues"`
}
