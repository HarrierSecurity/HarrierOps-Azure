package models

type ChainSourceDescriptor struct {
	Command       string   `json:"command"`
	MinimumFields []string `json:"minimum_fields"`
	Rationale     string   `json:"rationale"`
}

type ChainFamilyDescriptor struct {
	Family              string                  `json:"family"`
	State               string                  `json:"state"`
	Meaning             string                  `json:"meaning"`
	Summary             string                  `json:"summary"`
	AllowedClaim        string                  `json:"allowed_claim"`
	CurrentGap          string                  `json:"current_gap"`
	BestCurrentExamples []string                `json:"best_current_examples"`
	SourceCommands      []ChainSourceDescriptor `json:"source_commands"`
}

type ChainsOverviewOutput struct {
	Metadata               ScopedCommandMetadata   `json:"metadata"`
	GroupedCommandName     string                  `json:"grouped_command_name"`
	CommandState           string                  `json:"command_state"`
	CurrentBehavior        string                  `json:"current_behavior"`
	PlannedInputModes      []string                `json:"planned_input_modes"`
	PreferredArtifactOrder []string                `json:"preferred_artifact_order"`
	SelectedFamily         *string                 `json:"selected_family"`
	Families               []ChainFamilyDescriptor `json:"families"`
	Issues                 []Issue                 `json:"issues"`
}

type ChainSourceArtifact struct {
	Command      string `json:"command"`
	ArtifactType string `json:"artifact_type"`
	Path         string `json:"path"`
}

type ChainPathRecord struct {
	ChainID                        string   `json:"chain_id"`
	AssetID                        string   `json:"asset_id"`
	AssetName                      string   `json:"asset_name"`
	AssetKind                      string   `json:"asset_kind"`
	Location                       *string  `json:"location,omitempty"`
	Surface                        *string  `json:"surface,omitempty"`
	PersistenceType                *string  `json:"persistence_type,omitempty"`
	Classification                 *string  `json:"classification,omitempty"`
	Durability                     *string  `json:"durability,omitempty"`
	WhatPersists                   *string  `json:"what_persists,omitempty"`
	FootholdAnchor                 *string  `json:"foothold_anchor,omitempty"`
	SurvivesHostRebuild            *bool    `json:"survives_host_rebuild,omitempty"`
	SurvivesOriginalAccountCleanup *bool    `json:"survives_original_account_cleanup,omitempty"`
	CurrentEvidence                *string  `json:"current_evidence,omitempty"`
	MissingProof                   *string  `json:"missing_proof,omitempty"`
	OperatorActionability          *string  `json:"operator_actionability,omitempty"`
	RecommendedFixFocus            *string  `json:"recommended_fix_focus,omitempty"`
	StartingFoothold               *string  `json:"starting_foothold,omitempty"`
	SourceCommand                  *string  `json:"source_command,omitempty"`
	SourceContext                  *string  `json:"source_context,omitempty"`
	Source                         *string  `json:"source,omitempty"`
	SettingName                    *string  `json:"setting_name,omitempty"`
	ClueType                       string   `json:"clue_type"`
	ConfirmationBasis              *string  `json:"confirmation_basis,omitempty"`
	Priority                       string   `json:"priority"`
	Urgency                        *string  `json:"urgency,omitempty"`
	Actionability                  *string  `json:"actionability,omitempty"`
	ActionabilityState             *string  `json:"actionability_state,omitempty"`
	VisiblePath                    string   `json:"visible_path"`
	InsertionPoint                 *string  `json:"insertion_point,omitempty"`
	InsertionPointLabel            *string  `json:"insertion_point_display,omitempty"`
	PathConcept                    *string  `json:"path_concept,omitempty"`
	PathType                       *string  `json:"path_type,omitempty"`
	PrimarySurface                 *string  `json:"primary_injection_surface,omitempty"`
	PrimaryInputRef                *string  `json:"primary_trusted_input_ref,omitempty"`
	StrongerOutcome                *string  `json:"stronger_outcome,omitempty"`
	WhyCare                        *string  `json:"why_care,omitempty"`
	LikelyImpact                   *string  `json:"likely_impact,omitempty"`
	LikelyAzureImpact              *string  `json:"likely_azure_impact,omitempty"`
	ConfidenceBoundary             *string  `json:"confidence_boundary,omitempty"`
	WhatsMissing                   *string  `json:"whats_missing,omitempty"`
	Note                           *string  `json:"note,omitempty"`
	TargetService                  string   `json:"target_service"`
	TargetResolution               string   `json:"target_resolution"`
	EvidenceCommands               []string `json:"evidence_commands"`
	JoinedSurfaceTypes             []string `json:"joined_surface_types"`
	TargetCount                    int      `json:"target_count"`
	TargetIDs                      []string `json:"target_ids"`
	TargetNames                    []string `json:"target_names"`
	TargetVisibility               *string  `json:"target_visibility_issue"`
	NextReview                     string   `json:"next_review"`
	Summary                        string   `json:"summary"`
	MissingConfirmation            string   `json:"missing_confirmation"`
	RelatedIDs                     []string `json:"related_ids"`
}

type ChainsOutput struct {
	Metadata                ScopedCommandMetadata `json:"metadata"`
	GroupedCommandName      string                `json:"grouped_command_name"`
	Family                  string                `json:"family"`
	InputMode               string                `json:"input_mode"`
	CommandState            string                `json:"command_state"`
	Summary                 string                `json:"summary"`
	ClaimBoundary           string                `json:"claim_boundary"`
	CurrentGap              *string               `json:"current_gap,omitempty"`
	ArtifactPreferenceOrder []string              `json:"artifact_preference_order"`
	BackingCommands         []string              `json:"backing_commands"`
	SourceArtifacts         []ChainSourceArtifact `json:"source_artifacts"`
	Paths                   []ChainPathRecord     `json:"paths"`
	Issues                  []Issue               `json:"issues"`
}
