package providers

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestManagedIdentityNarrativeUsesLogicAppSpecificFollowup(t *testing.T) {
	operatorSignal, nextReview, summary := managedIdentityNarrative(
		"LogicApp",
		"la-inbound-redeploy",
		"la-inbound-redeploy-identity",
		models.WorkloadExposureNone,
		true,
		false,
		[]string{"Contributor"},
	)

	if !strings.Contains(operatorSignal, "Logic App") {
		t.Fatalf("expected Logic App operator signal, got %q", operatorSignal)
	}
	if strings.Contains(nextReview, "env-vars") {
		t.Fatalf("expected Logic App-specific next review, got %q", nextReview)
	}
	if !strings.Contains(nextReview, "logic-apps") {
		t.Fatalf("expected logic-apps next review guidance, got %q", nextReview)
	}
	if !strings.Contains(summary, "Logic App 'la-inbound-redeploy'") {
		t.Fatalf("expected Logic App summary, got %q", summary)
	}
}
