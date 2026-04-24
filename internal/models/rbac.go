package models

type RoleAssignment struct {
	ID               string   `json:"id"`
	PrincipalID      string   `json:"principal_id"`
	PrincipalType    string   `json:"principal_type"`
	RoleDefinitionID string   `json:"role_definition_id"`
	RoleName         string   `json:"role_name"`
	ScopeID          string   `json:"scope_id"`
	Actions          []string `json:"-"`
	NotActions       []string `json:"-"`
	DataActions      []string `json:"-"`
	NotDataActions   []string `json:"-"`
}

type RbacOutput struct {
	Issues          []Issue          `json:"issues"`
	Metadata        Metadata         `json:"metadata"`
	Principals      []Principal      `json:"principals"`
	RoleAssignments []RoleAssignment `json:"role_assignments"`
	Scopes          []ScopeRef       `json:"scopes"`
}
