package render

import (
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

func persistenceTableRenderer(payload any) (string, error) {
	switch out := payload.(type) {
	case models.PersistenceOverviewOutput:
		return persistenceOverviewTable(out), nil
	case models.PersistenceAutomationOutput:
		return persistenceAutomationTable(out), nil
	case models.PersistenceAppServiceOutput:
		return persistenceAppServiceTable(out), nil
	case models.PersistenceWebJobsOutput:
		return persistenceWebJobsTable(out), nil
	case models.PersistenceContainerAppsJobsOutput:
		return persistenceContainerAppsJobsTable(out), nil
	case models.PersistenceVMExtensionsOutput:
		return persistenceVMExtensionsTable(out), nil
	case models.PersistenceAzureMLOutput:
		return persistenceAzureMLTable(out), nil
	case models.PersistenceFunctionsOutput:
		return persistenceFunctionsTable(out), nil
	case models.PersistenceLogicAppsOutput:
		return persistenceLogicAppsTable(out), nil
	default:
		return "", fmt.Errorf("unexpected payload type for persistence: %T", payload)
	}
}

func persistenceOverviewTable(payload models.PersistenceOverviewOutput) string {
	rows := make([][]string, 0, len(payload.Surfaces))
	for _, surface := range payload.Surfaces {
		rows = append(rows, []string{
			surface.Surface,
			surface.Summary,
		})
	}
	return renderListTable(
		"ho-azure persistence",
		[]string{"surface", "summary"},
		rows,
		[]string{"no persistence surfaces available", ""},
		persistenceOverviewTakeaway(payload),
	)
}

func persistenceAutomationTable(payload models.PersistenceAutomationOutput) string {
	if len(payload.AutomationAccounts) == 0 {
		return renderListTable(
			"ho-azure persistence automation",
			[]string{"automation account", "status"},
			nil,
			[]string{"No visible Automation accounts were confirmed from current scope.", ""},
			"0 Automation accounts visible; no Azure Automation persistence surface was confirmed from current scope.",
		)
	}

	lead := persistenceAutomationLeadAccount(payload.AutomationAccounts)
	lines := []string{
		"Automation capability",
		"",
		renderPersistenceSectionTable(
			[]string{"action", "api surface", "status"},
			persistenceCapabilityRows(lead.CapabilitySteps),
		),
	}
	if len(payload.AutomationAccounts) > 1 {
		lines = append(lines, "This walkthrough shows the strongest currently visible Automation persistence path. The inventory below lists the other visible accounts without repeating the same narrative.")
	}
	lines = append(lines,
		persistenceAutomationExplanation(lead),
		"",
		"Visible Automation Accounts",
		renderAlignedPipeTable(
			[]string{"automation account", "resource group", "visible state", "execution context"},
			persistenceAutomationInventoryRows(payload.AutomationAccounts),
		),
	)
	if unmapped := persistenceAutomationCombinedStillUnmapped(payload.AutomationAccounts); unmapped != "" {
		lines = append(lines, "", "Not collected by default", unmapped)
	}

	return strings.Join(lines, "\n")
}

func persistenceAppServiceTable(payload models.PersistenceAppServiceOutput) string {
	if len(payload.AppServices) == 0 {
		return renderListTable(
			"ho-azure persistence app-service",
			[]string{"app service", "status"},
			nil,
			[]string{"No visible App Services were confirmed from current scope.", ""},
			"0 App Services visible; no App Service persistence surface was confirmed from current scope.",
		)
	}

	lead := persistenceAppServiceLeadApp(payload.AppServices)
	lines := []string{
		"App Service capability",
		"",
		renderPersistenceSectionTable(
			[]string{"action", "api surface", "status"},
			persistenceCapabilityRows(lead.CapabilitySteps),
		),
	}
	if len(payload.AppServices) > 1 {
		lines = append(lines, "This walkthrough shows the strongest currently visible App Service persistence path. The inventory below lists the other visible App Services without repeating the same narrative.")
	}
	lines = append(lines,
		persistenceAppServiceExplanation(lead),
		"",
		"Visible App Services",
		renderAlignedPipeTable(
			[]string{"app service", "resource group", "visible state", "execution context"},
			persistenceAppServiceInventoryRows(payload.AppServices),
		),
	)
	if items := persistenceAppServiceCombinedStillUnmapped(payload.AppServices); len(items) > 0 {
		lines = append(lines, "", "Not collected by default", renderBulletList(items))
	}

	return strings.Join(lines, "\n")
}

func persistenceWebJobsTable(payload models.PersistenceWebJobsOutput) string {
	if len(payload.WebJobs) == 0 {
		return renderListTable(
			"ho-azure persistence webjobs",
			[]string{"webjob", "status"},
			nil,
			[]string{"No visible WebJobs were confirmed from current scope.", ""},
			"0 WebJobs visible; no WebJobs persistence surface was confirmed from current scope.",
		)
	}

	lead := persistenceWebJobsLeadJob(payload.WebJobs)
	lines := []string{
		"WebJob capability",
		"",
		renderPersistenceSectionTable(
			[]string{"action", "api surface", "status"},
			persistenceCapabilityRows(lead.CapabilitySteps),
		),
	}
	if len(payload.WebJobs) > 1 {
		lines = append(lines, "This walkthrough shows the strongest currently visible WebJobs persistence path. The inventory below lists the other visible WebJobs without repeating the same narrative.")
	}
	lines = append(lines,
		persistenceWebJobsExplanation(lead),
		"",
		"Visible WebJobs",
		renderAlignedPipeTable(
			[]string{"webjob", "parent app", "visible state", "execution context"},
			persistenceWebJobsInventoryRows(payload.WebJobs),
		),
	)
	if items := persistenceWebJobsCombinedStillUnmapped(payload.WebJobs); len(items) > 0 {
		lines = append(lines, "", "Not collected by default", renderBulletList(items))
	}

	return strings.Join(lines, "\n")
}

func persistenceContainerAppsJobsTable(payload models.PersistenceContainerAppsJobsOutput) string {
	if len(payload.ContainerAppsJobs) == 0 {
		return renderListTable(
			"ho-azure persistence container-apps-jobs",
			[]string{"container apps job", "status"},
			nil,
			[]string{"No visible Container Apps Jobs were confirmed from current scope.", ""},
			"0 Container Apps Jobs visible; no Container Apps Jobs persistence surface was confirmed from current scope.",
		)
	}

	lead := persistenceContainerAppsJobsLeadJob(payload.ContainerAppsJobs)
	lines := []string{
		"Container Apps Jobs capability",
		"",
		renderPersistenceSectionTable(
			[]string{"action", "api surface", "status"},
			persistenceCapabilityRows(lead.CapabilitySteps),
		),
	}
	if len(payload.ContainerAppsJobs) > 1 {
		lines = append(lines, "This walkthrough shows the strongest currently visible Container Apps Jobs persistence path. The inventory below lists the other visible jobs without repeating the same narrative.")
	}
	lines = append(lines,
		persistenceContainerAppsJobsExplanation(lead),
		"",
		"Visible Container Apps Jobs",
		renderAlignedPipeTable(
			[]string{"container apps job", "trigger", "visible state", "execution context"},
			persistenceContainerAppsJobsInventoryRows(payload.ContainerAppsJobs),
		),
	)
	if items := persistenceContainerAppsJobsCombinedStillUnmapped(payload.ContainerAppsJobs); len(items) > 0 {
		lines = append(lines, "", "Not collected by default", renderBulletList(items))
	}

	return strings.Join(lines, "\n")
}

func persistenceVMExtensionsTable(payload models.PersistenceVMExtensionsOutput) string {
	if len(payload.VMExtensions) == 0 {
		return renderListTable(
			"ho-azure persistence vm-extensions",
			[]string{"vm extension", "status"},
			nil,
			[]string{"No visible VM Extensions were confirmed from current scope.", ""},
			"0 VM Extensions visible; no VM Extensions persistence surface was confirmed from current scope.",
		)
	}

	lead := persistenceVMExtensionsLeadExtension(payload.VMExtensions)
	lines := []string{
		"VM Extensions capability",
		"",
		renderPersistenceSectionTable(
			[]string{"action", "api surface", "status"},
			persistenceCapabilityRows(lead.CapabilitySteps),
		),
	}
	if len(payload.VMExtensions) > 1 {
		lines = append(lines, "This walkthrough shows the strongest currently visible VM Extensions persistence path. The inventory below lists the other visible VM and VMSS extensions without repeating the same narrative.")
	}
	lines = append(lines,
		persistenceVMExtensionsExplanation(lead),
		"",
		"Visible VM Extensions",
		renderAlignedPipeTable(
			[]string{"extension", "target", "visible state", "execution context"},
			persistenceVMExtensionsInventoryRows(payload.VMExtensions),
		),
	)
	if items := persistenceVMExtensionsCombinedStillUnmapped(payload.VMExtensions); len(items) > 0 {
		lines = append(lines, "", "Not collected by default", renderBulletList(items))
	}

	return strings.Join(lines, "\n")
}

func persistenceAzureMLTable(payload models.PersistenceAzureMLOutput) string {
	if len(payload.Workspaces) == 0 {
		return renderListTable(
			"ho-azure persistence azure-ml",
			[]string{"workspace", "status"},
			nil,
			[]string{"No visible Azure ML workspaces were confirmed from current scope.", ""},
			"0 Azure ML workspaces visible; no Azure ML persistence surface was confirmed from current scope.",
		)
	}

	lead := persistenceAzureMLLeadWorkspace(payload.Workspaces)
	lines := []string{
		"Azure ML capability",
		"",
		renderPersistenceSectionTable(
			[]string{"action", "api surface", "status"},
			persistenceCapabilityRows(lead.CapabilitySteps),
		),
	}
	if len(payload.Workspaces) > 1 {
		lines = append(lines, "This walkthrough shows the strongest currently visible Azure ML persistence path. The inventory below lists the other visible workspaces without repeating the same narrative.")
	}
	lines = append(lines,
		persistenceAzureMLExplanation(lead),
		"",
		"Visible Azure ML Workspaces",
		renderAlignedPipeTable(
			[]string{"workspace", "resource group", "visible state", "execution context"},
			persistenceAzureMLInventoryRows(payload.Workspaces),
		),
	)
	defaultItems, currentGapItems := persistenceAzureMLBoundarySections(payload.Workspaces)
	if len(defaultItems) > 0 {
		lines = append(lines, "", "Not collected by default", renderBulletList(defaultItems))
	}
	if len(currentGapItems) > 0 {
		lines = append(lines, "", "Current output gap", renderBulletList(currentGapItems))
	}
	return strings.Join(lines, "\n")
}

func persistenceLogicAppsTable(payload models.PersistenceLogicAppsOutput) string {
	if len(payload.Workflows) == 0 {
		return renderListTable(
			"ho-azure persistence logic-apps",
			[]string{"logic app", "status"},
			nil,
			[]string{"No visible Logic Apps were confirmed from current scope.", ""},
			"0 Logic Apps visible; no workflow persistence surface was confirmed from current scope.",
		)
	}

	lead := persistenceLogicAppLeadWorkflow(payload.Workflows)
	lines := []string{
		"Logic Apps capability",
		"",
		renderPersistenceSectionTable(
			[]string{"action", "api surface", "status"},
			persistenceCapabilityRows(lead.CapabilitySteps),
		),
	}
	if len(payload.Workflows) > 1 {
		lines = append(lines, "This walkthrough shows the strongest currently visible Logic App persistence path. The inventory below lists the other visible workflows without repeating the same narrative.")
	}
	lines = append(lines,
		persistenceLogicAppExplanation(lead),
		"",
		"Visible Logic Apps",
		renderAlignedPipeTable(
			[]string{"logic app", "resource group", "visible state", "execution context"},
			persistenceLogicAppInventoryRows(payload.Workflows),
		),
	)
	if items := persistenceLogicAppCombinedStillUnmapped(payload.Workflows); len(items) > 0 {
		lines = append(lines, "", "Not collected by default", renderBulletList(items))
	}

	return strings.Join(lines, "\n")
}

func persistenceFunctionsTable(payload models.PersistenceFunctionsOutput) string {
	if len(payload.FunctionApps) == 0 {
		return renderListTable(
			"ho-azure persistence functions",
			[]string{"function app", "status"},
			nil,
			[]string{"No visible Function Apps were confirmed from current scope.", ""},
			"0 Function Apps visible; no Azure Functions persistence surface was confirmed from current scope.",
		)
	}

	lead := persistenceFunctionsLeadApp(payload.FunctionApps)
	lines := []string{
		"Azure Functions capability",
		"",
		renderPersistenceSectionTable(
			[]string{"action", "api surface", "status"},
			persistenceCapabilityRows(lead.CapabilitySteps),
		),
	}
	if len(payload.FunctionApps) > 1 {
		lines = append(lines, "This walkthrough shows the strongest currently visible Azure Functions persistence path. The inventory below lists the other visible Function Apps without repeating the same narrative.")
	}
	lines = append(lines,
		persistenceFunctionsExplanation(lead),
		"",
		"Visible Function Apps",
		renderAlignedPipeTable(
			[]string{"function app", "resource group", "visible state", "execution context"},
			persistenceFunctionsInventoryRows(payload.FunctionApps),
		),
	)
	defaultItems, currentGapItems := persistenceFunctionsBoundarySections(payload.FunctionApps)
	if len(defaultItems) > 0 {
		lines = append(lines, "", "Not collected by default", renderBulletList(defaultItems))
	}
	if len(currentGapItems) > 0 {
		lines = append(lines, "", "Current output gap", renderBulletList(currentGapItems))
	}

	return strings.Join(lines, "\n")
}

func persistenceCapabilityRows(steps []models.PersistenceCapabilityStep) [][]string {
	rows := make([][]string, 0, len(steps))
	for _, step := range steps {
		rows = append(rows, []string{step.Action, step.APISurface, step.Status})
	}
	return rows
}

func persistenceOverviewTakeaway(payload models.PersistenceOverviewOutput) string {
	if len(payload.Surfaces) == 0 {
		return "No persistence surfaces are available yet."
	}
	if len(payload.Surfaces) == 1 {
		return fmt.Sprintf("1 persistence surface is available; start with %s.", payload.Surfaces[0].Surface)
	}
	return fmt.Sprintf("%d persistence surfaces are available; start with the service that best matches your current question.", len(payload.Surfaces))
}

func renderAlignedPipeTable(headers []string, rows [][]string) string {
	widths := make([]int, len(headers))
	for index, header := range headers {
		widths[index] = len(header)
	}
	for _, row := range rows {
		for index, cell := range row {
			if index >= len(widths) {
				continue
			}
			for _, line := range strings.Split(cell, "\n") {
				if len(line) > widths[index] {
					widths[index] = len(line)
				}
			}
		}
	}

	renderRow := func(cells []string) string {
		cellLines := make([][]string, len(widths))
		height := 1
		for index := range widths {
			lines := []string{""}
			if index < len(cells) && strings.TrimSpace(cells[index]) != "" {
				lines = strings.Split(cells[index], "\n")
			}
			cellLines[index] = lines
			if len(lines) > height {
				height = len(lines)
			}
		}
		rendered := make([]string, 0, height)
		for lineIndex := 0; lineIndex < height; lineIndex++ {
			parts := make([]string, len(widths))
			for index := range widths {
				cell := ""
				if lineIndex < len(cellLines[index]) {
					cell = cellLines[index][lineIndex]
				}
				parts[index] = padRight(cell, widths[index])
			}
			rendered = append(rendered, strings.Join(parts, " | "))
		}
		return trimTrailingLineSpaces(strings.Join(rendered, "\n"))
	}

	var builder strings.Builder
	builder.WriteString(renderRow(headers))
	builder.WriteString("\n\n")
	for index, row := range rows {
		if index > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(renderRow(row))
	}
	builder.WriteString("\n")
	return builder.String()
}

func renderPersistenceSectionTable(headers []string, rows [][]string) string {
	return strings.TrimRight(renderStructuredTableWithTitle("", headers, rows, false), "\n")
}

func padRight(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-len(value))
}

