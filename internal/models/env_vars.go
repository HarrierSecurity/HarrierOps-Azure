package models

type EnvVarSummary struct {
	AssetID                   string                `json:"asset_id"`
	AssetKind                 string                `json:"asset_kind"`
	AssetName                 string                `json:"asset_name"`
	KeyVaultReferenceIdentity *string               `json:"key_vault_reference_identity"`
	Location                  string                `json:"location"`
	LooksSensitive            bool                  `json:"looks_sensitive"`
	ReferenceTarget           *string               `json:"reference_target"`
	RelatedIDs                []string              `json:"related_ids"`
	ResourceGroup             string                `json:"resource_group"`
	SettingName               string                `json:"setting_name"`
	Summary                   string                `json:"summary"`
	ValueType                 string                `json:"value_type"`
	WorkloadClientID          *string               `json:"workload_client_id"`
	WorkloadIdentityIDs       []string              `json:"workload_identity_ids"`
	WorkloadIdentityType      *string               `json:"workload_identity_type"`
	WorkloadPrincipalID       *string               `json:"workload_principal_id"`
	TargetServices            []EnvVarTargetService `json:"-"`
}

type EnvVarFinding struct {
	Description string   `json:"description"`
	ID          string   `json:"id"`
	RelatedIDs  []string `json:"related_ids"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
}

type EnvVarsOutput struct {
	EnvVars  []EnvVarSummary `json:"env_vars"`
	Findings []EnvVarFinding `json:"findings"`
	Issues   []Issue         `json:"issues"`
	Metadata Metadata        `json:"metadata"`
}
