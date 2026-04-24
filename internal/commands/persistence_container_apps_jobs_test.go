package commands

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestPersistenceContainerAppsJobControlUsesAttachedRoleActions(t *testing.T) {
	resourceID := "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/jobs/nightly-reconcile"
	control, ok := persistenceContainerAppsJobControl(resourceID, []models.RoleAssignment{{
		RoleName: "Container Apps Job Operator",
		ScopeID:  "/subscriptions/sub/resourceGroups/rg",
		Actions:  []string{"Microsoft.App/jobs/*"},
	}})

	if !ok {
		t.Fatal("expected attached role action list to prove Container Apps job control")
	}
	if !strings.Contains(control.RoleName, "Container Apps Job Operator") {
		t.Fatalf("expected control role label to preserve custom role name, got %q", control.RoleName)
	}
}

func TestPersistenceContainerAppsJobControlDoesNotInferUncollectedCustomRoleActions(t *testing.T) {
	resourceID := "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/jobs/nightly-reconcile"
	_, ok := persistenceContainerAppsJobControl(resourceID, []models.RoleAssignment{{
		RoleName: "Container Apps Job Operator",
		ScopeID:  "/subscriptions/sub/resourceGroups/rg",
	}})

	if ok {
		t.Fatal("expected custom role without collected action list to stay unproven")
	}
}

func TestPersistenceRoleAssignmentAllowsManagementActionHonorsNotActions(t *testing.T) {
	assignment := models.RoleAssignment{
		Actions:    []string{"Microsoft.App/jobs/*"},
		NotActions: []string{"Microsoft.App/jobs/write"},
	}

	if persistenceRoleAssignmentAllowsManagementAction(assignment, "Microsoft.App/jobs/write") {
		t.Fatal("expected notActions to subtract the matching management action")
	}
}

func TestPersistenceContainerAppsJobControlHonorsNotActionsOnNamedRoles(t *testing.T) {
	resourceID := "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/jobs/nightly-reconcile"
	_, ok := persistenceContainerAppsJobControl(resourceID, []models.RoleAssignment{{
		RoleName:   "Contributor",
		ScopeID:    "/subscriptions/sub/resourceGroups/rg",
		NotActions: []string{"Microsoft.App/jobs/write"},
	}})

	if ok {
		t.Fatal("expected notActions to subtract Container Apps job write even when role name is Contributor")
	}
}
