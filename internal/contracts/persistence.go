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
		Summary:          "Walk the current identity through the end-to-end Azure Automation steps that can store runnable code and make Azure invoke it again later.",
		OperatorQuestion: "Given the current identity, how far can I go in Azure Automation to create, modify, publish, and trigger automation that Azure will run again later?",
		BackingCommands:  []string{"automation", "permissions", "rbac"},
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
