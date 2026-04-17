package commands

import (
	"context"
	"testing"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
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

func TestBuildPersistenceOverviewIncludesLogicAppsSurface(t *testing.T) {
	output := buildPersistenceOverview(func() time.Time { return time.Unix(0, 0) }, Request{}, nil)

	for _, surface := range output.Surfaces {
		if surface.Surface == "logic-apps" {
			return
		}
	}

	t.Fatalf("expected persistence overview to include logic-apps surface, got %#v", output.Surfaces)
}

func TestBuildPersistenceLogicAppsOutputResolvesVisibleExecutionContextRoleContext(t *testing.T) {
	contract, ok := contracts.PersistenceSurface("logic-apps")
	if !ok {
		t.Fatal("expected logic-apps persistence surface contract")
	}

	outputAny, err := buildPersistenceLogicAppsOutput(
		context.Background(),
		providers.NewStaticProvider(),
		func() time.Time { return time.Unix(0, 0) },
		Request{},
		contract,
	)
	if err != nil {
		t.Fatalf("buildPersistenceLogicAppsOutput returned error: %v", err)
	}

	output, ok := outputAny.(models.PersistenceLogicAppsOutput)
	if !ok {
		t.Fatalf("unexpected output type %T", outputAny)
	}

	for _, workflow := range output.Workflows {
		if workflow.Name != "la-inbound-redeploy" {
			continue
		}
		if workflow.CurrentState.StrongestVisibleExecutionContext == nil {
			t.Fatalf("expected visible execution context for %q", workflow.Name)
		}
		if got := workflow.CurrentState.StrongestVisibleExecutionContext.RoleNames; len(got) == 0 {
			t.Fatalf("expected role context for %q, got none", workflow.Name)
		}
		return
	}

	t.Fatalf("expected la-inbound-redeploy workflow in output")
}

func TestPersistenceLogicAppSummaryDoesNotOverclaimDisabledWorkflowAsDurable(t *testing.T) {
	workflow := models.LogicAppWorkflowAsset{
		Name:                             "la-disabled",
		Classification:                   "persistence-capable",
		ExternallyCallableRequestTrigger: true,
		State:                            models.StringPtr("Disabled"),
	}

	got := persistenceLogicAppSummary(workflow, true, nil, false)
	want := "Current identity can build or repurpose Logic App 'la-disabled', but the current workflow definition does not yet show a durable request or recurrence trigger."
	if got != want {
		t.Fatalf("unexpected summary %q, want %q", got, want)
	}
}
