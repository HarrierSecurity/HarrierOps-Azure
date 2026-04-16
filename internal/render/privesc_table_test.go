package render

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestPrivescTableShowsPreferredFootholdWithoutOverclaiming(t *testing.T) {
	output, err := Table("privesc", models.PrivescOutput{
		Paths: []models.PrivescPathSummary{
			{
				CurrentIdentity: true,
				PathType:        "current-foothold-direct-control",
				Priority:        "high",
				Principal:       "svc-current",
				PrincipalType:   "ServicePrincipal",
				Preferred:       true,
				PreferredReason: "Preferred foothold: current foothold svc-current (ServicePrincipal). It already has direct high-impact RBAC on visible scope.",
				ProvenPath:      "Current foothold 'svc-current' already holds high-impact RBAC (Owner) on visible scope.",
				MissingProof:    "HO-Azure does not prove which exact abuse action is the best next step from this row alone.",
			},
			{
				Asset:          models.StringPtr("vm-public"),
				PathType:       "ingress-backed-workload-identity",
				Priority:       "medium",
				Principal:      "mi-nearby",
				PrincipalType:  "ManagedIdentity",
				ProvenPath:     "Public workload 'vm-public' carries identity 'mi-nearby' with high-impact RBAC (Owner).",
				MissingProof:   "HO-Azure does not prove control of the workload or successful token use from it.",
				OperatorSignal: "Visible ingress-backed lead; not yet rooted in current foothold.",
			},
		},
	}, models.RenderContext{})
	if err != nil {
		t.Fatalf("Table(privesc) returned error: %v", err)
	}

	if !strings.Contains(output, "Preferred foothold: current foothold svc-current (ServicePrincipal).") {
		t.Fatalf("expected preferred-foothold explanation in output, got:\n%s", output)
	}
	if strings.Contains(output, "│ operator signal ") {
		t.Fatalf("did not expect operator signal column in output, got:\n%s", output)
	}
	if strings.Contains(output, "│ next review ") {
		t.Fatalf("did not expect next review column in output, got:\n%s", output)
	}
	if !strings.Contains(output, "│ note ") {
		t.Fatalf("expected wrapped note row in output, got:\n%s", output)
	}
	if !strings.Contains(output, "managed identity mi-nearby via vm-public") {
		t.Fatalf("expected managed identity target label in output, got:\n%s", output)
	}
	if !strings.Contains(output, "2 privilege-escalation paths surfaced; 1 current-identity-rooted, 1 visible-only lead, 1 high, 1 medium.") {
		t.Fatalf("expected compact count takeaway in output, got:\n%s", output)
	}
}
