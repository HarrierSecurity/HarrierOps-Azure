package contracts

import "sort"

type ResourceHijackingSurfaceContract = SurfaceContract

var resourceHijackingSurfaceContracts = map[string]ResourceHijackingSurfaceContract{
	"api-mgmt": {
		GroupCommand:     "resourcehijacking",
		Name:             "api-mgmt",
		Status:           StatusImplemented,
		Summary:          "Review API Management services for gateway, backend, subscription, named-value, and identity posture that can redirect a trusted API surface.",
		OperatorQuestion: "How far can current access take me through APIM backend and routing-control levers before the proof boundary moves into policy bodies, traffic logs, or backend ownership?",
		BackingCommands:  []string{"api-mgmt", "permissions", "rbac"},
	},
	"automation": {
		GroupCommand:     "resourcehijacking",
		Name:             "automation",
		Status:           StatusImplemented,
		Summary:          "Review Azure Automation accounts for published runbook, schedule, webhook, hybrid worker, secure asset, and identity posture that can repurpose trusted operations automation.",
		OperatorQuestion: "How far can current access take me through Automation runbook, trigger, execution context, and account-control levers before the proof boundary moves into script content, job output, or runtime host state?",
		BackingCommands:  []string{"automation", "permissions", "rbac"},
	},
	"logic-apps": {
		GroupCommand:     "resourcehijacking",
		Name:             "logic-apps",
		Status:           StatusImplemented,
		Summary:          "Review Logic Apps for workflow definition, trigger, downstream action, connector, and identity posture that can repurpose trusted automation.",
		OperatorQuestion: "How far can current access take me through Logic App trigger, workflow, action, and identity levers before the proof boundary moves into run history, connector data, or secret material?",
		BackingCommands:  []string{"logic-apps", "permissions", "rbac"},
	},
}

func ResourceHijackingSurface(name string) (ResourceHijackingSurfaceContract, bool) {
	contract, ok := resourceHijackingSurfaceContracts[name]
	return contract, ok
}

func ResourceHijackingSurfaceNames() []string {
	names := make([]string, 0, len(resourceHijackingSurfaceContracts))
	for name := range resourceHijackingSurfaceContracts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
