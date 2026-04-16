package models

type AzureMLWorkspaceAsset struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"workspace"`
	Runtime               *string  `json:"runtime,omitempty"`
	Serving               *string  `json:"serving,omitempty"`
	Identity              *string  `json:"identity,omitempty"`
	Storage               *string  `json:"storage,omitempty"`
	Classification        string   `json:"classification"`
	ResourceGroup         string   `json:"resource_group"`
	Location              *string  `json:"location,omitempty"`
	WorkspaceKind         *string  `json:"workspace_kind,omitempty"`
	State                 *string  `json:"state,omitempty"`
	PublicNetworkAccess   *string  `json:"public_network_access,omitempty"`
	IdentityType          *string  `json:"identity_type,omitempty"`
	PrincipalID           *string  `json:"principal_id,omitempty"`
	IdentityIDs           []string `json:"identity_ids"`
	ComputeCount          int      `json:"compute_count"`
	ComputeTypes          []string `json:"compute_types"`
	JobCount              int      `json:"job_count"`
	JobTypes              []string `json:"job_types"`
	ScheduleCount         int      `json:"schedule_count"`
	ScheduleTriggerTypes  []string `json:"schedule_trigger_types"`
	EndpointCount         int      `json:"endpoint_count"`
	EndpointAuthModes     []string `json:"endpoint_auth_modes"`
	EndpointPublicAccess  []string `json:"endpoint_public_access"`
	DatastoreCount        int      `json:"datastore_count"`
	DatastoreTypes        []string `json:"datastore_types"`
	StorageAccountID      *string  `json:"storage_account_id,omitempty"`
	KeyVaultID            *string  `json:"key_vault_id,omitempty"`
	ContainerRegistryID   *string  `json:"container_registry_id,omitempty"`
	ApplicationInsightsID *string  `json:"application_insights_id,omitempty"`
	Summary               string   `json:"summary"`
	RelatedIDs            []string `json:"related_ids"`
}

type AzureMLMetadata = RuntimeCommandMetadata

type AzureMLOutput struct {
	Findings   []Finding               `json:"findings"`
	Issues     []Issue                 `json:"issues"`
	Metadata   AzureMLMetadata         `json:"metadata"`
	Workspaces []AzureMLWorkspaceAsset `json:"workspaces"`
}
