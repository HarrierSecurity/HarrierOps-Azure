package commands

import (
	"context"
	"strings"
	"testing"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func singleRoleAssignment(principalID, roleName, scopeID string) map[string][]models.RoleAssignment {
	return map[string][]models.RoleAssignment{
		principalID: {{
			PrincipalID: principalID,
			RoleName:    roleName,
			ScopeID:     scopeID,
		}},
	}
}

func assertRBACOnlyFallbackHedged(t *testing.T, context *models.PersistenceRoleContext, carriesAzureControl bool) {
	t.Helper()
	if context == nil {
		t.Fatal("expected execution context from RBAC-only fallback")
	}
	if carriesAzureControl {
		t.Fatal("expected RBAC-only fallback to avoid claiming resolved Azure control")
	}
	if !strings.Contains(context.Summary, "raw Azure role-assignment rows for its principal ID suggest stronger Azure control") {
		t.Fatalf("expected hedged RBAC-only fallback summary, got %q", context.Summary)
	}
}

func assertRBACOnlyFallbackCarriesControl(t *testing.T, context *models.PersistenceRoleContext, carriesAzureControl bool, summaryNeedle string) {
	t.Helper()
	if context == nil {
		t.Fatal("expected execution context from RBAC-only fallback")
	}
	if !carriesAzureControl {
		t.Fatal("expected RBAC-only fallback to carry Azure-control signal")
	}
	if !strings.Contains(context.Summary, summaryNeedle) {
		t.Fatalf("expected RBAC-only fallback summary to keep visible assignment strength, got %q", context.Summary)
	}
}

func TestPersistenceAppServiceExecutionContextDoesNotOverclaimRBACOnlyFallback(t *testing.T) {
	app := models.AppServiceAsset{
		Name:                 "app-orders",
		WorkloadIdentityType: models.StringPtr("SystemAssigned"),
		WorkloadPrincipalID:  models.StringPtr("principal-app"),
	}
	context, carriesAzureControl := persistenceAppServiceExecutionContext(app, nil, map[string]models.PermissionRow{}, singleRoleAssignment("principal-app", "Owner", "/subscriptions/test"))
	assertRBACOnlyFallbackHedged(t, context, carriesAzureControl)
}

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

func TestBuildPersistencePrincipalEvidenceIndexesCurrentIdentityAndAssignments(t *testing.T) {
	permissions := []models.PermissionRow{
		{PrincipalID: "current-principal", DisplayName: "azurefox-lab-sp", IsCurrentIdentity: true},
		{PrincipalID: "other-principal", DisplayName: "other"},
	}
	assignments := []models.RoleAssignment{
		{PrincipalID: "current-principal", RoleName: "Owner", ScopeID: "/subscriptions/test"},
		{PrincipalID: "other-principal", RoleName: "Reader", ScopeID: "/subscriptions/test/resourceGroups/rg-apps"},
	}

	evidence := buildPersistencePrincipalEvidence(permissions, assignments)

	if !evidence.currentIdentityVisible {
		t.Fatal("expected current identity to be indexed")
	}
	if got := evidence.currentIdentity.DisplayName; got != "azurefox-lab-sp" {
		t.Fatalf("expected current identity display name, got %q", got)
	}
	if len(evidence.currentIdentityAssignments) != 1 {
		t.Fatalf("expected one current-identity assignment, got %d", len(evidence.currentIdentityAssignments))
	}
	if got := evidence.permissionsByPrincipal["other-principal"].DisplayName; got != "other" {
		t.Fatalf("expected permission lookup for other principal, got %q", got)
	}
	if len(evidence.assignmentsByPrincipal["other-principal"]) != 1 {
		t.Fatalf("expected assignment lookup for other principal, got %d rows", len(evidence.assignmentsByPrincipal["other-principal"]))
	}
}

func TestPersistenceAutomationExecutionContextCarriesRBACOnlyFallbackWhenAssignmentsVisible(t *testing.T) {
	account := models.AutomationAccountAsset{
		Name:         "aa-prod",
		IdentityType: models.StringPtr("SystemAssigned"),
		PrincipalID:  models.StringPtr("principal-auto"),
	}
	context, carriesAzureControl := persistenceAutomationExecutionContext(account, map[string]models.PermissionRow{}, singleRoleAssignment("principal-auto", "Owner", "/subscriptions/test"))
	assertRBACOnlyFallbackCarriesControl(t, context, carriesAzureControl, "already holds Owner at subscription scope")
}

func TestBuildPersistenceOverviewIncludesLogicAppsSurface(t *testing.T) {
	output := buildPersistenceOverview(func() time.Time { return time.Unix(0, 0) }, Request{}, nil)

	sawLogicApps := false
	sawFunctions := false
	sawAzureML := false
	sawWebJobs := false
	for _, surface := range output.Surfaces {
		switch surface.Surface {
		case "azure-ml":
			sawAzureML = true
		case "logic-apps":
			sawLogicApps = true
		case "functions":
			sawFunctions = true
		case "webjobs":
			sawWebJobs = true
		}
	}

	if !sawLogicApps {
		t.Fatalf("expected persistence overview to include logic-apps surface, got %#v", output.Surfaces)
	}
	if !sawFunctions {
		t.Fatalf("expected persistence overview to include functions surface, got %#v", output.Surfaces)
	}
	if !sawAzureML {
		t.Fatalf("expected persistence overview to include azure-ml surface, got %#v", output.Surfaces)
	}
	if !sawWebJobs {
		t.Fatalf("expected persistence overview to include webjobs surface, got %#v", output.Surfaces)
	}
}

func TestBuildPersistenceAzureMLOutputResolvesVisibleExecutionContextRoleContext(t *testing.T) {
	contract, ok := contracts.PersistenceSurface("azure-ml")
	if !ok {
		t.Fatal("expected azure-ml persistence surface contract")
	}

	outputAny, err := buildPersistenceAzureMLOutput(
		context.Background(),
		providers.NewStaticProvider(),
		func() time.Time { return time.Unix(0, 0) },
		Request{},
		contract,
	)
	if err != nil {
		t.Fatalf("buildPersistenceAzureMLOutput returned error: %v", err)
	}

	output, ok := outputAny.(models.PersistenceAzureMLOutput)
	if !ok {
		t.Fatalf("unexpected output type %T", outputAny)
	}

	for _, workspace := range output.Workspaces {
		if workspace.Name != "ml-ops-hub" {
			continue
		}
		if workspace.CurrentState.StrongestVisibleExecutionContext == nil {
			t.Fatalf("expected visible execution context for %q", workspace.Name)
		}
		if got := workspace.CurrentState.StrongestVisibleExecutionContext.Name; got != "ua-ml-ops" {
			t.Fatalf("expected workspace-specific execution context label, got %q", got)
		}
		foundGap := false
		for _, item := range workspace.StillUnmapped {
			if strings.Contains(item, "current output does not yet resolve their backing principals") {
				foundGap = true
				break
			}
		}
		if foundGap {
			t.Fatalf("expected user-assigned identity ranking gap to be resolved for %q", workspace.Name)
		}
		if len(workspace.CurrentState.VisibleIdentityNames) == 0 {
			t.Fatalf("expected visible Azure ML identity names for %q", workspace.Name)
		}
		return
	}

	t.Fatalf("expected ml-ops-hub workspace in output")
}

func TestManagedIdentitiesOutputIncludesAzureMLAttachments(t *testing.T) {
	outputAny, err := managedIdentitiesHandler(providers.NewStaticProvider(), func() time.Time { return time.Unix(0, 0) })(
		context.Background(),
		Request{},
	)
	if err != nil {
		t.Fatalf("managedIdentitiesHandler returned error: %v", err)
	}

	output, ok := outputAny.(models.ManagedIdentitiesOutput)
	if !ok {
		t.Fatalf("unexpected output type %T", outputAny)
	}

	for _, identity := range output.Identities {
		if identity.Name != "ua-ml-ops" {
			continue
		}
		if len(identity.AttachedTo) == 0 || !strings.Contains(identity.AttachedTo[0], "/Microsoft.MachineLearningServices/workspaces/ml-ops-hub") {
			t.Fatalf("expected Azure ML attachment for %q, got %#v", identity.Name, identity.AttachedTo)
		}
		if identity.PrincipalID == nil || *identity.PrincipalID == "" {
			t.Fatalf("expected resolved principal for %q", identity.Name)
		}
		return
	}

	t.Fatalf("expected Azure ML managed identity attachment in output")
}

func TestBuildPersistenceFunctionsOutputResolvesVisibleExecutionContextRoleContext(t *testing.T) {
	contract, ok := contracts.PersistenceSurface("functions")
	if !ok {
		t.Fatal("expected functions persistence surface contract")
	}

	outputAny, err := buildPersistenceFunctionsOutput(
		context.Background(),
		providers.NewStaticProvider(),
		func() time.Time { return time.Unix(0, 0) },
		Request{},
		contract,
	)
	if err != nil {
		t.Fatalf("buildPersistenceFunctionsOutput returned error: %v", err)
	}

	output, ok := outputAny.(models.PersistenceFunctionsOutput)
	if !ok {
		t.Fatalf("unexpected output type %T", outputAny)
	}

	for _, app := range output.FunctionApps {
		if app.Name != "func-orders" {
			continue
		}
		if app.CurrentState.StrongestVisibleExecutionContext == nil {
			t.Fatalf("expected visible execution context for %q", app.Name)
		}
		if got := app.CurrentState.StrongestVisibleExecutionContext.RoleNames; len(got) == 0 {
			t.Fatalf("expected role context for %q, got none", app.Name)
		}
		if got := app.CurrentState.StrongestVisibleExecutionContext.Name; got != "ua-orders" {
			t.Fatalf("expected strongest visible execution context to prefer user-assigned identity, got %q", got)
		}
		for _, item := range app.StillUnmapped {
			if strings.Contains(item, "does not yet rank them into the strongest visible execution context") {
				t.Fatalf("expected user-assigned identity ranking gap to be resolved, got still-unmapped item %q", item)
			}
		}
		return
	}

	t.Fatalf("expected func-orders function app in output")
}

func TestPersistenceFunctionExecutionContextDoesNotOverclaimRBACOnlyFallback(t *testing.T) {
	app := models.FunctionAppAsset{
		Name:                 "func-orders",
		WorkloadIdentityType: models.StringPtr("SystemAssigned"),
		WorkloadPrincipalID:  models.StringPtr("principal-func"),
	}
	context, carriesAzureControl := persistenceFunctionExecutionContext(app, map[string]models.PermissionRow{}, singleRoleAssignment("principal-func", "Owner", "/subscriptions/test"))
	assertRBACOnlyFallbackHedged(t, context, carriesAzureControl)
}

func TestPersistenceFunctionSummaryDoesNotOverclaimExactTriggerDefinitions(t *testing.T) {
	app := models.FunctionAppAsset{
		Name:                 "func-orders",
		DefaultHostname:      models.StringPtr("func-orders.azurewebsites.net"),
		PublicNetworkAccess:  models.StringPtr("Enabled"),
		Runtime:              models.StringPtr("PYTHON|3.11; functions=~4"),
		Deployment:           models.StringPtr("storage=plain-text; kv-refs=1"),
		WorkloadIdentityType: models.StringPtr("SystemAssigned"),
		WorkloadPrincipalID:  models.StringPtr("cccc2222-2222-2222-2222-222222222222"),
		WorkloadIdentityIDs:  []string{"/subscriptions/test/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders"},
	}

	got := persistenceFunctionSummary(app, true, nil, false)
	want := "Current identity can repurpose Function App 'func-orders' as reusable Azure Functions persistence from current RBAC evidence."
	if got != want {
		t.Fatalf("unexpected summary %q, want %q", got, want)
	}
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

func TestPersistenceLogicAppExecutionContextCarriesRBACOnlyFallbackWhenAssignmentsVisible(t *testing.T) {
	workflow := models.LogicAppWorkflowAsset{
		Name:         "la-redeploy",
		IdentityType: models.StringPtr("SystemAssigned"),
		PrincipalID:  models.StringPtr("principal-logic"),
	}
	context, carriesAzureControl := persistenceLogicAppExecutionContext(workflow, map[string]models.PermissionRow{}, singleRoleAssignment("principal-logic", "Contributor", "/subscriptions/test/resourceGroups/rg-apps"))
	assertRBACOnlyFallbackCarriesControl(t, context, carriesAzureControl, "already holds Contributor at resource group rg-apps")
}

func TestPersistenceAzureMLExecutionContextDoesNotOverclaimRBACOnlyFallback(t *testing.T) {
	workspace := models.AzureMLWorkspaceAsset{
		ID:           "/subscriptions/test/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops",
		Name:         "ml-ops",
		IdentityType: models.StringPtr("SystemAssigned"),
		PrincipalID:  models.StringPtr("principal-ml"),
	}
	context, carriesAzureControl := persistenceAzureMLExecutionContext(workspace, nil, map[string]models.PermissionRow{}, singleRoleAssignment("principal-ml", "Owner", "/subscriptions/test"))
	assertRBACOnlyFallbackHedged(t, context, carriesAzureControl)
}

func TestBuildPersistenceWebJobsOutputResolvesInheritedExecutionContext(t *testing.T) {
	contract, ok := contracts.PersistenceSurface("webjobs")
	if !ok {
		t.Fatal("expected webjobs persistence surface contract")
	}

	outputAny, err := buildPersistenceWebJobsOutput(
		context.Background(),
		providers.NewStaticProvider(),
		func() time.Time { return time.Unix(0, 0) },
		Request{},
		contract,
	)
	if err != nil {
		t.Fatalf("buildPersistenceWebJobsOutput returned error: %v", err)
	}

	output, ok := outputAny.(models.PersistenceWebJobsOutput)
	if !ok {
		t.Fatalf("unexpected output type %T", outputAny)
	}

	for _, job := range output.WebJobs {
		if job.Name != "nightly-reconcile" {
			continue
		}
		if job.CurrentState.StrongestVisibleExecutionContext == nil {
			t.Fatalf("expected visible inherited execution context for %q", job.Name)
		}
		if got := job.CurrentState.StrongestVisibleExecutionContext.Name; got != "app-public-api-system" {
			t.Fatalf("expected strongest visible execution context to resolve parent app identity, got %q", got)
		}
		if got := job.CurrentState.ParentAppName; got != "app-public-api" {
			t.Fatalf("expected parent App Service name, got %q", got)
		}
		return
	}

	t.Fatalf("expected nightly-reconcile webjob in output")
}

func TestPersistenceAppServiceStillUnmappedDoesNotDuplicateWebJobsSurfaceBoundary(t *testing.T) {
	items := persistenceAppServiceStillUnmapped()
	for _, item := range items {
		if strings.Contains(item, "`persistence webjobs`") {
			t.Fatalf("expected WebJobs boundary to stay in table walkthrough rather than Not collected by default, got %#v", items)
		}
	}
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
