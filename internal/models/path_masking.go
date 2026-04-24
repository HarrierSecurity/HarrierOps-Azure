package models

type PathMaskingSurfaceDescriptor = FamilySurfaceDescriptor

type PathMaskingOverviewOutput struct {
	Metadata               ScopedCommandMetadata          `json:"metadata"`
	GroupedCommandName     string                         `json:"grouped_command_name"`
	CommandState           string                         `json:"command_state"`
	CurrentBehavior        string                         `json:"current_behavior"`
	PlannedInputModes      []string                       `json:"planned_input_modes"`
	PreferredArtifactOrder []string                       `json:"preferred_artifact_order"`
	SelectedSurface        *string                        `json:"selected_surface"`
	Surfaces               []PathMaskingSurfaceDescriptor `json:"surfaces"`
	Issues                 []Issue                        `json:"issues"`
}

type PathMaskingCapabilityStep = FamilyCapabilityStep

type PathMaskingRoleContext = FamilyRoleContext

type PathMaskingBoundaryNote = FamilyBoundaryNote

type PathMaskingAPIMState struct {
	GatewayHostnames        []string `json:"gateway_hostnames"`
	BackendHostnames        []string `json:"backend_hostnames"`
	APICount                *int     `json:"api_count,omitempty"`
	SubscriptionCount       *int     `json:"subscription_count,omitempty"`
	PolicyCount             *int     `json:"policy_count,omitempty"`
	PolicyControlTypes      []string `json:"policy_control_types"`
	NamedValueSecretCount   *int     `json:"named_value_secret_count,omitempty"`
	NamedValueKeyVaultCount *int     `json:"named_value_key_vault_count,omitempty"`
	PublicNetworkAccess     *string  `json:"public_network_access,omitempty"`
	VirtualNetworkType      *string  `json:"virtual_network_type,omitempty"`
	Posture                 string   `json:"posture"`
}

type PathMaskingRelayState struct {
	ServiceBusEndpoint     *string  `json:"service_bus_endpoint,omitempty"`
	HybridConnectionCount  *int     `json:"hybrid_connection_count,omitempty"`
	AuthorizationRuleCount *int     `json:"authorization_rule_count,omitempty"`
	HybridConnectionNames  []string `json:"hybrid_connection_names"`
	ListenerSummary        string   `json:"listener_summary"`
	AppServiceAttachments  []string `json:"app_service_attachments"`
	Posture                string   `json:"posture"`
}

type PathMaskingLogicAppState = FamilyLogicAppState

type PathMaskingAPIMTarget struct {
	ID                     string                      `json:"id"`
	Name                   string                      `json:"api_management_service"`
	ResourceGroup          string                      `json:"resource_group"`
	Location               *string                     `json:"location,omitempty"`
	MaskingRank            int                         `json:"masking_rank"`
	MaskingReason          string                      `json:"masking_reason"`
	CapabilitySteps        []PathMaskingCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext *PathMaskingRoleContext     `json:"current_identity_context,omitempty"`
	CurrentState           PathMaskingAPIMState        `json:"current_state"`
	NotCollectedByDefault  []PathMaskingBoundaryNote   `json:"not_collected_by_default"`
	Summary                string                      `json:"summary"`
	RelatedIDs             []string                    `json:"related_ids"`
}

type PathMaskingLogicAppTarget struct {
	ID                     string                      `json:"id"`
	Name                   string                      `json:"logic_app"`
	ResourceGroup          string                      `json:"resource_group"`
	Location               *string                     `json:"location,omitempty"`
	MaskingRank            int                         `json:"masking_rank"`
	MaskingReason          string                      `json:"masking_reason"`
	CapabilitySteps        []PathMaskingCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext *PathMaskingRoleContext     `json:"current_identity_context,omitempty"`
	CurrentState           PathMaskingLogicAppState    `json:"current_state"`
	NotCollectedByDefault  []PathMaskingBoundaryNote   `json:"not_collected_by_default"`
	Summary                string                      `json:"summary"`
	RelatedIDs             []string                    `json:"related_ids"`
}

type PathMaskingRelayTarget struct {
	ID                     string                      `json:"id"`
	Name                   string                      `json:"relay_namespace"`
	ResourceGroup          string                      `json:"resource_group"`
	Location               *string                     `json:"location,omitempty"`
	MaskingRank            int                         `json:"masking_rank"`
	MaskingReason          string                      `json:"masking_reason"`
	CapabilitySteps        []PathMaskingCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext *PathMaskingRoleContext     `json:"current_identity_context,omitempty"`
	CurrentState           PathMaskingRelayState       `json:"current_state"`
	NotCollectedByDefault  []PathMaskingBoundaryNote   `json:"not_collected_by_default"`
	Summary                string                      `json:"summary"`
	RelatedIDs             []string                    `json:"related_ids"`
}

type PathMaskingAPIMOutput struct {
	Metadata           ScopedCommandMetadata   `json:"metadata"`
	GroupedCommandName string                  `json:"grouped_command_name"`
	Surface            string                  `json:"surface"`
	InputMode          string                  `json:"input_mode"`
	CommandState       string                  `json:"command_state"`
	Summary            string                  `json:"summary"`
	BackingCommands    []string                `json:"backing_commands"`
	Targets            []PathMaskingAPIMTarget `json:"targets"`
	Issues             []Issue                 `json:"issues"`
}

type PathMaskingLogicAppsOutput struct {
	Metadata           ScopedCommandMetadata       `json:"metadata"`
	GroupedCommandName string                      `json:"grouped_command_name"`
	Surface            string                      `json:"surface"`
	InputMode          string                      `json:"input_mode"`
	CommandState       string                      `json:"command_state"`
	Summary            string                      `json:"summary"`
	BackingCommands    []string                    `json:"backing_commands"`
	Targets            []PathMaskingLogicAppTarget `json:"targets"`
	Issues             []Issue                     `json:"issues"`
}

type PathMaskingRelayOutput struct {
	Metadata           ScopedCommandMetadata    `json:"metadata"`
	GroupedCommandName string                   `json:"grouped_command_name"`
	Surface            string                   `json:"surface"`
	InputMode          string                   `json:"input_mode"`
	CommandState       string                   `json:"command_state"`
	Summary            string                   `json:"summary"`
	BackingCommands    []string                 `json:"backing_commands"`
	Targets            []PathMaskingRelayTarget `json:"targets"`
	Issues             []Issue                  `json:"issues"`
}
