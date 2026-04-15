package families

import "harrierops-azure/internal/contracts"

func Names() []string {
	return contracts.FamilyNames()
}

func Contract(name string) (contracts.FamilyContract, bool) {
	return contracts.Family(name)
}
