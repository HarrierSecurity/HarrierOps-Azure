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

type PersistenceAutomationAccountState struct {
	RunbookCount                     *int                    `json:"runbook_count,omitempty"`
	PublishedRunbookCount            *int                    `json:"published_runbook_count,omitempty"`
	PublishedRunbookNames            []string                `json:"published_runbook_names"`
	ScheduleCount                    *int                    `json:"schedule_count,omitempty"`
	ScheduleDefinitions              []string                `json:"schedule_definitions"`
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
	ID                      string                            `json:"id"`
	Name                    string                            `json:"automation_account"`
	ResourceGroup           string                            `json:"resource_group"`
	Location                *string                           `json:"location,omitempty"`
	CapabilitySteps         []PersistenceCapabilityStep       `json:"capability_steps"`
	CurrentIdentityContext  *PersistenceRoleContext           `json:"current_identity_context,omitempty"`
	ExecutionContextOptions []string                          `json:"execution_context_options"`
	CurrentState            PersistenceAutomationAccountState `json:"current_state"`
	StillUnmapped           []string                          `json:"still_unmapped"`
	Summary                 string                            `json:"summary"`
	RelatedIDs              []string                          `json:"related_ids"`
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

type PersistenceFunctionAppState struct {
	State                            *string                 `json:"state,omitempty"`
	Hostname                         *string                 `json:"hostname,omitempty"`
	PublicNetworkAccess              *string                 `json:"public_network_access,omitempty"`
	Runtime                          *string                 `json:"runtime,omitempty"`
	Deployment                       *string                 `json:"deployment,omitempty"`
	IdentityType                     *string                 `json:"identity_type,omitempty"`
	AlwaysOn                         *bool                   `json:"always_on,omitempty"`
	AzureWebJobsStorageValueType     *string                 `json:"azure_webjobs_storage_value_type,omitempty"`
	KeyVaultReferenceCount           *int                    `json:"key_vault_reference_count,omitempty"`
	RunFromPackage                   *bool                   `json:"run_from_package,omitempty"`
	TriggerTypes                     []string                `json:"trigger_types,omitempty"`
	VisibleFunctions                 []FunctionChildAsset    `json:"visible_functions,omitempty"`
	StrongestVisibleExecutionContext *PersistenceRoleContext `json:"strongest_visible_execution_context,omitempty"`
	NearbyThematicNames              []string                `json:"nearby_thematic_names,omitempty"`
}

type PersistenceFunctionApp struct {
	ID                      string                      `json:"id"`
	Name                    string                      `json:"function_app"`
	ResourceGroup           string                      `json:"resource_group"`
	Location                string                      `json:"location"`
	CapabilitySteps         []PersistenceCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext  *PersistenceRoleContext     `json:"current_identity_context,omitempty"`
	ExecutionContextOptions []string                    `json:"execution_context_options"`
	CurrentState            PersistenceFunctionAppState `json:"current_state"`
	StillUnmapped           []string                    `json:"still_unmapped"`
	Summary                 string                      `json:"summary"`
	RelatedIDs              []string                    `json:"related_ids"`
}

type PersistenceFunctionsOutput struct {
	Metadata           ScopedCommandMetadata    `json:"metadata"`
	GroupedCommandName string                   `json:"grouped_command_name"`
	Surface            string                   `json:"surface"`
	InputMode          string                   `json:"input_mode"`
	CommandState       string                   `json:"command_state"`
	Summary            string                   `json:"summary"`
	BackingCommands    []string                 `json:"backing_commands"`
	FunctionApps       []PersistenceFunctionApp `json:"function_apps"`
	Issues             []Issue                  `json:"issues"`
}

type PersistenceAppServiceState struct {
	State                            *string                 `json:"state,omitempty"`
	Hostname                         *string                 `json:"hostname,omitempty"`
	PublicNetworkAccess              *string                 `json:"public_network_access,omitempty"`
	Runtime                          *string                 `json:"runtime,omitempty"`
	Deployment                       *string                 `json:"deployment,omitempty"`
	DeploymentRepoURL                *string                 `json:"deployment_repo_url,omitempty"`
	DeploymentBranch                 *string                 `json:"deployment_branch,omitempty"`
	DeploymentIsGitHubAction         *bool                   `json:"deployment_is_github_action,omitempty"`
	DeploymentManualIntegration      *bool                   `json:"deployment_manual_integration,omitempty"`
	IdentityType                     *string                 `json:"identity_type,omitempty"`
	AppSettingsCount                 *int                    `json:"app_settings_count,omitempty"`
	KeyVaultReferenceCount           *int                    `json:"key_vault_reference_count,omitempty"`
	SensitiveSettingCount            *int                    `json:"sensitive_setting_count,omitempty"`
	ConnectionStringCount            *int                    `json:"connection_string_count,omitempty"`
	KeyVaultConnectionStringCount    *int                    `json:"key_vault_connection_string_count,omitempty"`
	ConnectionStringTypes            []string                `json:"connection_string_types,omitempty"`
	RunFromPackage                   *bool                   `json:"run_from_package,omitempty"`
	HTTPSOnly                        *bool                   `json:"https_only,omitempty"`
	MinTLSVersion                    *string                 `json:"min_tls_version,omitempty"`
	FTPSState                        *string                 `json:"ftps_state,omitempty"`
	VisibleSensitiveSettingNames     []string                `json:"visible_sensitive_setting_names,omitempty"`
	StrongestVisibleExecutionContext *PersistenceRoleContext `json:"strongest_visible_execution_context,omitempty"`
	NearbyThematicNames              []string                `json:"nearby_thematic_names,omitempty"`
}

type PersistenceAppService struct {
	ID                      string                      `json:"id"`
	Name                    string                      `json:"app_service"`
	ResourceGroup           string                      `json:"resource_group"`
	Location                string                      `json:"location"`
	CapabilitySteps         []PersistenceCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext  *PersistenceRoleContext     `json:"current_identity_context,omitempty"`
	ExecutionContextOptions []string                    `json:"execution_context_options"`
	CurrentState            PersistenceAppServiceState  `json:"current_state"`
	StillUnmapped           []string                    `json:"still_unmapped"`
	Summary                 string                      `json:"summary"`
	RelatedIDs              []string                    `json:"related_ids"`
}

type PersistenceAppServiceOutput struct {
	Metadata           ScopedCommandMetadata   `json:"metadata"`
	GroupedCommandName string                  `json:"grouped_command_name"`
	Surface            string                  `json:"surface"`
	InputMode          string                  `json:"input_mode"`
	CommandState       string                  `json:"command_state"`
	Summary            string                  `json:"summary"`
	BackingCommands    []string                `json:"backing_commands"`
	AppServices        []PersistenceAppService `json:"app_services"`
	Issues             []Issue                 `json:"issues"`
}

type PersistenceWebJobState struct {
	Mode                             string                  `json:"mode"`
	JobType                          *string                 `json:"job_type,omitempty"`
	Status                           *string                 `json:"status,omitempty"`
	DetailedStatus                   *string                 `json:"detailed_status,omitempty"`
	LatestRunStatus                  *string                 `json:"latest_run_status,omitempty"`
	LatestRunTrigger                 *string                 `json:"latest_run_trigger,omitempty"`
	RunCommand                       *string                 `json:"run_command,omitempty"`
	ScheduleExpression               *string                 `json:"schedule_expression,omitempty"`
	SchedulerLogsURL                 *string                 `json:"scheduler_logs_url,omitempty"`
	ParentAppName                    string                  `json:"parent_app_name"`
	ParentHostname                   *string                 `json:"parent_hostname,omitempty"`
	ParentRuntime                    *string                 `json:"parent_runtime,omitempty"`
	ParentPublicNetworkAccess        *string                 `json:"parent_public_network_access,omitempty"`
	ParentIdentityType               *string                 `json:"parent_identity_type,omitempty"`
	ParentAppSettingsCount           *int                    `json:"parent_app_settings_count,omitempty"`
	ParentKeyVaultReferenceCount     *int                    `json:"parent_key_vault_reference_count,omitempty"`
	ParentConnectionStringCount      *int                    `json:"parent_connection_string_count,omitempty"`
	StrongestVisibleExecutionContext *PersistenceRoleContext `json:"strongest_visible_execution_context,omitempty"`
	NearbyThematicNames              []string                `json:"nearby_thematic_names,omitempty"`
}

type PersistenceWebJob struct {
	ID                      string                      `json:"id"`
	Name                    string                      `json:"webjob"`
	ResourceGroup           string                      `json:"resource_group"`
	Location                string                      `json:"location"`
	CapabilitySteps         []PersistenceCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext  *PersistenceRoleContext     `json:"current_identity_context,omitempty"`
	ExecutionContextOptions []string                    `json:"execution_context_options"`
	CurrentState            PersistenceWebJobState      `json:"current_state"`
	StillUnmapped           []string                    `json:"still_unmapped"`
	Summary                 string                      `json:"summary"`
	RelatedIDs              []string                    `json:"related_ids"`
}

type PersistenceWebJobsOutput struct {
	Metadata           ScopedCommandMetadata `json:"metadata"`
	GroupedCommandName string                `json:"grouped_command_name"`
	Surface            string                `json:"surface"`
	InputMode          string                `json:"input_mode"`
	CommandState       string                `json:"command_state"`
	Summary            string                `json:"summary"`
	BackingCommands    []string              `json:"backing_commands"`
	WebJobs            []PersistenceWebJob   `json:"webjobs"`
	Issues             []Issue               `json:"issues"`
}

type PersistenceContainerAppsJobState struct {
	EnvironmentID                    *string                     `json:"environment_id,omitempty"`
	TriggerType                      *string                     `json:"trigger_type,omitempty"`
	ScheduleExpression               *string                     `json:"schedule_expression,omitempty"`
	EventRules                       []ContainerAppsJobEventRule `json:"event_rules,omitempty"`
	ContainerImages                  []string                    `json:"container_images,omitempty"`
	Command                          []string                    `json:"command,omitempty"`
	Parallelism                      *int                        `json:"parallelism,omitempty"`
	ReplicaCompletionCount           *int                        `json:"replica_completion_count,omitempty"`
	ReplicaRetryLimit                *int                        `json:"replica_retry_limit,omitempty"`
	ReplicaTimeout                   *int                        `json:"replica_timeout,omitempty"`
	IdentityType                     *string                     `json:"identity_type,omitempty"`
	WorkloadPrincipalID              *string                     `json:"workload_principal_id,omitempty"`
	WorkloadClientID                 *string                     `json:"workload_client_id,omitempty"`
	WorkloadIdentityIDs              []string                    `json:"workload_identity_ids,omitempty"`
	SecretCount                      *int                        `json:"secret_count,omitempty"`
	KeyVaultSecretCount              *int                        `json:"key_vault_secret_count,omitempty"`
	RegistryServers                  []string                    `json:"registry_servers,omitempty"`
	RegistryIdentityCount            *int                        `json:"registry_identity_count,omitempty"`
	RegistryPasswordRefCount         *int                        `json:"registry_password_ref_count,omitempty"`
	StrongestVisibleExecutionContext *PersistenceRoleContext     `json:"strongest_visible_execution_context,omitempty"`
	NearbyThematicNames              []string                    `json:"nearby_thematic_names,omitempty"`
}

type PersistenceContainerAppsJob struct {
	ID                      string                           `json:"id"`
	Name                    string                           `json:"container_apps_job"`
	ResourceGroup           string                           `json:"resource_group"`
	Location                string                           `json:"location"`
	CapabilitySteps         []PersistenceCapabilityStep      `json:"capability_steps"`
	CurrentIdentityContext  *PersistenceRoleContext          `json:"current_identity_context,omitempty"`
	ExecutionContextOptions []string                         `json:"execution_context_options"`
	CurrentState            PersistenceContainerAppsJobState `json:"current_state"`
	StillUnmapped           []string                         `json:"still_unmapped"`
	Summary                 string                           `json:"summary"`
	RelatedIDs              []string                         `json:"related_ids"`
}

type PersistenceContainerAppsJobsOutput struct {
	Metadata           ScopedCommandMetadata         `json:"metadata"`
	GroupedCommandName string                        `json:"grouped_command_name"`
	Surface            string                        `json:"surface"`
	InputMode          string                        `json:"input_mode"`
	CommandState       string                        `json:"command_state"`
	Summary            string                        `json:"summary"`
	BackingCommands    []string                      `json:"backing_commands"`
	ContainerAppsJobs  []PersistenceContainerAppsJob `json:"container_apps_jobs"`
	Issues             []Issue                       `json:"issues"`
}

type PersistenceVMExtensionState struct {
	TargetKind                       string                  `json:"target_kind"`
	TargetName                       string                  `json:"target_name"`
	TargetID                         string                  `json:"target_id"`
	Publisher                        *string                 `json:"publisher,omitempty"`
	ExtensionType                    *string                 `json:"extension_type,omitempty"`
	TypeHandlerVersion               *string                 `json:"type_handler_version,omitempty"`
	AutoUpgradeMinorVersion          *bool                   `json:"auto_upgrade_minor_version,omitempty"`
	EnableAutomaticUpgrade           *bool                   `json:"enable_automatic_upgrade,omitempty"`
	FileURIHosts                     []string                `json:"file_uri_hosts,omitempty"`
	FileURICount                     *int                    `json:"file_uri_count,omitempty"`
	CommandClue                      *string                 `json:"command_clue,omitempty"`
	PublicSettingKeys                []string                `json:"public_setting_keys,omitempty"`
	ProtectedSettingsPresent         *bool                   `json:"protected_settings_present,omitempty"`
	KeyVaultProtectedSettings        *bool                   `json:"key_vault_protected_settings,omitempty"`
	SuppressFailures                 *bool                   `json:"suppress_failures,omitempty"`
	ForceUpdateTag                   *string                 `json:"force_update_tag,omitempty"`
	RerunClues                       []string                `json:"rerun_clues,omitempty"`
	ProvisionAfterExtensions         []string                `json:"provision_after_extensions,omitempty"`
	ProvisioningState                *string                 `json:"provisioning_state,omitempty"`
	InstanceViewStatuses             []string                `json:"instance_view_statuses,omitempty"`
	TargetIdentityIDs                []string                `json:"target_identity_ids,omitempty"`
	StrongestVisibleExecutionContext *PersistenceRoleContext `json:"strongest_visible_execution_context,omitempty"`
	VMSSOrchestrationMode            *string                 `json:"vmss_orchestration_mode,omitempty"`
	VMSSUpgradeMode                  *string                 `json:"vmss_upgrade_mode,omitempty"`
	NearbyThematicNames              []string                `json:"nearby_thematic_names,omitempty"`
}

type PersistenceVMExtension struct {
	ID                      string                      `json:"id"`
	Name                    string                      `json:"vm_extension"`
	ResourceGroup           string                      `json:"resource_group"`
	Location                string                      `json:"location"`
	CapabilitySteps         []PersistenceCapabilityStep `json:"capability_steps"`
	CurrentIdentityContext  *PersistenceRoleContext     `json:"current_identity_context,omitempty"`
	ExecutionContextOptions []string                    `json:"execution_context_options"`
	CurrentState            PersistenceVMExtensionState `json:"current_state"`
	StillUnmapped           []string                    `json:"still_unmapped"`
	Summary                 string                      `json:"summary"`
	RelatedIDs              []string                    `json:"related_ids"`
}

type PersistenceVMExtensionsOutput struct {
	Metadata           ScopedCommandMetadata    `json:"metadata"`
	GroupedCommandName string                   `json:"grouped_command_name"`
	Surface            string                   `json:"surface"`
	InputMode          string                   `json:"input_mode"`
	CommandState       string                   `json:"command_state"`
	Summary            string                   `json:"summary"`
	BackingCommands    []string                 `json:"backing_commands"`
	VMExtensions       []PersistenceVMExtension `json:"vm_extensions"`
	Issues             []Issue                  `json:"issues"`
}

type PersistenceAzureMLWorkspaceState struct {
	Classification                   string                  `json:"classification"`
	State                            *string                 `json:"state,omitempty"`
	PublicNetworkAccess              *string                 `json:"public_network_access,omitempty"`
	IdentityType                     *string                 `json:"identity_type,omitempty"`
	VisibleIdentityNames             []string                `json:"visible_identity_names"`
	ComputeCount                     *int                    `json:"compute_count,omitempty"`
	ComputeTypes                     []string                `json:"compute_types"`
	JobCount                         *int                    `json:"job_count,omitempty"`
	JobTypes                         []string                `json:"job_types"`
	ScheduleCount                    *int                    `json:"schedule_count,omitempty"`
	ScheduleTriggerTypes             []string                `json:"schedule_trigger_types"`
	EndpointCount                    *int                    `json:"endpoint_count,omitempty"`
	EndpointAuthModes                []string                `json:"endpoint_auth_modes"`
	EndpointPublicAccess             []string                `json:"endpoint_public_access"`
	DatastoreCount                   *int                    `json:"datastore_count,omitempty"`
	DatastoreTypes                   []string                `json:"datastore_types"`
	StrongestVisibleExecutionContext *PersistenceRoleContext `json:"strongest_visible_execution_context,omitempty"`
	NearbyThematicNames              []string                `json:"nearby_thematic_names,omitempty"`
}

type PersistenceAzureMLWorkspace struct {
	ID                      string                           `json:"id"`
	Name                    string                           `json:"workspace"`
	ResourceGroup           string                           `json:"resource_group"`
	Location                *string                          `json:"location,omitempty"`
	CapabilitySteps         []PersistenceCapabilityStep      `json:"capability_steps"`
	CurrentIdentityContext  *PersistenceRoleContext          `json:"current_identity_context,omitempty"`
	ExecutionContextOptions []string                         `json:"execution_context_options"`
	CurrentState            PersistenceAzureMLWorkspaceState `json:"current_state"`
	StillUnmapped           []string                         `json:"still_unmapped"`
	Summary                 string                           `json:"summary"`
	RelatedIDs              []string                         `json:"related_ids"`
}

type PersistenceAzureMLOutput struct {
	Metadata           ScopedCommandMetadata         `json:"metadata"`
	GroupedCommandName string                        `json:"grouped_command_name"`
	Surface            string                        `json:"surface"`
	InputMode          string                        `json:"input_mode"`
	CommandState       string                        `json:"command_state"`
	Summary            string                        `json:"summary"`
	BackingCommands    []string                      `json:"backing_commands"`
	Workspaces         []PersistenceAzureMLWorkspace `json:"workspaces"`
	Issues             []Issue                       `json:"issues"`
}

type PersistenceLogicAppWorkflowState struct {
	Classification                   string                  `json:"classification"`
	Platform                         *string                 `json:"platform,omitempty"`
	WorkflowKind                     *string                 `json:"workflow_kind,omitempty"`
	State                            *string                 `json:"state,omitempty"`
	TriggerTypes                     []string                `json:"trigger_types"`
	ExternallyCallableRequestTrigger bool                    `json:"externally_callable_request_trigger"`
	RecurrenceSummary                *string                 `json:"recurrence_summary,omitempty"`
	IdentityType                     *string                 `json:"identity_type,omitempty"`
	StrongestVisibleExecutionContext *PersistenceRoleContext `json:"strongest_visible_execution_context,omitempty"`
	NearbyThematicNames              []string                `json:"nearby_thematic_names,omitempty"`
	DownstreamActionKinds            []string                `json:"downstream_action_kinds"`
}

type PersistenceLogicAppWorkflow struct {
	ID                      string                           `json:"id"`
	Name                    string                           `json:"logic_app"`
	ResourceGroup           string                           `json:"resource_group"`
	Location                *string                          `json:"location,omitempty"`
	CapabilitySteps         []PersistenceCapabilityStep      `json:"capability_steps"`
	CurrentIdentityContext  *PersistenceRoleContext          `json:"current_identity_context,omitempty"`
	ExecutionContextOptions []string                         `json:"execution_context_options"`
	CurrentState            PersistenceLogicAppWorkflowState `json:"current_state"`
	StillUnmapped           []string                         `json:"still_unmapped"`
	Summary                 string                           `json:"summary"`
	RelatedIDs              []string                         `json:"related_ids"`
}

type PersistenceLogicAppsOutput struct {
	Metadata           ScopedCommandMetadata         `json:"metadata"`
	GroupedCommandName string                        `json:"grouped_command_name"`
	Surface            string                        `json:"surface"`
	InputMode          string                        `json:"input_mode"`
	CommandState       string                        `json:"command_state"`
	Summary            string                        `json:"summary"`
	BackingCommands    []string                      `json:"backing_commands"`
	Workflows          []PersistenceLogicAppWorkflow `json:"workflows"`
	Issues             []Issue                       `json:"issues"`
}
