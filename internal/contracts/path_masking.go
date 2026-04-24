package contracts

import "sort"

type PathMaskingSurfaceContract = SurfaceContract

var pathMaskingSurfaceContracts = map[string]PathMaskingSurfaceContract{
	"api-mgmt": {
		GroupCommand:     "pathmasking",
		Name:             "api-mgmt",
		Status:           StatusImplemented,
		Summary:          "Review API Management services for gateway, backend, hostname, subscription, and named-value posture that can mask the true public-to-backend path.",
		OperatorQuestion: "How far can current access take me through APIM gateway, route, transform, and backend indirection before the proof boundary moves into policy bodies, live traffic, or backend ownership?",
		BackingCommands:  []string{"api-mgmt", "permissions", "rbac"},
	},
	"logic-apps": {
		GroupCommand:     "pathmasking",
		Name:             "logic-apps",
		Status:           StatusImplemented,
		Summary:          "Review Logic Apps for request, schedule, connector, HTTP action, and identity posture that can relay activity through trusted integration workflows.",
		OperatorQuestion: "Which visible workflows can current access reuse or modify as a trusted relay path before the proof boundary moves into run history, connector payloads, or credential material?",
		BackingCommands:  []string{"logic-apps", "permissions", "rbac"},
	},
	"relay": {
		GroupCommand:     "pathmasking",
		Name:             "relay",
		Status:           StatusImplemented,
		Summary:          "Review Azure Relay namespaces and Hybrid Connections for cloud rendezvous points that can blur the path to private listeners or internal services.",
		OperatorQuestion: "Which Relay namespaces and Hybrid Connections give current access a visible private-path rendezvous before the proof boundary moves into listener runtime, backend process, or traffic contents?",
		BackingCommands:  []string{"relay", "permissions", "rbac"},
	},
}

func PathMaskingSurface(name string) (PathMaskingSurfaceContract, bool) {
	contract, ok := pathMaskingSurfaceContracts[name]
	return contract, ok
}

func PathMaskingSurfaceNames() []string {
	names := make([]string, 0, len(pathMaskingSurfaceContracts))
	for name := range pathMaskingSurfaceContracts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
