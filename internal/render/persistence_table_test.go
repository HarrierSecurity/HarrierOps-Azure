package render

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func intPtr(value int) *int {
	return &value
}

var (
	automationStepActions = []string{
		"create or modify account",
		"add or edit runbook",
		"upload or replace code",
		"publish runbook",
		"attach or reuse exec ctx",
		"create schedule",
		"link schedule to runbook",
		"create webhook",
	}
	appServiceStepActions = []string{
		"create or reuse app service",
		"set or reuse deployment path",
		"change app settings or identity attachment",
		"deploy or replace application code",
		"expose or reuse HTTP/HTTPS entry path",
	}
	webJobStepActions = []string{
		"create or reuse parent app service",
		"add or replace webjob package",
		"set or reuse webjob mode",
		"reuse inherited app execution context",
		"leave or repurpose rerun path",
	}
	logicAppStepActions = []string{
		"create or modify workflow",
		"edit workflow definition",
		"attach or reuse exec ctx",
		"define or modify trigger",
		"enable workflow",
		"add or repurpose downstream actions",
	}
	functionStepActions = []string{
		"create or modify function app",
		"deploy or replace code",
		"attach or reuse exec ctx",
		"change app settings or deployment config",
		"repurpose trigger posture",
		"restart or enable function host",
	}
	azureMLStepActions = []string{
		"create or modify workspace",
		"attach or reuse compute",
		"add or modify jobs or pipelines",
		"create or modify schedule",
		"attach or reuse exec ctx",
		"expose or reuse endpoint",
	}
)

func capabilitySteps(actions []string, defaultStatus string, overrides map[string]string) []models.PersistenceCapabilityStep {
	steps := make([]models.PersistenceCapabilityStep, 0, len(actions))
	for _, action := range actions {
		status := defaultStatus
		if override, ok := overrides[action]; ok {
			status = override
		}
		steps = append(steps, models.PersistenceCapabilityStep{Action: action, Status: status})
	}
	return steps
}

func TestPersistenceAutomationTableUsesSingleWalkthroughForMultipleAccounts(t *testing.T) {
	output := persistenceAutomationTable(models.PersistenceAutomationOutput{
		AutomationAccounts: []models.PersistenceAutomationAccount{
			{
				Name:            "aa-one",
				CapabilitySteps: capabilitySteps(automationStepActions, "yes", nil),
				CurrentState:    models.PersistenceAutomationAccountState{},
				Summary:         "summary one",
			},
			{
				Name:            "aa-two",
				CapabilitySteps: capabilitySteps(automationStepActions, "yes", nil),
				CurrentState:    models.PersistenceAutomationAccountState{},
				Summary:         "summary two",
			},
		},
	})

	if strings.Count(output, "Automation capability") != 1 {
		t.Fatalf("expected one shared Automation walkthrough, got:\n%s", output)
	}
	if strings.Contains(output, "Reminder: ") {
		t.Fatalf("expected Automation body renderer to stay free of reminder clutter, got:\n%s", output)
	}
	if !strings.Contains(output, "Visible Automation Accounts") {
		t.Fatalf("expected compact Automation inventory section, got:\n%s", output)
	}
	if !strings.Contains(output, "aa-one") || !strings.Contains(output, "aa-two") {
		t.Fatalf("expected both Automation accounts in compact inventory, got:\n%s", output)
	}
}

