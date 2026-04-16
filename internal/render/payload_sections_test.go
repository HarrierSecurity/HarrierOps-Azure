package render

import (
	"testing"

	"harrierops-azure/internal/models"
)

func TestPayloadFindingsConvertsCustomFindingTypes(t *testing.T) {
	findings := payloadFindings(models.EnvVarsOutput{
		Findings: []models.EnvVarFinding{{
			ID:          "finding-1",
			Title:       "Sensitive setting",
			Severity:    "high",
			Description: "Plain-text secret is visible.",
			RelatedIDs:  []string{"app-1"},
		}},
	})

	if len(findings) != 1 {
		t.Fatalf("payloadFindings() len = %d, want 1", len(findings))
	}
	if findings[0].Title != "Sensitive setting" || findings[0].Severity != "high" {
		t.Fatalf("payloadFindings() = %#v, want converted finding", findings[0])
	}
}

func TestPayloadIssuesHandlesKnownOutputTypesExplicitly(t *testing.T) {
	issues := payloadIssues(models.AppCredentialsOutput{
		Issues: []models.Issue{{
			Kind:    "graph",
			Message: "owners query failed",
		}},
	})

	if len(issues) != 1 {
		t.Fatalf("payloadIssues() len = %d, want 1", len(issues))
	}
	if issues[0].Kind != "graph" || issues[0].Message != "owners query failed" {
		t.Fatalf("payloadIssues() = %#v, want copied issue", issues[0])
	}
}
