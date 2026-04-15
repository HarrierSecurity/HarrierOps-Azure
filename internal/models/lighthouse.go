package models

type LighthouseDelegationAsset struct {
	AuthorizationCount          int      `json:"authorization_count"`
	DefinitionProvisioningState *string  `json:"definition_provisioning_state"`
	Description                 *string  `json:"description"`
	EligibleAuthorizationCount  int      `json:"eligible_authorization_count"`
	EligiblePrincipalCount      int      `json:"eligible_principal_count"`
	EligibleRoleNames           []string `json:"eligible_role_names"`
	HasDelegatedRoleAssignments bool     `json:"has_delegated_role_assignments"`
	HasOwnerRole                bool     `json:"has_owner_role"`
	HasUserAccessAdministrator  bool     `json:"has_user_access_administrator"`
	ID                          string   `json:"id"`
	ManagedByTenantID           *string  `json:"managed_by_tenant_id"`
	ManagedByTenantName         *string  `json:"managed_by_tenant_name"`
	ManageeTenantID             *string  `json:"managee_tenant_id"`
	ManageeTenantName           *string  `json:"managee_tenant_name"`
	Name                        string   `json:"name"`
	PlanName                    *string  `json:"plan_name"`
	PlanProduct                 *string  `json:"plan_product"`
	PlanPublisher               *string  `json:"plan_publisher"`
	PrincipalCount              int      `json:"principal_count"`
	ProvisioningState           *string  `json:"provisioning_state"`
	RegistrationDefinitionID    *string  `json:"registration_definition_id"`
	RegistrationDefinitionName  *string  `json:"registration_definition_name"`
	RelatedIDs                  []string `json:"related_ids"`
	ResourceGroup               *string  `json:"resource_group"`
	RoleNames                   []string `json:"role_names"`
	ScopeDisplayName            *string  `json:"scope_display_name"`
	ScopeID                     string   `json:"scope_id"`
	ScopeType                   string   `json:"scope_type"`
	StrongestRoleName           *string  `json:"strongest_role_name"`
	Summary                     string   `json:"summary"`
}

type LighthouseMetadata = Metadata

type LighthouseOutput struct {
	Findings              []Finding                   `json:"findings"`
	Issues                []Issue                     `json:"issues"`
	LighthouseDelegations []LighthouseDelegationAsset `json:"lighthouse_delegations"`
	Metadata              LighthouseMetadata          `json:"metadata"`
}
