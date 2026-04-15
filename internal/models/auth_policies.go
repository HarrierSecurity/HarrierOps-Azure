package models

type AuthPolicySummary struct {
	Controls   []string `json:"controls"`
	Name       string   `json:"name"`
	PolicyType string   `json:"policy_type"`
	RelatedIDs []string `json:"related_ids"`
	Scope      *string  `json:"scope"`
	State      string   `json:"state"`
	Summary    string   `json:"summary"`
}

type AuthPoliciesMetadata struct {
	Command        string  `json:"command"`
	GeneratedAt    string  `json:"generated_at"`
	SchemaVersion  string  `json:"schema_version"`
	SubscriptionID *string `json:"subscription_id"`
	TenantID       *string `json:"tenant_id"`
	TokenSource    *string `json:"token_source"`
}

type AuthPolicyFinding struct {
	Description string   `json:"description"`
	ID          string   `json:"id"`
	RelatedIDs  []string `json:"related_ids"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
}

type AuthPoliciesOutput struct {
	AuthPolicies []AuthPolicySummary  `json:"auth_policies"`
	Findings     []AuthPolicyFinding  `json:"findings"`
	Issues       []Issue              `json:"issues"`
	Metadata     AuthPoliciesMetadata `json:"metadata"`
}
