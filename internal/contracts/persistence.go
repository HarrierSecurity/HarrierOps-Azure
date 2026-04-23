package contracts

import "sort"

type PersistenceSurfaceContract struct {
	GroupCommand     string
	Name             string
	Status           string
	Summary          string
	OperatorQuestion string
	BackingCommands  []string
}

var persistenceSurfaceContracts = map[string]PersistenceSurfaceContract{
	"automation": {
		GroupCommand:     "persistence",
		Name:             "automation",
		Status:           StatusImplemented,
		Summary:          "Review Azure Automation for reusable code, execution context, and trigger paths.",
		OperatorQuestion: "How far can current access take me in Azure Automation before I hit a permission boundary?",
		BackingCommands:  []string{"automation", "permissions", "rbac"},
	},
	"app-service": {
		GroupCommand:     "persistence",
		Name:             "app-service",
		Status:           StatusImplemented,
		Summary:          "Review App Service for deployment path, configuration power, code replacement, and reachable re-entry posture.",
		OperatorQuestion: "How far can current access take me in App Service before I hit the permission boundary between app control, deployment/config control, and later HTTP-backed reuse?",
		BackingCommands:  []string{"app-services", "env-vars", "managed-identities", "permissions", "rbac"},
	},
	"azure-ml": {
		GroupCommand:     "persistence",
		Name:             "azure-ml",
		Status:           StatusImplemented,
		Summary:          "Review Azure ML for reusable compute, job, schedule, endpoint, and identity paths.",
		OperatorQuestion: "How far can current access take me in Azure ML before I hit a permission boundary?",
		BackingCommands:  []string{"azure-ml", "managed-identities", "permissions", "rbac"},
	},
	"functions": {
		GroupCommand:     "persistence",
		Name:             "functions",
		Status:           StatusImplemented,
		Summary:          "Review Function Apps for reusable code, identity, and trigger paths.",
		OperatorQuestion: "How far can current access take me in Azure Functions before I hit a permission boundary?",
		BackingCommands:  []string{"functions", "permissions", "rbac"},
	},
	"logic-apps": {
		GroupCommand:     "persistence",
		Name:             "logic-apps",
		Status:           StatusImplemented,
		Summary:          "Review Logic Apps for reusable workflow and trigger paths.",
		OperatorQuestion: "How far can current access take me in Azure Logic Apps before I hit a permission boundary?",
		BackingCommands:  []string{"logic-apps", "permissions", "rbac"},
	},
	"webjobs": {
		GroupCommand:     "persistence",
		Name:             "webjobs",
		Status:           StatusImplemented,
		Summary:          "Review App Service WebJobs for reusable background code, mode, inherited app context, and rerun paths.",
		OperatorQuestion: "How far can current access take me in App Service WebJobs before I hit the permission boundary between parent app control, WebJob content, rerun mode, and inherited execution context?",
		BackingCommands:  []string{"webjobs", "app-services", "managed-identities", "permissions", "rbac"},
	},
}

func PersistenceSurface(name string) (PersistenceSurfaceContract, bool) {
	contract, ok := persistenceSurfaceContracts[name]
	return contract, ok
}

func PersistenceSurfaceNames() []string {
	names := make([]string, 0, len(persistenceSurfaceContracts))
	for name := range persistenceSurfaceContracts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