func persistenceAutomationExplanation(account models.PersistenceAutomationAccount) string {
	lines := []string{"- " + persistenceAutomationAccountBullet(account)}
	if persistenceCapabilityStatus(account.CapabilitySteps, "create or modify account") != "yes" {
		return persistenceTruncatedWalkthrough(lines, []string{"  " + persistenceAutomationVisibilityLine(account)}, account.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, "  The Automation account is the Azure-side container for runbooks, schedules, webhooks, identity, and secure assets; no VM or host login is required to keep this path in Azure.")

	lines = append(lines, "- "+persistenceAutomationRunbookBullet(account))
	if persistenceCapabilityStatus(account.CapabilitySteps, "add or edit runbook") != "yes" {
		return persistenceTruncatedWalkthrough(lines, []string{"  " + persistenceAutomationVisibilityLine(account)}, account.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, "  A runbook is the stored container first; it becomes useful execution only after content is added and a published version exists.")

	lines = append(lines, "- "+persistenceAutomationCodeBullet(account))
	if persistenceCapabilityStatus(account.CapabilitySteps, "upload or replace code") != "yes" {
		return persistenceTruncatedWalkthrough(lines, []string{"  " + persistenceAutomationVisibilityLine(account)}, account.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, "  This is the runnable content layer: PowerShell or Python runbook content can call Azure APIs, reach storage or Key Vault, make outbound calls, or drive host actions through control-plane paths.")

	lines = append(lines, "- "+persistenceAutomationPublishBullet(account))
	if persistenceCapabilityStatus(account.CapabilitySteps, "publish runbook") != "yes" {
		return persistenceTruncatedWalkthrough(lines, []string{"  " + persistenceAutomationVisibilityLine(account)}, account.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, "  Automation keeps draft and published runbook versions; publishing is the step that makes the stored content runnable in Azure.")

	lines = append(lines, "- "+persistenceAutomationExecutionContextBullet(account))
	if persistenceCapabilityStatus(account.CapabilitySteps, "attach or reuse exec ctx") != "yes" {
		return persistenceTruncatedWalkthrough(lines, []string{"  " + persistenceAutomationVisibilityLine(account)}, account.CurrentState.NearbyThematicNames)
	}

	lines = append(lines, "  Managed identity, stored credentials, connections, certificates, variables, or other Automation assets may provide that execution context.")
	if account.CurrentIdentityContext != nil && strings.TrimSpace(account.CurrentIdentityContext.Summary) != "" {
		lines = append(lines, "  "+account.CurrentIdentityContext.Summary)
	}
	if ctx := account.CurrentState.StrongestVisibleExecutionContext; ctx != nil && strings.TrimSpace(ctx.Summary) != "" {
		lines = append(lines, "  "+ctx.Summary)
	}
	lines = append(lines, "- "+persistenceAutomationTriggerBullet(account))
	if persistenceCapabilityStatus(account.CapabilitySteps, "create schedule") != "yes" ||
		persistenceCapabilityStatus(account.CapabilitySteps, "link schedule to runbook") != "yes" ||
		persistenceCapabilityStatus(account.CapabilitySteps, "create webhook") != "yes" {
		return persistenceTruncatedWalkthrough(lines, []string{"  " + persistenceAutomationVisibilityLine(account)}, account.CurrentState.NearbyThematicNames)
	}
	if len(account.CurrentState.ScheduleDefinitions) > 0 {
		lines = append(lines, "  Visible schedule definitions here include "+persistenceAutomationScheduleDefinitionSummary(account.CurrentState.ScheduleDefinitions)+".")
	}
	lines = append(lines, "  Schedules, job schedules, webhooks, or upstream services such as Logic Apps and Functions are the durable rerun anchors; a runbook without one is stored code but not a complete persistence path.")
	lines = append(lines, "- "+persistenceAutomationRepurposeBullet(account))
	lines = append(lines, "  When triggered, Azure spins up a worker, loads the published runbook, executes under the selected identity or credential context, and then stops; persistence is the code, identity, and trigger remaining configured.")
	if nearby := persistenceAutomationNearbyNamesLine(account.CurrentState.NearbyThematicNames); nearby != "" {
		lines = append(lines, "  "+nearby)
	}
	return renderPersistenceWalkthrough(lines)
}

func persistenceAutomationScheduleDefinitionSummary(definitions []string) string {
	if len(definitions) <= 3 {
		return renderNaturalJoin(quoteInlineValues(definitions))
	}
	visible := quoteInlineValues(definitions[:3])
	return renderNaturalJoin(visible) + fmt.Sprintf(", plus %d more", len(definitions)-3)
}

func persistenceAutomationAccountBullet(account models.PersistenceAutomationAccount) string {
	switch persistenceCapabilityStatus(account.CapabilitySteps, "create or modify account") {
	case "yes":
		return "Current identity can create or modify an Azure Automation Account."
	default:
		return "Current identity does not yet have a proven path to create or modify this Azure Automation Account."
	}
}

func persistenceAutomationRunbookBullet(account models.PersistenceAutomationAccount) string {
	switch persistenceCapabilityStatus(account.CapabilitySteps, "add or edit runbook") {
	case "yes":
		return "Current identity can add or edit a runbook inside an existing Azure Automation Account."
	default:
		return "Current identity does not yet have a proven path to add or edit a runbook inside this Azure Automation Account."
	}
}

func persistenceAutomationCodeBullet(account models.PersistenceAutomationAccount) string {
	switch persistenceCapabilityStatus(account.CapabilitySteps, "upload or replace code") {
	case "yes":
		return "Current identity can upload or replace the code inside a runbook."
	default:
		return "Current identity does not yet have a proven path to upload or replace code inside a runbook."
	}
}

func persistenceAutomationPublishBullet(account models.PersistenceAutomationAccount) string {
	switch persistenceCapabilityStatus(account.CapabilitySteps, "publish runbook") {
	case "yes":
		return "Current identity can publish runnable automation so Azure can execute it later."
	default:
		return "Current identity does not yet have a proven path to publish runnable automation here."
	}
}

func persistenceAutomationExecutionContextBullet(account models.PersistenceAutomationAccount) string {
	switch persistenceCapabilityStatus(account.CapabilitySteps, "attach or reuse exec ctx") {
	case "yes":
		return "Current identity can attach or reuse execution context for that runbook."
	default:
		return "Current identity does not yet have a proven path to attach or reuse execution context for that runbook."
	}
}

func persistenceAutomationTriggerBullet(account models.PersistenceAutomationAccount) string {
	statuses := []string{
		persistenceCapabilityStatus(account.CapabilitySteps, "create schedule"),
		persistenceCapabilityStatus(account.CapabilitySteps, "link schedule to runbook"),
		persistenceCapabilityStatus(account.CapabilitySteps, "create webhook"),
	}
	for _, status := range statuses {
		if status != "yes" {
			return "Current identity does not yet have a proven path to create durable triggers for the runbook, including schedules, schedule links, and webhooks."
		}
	}
	return "Current identity can create durable triggers for the runbook, including schedules, schedule links, and webhooks."
}

func persistenceAutomationRepurposeBullet(account models.PersistenceAutomationAccount) string {
	switch persistenceCapabilityStatus(account.CapabilitySteps, "create or modify account") {
	case "yes":
		return "Current identity can repurpose an existing Azure Automation Account instead of creating a new one."
	default:
		return "Current identity does not yet have a proven path to repurpose this Azure Automation Account."
	}
}

func persistenceAutomationVisibilityLine(account models.PersistenceAutomationAccount) string {
	state := strings.TrimSpace(persistenceAutomationInventoryState(account))
	execution := strings.TrimSpace(persistenceAutomationInventoryExecutionContext(account))
	switch {
	case state != "" && state != "none visible" && execution != "" && execution != "none visible":
		return "Visibility still shows " + state + " with execution context " + execution + "; that is enough to judge whether this account already has runnable automation, trigger posture, or reuse value if stronger control is obtained later."
	case state != "" && state != "none visible":
		return "Visibility still shows " + state + "; that is enough to judge whether this account already has runnable automation, trigger posture, or reuse value if stronger control is obtained later."
	case execution != "" && execution != "none visible":
		return "Visibility still shows execution context " + execution + "; that is enough to judge whether this account is worth revisiting if stronger control is obtained later."
	default:
		return "Visibility still confirms this Automation account exists, even though the current identity does not yet have a proven write path here."
	}
}

func persistenceAutomationNearbyNamesLine(names []string) string {
	if len(names) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, "`"+name+"`")
	}
	return "Nearby maintenance- or schedule-themed names visible from the current environment include " + renderNaturalJoin(quoted) + "."
}

func persistenceCapabilityStatus(steps []models.PersistenceCapabilityStep, action string) string {
	for _, step := range steps {
		if step.Action == action {
			return step.Status
		}
	}
	return "not proven"
}

func persistenceLogicAppExplanation(workflow models.PersistenceLogicAppWorkflow) string {
	visibilityLines := persistenceLogicAppVisibilityLines(workflow)
	lines := []string{"- " + persistenceLogicAppWorkflowBullet(workflow)}
	if persistenceCapabilityStatus(workflow.CapabilitySteps, "create or modify workflow") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workflow.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, "  A Logic App is a workflow resource stored in Azure: the trigger starts it, and the actions decide what it does next.")

	lines = append(lines, "- "+persistenceLogicAppDefinitionBullet(workflow))
	if persistenceCapabilityStatus(workflow.CapabilitySteps, "edit workflow definition") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workflow.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, "  In Logic Apps, the payload is the stored workflow definition and action chain Azure will execute later.")
	lines = append(lines, "  Consumption-style workflows are managed directly from the workflow definition; Standard Logic Apps behave more like a host with workflows, app settings, and package or deployment paths inside it.")

	lines = append(lines, "- "+persistenceLogicAppExecutionContextBullet(workflow))
	if persistenceCapabilityStatus(workflow.CapabilitySteps, "attach or reuse exec ctx") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workflow.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, "  Managed identity or connector-backed actions may provide that execution context.")
	lines = append(lines, "  That identity or connection is the power layer: it determines which Azure services, secrets, storage paths, external endpoints, or other automation the workflow can reach.")
	if workflow.CurrentIdentityContext != nil && strings.TrimSpace(workflow.CurrentIdentityContext.Summary) != "" {
		lines = append(lines, "  "+workflow.CurrentIdentityContext.Summary)
	}
	if ctx := workflow.CurrentState.StrongestVisibleExecutionContext; ctx != nil && strings.TrimSpace(ctx.Summary) != "" {
		lines = append(lines, "  "+ctx.Summary)
	}

	lines = append(lines, "- "+persistenceLogicAppTriggerBullet(workflow))
	if persistenceCapabilityStatus(workflow.CapabilitySteps, "define or modify trigger") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workflow.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceLogicAppTriggerWalkthrough(workflow)...)

	lines = append(lines, "- "+persistenceLogicAppEnableBullet(workflow))
	if persistenceCapabilityStatus(workflow.CapabilitySteps, "enable workflow") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workflow.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, "  Once saved and enabled, Azure listens for the trigger and starts the workflow when the trigger fires; no user needs to stay logged in.")

	lines = append(lines, "- "+persistenceLogicAppActionBullet(workflow))
	if persistenceCapabilityStatus(workflow.CapabilitySteps, "add or repurpose downstream actions") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workflow.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceLogicAppActionWalkthrough(workflow)...)

	lines = append(lines, "- "+persistenceLogicAppRepurposeBullet(workflow))
	lines = append(lines, "  Persistence here is the stored workflow, reachable trigger, and valid identity or connector context remaining in Azure so the path can be reused later.")
	if nearby := persistenceAutomationNearbyNamesLine(workflow.CurrentState.NearbyThematicNames); nearby != "" {
		lines = append(lines, "  "+nearby)
	}
	return renderPersistenceWalkthrough(lines)
}

func persistenceLogicAppTriggerWalkthrough(workflow models.PersistenceLogicAppWorkflow) []string {
	lines := []string{}
	if len(workflow.CurrentState.TriggerTypes) > 0 {
		lines = append(lines, "  Visible trigger types here include "+persistenceJoinedOrNone(workflow.CurrentState.TriggerTypes)+".")
	}
	if workflow.CurrentState.ExternallyCallableRequestTrigger {
		lines = append(lines, "  The visible request trigger makes this workflow externally callable if the callback URL or caller path is usable; this command does not print trigger secret material.")
	}
	if recurrence := strings.TrimSpace(valueOrEmpty(workflow.CurrentState.RecurrenceSummary)); recurrence != "" {
		lines = append(lines, "  Visible recurrence posture here is "+recurrence+".")
	}
	if len(lines) == 0 {
		lines = append(lines, "  Trigger posture is the re-entry anchor: HTTP request, schedule, connector, or event triggers decide how this workflow can run again later.")
	}
	return lines
}

func persistenceLogicAppActionWalkthrough(workflow models.PersistenceLogicAppWorkflow) []string {
	lines := []string{
		"  Logic Apps do not need a traditional script to be useful; the action graph is the execution logic.",
	}
	if len(workflow.CurrentState.DownstreamActionKinds) > 0 {
		lines = append(lines, "  Visible downstream action kinds here include "+persistenceJoinedOrNone(workflow.CurrentState.DownstreamActionKinds)+".")
	}
	lines = append(lines, "  Actions can call Azure APIs, send HTTP requests, read or write storage, invoke other automation, or branch through connector-backed workflows when those mechanics are present.")
	return lines
}

func persistenceLogicAppVisibilityLines(workflow models.PersistenceLogicAppWorkflow) []string {
	return persistenceVisibilityFallbackLines(
		strings.TrimSpace(persistenceLogicAppInventoryState(workflow)),
		strings.TrimSpace(persistenceLogicAppInventoryExecutionContext(workflow)),
		"this workflow already has trigger posture, downstream action shape, or reuse value if stronger control is obtained later.",
		"this workflow is worth revisiting if stronger control is obtained later.",
		"  Visibility still confirms this Logic App exists, even though the current identity does not yet have a proven write path here.",
	)
}

func persistenceFunctionsExplanation(app models.PersistenceFunctionApp) string {
	lines := []string{"- " + persistenceFunctionsHostBullet(app)}
	if persistenceCapabilityStatus(app.CapabilitySteps, "create or modify function app") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceFunctionsVisibilityLines(app), app.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceFunctionsHostWalkthrough(app)...)

	lines = append(lines, "- "+persistenceFunctionsCodeBullet(app))
	if persistenceCapabilityStatus(app.CapabilitySteps, "deploy or replace code") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceFunctionsVisibilityLines(app), app.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceFunctionsCodeWalkthrough(app)...)

	lines = append(lines, "- "+persistenceFunctionsTriggerBullet(app))
	if persistenceCapabilityStatus(app.CapabilitySteps, "repurpose trigger posture") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceFunctionsVisibilityLines(app), app.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceFunctionsTriggerWalkthrough(app)...)

	lines = append(lines, "- "+persistenceFunctionsConfigBullet(app))
	if persistenceCapabilityStatus(app.CapabilitySteps, "change app settings or deployment config") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceFunctionsVisibilityLines(app), app.CurrentState.NearbyThematicNames)
	}
	lines = append(lines,
		"  This is where runtime behavior, package behavior, and connection material get shaped for this Function App.",
		"  App settings can control service endpoints, feature toggles, and other runtime behavior the function will use when the trigger fires.",
	)

	lines = append(lines, "- "+persistenceFunctionsExecutionContextBullet(app))
	if persistenceCapabilityStatus(app.CapabilitySteps, "attach or reuse exec ctx") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceFunctionsVisibilityLines(app), app.CurrentState.NearbyThematicNames)
	}
	lines = append(lines,
		"  In Functions, execution context usually comes from the managed identity attached to this Function App.",
		"  That can be a system-assigned identity, a user-assigned identity, or connection material referenced through settings the code will use later.",
	)
	if app.CurrentIdentityContext != nil && strings.TrimSpace(app.CurrentIdentityContext.Summary) != "" {
		lines = append(lines, "  "+app.CurrentIdentityContext.Summary)
	}
	if ctx := app.CurrentState.StrongestVisibleExecutionContext; ctx != nil && strings.TrimSpace(ctx.Summary) != "" {
		lines = append(lines, "  "+ctx.Summary)
	}

	lines = append(lines, "- "+persistenceFunctionsEnableBullet(app))
	if persistenceCapabilityStatus(app.CapabilitySteps, "restart or enable function host") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceFunctionsVisibilityLines(app), app.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceFunctionsEnableWalkthrough(app)...)

	lines = append(lines, "- "+persistenceFunctionsRepurposeBullet(app))
	lines = append(lines, persistenceFunctionsRepurposeWalkthrough(app)...)
	if nearby := persistenceAutomationNearbyNamesLine(app.CurrentState.NearbyThematicNames); nearby != "" {
		lines = append(lines, "  "+nearby)
	}
	return renderPersistenceWalkthrough(lines)
}

func persistenceAppServiceExplanation(app models.PersistenceAppService) string {
	lines := []string{"- " + persistenceAppServiceHostBullet(app)}
	if persistenceCapabilityStatus(app.CapabilitySteps, "create or reuse app service") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceAppServiceVisibilityLines(app), app.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAppServiceHostWalkthrough(app)...)

	lines = append(lines, "- "+persistenceAppServiceDeploymentPathBullet(app))
	if persistenceCapabilityStatus(app.CapabilitySteps, "set or reuse deployment path") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceAppServiceVisibilityLines(app), app.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAppServiceDeploymentPathWalkthrough(app)...)

	lines = append(lines, "- "+persistenceAppServiceConfigBullet(app))
	if persistenceCapabilityStatus(app.CapabilitySteps, "change app settings or identity attachment") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceAppServiceVisibilityLines(app), app.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAppServiceConfigWalkthrough(app)...)

	lines = append(lines, "- "+persistenceAppServiceCodeBullet(app))
	if persistenceCapabilityStatus(app.CapabilitySteps, "deploy or replace application code") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceAppServiceVisibilityLines(app), app.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAppServiceCodeWalkthrough(app)...)

	lines = append(lines, "- "+persistenceAppServiceExposureBullet(app))
	if persistenceCapabilityStatus(app.CapabilitySteps, "expose or reuse HTTP/HTTPS entry path") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceAppServiceVisibilityLines(app), app.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAppServiceExposureWalkthrough(app)...)

	lines = append(lines,
		"- This App Service view stops at the main web host; use `persistence webjobs` when you need App Service WebJobs background-execution depth.",
		"- Because the app stays deployed, reachable, and trusted until changed, this remains reusable App Service persistence even after the original session is gone.",
	)
	if nearby := persistenceAutomationNearbyNamesLine(app.CurrentState.NearbyThematicNames); nearby != "" {
		lines = append(lines, "  "+nearby)
	}
	return renderPersistenceWalkthrough(lines)
}

func persistenceWebJobsExplanation(job models.PersistenceWebJob) string {
	lines := []string{"- " + persistenceWebJobsHostBullet(job)}
	if persistenceCapabilityStatus(job.CapabilitySteps, "create or reuse parent app service") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceWebJobsVisibilityLines(job), job.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceWebJobsHostWalkthrough(job)...)

	lines = append(lines, "- "+persistenceWebJobsPackageBullet(job))
	if persistenceCapabilityStatus(job.CapabilitySteps, "add or replace webjob package") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceWebJobsVisibilityLines(job), job.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceWebJobsPackageWalkthrough(job)...)

	lines = append(lines, "- "+persistenceWebJobsModeBullet(job))
	if persistenceCapabilityStatus(job.CapabilitySteps, "set or reuse webjob mode") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceWebJobsVisibilityLines(job), job.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceWebJobsModeWalkthrough(job)...)

	lines = append(lines, "- "+persistenceWebJobsExecutionContextBullet(job))
	if persistenceCapabilityStatus(job.CapabilitySteps, "reuse inherited app execution context") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceWebJobsVisibilityLines(job), job.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceWebJobsExecutionContextWalkthrough(job)...)

	lines = append(lines, "- "+persistenceWebJobsRerunBullet(job))
	if persistenceCapabilityStatus(job.CapabilitySteps, "leave or repurpose rerun path") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceWebJobsVisibilityLines(job), job.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceWebJobsRerunWalkthrough(job)...)

	lines = append(lines, "- "+persistenceWebJobsRepurposeBullet(job))
	lines = append(lines, persistenceWebJobsRepurposeWalkthrough(job)...)
	if nearby := persistenceWebJobNearbyNamesLine(job.CurrentState.NearbyThematicNames); nearby != "" {
		lines = append(lines, "  "+nearby)
	}
	return renderPersistenceWalkthrough(lines)
}

