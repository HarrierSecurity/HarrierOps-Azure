package render

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestPrivescTableThemeTieBreakExample(t *testing.T) {
	output, err := Table("privesc", models.PrivescOutput{
		Paths: []models.PrivescPathSummary{
			{
				CurrentIdentity:  false,
				StartingFoothold: "dev-user (current foothold)",
				PathType:         "visible-privileged-lead",
				Priority:         "medium",
				Principal:        "aaa-pipeline-sp",
				PrincipalType:    "ServicePrincipal",
				Preferred:        true,
				PreferredReason:  "Preferred foothold: service principal aaa-pipeline-sp. It edges out otherwise similar alternatives because its naming/context looks pipeline-themed.",
				OperatorSignal:   "Visible privileged lead; not yet rooted in current foothold.",
				ProvenPath:       "Visible principal 'aaa-pipeline-sp' already holds high-impact RBAC (Owner) on visible scope.",
				MissingProof:     "HO-Azure does not prove the current identity can act as or control this principal.",
				NextReview:       "Check role-trusts for paths that could let the current identity influence this privileged principal.",
			},
			{
				CurrentIdentity:  false,
				StartingFoothold: "dev-user (current foothold)",
				PathType:         "visible-privileged-lead",
				Priority:         "medium",
				Principal:        "zzz-generic-sp",
				PrincipalType:    "ServicePrincipal",
				OperatorSignal:   "Visible privileged lead; not yet rooted in current foothold.",
				ProvenPath:       "Visible principal 'zzz-generic-sp' already holds high-impact RBAC (Owner) on visible scope.",
				MissingProof:     "HO-Azure does not prove the current identity can act as or control this principal.",
				NextReview:       "Check role-trusts for paths that could let the current identity influence this privileged principal.",
			},
		},
	}, models.RenderContext{})
	if err != nil {
		t.Fatalf("Table(privesc) returned error: %v", err)
	}

	if !strings.Contains(output, "pipeline-themed") {
		t.Fatalf("expected theme tie-break wording in output, got:\n%s", output)
	}
	t.Log("\n" + output)
}
