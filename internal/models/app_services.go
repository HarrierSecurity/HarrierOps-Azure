package models

type AppServiceAsset struct {
	AppSettingsCount              *int     `json:"app_settings_count"`
	AppServicePlanID              *string  `json:"app_service_plan_id"`
	ClientCertEnabled             bool     `json:"client_cert_enabled"`
	ConnectionStringCount         *int     `json:"connection_string_count"`
	ConnectionStringTypes         []string `json:"connection_string_types,omitempty"`
	DefaultHostname               *string  `json:"default_hostname"`
	Deployment                    *string  `json:"deployment,omitempty"`
	DeploymentBranch              *string  `json:"deployment_branch,omitempty"`
	DeploymentIsGitHubAction      *bool    `json:"deployment_is_github_action,omitempty"`
	DeploymentManualIntegration   *bool    `json:"deployment_manual_integration,omitempty"`
	DeploymentRepoURL             *string  `json:"deployment_repo_url,omitempty"`
	FTPSState                     *string  `json:"ftps_state"`
	HTTPSOnly                     bool     `json:"https_only"`
	ID                            string   `json:"id"`
	KeyVaultConnectionStringCount *int     `json:"key_vault_connection_string_count"`
	KeyVaultReferenceCount        *int     `json:"key_vault_reference_count"`
	Location                      string   `json:"location"`
	MinTLSVersion                 *string  `json:"min_tls_version"`
	Name                          string   `json:"name"`
	PublicNetworkAccess           *string  `json:"public_network_access"`
	RelatedIDs                    []string `json:"related_ids"`
	ResourceGroup                 string   `json:"resource_group"`
	RunFromPackage                *bool    `json:"run_from_package"`
	RuntimeStack                  *string  `json:"runtime_stack"`
	SensitiveSettingCount         *int     `json:"sensitive_setting_count"`
	State                         *string  `json:"state"`
	Summary                       string   `json:"summary"`
	WorkloadClientID              *string  `json:"workload_client_id"`
	WorkloadIdentityIDs           []string `json:"workload_identity_ids"`
	WorkloadIdentityType          *string  `json:"workload_identity_type"`
	WorkloadPrincipalID           *string  `json:"workload_principal_id"`
}

type AppServicesMetadata = RuntimeCommandMetadata

type AppServicesOutput struct {
	AppServices []AppServiceAsset   `json:"app_services"`
	Findings    []Finding           `json:"findings"`
	Issues      []Issue             `json:"issues"`
	Metadata    AppServicesMetadata `json:"metadata"`
}
