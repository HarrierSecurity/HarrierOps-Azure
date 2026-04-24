package commands

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestPathMaskingAPIMRankDoesNotCallUnknownPublicPosturePublic(t *testing.T) {
	disabled := "Disabled"
	apiCount := 1
	backendCount := 1
	_, reason := pathMaskingAPIMRank(models.ApiMgmtServiceAsset{
		GatewayHostnames:    []string{"internal.contoso.test"},
		BackendCount:        &backendCount,
		APICount:            &apiCount,
		PublicNetworkAccess: &disabled,
	}, false)

	if strings.Contains(reason, "public gateway") {
		t.Fatalf("expected non-public APIM posture not to be labeled public, got %q", reason)
	}
	if !strings.Contains(reason, "gateway, backend indirection") {
		t.Fatalf("expected neutral gateway wording, got %q", reason)
	}
}
