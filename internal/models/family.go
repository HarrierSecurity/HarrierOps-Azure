package models

type FamilySurfaceDescriptor struct {
	Surface          string   `json:"surface"`
	State            string   `json:"state"`
	Summary          string   `json:"summary"`
	OperatorQuestion string   `json:"operator_question"`
	BackingCommands  []string `json:"backing_commands"`
}

type FamilyCapabilityStep struct {
	Action           string `json:"action"`
	APISurface       string `json:"api_surface"`
	Status           string `json:"status"`
	CanAct           bool   `json:"-"`
	DownstreamEffect string `json:"downstream_effect"`
	Boundary         string `json:"boundary"`
}

type FamilyRoleContext struct {
	Name         string   `json:"name"`
	Kind         string   `json:"kind"`
	PrincipalID  *string  `json:"principal_id,omitempty"`
	RoleNames    []string `json:"role_names"`
	ScopeIDs     []string `json:"scope_ids"`
	ControlLabel string   `json:"control_label,omitempty"`
	Summary      string   `json:"summary"`
}

type FamilyBoundaryNote struct {
	Name           string `json:"name"`
	Classification string `json:"classification"`
	Reason         string `json:"reason"`
}

type FamilyLogicAppState struct {
	Platform                         *string  `json:"platform,omitempty"`
	State                            *string  `json:"state,omitempty"`
	TriggerTypes                     []string `json:"trigger_types"`
	ExternallyCallableRequestTrigger bool     `json:"externally_callable_request_trigger"`
	RecurrenceSummary                *string  `json:"recurrence_summary,omitempty"`
	DownstreamActionKinds            []string `json:"downstream_action_kinds"`
	ConnectorReferences              []string `json:"connector_references"`
	ParameterNames                   []string `json:"parameter_names"`
	DownstreamResourceReferences     []string `json:"downstream_resource_references"`
	IdentityType                     *string  `json:"identity_type,omitempty"`
	IdentityIDs                      []string `json:"identity_ids"`
	Posture                          string   `json:"posture"`
}
