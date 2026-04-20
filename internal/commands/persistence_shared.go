package commands

import (
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

type persistenceCurrentIdentityControl struct {
	RoleName string
	ScopeID  string
}

func persistenceAutomationControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	bestRank := 99
	best := persistenceCurrentIdentityControl{}
	for _, assignment := range assignments {
		role := strings.ToLower(strings.TrimSpace(assignment.RoleName))
		if role != "owner" && role != "contributor" && role != "automation contributor" {
			continue
		}
		rank, ok := persistenceScopeRank(assignment.ScopeID, resourceID)
		if !ok || rank >= bestRank {
			continue
		}
		bestRank = rank
		best = persistenceCurrentIdentityControl{
			RoleName: fmt.Sprintf("%s at %s", assignment.RoleName, persistenceScopeLabel(assignment.ScopeID)),
			ScopeID:  assignment.ScopeID,
		}
	}
	return best, bestRank != 99
}

func persistenceScopeRank(scopeID string, resourceID string) (int, bool) {
	scopeID = strings.TrimSpace(scopeID)
	resourceID = strings.TrimSpace(resourceID)
	if scopeID == "" || resourceID == "" {
		return 0, false
	}
	scopeLower := strings.ToLower(strings.TrimRight(scopeID, "/"))
	resourceLower := strings.ToLower(strings.TrimRight(resourceID, "/"))
	if scopeLower == resourceLower {
		return 0, true
	}
	if !strings.HasPrefix(resourceLower, scopeLower+"/") {
		return 0, false
	}
	if strings.Contains(scopeLower, "/resourcegroups/") {
		return 1, true
	}
	if strings.Contains(scopeLower, "/subscriptions/") {
		return 2, true
	}
	return 3, true
}

func persistenceScopeLabel(scopeID string) string {
	if strings.Contains(scopeID, "/subscriptions/") && !strings.Contains(scopeID, "/resourceGroups/") {
		return "subscription scope"
	}
	if strings.Contains(scopeID, "/resourceGroups/") {
		return "resource group " + armScopeName(scopeID)
	}
	return "a parent scope of this resource"
}