func TestPersistenceAppServiceTableUsesSingleWalkthroughForMultipleApps(t *testing.T) {
	trueValue := true
	falseValue := false
	output := persistenceAppServiceTable(models.PersistenceAppServiceOutput{
		AppServices: []models.PersistenceAppService{
			{
				Name:            "app-one",
				CapabilitySteps: capabilitySteps(appServiceStepActions, "yes", nil),
				CurrentState: models.PersistenceAppServiceState{
					Hostname:              models.StringPtr("app-one.azurewebsites.net"),
					PublicNetworkAccess:   models.StringPtr("Enabled"),
					Deployment:            models.StringPtr("repo github.com/contoso/app-one, branch main, run-from-package enabled"),
					AppSettingsCount:      intPtr(4),
					ConnectionStringCount: intPtr(1),
					RunFromPackage:        &trueValue,
					HTTPSOnly:             &trueValue,
				},
				Summary: "summary one",
			},
			{
				Name:            "app-two",
				CapabilitySteps: capabilitySteps(appServiceStepActions, "yes", nil),
				CurrentState: models.PersistenceAppServiceState{
					Hostname:              models.StringPtr("app-two.azurewebsites.net"),
					PublicNetworkAccess:   models.StringPtr("Enabled"),
					Deployment:            models.StringPtr("run-from-package disabled"),
					AppSettingsCount:      intPtr(2),
					ConnectionStringCount: intPtr(0),
					RunFromPackage:        &falseValue,
					HTTPSOnly:             &falseValue,
				},
				Summary: "summary two",
			},
		},
	})

	if strings.Count(output, "App Service capability") != 1 {
		t.Fatalf("expected one shared App Service walkthrough, got:\n%s", output)
	}
	if strings.Contains(output, "Reminder: ") {
		t.Fatalf("expected App Service body renderer to stay free of reminder clutter, got:\n%s", output)
	}
	if !strings.Contains(output, "Visible App Services") {
		t.Fatalf("expected compact App Service inventory section, got:\n%s", output)
	}
	if !strings.Contains(output, "app-one") || !strings.Contains(output, "app-two") {
		t.Fatalf("expected both App Services in compact inventory, got:\n%s", output)
	}
	if !strings.Contains(output, "This App Service view stops at the main web host; use `persistence webjobs` when you need App Service WebJobs background-execution depth.") {
		t.Fatalf("expected App Service-specific WebJobs boundary callout, got:\n%s", output)
	}
}

func TestPersistenceAppServiceTableCarriesNearbyThematicNames(t *testing.T) {
	trueValue := true
	output := persistenceAppServiceTable(models.PersistenceAppServiceOutput{
		AppServices: []models.PersistenceAppService{
			{
				Name:            "app-public-api",
				CapabilitySteps: capabilitySteps(appServiceStepActions, "yes", nil),
				CurrentState: models.PersistenceAppServiceState{
					Hostname:            models.StringPtr("app-public-api.azurewebsites.net"),
					PublicNetworkAccess: models.StringPtr("Enabled"),
					Deployment:          models.StringPtr("repo github.com/contoso/customer-portal, branch main"),
					RunFromPackage:      &trueValue,
					HTTPSOnly:           &trueValue,
					NearbyThematicNames: []string{"app-nightly-sync", "app-maintenance-api"},
				},
			},
		},
	})

	if !strings.Contains(output, "Nearby maintenance- or schedule-themed names visible from the current environment include `app-nightly-sync` and `app-maintenance-api`.") {
		t.Fatalf("expected App Service walkthrough to carry nearby thematic names line, got:\n%s", output)
	}
}

