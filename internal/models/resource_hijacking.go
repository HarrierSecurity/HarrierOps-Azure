package models

type ResourceHijackingSurfaceDescriptor = FamilySurfaceDescriptor

type ResourceHijackingOverviewOutput struct {
	Metadata               ScopedCommandMetadata                `json:"metadata"`
	GroupedCommandName     string                               `json:"grouped_command_name"`
	CommandState           string                               `json:"command_state"`
	CurrentBehavior        string                               `json:"current_behavior"`
	PlannedInputModes      []string                             `json:"planned_input_modes"`
	PreferredArtifactOrder []string                             `json:"preferred_artifact_order"`
	SelectedSurface        *string                              `json:"selected_surface"`
	Surfaces               []ResourceHijackingSurfaceDescriptor `json:"surfaces"`
	Issues                 []Issue                              `json:"issues"`
}

type ResourceHijackingCapabilityStep = FamilyCapabilityStep

type ResourceHijackingRoleContext = FamilyRoleContext

type ResourceHijackingBoundaryNote = FamilyBoundaryNote

type ResourceHijackingAPIMState struct {
	State                   *string  `json:"state,omitempty"`
	PublicNetworkAccess     *string  `json:"public_network_access,omitempty"`
	VirtualNetworkType      *string  `json:"virtual_network_type,omitempty"`
	GatewayHostnames        []string `json:"gateway_hostnames"`
	BackendHostnames        []string `json:"backend_hostnames"`
	APICount                *int     `json:"api_count,omitempty"`
	SubscriptionCount       *int     `json:"subscription_count,omitempty"`
	ActiveSubscriptionCount *int     `json:"active_subscription_count,omitempty"`
	BackendCount            *int     `json:"backend_count,omitempty"`
	PolicyCount             *int     `json:"policy_count,omitempty"`
	PolicyControlTypes      []string `json:"policy_control_types"`
	NamedValueSecretCount   *int     `json:"named_value_secret_count,omitempty"`
	NamedValueKeyVaultCount *int     `json:"named_value_key_vault_count,omitempty"`
	WorkloadIdentityType    *string  `json:"workload_identity_type,omitempty"`
	Posture                 string   `json:"posture"`
}

type ResourceHijackingAPIMTarget struct {
	ID                     string                            `json:"id"`
	Name                   string                            `json:"api_management_service"`
	ResourceGroup          string                            `json:"resource_group"`
	Location               *string                           `json:"location,omitempty"`
	TakeoverRank           int                               `json:"takeover_rank"`
	TakeoverReason         string                            `json:"takeover_reason"`
	CapabilitySteps        []ResourceHijackingCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext *ResourceHijackingRoleContext     `json:"current_identity_context,omitempty"`
	CurrentState           ResourceHijackingAPIMState        `json:"current_state"`
	NotCollectedByDefault  []ResourceHijackingBoundaryNote   `json:"not_collected_by_default"`
	Summary                string                            `json:"summary"`
	RelatedIDs             []string                          `json:"related_ids"`
}

type ResourceHijackingLogicAppState = FamilyLogicAppState

type ResourceHijackingLogicAppTarget struct {
	ID                     string                            `json:"id"`
	Name                   string                            `json:"logic_app"`
	ResourceGroup          string                            `json:"resource_group"`
	Location               *string                           `json:"location,omitempty"`
	TakeoverRank           int                               `json:"takeover_rank"`
	TakeoverReason         string                            `json:"takeover_reason"`
	CapabilitySteps        []ResourceHijackingCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext *ResourceHijackingRoleContext     `json:"current_identity_context,omitempty"`
	CurrentState           ResourceHijackingLogicAppState    `json:"current_state"`
	NotCollectedByDefault  []ResourceHijackingBoundaryNote   `json:"not_collected_by_default"`
	Summary                string                            `json:"summary"`
	RelatedIDs             []string                          `json:"related_ids"`
}

type ResourceHijackingAutomationState struct {
	State                  *string  `json:"state,omitempty"`
	IdentityType           *string  `json:"identity_type,omitempty"`
	PublishedRunbookCount  *int     `json:"published_runbook_count,omitempty"`
	PublishedRunbookNames  []string `json:"published_runbook_names"`
	RunbookTypes           []string `json:"runbook_types"`
	RunbookCommandClues    []string `json:"runbook_command_clues"`
	RunbookResourceClues   []string `json:"runbook_resource_clues"`
	ScheduleCount          *int     `json:"schedule_count,omitempty"`
	JobScheduleCount       *int     `json:"job_schedule_count,omitempty"`
	WebhookCount           *int     `json:"webhook_count,omitempty"`
	HybridWorkerGroupCount *int     `json:"hybrid_worker_group_count,omitempty"`
	PrimaryStartMode       *string  `json:"primary_start_mode,omitempty"`
	PrimaryRunbookName     *string  `json:"primary_runbook_name,omitempty"`
	ScheduleRunbookNames   []string `json:"schedule_runbook_names"`
	WebhookRunbookNames    []string `json:"webhook_runbook_names"`
	ConsequenceTypes       []string `json:"consequence_types"`
	Posture                string   `json:"posture"`
}

type ResourceHijackingAutomationTarget struct {
	ID                     string                            `json:"id"`
	Name                   string                            `json:"automation_account"`
	ResourceGroup          string                            `json:"resource_group"`
	Location               *string                           `json:"location,omitempty"`
	TakeoverRank           int                               `json:"takeover_rank"`
	TakeoverReason         string                            `json:"takeover_reason"`
	CapabilitySteps        []ResourceHijackingCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext *ResourceHijackingRoleContext     `json:"current_identity_context,omitempty"`
	CurrentState           ResourceHijackingAutomationState  `json:"current_state"`
	NotCollectedByDefault  []ResourceHijackingBoundaryNote   `json:"not_collected_by_default"`
	Summary                string                            `json:"summary"`
	RelatedIDs             []string                          `json:"related_ids"`
}

type ResourceHijackingAPIMOutput struct {
	Metadata           ScopedCommandMetadata         `json:"metadata"`
	GroupedCommandName string                        `json:"grouped_command_name"`
	Surface            string                        `json:"surface"`
	InputMode          string                        `json:"input_mode"`
	CommandState       string                        `json:"command_state"`
	Summary            string                        `json:"summary"`
	BackingCommands    []string                      `json:"backing_commands"`
	Targets            []ResourceHijackingAPIMTarget `json:"targets"`
	Issues             []Issue                       `json:"issues"`
}

type ResourceHijackingLogicAppsOutput struct {
	Metadata           ScopedCommandMetadata             `json:"metadata"`
	GroupedCommandName string                            `json:"grouped_command_name"`
	Surface            string                            `json:"surface"`
	InputMode          string                            `json:"input_mode"`
	CommandState       string                            `json:"command_state"`
	Summary            string                            `json:"summary"`
	BackingCommands    []string                          `json:"backing_commands"`
	Targets            []ResourceHijackingLogicAppTarget `json:"targets"`
	Issues             []Issue                           `json:"issues"`
}

type ResourceHijackingAutomationOutput struct {
	Metadata           ScopedCommandMetadata               `json:"metadata"`
	GroupedCommandName string                              `json:"grouped_command_name"`
	Surface            string                              `json:"surface"`
	InputMode          string                              `json:"input_mode"`
	CommandState       string                              `json:"command_state"`
	Summary            string                              `json:"summary"`
	BackingCommands    []string                            `json:"backing_commands"`
	Targets            []ResourceHijackingAutomationTarget `json:"targets"`
	Issues             []Issue                             `json:"issues"`
}
