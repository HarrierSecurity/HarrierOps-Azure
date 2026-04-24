package render

import (
	"fmt"

	"harrierops-azure/internal/models"
)

func appInsightsTable(payload models.AppInsightsOutput) string {
	rows := make([][]string, 0, len(payload.Targets))
	for _, target := range payload.Targets {
		rows = append(rows, []string{
			target.Name,
			target.Kind,
			joinOrNone(target.SamplingClues),
			joinOrNone(target.FilteringClues),
			joinOrNone(target.LoggingLevelClues),
		})
	}
	output := renderListTable(
		"ho-azure appinsights",
		[]string{"target", "kind", "sampling", "filtering", "logging"},
		rows,
		[]string{"no instrumented app setting clues", "", "", "", ""},
		appInsightsTakeaway(payload),
	)
	if len(payload.Components) > 0 {
		componentRows := make([][]string, 0, len(payload.Components))
		for _, component := range payload.Components {
			componentRows = append(componentRows, []string{
				component.Name,
				component.ResourceGroup,
				valueOrFallback(component.IngestionMode, "unknown"),
				resourceNameFromIDForTable(stringPtrValue(component.WorkspaceResourceID)),
			})
		}
		output += "\nComponents\n" + renderAlignedPipeTable([]string{"component", "resource group", "ingestion", "workspace"}, componentRows)
	}
	output += "\nNot collected by default\n"
	output += renderAlignedPipeTable(
		[]string{"item", "classification", "reason"},
		[][]string{
			{"setting values", "recon safety", "Default output uses setting names as posture clues and does not print instrumentation keys or connection strings."},
			{"code-level processors", "proof boundary", "Telemetry processor bodies usually live in source code or binaries, outside management-plane posture."},
			{"true unsampled count", "proof boundary", "Current posture cannot prove how many events were dropped or retained."},
			{"host.json body", "collector issue", "Function sampling can live in host.json; this helper only uses visible app setting names by default."},
			{"detector failure", "proof boundary", "The command does not inspect detections, so it cannot claim a rule missed activity."},
		},
	)
	return output
}

func appInsightsTakeaway(payload models.AppInsightsOutput) string {
	sampling := 0
	filtering := 0
	for _, target := range payload.Targets {
		if len(target.SamplingClues) > 0 {
			sampling++
		}
		if len(target.FilteringClues) > 0 {
			filtering++
		}
	}
	return fmt.Sprintf("%d component(s) and %d instrumented target(s) visible; %d target(s) show sampling clues and %d show filtering clues.", len(payload.Components), len(payload.Targets), sampling, filtering)
}
