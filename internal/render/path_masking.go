package render

import (
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

func pathMaskingTableRenderer(payload any) (string, error) {
	switch out := payload.(type) {
	case models.PathMaskingOverviewOutput:
		return pathMaskingOverviewTable(out), nil
	case models.PathMaskingAPIMOutput:
		return pathMaskingAPIMTable(out), nil
	case models.PathMaskingLogicAppsOutput:
		return pathMaskingLogicAppsTable(out), nil
	case models.PathMaskingRelayOutput:
		return pathMaskingRelayTable(out), nil
	default:
		return "", fmt.Errorf("unexpected payload type for pathmasking: %T", payload)
	}
}

func pathMaskingOverviewTable(payload models.PathMaskingOverviewOutput) string {
	rows := make([][]string, 0, len(payload.Surfaces))
	for _, surface := range payload.Surfaces {
		rows = append(rows, []string{surface.Surface, surface.Summary})
	}
	return renderListTable(
		"ho-azure pathmasking",
		[]string{"surface", "summary"},
		rows,
		[]string{"no pathmasking surfaces available", ""},
		pathMaskingOverviewTakeaway(payload),
	)
}

func pathMaskingAPIMTable(payload models.PathMaskingAPIMOutput) string {
	if len(payload.Targets) == 0 {
		return renderFamilySurfaceTable(familySurfaceTableConfig{
			Title:         "ho-azure pathmasking api-mgmt",
			EmptyHeaders:  []string{"api management service", "status"},
			EmptyRow:      []string{"No visible API Management services were confirmed from current scope.", ""},
			EmptyTakeaway: "0 APIM services visible; no APIM pathmasking surface was confirmed from current scope.",
		})
	}

	lead := payload.Targets[0]
	return renderFamilySurfaceTable(familySurfaceTableConfig{
		Title:             "ho-azure pathmasking api-mgmt",
		CapabilityTitle:   "APIM pathmasking capability",
		CapabilitySteps:   lead.CapabilitySteps,
		MultiTargetNote:   "This walkthrough shows the strongest currently visible APIM pathmasking posture. The inventory below lists the other visible services without repeating the same narrative.",
		TargetCount:       len(payload.Targets),
		Explanation:       pathMaskingAPIMExplanation(lead),
		ReducedVisibility: familyReducedVisibilityExplanation("APIM gateway and backend-indirection", "APIM management-plane", "pathmasking", lead.CurrentIdentityContext),
		InventoryTitle:    "Visible APIM Services",
		InventoryHeaders:  []string{"service", "rank", "gateways", "backends", "subscriptions", "current identity"},
		InventoryRows:     pathMaskingAPIMInventoryRows(payload.Targets),
		BoundaryNotes:     lead.NotCollectedByDefault,
	})
}

func pathMaskingLogicAppsTable(payload models.PathMaskingLogicAppsOutput) string {
	if len(payload.Targets) == 0 {
		return renderFamilySurfaceTable(familySurfaceTableConfig{
			Title:         "ho-azure pathmasking logic-apps",
			EmptyHeaders:  []string{"logic app", "status"},
			EmptyRow:      []string{"No visible Logic Apps were confirmed from current scope.", ""},
			EmptyTakeaway: "0 Logic Apps visible; no Logic Apps pathmasking surface was confirmed from current scope.",
		})
	}

	lead := payload.Targets[0]
	return renderFamilySurfaceTable(familySurfaceTableConfig{
		Title:             "ho-azure pathmasking logic-apps",
		CapabilityTitle:   "Logic Apps pathmasking capability",
		CapabilitySteps:   lead.CapabilitySteps,
		MultiTargetNote:   "This walkthrough shows the strongest currently visible Logic App relay path. The inventory below lists the other visible workflows without repeating the same narrative.",
		TargetCount:       len(payload.Targets),
		Explanation:       pathMaskingLogicAppsExplanation(lead),
		ReducedVisibility: familyReducedVisibilityExplanation("Logic App path-shaping", "Logic Apps management-plane", "pathmasking", lead.CurrentIdentityContext),
		InventoryTitle:    "Visible Logic Apps",
		InventoryHeaders:  []string{"workflow", "rank", "triggers", "actions", "identity", "current identity"},
		InventoryRows:     pathMaskingLogicAppsInventoryRows(payload.Targets),
		BoundaryNotes:     lead.NotCollectedByDefault,
	})
}

func pathMaskingRelayTable(payload models.PathMaskingRelayOutput) string {
	if len(payload.Targets) == 0 {
		return renderFamilySurfaceTable(familySurfaceTableConfig{
			Title:         "ho-azure pathmasking relay",
			EmptyHeaders:  []string{"relay namespace", "status"},
			EmptyRow:      []string{"No visible Relay namespaces were confirmed from current scope.", ""},
			EmptyTakeaway: "0 Relay namespaces visible; no Relay pathmasking surface was confirmed from current scope.",
		})
	}

	lead := payload.Targets[0]
	return renderFamilySurfaceTable(familySurfaceTableConfig{
		Title:             "ho-azure pathmasking relay",
		CapabilityTitle:   "Relay pathmasking capability",
		CapabilitySteps:   lead.CapabilitySteps,
		MultiTargetNote:   "This walkthrough shows the strongest currently visible Relay private-path posture. The inventory below lists the other visible namespaces without repeating the same narrative.",
		TargetCount:       len(payload.Targets),
		Explanation:       pathMaskingRelayExplanation(lead),
		ReducedVisibility: familyReducedVisibilityExplanation("Relay namespace and Hybrid Connection", "Relay management-plane", "pathmasking", lead.CurrentIdentityContext),
		InventoryTitle:    "Visible Relay Namespaces",
		InventoryHeaders:  []string{"namespace", "rank", "hybrid connections", "auth rules", "listeners", "current identity"},
		InventoryRows:     pathMaskingRelayInventoryRows(payload.Targets),
		BoundaryNotes:     lead.NotCollectedByDefault,
	})
}

func pathMaskingAPIMExplanation(target models.PathMaskingAPIMTarget) string {
	lines := []string{
		"",
		"Operator read",
		target.Summary,
		"Current identity: " + familyRoleSummary(target.CurrentIdentityContext),
		"Downstream effect: " + target.MaskingReason,
		"First boundary: this is APIM management-plane posture, not policy-body proof, live request proof, or backend ownership proof.",
		"Posture: " + target.CurrentState.Posture + ".",
	}
	return strings.Join(lines, "\n")
}

func pathMaskingLogicAppsExplanation(target models.PathMaskingLogicAppTarget) string {
	lines := []string{
		"",
		"Operator read",
		target.Summary,
		"Current identity: " + familyRoleSummary(target.CurrentIdentityContext),
		"Downstream effect: " + target.MaskingReason,
		"First boundary: this is Logic App management-plane posture, not run-history proof, connector payload proof, or credential-material proof.",
		"Posture: " + target.CurrentState.Posture + ".",
	}
	return strings.Join(lines, "\n")
}

func pathMaskingRelayExplanation(target models.PathMaskingRelayTarget) string {
	lines := []string{
		"",
		"Operator read",
		target.Summary,
		"Current identity: " + familyRoleSummary(target.CurrentIdentityContext),
		"Downstream effect: " + target.MaskingReason,
		"First boundary: this is Relay management-plane posture, not listener-runtime proof, backend process proof, or traffic-content proof.",
		"Posture: " + target.CurrentState.Posture + ".",
	}
	return strings.Join(lines, "\n")
}

func pathMaskingAPIMInventoryRows(targets []models.PathMaskingAPIMTarget) [][]string {
	rows := make([][]string, 0, len(targets))
	for _, target := range targets {
		rows = append(rows, []string{
			target.Name,
			fmt.Sprintf("%d/5", target.MaskingRank),
			joinOrNone(target.CurrentState.GatewayHostnames),
			joinOrNone(target.CurrentState.BackendHostnames),
			intPtrString(target.CurrentState.SubscriptionCount),
			familyRoleControlLabel(target.CurrentIdentityContext),
		})
	}
	return rows
}

func pathMaskingLogicAppsInventoryRows(targets []models.PathMaskingLogicAppTarget) [][]string {
	rows := make([][]string, 0, len(targets))
	for _, target := range targets {
		rows = append(rows, []string{
			target.Name,
			fmt.Sprintf("%d/5", target.MaskingRank),
			joinOrNone(target.CurrentState.TriggerTypes),
			joinOrNone(target.CurrentState.DownstreamActionKinds),
			valueOrEmpty(target.CurrentState.IdentityType),
			familyRoleControlLabel(target.CurrentIdentityContext),
		})
	}
	return rows
}

func pathMaskingRelayInventoryRows(targets []models.PathMaskingRelayTarget) [][]string {
	rows := make([][]string, 0, len(targets))
	for _, target := range targets {
		rows = append(rows, []string{
			target.Name,
			fmt.Sprintf("%d/5", target.MaskingRank),
			intPtrString(target.CurrentState.HybridConnectionCount),
			intPtrString(target.CurrentState.AuthorizationRuleCount),
			target.CurrentState.ListenerSummary,
			familyRoleControlLabel(target.CurrentIdentityContext),
		})
	}
	return rows
}

func pathMaskingOverviewTakeaway(payload models.PathMaskingOverviewOutput) string {
	return fmt.Sprintf("%d pathmasking surface(s) available; run a surface to rank visible posture by path ambiguity and attribution-blur value.", len(payload.Surfaces))
}