func persistenceTruncatedWalkthrough(lines []string, visibilityLines []string, nearbyNames []string) string {
	for _, line := range visibilityLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}
	lines = append(lines, "  Higher permissions are required to complete the remaining persistence steps for this path.")
	if nearby := persistenceAutomationNearbyNamesLine(nearbyNames); nearby != "" {
		lines = append(lines, "  "+nearby)
	}
	return renderPersistenceWalkthrough(lines)
}

func persistenceVisibilityFallbackLines(
	state string,
	execution string,
	stateCapability string,
	revisitCapability string,
	defaultLine string,
) []string {
	switch {
	case state != "" && state != "none visible" && execution != "" && execution != "none visible":
		return []string{
			"  Visibility still shows " + state + ".",
			"  Visible execution context here is " + execution + ".",
			"  That is enough to judge whether " + stateCapability,
		}
	case state != "" && state != "none visible":
		return []string{
			"  Visibility still shows " + state + ".",
			"  That is enough to judge whether " + stateCapability,
		}
	case execution != "" && execution != "none visible":
		return []string{
			"  Visible execution context here is " + execution + ".",
			"  That is enough to judge whether " + revisitCapability,
		}
	default:
		return []string{defaultLine}
	}
}

func renderPersistenceWalkthrough(lines []string) string {
	formatted := make([]string, 0, len(lines)+8)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "- ") {
			if len(formatted) > 0 {
				formatted = append(formatted, "")
			}
			formatted = append(formatted, line)
			continue
		}
		if strings.HasPrefix(line, "  ") {
			formatted = append(formatted, "   "+strings.TrimPrefix(line, "  "))
			continue
		}
		formatted = append(formatted, line)
	}
	return strings.Join(formatted, "\n")
}

func persistenceFunctionsHostBullet(app models.PersistenceFunctionApp) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "create or modify function app") {
	case "yes":
		return "Current identity can create a new Function App or reuse this existing Function App."
	default:
		return "Current identity does not yet have a proven path to create a new Function App or reuse this existing Function App."
	}
}

func persistenceFunctionsHostWalkthrough(app models.PersistenceFunctionApp) []string {
	return []string{
		"  The Function App is the Azure-side host for function code, app settings, identity, trigger configuration, and the deployed package.",
	}
}

func persistenceFunctionsCodeBullet(app models.PersistenceFunctionApp) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "deploy or replace code") {
	case "yes":
		return "Current identity can deploy or replace the function package Azure will load in this Function App."
	default:
		return "Current identity does not yet have a proven path to deploy or replace the function package Azure would load in this Function App."
	}
}

func persistenceFunctionsCodeWalkthrough(app models.PersistenceFunctionApp) []string {
	if persistenceCapabilityStatus(app.CapabilitySteps, "deploy or replace code") != "yes" {
		return []string{"  The current read path does not yet tie Function App control to a defended deploy path here."}
	}

	lines := []string{
		"  Because the current identity already controls this Function App, zip deploy, publish, or package replacement are part of the defended Functions persistence path here.",
		"  Common deploy paths here include ZIP package deployment, pipeline deployment, run-from-package, or local project publish.",
		"  The Function App can exist without meaningful deployed logic; the package or project is the runnable payload Azure loads when a trigger fires.",
	}
	if deployment := strings.TrimSpace(valueOrEmpty(app.CurrentState.Deployment)); deployment != "" {
		for _, item := range strings.Split(deployment, ";") {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" {
				continue
			}
			lines = append(lines, "  Visible deployment posture includes "+trimmed+".")
		}
	}
	if app.CurrentState.RunFromPackage != nil && *app.CurrentState.RunFromPackage {
		lines = append(lines, "  Run-from-package is already enabled.")
	}
	if value := strings.TrimSpace(valueOrEmpty(app.CurrentState.AzureWebJobsStorageValueType)); value != "" {
		lines = append(lines, "  AzureWebJobsStorage is "+value+".")
	}
	return lines
}

func persistenceFunctionsExecutionContextBullet(app models.PersistenceFunctionApp) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "attach or reuse exec ctx") {
	case "yes":
		return "Current identity can attach or reuse execution context for this Function App."
	default:
		return "Current identity does not yet have a proven path to attach or reuse execution context for this Function App."
	}
}

func persistenceFunctionsConfigBullet(app models.PersistenceFunctionApp) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "change app settings or deployment config") {
	case "yes":
		return "Current identity can change app settings, identity attachment, and deployment configuration for this Function App."
	default:
		return "Current identity does not yet have a proven path to change app settings, identity attachment, or deployment configuration for this Function App."
	}
}

func persistenceFunctionsTriggerBullet(app models.PersistenceFunctionApp) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "repurpose trigger posture") {
	case "yes":
		if strings.TrimSpace(valueOrEmpty(app.CurrentState.Hostname)) != "" || strings.EqualFold(valueOrEmpty(app.CurrentState.PublicNetworkAccess), "Enabled") {
			return "Current identity can repurpose this Function App's trigger exposure so Azure has a way to run it again later, including HTTP-style externally reachable entrypoints."
		}
		return "Current identity can repurpose this Function App's trigger posture so Azure has a way to run it again later."
	default:
		return "Current identity does not yet have a proven path to repurpose this Function App's trigger posture."
	}
}

func persistenceFunctionsTriggerWalkthrough(app models.PersistenceFunctionApp) []string {
	if len(app.CurrentState.TriggerTypes) == 0 && len(app.CurrentState.VisibleFunctions) == 0 {
		return []string{
			"  Common trigger paths here include HTTP, timer, queue, Service Bus, or storage/event-driven execution.",
			"  The current collector now asks Azure for child functions, but no per-function trigger detail was confirmed from the current read path here.",
		}
	}

	lines := []string{}
	if len(app.CurrentState.TriggerTypes) > 0 {
		lines = append(lines, "  Visible child functions here show "+persistenceJoinedOrNone(app.CurrentState.TriggerTypes)+" trigger paths.")
	}
	if len(app.CurrentState.VisibleFunctions) > 0 {
		lines = append(lines, "  Current visible functions include "+persistenceFunctionsVisibleFunctionSummary(app.CurrentState.VisibleFunctions)+".")
	}
	if detail := persistenceFunctionsTriggerBoundaryLine(app.CurrentState.VisibleFunctions); detail != "" {
		for _, line := range strings.Split(detail, "\n") {
			lines = append(lines, line)
		}
	}
	lines = append(lines, "  The remaining gap is data-plane and runtime-side validation the current management-plane collector does not perform.")
	lines = append(lines, "  That includes function keys or caller auth actually in hand, upstream Service Bus, queue, storage, or binding access, and any runtime-side restriction beyond the visible trigger metadata.")
	return lines
}

func persistenceFunctionsEnableBullet(app models.PersistenceFunctionApp) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "restart or enable function host") {
	case "yes":
		return "Current identity can restart or enable the Function App so Azure is ready to run it again when the trigger path is hit."
	default:
		return "Current identity does not yet have a proven path to restart or enable this Function App for later trigger-driven execution."
	}
}

func persistenceFunctionsEnableWalkthrough(app models.PersistenceFunctionApp) []string {
	return []string{
		"  Once the package is in place and the Function App is active, Azure can invoke it whenever the trigger condition is met.",
	}
}

func persistenceFunctionsRepurposeBullet(app models.PersistenceFunctionApp) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "create or modify function app") {
	case "yes":
		return "Current identity can repurpose an existing Function App instead of creating a new one."
	default:
		return "Current identity does not yet have a proven path to repurpose this existing Function App."
	}
}

func persistenceFunctionsRepurposeWalkthrough(app models.PersistenceFunctionApp) []string {
	return []string{
		"  Reusing an existing Function App can blend in better than standing up a brand-new serverless entrypoint.",
	}
}

func persistenceFunctionsVisibilityLines(app models.PersistenceFunctionApp) []string {
	state := strings.TrimSpace(persistenceFunctionsInventoryState(app))
	execution := strings.TrimSpace(persistenceFunctionsInventoryExecutionContext(app))
	return persistenceVisibilityFallbackLines(
		state,
		execution,
		"this Function App already has trigger exposure, deployment signals, or reuse value if stronger control is obtained later.",
		"this Function App is worth revisiting if stronger control is obtained later.",
		"  Visibility still confirms this Function App exists, even though the current identity does not yet have a proven write path here.",
	)
}

func persistenceAppServiceHostBullet(app models.PersistenceAppService) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "create or reuse app service") {
	case "yes":
		return "Current identity can create a new App Service app or reuse this existing app."
	default:
		return "Current identity does not yet have a proven path to create a new App Service app or reuse this existing app."
	}
}

func persistenceAppServiceHostWalkthrough(app models.PersistenceAppService) []string {
	return []string{
		"  The App Service app is the Azure-side host that keeps the web workload deployed, configurable, identity-backed, and reachable over time.",
	}
}

func persistenceAppServiceDeploymentPathBullet(app models.PersistenceAppService) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "set or reuse deployment path") {
	case "yes":
		return "Current identity can set or reuse the deployment path this App Service will load."
	default:
		return "Current identity does not yet have a proven path to set or reuse the deployment path this App Service will load."
	}
}

func persistenceAppServiceDeploymentPathWalkthrough(app models.PersistenceAppService) []string {
	lines := []string{
		"  This is distinct from the application content itself: it is the trust path Azure uses to decide where code arrives from.",
	}
	if deployment := strings.TrimSpace(valueOrEmpty(app.CurrentState.Deployment)); deployment != "" {
		lines = append(lines, "  Visible deployment signals here include "+deployment+".")
	}
	return lines
}

func persistenceAppServiceConfigBullet(app models.PersistenceAppService) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "change app settings or identity attachment") {
	case "yes":
		return "Current identity can change app settings and attach or reuse managed identity for this App Service."
	default:
		return "Current identity does not yet have a proven path to change app settings or attach or reuse managed identity for this App Service."
	}
}

func persistenceAppServiceConfigWalkthrough(app models.PersistenceAppService) []string {
	lines := []string{
		"  This is the execution-power layer for App Service: settings, connection strings, and workload identity shape what the deployed app can reach later.",
	}
	configParts := []string{}
	if count := intPtrString(app.CurrentState.AppSettingsCount); count != "" {
		configParts = append(configParts, count+" app setting(s)")
	}
	if count := intPtrString(app.CurrentState.KeyVaultReferenceCount); count != "" && count != "0" {
		configParts = append(configParts, count+" Key Vault-backed setting(s)")
	}
	if count := intPtrString(app.CurrentState.SensitiveSettingCount); count != "" && count != "0" {
		configParts = append(configParts, count+" sensitive-looking setting name(s)")
	}
	if count := intPtrString(app.CurrentState.ConnectionStringCount); count != "" && count != "0" {
		configParts = append(configParts, count+" connection string(s)")
	}
	if count := intPtrString(app.CurrentState.KeyVaultConnectionStringCount); count != "" && count != "0" {
		configParts = append(configParts, count+" Key Vault-backed connection string(s)")
	}
	if len(app.CurrentState.ConnectionStringTypes) > 0 {
		configParts = append(configParts, "connection types "+strings.Join(app.CurrentState.ConnectionStringTypes, ", "))
	}
	if len(configParts) > 0 {
		lines = append(lines, "  Visible config posture here includes "+strings.Join(configParts, ", ")+".")
	}
	if len(app.CurrentState.VisibleSensitiveSettingNames) > 0 {
		quoted := make([]string, 0, len(app.CurrentState.VisibleSensitiveSettingNames))
		for _, item := range app.CurrentState.VisibleSensitiveSettingNames {
			quoted = append(quoted, "`"+item+"`")
		}
		lines = append(lines, "  Visible sensitive-looking setting names include "+renderNaturalJoin(quoted)+".")
	}
	if app.CurrentIdentityContext != nil && strings.TrimSpace(app.CurrentIdentityContext.Summary) != "" {
		lines = append(lines, "  "+app.CurrentIdentityContext.Summary)
	}
	if ctx := app.CurrentState.StrongestVisibleExecutionContext; ctx != nil && strings.TrimSpace(ctx.Summary) != "" {
		lines = append(lines, "  "+ctx.Summary)
	}
	return lines
}

func persistenceAppServiceCodeBullet(app models.PersistenceAppService) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "deploy or replace application code") {
	case "yes":
		return "Current identity can deploy or replace the application code this App Service will run."
	default:
		return "Current identity does not yet have a proven path to deploy or replace the application code this App Service will run."
	}
}

func persistenceAppServiceCodeWalkthrough(app models.PersistenceAppService) []string {
	lines := []string{
		"  This is the runnable content layer, separate from the deployment path that feeds it.",
		"  Because the current identity already controls this App Service, package replacement, source-based deployment, or pipeline-backed redeploy are part of the defended App Service persistence path here.",
	}
	if app.CurrentState.RunFromPackage != nil {
		if *app.CurrentState.RunFromPackage {
			lines = append(lines, "  Run-from-package is already enabled.")
		} else {
			lines = append(lines, "  Run-from-package is not currently enabled.")
		}
	}
	return lines
}

func persistenceAppServiceExposureBullet(app models.PersistenceAppService) string {
	switch persistenceCapabilityStatus(app.CapabilitySteps, "expose or reuse HTTP/HTTPS entry path") {
	case "yes":
		return "Current identity can expose or reuse this app's HTTP or HTTPS entry path so it remains reachable later."
	default:
		return "Current identity does not yet have a proven path to expose or reuse this app's HTTP or HTTPS entry path."
	}
}

func persistenceAppServiceExposureWalkthrough(app models.PersistenceAppService) []string {
	lines := []string{
		"  App Service re-entry is usually the app's reachable HTTP or HTTPS path, not a background process on a VM.",
	}
	if hostname := strings.TrimSpace(valueOrEmpty(app.CurrentState.Hostname)); hostname != "" {
		lines = append(lines, "  Visible hostname here is `"+hostname+"`.")
	}
	postureParts := []string{}
	if network := strings.TrimSpace(valueOrEmpty(app.CurrentState.PublicNetworkAccess)); network != "" {
		postureParts = append(postureParts, "public network access "+network)
	}
	if app.CurrentState.HTTPSOnly != nil {
		if *app.CurrentState.HTTPSOnly {
			postureParts = append(postureParts, "HTTPS-only enabled")
		} else {
			postureParts = append(postureParts, "HTTPS-only disabled")
		}
	}
	if tls := strings.TrimSpace(valueOrEmpty(app.CurrentState.MinTLSVersion)); tls != "" {
		postureParts = append(postureParts, "TLS "+tls)
	}
	if len(postureParts) > 0 {
		lines = append(lines, "  Visible entry posture includes "+strings.Join(postureParts, ", ")+".")
	}
	return lines
}

func persistenceAppServiceVisibilityLines(app models.PersistenceAppService) []string {
	state := strings.TrimSpace(persistenceAppServiceInventoryState(app))
	execution := strings.TrimSpace(persistenceAppServiceInventoryExecutionContext(app))
	return persistenceVisibilityFallbackLines(
		state,
		execution,
		"this App Service already has deployment path, configuration power, or reuse value if stronger control is obtained later.",
		"this App Service is worth revisiting if stronger control is obtained later.",
		"  Visibility still confirms this App Service exists, even though the current identity does not yet have a proven write path here.",
	)
}

func persistenceWebJobsHostBullet(job models.PersistenceWebJob) string {
	switch persistenceCapabilityStatus(job.CapabilitySteps, "create or reuse parent app service") {
	case "yes":
		return "Current identity can create or reuse the parent App Service host that carries this WebJob."
	default:
		return "Current identity does not yet have a proven path to create or reuse the parent App Service host that carries this WebJob."
	}
}

func persistenceWebJobsHostWalkthrough(job models.PersistenceWebJob) []string {
	return []string{
		"  WebJobs run alongside the existing web, API, or mobile app path in that App Service instead of replacing the main endpoint.",
	}
}

func persistenceWebJobsPackageBullet(job models.PersistenceWebJob) string {
	switch persistenceCapabilityStatus(job.CapabilitySteps, "add or replace webjob package") {
	case "yes":
		return "Current identity can add or replace the WebJob script or executable this App Service keeps deployed."
	default:
		return "Current identity does not yet have a proven path to add or replace the WebJob script or executable this App Service keeps deployed."
	}
}

func persistenceWebJobsPackageWalkthrough(job models.PersistenceWebJob) []string {
	lines := []string{
		"  This is the durable WebJob content layer, separate from the parent app's main web route.",
	}
	if command := strings.TrimSpace(valueOrEmpty(job.CurrentState.RunCommand)); command != "" {
		lines = append(lines, "  Visible run command here is `"+command+"`.")
	}
	return lines
}

func persistenceWebJobsModeBullet(job models.PersistenceWebJob) string {
	modeLabel := persistenceWebJobModePhrase(job.CurrentState.Mode)
	switch persistenceCapabilityStatus(job.CapabilitySteps, "set or reuse webjob mode") {
	case "yes":
		return "Current identity can leave this WebJob on a " + modeLabel + " that Kudu and the App Service runtime can discover and run again later."
	default:
		return "Current identity does not yet have a proven path to leave this WebJob on a " + modeLabel + " the platform can discover and run again later."
	}
}

