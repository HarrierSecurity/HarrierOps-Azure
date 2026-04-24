package models

type ContainerAppsJobEventRule struct {
	AuthSecretRefs []string `json:"auth_secret_refs"`
	Identity       *string  `json:"identity"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
}

type ContainerAppsJobAsset struct {
	Command                  []string                    `json:"command"`
	ContainerImages          []string                    `json:"container_images"`
	EnvironmentID            *string                     `json:"environment_id"`
	EventRules               []ContainerAppsJobEventRule `json:"event_rules"`
	ID                       string                      `json:"id"`
	KeyVaultSecretCount      *int                        `json:"key_vault_secret_count"`
	Location                 string                      `json:"location"`
	Name                     string                      `json:"name"`
	Parallelism              *int                        `json:"parallelism"`
	RegistryIdentityCount    *int                        `json:"registry_identity_count"`
	RegistryPasswordRefCount *int                        `json:"registry_password_ref_count"`
	RegistryServers          []string                    `json:"registry_servers"`
	RelatedIDs               []string                    `json:"related_ids"`
	ReplicaCompletionCount   *int                        `json:"replica_completion_count"`
	ReplicaRetryLimit        *int                        `json:"replica_retry_limit"`
	ReplicaTimeout           *int                        `json:"replica_timeout"`
	ResourceGroup            string                      `json:"resource_group"`
	ScheduleExpression       *string                     `json:"schedule_expression,omitempty"`
	SecretCount              *int                        `json:"secret_count"`
	Summary                  string                      `json:"summary"`
	TriggerType              *string                     `json:"trigger_type"`
	WorkloadClientID         *string                     `json:"workload_client_id"`
	WorkloadIdentityIDs      []string                    `json:"workload_identity_ids"`
	WorkloadIdentityType     *string                     `json:"workload_identity_type"`
	WorkloadPrincipalID      *string                     `json:"workload_principal_id"`
}

type ContainerAppsJobsMetadata = RuntimeCommandMetadata

type ContainerAppsJobsOutput struct {
	ContainerAppsJobs []ContainerAppsJobAsset   `json:"container_apps_jobs"`
	Findings          []Finding                 `json:"findings"`
	Issues            []Issue                   `json:"issues"`
	Metadata          ContainerAppsJobsMetadata `json:"metadata"`
}
