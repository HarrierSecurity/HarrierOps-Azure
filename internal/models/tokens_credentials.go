package models

type TokenCredentialSurfaceSummary struct {
	AccessPath     string                     `json:"access_path"`
	AssetID        string                     `json:"asset_id"`
	AssetKind      string                     `json:"kind"`
	AssetName      string                     `json:"asset"`
	Location       *string                    `json:"location"`
	OperatorSignal string                     `json:"operator_signal"`
	Priority       string                     `json:"priority"`
	RelatedIDs     []string                   `json:"related_ids"`
	ResourceGroup  *string                    `json:"resource_group"`
	Summary        string                     `json:"summary"`
	SurfaceType    TokenCredentialSurfaceType `json:"surface"`

	NextReviewKind    TokenCredentialNextReviewKind `json:"-"`
	PubliclyReachable bool                          `json:"-"`
}

type TokenCredentialFinding struct {
	Description string   `json:"description"`
	ID          string   `json:"id"`
	RelatedIDs  []string `json:"related_ids"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
}

type TokensCredentialsOutput struct {
	Findings []TokenCredentialFinding        `json:"findings"`
	Issues   []Issue                         `json:"issues"`
	Metadata ScopedCommandMetadata           `json:"metadata"`
	Surfaces []TokenCredentialSurfaceSummary `json:"surfaces"`
}
