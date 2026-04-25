package providers

import (
	"testing"

	"harrierops-azure/internal/models"
)

func TestMergeArtifactIdentityFactsMarksMismatchedSourceContext(t *testing.T) {
	first := ArtifactIdentityFacts{
		CurrentPrincipal: models.Principal{ID: "principal-a", PrincipalType: "ServicePrincipal", TenantID: "tenant-a"},
		AuthMode:         "azure_cli",
		TokenSource:      "cli",
	}
	second := ArtifactIdentityFacts{
		CurrentPrincipal: models.Principal{ID: "principal-b", PrincipalType: "ServicePrincipal", TenantID: "tenant-a"},
		AuthMode:         "azure_cli",
		TokenSource:      "cli",
	}

	merged, issues := MergeArtifactIdentityFacts(first, second)
	if merged.CurrentPrincipal.ID != "principal-a" {
		t.Fatalf("expected first visible identity to remain selected, got %#v", merged)
	}
	if len(issues) != 1 || issues[0].Kind != "artifact_identity_mismatch" {
		t.Fatalf("expected artifact identity mismatch issue, got %#v", issues)
	}
}
