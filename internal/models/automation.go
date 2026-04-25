package models

type AutomationAccountAsset struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	ResourceGroup          string   `json:"resource_group"`
	Location               *string  `json:"location"`
	State                  *string  `json:"state"`
	SKUName                *string  `json:"sku_name"`
	IdentityType           *string  `json:"identity_type"`
	PrincipalID            *string  `json:"principal_id"`
	ClientID               *string  `json:"client_id"`
	IdentityIDs            []string `json:"identity_ids"`
	RunbookCount           *int     `json:"runbook_count"`
	PublishedRunbookCount  *int     `json:"published_runbook_count"`
	PublishedRunbookNames  []string `json:"published_runbook_names"`
	RunbookTypes           []string `json:"runbook_types"`
	RunbookCommandClues    []string `json:"runbook_command_clues"`
	RunbookResourceClues   []string `json:"runbook_resource_clues"`
	ScheduleCount          *int     `json:"schedule_count"`
	ScheduleDefinitions    []string `json:"schedule_definitions"`
	JobScheduleCount       *int     `json:"job_schedule_count"`
	WebhookCount           *int     `json:"webhook_count"`
	HybridWorkerGroupCount *int     `json:"hybrid_worker_group_count"`
	CredentialCount        *int     `json:"credential_count"`
	CertificateCount       *int     `json:"certificate_count"`
	ConnectionCount        *int     `json:"connection_count"`
	VariableCount          *int     `json:"variable_count"`
	EncryptedVariableCount *int     `json:"encrypted_variable_count"`
	StartModes             []string `json:"start_modes"`
	PrimaryStartMode       *string  `json:"primary_start_mode"`
	PrimaryRunbookName     *string  `json:"primary_runbook_name"`
	ScheduleRunbookNames   []string `json:"schedule_runbook_names"`
	WebhookRunbookNames    []string `json:"webhook_runbook_names"`
	HybridWorkerGroupIDs   []string `json:"hybrid_worker_group_ids"`
	TriggerJoinIDs         []string `json:"trigger_join_ids"`
	IdentityJoinIDs        []string `json:"identity_join_ids"`
	SecretSupportTypes     []string `json:"secret_support_types"`
	SecretDependencyIDs    []string `json:"secret_dependency_ids"`
	ConsequenceTypes       []string `json:"consequence_types"`
	MissingExecutionPath   bool     `json:"missing_execution_path"`
	MissingTargetMapping   bool     `json:"missing_target_mapping"`
	Summary                string   `json:"summary"`
	RelatedIDs             []string `json:"related_ids"`
}

type AutomationMetadata struct {
	SchemaVersion    string            `json:"schema_version"`
	Command          string            `json:"command"`
	GeneratedAt      string            `json:"generated_at"`
	TenantID         *string           `json:"tenant_id"`
	SubscriptionID   *string           `json:"subscription_id"`
	TokenSource      *string           `json:"token_source"`
	AuthMode         *string           `json:"auth_mode,omitempty"`
	ArtifactContext  *ArtifactContext  `json:"artifact_context,omitempty"`
	SessionArtifacts []SessionArtifact `json:"session_artifacts,omitempty"`
}

type AutomationOutput struct {
	Metadata           AutomationMetadata       `json:"metadata"`
	AutomationAccounts []AutomationAccountAsset `json:"automation_accounts"`
	Findings           []Finding                `json:"findings"`
	Issues             []Issue                  `json:"issues"`
}
