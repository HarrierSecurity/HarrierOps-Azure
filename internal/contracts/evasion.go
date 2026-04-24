package contracts

import "sort"

type SurfaceContract struct {
	GroupCommand     string
	Name             string
	Status           string
	Summary          string
	OperatorQuestion string
	BackingCommands  []string
}

type EvasionSurfaceContract = SurfaceContract

var evasionSurfaceContracts = map[string]EvasionSurfaceContract{
	"dcr": {
		GroupCommand:     "evasion",
		Name:             "dcr",
		Status:           StatusImplemented,
		Summary:          "Review Data Collection Rules for collection, stream, destination, association, and transformation posture that can quietly reshape monitoring truth.",
		OperatorQuestion: "How far can current access take me through DCR collection, routing, association, and transformation levers before the proof boundary moves into runtime logs or agent state?",
		BackingCommands:  []string{"dcr", "permissions", "rbac"},
	},
	"diagnostic-settings": {
		GroupCommand:     "evasion",
		Name:             "diagnostic-settings",
		Status:           StatusImplemented,
		Summary:          "Review diagnostic settings for source resources, exported categories, metrics, destinations, and visible telemetry-routing posture.",
		OperatorQuestion: "How far can current access take me through diagnostic setting category and destination levers before the proof boundary moves into sink contents, history, or detector wiring?",
		BackingCommands:  []string{"diagnostic-settings", "permissions", "rbac"},
	},
	"appinsights": {
		GroupCommand:     "evasion",
		Name:             "appinsights",
		Status:           StatusImplemented,
		Summary:          "Review Application Insights components and instrumented app settings for visible sampling, filtering, and logging posture clues.",
		OperatorQuestion: "How far can current access take me through visible Application Insights instrumentation, sampling, filtering, and logging-level levers before the proof boundary moves into code, runtime, or telemetry content?",
		BackingCommands:  []string{"appinsights", "permissions", "rbac"},
	},
}

func EvasionSurface(name string) (EvasionSurfaceContract, bool) {
	contract, ok := evasionSurfaceContracts[name]
	return contract, ok
}

func EvasionSurfaceNames() []string {
	names := make([]string, 0, len(evasionSurfaceContracts))
	for name := range evasionSurfaceContracts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