func persistenceWebJobsModeWalkthrough(job models.PersistenceWebJob) []string {
	lines := []string{
		"  This is the WebJobs-specific control point: the deployed job path plus its mode tells the platform how it comes back later.",
	}
	mode := strings.TrimSpace(job.CurrentState.Mode)
	if mode != "" {
		lines = append(lines, "  Visible mode here is "+mode+".")
	}
	if trigger := strings.TrimSpace(valueOrEmpty(job.CurrentState.LatestRunTrigger)); trigger != "" {
		lines = append(lines, "  Latest visible trigger here is "+trigger+".")
	}
	if schedule := strings.TrimSpace(valueOrEmpty(job.CurrentState.ScheduleExpression)); schedule != "" {
		lines = append(lines, "  Visible schedule expression here is `"+schedule+"`.")
	}
	if job.CurrentState.SchedulerLogsURL != nil && strings.TrimSpace(*job.CurrentState.SchedulerLogsURL) != "" {
		lines = append(lines, "  Scheduler log visibility is already present for this rerun path.")
	}
	lines = append(lines, "  Kudu and the App Service runtime discover the job from the deployed WebJobs path and run it again according to that mode.")
	return lines
}

func persistenceWebJobsExecutionContextBullet(job models.PersistenceWebJob) string {
	switch persistenceCapabilityStatus(job.CapabilitySteps, "reuse inherited app execution context") {
	case "yes":
		return "Current identity can reuse the parent App Service execution context this WebJob inherits."
	default:
		return "Current identity does not yet have a proven path to reuse the parent App Service execution context this WebJob inherits."
	}
}

func persistenceWebJobsExecutionContextWalkthrough(job models.PersistenceWebJob) []string {
	lines := []string{
		"  This is the inherited power layer for WebJobs: managed identity, app settings, and connection material come from the parent App Service.",
	}
	configParts := []string{}
	if count := intPtrString(job.CurrentState.ParentAppSettingsCount); count != "" && count != "0" {
		configParts = append(configParts, count+" parent app setting(s)")
	}
	if count := intPtrString(job.CurrentState.ParentKeyVaultReferenceCount); count != "" && count != "0" {
		configParts = append(configParts, count+" Key Vault-backed setting(s)")
	}
	if count := intPtrString(job.CurrentState.ParentConnectionStringCount); count != "" && count != "0" {
		configParts = append(configParts, count+" connection string(s)")
	}
	if len(configParts) > 0 {
		lines = append(lines, "  Visible inherited app context here includes "+strings.Join(configParts, ", ")+".")
	}
	if job.CurrentIdentityContext != nil && strings.TrimSpace(job.CurrentIdentityContext.Summary) != "" {
		lines = append(lines, "  "+job.CurrentIdentityContext.Summary)
	}
	if ctx := job.CurrentState.StrongestVisibleExecutionContext; ctx != nil && strings.TrimSpace(ctx.Summary) != "" {
		lines = append(lines, "  "+ctx.Summary)
	}
	return lines
}

func persistenceWebJobsRerunBullet(job models.PersistenceWebJob) string {
	switch persistenceCapabilityStatus(job.CapabilitySteps, "leave or repurpose rerun path") {
	case "yes":
		return "As long as the job stays deployed, that rerun path can bring it back later without relying on the main web endpoint."
	default:
		return "Current identity does not yet have a proven path to leave this rerun path in place for later WebJob execution."
	}
}

func persistenceWebJobsRerunWalkthrough(job models.PersistenceWebJob) []string {
	lines := []string{}
	switch strings.ToLower(strings.TrimSpace(job.CurrentState.Mode)) {
	case "continuous":
		lines = append(lines, "  Continuous mode means the job stays on a background execution path inside the parent app.")
	case "scheduled":
		lines = append(lines, "  Scheduled mode means the job comes back when that scheduled rerun path fires again.")
	case "triggered/manual":
		lines = append(lines, "  Triggered/manual mode means the job comes back through that trigger path instead of the main web route.")
	default:
		lines = append(lines, "  The rerun story here comes from the job mode and the platform path that discovers it again later.")
	}
	postureParts := []string{}
	if hostname := strings.TrimSpace(valueOrEmpty(job.CurrentState.ParentHostname)); hostname != "" {
		postureParts = append(postureParts, "parent hostname `"+hostname+"`")
	}
	if network := strings.TrimSpace(valueOrEmpty(job.CurrentState.ParentPublicNetworkAccess)); network != "" {
		postureParts = append(postureParts, "public network access "+network)
	}
	if len(postureParts) > 0 {
		lines = append(lines, "  Visible parent-app posture includes "+strings.Join(postureParts, ", ")+".")
	}
	return lines
}

func persistenceWebJobsRepurposeBullet(job models.PersistenceWebJob) string {
	switch persistenceCapabilityStatus(job.CapabilitySteps, "create or reuse parent app service") {
	case "yes":
		return "This remains reusable background persistence until the job package, mode, or parent app context is changed."
	default:
		return "This WebJob is still visible background execution, but the current identity does not yet have a proven path to repurpose it here."
	}
}

func persistenceWebJobsRepurposeWalkthrough(job models.PersistenceWebJob) []string {
	return []string{
		"  The operator story here is the parent App Service plus the deployed WebJob plus the rerun mode the platform will keep honoring later.",
	}
}

func persistenceWebJobsVisibilityLines(job models.PersistenceWebJob) []string {
	state := strings.TrimSpace(persistenceWebJobsInventoryState(job))
	execution := strings.TrimSpace(persistenceWebJobsInventoryExecutionContext(job))
	return persistenceVisibilityFallbackLines(
		state,
		execution,
		"this WebJob already has rerun posture, inherited app power, or reuse value if stronger control is obtained later.",
		"this WebJob is worth revisiting if stronger control is obtained later.",
		"  Visibility still confirms this WebJob exists, even though the current identity does not yet have a proven write path here.",
	)
}

func persistenceContainerAppsJobsExplanation(job models.PersistenceContainerAppsJob) string {
	lines := []string{"- " + persistenceContainerAppsJobsJobBullet(job)}
	if persistenceCapabilityStatus(job.CapabilitySteps, "create or reuse job in environment") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceContainerAppsJobsVisibilityLines(job), job.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceContainerAppsJobsJobWalkthrough(job)...)

	lines = append(lines, "- "+persistenceContainerAppsJobsPayloadBullet(job))
	if persistenceCapabilityStatus(job.CapabilitySteps, "point job at image or command") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceContainerAppsJobsVisibilityLines(job), job.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceContainerAppsJobsPayloadWalkthrough(job)...)

	lines = append(lines, "- "+persistenceContainerAppsJobsTriggerBullet(job))
	if persistenceCapabilityStatus(job.CapabilitySteps, "choose trigger mode") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceContainerAppsJobsVisibilityLines(job), job.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceContainerAppsJobsTriggerWalkthrough(job)...)

	lines = append(lines, "- "+persistenceContainerAppsJobsExecutionShapeBullet(job))
	if persistenceCapabilityStatus(job.CapabilitySteps, "set execution shape and access posture") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceContainerAppsJobsVisibilityLines(job), job.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceContainerAppsJobsExecutionShapeWalkthrough(job)...)

	lines = append(lines, "- "+persistenceContainerAppsJobsDeployBullet(job))
	if persistenceCapabilityStatus(job.CapabilitySteps, "deploy or update stored job definition") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceContainerAppsJobsVisibilityLines(job), job.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceContainerAppsJobsDeployWalkthrough(job)...)

	lines = append(lines, "- "+persistenceContainerAppsJobsRerunBullet(job))
	if persistenceCapabilityStatus(job.CapabilitySteps, "start or rely on later executions") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceContainerAppsJobsVisibilityLines(job), job.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceContainerAppsJobsRerunWalkthrough(job)...)

	lines = append(lines, "- "+persistenceContainerAppsJobsPreserveBullet(job))
	lines = append(lines, persistenceContainerAppsJobsPreserveWalkthrough(job)...)
	if nearby := persistenceContainerAppsJobNearbyNamesLine(job.CurrentState.NearbyThematicNames); nearby != "" {
		lines = append(lines, "  "+nearby)
	}
	return renderPersistenceWalkthrough(lines)
}

func persistenceContainerAppsJobsJobBullet(job models.PersistenceContainerAppsJob) string {
	switch persistenceCapabilityStatus(job.CapabilitySteps, "create or reuse job in environment") {
	case "yes":
		return "Current identity can create a new Container Apps job or reuse this existing job definition in its environment."
	default:
		return "Current identity does not yet have a proven path to create a new Container Apps job or reuse this existing job definition in its environment."
	}
}

func persistenceContainerAppsJobsJobWalkthrough(job models.PersistenceContainerAppsJob) []string {
	lines := []string{
		"  The job definition is the Azure-side object that ties together the environment, trigger, image, execution settings, and identity.",
	}
	if environment := strings.TrimSpace(valueOrEmpty(job.CurrentState.EnvironmentID)); environment != "" {
		lines = append(lines, "  Visible Container Apps environment here is `"+resourceNameFromDisplayID(environment)+"`.")
	}
	return lines
}

func persistenceContainerAppsJobsPayloadBullet(job models.PersistenceContainerAppsJob) string {
	switch persistenceCapabilityStatus(job.CapabilitySteps, "point job at image or command") {
	case "yes":
		return "Current identity can point this job at the container image and command Azure will execute for each run."
	default:
		return "Current identity does not yet have a proven path to point this job at the container image and command Azure would execute for each run."
	}
}

func persistenceContainerAppsJobsPayloadWalkthrough(job models.PersistenceContainerAppsJob) []string {
	lines := []string{
		"  This is the stored execution payload, separate from the trigger mode that brings the job back later.",
	}
	if len(job.CurrentState.ContainerImages) > 0 {
		lines = append(lines, "  Visible image clue(s) here include "+strings.Join(job.CurrentState.ContainerImages, ", ")+".")
	}
	if len(job.CurrentState.Command) > 0 {
		lines = append(lines, "  Visible command clue(s) here include `"+strings.Join(job.CurrentState.Command, "` and `")+"`.")
	}
	return lines
}

func persistenceContainerAppsJobsTriggerBullet(job models.PersistenceContainerAppsJob) string {
	trigger := persistenceContainerAppsJobsTriggerPhrase(job)
	switch persistenceCapabilityStatus(job.CapabilitySteps, "choose trigger mode") {
	case "yes":
		return "Current identity can leave this as a " + trigger + " Container Apps job."
	default:
		return "Current identity does not yet have a proven path to leave this as a " + trigger + " Container Apps job."
	}
}

func persistenceContainerAppsJobsTriggerWalkthrough(job models.PersistenceContainerAppsJob) []string {
	lines := []string{
		"  Trigger mode is the service-specific re-entry choice for Container Apps Jobs: manual start, cron schedule, or event scaler.",
	}
	if trigger := strings.TrimSpace(valueOrEmpty(job.CurrentState.TriggerType)); trigger != "" {
		lines = append(lines, "  Visible trigger type here is "+trigger+".")
	}
	if schedule := strings.TrimSpace(valueOrEmpty(job.CurrentState.ScheduleExpression)); schedule != "" {
		lines = append(lines, "  Visible schedule expression here is `"+schedule+"`.")
	}
	if len(job.CurrentState.EventRules) > 0 {
		lines = append(lines, "  Visible event rule posture here includes "+persistenceContainerAppsJobsEventRulesLine(job.CurrentState.EventRules)+".")
	}
	return lines
}

func persistenceContainerAppsJobsTriggerPhrase(job models.PersistenceContainerAppsJob) string {
	switch strings.ToLower(strings.TrimSpace(valueOrEmpty(job.CurrentState.TriggerType))) {
	case "schedule", "scheduled":
		if strings.TrimSpace(valueOrEmpty(job.CurrentState.ScheduleExpression)) != "" {
			return "scheduled"
		}
		return "scheduled"
	case "event", "event-driven":
		return "event-driven"
	case "manual":
		return "manual"
	default:
		return "manual, scheduled, or event-driven"
	}
}

func persistenceContainerAppsJobsEventRulesLine(rules []models.ContainerAppsJobEventRule) string {
	parts := make([]string, 0, len(rules))
	for _, rule := range rules {
		label := strings.TrimSpace(rule.Name)
		if label == "" {
			label = "unnamed rule"
		}
		if strings.TrimSpace(rule.Type) != "" {
			label += " (" + rule.Type + ")"
		}
		if len(rule.AuthSecretRefs) > 0 {
			label += fmt.Sprintf(" with %d auth secret reference(s)", len(rule.AuthSecretRefs))
		}
		if strings.TrimSpace(valueOrEmpty(rule.Identity)) != "" {
			label += " using identity " + resourceNameFromDisplayID(*rule.Identity)
		}
		parts = append(parts, label)
	}
	if len(parts) == 0 {
		return "no event rules"
	}
	return strings.Join(parts, "; ")
}

func persistenceContainerAppsJobsExecutionShapeBullet(job models.PersistenceContainerAppsJob) string {
	switch persistenceCapabilityStatus(job.CapabilitySteps, "set execution shape and access posture") {
	case "yes":
		return "Current identity can set the execution shape through retries, timeout, replica counts, secrets, registry posture, and attached identity."
	default:
		return "Current identity does not yet have a proven path to set execution shape, safe access posture, or attached identity for this job."
	}
}

func persistenceContainerAppsJobsExecutionShapeWalkthrough(job models.PersistenceContainerAppsJob) []string {
	lines := []string{
		"  This is the execution-power layer for the job: how many tasks run, how they retry, and what identity or secret-backed material they can use.",
	}
	executionParts := []string{}
	if count := intPtrString(job.CurrentState.Parallelism); count != "" {
		executionParts = append(executionParts, "parallelism="+count)
	}
	if count := intPtrString(job.CurrentState.ReplicaCompletionCount); count != "" {
		executionParts = append(executionParts, "completions="+count)
	}
	if count := intPtrString(job.CurrentState.ReplicaRetryLimit); count != "" {
		executionParts = append(executionParts, "retries="+count)
	}
	if count := intPtrString(job.CurrentState.ReplicaTimeout); count != "" {
		executionParts = append(executionParts, "timeout="+count+"s")
	}
	if len(executionParts) > 0 {
		lines = append(lines, "  Visible execution settings here include "+strings.Join(executionParts, ", ")+".")
	}
	postureParts := []string{}
	if count := intPtrString(job.CurrentState.SecretCount); count != "" {
		postureParts = append(postureParts, count+" secret reference(s)")
	}
	if count := intPtrString(job.CurrentState.KeyVaultSecretCount); count != "" && count != "0" {
		postureParts = append(postureParts, count+" Key Vault-backed secret reference(s)")
	}
	if len(job.CurrentState.RegistryServers) > 0 {
		postureParts = append(postureParts, "registry server(s) "+strings.Join(job.CurrentState.RegistryServers, ", "))
	}
	if count := intPtrString(job.CurrentState.RegistryIdentityCount); count != "" && count != "0" {
		postureParts = append(postureParts, count+" registry identity reference(s)")
	}
	if count := intPtrString(job.CurrentState.RegistryPasswordRefCount); count != "" && count != "0" {
		postureParts = append(postureParts, count+" registry password reference(s)")
	}
	if len(postureParts) > 0 {
		lines = append(lines, "  Safe access posture here includes "+strings.Join(postureParts, ", ")+".")
	}
	if job.CurrentIdentityContext != nil && strings.TrimSpace(job.CurrentIdentityContext.Summary) != "" {
		lines = append(lines, "  "+job.CurrentIdentityContext.Summary)
	}
	if ctx := job.CurrentState.StrongestVisibleExecutionContext; ctx != nil && strings.TrimSpace(ctx.Summary) != "" {
		lines = append(lines, "  "+ctx.Summary)
	}
	return lines
}

func persistenceContainerAppsJobsDeployBullet(job models.PersistenceContainerAppsJob) string {
	switch persistenceCapabilityStatus(job.CapabilitySteps, "deploy or update stored job definition") {
	case "yes":
		return "Current identity can deploy or update the job so the Container Apps environment retains the job definition, trigger, image, and execution settings."
	default:
		return "Current identity does not yet have a proven path to deploy or update the stored Container Apps job definition."
	}
}

func persistenceContainerAppsJobsDeployWalkthrough(job models.PersistenceContainerAppsJob) []string {
	return []string{
		"  This is the durable object layer: the stored Microsoft.App/jobs definition remains in the control plane until changed or removed.",
	}
}

func persistenceContainerAppsJobsRerunBullet(job models.PersistenceContainerAppsJob) string {
	switch persistenceCapabilityStatus(job.CapabilitySteps, "start or rely on later executions") {
	case "yes":
		return "Current identity can start the job manually, or rely on the configured schedule or event rule to create later executions from the same stored definition."
	default:
		return "Current identity does not yet have a proven path to start this job manually or rely on its later trigger-driven executions."
	}
}

func persistenceContainerAppsJobsRerunWalkthrough(job models.PersistenceContainerAppsJob) []string {
	lines := []string{}
	switch strings.ToLower(strings.TrimSpace(valueOrEmpty(job.CurrentState.TriggerType))) {
	case "schedule", "scheduled":
		lines = append(lines, "  Scheduled trigger mode means the stored job can come back when the cron expression fires again.")
	case "event", "event-driven":
		lines = append(lines, "  Event trigger mode means the stored job can come back when the configured scaler rule fires again.")
	case "manual":
		lines = append(lines, "  Manual trigger mode means the stored job can be started again without redefining the container payload.")
	default:
		lines = append(lines, "  The rerun story here comes from the stored trigger mode and the job definition Azure keeps for later executions.")
	}
	return lines
}

func persistenceContainerAppsJobsPreserveBullet(job models.PersistenceContainerAppsJob) string {
	switch persistenceCapabilityStatus(job.CapabilitySteps, "preserve or reuse execution path") {
	case "yes":
		return "Current identity can preserve or reuse a Container Apps Jobs execution path by keeping the stored job definition, trigger, image, and execution context in place."
	default:
		return "This Container Apps job is still visible as a stored execution path, but the current identity does not yet have a proven path to preserve or reuse it here."
	}
}

