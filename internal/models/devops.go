package models

type DevopsTrustedInput struct {
	InputType                    string   `json:"input_type"`
	Ref                          string   `json:"ref"`
	VisibilityState              string   `json:"visibility_state"`
	CurrentOperatorAccessState   string   `json:"current_operator_access_state"`
	CurrentOperatorCanPoison     bool     `json:"current_operator_can_poison"`
	TrustedInputEvidenceBasis    string   `json:"trusted_input_evidence_basis"`
	TrustedInputPermissionSource string   `json:"trusted_input_permission_source"`
	TrustedInputPermissionDetail string   `json:"trusted_input_permission_detail"`
	SurfaceTypes                 []string `json:"surface_types"`
	JoinIDs                      []string `json:"join_ids"`
}

type DevopsPipelineAsset struct {
	ID                                    string               `json:"id"`
	DefinitionID                          string               `json:"definition_id"`
	Name                                  string               `json:"name"`
	ProjectID                             string               `json:"project_id"`
	ProjectName                           string               `json:"project_name"`
	Path                                  string               `json:"path"`
	RepositoryID                          *string              `json:"repository_id"`
	RepositoryName                        string               `json:"repository_name"`
	RepositoryType                        string               `json:"repository_type"`
	RepositoryURL                         string               `json:"repository_url"`
	RepositoryHostType                    string               `json:"repository_host_type"`
	SourceVisibilityState                 string               `json:"source_visibility_state"`
	DefaultBranch                         string               `json:"default_branch"`
	TriggerTypes                          []string             `json:"trigger_types"`
	VariableGroupNames                    []string             `json:"variable_group_names"`
	SecretVariableCount                   int                  `json:"secret_variable_count"`
	SecretVariableNames                   []string             `json:"secret_variable_names"`
	KeyVaultGroupNames                    []string             `json:"key_vault_group_names"`
	KeyVaultNames                         []string             `json:"key_vault_names"`
	AzureServiceConnectionNames           []string             `json:"azure_service_connection_names"`
	AzureServiceConnectionTypes           []string             `json:"azure_service_connection_types"`
	AzureServiceConnectionAuthSchemes     []string             `json:"azure_service_connection_auth_schemes"`
	AzureServiceConnectionIDs             []string             `json:"azure_service_connection_ids"`
	AzureServiceConnectionPrincipalIDs    []string             `json:"azure_service_connection_principal_ids"`
	AzureServiceConnectionClientIDs       []string             `json:"azure_service_connection_client_ids"`
	AzureServiceConnectionTenantIDs       []string             `json:"azure_service_connection_tenant_ids"`
	AzureServiceConnectionSubscriptionIDs []string             `json:"azure_service_connection_subscription_ids"`
	TargetClues                           []string             `json:"target_clues"`
	RiskCues                              []string             `json:"risk_cues"`
	ExecutionModes                        []string             `json:"execution_modes"`
	UpstreamSources                       []string             `json:"upstream_sources"`
	TrustedInputs                         []DevopsTrustedInput `json:"trusted_inputs"`
	TrustedInputTypes                     []string             `json:"trusted_input_types"`
	TrustedInputRefs                      []string             `json:"trusted_input_refs"`
	TrustedInputJoinIDs                   []string             `json:"trusted_input_join_ids"`
	PrimaryInjectionSurface               string               `json:"primary_injection_surface"`
	PrimaryTrustedInputRef                string               `json:"primary_trusted_input_ref"`
	SourceJoinIDs                         []string             `json:"source_join_ids"`
	TriggerJoinIDs                        []string             `json:"trigger_join_ids"`
	IdentityJoinIDs                       []string             `json:"identity_join_ids"`
	SecretSupportTypes                    []string             `json:"secret_support_types"`
	SecretDependencyIDs                   []string             `json:"secret_dependency_ids"`
	InjectionSurfaceTypes                 []string             `json:"injection_surface_types"`
	CurrentOperatorInjectionSurfaceTypes  []string             `json:"current_operator_injection_surface_types"`
	EditPathState                         string               `json:"edit_path_state"`
	QueuePathState                        string               `json:"queue_path_state"`
	RerunPathState                        string               `json:"rerun_path_state"`
	ApprovalPathState                     string               `json:"approval_path_state"`
	CurrentOperatorCanViewDefinition      *bool                `json:"current_operator_can_view_definition"`
	CurrentOperatorCanQueue               *bool                `json:"current_operator_can_queue"`
	CurrentOperatorCanEdit                *bool                `json:"current_operator_can_edit"`
	CurrentOperatorCanViewSource          *bool                `json:"current_operator_can_view_source"`
	CurrentOperatorCanContributeSource    *bool                `json:"current_operator_can_contribute_source"`
	ConsequenceTypes                      []string             `json:"consequence_types"`
	MissingExecutionPath                  bool                 `json:"missing_execution_path"`
	MissingInjectionPoint                 bool                 `json:"missing_injection_point"`
	MissingTargetMapping                  bool                 `json:"missing_target_mapping"`
	PartialRead                           bool                 `json:"partial_read"`
	Summary                               string               `json:"summary"`
	RelatedIDs                            []string             `json:"related_ids"`
}

type DevopsMetadata struct {
	SchemaVersion      string  `json:"schema_version"`
	Command            string  `json:"command"`
	GeneratedAt        string  `json:"generated_at"`
	TenantID           *string `json:"tenant_id"`
	SubscriptionID     *string `json:"subscription_id"`
	DevOpsOrganization *string `json:"devops_organization,omitempty"`
	TokenSource        *string `json:"token_source"`
	AuthMode           *string `json:"auth_mode,omitempty"`
}

type DevopsOutput struct {
	Metadata  DevopsMetadata        `json:"metadata"`
	Pipelines []DevopsPipelineAsset `json:"pipelines"`
	Findings  []Finding             `json:"findings"`
	Issues    []Issue               `json:"issues"`
}
