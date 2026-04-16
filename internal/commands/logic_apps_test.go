package commands

import (
	"testing"

	"harrierops-azure/internal/models"
)

func TestLogicAppArtifactTriggerDedupesStructuredTriggerLabels(t *testing.T) {
	trigger := logicAppArtifactTrigger(models.LogicAppWorkflowAsset{
		ExternallyCallableRequestTrigger: true,
		RecurrenceSummary:                models.StringPtr("Day/1"),
		TriggerTypes:                     []string{"request", "recurrence", "api-connection"},
	})

	if trigger != "request(external); recurrence; api-connection" {
		t.Fatalf("unexpected trigger summary %q", trigger)
	}
}

func TestPersistenceAutomationSummaryUsesStructuredPrivilegeSignal(t *testing.T) {
	summary := persistenceAutomationSummary(
		models.AutomationAccountAsset{Name: "auto-prod"},
		true,
		&models.PersistenceRoleContext{
			Name:      "auto-prod-identity",
			Kind:      "automation-execution-context",
			RoleNames: []string{"Reader"},
			Summary:   "Automation identity is visible.",
		},
		false,
	)

	want := "Current identity can manage Automation Account 'auto-prod' end to end from current RBAC evidence."
	if summary != want {
		t.Fatalf("unexpected persistence summary %q, want %q", summary, want)
	}
}
