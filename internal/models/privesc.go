package models

type PrivescPathSummary struct {
	Asset            *string  `json:"asset"`
	CurrentIdentity  bool     `json:"current_identity"`
	ImpactRoles      []string `json:"impact_roles"`
	StartingFoothold string   `json:"starting_foothold"`
	MissingProof     string   `json:"missing_proof"`
	NextReview       string   `json:"next_review"`
	OperatorSignal   string   `json:"operator_signal"`
	PathType         string   `json:"path_type"`
	Target           string   `json:"target"`
	Preferred        bool     `json:"preferred"`
	PreferredReason  string   `json:"preferred_reason"`
	Priority         string   `json:"priority"`
	Principal        string   `json:"principal"`
	PrincipalID      string   `json:"principal_id"`
	PrincipalType    string   `json:"principal_type"`
	ProvenPath       string   `json:"proven_path"`
	RelatedIDs       []string `json:"related_ids"`
	Summary          string   `json:"summary"`
}

type PrivescOutput struct {
	Issues   []Issue              `json:"issues"`
	Metadata PrincipalsMetadata   `json:"metadata"`
	Paths    []PrivescPathSummary `json:"paths"`
}