func persistenceContainerAppsJobsPreserveWalkthrough(job models.PersistenceContainerAppsJob) []string {
	return []string{
		"  The operator story here is the stored job definition plus trigger mode plus container image and execution context Azure can run again later.",
	}
}

func persistenceContainerAppsJobsVisibilityLines(job models.PersistenceContainerAppsJob) []string {
	state := strings.TrimSpace(persistenceContainerAppsJobsInventoryState(job))
	execution := strings.TrimSpace(persistenceContainerAppsJobsInventoryExecutionContext(job))
	return persistenceVisibilityFallbackLines(
		state,
		execution,
		"this Container Apps job already has trigger, image, execution-setting, or reuse value if stronger control is obtained later.",
		"this Container Apps job is worth revisiting if stronger control is obtained later.",
		"  Visibility still confirms this Container Apps job exists, even though the current identity does not yet have a proven write path here.",
	)
}

func persistenceContainerAppsJobNearbyNamesLine(names []string) string {
	if len(names) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, "`"+name+"`")
	}
	return "Nearby batch-, sync-, worker-, or runner-themed Container Apps job names visible from the current environment include " + renderNaturalJoin(quoted) + "."
}

func persistenceVMExtensionsExplanation(extension models.PersistenceVMExtension) string {
	lines := []string{"- " + persistenceVMExtensionsControlBullet(extension)}
	if persistenceCapabilityStatus(extension.CapabilitySteps, "modify VM extension configuration") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceVMExtensionsVisibilityLines(extension), extension.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceVMExtensionsControlWalkthrough(extension)...)

	lines = append(lines, "- "+persistenceVMExtensionsTargetBullet(extension))
	if persistenceCapabilityStatus(extension.CapabilitySteps, "reuse VM or VMSS target") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceVMExtensionsVisibilityLines(extension), extension.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceVMExtensionsTargetWalkthrough(extension)...)

	lines = append(lines, "- "+persistenceVMExtensionsAttachmentBullet(extension))
	if persistenceCapabilityStatus(extension.CapabilitySteps, "add or modify extension attachment") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceVMExtensionsVisibilityLines(extension), extension.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceVMExtensionsAttachmentWalkthrough(extension)...)

	lines = append(lines, "- "+persistenceVMExtensionsSourceBullet(extension))
	if persistenceCapabilityStatus(extension.CapabilitySteps, "provide script or command source") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceVMExtensionsVisibilityLines(extension), extension.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceVMExtensionsSourceWalkthrough(extension)...)

	lines = append(lines, "- "+persistenceVMExtensionsSettingsBullet(extension))
	if persistenceCapabilityStatus(extension.CapabilitySteps, "configure extension execution") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceVMExtensionsVisibilityLines(extension), extension.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceVMExtensionsSettingsWalkthrough(extension)...)

	lines = append(lines, "- "+persistenceVMExtensionsDeliveryBullet(extension))
	if persistenceCapabilityStatus(extension.CapabilitySteps, "deliver config to VM agent") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceVMExtensionsVisibilityLines(extension), extension.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceVMExtensionsDeliveryWalkthrough(extension)...)

	lines = append(lines, "- "+persistenceVMExtensionsGuestExecutionBullet(extension))
	if persistenceCapabilityStatus(extension.CapabilitySteps, "hand off extension execution to VM agent") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceVMExtensionsVisibilityLines(extension), extension.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceVMExtensionsGuestExecutionWalkthrough(extension)...)

	lines = append(lines, "- "+persistenceVMExtensionsUpdateBullet(extension))
	if persistenceCapabilityStatus(extension.CapabilitySteps, "update extension later") != "yes" {
		return persistenceTruncatedWalkthrough(lines, persistenceVMExtensionsVisibilityLines(extension), extension.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceVMExtensionsUpdateWalkthrough(extension)...)

	lines = append(lines, "- "+persistenceVMExtensionsPreserveBullet(extension))
	lines = append(lines, persistenceVMExtensionsPreserveWalkthrough(extension)...)
	if nearby := persistenceVMExtensionsNearbyNamesLine(extension.CurrentState.NearbyThematicNames); nearby != "" {
		lines = append(lines, "  "+nearby)
	}
	return renderPersistenceWalkthrough(lines)
}

func persistenceVMExtensionsControlBullet(extension models.PersistenceVMExtension) string {
	switch persistenceCapabilityStatus(extension.CapabilitySteps, "modify VM extension configuration") {
	case "yes":
		return "Current identity can modify VM extension configuration on this VM or VMSS."
	default:
		return "Current identity does not yet have a proven path to modify VM extension configuration on this VM or VMSS."
	}
}

func persistenceVMExtensionsControlWalkthrough(extension models.PersistenceVMExtension) []string {
	if extension.CurrentIdentityContext == nil {
		return nil
	}
	return []string{"  " + extension.CurrentIdentityContext.Summary}
}

func persistenceVMExtensionsTargetBullet(extension models.PersistenceVMExtension) string {
	switch persistenceCapabilityStatus(extension.CapabilitySteps, "reuse VM or VMSS target") {
	case "yes":
		return "Current identity can reuse this visible " + persistenceVMExtensionsTargetKindLabel(extension) + " target without interactive guest sign-in."
	default:
		return "Current identity does not yet have a proven path to reuse this visible " + persistenceVMExtensionsTargetKindLabel(extension) + " target."
	}
}

func persistenceVMExtensionsTargetWalkthrough(extension models.PersistenceVMExtension) []string {
	parts := []string{persistenceVMExtensionsTargetKindLabel(extension) + " `" + firstNonEmptyText(extension.CurrentState.TargetName, "unknown") + "`"}
	if len(extension.CurrentState.TargetIdentityIDs) > 0 {
		identityWord := "identities"
		if len(extension.CurrentState.TargetIdentityIDs) == 1 {
			identityWord = "identity"
		}
		parts = append(parts, fmt.Sprintf("%d target %s attached", len(extension.CurrentState.TargetIdentityIDs), identityWord))
	}
	if strings.EqualFold(extension.CurrentState.TargetKind, "vmss") {
		vmssParts := nonEmptyStrings(valueOrEmpty(extension.CurrentState.VMSSOrchestrationMode), valueOrEmpty(extension.CurrentState.VMSSUpgradeMode))
		if len(vmssParts) > 0 {
			parts = append(parts, "VMSS posture "+strings.Join(vmssParts, "/"))
		}
	}
	lines := []string{"  Visible target posture here is " + strings.Join(parts, ", ") + "."}
	if ctx := extension.CurrentState.StrongestVisibleExecutionContext; ctx != nil && strings.TrimSpace(ctx.Summary) != "" {
		lines = append(lines, "  "+ctx.Summary+" This does not prove the extension payload used that identity.")
	}
	return lines
}

func persistenceVMExtensionsAttachmentBullet(extension models.PersistenceVMExtension) string {
	switch persistenceCapabilityStatus(extension.CapabilitySteps, "add or modify extension attachment") {
	case "yes":
		return "Current identity can attach or change an extension definition that Azure stores on the VM or VMSS control-plane resource."
	default:
		return "Current identity does not yet have a proven path to attach or change the Azure-stored extension definition."
	}
}

func persistenceVMExtensionsAttachmentWalkthrough(extension models.PersistenceVMExtension) []string {
	handler := persistenceVMExtensionsHandlerLabel(extension)
	lines := []string{"  The extension attachment is the Azure-side object that binds handler, settings, source clues, and target together."}
	if handler != "" {
		lines = append(lines, "  Visible handler here is "+handler+".")
	}
	return lines
}

func persistenceVMExtensionsSourceBullet(extension models.PersistenceVMExtension) string {
	switch persistenceCapabilityStatus(extension.CapabilitySteps, "provide script or command source") {
	case "yes":
		return "Current identity can provide a script or command source directly, or point the extension at a reachable source such as storage, GitHub, or another host."
	default:
		return "Current identity does not yet have a proven path to provide or repoint the script or command source."
	}
}

func persistenceVMExtensionsSourceWalkthrough(extension models.PersistenceVMExtension) []string {
	source := persistenceVMExtensionsSourceState(extension)
	if source == "none visible" {
		return []string{"  No public script or command source clue is visible for this extension from the current read path."}
	}
	return []string{
		"  This is the script or command source layer, separate from the settings that tell the handler how to apply it.",
		"  Visible source posture here includes " + source + ".",
	}
}

func persistenceVMExtensionsSettingsBullet(extension models.PersistenceVMExtension) string {
	switch persistenceCapabilityStatus(extension.CapabilitySteps, "configure extension execution") {
	case "yes":
		return "Current identity can set the extension settings that tell the handler what to run and how to apply it."
	default:
		return "Current identity does not yet have a proven path to set the extension settings that shape execution."
	}
}

func persistenceVMExtensionsSettingsWalkthrough(extension models.PersistenceVMExtension) []string {
	settings := persistenceVMExtensionsSettingsState(extension)
	if settings == "none visible" {
		return []string{"  No public setting keys or protected-settings posture are visible for this extension from the current read path."}
	}
	return []string{"  Visible settings posture here includes " + settings + "."}
}

func persistenceVMExtensionsDeliveryBullet(extension models.PersistenceVMExtension) string {
	switch persistenceCapabilityStatus(extension.CapabilitySteps, "deliver config to VM agent") {
	case "yes":
		return "Azure stores the extension configuration and the VM agent receives and applies it on the target."
	default:
		return "Azure stores extension configuration, but this path does not yet have proven current-identity control here."
	}
}

func persistenceVMExtensionsDeliveryWalkthrough(extension models.PersistenceVMExtension) []string {
	status := persistenceVMExtensionsStatusState(extension)
	if status == "none visible" {
		return []string{"  Azure-visible status is not present in this row, so delivery conclusions stay at the extension resource and target linkage."}
	}
	return []string{"  Visible Azure status here includes " + status + "."}
}

func persistenceVMExtensionsGuestExecutionBullet(extension models.PersistenceVMExtension) string {
	switch persistenceCapabilityStatus(extension.CapabilitySteps, "hand off extension execution to VM agent") {
	case "yes":
		return "Execution happens locally on the VM through the VM agent, while the trigger and configuration come from Azure."
	default:
		return "Execution would happen locally through the VM agent, but the current identity does not yet have a proven path to control this extension."
	}
}

func persistenceVMExtensionsGuestExecutionWalkthrough(extension models.PersistenceVMExtension) []string {
	return []string{"  Default output does not prove guest-side success, logs, filesystem changes, or runtime effects."}
}

func persistenceVMExtensionsUpdateBullet(extension models.PersistenceVMExtension) string {
	switch persistenceCapabilityStatus(extension.CapabilitySteps, "update extension later") {
	case "yes":
		return "Current identity can update the extension later, which may start another extension apply cycle."
	default:
		return "Current identity does not yet have a proven path to update this extension later."
	}
}

func persistenceVMExtensionsUpdateWalkthrough(extension models.PersistenceVMExtension) []string {
	rerun := persistenceVMExtensionsRerunState(extension)
	if rerun == "none visible" {
		return []string{"  No explicit rerun clue is visible here, so the reapply story is the normal extension update path."}
	}
	return []string{
		"  Visible rerun posture here includes " + rerun + ".",
		"  This does not prove the same command or payload still succeeds.",
	}
}

func persistenceVMExtensionsPreserveBullet(extension models.PersistenceVMExtension) string {
	switch persistenceCapabilityStatus(extension.CapabilitySteps, "preserve control-plane execution path") {
	case "yes":
		return "This acts like persistence because the extension attachment and settings remain in Azure control-plane state until changed."
	default:
		return "This extension is still visible Azure-side configuration, but the current identity does not yet have a proven path to preserve or repurpose it here."
	}
}

func persistenceVMExtensionsPreserveWalkthrough(extension models.PersistenceVMExtension) []string {
	return []string{
		"  Defender review should start with the VM or VMSS target, extension handler, source clues, settings posture, rerun clues, and visible status.",
	}
}

func persistenceVMExtensionsVisibilityLines(extension models.PersistenceVMExtension) []string {
	state := strings.TrimSpace(persistenceVMExtensionsInventoryState(extension))
	execution := strings.TrimSpace(persistenceVMExtensionsInventoryExecutionContext(extension))
	return persistenceVisibilityFallbackLines(
		state,
		execution,
		"this VM Extension already has handler, source, settings, rerun, or reuse value if stronger control is obtained later.",
		"this VM Extension is worth revisiting if stronger control is obtained later.",
		"  Visibility still confirms this VM Extension exists, even though the current identity does not yet have a proven write path here.",
	)
}

func persistenceVMExtensionsNearbyNamesLine(names []string) string {
	if len(names) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, "`"+name+"`")
	}
	return "Nearby maintenance-, bootstrap-, dependency-, or configuration-themed VM Extension names visible from the current environment include " + renderNaturalJoin(quoted) + "."
}

func persistenceLogicAppWorkflowBullet(workflow models.PersistenceLogicAppWorkflow) string {
	switch persistenceCapabilityStatus(workflow.CapabilitySteps, "create or modify workflow") {
	case "yes":
		return "Current identity can create a new Logic App or modify this existing workflow."
	default:
		return "Current identity does not yet have a proven path to create a new Logic App or modify this existing workflow."
	}
}

func persistenceLogicAppDefinitionBullet(workflow models.PersistenceLogicAppWorkflow) string {
	switch persistenceCapabilityStatus(workflow.CapabilitySteps, "edit workflow definition") {
	case "yes":
		return "Current identity can change the stored workflow definition Azure will execute here."
	default:
		return "Current identity does not yet have a proven path to change the stored workflow definition here."
	}
}

func persistenceWebJobNearbyNamesLine(names []string) string {
	if len(names) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, "`"+name+"`")
	}
	return "Nearby maintenance- or sync-themed WebJob names visible from the current environment include " + renderNaturalJoin(quoted) + "."
}

func persistenceLogicAppExecutionContextBullet(workflow models.PersistenceLogicAppWorkflow) string {
	switch persistenceCapabilityStatus(workflow.CapabilitySteps, "attach or reuse exec ctx") {
	case "yes":
		return "Current identity can attach or reuse execution context for this Logic App."
	default:
		return "Current identity does not yet have a proven path to attach or reuse execution context for this Logic App."
	}
}

func persistenceLogicAppTriggerBullet(workflow models.PersistenceLogicAppWorkflow) string {
	switch persistenceCapabilityStatus(workflow.CapabilitySteps, "define or modify trigger") {
	case "yes":
		return "Current identity can define or modify request, recurrence, or event trigger posture for this Logic App."
	default:
		return "Current identity does not yet have a proven path to define or modify durable trigger posture for this Logic App."
	}
}

func persistenceLogicAppEnableBullet(workflow models.PersistenceLogicAppWorkflow) string {
	switch persistenceCapabilityStatus(workflow.CapabilitySteps, "enable workflow") {
	case "yes":
		return "Current identity can enable the workflow so Azure will listen for the trigger and run it later."
	default:
		return "Current identity does not yet have a proven path to enable this workflow for later trigger-driven execution."
	}
}

func persistenceLogicAppActionBullet(workflow models.PersistenceLogicAppWorkflow) string {
	switch persistenceCapabilityStatus(workflow.CapabilitySteps, "add or repurpose downstream actions") {
	case "yes":
		return "Current identity can add or repurpose the downstream action paths this Logic App will carry out after the trigger fires."
	default:
		return "Current identity does not yet have a proven path to add or repurpose the downstream action paths this Logic App exposes."
	}
}

func persistenceLogicAppRepurposeBullet(workflow models.PersistenceLogicAppWorkflow) string {
	switch persistenceCapabilityStatus(workflow.CapabilitySteps, "create or modify workflow") {
	case "yes":
		return "Current identity can repurpose an existing Logic App instead of creating a new one."
	default:
		return "Current identity does not yet have a proven path to repurpose this existing Logic App."
	}
}

func renderBulletList(items []string) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, "- "+item)
	}
	return strings.Join(lines, "\n")
}

func persistenceRoleContextLine(context *models.PersistenceRoleContext) string {
	if context == nil {
		return "none visible"
	}
	if len(context.RoleNames) > 0 {
		return fmt.Sprintf("`%s` with %s", context.Name, renderNaturalJoin(context.RoleNames))
	}
	return fmt.Sprintf("`%s` with no Azure role-assignment rows found for its principal ID", context.Name)
}

func persistenceRoleContextLabel(context models.PersistenceRoleContext) string {
	label := "`" + context.Name + "`"
	if len(context.RoleNames) == 0 {
		return label
	}
	return label + " with " + persistenceRoleSummaryForRender(context.RoleNames, context.ScopeIDs)
}

func persistenceRoleSummaryForRender(roleNames []string, scopeIDs []string) string {
	roleText := strings.Join(roleNames, ", ")
	if roleText == "" {
		roleText = "visible Azure roles"
	}
	if len(scopeIDs) == 0 {
		return roleText
	}
	if len(scopeIDs) == 1 {
		return roleText + " at " + persistenceScopeLabelForRender(scopeIDs[0])
	}
	return fmt.Sprintf("%s across %d visible scopes", roleText, len(scopeIDs))
}

func persistenceScopeLabelForRender(scopeID string) string {
	if strings.Contains(scopeID, "/subscriptions/") && !strings.Contains(scopeID, "/resourceGroups/") {
		return "subscription scope"
	}
	if strings.Contains(scopeID, "/resourceGroups/") {
		return "resource group `" + renderScopeName(scopeID) + "`"
	}
	return "a parent scope"
}

func persistenceJoinedOrNone(values []string) string {
	if len(values) == 0 {
		return "none visible"
	}
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, "`"+value+"`")
	}
	return renderNaturalJoin(quoted)
}

func textOrNone(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "none visible"
	}
	return value
}

