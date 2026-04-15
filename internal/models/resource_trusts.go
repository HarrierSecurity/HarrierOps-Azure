package models

type ResourceTrustSummary struct {
	Confidence   string   `json:"confidence"`
	Exposure     string   `json:"exposure"`
	RelatedIDs   []string `json:"related_ids"`
	ResourceID   string   `json:"resource_id"`
	ResourceName string   `json:"resource_name"`
	ResourceType string   `json:"resource_type"`
	Summary      string   `json:"summary"`
	Target       string   `json:"target"`
	TrustType    string   `json:"trust_type"`
}

type ResourceTrustFinding struct {
	Description string   `json:"description"`
	ID          string   `json:"id"`
	RelatedIDs  []string `json:"related_ids"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
}

type ResourceTrustsMetadata = Metadata

type ResourceTrustsOutput struct {
	Findings       []ResourceTrustFinding `json:"findings"`
	Issues         []Issue                `json:"issues"`
	Metadata       ResourceTrustsMetadata `json:"metadata"`
	ResourceTrusts []ResourceTrustSummary `json:"resource_trusts"`
}
