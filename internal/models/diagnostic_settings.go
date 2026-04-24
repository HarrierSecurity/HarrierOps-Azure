package models

type DiagnosticSettingsDestination struct {
	Type       string  `json:"type"`
	ResourceID *string `json:"resource_id,omitempty"`
	Detail     *string `json:"detail,omitempty"`
}

type DiagnosticSettingsCategory struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

type DiagnosticSettingAsset struct {
	ID                   string                          `json:"id"`
	Name                 string                          `json:"name"`
	SourceResourceID     string                          `json:"source_resource_id"`
	Destinations         []DiagnosticSettingsDestination `json:"destinations"`
	Logs                 []DiagnosticSettingsCategory    `json:"logs"`
	Metrics              []DiagnosticSettingsCategory    `json:"metrics"`
	EnabledCategories    []string                        `json:"enabled_categories"`
	DisabledCategories   []string                        `json:"disabled_categories"`
	CategoryGroups       []string                        `json:"category_groups"`
	HighSignalCategories []string                        `json:"high_signal_categories"`
	DestinationTypes     []string                        `json:"destination_types"`
	RelatedIDs           []string                        `json:"related_ids"`
	Summary              string                          `json:"summary"`
}

type DiagnosticSettingsSource struct {
	ID                         string                   `json:"id"`
	Name                       string                   `json:"name"`
	Type                       string                   `json:"type"`
	ResourceGroup              string                   `json:"resource_group"`
	Location                   string                   `json:"location"`
	DiagnosticSettings         []DiagnosticSettingAsset `json:"diagnostic_settings"`
	DiagnosticSettingCount     int                      `json:"diagnostic_setting_count"`
	EnabledCategories          []string                 `json:"enabled_categories"`
	DisabledCategories         []string                 `json:"disabled_categories"`
	SupportedCategories        []string                 `json:"supported_categories"`
	NotExportedSupported       []string                 `json:"not_exported_supported_categories"`
	SupportedCategoryCatalog   bool                     `json:"supported_category_catalog"`
	CategoryGroups             []string                 `json:"category_groups"`
	HighSignalCategories       []string                 `json:"high_signal_categories"`
	DestinationTypes           []string                 `json:"destination_types"`
	HasDiagnosticSettings      bool                     `json:"has_diagnostic_settings"`
	HasPartialLogPosture       bool                     `json:"has_partial_log_posture"`
	HasHighSignalLogPosture    bool                     `json:"has_high_signal_log_posture"`
	HasNonWorkspaceDestination bool                     `json:"has_non_workspace_destination"`
	RelatedIDs                 []string                 `json:"related_ids"`
	Summary                    string                   `json:"summary"`
}

type DiagnosticSettingsMetadata = RuntimeCommandMetadata

type DiagnosticSettingsOutput struct {
	Sources  []DiagnosticSettingsSource `json:"sources"`
	Findings []Finding                  `json:"findings"`
	Issues   []Issue                    `json:"issues"`
	Metadata DiagnosticSettingsMetadata `json:"metadata"`
}
