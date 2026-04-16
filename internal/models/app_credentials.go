package models

type AppCredentialSummary struct {
	RowClass                    string   `json:"row_class"`
	TargetObjectType            string   `json:"target_object_type"`
	TargetObjectID              string   `json:"target_object_id"`
	TargetObjectName            string   `json:"target_object_name"`
	BackingServicePrincipalID   *string  `json:"backing_service_principal_id,omitempty"`
	BackingServicePrincipalName *string  `json:"backing_service_principal_name,omitempty"`
	CredentialType              *string  `json:"credential_type,omitempty"`
	ControlPath                 string   `json:"control_path"`
	RoleContext                 string   `json:"role_context"`
	TenantContext               string   `json:"tenant_context"`
	CurrentEvidence             string   `json:"current_evidence"`
	MissingProof                string   `json:"missing_proof"`
	OperatorActionability       string   `json:"operator_actionability"`
	RecommendedFixFocus         string   `json:"recommended_fix_focus"`
	Summary                     string   `json:"summary"`
	RelatedIDs                  []string `json:"related_ids"`
}

type AppCredentialsOutput struct {
	Metadata       ScopedCommandMetadata  `json:"metadata"`
	AppCredentials []AppCredentialSummary `json:"app_credentials"`
	Issues         []Issue                `json:"issues"`
}
