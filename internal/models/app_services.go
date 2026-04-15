package models

type AppServiceAsset struct {
	AppServicePlanID     *string  `json:"app_service_plan_id"`
	ClientCertEnabled    bool     `json:"client_cert_enabled"`
	DefaultHostname      *string  `json:"default_hostname"`
	FTPSState            *string  `json:"ftps_state"`
	HTTPSOnly            bool     `json:"https_only"`
	ID                   string   `json:"id"`
	Location             string   `json:"location"`
	MinTLSVersion        *string  `json:"min_tls_version"`
	Name                 string   `json:"name"`
	PublicNetworkAccess  *string  `json:"public_network_access"`
	RelatedIDs           []string `json:"related_ids"`
	ResourceGroup        string   `json:"resource_group"`
	RuntimeStack         *string  `json:"runtime_stack"`
	State                *string  `json:"state"`
	Summary              string   `json:"summary"`
	WorkloadClientID     *string  `json:"workload_client_id"`
	WorkloadIdentityIDs  []string `json:"workload_identity_ids"`
	WorkloadIdentityType *string  `json:"workload_identity_type"`
	WorkloadPrincipalID  *string  `json:"workload_principal_id"`
}

type AppServicesMetadata = RuntimeCommandMetadata

type AppServicesOutput struct {
	AppServices []AppServiceAsset   `json:"app_services"`
	Findings    []Finding           `json:"findings"`
	Issues      []Issue             `json:"issues"`
	Metadata    AppServicesMetadata `json:"metadata"`
}
