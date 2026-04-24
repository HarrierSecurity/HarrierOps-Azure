package models

type EvasionSurfaceDescriptor = FamilySurfaceDescriptor

type EvasionOverviewOutput struct {
	Metadata               ScopedCommandMetadata      `json:"metadata"`
	GroupedCommandName     string                     `json:"grouped_command_name"`
	CommandState           string                     `json:"command_state"`
	CurrentBehavior        string                     `json:"current_behavior"`
	PlannedInputModes      []string                   `json:"planned_input_modes"`
	PreferredArtifactOrder []string                   `json:"preferred_artifact_order"`
	SelectedSurface        *string                    `json:"selected_surface"`
	Surfaces               []EvasionSurfaceDescriptor `json:"surfaces"`
	Issues                 []Issue                    `json:"issues"`
}

type EvasionCapabilityStep = FamilyCapabilityStep

type EvasionRoleContext = FamilyRoleContext

type EvasionBoundaryNote = FamilyBoundaryNote

type EvasionDCRState struct {
	DataSourceTypes       []string `json:"data_source_types"`
	Streams               []string `json:"streams"`
	HighSignalStreams     []string `json:"high_signal_streams"`
	DestinationTypes      []string `json:"destination_types"`
	AssociationTargets    []string `json:"association_targets"`
	TransformationCount   int      `json:"transformation_count"`
	AssociationCount      int      `json:"association_count"`
	TransformationPosture string   `json:"transformation_posture"`
	DestinationPosture    string   `json:"destination_posture"`
}

type EvasionDiagnosticSettingsState struct {
	SourceType             string   `json:"source_type"`
	DiagnosticSettingCount int      `json:"diagnostic_setting_count"`
	EnabledCategories      []string `json:"enabled_categories"`
	NotExportedCategories  []string `json:"not_exported_categories"`
	SupportedCategories    []string `json:"supported_categories"`
	SupportedCategoryProof bool     `json:"supported_category_proof"`
	CategoryGroups         []string `json:"category_groups"`
	HighSignalCategories   []string `json:"high_signal_categories"`
	DestinationTypes       []string `json:"destination_types"`
	HasNonWorkspaceSink    bool     `json:"has_non_workspace_sink"`
	ExportPosture          string   `json:"export_posture"`
	DestinationPosture     string   `json:"destination_posture"`
}

type EvasionAppInsightsState struct {
	Kind                  string   `json:"kind"`
	InstrumentationClues  []string `json:"instrumentation_clues"`
	SamplingClues         []string `json:"sampling_clues"`
	FilteringClues        []string `json:"filtering_clues"`
	LoggingLevelClues     []string `json:"logging_level_clues"`
	VisibleTelemetryTypes []string `json:"visible_telemetry_types"`
	Posture               string   `json:"posture"`
}

type EvasionDCR struct {
	ID                     string                  `json:"id"`
	Name                   string                  `json:"dcr"`
	ResourceGroup          string                  `json:"resource_group"`
	Location               string                  `json:"location"`
	DisruptionRank         int                     `json:"disruption_rank"`
	DisruptionReason       string                  `json:"disruption_reason"`
	CapabilitySteps        []EvasionCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext *EvasionRoleContext     `json:"current_identity_context,omitempty"`
	CurrentState           EvasionDCRState         `json:"current_state"`
	NotCollectedByDefault  []EvasionBoundaryNote   `json:"not_collected_by_default"`
	Summary                string                  `json:"summary"`
	RelatedIDs             []string                `json:"related_ids"`
}

type EvasionDiagnosticSettingsSource struct {
	ID                     string                         `json:"id"`
	Name                   string                         `json:"source"`
	ResourceGroup          string                         `json:"resource_group"`
	Location               string                         `json:"location"`
	DisruptionRank         int                            `json:"disruption_rank"`
	DisruptionReason       string                         `json:"disruption_reason"`
	CapabilitySteps        []EvasionCapabilityStep        `json:"capability_steps"`
	CurrentIdentityContext *EvasionRoleContext            `json:"current_identity_context,omitempty"`
	CurrentState           EvasionDiagnosticSettingsState `json:"current_state"`
	NotCollectedByDefault  []EvasionBoundaryNote          `json:"not_collected_by_default"`
	Summary                string                         `json:"summary"`
	RelatedIDs             []string                       `json:"related_ids"`
}

type EvasionAppInsightsTarget struct {
	ID                     string                  `json:"id"`
	Name                   string                  `json:"target"`
	ResourceGroup          string                  `json:"resource_group"`
	Location               string                  `json:"location"`
	DisruptionRank         int                     `json:"disruption_rank"`
	DisruptionReason       string                  `json:"disruption_reason"`
	CapabilitySteps        []EvasionCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext *EvasionRoleContext     `json:"current_identity_context,omitempty"`
	CurrentState           EvasionAppInsightsState `json:"current_state"`
	NotCollectedByDefault  []EvasionBoundaryNote   `json:"not_collected_by_default"`
	Summary                string                  `json:"summary"`
	RelatedIDs             []string                `json:"related_ids"`
}

type EvasionDCROutput struct {
	Metadata           ScopedCommandMetadata `json:"metadata"`
	GroupedCommandName string                `json:"grouped_command_name"`
	Surface            string                `json:"surface"`
	InputMode          string                `json:"input_mode"`
	CommandState       string                `json:"command_state"`
	Summary            string                `json:"summary"`
	BackingCommands    []string              `json:"backing_commands"`
	MonitoringSinks    []MonitoringSinkAsset `json:"monitoring_sinks"`
	DCRs               []EvasionDCR          `json:"dcrs"`
	Issues             []Issue               `json:"issues"`
}

type EvasionDiagnosticSettingsOutput struct {
	Metadata           ScopedCommandMetadata             `json:"metadata"`
	GroupedCommandName string                            `json:"grouped_command_name"`
	Surface            string                            `json:"surface"`
	InputMode          string                            `json:"input_mode"`
	CommandState       string                            `json:"command_state"`
	Summary            string                            `json:"summary"`
	BackingCommands    []string                          `json:"backing_commands"`
	MonitoringSinks    []MonitoringSinkAsset             `json:"monitoring_sinks"`
	Sources            []EvasionDiagnosticSettingsSource `json:"sources"`
	Issues             []Issue                           `json:"issues"`
}

type EvasionAppInsightsOutput struct {
	Metadata           ScopedCommandMetadata      `json:"metadata"`
	GroupedCommandName string                     `json:"grouped_command_name"`
	Surface            string                     `json:"surface"`
	InputMode          string                     `json:"input_mode"`
	CommandState       string                     `json:"command_state"`
	Summary            string                     `json:"summary"`
	BackingCommands    []string                   `json:"backing_commands"`
	Targets            []EvasionAppInsightsTarget `json:"targets"`
	Components         []AppInsightsComponent     `json:"components"`
	Issues             []Issue                    `json:"issues"`
}
