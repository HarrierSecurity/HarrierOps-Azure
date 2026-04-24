package render

import (
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

func diagnosticSettingsTable(payload models.DiagnosticSettingsOutput) string {
	rows := make([][]string, 0, len(payload.Sources))
	for _, source := range payload.Sources {
		rows = append(rows, []string{
			source.Name,
			source.Type,
			diagnosticSettingsExportContext(source),
			joinOrNone(source.DestinationTypes),
			diagnosticSettingsCategoryContext(source),
		})
	}
	output := renderListTable(
		"ho-azure diagnostic-settings",
		[]string{"source", "type", "settings", "destinations", "categories"},
		rows,
		[]string{"no visible resources", "", "", "", ""},
		diagnosticSettingsTakeaway(payload),
	)
	output += "\nNot collected by default\n"
	output += renderAlignedPipeTable(
		[]string{"item", "classification", "reason"},
		[][]string{
			{"unsupported category proof", "proof boundary", "When Azure rejects category catalog reads for a source type, the helper reports a collection issue instead of treating omitted categories as unsupported."},
			{"activity-log history", "API/noise", "Change timing and actor proof require history collection, which is not needed for the default posture view."},
			{"sink contents", "proof boundary", "Log Analytics, Storage, Event Hub, or partner sink contents are data/content evidence, not management-plane posture."},
			{"detector wiring", "proof boundary", "The command does not inspect Sentinel or defender rule dependencies, so it cannot claim a detection failed."},
			{"expected SOC destination baseline", "scope/sequencing", "Destination drift needs an expected sink model before the tool can call the current sink wrong."},
		},
	)
	return output
}

func diagnosticSettingsExportContext(source models.DiagnosticSettingsSource) string {
	if !source.HasDiagnosticSettings {
		return "none visible"
	}
	return fmt.Sprintf("%d visible", source.DiagnosticSettingCount)
}

func diagnosticSettingsCategoryContext(source models.DiagnosticSettingsSource) string {
	parts := []string{}
	if len(source.EnabledCategories) > 0 {
		parts = append(parts, "enabled: "+strings.Join(source.EnabledCategories, ", "))
	}
	if len(source.DisabledCategories) > 0 {
		parts = append(parts, "not exported by visible setting: "+strings.Join(source.DisabledCategories, ", "))
	}
	if len(source.NotExportedSupported) > 0 {
		parts = append(parts, "supported not exported: "+strings.Join(source.NotExportedSupported, ", "))
	}
	if len(source.HighSignalCategories) > 0 {
		parts = append(parts, "high-signal: "+strings.Join(source.HighSignalCategories, ", "))
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return strings.Join(parts, "; ")
}

func diagnosticSettingsTakeaway(payload models.DiagnosticSettingsOutput) string {
	if len(payload.Sources) == 0 {
		return "no source resources were visible from the current read path."
	}
	withSettings := 0
	partial := 0
	nonWorkspace := 0
	for _, source := range payload.Sources {
		if source.HasDiagnosticSettings {
			withSettings++
		}
		if source.HasPartialLogPosture {
			partial++
		}
		if source.HasNonWorkspaceDestination {
			nonWorkspace++
		}
	}
	return fmt.Sprintf("%d source(s) visible; %d have diagnostic settings, %d show partial category posture, and %d route to non-Log Analytics destinations.", len(payload.Sources), withSettings, partial, nonWorkspace)
}
