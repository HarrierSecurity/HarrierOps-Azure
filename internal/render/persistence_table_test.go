package render

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestPersistenceAutomationTableUsesSingleWalkthroughForMultipleAccounts(t *testing.T) {
	output := persistenceAutomationTable(models.PersistenceAutomationOutput{
		AutomationAccounts: []models.PersistenceAutomationAccount{
			{
				Name: "aa-one",
				CapabilitySteps: []models.PersistenceCapabilityStep{
					{Action: "create or modify account", Status: "yes"},
					{Action: "add or edit runbook", Status: "yes"},
					{Action: "upload or replace code", Status: "yes"},
					{Action: "publish runbook", Status: "yes"},
					{Action: "attach or reuse exec ctx", Status: "yes"},
					{Action: "create schedule", Status: "yes"},
					{Action: "link schedule to runbook", Status: "yes"},
					{Action: "create webhook", Status: "yes"},
				},
				CurrentState: models.PersistenceAutomationState{},
				Summary:      "summary one",
			},
			{
				Name: "aa-two",
				CapabilitySteps: []models.PersistenceCapabilityStep{
					{Action: "create or modify account", Status: "yes"},
					{Action: "add or edit runbook", Status: "yes"},
					{Action: "upload or replace code", Status: "yes"},
					{Action: "publish runbook", Status: "yes"},
					{Action: "attach or reuse exec ctx", Status: "yes"},
					{Action: "create schedule", Status: "yes"},
					{Action: "link schedule to runbook", Status: "yes"},
					{Action: "create webhook", Status: "yes"},
				},
				CurrentState: models.PersistenceAutomationState{},
				Summary:      "summary two",
			},
		},
	})

	if strings.Count(output, "Automation capability") != 1 {
		t.Fatalf("expected one shared Automation walkthrough, got:\n%s", output)
	}
	if strings.Count(output, "Reminder: a runbook does not run continuously") != 1 {
		t.Fatalf("expected one shared Automation reminder, got:\n%s", output)
	}
	if !strings.Contains(output, "Visible Automation Accounts") {
		t.Fatalf("expected compact Automation inventory section, got:\n%s", output)
	}
	if !strings.Contains(output, "aa-one") || !strings.Contains(output, "aa-two") {
		t.Fatalf("expected both Automation accounts in compact inventory, got:\n%s", output)
	}
}

func TestPersistenceLogicAppsTableUsesSingleWalkthroughForMultipleWorkflows(t *testing.T) {
	output := persistenceLogicAppsTable(models.PersistenceLogicAppsOutput{
		Workflows: []models.PersistenceLogicAppWorkflow{
			{
				Name: "la-one",
				CapabilitySteps: []models.PersistenceCapabilityStep{
					{Action: "create or modify workflow", Status: "yes"},
					{Action: "edit workflow definition", Status: "yes"},
					{Action: "attach or reuse exec ctx", Status: "yes"},
					{Action: "define or modify trigger", Status: "yes"},
					{Action: "enable workflow", Status: "yes"},
					{Action: "add or repurpose downstream actions", Status: "yes"},
				},
				CurrentState: models.PersistenceLogicAppState{
					Classification:                   "persistence-capable",
					ExternallyCallableRequestTrigger: true,
					NearbyThematicNames:              []string{"nightly-sync", "maintenance-router"},
				},
				Summary: "logic summary one",
			},
			{
				Name: "la-two",
				CapabilitySteps: []models.PersistenceCapabilityStep{
					{Action: "create or modify workflow", Status: "yes"},
					{Action: "edit workflow definition", Status: "yes"},
					{Action: "attach or reuse exec ctx", Status: "yes"},
					{Action: "define or modify trigger", Status: "yes"},
					{Action: "enable workflow", Status: "yes"},
					{Action: "add or repurpose downstream actions", Status: "yes"},
				},
				CurrentState: models.PersistenceLogicAppState{
					Classification: "persistence-capable",
				},
				Summary: "logic summary two",
			},
		},
	})

	if strings.Count(output, "Workflow capability") != 1 {
		t.Fatalf("expected one shared Logic Apps walkthrough, got:\n%s", output)
	}
	if strings.Count(output, "Reminder: Logic App persistence is about a stored workflow plus a trigger plus access that remains valid") != 1 {
		t.Fatalf("expected one shared Logic Apps reminder, got:\n%s", output)
	}
	if !strings.Contains(output, "Visible Logic Apps") {
		t.Fatalf("expected compact Logic Apps inventory section, got:\n%s", output)
	}
	if !strings.Contains(output, "la-one") || !strings.Contains(output, "la-two") {
		t.Fatalf("expected both Logic Apps in compact inventory, got:\n%s", output)
	}
	if !strings.Contains(output, "Nearby maintenance- or schedule-themed names visible from the current environment include `nightly-sync` and `maintenance-router`.") {
		t.Fatalf("expected nearby thematic Logic App names line, got:\n%s", output)
	}
}
