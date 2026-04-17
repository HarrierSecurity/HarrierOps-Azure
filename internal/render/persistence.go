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
			surface.State,
			surface.OperatorQuestion,
			strings.Join(surface.BackingCommands, ", "),
		})
	}
	return renderListTable(
		"ho-azure persistence",
		[]string{"surface", "state", "operator question", "backing commands"},
		rows,
		[]string{"no persistence surfaces implemented", "", "", ""},
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
		renderAlignedPipeTable(
			[]string{"action", "api surface", "status"},
			persistenceAutomationCapabilityRows(lead.CapabilitySteps),
		),
	}
	if len(payload.AutomationAccounts) > 1 {
		lines = append(lines, "This walkthrough shows one currently visible Automation persistence path. The inventory below lists the other visible accounts without repeating the same narrative.")
	}
	lines = append(lines,
		persistenceAutomationExplanation(lead),
		"",
		"Reminder: a runbook does not run continuously and is not a backdoor listening on a port. In this context, persistence means Azure stores code plus execution context plus a trigger that can invoke it again later.",
		"",
		"Visible Automation Accounts",
		renderAlignedPipeTable(
			[]string{"automation account", "resource group", "visible state", "execution context"},
			persistenceAutomationInventoryRows(payload.AutomationAccounts),
		),
	)
	if unmapped := persistenceAutomationCombinedStillUnmapped(payload.AutomationAccounts); unmapped != "" {
		lines = append(lines, "", "Still unmapped", unmapped)
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
		"Workflow capability",
		renderAlignedPipeTable(
			[]string{"action", "api surface", "status"},
			persistenceLogicAppCapabilityRows(lead.CapabilitySteps),
		),
	}
	if len(payload.Workflows) > 1 {
		lines = append(lines, "This walkthrough shows one currently visible Logic App persistence path. The inventory below lists the other visible workflows without repeating the same narrative.")
	}
	lines = append(lines,
		persistenceLogicAppExplanation(lead),
		"",
		"Reminder: Logic App persistence is about a stored workflow plus a trigger plus access that remains valid, not malware living on a host.",
		"",
		"Visible Logic Apps",
		renderAlignedPipeTable(
			[]string{"logic app", "resource group", "visible state", "execution context"},
			persistenceLogicAppInventoryRows(payload.Workflows),
		),
	)
	if items := persistenceLogicAppCombinedStillUnmapped(payload.Workflows); len(items) > 0 {
		lines = append(lines, "", "Still unmapped", renderBulletList(items))
	}

	return strings.Join(lines, "\n")
}

func persistenceOverviewTakeaway(payload models.PersistenceOverviewOutput) string {
	if len(payload.Surfaces) == 0 {
		return "No persistence surfaces are implemented yet."
	}
	if len(payload.Surfaces) == 1 {
		return fmt.Sprintf("1 persistence surface is implemented; start with %s.", payload.Surfaces[0].Surface)
	}
	return fmt.Sprintf("%d persistence surfaces are implemented; start with the service that best matches your current question.", len(payload.Surfaces))
}

func persistenceAutomationCapabilityRows(steps []models.PersistenceCapabilityStep) [][]string {
	rows := make([][]string, 0, len(steps))
	for _, step := range steps {
		rows = append(rows, []string{step.Action, step.APISurface, step.Status})
	}
	return rows
}

func persistenceLogicAppCapabilityRows(steps []models.PersistenceCapabilityStep) [][]string {
	rows := make([][]string, 0, len(steps))
	for _, step := range steps {
		rows = append(rows, []string{step.Action, step.APISurface, step.Status})
	}
	return rows
}

func renderAlignedPipeTable(headers []string, rows [][]string) string {
	widths := make([]int, len(headers))
	for index, header := range headers {
		widths[index] = len(header)
	}
	for _, row := range rows {
		for index, cell := range row {
			if index < len(widths) && len(cell) > widths[index] {
				widths[index] = len(cell)
			}
		}
	}

	renderRow := func(cells []string) string {
		parts := make([]string, len(widths))
		for index := range widths {
			cell := ""
			if index < len(cells) {
				cell = cells[index]
			}
			parts[index] = padRight(cell, widths[index])
		}
		return strings.Join(parts, " | ")
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

func padRight(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-len(value))
}

func persistenceAutomationExplanation(account models.PersistenceAutomationAccount) string {
	lines := []string{
		"- " + persistenceAutomationAccountBullet(account),
		"- " + persistenceAutomationRunbookBullet(account),
		"- " + persistenceAutomationCodeBullet(account),
		"- " + persistenceAutomationPublishBullet(account),
		"- " + persistenceAutomationExecutionContextBullet(account),
		"  Managed identity, stored credentials, connections, certificates, variables, or other Automation assets may provide that execution context.",
	}
	if account.CurrentIdentityContext != nil && strings.TrimSpace(account.CurrentIdentityContext.Summary) != "" {
		lines = append(lines, "  "+account.CurrentIdentityContext.Summary)
	}
	if ctx := account.CurrentState.StrongestVisibleExecutionContext; ctx != nil && strings.TrimSpace(ctx.Summary) != "" {
		lines = append(lines, "  "+ctx.Summary)
	}
	lines = append(lines, "- "+persistenceAutomationTriggerBullet(account))
	lines = append(lines, "- "+persistenceAutomationRepurposeBullet(account))
	if nearby := persistenceAutomationNearbyNamesLine(account.CurrentState.NearbyThematicNames); nearby != "" {
		lines = append(lines, "  "+nearby)
	}
	return strings.Join(lines, "\n")
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
	lines := []string{
		"- " + persistenceLogicAppWorkflowBullet(workflow),
		"- " + persistenceLogicAppDefinitionBullet(workflow),
		"- " + persistenceLogicAppExecutionContextBullet(workflow),
		"  Managed identity or connector-backed actions may provide that execution context.",
		"  In Logic Apps, the payload is the stored workflow definition and action chain Azure will execute later.",
	}
	if workflow.CurrentIdentityContext != nil && strings.TrimSpace(workflow.CurrentIdentityContext.Summary) != "" {
		lines = append(lines, "  "+workflow.CurrentIdentityContext.Summary)
	}
	if ctx := workflow.CurrentState.StrongestVisibleExecutionContext; ctx != nil && strings.TrimSpace(ctx.Summary) != "" {
		lines = append(lines, "  "+ctx.Summary)
	}
	lines = append(lines,
		"- "+persistenceLogicAppTriggerBullet(workflow),
		"- "+persistenceLogicAppEnableBullet(workflow),
		"- "+persistenceLogicAppActionBullet(workflow),
		"- "+persistenceLogicAppRepurposeBullet(workflow),
	)
	if nearby := persistenceAutomationNearbyNamesLine(workflow.CurrentState.NearbyThematicNames); nearby != "" {
		lines = append(lines, "  "+nearby)
	}
	return strings.Join(lines, "\n")
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
	return fmt.Sprintf("`%s` with no confirmed Azure role context", context.Name)
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

func renderScopeName(scopeID string) string {
	parts := strings.Split(scopeID, "/")
	for index := 0; index < len(parts)-1; index++ {
		if strings.EqualFold(parts[index], "resourceGroups") {
			return parts[index+1]
		}
	}
	return "unknown"
}