func TestPersistenceAppServiceTableStopsWalkthroughAtFirstBrokenStep(t *testing.T) {
	trueValue := true
	output := persistenceAppServiceTable(models.PersistenceAppServiceOutput{
		AppServices: []models.PersistenceAppService{
			{
				Name: "app-public-api",
				CapabilitySteps: capabilitySteps(appServiceStepActions, "yes", map[string]string{
					"deploy or replace application code": "not proven",
				}),
				CurrentState: models.PersistenceAppServiceState{
					State:                         models.StringPtr("Running"),
					Hostname:                      models.StringPtr("app-public-api.azurewebsites.net"),
					PublicNetworkAccess:           models.StringPtr("Enabled"),
					Runtime:                       models.StringPtr("NODE|20-lts"),
					Deployment:                    models.StringPtr("repo github.com/contoso/customer-portal, branch main, GitHub Actions, continuous integration, run-from-package enabled"),
					AppSettingsCount:              intPtr(4),
					KeyVaultReferenceCount:        intPtr(2),
					SensitiveSettingCount:         intPtr(1),
					ConnectionStringCount:         intPtr(2),
					KeyVaultConnectionStringCount: intPtr(1),
					ConnectionStringTypes:         []string{"Custom", "SQLAzure"},
					RunFromPackage:                &trueValue,
					HTTPSOnly:                     &trueValue,
					VisibleSensitiveSettingNames:  []string{"DB_PASSWORD"},
				},
			},
		},
	})

	if !strings.Contains(output, "Current identity can set or reuse the deployment path this App Service will load.") {
		t.Fatalf("expected App Service walkthrough to keep deployment path distinct, got:\n%s", output)
	}
	if !strings.Contains(output, "Visible deployment signals here include repo github.com/contoso/customer-portal, branch main, GitHub Actions, continuous integration, run-from-package enabled.") {
		t.Fatalf("expected App Service walkthrough to surface deployment path truth, got:\n%s", output)
	}
	if !strings.Contains(output, "Current identity can change app settings and attach or reuse managed identity for this App Service.") {
		t.Fatalf("expected App Service walkthrough to keep config and identity bullet, got:\n%s", output)
	}
	if !strings.Contains(output, "Visible config posture here includes 4 app setting(s), 2 Key Vault-backed setting(s), 1 sensitive-looking setting name(s), 2 connection string(s), 1 Key Vault-backed connection string(s), connection types Custom, SQLAzure.") {
		t.Fatalf("expected App Service walkthrough to show config posture, got:\n%s", output)
	}
	if !strings.Contains(output, "Visible sensitive-looking setting names include `DB_PASSWORD`.") {
		t.Fatalf("expected App Service walkthrough to use visible sensitive setting names when present, got:\n%s", output)
	}
	if strings.Contains(output, "Current identity can expose or reuse this app's HTTP or HTTPS entry path so it remains reachable later.") {
		t.Fatalf("expected App Service walkthrough to stop before later bullets when code replacement is not proven, got:\n%s", output)
	}
}

func TestPersistenceWebJobsTableUsesSingleWalkthroughForMultipleJobs(t *testing.T) {
	output := persistenceWebJobsTable(models.PersistenceWebJobsOutput{
		WebJobs: []models.PersistenceWebJob{
			{
				Name:            "queue-worker",
				CapabilitySteps: capabilitySteps(webJobStepActions, "yes", nil),
				CurrentState: models.PersistenceWebJobState{
					Mode:          "continuous",
					RunCommand:    models.StringPtr("node /home/site/wwwroot/app_data/jobs/continuous/queue-worker/index.js"),
					ParentAppName: "app-public-api",
				},
				Summary: "summary one",
			},
			{
				Name:            "nightly-reconcile",
				CapabilitySteps: capabilitySteps(webJobStepActions, "yes", nil),
				CurrentState: models.PersistenceWebJobState{
					Mode:             "scheduled",
					LatestRunTrigger: models.StringPtr("Schedule"),
					ParentAppName:    "app-public-api",
				},
				Summary: "summary two",
			},
		},
	})

	if strings.Count(output, "WebJob capability") != 1 {
		t.Fatalf("expected one shared WebJobs walkthrough, got:\n%s", output)
	}
	if strings.Contains(output, "Reminder: ") {
		t.Fatalf("expected WebJobs body renderer to stay free of reminder clutter, got:\n%s", output)
	}
	if !strings.Contains(output, "Visible WebJobs") {
		t.Fatalf("expected compact WebJobs inventory section, got:\n%s", output)
	}
	if !strings.Contains(output, "queue-worker") || !strings.Contains(output, "nightly-reconcile") {
		t.Fatalf("expected both WebJobs in compact inventory, got:\n%s", output)
	}
	if !strings.Contains(output, "Kudu and the App Service runtime discover the job from the deployed WebJobs path and run it again according to that mode.") {
		t.Fatalf("expected WebJobs walkthrough to keep platform discovery detail, got:\n%s", output)
	}
}

