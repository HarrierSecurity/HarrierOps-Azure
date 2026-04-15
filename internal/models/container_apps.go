package models

type ContainerAppAsset struct {
	DefaultHostname         *string  `json:"default_hostname"`
	EnvironmentID           *string  `json:"environment_id"`
	ExternalIngressEnabled  *bool    `json:"external_ingress_enabled"`
	ID                      string   `json:"id"`
	IngressTargetPort       *int     `json:"ingress_target_port"`
	IngressTransport        *string  `json:"ingress_transport"`
	LatestReadyRevisionName *string  `json:"latest_ready_revision_name"`
	LatestRevisionName      *string  `json:"latest_revision_name"`
	Location                string   `json:"location"`
	Name                    string   `json:"name"`
	RelatedIDs              []string `json:"related_ids"`
	ResourceGroup           string   `json:"resource_group"`
	RevisionMode            *string  `json:"revision_mode"`
	Summary                 string   `json:"summary"`
	WorkloadClientID        *string  `json:"workload_client_id"`
	WorkloadIdentityIDs     []string `json:"workload_identity_ids"`
	WorkloadIdentityType    *string  `json:"workload_identity_type"`
	WorkloadPrincipalID     *string  `json:"workload_principal_id"`
}

type ContainerAppsMetadata = RuntimeCommandMetadata

type ContainerAppsOutput struct {
	ContainerApps []ContainerAppAsset   `json:"container_apps"`
	Findings      []Finding             `json:"findings"`
	Issues        []Issue               `json:"issues"`
	Metadata      ContainerAppsMetadata `json:"metadata"`
}
