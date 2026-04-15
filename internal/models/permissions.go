package models

type PermissionRow struct {
	PrincipalID         string   `json:"principal_id"`
	DisplayName         string   `json:"display_name"`
	PrincipalType       string   `json:"principal_type"`
	Priority            string   `json:"priority"`
	HighImpactRoles     []string `json:"high_impact_roles"`
	AllRoleNames        []string `json:"all_role_names"`
	RoleAssignmentCount int      `json:"role_assignment_count"`
	ScopeCount          int      `json:"scope_count"`
	ScopeIDs            []string `json:"scope_ids"`
	Privileged          bool     `json:"privileged"`
	IsCurrentIdentity   bool     `json:"is_current_identity"`
	OperatorSignal      string   `json:"operator_signal"`
	NextReview          string   `json:"next_review"`
	Summary             string   `json:"summary"`
}

type PermissionsOutput struct {
	Metadata    PermissionsMetadata `json:"metadata"`
	Permissions []PermissionRow     `json:"permissions"`
	Issues      []Issue             `json:"issues"`
}
