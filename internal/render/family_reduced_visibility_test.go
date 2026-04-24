package render

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func reducedVisibilityFamilySteps(actions ...string) []models.FamilyCapabilityStep {
	steps := make([]models.FamilyCapabilityStep, 0, len(actions))
	for index, action := range actions {
		status := "not proven"
		if index == 0 {
			status = "visible posture only"
		}
		steps = append(steps, models.FamilyCapabilityStep{
			Action: action,
			Status: status,
			CanAct: status == "yes" || status == "partial",
		})
	}
	return steps
}

func reducedVisibilityRoleContext() *models.FamilyRoleContext {
	return &models.FamilyRoleContext{
		ControlLabel: "not proven",
		Summary:      "Current foothold identity is visible, but write control is not proven here.",
	}
}

func TestFamilyTablesStopOperatorWalkthroughWhenOnlyReadVisibilityIsProven(t *testing.T) {
	const marker = "FULL_WALKTHROUGH_MARKER"
	tests := []struct {
		name   string
		output string
	}{
		{
			name: "evasion dcr",
			output: evasionDCRTable(models.EvasionDCROutput{DCRs: []models.EvasionDCR{{
				Name:                   "dcr-prod",
				Summary:                marker,
				DisruptionReason:       marker,
				CapabilitySteps:        reducedVisibilityFamilySteps("choose or create DCR", "save or re-associate rule"),
				CurrentIdentityContext: reducedVisibilityRoleContext(),
			}}}),
		},
		{
			name: "evasion diagnostic-settings",
			output: evasionDiagnosticSettingsTable(models.EvasionDiagnosticSettingsOutput{Sources: []models.EvasionDiagnosticSettingsSource{{
				Name:                   "kv-prod",
				Summary:                marker,
				DisruptionReason:       marker,
				CapabilitySteps:        reducedVisibilityFamilySteps("pick source resource", "save or edit setting"),
				CurrentIdentityContext: reducedVisibilityRoleContext(),
			}}}),
		},
		{
			name: "evasion appinsights",
			output: evasionAppInsightsTable(models.EvasionAppInsightsOutput{Targets: []models.EvasionAppInsightsTarget{{
				Name:                   "app-prod",
				Summary:                marker,
				DisruptionReason:       marker,
				CapabilitySteps:        reducedVisibilityFamilySteps("choose telemetry target", "configure sampling"),
				CurrentIdentityContext: reducedVisibilityRoleContext(),
			}}}),
		},
		{
			name: "resourcehijacking api-mgmt",
			output: resourceHijackingAPIMTable(models.ResourceHijackingAPIMOutput{Targets: []models.ResourceHijackingAPIMTarget{{
				Name:                   "apim-prod",
				Summary:                marker,
				TakeoverReason:         marker,
				CapabilitySteps:        reducedVisibilityFamilySteps("select trusted gateway", "change backend or routing policy"),
				CurrentIdentityContext: reducedVisibilityRoleContext(),
			}}}),
		},
		{
			name: "resourcehijacking automation",
			output: resourceHijackingAutomationTable(models.ResourceHijackingAutomationOutput{Targets: []models.ResourceHijackingAutomationTarget{{
				Name:                   "aa-prod",
				Summary:                marker,
				TakeoverReason:         marker,
				CapabilitySteps:        reducedVisibilityFamilySteps("select trusted automation account", "edit published runbook"),
				CurrentIdentityContext: reducedVisibilityRoleContext(),
			}}}),
		},
		{
			name: "resourcehijacking logic-apps",
			output: resourceHijackingLogicAppsTable(models.ResourceHijackingLogicAppsOutput{Targets: []models.ResourceHijackingLogicAppTarget{{
				Name:                   "wf-prod",
				Summary:                marker,
				TakeoverReason:         marker,
				CapabilitySteps:        reducedVisibilityFamilySteps("select trusted workflow", "edit workflow definition"),
				CurrentIdentityContext: reducedVisibilityRoleContext(),
			}}}),
		},
		{
			name: "pathmasking api-mgmt",
			output: pathMaskingAPIMTable(models.PathMaskingAPIMOutput{Targets: []models.PathMaskingAPIMTarget{{
				Name:                   "apim-prod",
				Summary:                marker,
				MaskingReason:          marker,
				CapabilitySteps:        reducedVisibilityFamilySteps("select public gateway", "apply route or transform policy"),
				CurrentIdentityContext: reducedVisibilityRoleContext(),
			}}}),
		},
		{
			name: "pathmasking logic-apps",
			output: pathMaskingLogicAppsTable(models.PathMaskingLogicAppsOutput{Targets: []models.PathMaskingLogicAppTarget{{
				Name:                   "wf-prod",
				Summary:                marker,
				MaskingReason:          marker,
				CapabilitySteps:        reducedVisibilityFamilySteps("select trusted workflow", "change workflow route"),
				CurrentIdentityContext: reducedVisibilityRoleContext(),
			}}}),
		},
		{
			name: "pathmasking relay",
			output: pathMaskingRelayTable(models.PathMaskingRelayOutput{Targets: []models.PathMaskingRelayTarget{{
				Name:                   "relay-prod",
				Summary:                marker,
				MaskingReason:          marker,
				CapabilitySteps:        reducedVisibilityFamilySteps("select Relay namespace", "change namespace or connection posture"),
				CurrentIdentityContext: reducedVisibilityRoleContext(),
			}}}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !strings.Contains(test.output, "Higher permissions are required") {
				t.Fatalf("expected reduced-visibility stop line, got:\n%s", test.output)
			}
			if strings.Contains(test.output, marker) {
				t.Fatalf("expected read-only output to suppress full operator walkthrough details, got:\n%s", test.output)
			}
			if strings.Contains(test.output, "Downstream effect:") {
				t.Fatalf("expected read-only output to stop before downstream-effect narration, got:\n%s", test.output)
			}
		})
	}
}