func persistenceAutomationLeadAccount(accounts []models.PersistenceAutomationAccount) models.PersistenceAutomationAccount {
	lead := accounts[0]
	for _, candidate := range accounts[1:] {
		if persistenceAutomationAccountRanksBefore(candidate, lead) {
			lead = candidate
		}
	}
	return lead
}

func persistenceAutomationAccountRanksBefore(left, right models.PersistenceAutomationAccount) bool {
	leftHasExecRole := left.CurrentState.StrongestVisibleExecutionContext != nil && len(left.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	rightHasExecRole := right.CurrentState.StrongestVisibleExecutionContext != nil && len(right.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	switch {
	case leftHasExecRole != rightHasExecRole:
		return leftHasExecRole
	case intPtrValue(left.CurrentState.PublishedRunbookCount) != intPtrValue(right.CurrentState.PublishedRunbookCount):
		return intPtrValue(left.CurrentState.PublishedRunbookCount) > intPtrValue(right.CurrentState.PublishedRunbookCount)
	case intPtrValue(left.CurrentState.WebhookCount) != intPtrValue(right.CurrentState.WebhookCount):
		return intPtrValue(left.CurrentState.WebhookCount) > intPtrValue(right.CurrentState.WebhookCount)
	case intPtrValue(left.CurrentState.ScheduleCount) != intPtrValue(right.CurrentState.ScheduleCount):
		return intPtrValue(left.CurrentState.ScheduleCount) > intPtrValue(right.CurrentState.ScheduleCount)
	case intPtrValue(left.CurrentState.RunbookCount) != intPtrValue(right.CurrentState.RunbookCount):
		return intPtrValue(left.CurrentState.RunbookCount) > intPtrValue(right.CurrentState.RunbookCount)
	default:
		return left.Name < right.Name
	}
}

func persistenceAutomationInventoryRows(accounts []models.PersistenceAutomationAccount) [][]string {
	rows := make([][]string, 0, len(accounts))
	for _, account := range accounts {
		rows = append(rows, []string{
			account.Name,
			account.ResourceGroup,
			persistenceAutomationInventoryState(account),
			persistenceAutomationInventoryExecutionContext(account),
		})
	}
	return rows
}

func persistenceAutomationInventoryState(account models.PersistenceAutomationAccount) string {
	parts := []string{
		fmt.Sprintf("%s/%s published", intPtrString(account.CurrentState.PublishedRunbookCount), intPtrString(account.CurrentState.RunbookCount)),
		fmt.Sprintf("schedules %s", intPtrString(account.CurrentState.ScheduleCount)),
		fmt.Sprintf("webhooks %s", intPtrString(account.CurrentState.WebhookCount)),
	}
	if mode := textOrNone(valueOrEmpty(account.CurrentState.PrimaryStartMode)); mode != "none visible" {
		parts = append(parts, "primary "+mode)
	}
	return strings.Join(parts, "; ")
}

func persistenceAutomationInventoryExecutionContext(account models.PersistenceAutomationAccount) string {
	if ctx := account.CurrentState.StrongestVisibleExecutionContext; ctx != nil {
		if len(ctx.RoleNames) > 0 {
			return persistenceRoleContextLabel(*ctx)
		}
		return persistenceRoleContextLine(ctx)
	}
	if len(account.ExecutionContextOptions) > 0 {
		return strings.Join(account.ExecutionContextOptions, ", ")
	}
	return "none visible"
}

func persistenceAutomationCombinedStillUnmapped(accounts []models.PersistenceAutomationAccount) string {
	items := []string{}
	seen := map[string]struct{}{}
	for _, account := range accounts {
		for _, item := range account.StillUnmapped {
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			items = append(items, item)
		}
	}
	return renderBulletList(items)
}

func persistenceLogicAppLeadWorkflow(workflows []models.PersistenceLogicAppWorkflow) models.PersistenceLogicAppWorkflow {
	lead := workflows[0]
	for _, candidate := range workflows[1:] {
		if persistenceLogicAppRanksBefore(candidate, lead) {
			lead = candidate
		}
	}
	return lead
}

func persistenceLogicAppRanksBefore(left, right models.PersistenceLogicAppWorkflow) bool {
	leftDurable := left.CurrentState.Classification == "persistence-capable"
	rightDurable := right.CurrentState.Classification == "persistence-capable"
	leftHasRequest := left.CurrentState.ExternallyCallableRequestTrigger
	rightHasRequest := right.CurrentState.ExternallyCallableRequestTrigger
	leftHasRecurrence := strings.TrimSpace(valueOrEmpty(left.CurrentState.RecurrenceSummary)) != ""
	rightHasRecurrence := strings.TrimSpace(valueOrEmpty(right.CurrentState.RecurrenceSummary)) != ""
	leftHasExecRole := left.CurrentState.StrongestVisibleExecutionContext != nil && len(left.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	rightHasExecRole := right.CurrentState.StrongestVisibleExecutionContext != nil && len(right.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	switch {
	case leftDurable != rightDurable:
		return leftDurable
	case leftHasRequest != rightHasRequest:
		return leftHasRequest
	case leftHasRecurrence != rightHasRecurrence:
		return leftHasRecurrence
	case leftHasExecRole != rightHasExecRole:
		return leftHasExecRole
	case len(left.CurrentState.DownstreamActionKinds) != len(right.CurrentState.DownstreamActionKinds):
		return len(left.CurrentState.DownstreamActionKinds) > len(right.CurrentState.DownstreamActionKinds)
	default:
		return left.Name < right.Name
	}
}

func persistenceLogicAppInventoryRows(workflows []models.PersistenceLogicAppWorkflow) [][]string {
	rows := make([][]string, 0, len(workflows))
	for _, workflow := range workflows {
		rows = append(rows, []string{
			workflow.Name,
			workflow.ResourceGroup,
			persistenceLogicAppInventoryState(workflow),
			persistenceLogicAppInventoryExecutionContext(workflow),
		})
	}
	return rows
}

func persistenceLogicAppInventoryState(workflow models.PersistenceLogicAppWorkflow) string {
	parts := []string{textOrNone(workflow.CurrentState.Classification)}
	switch {
	case workflow.CurrentState.ExternallyCallableRequestTrigger:
		parts = append(parts, "request(external)")
	case strings.TrimSpace(valueOrEmpty(workflow.CurrentState.RecurrenceSummary)) != "":
		parts = append(parts, "recurrence "+valueOrEmpty(workflow.CurrentState.RecurrenceSummary))
	default:
		triggerTypes := strings.ReplaceAll(persistenceJoinedOrNone(workflow.CurrentState.TriggerTypes), "`", "")
		if strings.TrimSpace(triggerTypes) != "" && triggerTypes != "none visible" {
			parts = append(parts, triggerTypes)
		}
	}
	return strings.Join(parts, "; ")
}

func persistenceLogicAppInventoryExecutionContext(workflow models.PersistenceLogicAppWorkflow) string {
	if ctx := workflow.CurrentState.StrongestVisibleExecutionContext; ctx != nil {
		if len(ctx.RoleNames) > 0 {
			return persistenceRoleContextLabel(*ctx)
		}
		return persistenceRoleContextLine(ctx)
	}
	if len(workflow.ExecutionContextOptions) > 0 {
		return strings.Join(workflow.ExecutionContextOptions, ", ")
	}
	return "none visible"
}

func persistenceLogicAppCombinedStillUnmapped(workflows []models.PersistenceLogicAppWorkflow) []string {
	items := []string{}
	seen := map[string]struct{}{}
	for _, workflow := range workflows {
		for _, item := range workflow.StillUnmapped {
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			items = append(items, item)
		}
	}
	return items
}

func persistenceFunctionsLeadApp(apps []models.PersistenceFunctionApp) models.PersistenceFunctionApp {
	lead := apps[0]
	for _, candidate := range apps[1:] {
		if persistenceFunctionsRanksBefore(candidate, lead) {
			lead = candidate
		}
	}
	return lead
}

func persistenceAppServiceLeadApp(apps []models.PersistenceAppService) models.PersistenceAppService {
	lead := apps[0]
	for _, candidate := range apps[1:] {
		if persistenceAppServiceRanksBefore(candidate, lead) {
			lead = candidate
		}
	}
	return lead
}

func persistenceWebJobsLeadJob(jobs []models.PersistenceWebJob) models.PersistenceWebJob {
	lead := jobs[0]
	for _, candidate := range jobs[1:] {
		if persistenceWebJobsRanksBefore(candidate, lead) {
			lead = candidate
		}
	}
	return lead
}

func persistenceContainerAppsJobsLeadJob(jobs []models.PersistenceContainerAppsJob) models.PersistenceContainerAppsJob {
	lead := jobs[0]
	for _, candidate := range jobs[1:] {
		if persistenceContainerAppsJobsRanksBefore(candidate, lead) {
			lead = candidate
		}
	}
	return lead
}

func persistenceVMExtensionsLeadExtension(extensions []models.PersistenceVMExtension) models.PersistenceVMExtension {
	lead := extensions[0]
	for _, candidate := range extensions[1:] {
		if persistenceVMExtensionsRanksBefore(candidate, lead) {
			lead = candidate
		}
	}
	return lead
}

func persistenceAppServiceRanksBefore(left, right models.PersistenceAppService) bool {
	leftHasExecRole := left.CurrentState.StrongestVisibleExecutionContext != nil && len(left.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	rightHasExecRole := right.CurrentState.StrongestVisibleExecutionContext != nil && len(right.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	leftPublic := strings.EqualFold(valueOrEmpty(left.CurrentState.PublicNetworkAccess), "Enabled") || strings.TrimSpace(valueOrEmpty(left.CurrentState.Hostname)) != ""
	rightPublic := strings.EqualFold(valueOrEmpty(right.CurrentState.PublicNetworkAccess), "Enabled") || strings.TrimSpace(valueOrEmpty(right.CurrentState.Hostname)) != ""
	leftHasDeployment := strings.TrimSpace(valueOrEmpty(left.CurrentState.Deployment)) != ""
	rightHasDeployment := strings.TrimSpace(valueOrEmpty(right.CurrentState.Deployment)) != ""
	leftDeploymentScore := len(strings.TrimSpace(valueOrEmpty(left.CurrentState.Deployment)))
	rightDeploymentScore := len(strings.TrimSpace(valueOrEmpty(right.CurrentState.Deployment)))
	leftConfigCount := intPtrValue(left.CurrentState.AppSettingsCount) + intPtrValue(left.CurrentState.ConnectionStringCount)
	rightConfigCount := intPtrValue(right.CurrentState.AppSettingsCount) + intPtrValue(right.CurrentState.ConnectionStringCount)
	switch {
	case leftPublic != rightPublic:
		return leftPublic
	case leftHasDeployment != rightHasDeployment:
		return leftHasDeployment
	case leftDeploymentScore != rightDeploymentScore:
		return leftDeploymentScore > rightDeploymentScore
	case leftConfigCount != rightConfigCount:
		return leftConfigCount > rightConfigCount
	case leftHasExecRole != rightHasExecRole:
		return leftHasExecRole
	default:
		return left.Name < right.Name
	}
}

func persistenceWebJobsRanksBefore(left, right models.PersistenceWebJob) bool {
	leftHasExecRole := left.CurrentState.StrongestVisibleExecutionContext != nil && len(left.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	rightHasExecRole := right.CurrentState.StrongestVisibleExecutionContext != nil && len(right.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	leftHostname := strings.TrimSpace(valueOrEmpty(left.CurrentState.ParentHostname)) != ""
	rightHostname := strings.TrimSpace(valueOrEmpty(right.CurrentState.ParentHostname)) != ""
	leftCommand := strings.TrimSpace(valueOrEmpty(left.CurrentState.RunCommand)) != ""
	rightCommand := strings.TrimSpace(valueOrEmpty(right.CurrentState.RunCommand)) != ""
	leftModeRank := persistenceWebJobModeRank(left.CurrentState.Mode)
	rightModeRank := persistenceWebJobModeRank(right.CurrentState.Mode)
	switch {
	case leftHasExecRole != rightHasExecRole:
		return leftHasExecRole
	case leftModeRank != rightModeRank:
		return leftModeRank < rightModeRank
	case leftCommand != rightCommand:
		return leftCommand
	case leftHostname != rightHostname:
		return leftHostname
	default:
		return left.Name < right.Name
	}
}

func persistenceContainerAppsJobsRanksBefore(left, right models.PersistenceContainerAppsJob) bool {
	leftHasExecRole := left.CurrentState.StrongestVisibleExecutionContext != nil && len(left.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	rightHasExecRole := right.CurrentState.StrongestVisibleExecutionContext != nil && len(right.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	leftTriggerRank := persistenceContainerAppsJobTriggerRank(left.CurrentState.TriggerType)
	rightTriggerRank := persistenceContainerAppsJobTriggerRank(right.CurrentState.TriggerType)
	leftPayload := len(left.CurrentState.ContainerImages) > 0 || len(left.CurrentState.Command) > 0
	rightPayload := len(right.CurrentState.ContainerImages) > 0 || len(right.CurrentState.Command) > 0
	leftAccessPosture := intPtrValue(left.CurrentState.SecretCount) + intPtrValue(left.CurrentState.RegistryIdentityCount) + intPtrValue(left.CurrentState.RegistryPasswordRefCount)
	rightAccessPosture := intPtrValue(right.CurrentState.SecretCount) + intPtrValue(right.CurrentState.RegistryIdentityCount) + intPtrValue(right.CurrentState.RegistryPasswordRefCount)
	switch {
	case leftHasExecRole != rightHasExecRole:
		return leftHasExecRole
	case leftTriggerRank != rightTriggerRank:
		return leftTriggerRank < rightTriggerRank
	case leftPayload != rightPayload:
		return leftPayload
	case leftAccessPosture != rightAccessPosture:
		return leftAccessPosture > rightAccessPosture
	default:
		return left.Name < right.Name
	}
}

func persistenceVMExtensionsRanksBefore(left, right models.PersistenceVMExtension) bool {
	leftHasCurrentControl := left.CurrentIdentityContext != nil && len(left.CurrentIdentityContext.RoleNames) > 0
	rightHasCurrentControl := right.CurrentIdentityContext != nil && len(right.CurrentIdentityContext.RoleNames) > 0
	leftHasExecRole := left.CurrentState.StrongestVisibleExecutionContext != nil && len(left.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	rightHasExecRole := right.CurrentState.StrongestVisibleExecutionContext != nil && len(right.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	leftCustomScript := persistenceVMExtensionsIsCustomScript(left)
	rightCustomScript := persistenceVMExtensionsIsCustomScript(right)
	leftCommand := strings.TrimSpace(valueOrEmpty(left.CurrentState.CommandClue)) != ""
	rightCommand := strings.TrimSpace(valueOrEmpty(right.CurrentState.CommandClue)) != ""
	leftSources := len(left.CurrentState.FileURIHosts)
	rightSources := len(right.CurrentState.FileURIHosts)
	leftRerun := len(left.CurrentState.RerunClues) > 0 || strings.TrimSpace(valueOrEmpty(left.CurrentState.ForceUpdateTag)) != ""
	rightRerun := len(right.CurrentState.RerunClues) > 0 || strings.TrimSpace(valueOrEmpty(right.CurrentState.ForceUpdateTag)) != ""
	leftProtected := boolPtrValue(left.CurrentState.ProtectedSettingsPresent) || boolPtrValue(left.CurrentState.KeyVaultProtectedSettings)
	rightProtected := boolPtrValue(right.CurrentState.ProtectedSettingsPresent) || boolPtrValue(right.CurrentState.KeyVaultProtectedSettings)
	switch {
	case leftHasExecRole != rightHasExecRole:
		return leftHasExecRole
	case leftHasCurrentControl != rightHasCurrentControl:
		return leftHasCurrentControl
	case leftCustomScript != rightCustomScript:
		return leftCustomScript
	case leftCommand != rightCommand:
		return leftCommand
	case leftSources != rightSources:
		return leftSources > rightSources
	case leftRerun != rightRerun:
		return leftRerun
	case leftProtected != rightProtected:
		return leftProtected
	default:
		return left.Name < right.Name
	}
}

func persistenceContainerAppsJobTriggerRank(trigger *string) int {
	switch strings.ToLower(strings.TrimSpace(valueOrEmpty(trigger))) {
	case "schedule", "scheduled":
		return 0
	case "event", "event-driven":
		return 1
	case "manual":
		return 2
	default:
		return 3
	}
}

func persistenceWebJobModeRank(mode string) int {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "continuous":
		return 0
	case "scheduled":
		return 1
	case "triggered/manual":
		return 2
	default:
		return 3
	}
}

func persistenceAppServiceInventoryRows(apps []models.PersistenceAppService) [][]string {
	rows := make([][]string, 0, len(apps))
	for _, app := range apps {
		rows = append(rows, []string{
			app.Name,
			app.ResourceGroup,
			persistenceAppServiceInventoryState(app),
			persistenceAppServiceInventoryExecutionContext(app),
		})
	}
	return rows
}

func persistenceWebJobsInventoryRows(jobs []models.PersistenceWebJob) [][]string {
	rows := make([][]string, 0, len(jobs))
	for _, job := range jobs {
		rows = append(rows, []string{
			job.Name,
			job.CurrentState.ParentAppName,
			persistenceWebJobsInventoryState(job),
			persistenceWebJobsInventoryExecutionContext(job),
		})
	}
	return rows
}

func persistenceContainerAppsJobsInventoryRows(jobs []models.PersistenceContainerAppsJob) [][]string {
	rows := make([][]string, 0, len(jobs))
	for _, job := range jobs {
		rows = append(rows, []string{
			job.Name,
			persistenceContainerAppsJobsTriggerInventory(job),
			persistenceContainerAppsJobsInventoryState(job),
			persistenceContainerAppsJobsInventoryExecutionContext(job),
		})
	}
	return rows
}

func persistenceVMExtensionsInventoryRows(extensions []models.PersistenceVMExtension) [][]string {
	rows := make([][]string, 0, len(extensions))
	for _, extension := range extensions {
		rows = append(rows, []string{
			extension.Name,
			persistenceVMExtensionsTargetInventory(extension),
			persistenceVMExtensionsInventoryState(extension),
			persistenceVMExtensionsInventoryExecutionContext(extension),
		})
	}
	return rows
}

func persistenceAppServiceInventoryState(app models.PersistenceAppService) string {
	parts := []string{}
	if state := strings.TrimSpace(valueOrEmpty(app.CurrentState.State)); state != "" {
		parts = append(parts, state)
	}
	if hostname := strings.TrimSpace(valueOrEmpty(app.CurrentState.Hostname)); hostname != "" {
		parts = append(parts, "hostname "+hostname)
	}
	if runtime := strings.TrimSpace(valueOrEmpty(app.CurrentState.Runtime)); runtime != "" {
		parts = append(parts, runtime)
	}
	if deployment := strings.TrimSpace(valueOrEmpty(app.CurrentState.Deployment)); deployment != "" {
		parts = append(parts, deployment)
	}
	configParts := []string{}
	if count := intPtrValue(app.CurrentState.AppSettingsCount); count > 0 {
		configParts = append(configParts, fmt.Sprintf("settings=%d", count))
	}
	if count := intPtrValue(app.CurrentState.KeyVaultReferenceCount); count > 0 {
		configParts = append(configParts, fmt.Sprintf("kv=%d", count))
	}
	if count := intPtrValue(app.CurrentState.ConnectionStringCount); count > 0 {
		configParts = append(configParts, fmt.Sprintf("conn=%d", count))
	}
	if len(configParts) > 0 {
		parts = append(parts, strings.Join(configParts, "; "))
	}
	if network := strings.TrimSpace(valueOrEmpty(app.CurrentState.PublicNetworkAccess)); network != "" {
		parts = append(parts, "public "+network)
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return strings.Join(parts, "; ")
}

func persistenceAppServiceInventoryExecutionContext(app models.PersistenceAppService) string {
	if ctx := app.CurrentState.StrongestVisibleExecutionContext; ctx != nil {
		if len(ctx.RoleNames) > 0 {
			return persistenceRoleContextLabel(*ctx)
		}
		return persistenceRoleContextLine(ctx)
	}
	if len(app.ExecutionContextOptions) > 0 {
		return strings.Join(app.ExecutionContextOptions, ", ")
	}
	return "none visible"
}

func persistenceWebJobsInventoryState(job models.PersistenceWebJob) string {
	parts := []string{}
	if mode := strings.TrimSpace(job.CurrentState.Mode); mode != "" {
		parts = append(parts, mode)
	}
	if status := strings.TrimSpace(valueOrEmpty(job.CurrentState.Status)); status != "" {
		parts = append(parts, "status "+status)
	}
	if trigger := strings.TrimSpace(valueOrEmpty(job.CurrentState.LatestRunTrigger)); trigger != "" {
		parts = append(parts, "latest trigger "+trigger)
	}
	if schedule := strings.TrimSpace(valueOrEmpty(job.CurrentState.ScheduleExpression)); schedule != "" {
		parts = append(parts, "schedule "+schedule)
	}
	if command := strings.TrimSpace(valueOrEmpty(job.CurrentState.RunCommand)); command != "" {
		parts = append(parts, "run command visible")
	}
	if hostname := strings.TrimSpace(valueOrEmpty(job.CurrentState.ParentHostname)); hostname != "" {
		parts = append(parts, "parent hostname "+hostname)
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return strings.Join(parts, "; ")
}

func persistenceWebJobsInventoryExecutionContext(job models.PersistenceWebJob) string {
	if ctx := job.CurrentState.StrongestVisibleExecutionContext; ctx != nil {
		if len(ctx.RoleNames) > 0 {
			return persistenceRoleContextLabel(*ctx)
		}
		return persistenceRoleContextLine(ctx)
	}
	if len(job.ExecutionContextOptions) > 0 {
		return strings.Join(job.ExecutionContextOptions, ", ")
	}
	return "none visible"
}

func persistenceContainerAppsJobsTriggerInventory(job models.PersistenceContainerAppsJob) string {
	parts := []string{}
	if trigger := strings.TrimSpace(valueOrEmpty(job.CurrentState.TriggerType)); trigger != "" {
		parts = append(parts, trigger)
	}
	if schedule := strings.TrimSpace(valueOrEmpty(job.CurrentState.ScheduleExpression)); schedule != "" {
		parts = append(parts, "schedule "+schedule)
	}
	if len(job.CurrentState.EventRules) > 0 {
		parts = append(parts, fmt.Sprintf("%d event rule(s)", len(job.CurrentState.EventRules)))
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return strings.Join(parts, "; ")
}

func persistenceContainerAppsJobsInventoryState(job models.PersistenceContainerAppsJob) string {
	parts := []string{}
	if environment := strings.TrimSpace(valueOrEmpty(job.CurrentState.EnvironmentID)); environment != "" {
		parts = append(parts, "environment "+resourceNameFromDisplayID(environment))
	}
	if len(job.CurrentState.ContainerImages) > 0 {
		parts = append(parts, fmt.Sprintf("%d image clue(s)", len(job.CurrentState.ContainerImages)))
	}
	if len(job.CurrentState.Command) > 0 {
		parts = append(parts, "command clue visible")
	}
	if count := intPtrValue(job.CurrentState.Parallelism); count > 0 {
		parts = append(parts, fmt.Sprintf("parallelism=%d", count))
	}
	if count := intPtrValue(job.CurrentState.SecretCount); count > 0 {
		parts = append(parts, fmt.Sprintf("secrets=%d", count))
	}
	if len(job.CurrentState.RegistryServers) > 0 {
		parts = append(parts, "registries="+strings.Join(job.CurrentState.RegistryServers, ", "))
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return wrapTableNote(strings.Join(parts, "; "), 72)
}

func persistenceContainerAppsJobsInventoryExecutionContext(job models.PersistenceContainerAppsJob) string {
	if ctx := job.CurrentState.StrongestVisibleExecutionContext; ctx != nil {
		if len(ctx.RoleNames) > 0 {
			return wrapTableNote(persistenceRoleContextLabel(*ctx), 72)
		}
		return wrapTableNote(persistenceRoleContextLine(ctx), 72)
	}
	if len(job.ExecutionContextOptions) > 0 {
		return wrapTableNote(strings.Join(job.ExecutionContextOptions, ", "), 72)
	}
	return "none visible"
}

func persistenceVMExtensionsTargetInventory(extension models.PersistenceVMExtension) string {
	target := persistenceVMExtensionsTargetKindLabel(extension)
	if name := strings.TrimSpace(extension.CurrentState.TargetName); name != "" {
		target += "=" + name
	}
	if len(extension.CurrentState.TargetIdentityIDs) > 0 {
		target += fmt.Sprintf("; identities=%d", len(extension.CurrentState.TargetIdentityIDs))
	}
	return target
}

func persistenceVMExtensionsInventoryState(extension models.PersistenceVMExtension) string {
	parts := []string{}
	if handler := persistenceVMExtensionsHandlerLabel(extension); handler != "" {
		parts = append(parts, handler)
	}
	if source := persistenceVMExtensionsSourceState(extension); source != "none visible" {
		parts = append(parts, source)
	}
	if settings := persistenceVMExtensionsSettingsState(extension); settings != "none visible" {
		parts = append(parts, settings)
	}
	if rerun := persistenceVMExtensionsRerunState(extension); rerun != "none visible" {
		parts = append(parts, rerun)
	}
	if status := persistenceVMExtensionsStatusState(extension); status != "none visible" {
		parts = append(parts, status)
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return wrapTableNote(strings.Join(parts, "; "), 72)
}

func persistenceVMExtensionsInventoryExecutionContext(extension models.PersistenceVMExtension) string {
	if ctx := extension.CurrentState.StrongestVisibleExecutionContext; ctx != nil {
		if len(ctx.RoleNames) > 0 {
			return wrapTableNote(persistenceRoleContextLabel(*ctx), 72)
		}
		return wrapTableNote(persistenceRoleContextLine(ctx), 72)
	}
	if len(extension.ExecutionContextOptions) > 0 {
		return wrapTableNote(strings.Join(extension.ExecutionContextOptions, ", "), 72)
	}
	return "none visible"
}

func persistenceVMExtensionsHandlerLabel(extension models.PersistenceVMExtension) string {
	handler := firstNonEmptyText(valueOrEmpty(extension.CurrentState.Publisher), "unknown publisher") + "/" + firstNonEmptyText(valueOrEmpty(extension.CurrentState.ExtensionType), "unknown extension")
	if version := strings.TrimSpace(valueOrEmpty(extension.CurrentState.TypeHandlerVersion)); version != "" {
		handler += " " + version
	}
	return handler
}

func persistenceVMExtensionsSourceState(extension models.PersistenceVMExtension) string {
	parts := []string{}
	if len(extension.CurrentState.FileURIHosts) > 0 {
		parts = append(parts, "file-hosts="+strings.Join(extension.CurrentState.FileURIHosts, ", "))
	}
	if count := intPtrValue(extension.CurrentState.FileURICount); count > 0 {
		parts = append(parts, fmt.Sprintf("fileUris=%d", count))
	}
	if command := strings.TrimSpace(valueOrEmpty(extension.CurrentState.CommandClue)); command != "" {
		parts = append(parts, "command clue `"+command+"`")
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return strings.Join(parts, "; ")
}

func persistenceVMExtensionsSettingsState(extension models.PersistenceVMExtension) string {
	parts := []string{}
	if len(extension.CurrentState.PublicSettingKeys) > 0 {
		parts = append(parts, "public="+strings.Join(extension.CurrentState.PublicSettingKeys, ", "))
	}
	if extension.CurrentState.ProtectedSettingsPresent != nil {
		parts = append(parts, "protected="+boolWord(boolPtrValue(extension.CurrentState.ProtectedSettingsPresent)))
	}
	if boolPtrValue(extension.CurrentState.KeyVaultProtectedSettings) {
		parts = append(parts, "kv-protected=yes")
	}
	if extension.CurrentState.SuppressFailures != nil {
		parts = append(parts, "suppress-failures="+boolWord(boolPtrValue(extension.CurrentState.SuppressFailures)))
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return strings.Join(parts, "; ")
}

func persistenceVMExtensionsRerunState(extension models.PersistenceVMExtension) string {
	parts := append([]string{}, extension.CurrentState.RerunClues...)
	if tag := strings.TrimSpace(valueOrEmpty(extension.CurrentState.ForceUpdateTag)); tag != "" && !persistenceStringSliceContains(parts, "forceUpdateTag="+tag) {
		parts = append(parts, "forceUpdateTag="+tag)
	}
	if len(extension.CurrentState.ProvisionAfterExtensions) > 0 {
		parts = append(parts, "after="+strings.Join(extension.CurrentState.ProvisionAfterExtensions, ", "))
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return strings.Join(parts, "; ")
}

func persistenceVMExtensionsStatusState(extension models.PersistenceVMExtension) string {
	parts := []string{}
	if status := strings.TrimSpace(valueOrEmpty(extension.CurrentState.ProvisioningState)); status != "" {
		parts = append(parts, "provisioning="+status)
	}
	if len(extension.CurrentState.InstanceViewStatuses) > 0 {
		parts = append(parts, "instance="+strings.Join(extension.CurrentState.InstanceViewStatuses, ", "))
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return strings.Join(parts, "; ")
}

func persistenceStringSliceContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func persistenceVMExtensionsTargetKindLabel(extension models.PersistenceVMExtension) string {
	switch strings.ToLower(strings.TrimSpace(extension.CurrentState.TargetKind)) {
	case "vm":
		return "VM"
	case "vmss":
		return "VMSS"
	default:
		return firstNonEmptyText(strings.ToUpper(strings.TrimSpace(extension.CurrentState.TargetKind)), "target")
	}
}

func persistenceVMExtensionsIsCustomScript(extension models.PersistenceVMExtension) bool {
	return strings.Contains(strings.ToLower(valueOrEmpty(extension.CurrentState.ExtensionType)), "customscript")
}

func persistenceAppServiceCombinedStillUnmapped(apps []models.PersistenceAppService) []string {
	items := []string{}
	seen := map[string]struct{}{}
	for _, app := range apps {
		for _, item := range app.StillUnmapped {
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			items = append(items, item)
		}
	}
	return items
}

func persistenceWebJobsCombinedStillUnmapped(jobs []models.PersistenceWebJob) []string {
	return persistenceCombinedStillUnmapped(jobs, func(job models.PersistenceWebJob) []string {
		return job.StillUnmapped
	})
}

func persistenceContainerAppsJobsCombinedStillUnmapped(jobs []models.PersistenceContainerAppsJob) []string {
	return persistenceCombinedStillUnmapped(jobs, func(job models.PersistenceContainerAppsJob) []string {
		return job.StillUnmapped
	})
}

func persistenceVMExtensionsCombinedStillUnmapped(extensions []models.PersistenceVMExtension) []string {
	return persistenceCombinedStillUnmapped(extensions, func(extension models.PersistenceVMExtension) []string {
		return extension.StillUnmapped
	})
}

func persistenceCombinedStillUnmapped[T any](values []T, itemsFor func(T) []string) []string {
	items := []string{}
	seen := map[string]struct{}{}
	for _, value := range values {
		for _, item := range itemsFor(value) {
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			items = append(items, item)
		}
	}
	return items
}

func persistenceWebJobModePhrase(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "continuous":
		return "continuous path"
	case "scheduled":
		return "scheduled path"
	case "triggered/manual":
		return "triggered or manual path"
	default:
		return "rerun path"
	}
}

func persistenceFunctionsRanksBefore(left, right models.PersistenceFunctionApp) bool {
	leftHasExecRole := left.CurrentState.StrongestVisibleExecutionContext != nil && len(left.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	rightHasExecRole := right.CurrentState.StrongestVisibleExecutionContext != nil && len(right.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	leftPublic := strings.EqualFold(valueOrEmpty(left.CurrentState.PublicNetworkAccess), "Enabled") || strings.TrimSpace(valueOrEmpty(left.CurrentState.Hostname)) != ""
	rightPublic := strings.EqualFold(valueOrEmpty(right.CurrentState.PublicNetworkAccess), "Enabled") || strings.TrimSpace(valueOrEmpty(right.CurrentState.Hostname)) != ""
	leftHasDeployment := strings.TrimSpace(valueOrEmpty(left.CurrentState.Deployment)) != ""
	rightHasDeployment := strings.TrimSpace(valueOrEmpty(right.CurrentState.Deployment)) != ""
	switch {
	case leftHasExecRole != rightHasExecRole:
		return leftHasExecRole
	case leftPublic != rightPublic:
		return leftPublic
	case leftHasDeployment != rightHasDeployment:
		return leftHasDeployment
	default:
		return left.Name < right.Name
	}
}

func persistenceFunctionsInventoryRows(apps []models.PersistenceFunctionApp) [][]string {
	rows := make([][]string, 0, len(apps))
	for _, app := range apps {
		rows = append(rows, []string{
			app.Name,
			app.ResourceGroup,
			persistenceFunctionsInventoryState(app),
			persistenceFunctionsInventoryExecutionContext(app),
		})
	}
	return rows
}

func persistenceFunctionsInventoryState(app models.PersistenceFunctionApp) string {
	parts := []string{}
	if state := strings.TrimSpace(valueOrEmpty(app.CurrentState.State)); state != "" {
		parts = append(parts, state)
	}
	if strings.TrimSpace(valueOrEmpty(app.CurrentState.Hostname)) != "" {
		parts = append(parts, "hostname visible")
	}
	if runtime := strings.TrimSpace(valueOrEmpty(app.CurrentState.Runtime)); runtime != "" {
		parts = append(parts, runtime)
	}
	if deployment := strings.TrimSpace(valueOrEmpty(app.CurrentState.Deployment)); deployment != "" {
		parts = append(parts, deployment)
	}
	if network := strings.TrimSpace(valueOrEmpty(app.CurrentState.PublicNetworkAccess)); network != "" {
		parts = append(parts, "public "+network)
	}
	if len(app.CurrentState.TriggerTypes) > 0 {
		parts = append(parts, "triggers="+strings.Join(app.CurrentState.TriggerTypes, ", "))
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return strings.Join(parts, "; ")
}

func persistenceFunctionsInventoryExecutionContext(app models.PersistenceFunctionApp) string {
	if ctx := app.CurrentState.StrongestVisibleExecutionContext; ctx != nil {
		if len(ctx.RoleNames) > 0 {
			return persistenceRoleContextLabel(*ctx)
		}
		return persistenceRoleContextLine(ctx)
	}
	if len(app.ExecutionContextOptions) > 0 {
		return strings.Join(app.ExecutionContextOptions, ", ")
	}
	return "none visible"
}

func persistenceFunctionsCombinedStillUnmapped(apps []models.PersistenceFunctionApp) []string {
	items := []string{}
	seen := map[string]struct{}{}
	for _, app := range apps {
		for _, item := range app.StillUnmapped {
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			items = append(items, item)
		}
	}
	return items
}

func persistenceFunctionsBoundarySections(apps []models.PersistenceFunctionApp) ([]string, []string) {
	defaultItems := []string{}
	currentGapItems := []string{}
	seenDefault := map[string]struct{}{}
	seenGap := map[string]struct{}{}
	for _, item := range persistenceFunctionsCombinedStillUnmapped(apps) {
		if strings.HasPrefix(item, "attached user-assigned identities are visible on this Function App") {
			if _, ok := seenGap[item]; ok {
				continue
			}
			seenGap[item] = struct{}{}
			currentGapItems = append(currentGapItems, item)
			continue
		}
		if _, ok := seenDefault[item]; ok {
			continue
		}
		seenDefault[item] = struct{}{}
		defaultItems = append(defaultItems, item)
	}
	return defaultItems, currentGapItems
}

func persistenceAzureMLExplanation(workspace models.PersistenceAzureMLWorkspace) string {
	visibilityLines := []string{"  " + persistenceAzureMLVisibilityLine(workspace)}
	lines := []string{persistenceAzureMLWorkspaceBullet(workspace)}
	if persistenceCapabilityStatus(workspace.CapabilitySteps, "create or modify workspace") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workspace.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAzureMLWorkspaceWalkthrough(workspace)...)

	lines = append(lines, persistenceAzureMLComputeBullet(workspace))
	if persistenceCapabilityStatus(workspace.CapabilitySteps, "attach or reuse compute") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workspace.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAzureMLComputeWalkthrough(workspace)...)

	lines = append(lines, persistenceAzureMLCodeBullet(workspace))
	if persistenceCapabilityStatus(workspace.CapabilitySteps, "add or modify jobs or pipelines") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workspace.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAzureMLCodeWalkthrough(workspace)...)

	lines = append(lines, persistenceAzureMLExecutionContextBullet(workspace))
	if persistenceCapabilityStatus(workspace.CapabilitySteps, "attach or reuse exec ctx") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workspace.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAzureMLExecutionContextWalkthrough(workspace)...)

	lines = append(lines, persistenceAzureMLScheduleBullet(workspace))
	if persistenceCapabilityStatus(workspace.CapabilitySteps, "create or modify schedule") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workspace.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAzureMLScheduleWalkthrough(workspace)...)

	lines = append(lines, persistenceAzureMLEndpointBullet(workspace))
	if persistenceCapabilityStatus(workspace.CapabilitySteps, "expose or reuse endpoint") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workspace.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAzureMLEndpointWalkthrough(workspace)...)

	lines = append(lines, persistenceAzureMLRepurposeBullet(workspace))
	if persistenceCapabilityStatus(workspace.CapabilitySteps, "create or modify workspace") != "yes" {
		return persistenceTruncatedWalkthrough(lines, visibilityLines, workspace.CurrentState.NearbyThematicNames)
	}
	lines = append(lines, persistenceAzureMLRepurposeWalkthrough(workspace)...)
	if nearby := persistenceAutomationNearbyNamesLine(workspace.CurrentState.NearbyThematicNames); nearby != "" {
		lines = append(lines, "  "+nearby)
	}
	return renderPersistenceWalkthrough(lines)
}

func persistenceAzureMLWorkspaceBullet(workspace models.PersistenceAzureMLWorkspace) string {
	switch persistenceCapabilityStatus(workspace.CapabilitySteps, "create or modify workspace") {
	case "yes":
		return "- Current identity can create a new Azure ML workspace or reuse this existing workspace."
	case "not proven":
		return "- Current identity does not have a proven path to create or modify this Azure ML workspace from current RBAC evidence."
	default:
		return ""
	}
}

func persistenceAzureMLWorkspaceWalkthrough(workspace models.PersistenceAzureMLWorkspace) []string {
	return []string{
		"  This is the Azure ML object you would reuse instead of standing up a new ML path from scratch.",
	}
}

func persistenceAzureMLComputeBullet(workspace models.PersistenceAzureMLWorkspace) string {
	switch persistenceCapabilityStatus(workspace.CapabilitySteps, "attach or reuse compute") {
	case "yes":
		return "- Current identity can attach or reuse Azure ML compute for this workspace, including long-lived instances or cluster-backed execution."
	case "not proven":
		return "- Current identity does not have a proven path to attach or reuse Azure ML compute in this workspace from current RBAC evidence."
	default:
		return ""
	}
}

func persistenceAzureMLComputeWalkthrough(workspace models.PersistenceAzureMLWorkspace) []string {
	detail := []string{
		"  Compute is the runtime side of Azure ML that gives saved notebooks, jobs, or pipelines a place to run later.",
		"  Long-lived compute instances can stay behind, while cluster-backed execution can launch saved jobs or pipelines again when Azure ML re-triggers them.",
	}
	return detail
}

func persistenceAzureMLCodeBullet(workspace models.PersistenceAzureMLWorkspace) string {
	switch persistenceCapabilityStatus(workspace.CapabilitySteps, "add or modify jobs or pipelines") {
	case "yes":
		return "- Current identity can add or modify Azure ML jobs or pipelines that hold stored execution logic this workspace can run later."
	case "not proven":
		return "- Current identity does not have a proven path to add or modify Azure ML jobs or pipelines in this workspace from current RBAC evidence."
	default:
		return ""
	}
}

func persistenceAzureMLCodeWalkthrough(workspace models.PersistenceAzureMLWorkspace) []string {
	detail := []string{
		"  In Azure ML, persistence can live in saved notebooks, jobs, pipelines, scheduled jobs, and environment definitions.",
		"  Notebooks are interactive code surfaces, while jobs and pipelines are the scheduled or triggered execution surfaces.",
		"  Those are the stored execution surfaces that can remain in the workspace even when no host is persistently compromised.",
	}
	return detail
}

func persistenceAzureMLScheduleBullet(workspace models.PersistenceAzureMLWorkspace) string {
	switch persistenceCapabilityStatus(workspace.CapabilitySteps, "create or modify schedule") {
	case "yes":
		return "- Current identity can create or modify Azure ML schedules so this workspace can run again later."
	case "not proven":
		return "- Current identity does not have a proven path to create or modify Azure ML schedules in this workspace from current RBAC evidence."
	default:
		return ""
	}
}

func persistenceAzureMLScheduleWalkthrough(workspace models.PersistenceAzureMLWorkspace) []string {
	detail := []string{}
	if len(workspace.CurrentState.ScheduleTriggerTypes) > 0 {
		detail = append(detail, "  Visible schedule trigger types here include "+strings.Join(workspace.CurrentState.ScheduleTriggerTypes, ", ")+".")
	} else {
		detail = append(detail, "  Schedules are the clearest Azure-native re-entry anchor in Azure ML when jobs or pipelines need to run again later.")
	}
	detail = append(detail, "  A saved schedule can re-run that stored execution path later on the attached compute without requiring a compromised host to stay resident.")
	return detail
}

func persistenceAzureMLEndpointBullet(workspace models.PersistenceAzureMLWorkspace) string {
	switch persistenceCapabilityStatus(workspace.CapabilitySteps, "expose or reuse endpoint") {
	case "yes":
		return "- Current identity can expose or reuse Azure ML online endpoints as a serving or externally reachable re-entry path."
	case "not proven":
		return "- Current identity does not have a proven path to expose or reuse Azure ML online endpoints in this workspace from current RBAC evidence."
	default:
		return ""
	}
}

func persistenceAzureMLEndpointWalkthrough(workspace models.PersistenceAzureMLWorkspace) []string {
	detail := []string{}
	if len(workspace.CurrentState.EndpointAuthModes) > 0 || len(workspace.CurrentState.EndpointPublicAccess) > 0 {
		parts := []string{}
		if len(workspace.CurrentState.EndpointAuthModes) > 0 {
			parts = append(parts, "auth "+strings.Join(workspace.CurrentState.EndpointAuthModes, ", "))
		}
		if len(workspace.CurrentState.EndpointPublicAccess) > 0 {
			parts = append(parts, "public access "+strings.Join(workspace.CurrentState.EndpointPublicAccess, ", "))
		}
		detail = append(detail, "  Visible endpoint posture here includes "+strings.Join(parts, "; ")+".")
	}
	detail = append(detail, "  API-driven or endpoint-driven paths are another way Azure ML can be re-entered later, separate from saved schedules.")
	return detail
}

func persistenceAzureMLExecutionContextBullet(workspace models.PersistenceAzureMLWorkspace) string {
	switch persistenceCapabilityStatus(workspace.CapabilitySteps, "attach or reuse exec ctx") {
	case "yes":
		return "- Current identity can attach or reuse execution context for this Azure ML workspace."
	case "not proven":
		return "- Current identity does not have a proven path to attach or reuse execution context for this Azure ML workspace from current RBAC evidence."
	default:
		return ""
	}
}

func persistenceAzureMLExecutionContextWalkthrough(workspace models.PersistenceAzureMLWorkspace) []string {
	detail := []string{
		"  When a notebook, job, or pipeline runs later, it executes with the attached identity plus the linked workspace resources Azure ML will use at runtime.",
		"  That execution context determines what Azure actions, storage access, and downstream calls the re-triggered path can actually make.",
	}
	if len(workspace.ExecutionContextOptions) > 0 {
		detail = append(detail, "  Linked workspace context here already includes "+strings.Join(workspace.ExecutionContextOptions, ", ")+".")
	}
	if context := workspace.CurrentState.StrongestVisibleExecutionContext; context != nil {
		detail = append(detail, "  "+context.Summary)
	}
	return detail
}

func persistenceAzureMLRepurposeBullet(workspace models.PersistenceAzureMLWorkspace) string {
	switch persistenceCapabilityStatus(workspace.CapabilitySteps, "create or modify workspace") {
	case "yes":
		return "- Current identity can repurpose an existing Azure ML workspace instead of standing up a brand-new ML control path."
	case "not proven":
		return ""
	default:
		return ""
	}
}

func persistenceAzureMLRepurposeWalkthrough(workspace models.PersistenceAzureMLWorkspace) []string {
	return []string{
		"  Reusing an existing Azure ML workspace can blend in better than creating a new ML runtime surface from scratch.",
		"  The persistence story here is the workspace plus compute plus stored code and re-entry paths that can all remain in place for later execution.",
	}
}

func persistenceAzureMLVisibilityLine(workspace models.PersistenceAzureMLWorkspace) string {
	state := strings.TrimSpace(persistenceAzureMLInventoryState(workspace))
	if state == "" {
		return ""
	}
	return "- Current scope still shows Azure ML runtime posture here: " + state + "."
}

func persistenceAzureMLLeadWorkspace(workspaces []models.PersistenceAzureMLWorkspace) models.PersistenceAzureMLWorkspace {
	lead := workspaces[0]
	for _, candidate := range workspaces[1:] {
		if persistenceAzureMLRanksBefore(candidate, lead) {
			lead = candidate
		}
	}
	return lead
}

func persistenceAzureMLRanksBefore(left, right models.PersistenceAzureMLWorkspace) bool {
	leftRank := persistenceAzureMLClassificationRank(left.CurrentState.Classification)
	rightRank := persistenceAzureMLClassificationRank(right.CurrentState.Classification)
	leftHasExecRole := left.CurrentState.StrongestVisibleExecutionContext != nil && len(left.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	rightHasExecRole := right.CurrentState.StrongestVisibleExecutionContext != nil && len(right.CurrentState.StrongestVisibleExecutionContext.RoleNames) > 0
	if leftRank != rightRank {
		return leftRank < rightRank
	}
	if leftHasExecRole != rightHasExecRole {
		return leftHasExecRole
	}
	if intPtrValue(left.CurrentState.ScheduleCount) != intPtrValue(right.CurrentState.ScheduleCount) {
		return intPtrValue(left.CurrentState.ScheduleCount) > intPtrValue(right.CurrentState.ScheduleCount)
	}
	if intPtrValue(left.CurrentState.JobCount) != intPtrValue(right.CurrentState.JobCount) {
		return intPtrValue(left.CurrentState.JobCount) > intPtrValue(right.CurrentState.JobCount)
	}
	if intPtrValue(left.CurrentState.ComputeCount) != intPtrValue(right.CurrentState.ComputeCount) {
		return intPtrValue(left.CurrentState.ComputeCount) > intPtrValue(right.CurrentState.ComputeCount)
	}
	if intPtrValue(left.CurrentState.EndpointCount) != intPtrValue(right.CurrentState.EndpointCount) {
		return intPtrValue(left.CurrentState.EndpointCount) > intPtrValue(right.CurrentState.EndpointCount)
	}
	return left.Name < right.Name
}

func persistenceAzureMLClassificationRank(classification string) int {
	switch classification {
	case "execution-capable":
		return 0
	case "supporting-persistence-context":
		return 1
	default:
		return 2
	}
}

func persistenceAzureMLInventoryRows(workspaces []models.PersistenceAzureMLWorkspace) [][]string {
	rows := make([][]string, 0, len(workspaces))
	for _, workspace := range workspaces {
		rows = append(rows, []string{
			workspace.Name,
			workspace.ResourceGroup,
			persistenceAzureMLInventoryState(workspace),
			persistenceAzureMLInventoryExecutionContext(workspace),
		})
	}
	return rows
}

func persistenceAzureMLInventoryState(workspace models.PersistenceAzureMLWorkspace) string {
	parts := []string{}
	if classification := strings.TrimSpace(workspace.CurrentState.Classification); classification != "" {
		parts = append(parts, classification)
	}
	if len(workspace.CurrentState.ComputeTypes) > 0 {
		parts = append(parts, "compute="+strings.Join(workspace.CurrentState.ComputeTypes, ","))
	} else if count := intPtrString(workspace.CurrentState.ComputeCount); count != "" && count != "0" {
		parts = append(parts, "compute "+count)
	}
	if count := intPtrString(workspace.CurrentState.JobCount); count != "" && count != "0" {
		parts = append(parts, "jobs "+count)
	}
	if count := intPtrString(workspace.CurrentState.ScheduleCount); count != "" && count != "0" {
		parts = append(parts, "schedules "+count)
	}
	if count := intPtrString(workspace.CurrentState.EndpointCount); count != "" && count != "0" {
		parts = append(parts, "endpoints "+count)
	}
	if len(workspace.CurrentState.EndpointPublicAccess) > 0 {
		parts = append(parts, "endpoint public "+strings.Join(workspace.CurrentState.EndpointPublicAccess, ","))
	}
	if network := strings.TrimSpace(valueOrEmpty(workspace.CurrentState.PublicNetworkAccess)); network != "" {
		parts = append(parts, "workspace public "+network)
	}
	return strings.Join(parts, "; ")
}

func persistenceAzureMLInventoryExecutionContext(workspace models.PersistenceAzureMLWorkspace) string {
	if ctx := workspace.CurrentState.StrongestVisibleExecutionContext; ctx != nil {
		if len(ctx.RoleNames) > 0 {
			return persistenceRoleContextLabel(*ctx)
		}
		return persistenceRoleContextLine(ctx)
	}
	if len(workspace.ExecutionContextOptions) > 0 {
		return strings.Join(workspace.ExecutionContextOptions, ", ")
	}
	return "-"
}

func persistenceAzureMLCombinedStillUnmapped(workspaces []models.PersistenceAzureMLWorkspace) []string {
	seen := map[string]struct{}{}
	items := []string{}
	for _, workspace := range workspaces {
		for _, item := range workspace.StillUnmapped {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			items = append(items, trimmed)
		}
	}
	return items
}

func persistenceAzureMLBoundarySections(workspaces []models.PersistenceAzureMLWorkspace) ([]string, []string) {
	defaultItems := []string{}
	currentGapItems := []string{}
	for _, item := range persistenceAzureMLCombinedStillUnmapped(workspaces) {
		if strings.Contains(item, "current output does not yet resolve") {
			currentGapItems = append(currentGapItems, item)
			continue
		}
		defaultItems = append(defaultItems, item)
	}
	return defaultItems, currentGapItems
}

func persistenceFunctionsVisibleFunctionSummary(values []models.FunctionChildAsset) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		name := strings.TrimSpace(value.Name)
		if name == "" {
			continue
		}
		details := []string{}
		if value.TriggerType != nil && strings.TrimSpace(*value.TriggerType) != "" {
			details = append(details, *value.TriggerType)
		}
		if value.IsDisabled != nil && *value.IsDisabled {
			details = append(details, "disabled")
		}
		if len(details) > 0 {
			name += " [" + strings.Join(details, "; ") + "]"
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, ", ")
}

func persistenceFunctionsTriggerBoundaryLine(values []models.FunctionChildAsset) string {
	httpVisible := false
	httpWithInvokeURL := false
	httpWithAuthLevel := false
	internalTriggers := []string{}
	for _, value := range values {
		triggerType := strings.TrimSpace(valueOrEmpty(value.TriggerType))
		switch {
		case strings.EqualFold(triggerType, "HTTP"):
			httpVisible = true
			if strings.TrimSpace(valueOrEmpty(value.InvokeURLTemplate)) != "" {
				httpWithInvokeURL = true
			}
			if authLevel := strings.TrimSpace(mapStringValueFromAny(value.Config, "authLevel")); authLevel != "" {
				httpWithAuthLevel = true
			}
		case triggerType != "":
			internalTriggers = append(internalTriggers, triggerType)
		}
	}

	parts := []string{}
	if httpVisible {
		httpDetail := "  HTTP-triggered functions are visible from management-plane metadata"
		if httpWithInvokeURL {
			httpDetail += ", including an invoke URL template"
		}
		if httpWithAuthLevel {
			httpDetail += " and visible auth-level metadata"
		}
		httpDetail += "."
		parts = append(parts, httpDetail)
	}
	if len(internalTriggers) > 0 {
		parts = append(parts, "  Timer, queue, Service Bus, or other event-driven triggers are visible from bindings, but they are not the same as a directly callable public entrypoint.")
	}
	return strings.Join(parts, "\n")
}

func mapStringValueFromAny(values map[string]any, key string) string {
	rawBindings, ok := values["bindings"].([]any)
	if !ok {
		return ""
	}
	for _, raw := range rawBindings {
		binding, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		rawType, ok := binding["type"].(string)
		if !ok || !strings.EqualFold(strings.TrimSpace(rawType), "httpTrigger") {
			continue
		}
		rawValue, ok := binding[key].(string)
		if !ok {
			continue
		}
		if value := strings.TrimSpace(rawValue); value != "" {
			return value
		}
	}
	return ""
}

func renderNaturalJoin(values []string) string {
	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	case 2:
		return values[0] + " and " + values[1]
	default:
		prefix := strings.Join(values[:len(values)-1], ", ")
		return prefix + ", and " + values[len(values)-1]
	}
}

func quoteInlineValues(values []string) []string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		quoted = append(quoted, "`"+trimmed+"`")
	}
	return quoted
}

func renderScopeName(scopeID string) string {
	parts := strings.Split(scopeID, "/")
	for index := 0; index < len(parts)-1; index++ {
		if strings.EqualFold(parts[index], "resourceGroups") {
			return parts[index+1]
		}
	}
	return "unknown"
}
