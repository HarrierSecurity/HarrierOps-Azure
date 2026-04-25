package models

type PrincipalSummary struct {
	AttachedTo          []string `json:"attached_to"`
	DisplayName         *string  `json:"display_name"`
	ID                  string   `json:"id"`
	IdentityNames       []string `json:"identity_names"`
	IdentityTypes       []string `json:"identity_types"`
	IsCurrentIdentity   bool     `json:"is_current_identity"`
	PrincipalType       string   `json:"principal_type"`
	RoleAssignmentCount int      `json:"role_assignment_count"`
	RoleNames           []string `json:"role_names"`
	ScopeIDs            []string `json:"scope_ids"`
	Sources             []string `json:"sources"`
	TenantID            *string  `json:"tenant_id"`
}

type PrincipalsMetadata struct {
	AuthMode           *string           `json:"auth_mode"`
	Command            string            `json:"command"`
	DevOpsOrganization *string           `json:"devops_organization"`
	GeneratedAt        string            `json:"generated_at"`
	SchemaVersion      string            `json:"schema_version"`
	SubscriptionID     *string           `json:"subscription_id"`
	TenantID           *string           `json:"tenant_id"`
	TokenSource        *string           `json:"token_source"`
	ArtifactContext    *ArtifactContext  `json:"artifact_context,omitempty"`
	SessionArtifacts   []SessionArtifact `json:"session_artifacts,omitempty"`
}

type PrincipalsOutput struct {
	Issues     []Issue            `json:"issues"`
	Metadata   PrincipalsMetadata `json:"metadata"`
	Principals []PrincipalSummary `json:"principals"`
}
