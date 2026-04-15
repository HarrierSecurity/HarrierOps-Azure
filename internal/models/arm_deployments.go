package models

type ArmDeploymentSummary struct {
	Duration            string   `json:"duration"`
	ID                  string   `json:"id"`
	Mode                string   `json:"mode"`
	Name                string   `json:"name"`
	OutputResourceCount int      `json:"output_resource_count"`
	OutputsCount        int      `json:"outputs_count"`
	ParametersLink      *string  `json:"parameters_link"`
	Providers           []string `json:"providers"`
	ProvisioningState   string   `json:"provisioning_state"`
	RelatedIDs          []string `json:"related_ids"`
	ResourceGroup       *string  `json:"resource_group"`
	Scope               string   `json:"scope"`
	ScopeType           string   `json:"scope_type"`
	Summary             string   `json:"summary"`
	TemplateLink        *string  `json:"template_link"`
	Timestamp           string   `json:"timestamp"`
}

type ArmDeploymentFinding struct {
	ID          string   `json:"id"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	RelatedIDs  []string `json:"related_ids"`
}

type ArmDeploymentsOutput struct {
	Deployments []ArmDeploymentSummary `json:"deployments"`
	Findings    []ArmDeploymentFinding `json:"findings"`
	Issues      []Issue                `json:"issues"`
	Metadata    Metadata               `json:"metadata"`
}
