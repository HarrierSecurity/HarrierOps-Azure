package providers

import "testing"

func TestWebJobTriggeredModeUsesSchedulerClues(t *testing.T) {
	mode := webJobTriggeredMode(
		stringPtr("https://example.scm.azurewebsites.net/api/triggeredwebjobs/nightly-reconcile/history"),
		nil,
		map[string]interface{}{},
	)

	if mode != "scheduled" {
		t.Fatalf("webJobTriggeredMode() = %q, want scheduled", mode)
	}
}

func TestWebJobTriggeredModeFallsBackToTriggeredManual(t *testing.T) {
	mode := webJobTriggeredMode(
		nil,
		stringPtr("Manual"),
		map[string]interface{}{},
	)

	if mode != "triggered/manual" {
		t.Fatalf("webJobTriggeredMode() = %q, want triggered/manual", mode)
	}
}
