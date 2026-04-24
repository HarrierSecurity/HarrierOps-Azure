package render

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestPathMaskingRelayTableUsesSurfaceNarration(t *testing.T) {
	output, err := Table("pathmasking", models.PathMaskingRelayOutput{}, models.RenderContext{})
	if err != nil {
		t.Fatalf("Table(pathmasking relay) returned error: %v", err)
	}
	if !strings.Contains(output, "Relay path masking means Azure exposes the cloud rendezvous point") {
		t.Fatalf("expected Relay-specific pathmasking narration, got:\n%s", output)
	}
	if strings.Contains(output, "Walking the current identity through Azure-native relay, proxy, and workflow surfaces") {
		t.Fatalf("expected surface narration instead of generic pathmasking narration, got:\n%s", output)
	}
}
