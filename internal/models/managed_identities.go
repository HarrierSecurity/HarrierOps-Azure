package models

type ManagedIdentity struct {
	ID                   string           `json:"id"`
	Name                 string           `json:"name"`
	IdentityType         string           `json:"identity_type"`
	PrincipalID          *string          `json:"principal_id"`
	ClientID             *string          `json:"client_id"`
	AttachedTo           []string         `json:"attached_to"`
	ScopeIDs             []string         `json:"scope_ids"`
	OperatorSignal       *string          `json:"operator_signal"`
	NextReview           *string          `json:"next_review"`
	Summary              *string          `json:"summary"`
	WorkloadExposure     WorkloadExposure `json:"-"`
	DirectControlVisible bool             `json:"-"`
}

type ManagedIdentityRoleAssignment struct {
	ID               string `json:"id"`
	ScopeID          string `json:"scope_id"`
	PrincipalID      string `json:"principal_id"`
	PrincipalType    string `json:"principal_type"`
	RoleDefinitionID string `json:"role_definition_id"`
	RoleName         string `json:"role_name"`
}

type ManagedIdentityFinding struct {
	ID          string   `json:"id"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	RelatedIDs  []string `json:"related_ids,omitempty"`
}

type ManagedIdentitiesOutput struct {
	Metadata        ScopedCommandMetadata           `json:"metadata"`
	Identities      []ManagedIdentity               `json:"identities"`
	RoleAssignments []ManagedIdentityRoleAssignment `json:"role_assignments"`
	Findings        []ManagedIdentityFinding        `json:"findings"`
	Issues          []Issue                         `json:"issues"`
}
