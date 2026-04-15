package models

type ContainerInstanceAsset struct {
	ContainerCount       *int     `json:"container_count"`
	ContainerImages      []string `json:"container_images"`
	ExposedPorts         []int    `json:"exposed_ports"`
	FQDN                 *string  `json:"fqdn"`
	ID                   string   `json:"id"`
	Location             string   `json:"location"`
	Name                 string   `json:"name"`
	OSType               *string  `json:"os_type"`
	ProvisioningState    *string  `json:"provisioning_state"`
	PublicIPAddress      *string  `json:"public_ip_address"`
	RelatedIDs           []string `json:"related_ids"`
	ResourceGroup        string   `json:"resource_group"`
	RestartPolicy        *string  `json:"restart_policy"`
	SubnetIDs            []string `json:"subnet_ids"`
	Summary              string   `json:"summary"`
	WorkloadClientID     *string  `json:"workload_client_id"`
	WorkloadIdentityIDs  []string `json:"workload_identity_ids"`
	WorkloadIdentityType *string  `json:"workload_identity_type"`
	WorkloadPrincipalID  *string  `json:"workload_principal_id"`
}

type ContainerInstancesMetadata = RuntimeCommandMetadata

type ContainerInstancesOutput struct {
	ContainerInstances []ContainerInstanceAsset   `json:"container_instances"`
	Findings           []Finding                  `json:"findings"`
	Issues             []Issue                    `json:"issues"`
	Metadata           ContainerInstancesMetadata `json:"metadata"`
}
