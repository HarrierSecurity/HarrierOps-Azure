package models

type FunctionBinding struct {
	Direction string `json:"direction,omitempty"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
}

type FunctionChildAsset struct {
	BindingTypes      []string          `json:"binding_types,omitempty"`
	Bindings          []FunctionBinding `json:"bindings,omitempty"`
	Config            map[string]any    `json:"config,omitempty"`
	ID                string            `json:"id"`
	InvokeURLTemplate *string           `json:"invoke_url_template,omitempty"`
	IsDisabled        *bool             `json:"is_disabled,omitempty"`
	Language          *string           `json:"language,omitempty"`
	Name              string            `json:"name"`
	TriggerType       *string           `json:"trigger_type,omitempty"`
}

type FunctionAttachedIdentity struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	PrincipalID *string `json:"principal_id,omitempty"`
	ClientID    *string `json:"client_id,omitempty"`
}

type FunctionAppAsset struct {
	AlwaysOn                           *bool                      `json:"always_on"`
	AppServicePlanID                   *string                    `json:"app_service_plan_id"`
	AzureWebJobsStorageReferenceTarget *string                    `json:"azure_webjobs_storage_reference_target"`
	AzureWebJobsStorageValueType       *string                    `json:"azure_webjobs_storage_value_type"`
	ClientCertEnabled                  bool                       `json:"client_cert_enabled"`
	DefaultHostname                    *string                    `json:"hostname"`
	Deployment                         *string                    `json:"deployment,omitempty"`
	FTPSState                          *string                    `json:"ftps_state"`
	FunctionsExtensionVersion          *string                    `json:"functions_extension_version"`
	HTTPSOnly                          bool                       `json:"https_only"`
	ID                                 string                     `json:"id"`
	Identity                           *string                    `json:"identity,omitempty"`
	KeyVaultReferenceCount             *int                       `json:"key_vault_reference_count"`
	Location                           string                     `json:"location"`
	MinTLSVersion                      *string                    `json:"min_tls_version"`
	Name                               string                     `json:"function_app"`
	PublicNetworkAccess                *string                    `json:"public_network_access"`
	RelatedIDs                         []string                   `json:"related_ids"`
	ResourceGroup                      string                     `json:"resource_group"`
	RunFromPackage                     *bool                      `json:"run_from_package"`
	Runtime                            *string                    `json:"runtime,omitempty"`
	RuntimeStack                       *string                    `json:"runtime_stack"`
	State                              *string                    `json:"state"`
	Summary                            string                     `json:"summary"`
	TriggerTypes                       []string                   `json:"trigger_types,omitempty"`
	UserAssignedIdentities             []FunctionAttachedIdentity `json:"user_assigned_identities,omitempty"`
	VisibleFunctions                   []FunctionChildAsset       `json:"visible_functions,omitempty"`
	WorkloadClientID                   *string                    `json:"workload_client_id"`
	WorkloadIdentityIDs                []string                   `json:"workload_identity_ids"`
	WorkloadIdentityType               *string                    `json:"workload_identity_type"`
	WorkloadPrincipalID                *string                    `json:"workload_principal_id"`
}

type FunctionsMetadata = RuntimeCommandMetadata

type FunctionsOutput struct {
	Findings     []Finding          `json:"findings"`
	FunctionApps []FunctionAppAsset `json:"function_apps"`
	Issues       []Issue            `json:"issues"`
	Metadata     FunctionsMetadata  `json:"metadata"`
}
