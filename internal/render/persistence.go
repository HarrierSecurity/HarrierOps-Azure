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

	sections := make([]string, 0, len(payload.AutomationAccounts))
	for _, account := range payload.AutomationAccounts {
		lines := []string{
			"Automation Account: " + account.Name,
			"",
			"Automation capability",
			renderAlignedPipeTable(
				[]string{"action", "api surface", "status"},
				persistenceAutomationCapabilityRows(account.CapabilitySteps),
			),
			persistenceAutomationExplanation(account),
			"",
			"Reminder: a runbook does not run continuously and is not a backdoor listening on a port. In this context, persistence means Azure stores code plus execution context plus a trigger that can invoke it again later.",
			"",
			"Current state",
			persistenceAutomationCurrentState(account),
		}
		if unmapped := persistenceAutomationStillUnmappedSection(account); unmapped != "" {
			lines = append(lines, "", "Still unmapped", unmapped)
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	return joinRenderedBlocks(sections)
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

func persistenceAutomationCurrentState(account models.PersistenceAutomationAccount) string {
	lines := []string{
		fmt.Sprintf("- runbooks: %s total, %s published", intPtrString(account.CurrentState.RunbookCount), intPtrString(account.CurrentState.PublishedRunbookCount)),
		"- published runbook names: " + persistenceJoinedOrNone(account.CurrentState.PublishedRunbookNames),
		fmt.Sprintf("- schedules: %s", intPtrString(account.CurrentState.ScheduleCount)),
		fmt.Sprintf("- job schedules: %s", intPtrString(account.CurrentState.JobScheduleCount)),
		fmt.Sprintf("- webhooks: %s", intPtrString(account.CurrentState.WebhookCount)),
		fmt.Sprintf("- primary start mode: %s", textOrNone(valueOrEmpty(account.CurrentState.PrimaryStartMode))),
		fmt.Sprintf("- primary runbook: %s", textOrNone(valueOrEmpty(account.CurrentState.PrimaryRunbookName))),
		fmt.Sprintf("- identity type: %s", textOrNone(valueOrEmpty(account.CurrentState.IdentityType))),
	}
	if len(account.ExecutionContextOptions) > 0 {
		lines = append(lines, "- execution context options: "+strings.Join(account.ExecutionContextOptions, ", "))
	}
	if ctx := account.CurrentState.StrongestVisibleExecutionContext; ctx != nil {
		lines = append(lines, "- strongest visible execution context: "+persistenceRoleContextLabel(*ctx))
	}
	lines = append(lines,
		fmt.Sprintf(
			"- secure assets: credentials %s, certificates %s, connections %s, variables %s, encrypted variables %s",
			intPtrString(account.CurrentState.CredentialCount),
			intPtrString(account.CurrentState.CertificateCount),
			intPtrString(account.CurrentState.ConnectionCount),
			intPtrString(account.CurrentState.VariableCount),
			intPtrString(account.CurrentState.EncryptedVariableCount),
		),
		fmt.Sprintf("- hybrid worker groups: %s", intPtrString(account.CurrentState.HybridWorkerGroupCount)),
	)
	return strings.Join(lines, "\n")
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

func persistenceAutomationStillUnmappedSection(account models.PersistenceAutomationAccount) string {
	if len(account.StillUnmapped) == 0 {
		return ""
	}
	lines := make([]string, 0, len(account.StillUnmapped))
	for _, item := range account.StillUnmapped {
		lines = append(lines, "- "+item)
	}
	return strings.Join(lines, "\n")
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
