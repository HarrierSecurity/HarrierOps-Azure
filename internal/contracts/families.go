package contracts

import "sort"

type FamilySourceContract struct {
	Command       string
	MinimumFields []string
	Rationale     string
}

type FamilyContract struct {
	GroupedCommand       string
	Name                 string
	Status               string
	Summary              string
	AllowedClaim         string
	CurrentGap           string
	BackingCommands      []string
	PreferredArtifacts   []string
	StructuredRowFields  []string
	OperatorQuestion     string
	SourceCommandMinimum []FamilySourceContract
}

var familyContracts = map[string]FamilyContract{
	"credential-path": {
		GroupedCommand:     "chains",
		Name:               "credential-path",
		Status:             StatusPlaceholder,
		Summary:            "Follow credential clues from surfaced secret-bearing or token-bearing evidence toward the likely downstream service.",
		AllowedClaim:       "Can claim that visible evidence suggests a likely credential path. Cannot claim confirmed downstream use or exploitation proof from recon-only signals.",
		CurrentGap:         "Family contract location exists, but the Go rewrite has not ported the real evidence joins yet.",
		BackingCommands:    []string{"env-vars", "tokens-credentials", "databases", "storage", "keyvault"},
		PreferredArtifacts: []string{"loot", "json"},
		StructuredRowFields: []string{
			"start_state",
			"source_command",
			"target_service",
			"confidence_boundary",
			"missing_proof",
			"next_review",
			"summary",
		},
		OperatorQuestion: "Which visible workload or config clue most plausibly leads to a downstream credential-bearing target?",
		SourceCommandMinimum: []FamilySourceContract{
			{Command: "env-vars", MinimumFields: []string{"asset_id", "setting_name", "value_type", "reference_target"}, Rationale: "Credential-shaped workload clue."},
			{Command: "tokens-credentials", MinimumFields: []string{"asset_id", "surface_type", "priority", "operator_signal"}, Rationale: "Direct token or credential surface proof."},
		},
	},
	"deployment-path": {
		GroupedCommand:     "chains",
		Name:               "deployment-path",
		Status:             StatusPlaceholder,
		Summary:            "Follow controllable deployment and automation paths toward the Azure footprint they are most likely to change next.",
		AllowedClaim:       "Can claim a visible Azure change path when the evidence supports it. Cannot claim successful execution or a proven exploit chain from recon-only signals.",
		CurrentGap:         "Real source-side actionability joins and downstream targeting have not been ported yet.",
		BackingCommands:    []string{"devops", "automation", "permissions", "rbac", "role-trusts"},
		PreferredArtifacts: []string{"loot", "json"},
		StructuredRowFields: []string{
			"start_state",
			"control_surface",
			"downstream_target",
			"confidence_boundary",
			"missing_proof",
			"next_review",
			"summary",
		},
		OperatorQuestion: "Which visible automation or supply-chain path can most honestly be described as a controllable Azure change path?",
	},
	"escalation-path": {
		GroupedCommand:     "chains",
		Name:               "escalation-path",
		Status:             StatusPlaceholder,
		Summary:            "Group privilege-escalation style paths with explicit proof boundaries and operator-ready next review steps.",
		AllowedClaim:       "Can claim the strongest defended escalation lead. Cannot overstate support clues as a proven admitted path.",
		CurrentGap:         "Family contract exists, but the Go port has not implemented the real row admission logic yet.",
		BackingCommands:    []string{"permissions", "privesc", "role-trusts"},
		PreferredArtifacts: []string{"loot", "json"},
		StructuredRowFields: []string{
			"starting_foothold",
			"target_identity",
			"path_type",
			"confidence_boundary",
			"missing_proof",
			"next_review",
			"summary",
		},
		OperatorQuestion: "What is the best defended escalation lead available from the current foothold?",
	},
	"compute-control": {
		GroupedCommand:     "chains",
		Name:               "compute-control",
		Status:             StatusPlaceholder,
		Summary:            "Group workload and deployment evidence into bounded control-path rows for Azure compute surfaces.",
		AllowedClaim:       "Can name defended compute-control leads. Cannot claim execution success beyond what the visible evidence proves.",
		CurrentGap:         "Real joins across workloads, identities, and permissions are still deferred.",
		BackingCommands:    []string{"workloads", "permissions", "managed-identities", "functions", "app-services", "aks"},
		PreferredArtifacts: []string{"loot", "json"},
		StructuredRowFields: []string{
			"start_state",
			"asset",
			"control_surface",
			"confidence_boundary",
			"missing_proof",
			"next_review",
			"summary",
		},
		OperatorQuestion: "Which compute surface offers the strongest current control lead without overstating what the evidence proves?",
	},
}

func Family(name string) (FamilyContract, bool) {
	contract, ok := familyContracts[name]
	return contract, ok
}

func FamilyNames() []string {
	names := make([]string, 0, len(familyContracts))
	for name := range familyContracts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