func TestPersistenceLogicAppsTableUsesSingleWalkthroughForMultipleWorkflows(t *testing.T) {
	output := persistenceLogicAppsTable(models.PersistenceLogicAppsOutput{
		Workflows: []models.PersistenceLogicAppWorkflow{
			{
				Name:            "la-one",
				CapabilitySteps: capabilitySteps(logicAppStepActions, "yes", nil),
				CurrentState: models.PersistenceLogicAppWorkflowState{
					Classification:                   "persistence-capable",
					ExternallyCallableRequestTrigger: true,
					NearbyThematicNames:              []string{"nightly-sync", "maintenance-router"},
				},
				Summary: "logic summary one",
			},
			{
				Name:            "la-two",
				CapabilitySteps: capabilitySteps(logicAppStepActions, "yes", nil),
				CurrentState: models.PersistenceLogicAppWorkflowState{
					Classification: "persistence-capable",
				},
				Summary: "logic summary two",
			},
		},
	})

	if strings.Count(output, "Logic Apps capability") != 1 {
		t.Fatalf("expected one shared Logic Apps walkthrough, got:\n%s", output)
	}
	if strings.Contains(output, "Reminder: ") {
		t.Fatalf("expected Logic Apps body renderer to stay free of reminder clutter, got:\n%s", output)
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

func TestPersistenceAutomationTableCarriesVisibilityWhenControlNotProven(t *testing.T) {
	output := persistenceAutomationTable(models.PersistenceAutomationOutput{
		AutomationAccounts: []models.PersistenceAutomationAccount{
			{
				Name:            "aa-quiet",
				CapabilitySteps: capabilitySteps(automationStepActions, "not proven", nil),
				CurrentState: models.PersistenceAutomationAccountState{
					PublishedRunbookCount: intPtr(1),
					RunbookCount:          intPtr(2),
					ScheduleCount:         intPtr(1),
					WebhookCount:          intPtr(0),
					PrimaryStartMode:      models.StringPtr("schedule"),
				},
			},
		},
	})

	if !strings.Contains(output, "Visibility still shows 1/2 published; schedules 1; webhooks 0; primary schedule;") {
		t.Fatalf("expected non-proven Automation walkthrough to keep visible state, got:\n%s", output)
	}
	if !strings.Contains(output, "trigger posture, or reuse value if stronger control is obtained later") {
		t.Fatalf("expected non-proven Automation walkthrough to explain visibility value, got:\n%s", output)
	}
	if strings.Contains(output, "Current identity does not yet have a proven path to add or edit a runbook inside this Azure Automation Account.") {
		t.Fatalf("expected non-proven Automation walkthrough to stop after the first unproven step, got:\n%s", output)
	}
}

func TestPersistenceFunctionsTableCarriesVisibilityWhenControlNotProven(t *testing.T) {
	httpTrigger := "HTTP"
	timerTrigger := "timer"
	output := persistenceFunctionsTable(models.PersistenceFunctionsOutput{
		FunctionApps: []models.PersistenceFunctionApp{
			{
				Name:            "func-orders",
				CapabilitySteps: capabilitySteps(functionStepActions, "not proven", nil),
				CurrentState: models.PersistenceFunctionAppState{
					State:                        models.StringPtr("Running"),
					Hostname:                     models.StringPtr("func-orders.azurewebsites.net"),
					Runtime:                      models.StringPtr("PYTHON|3.11; functions=~4"),
					Deployment:                   models.StringPtr("storage=plain-text; kv-refs=1"),
					PublicNetworkAccess:          models.StringPtr("Enabled"),
					AzureWebJobsStorageValueType: models.StringPtr("plain-text"),
					TriggerTypes:                 []string{httpTrigger, timerTrigger},
				},
			},
		},
	})

	if !strings.Contains(output, "Visibility still shows Running; hostname visible; PYTHON|3.11; functions=~4; storage=plain-text; kv-refs=1; public Enabled; triggers=HTTP, timer.") {
		t.Fatalf("expected non-proven Functions walkthrough to keep visible state, got:\n%s", output)
	}
	if !strings.Contains(output, "That is enough to judge whether this Function App already has trigger exposure, deployment signals, or reuse value if stronger control is obtained later.") {
		t.Fatalf("expected non-proven Functions walkthrough to explain visibility value, got:\n%s", output)
	}
	if !strings.Contains(output, "trigger exposure, deployment signals, or reuse value if stronger control is obtained later") {
		t.Fatalf("expected non-proven Functions walkthrough to explain visibility value, got:\n%s", output)
	}
	if strings.Contains(output, "Current identity does not yet have a proven path to deploy or replace the function package Azure would load in this Function App.") {
		t.Fatalf("expected non-proven Functions walkthrough to stop after the first unproven step, got:\n%s", output)
	}
}

func TestPersistenceFunctionsTableStopsWalkthroughAtFirstBrokenStep(t *testing.T) {
	httpTrigger := "HTTP"
	timerTrigger := "timer"
	falseValue := false
	trueValue := true
	output := persistenceFunctionsTable(models.PersistenceFunctionsOutput{
		FunctionApps: []models.PersistenceFunctionApp{
			{
				Name: "func-orders",
				CapabilitySteps: capabilitySteps(functionStepActions, "yes", map[string]string{
					"attach or reuse exec ctx":        "not proven",
					"restart or enable function host": "not proven",
				}),
				CurrentState: models.PersistenceFunctionAppState{
					State:                        models.StringPtr("Running"),
					Hostname:                     models.StringPtr("func-orders.azurewebsites.net"),
					Runtime:                      models.StringPtr("PYTHON|3.11; functions=~4"),
					Deployment:                   models.StringPtr("storage=plain-text; kv-refs=1"),
					PublicNetworkAccess:          models.StringPtr("Enabled"),
					AzureWebJobsStorageValueType: models.StringPtr("plain-text"),
					TriggerTypes:                 []string{httpTrigger, timerTrigger},
					VisibleFunctions: []models.FunctionChildAsset{
						{Name: "OrdersWebhook", TriggerType: &httpTrigger, IsDisabled: &falseValue},
						{Name: "NightlyReconcile", TriggerType: &timerTrigger, IsDisabled: &trueValue},
					},
				},
			},
		},
	})

	if !strings.Contains(output, "Current identity can deploy or replace the function package Azure will load in this Function App.") {
		t.Fatalf("expected Functions walkthrough to keep proven steps before the boundary, got:\n%s", output)
	}
	if !strings.Contains(output, "Because the current identity already controls this Function App, zip deploy, publish, or package replacement are part of the defended Functions persistence path here.") {
		t.Fatalf("expected Functions walkthrough to explain the defended deploy path on its own line, got:\n%s", output)
	}
	if !strings.Contains(output, "Visible deployment posture includes storage=plain-text.") {
		t.Fatalf("expected Functions walkthrough to split deployment posture into follow-on lines, got:\n%s", output)
	}
	if !strings.Contains(output, "Current identity can change app settings, identity attachment, and deployment configuration for this Function App.") {
		t.Fatalf("expected Functions walkthrough to keep the config step before the boundary, got:\n%s", output)
	}
	if !strings.Contains(output, "Current identity can repurpose this Function App's trigger exposure so Azure has a way to run it again later, including HTTP-style externally reachable entrypoints.") {
		t.Fatalf("expected Functions walkthrough to keep the trigger step before the boundary, got:\n%s", output)
	}
	if !strings.Contains(output, "Visible child functions here show `HTTP` and `timer` trigger paths.") {
		t.Fatalf("expected Functions walkthrough to use visible trigger truth, got:\n%s", output)
	}
	if !strings.Contains(output, "Current visible functions include OrdersWebhook [HTTP], NightlyReconcile [timer; disabled].") {
		t.Fatalf("expected Functions walkthrough to name visible child functions, got:\n%s", output)
	}
	if !strings.Contains(output, "HTTP-triggered functions are visible from management-plane metadata.") {
		t.Fatalf("expected Functions walkthrough to define what the HTTP trigger truth comes from, got:\n%s", output)
	}
	if !strings.Contains(output, "Timer, queue, Service Bus, or other event-driven triggers are visible from bindings, but they are not the same as a directly callable public entrypoint.") {
		t.Fatalf("expected Functions walkthrough to define the internal trigger boundary, got:\n%s", output)
	}
	if !strings.Contains(output, "The remaining gap is data-plane and runtime-side validation the current management-plane collector does not perform.") {
		t.Fatalf("expected Functions walkthrough to explain the management-plane boundary, got:\n%s", output)
	}
	if !strings.Contains(output, "Current identity does not yet have a proven path to attach or reuse execution context for this Function App.") {
		t.Fatalf("expected Functions walkthrough to show the first broken step, got:\n%s", output)
	}
	if strings.Index(output, "Current identity can deploy or replace the function package Azure will load in this Function App.") > strings.Index(output, "Current identity can repurpose this Function App's trigger exposure so Azure has a way to run it again later, including HTTP-style externally reachable entrypoints.") {
		t.Fatalf("expected Functions walkthrough to keep deploy before trigger, got:\n%s", output)
	}
	if strings.Contains(output, "Current identity does not yet have a proven path to restart or enable this Function App for later trigger-driven execution.") {
		t.Fatalf("expected Functions walkthrough to stop after the first broken step, got:\n%s", output)
	}
}

func TestPersistenceAutomationTableStopsWalkthroughAtFirstBrokenStep(t *testing.T) {
	output := persistenceAutomationTable(models.PersistenceAutomationOutput{
		AutomationAccounts: []models.PersistenceAutomationAccount{
			{
				Name: "aa-quiet",
				CapabilitySteps: capabilitySteps(automationStepActions, "not proven", map[string]string{
					"create or modify account": "yes",
					"add or edit runbook":      "yes",
				}),
				CurrentState: models.PersistenceAutomationAccountState{
					PublishedRunbookCount: intPtr(1),
					RunbookCount:          intPtr(2),
					ScheduleCount:         intPtr(1),
					WebhookCount:          intPtr(0),
					PrimaryStartMode:      models.StringPtr("schedule"),
				},
			},
		},
	})

	if !strings.Contains(output, "Current identity can add or edit a runbook inside an existing Azure Automation Account.") {
		t.Fatalf("expected Automation walkthrough to keep proven steps before the boundary, got:\n%s", output)
	}
	if !strings.Contains(output, "Current identity does not yet have a proven path to upload or replace code inside a runbook.") {
		t.Fatalf("expected Automation walkthrough to show the first broken step, got:\n%s", output)
	}
	if strings.Contains(output, "Current identity does not yet have a proven path to publish runnable automation here.") {
		t.Fatalf("expected Automation walkthrough to stop after the first broken step, got:\n%s", output)
	}
}

func TestPersistenceAzureMLTableUsesComputeAndResolvedIdentityTruth(t *testing.T) {
	output := persistenceAzureMLTable(models.PersistenceAzureMLOutput{
		Workspaces: []models.PersistenceAzureMLWorkspace{
			{
				Name:                    "ml-ops-hub",
				ResourceGroup:           "rg-ml",
				CapabilitySteps:         capabilitySteps(azureMLStepActions, "yes", nil),
				ExecutionContextOptions: []string{"managed identity", "workspace-linked storage", "workspace-linked key vault"},
				CurrentState: models.PersistenceAzureMLWorkspaceState{
					Classification:                   "execution-capable",
					PublicNetworkAccess:              models.StringPtr("Enabled"),
					VisibleIdentityNames:             []string{"ua-ml-ops", "ml-ops-hub-workspace-identity"},
					ComputeCount:                     intPtr(2),
					ComputeTypes:                     []string{"ComputeCluster", "ComputeInstance"},
					JobCount:                         intPtr(2),
					JobTypes:                         []string{"Command", "Pipeline"},
					ScheduleCount:                    intPtr(1),
					ScheduleTriggerTypes:             []string{"Cron"},
					EndpointCount:                    intPtr(1),
					EndpointAuthModes:                []string{"AADToken"},
					EndpointPublicAccess:             []string{"Enabled"},
					StrongestVisibleExecutionContext: &models.PersistenceRoleContext{Name: "ua-ml-ops", RoleNames: []string{"Owner"}, ScopeIDs: []string{"/subscriptions/test"}, Summary: "The strongest visible execution context here is the Azure ML identity `ua-ml-ops`, which already holds Owner at subscription scope."},
					NearbyThematicNames:              []string{"ml-nightly-train", "ml-catalog"},
				},
				StillUnmapped: []string{
					"the current command does not retrieve notebook content, model content, environment definitions, or job or pipeline payloads from Azure ML workspaces, so operator intent is not inferred from Azure ML content here",
				},
			},
		},
	})

	if !strings.Contains(output, "Current identity can attach or reuse Azure ML compute for this workspace, including long-lived instances or cluster-backed execution.") {
		t.Fatalf("expected Azure ML walkthrough to carry an explicit compute step, got:\n%s", output)
	}
	if !strings.Contains(output, "In Azure ML, persistence can live in saved notebooks, jobs, pipelines, scheduled jobs, and environment definitions.") {
		t.Fatalf("expected Azure ML walkthrough to mention stored execution logic locations, got:\n%s", output)
	}
	if !strings.Contains(output, "When a notebook, job, or pipeline runs later, it executes with the attached identity plus the linked workspace resources Azure ML will use at runtime.") {
		t.Fatalf("expected Azure ML walkthrough to explain re-triggered execution flow, got:\n%s", output)
	}
	if !strings.Contains(output, "The persistence story here is the workspace plus compute plus stored code and re-entry paths that can all remain in place for later execution.") {
		t.Fatalf("expected Azure ML walkthrough to explain why this acts like persistence, got:\n%s", output)
	}
	if !strings.Contains(output, "The strongest visible execution context here is the Azure ML identity `ua-ml-ops`, which already holds Owner at subscription scope.") {
		t.Fatalf("expected Azure ML walkthrough to keep the strongest execution-context proof, got:\n%s", output)
	}
	if !strings.Contains(output, "Nearby maintenance- or schedule-themed names visible from the current environment include `ml-nightly-train` and `ml-catalog`.") {
		t.Fatalf("expected Azure ML walkthrough to keep the nearby thematic names closer, got:\n%s", output)
	}
	if strings.Contains(output, "Visible compute types here already include ComputeCluster, ComputeInstance.") {
		t.Fatalf("expected Azure ML walkthrough to leave visible compute-type restatement to inventory, got:\n%s", output)
	}
	if strings.Contains(output, "Current workspace visibility already shows 2 compute target(s).") {
		t.Fatalf("expected Azure ML walkthrough to avoid duplicate compute counts, got:\n%s", output)
	}
	if strings.Contains(output, "Visible job types here already include Command, Pipeline.") {
		t.Fatalf("expected Azure ML walkthrough to leave visible job-type restatement to inventory, got:\n%s", output)
	}
	if strings.Contains(output, "Current workspace visibility already shows 2 job(s).") {
		t.Fatalf("expected Azure ML walkthrough to avoid duplicate job counts, got:\n%s", output)
	}
	if strings.Contains(output, "Visible attached identities here include `ua-ml-ops` and `ml-ops-hub-workspace-identity`.") {
		t.Fatalf("expected Azure ML walkthrough to avoid duplicate visible identity restatement, got:\n%s", output)
	}
	if strings.Contains(output, "Current workspace visibility already shows 1 online endpoint(s).") {
		t.Fatalf("expected Azure ML walkthrough to avoid duplicate endpoint counts, got:\n%s", output)
	}
	if strings.Contains(output, "Current output gap") {
		t.Fatalf("expected Azure ML table to avoid a current output gap when identities resolve cleanly, got:\n%s", output)
	}
}
