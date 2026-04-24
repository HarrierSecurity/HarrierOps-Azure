package models

type AppInsightsComponent struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	ResourceGroup       string   `json:"resource_group"`
	Location            string   `json:"location"`
	Kind                *string  `json:"kind,omitempty"`
	ApplicationType     *string  `json:"application_type,omitempty"`
	WorkspaceResourceID *string  `json:"workspace_resource_id,omitempty"`
	IngestionMode       *string  `json:"ingestion_mode,omitempty"`
	Summary             string   `json:"summary"`
	RelatedIDs          []string `json:"related_ids"`
}

type AppInsightsAppTarget struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Kind                  string   `json:"kind"`
	ResourceGroup         string   `json:"resource_group"`
	Location              string   `json:"location"`
	InstrumentationClues  []string `json:"instrumentation_clues"`
	SamplingClues         []string `json:"sampling_clues"`
	FilteringClues        []string `json:"filtering_clues"`
	LoggingLevelClues     []string `json:"logging_level_clues"`
	VisibleTelemetryTypes []string `json:"visible_telemetry_types"`
	Summary               string   `json:"summary"`
	RelatedIDs            []string `json:"related_ids"`
}

type AppInsightsOutput struct {
	Components []AppInsightsComponent `json:"components"`
	Targets    []AppInsightsAppTarget `json:"targets"`
	Findings   []Finding              `json:"findings"`
	Issues     []Issue                `json:"issues"`
	Metadata   RuntimeCommandMetadata `json:"metadata"`
}
