package providers

import (
	"strings"
	"testing"
)

func TestContainerInstanceSummaryPreservesPublicIPInRelatedIDsAndAzureFoxWording(t *testing.T) {
	resource := map[string]any{
		"id":       "/subscriptions/sub/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-web",
		"name":     "aci-web",
		"location": "eastus",
		"identity": map[string]any{
			"type":        "SystemAssigned",
			"principalId": "principal-id",
		},
		"properties": map[string]any{
			"osType":            "Linux",
			"restartPolicy":     "Always",
			"provisioningState": "Succeeded",
			"containers": []any{
				map[string]any{"properties": map[string]any{"image": "ghcr.io/harrierops/web:1.0"}},
			},
			"ipAddress": map[string]any{
				"fqdn": "aci-web.eastus.azurecontainer.io",
				"ip":   "4.150.147.0",
				"ports": []any{
					map[string]any{"port": 80},
				},
			},
		},
	}

	summary := containerInstanceSummary(resource)
	if !containsStringValue(summary.RelatedIDs, "4.150.147.0") {
		t.Fatalf("containerInstanceSummary().RelatedIDs = %#v, want public IP included", summary.RelatedIDs)
	}
	if !strings.Contains(summary.Summary, "Container group 'aci-web'") {
		t.Fatalf("containerInstanceSummary().Summary = %q, want Container group wording", summary.Summary)
	}
	if !strings.Contains(summary.Summary, "Visible posture: os Linux, restart Always, ports 80, containers 1.") {
		t.Fatalf("containerInstanceSummary().Summary = %q, want AzureFox-style posture wording", summary.Summary)
	}
}
