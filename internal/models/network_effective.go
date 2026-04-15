package models

type NetworkEffectiveSummary struct {
	AssetID              string   `json:"asset_id"`
	AssetName            string   `json:"asset_name"`
	ConstrainedPorts     []string `json:"constrained_ports"`
	EffectiveExposure    string   `json:"effective_exposure"`
	Endpoint             string   `json:"endpoint"`
	EndpointType         string   `json:"endpoint_type"`
	InternetExposedPorts []string `json:"internet_exposed_ports"`
	ObservedPaths        []string `json:"observed_paths"`
	RelatedIDs           []string `json:"related_ids"`
	Summary              string   `json:"summary"`
}

type NetworkEffectiveOutput struct {
	EffectiveExposures []NetworkEffectiveSummary `json:"effective_exposures"`
	Findings           []Finding                 `json:"findings"`
	Issues             []Issue                   `json:"issues"`
	Metadata           ScopedCommandMetadata     `json:"metadata"`
}
