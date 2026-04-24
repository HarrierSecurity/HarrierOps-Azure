package render

import (
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

func evasionTableRenderer(payload any) (string, error) {
	switch out := payload.(type) {
	case models.EvasionOverviewOutput:
		return evasionOverviewTable(out), nil
	case models.EvasionDCROutput:
		return evasionDCRTable(out), nil
	case models.EvasionDiagnosticSettingsOutput:
		return evasionDiagnosticSettingsTable(out), nil
	case models.EvasionAppInsightsOutput:
		return evasionAppInsightsTable(out), nil
	default:
		return "", fmt.Errorf("unexpected payload type for evasion: %T", payload)
	}
}

func evasionAppInsightsTable(payload models.EvasionAppInsightsOutput) string {
	if len(payload.Targets) == 0 {
		return renderFamilySurfaceTable(familySurfaceTableConfig{
			Title:         "ho-azure evasion appinsights",
			EmptyHeaders:  []string{"target", "status"},
			EmptyRow:      []string{"No visible instrumented App Insights targets were confirmed from current scope.", ""},
			EmptyTakeaway: "0 targets visible; no Application Insights evasion surface was confirmed from current scope.",
		})
	}
	lead := payload.Targets[0]
	return renderFamilySurfaceTable(familySurfaceTableConfig{
		Title:             "ho-azure evasion appinsights",
		CapabilityTitle:   "Application Insights evasion capability",
		CapabilitySteps:   lead.CapabilitySteps,
		MultiTargetNote:   "This walkthrough shows the strongest currently visible Application Insights truth-disruption path. The inventory below lists the other visible targets without repeating the same narrative.",
		TargetCount:       len(payload.Targets),
		Explanation:       evasionAppInsightsExplanation(lead),
		ReducedVisibility: familyReducedVisibilityExplanation("Application Insights and app telemetry", "Application Insights management-plane and app-setting", "evasion", lead.CurrentIdentityContext),
		InventoryTitle:    "Visible Targets",
		InventoryHeaders:  []string{"target", "rank", "kind", "sampling", "filtering", "current identity"},
		InventoryRows:     evasionAppInsightsInventoryRows(payload.Targets),
		BoundaryNotes:     lead.NotCollectedByDefault,
	})
}

func evasionOverviewTable(payload models.EvasionOverviewOutput) string {
	rows := make([][]string, 0, len(payload.Surfaces))
	for _, surface := range payload.Surfaces {
		rows = append(rows, []string{
			surface.Surface,
			surface.Summary,
		})
	}
	return renderListTable(
		"ho-azure evasion",
		[]string{"surface", "summary"},
		rows,
		[]string{"no evasion surfaces available", ""},
		evasionOverviewTakeaway(payload),
	)
}

func evasionDCRTable(payload models.EvasionDCROutput) string {
	if len(payload.DCRs) == 0 {
		return renderFamilySurfaceTable(familySurfaceTableConfig{
			Title:         "ho-azure evasion dcr",
			EmptyHeaders:  []string{"dcr", "status"},
			EmptyRow:      []string{"No visible DCRs were confirmed from current scope.", ""},
			EmptyTakeaway: "0 DCRs visible; no DCR evasion surface was confirmed from current scope.",
		})
	}
	lead := payload.DCRs[0]
	return renderFamilySurfaceTable(familySurfaceTableConfig{
		Title:             "ho-azure evasion dcr",
		CapabilityTitle:   "DCR evasion capability",
		CapabilitySteps:   lead.CapabilitySteps,
		MultiTargetNote:   "This walkthrough shows the strongest currently visible DCR truth-disruption path. The inventory below lists the other visible DCRs without repeating the same narrative.",
		TargetCount:       len(payload.DCRs),
		Explanation:       evasionDCRExplanation(lead),
		ReducedVisibility: familyReducedVisibilityExplanation("DCR and association", "DCR management-plane", "evasion", lead.CurrentIdentityContext),
		InventoryTitle:    "Visible DCRs",
		InventoryHeaders:  []string{"dcr", "rank", "streams", "destinations", "transforms", "current identity"},
		InventoryRows:     evasionDCRInventoryRows(payload.DCRs),
		BoundaryNotes:     lead.NotCollectedByDefault,
	})
}

func evasionDiagnosticSettingsTable(payload models.EvasionDiagnosticSettingsOutput) string {
	if len(payload.Sources) == 0 {
		return renderFamilySurfaceTable(familySurfaceTableConfig{
			Title:         "ho-azure evasion diagnostic-settings",
			EmptyHeaders:  []string{"source", "status"},
			EmptyRow:      []string{"No visible diagnostic settings sources were confirmed from current scope.", ""},
			EmptyTakeaway: "0 sources visible; no diagnostic-settings evasion surface was confirmed from current scope.",
		})
	}
	lead := payload.Sources[0]
	return renderFamilySurfaceTable(familySurfaceTableConfig{
		Title:             "ho-azure evasion diagnostic-settings",
		CapabilityTitle:   "Diagnostic settings evasion capability",
		CapabilitySteps:   lead.CapabilitySteps,
		MultiTargetNote:   "This walkthrough shows the strongest currently visible diagnostic-settings truth-disruption path. The inventory below lists the other visible sources without repeating the same narrative.",
		TargetCount:       len(payload.Sources),
		Explanation:       evasionDiagnosticSettingsExplanation(lead),
		ReducedVisibility: familyReducedVisibilityExplanation("diagnostic settings", "diagnostic-settings management-plane", "evasion", lead.CurrentIdentityContext),
		InventoryTitle:    "Visible Sources",
		InventoryHeaders:  []string{"source", "rank", "type", "destinations", "not exported", "current identity"},
		InventoryRows:     evasionDiagnosticSettingsInventoryRows(payload.Sources),
		BoundaryNotes:     lead.NotCollectedByDefault,
	})
}

func evasionDCRExplanation(dcr models.EvasionDCR) string {
	lines := []string{
		"",
		"Operator read",
		dcr.Summary,
		"Current identity: " + familyRoleSummary(dcr.CurrentIdentityContext),
		"Downstream effect: " + dcr.DisruptionReason,
		"First boundary: this is DCR management-plane posture, not log-content proof, runtime agent proof, or downstream detector failure.",
	}
	if dcr.CurrentState.TransformationPosture != "" {
		lines = append(lines, "Transformation posture: "+dcr.CurrentState.TransformationPosture+".")
	}
	if dcr.CurrentState.DestinationPosture != "" {
		lines = append(lines, "Destination posture: "+dcr.CurrentState.DestinationPosture+".")
	}
	if len(dcr.CurrentState.AssociationTargets) > 0 {
		lines = append(lines, "Association scope: "+strings.Join(shortResourceNames(dcr.CurrentState.AssociationTargets), ", ")+".")
	}
	return strings.Join(lines, "\n")
}

func evasionDiagnosticSettingsExplanation(source models.EvasionDiagnosticSettingsSource) string {
	lines := []string{
		"",
		"Operator read",
		source.Summary,
		"Current identity: " + familyRoleSummary(source.CurrentIdentityContext),
		"Downstream effect: " + source.DisruptionReason,
		"First boundary: this is diagnostic-settings management-plane posture, not sink-content proof, history proof, or detector-failure proof.",
	}
	if source.CurrentState.ExportPosture != "" {
		lines = append(lines, "Export posture: "+source.CurrentState.ExportPosture+".")
	}
	if source.CurrentState.DestinationPosture != "" {
		lines = append(lines, "Destination posture: "+source.CurrentState.DestinationPosture+".")
	}
	return strings.Join(lines, "\n")
}

func evasionAppInsightsExplanation(target models.EvasionAppInsightsTarget) string {
	lines := []string{
		"",
		"Operator read",
		target.Summary,
		"Current identity: " + familyRoleSummary(target.CurrentIdentityContext),
		"Downstream effect: " + target.DisruptionReason,
		"First boundary: this is visible Application Insights and app-setting posture, not code-body proof, runtime proof, or detector-failure proof.",
		"Posture: " + target.CurrentState.Posture + ".",
	}
	return strings.Join(lines, "\n")
}

func evasionDCRInventoryRows(dcrs []models.EvasionDCR) [][]string {
	rows := make([][]string, 0, len(dcrs))
	for _, dcr := range dcrs {
		rows = append(rows, []string{
			dcr.Name,
			fmt.Sprintf("%d/5", dcr.DisruptionRank),
			joinOrNone(dcr.CurrentState.Streams),
			joinOrNone(dcr.CurrentState.DestinationTypes),
			fmt.Sprintf("%d", dcr.CurrentState.TransformationCount),
			familyRoleControlLabel(dcr.CurrentIdentityContext),
		})
	}
	return rows
}

func evasionDiagnosticSettingsInventoryRows(sources []models.EvasionDiagnosticSettingsSource) [][]string {
	rows := make([][]string, 0, len(sources))
	for _, source := range sources {
		rows = append(rows, []string{
			source.Name,
			fmt.Sprintf("%d/5", source.DisruptionRank),
			source.CurrentState.SourceType,
			joinOrNone(source.CurrentState.DestinationTypes),
			joinOrNone(source.CurrentState.NotExportedCategories),
			familyRoleControlLabel(source.CurrentIdentityContext),
		})
	}
	return rows
}

func evasionAppInsightsInventoryRows(targets []models.EvasionAppInsightsTarget) [][]string {
	rows := make([][]string, 0, len(targets))
	for _, target := range targets {
		rows = append(rows, []string{
			target.Name,
			fmt.Sprintf("%d/5", target.DisruptionRank),
			target.CurrentState.Kind,
			joinOrNone(target.CurrentState.SamplingClues),
			joinOrNone(target.CurrentState.FilteringClues),
			familyRoleControlLabel(target.CurrentIdentityContext),
		})
	}
	return rows
}

func evasionOverviewTakeaway(payload models.EvasionOverviewOutput) string {
	return fmt.Sprintf("%d evasion surface(s) available; run a surface to rank visible posture by family-specific disruption value.", len(payload.Surfaces))
}

func shortResourceNames(ids []string) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, resourceNameFromIDForTable(id))
	}
	return values
}

func joinOrNone(values []string) string {
	if len(values) == 0 {
		return "none visible"
	}
	return strings.Join(values, ", ")
}
