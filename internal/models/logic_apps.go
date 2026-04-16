package models

type LogicAppWorkflowAsset struct {
	ID                               string   `json:"id"`
	Name                             string   `json:"logic_app"`
	Trigger                          *string  `json:"trigger,omitempty"`
	Identity                         *string  `json:"identity,omitempty"`
	Downstream                       *string  `json:"downstream,omitempty"`
	Classification                   string   `json:"classification"`
	ResourceGroup                    string   `json:"resource_group"`
	Location                         *string  `json:"location,omitempty"`
	Platform                         *string  `json:"platform,omitempty"`
	WorkflowKind                     *string  `json:"workflow_kind,omitempty"`
	State                            *string  `json:"state,omitempty"`
	IdentityType                     *string  `json:"identity_type,omitempty"`
	PrincipalID                      *string  `json:"principal_id,omitempty"`
	ClientID                         *string  `json:"client_id,omitempty"`
	IdentityIDs                      []string `json:"identity_ids"`
	TriggerTypes                     []string `json:"trigger_types"`
	ExternallyCallableRequestTrigger bool     `json:"externally_callable_request_trigger"`
	RecurrenceSummary                *string  `json:"recurrence_summary,omitempty"`
	DownstreamActionKinds            []string `json:"downstream_action_kinds"`
	Summary                          string   `json:"summary"`
	RelatedIDs                       []string `json:"related_ids"`
}

type LogicAppsMetadata = RuntimeCommandMetadata

type LogicAppsOutput struct {
	Findings  []Finding               `json:"findings"`
	Issues    []Issue                 `json:"issues"`
	Metadata  LogicAppsMetadata       `json:"metadata"`
	Workflows []LogicAppWorkflowAsset `json:"workflows"`
}
