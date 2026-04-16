package models

type PersistenceSurfaceDescriptor struct {
	Surface          string   `json:"surface"`
	State            string   `json:"state"`
	Summary          string   `json:"summary"`
	OperatorQuestion string   `json:"operator_question"`
	BackingCommands  []string `json:"backing_commands"`
}

type PersistenceOverviewOutput struct {
	Metadata               ScopedCommandMetadata          `json:"metadata"`
	GroupedCommandName     string                         `json:"grouped_command_name"`
	CommandState           string                         `json:"command_state"`
	CurrentBehavior        string                         `json:"current_behavior"`
	PlannedInputModes      []string                       `json:"planned_input_modes"`
	PreferredArtifactOrder []string                       `json:"preferred_artifact_order"`
	SelectedSurface        *string                        `json:"selected_surface"`
	Surfaces               []PersistenceSurfaceDescriptor `json:"surfaces"`
	Issues                 []Issue                        `json:"issues"`
}

type PersistenceCapabilityStep struct {
	Action     string `json:"action"`
	APISurface string `json:"api_surface"`
	Status     string `json:"status"`
}

type PersistenceRoleContext struct {
	Name         string   `json:"name"`
	Kind         string   `json:"kind"`
	PrincipalID  *string  `json:"principal_id,omitempty"`
	IdentityType *string  `json:"identity_type,omitempty"`
	RoleNames    []string `json:"role_names"`
	ScopeIDs     []string `json:"scope_ids"`
	Summary      string   `json:"summary"`
}

type PersistenceAutomationState struct {
	RunbookCount                     *int                    `json:"runbook_count,omitempty"`
	PublishedRunbookCount            *int                    `json:"published_runbook_count,omitempty"`
	PublishedRunbookNames            []string                `json:"published_runbook_names"`
	ScheduleCount                    *int                    `json:"schedule_count,omitempty"`
	JobScheduleCount                 *int                    `json:"job_schedule_count,omitempty"`
	WebhookCount                     *int                    `json:"webhook_count,omitempty"`
	HybridWorkerGroupCount           *int                    `json:"hybrid_worker_group_count,omitempty"`
	CredentialCount                  *int                    `json:"credential_count,omitempty"`
	CertificateCount                 *int                    `json:"certificate_count,omitempty"`
	ConnectionCount                  *int                    `json:"connection_count,omitempty"`
	VariableCount                    *int                    `json:"variable_count,omitempty"`
	EncryptedVariableCount           *int                    `json:"encrypted_variable_count,omitempty"`
	PrimaryStartMode                 *string                 `json:"primary_start_mode,omitempty"`
	PrimaryRunbookName               *string                 `json:"primary_runbook_name,omitempty"`
	IdentityType                     *string                 `json:"identity_type,omitempty"`
	StrongestVisibleExecutionContext *PersistenceRoleContext `json:"strongest_visible_execution_context,omitempty"`
	NearbyThematicNames              []string                `json:"nearby_thematic_names"`
	MissingTargetMapping             bool                    `json:"missing_target_mapping"`
}

type PersistenceAutomationAccount struct {
	ID                      string                      `json:"id"`
	Name                    string                      `json:"automation_account"`
	ResourceGroup           string                      `json:"resource_group"`
	Location                *string                     `json:"location,omitempty"`
	CapabilitySteps         []PersistenceCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext  *PersistenceRoleContext     `json:"current_identity_context,omitempty"`
	ExecutionContextOptions []string                    `json:"execution_context_options"`
	CurrentState            PersistenceAutomationState  `json:"current_state"`
	StillUnmapped           []string                    `json:"still_unmapped"`
	Summary                 string                      `json:"summary"`
	RelatedIDs              []string                    `json:"related_ids"`
}

type PersistenceAutomationOutput struct {
	Metadata           ScopedCommandMetadata          `json:"metadata"`
	GroupedCommandName string                         `json:"grouped_command_name"`
	Surface            string                         `json:"surface"`
	InputMode          string                         `json:"input_mode"`
	CommandState       string                         `json:"command_state"`
	Summary            string                         `json:"summary"`
	BackingCommands    []string                       `json:"backing_commands"`
	AutomationAccounts []PersistenceAutomationAccount `json:"automation_accounts"`
	Issues             []Issue                        `json:"issues"`
}
