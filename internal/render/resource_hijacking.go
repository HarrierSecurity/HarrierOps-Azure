package render

import (
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

func resourceHijackingTableRenderer(payload any) (string, error) {
	switch out := payload.(type) {
	case models.ResourceHijackingOverviewOutput:
		return resourceHijackingOverviewTable(out), nil
	case models.ResourceHijackingAPIMOutput:
		return resourceHijackingAPIMTable(out), nil
	case models.ResourceHijackingAutomationOutput:
		return resourceHijackingAutomationTable(out), nil
	case models.ResourceHijackingLogicAppsOutput:
		return resourceHijackingLogicAppsTable(out), nil
	default:
		return "", fmt.Errorf("unexpected payload type for resourcehijacking: %T", payload)
	}
}

func resourceHijackingAutomationTable(payload models.ResourceHijackingAutomationOutput) string {
	if len(payload.Targets) == 0 {
		return renderFamilySurfaceTable(familySurfaceTableConfig{
			Title:         "ho-azure resourcehijacking automation",
			EmptyHeaders:  []string{"automation account", "status"},
			EmptyRow:      []string{"No visible Automation accounts were confirmed from current scope.", ""},
			EmptyTakeaway: "0 Automation accounts visible; no Automation resourcehijacking surface was confirmed from current scope.",
		})
	}

	lead := payload.Targets[0]
	return renderFamilySurfaceTable(familySurfaceTableConfig{
		Title:             "ho-azure resourcehijacking automation",
		CapabilityTitle:   "Automation resourcehijacking capability",
		CapabilitySteps:   lead.CapabilitySteps,
		MultiTargetNote:   "This walkthrough shows the strongest currently visible Automation takeover path. The inventory below lists the other visible accounts without repeating the same narrative.",
		TargetCount:       len(payload.Targets),
		Explanation:       resourceHijackingAutomationExplanation(lead),
		ReducedVisibility: familyReducedVisibilityExplanation("Automation account", "Automation management-plane", "resourcehijacking", lead.CurrentIdentityContext),
		InventoryTitle:    "Visible Automation Accounts",
		InventoryHeaders:  []string{"account", "rank", "runbooks", "job schedules", "webhooks", "current identity"},
		InventoryRows:     resourceHijackingAutomationInventoryRows(payload.Targets),
		BoundaryNotes:     lead.NotCollectedByDefault,
	})
}

func resourceHijackingLogicAppsTable(payload models.ResourceHijackingLogicAppsOutput) string {
	if len(payload.Targets) == 0 {
		return renderFamilySurfaceTable(familySurfaceTableConfig{
			Title:         "ho-azure resourcehijacking logic-apps",
			EmptyHeaders:  []string{"logic app", "status"},
			EmptyRow:      []string{"No visible Logic Apps were confirmed from current scope.", ""},
			EmptyTakeaway: "0 Logic Apps visible; no Logic Apps resourcehijacking surface was confirmed from current scope.",
		})
	}

	lead := payload.Targets[0]
	return renderFamilySurfaceTable(familySurfaceTableConfig{
		Title:             "ho-azure resourcehijacking logic-apps",
		CapabilityTitle:   "Logic Apps resourcehijacking capability",
		CapabilitySteps:   lead.CapabilitySteps,
		MultiTargetNote:   "This walkthrough shows the strongest currently visible Logic App takeover path. The inventory below lists the other visible workflows without repeating the same narrative.",
		TargetCount:       len(payload.Targets),
		Explanation:       resourceHijackingLogicAppsExplanation(lead),
		ReducedVisibility: familyReducedVisibilityExplanation("Logic App workflow", "Logic Apps management-plane", "resourcehijacking", lead.CurrentIdentityContext),
		InventoryTitle:    "Visible Logic Apps",
		InventoryHeaders:  []string{"workflow", "rank", "triggers", "actions", "identity", "current identity"},
		InventoryRows:     resourceHijackingLogicAppsInventoryRows(payload.Targets),
		BoundaryNotes:     lead.NotCollectedByDefault,
	})
}

func resourceHijackingOverviewTable(payload models.ResourceHijackingOverviewOutput) string {
	rows := make([][]string, 0, len(payload.Surfaces))
	for _, surface := range payload.Surfaces {
		rows = append(rows, []string{surface.Surface, surface.Summary})
	}
	return renderListTable(
		"ho-azure resourcehijacking",
		[]string{"surface", "summary"},
		rows,
		[]string{"no resourcehijacking surfaces available", ""},
		resourceHijackingOverviewTakeaway(payload),
	)
}

func resourceHijackingAPIMTable(payload models.ResourceHijackingAPIMOutput) string {
	if len(payload.Targets) == 0 {
		return renderFamilySurfaceTable(familySurfaceTableConfig{
			Title:         "ho-azure resourcehijacking api-mgmt",
			EmptyHeaders:  []string{"api management service", "status"},
			EmptyRow:      []string{"No visible API Management services were confirmed from current scope.", ""},
			EmptyTakeaway: "0 APIM services visible; no APIM resourcehijacking surface was confirmed from current scope.",
		})
	}

	lead := payload.Targets[0]
	return renderFamilySurfaceTable(familySurfaceTableConfig{
		Title:             "ho-azure resourcehijacking api-mgmt",
		CapabilityTitle:   "APIM resourcehijacking capability",
		CapabilitySteps:   lead.CapabilitySteps,
		MultiTargetNote:   "This walkthrough shows the strongest currently visible APIM takeover path. The inventory below lists the other visible services without repeating the same narrative.",
		TargetCount:       len(payload.Targets),
		Explanation:       resourceHijackingAPIMExplanation(lead),
		ReducedVisibility: familyReducedVisibilityExplanation("APIM service", "APIM management-plane", "resourcehijacking", lead.CurrentIdentityContext),
		InventoryTitle:    "Visible APIM Services",
		InventoryHeaders:  []string{"service", "rank", "gateways", "backends", "active subscriptions", "current identity"},
		InventoryRows:     resourceHijackingAPIMInventoryRows(payload.Targets),
		BoundaryNotes:     lead.NotCollectedByDefault,
	})
}

func resourceHijackingAPIMExplanation(target models.ResourceHijackingAPIMTarget) string {
	lines := []string{
		"",
		"Operator read",
		target.Summary,
		"Current identity: " + familyRoleSummary(target.CurrentIdentityContext),
		"Downstream effect: " + target.TakeoverReason,
		"First boundary: this is APIM management-plane posture, not policy-body proof, live traffic proof, or backend ownership proof.",
		"Posture: " + target.CurrentState.Posture + ".",
	}
	return strings.Join(lines, "\n")
}

func resourceHijackingAutomationExplanation(target models.ResourceHijackingAutomationTarget) string {
	lines := []string{
		"",
		"Operator read",
		target.Summary,
		"Current identity: " + familyRoleSummary(target.CurrentIdentityContext),
		"Downstream effect: " + target.TakeoverReason,
		"First boundary: this is Automation management-plane posture, not runbook script proof, job-output proof, or hybrid worker host proof.",
		"Posture: " + target.CurrentState.Posture + ".",
	}
	return strings.Join(lines, "\n")
}

func resourceHijackingLogicAppsExplanation(target models.ResourceHijackingLogicAppTarget) string {
	lines := []string{
		"",
		"Operator read",
		target.Summary,
		"Current identity: " + familyRoleSummary(target.CurrentIdentityContext),
		"Downstream effect: " + target.TakeoverReason,
		"First boundary: this is Logic App management-plane posture, not run-history proof, connector data proof, or secret-material proof.",
		"Posture: " + target.CurrentState.Posture + ".",
	}
	return strings.Join(lines, "\n")
}

func resourceHijackingAPIMInventoryRows(targets []models.ResourceHijackingAPIMTarget) [][]string {
	rows := make([][]string, 0, len(targets))
	for _, target := range targets {
		rows = append(rows, []string{
			target.Name,
			fmt.Sprintf("%d/5", target.TakeoverRank),
			joinOrNone(target.CurrentState.GatewayHostnames),
			joinOrNone(target.CurrentState.BackendHostnames),
			intPtrString(target.CurrentState.ActiveSubscriptionCount),
			familyRoleControlLabel(target.CurrentIdentityContext),
		})
	}
	return rows
}

func resourceHijackingAutomationInventoryRows(targets []models.ResourceHijackingAutomationTarget) [][]string {
	rows := make([][]string, 0, len(targets))
	for _, target := range targets {
		rows = append(rows, []string{
			target.Name,
			fmt.Sprintf("%d/5", target.TakeoverRank),
			intPtrString(target.CurrentState.PublishedRunbookCount),
			intPtrString(target.CurrentState.JobScheduleCount),
			intPtrString(target.CurrentState.WebhookCount),
			familyRoleControlLabel(target.CurrentIdentityContext),
		})
	}
	return rows
}

func resourceHijackingLogicAppsInventoryRows(targets []models.ResourceHijackingLogicAppTarget) [][]string {
	rows := make([][]string, 0, len(targets))
	for _, target := range targets {
		rows = append(rows, []string{
			target.Name,
			fmt.Sprintf("%d/5", target.TakeoverRank),
			joinOrNone(target.CurrentState.TriggerTypes),
			joinOrNone(target.CurrentState.DownstreamActionKinds),
			valueOrEmpty(target.CurrentState.IdentityType),
			familyRoleControlLabel(target.CurrentIdentityContext),
		})
	}
	return rows
}

func resourceHijackingOverviewTakeaway(payload models.ResourceHijackingOverviewOutput) string {
	return fmt.Sprintf("%d resourcehijacking surface(s) available; run a surface to rank visible posture by takeover value.", len(payload.Surfaces))
}
