package models

type CrossTenantPathSummary struct {
	AttackPath string   `json:"attack_path"`
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Posture    *string  `json:"posture"`
	Priority   string   `json:"priority"`
	RelatedIDs []string `json:"related_ids,omitempty"`
	Scope      *string  `json:"scope"`
	SignalType string   `json:"signal_type"`
	Summary    string   `json:"summary"`
	TenantID   *string  `json:"tenant_id"`
	TenantName *string  `json:"tenant_name"`
}

type CrossTenantMetadata = Metadata

type CrossTenantOutput struct {
	CrossTenantPaths []CrossTenantPathSummary `json:"cross_tenant_paths"`
	Findings         []Finding                `json:"findings"`
	Issues           []Issue                  `json:"issues"`
	Metadata         CrossTenantMetadata      `json:"metadata"`
}
