package models

type RoleTrustSummary struct {
	TrustType                   string                `json:"trust_type"`
	SourceObjectID              string                `json:"source_object_id"`
	SourceName                  *string               `json:"source_name"`
	SourceType                  string                `json:"source_type"`
	TargetObjectID              string                `json:"target_object_id"`
	TargetName                  *string               `json:"target_name"`
	TargetType                  string                `json:"target_type"`
	EvidenceType                string                `json:"evidence_type"`
	Confidence                  string                `json:"confidence"`
	ControlPrimitive            *string               `json:"control_primitive"`
	ControlledObjectType        *string               `json:"controlled_object_type"`
	ControlledObjectName        *string               `json:"controlled_object_name"`
	BackingServicePrincipalID   *string               `json:"backing_service_principal_id"`
	BackingServicePrincipalName *string               `json:"backing_service_principal_name"`
	EscalationMechanism         *string               `json:"escalation_mechanism"`
	UsableIdentityResult        *string               `json:"usable_identity_result"`
	DefenderCutPoint            *string               `json:"defender_cut_point"`
	OperatorSignal              *string               `json:"operator_signal"`
	NextReview                  *string               `json:"next_review"`
	Summary                     string                `json:"summary"`
	RelatedIDs                  []string              `json:"related_ids"`
	FollowOnKind                RoleTrustFollowOnKind `json:"-"`
}

type RoleTrustsOutput struct {
	Metadata ScopedCommandMetadata `json:"metadata"`
	Mode     RoleTrustsMode        `json:"mode"`
	Trusts   []RoleTrustSummary    `json:"trusts"`
	Issues   []Issue               `json:"issues"`
}
