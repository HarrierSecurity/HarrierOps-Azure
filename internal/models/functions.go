package models

type FunctionAppAsset struct {
	AlwaysOn                           *bool    `json:"always_on"`
	AppServicePlanID                   *string  `json:"app_service_plan_id"`
	AzureWebJobsStorageReferenceTarget *string  `json:"azure_webjobs_storage_reference_target"`
	AzureWebJobsStorageValueType       *string  `json:"azure_webjobs_storage_value_type"`
	ClientCertEnabled                  bool     `json:"client_cert_enabled"`
	DefaultHostname                    *string  `json:"default_hostname"`
	FTPSState                          *string  `json:"ftps_state"`
	FunctionsExtensionVersion          *string  `json:"functions_extension_version"`
	HTTPSOnly                          bool     `json:"https_only"`
	ID                                 string   `json:"id"`
	KeyVaultReferenceCount             *int     `json:"key_vault_reference_count"`
	Location                           string   `json:"location"`
	MinTLSVersion                      *string  `json:"min_tls_version"`
	Name                               string   `json:"name"`
	PublicNetworkAccess                *string  `json:"public_network_access"`
	RelatedIDs                         []string `json:"related_ids"`
	ResourceGroup                      string   `json:"resource_group"`
	RunFromPackage                     *bool    `json:"run_from_package"`
	RuntimeStack                       *string  `json:"runtime_stack"`
	State                              *string  `json:"state"`
	Summary                            string   `json:"summary"`
	WorkloadClientID                   *string  `json:"workload_client_id"`
	WorkloadIdentityIDs                []string `json:"workload_identity_ids"`
	WorkloadIdentityType               *string  `json:"workload_identity_type"`
	WorkloadPrincipalID                *string  `json:"workload_principal_id"`
}

type FunctionsMetadata = RuntimeCommandMetadata

type FunctionsOutput struct {
	Findings     []Finding          `json:"findings"`
	FunctionApps []FunctionAppAsset `json:"function_apps"`
	Issues       []Issue            `json:"issues"`
	Metadata     FunctionsMetadata  `json:"metadata"`
}
