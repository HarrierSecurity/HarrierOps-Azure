package providers

import "testing"

func TestAksClusterNeedsHydrationWhenPublicFQDNFieldIsMissing(t *testing.T) {
	cluster := map[string]any{
		"properties": map[string]any{
			"apiServerAccessProfile": map[string]any{
				"enablePrivateCluster": false,
			},
			"securityProfile": map[string]any{
				"workloadIdentity": map[string]any{
					"enabled": false,
				},
			},
			"ingressProfile": map[string]any{
				"webAppRouting": map[string]any{
					"enabled": false,
				},
			},
		},
	}

	if !aksClusterNeedsHydration(cluster) {
		t.Fatal("aksClusterNeedsHydration() = false, want true when public FQDN field is still missing")
	}
}

func TestAksClusterSummaryPreservesFalsePublicFQDNFromDirectShape(t *testing.T) {
	cluster := map[string]any{
		"id":   "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/aks-ops",
		"name": "aks-ops",
		"properties": map[string]any{
			"apiServerAccessProfile": map[string]any{
				"enablePrivateCluster":           false,
				"enablePrivateClusterPublicFQDN": false,
			},
		},
	}

	summary := aksClusterSummary(cluster)
	if summary.PrivateClusterEnabled == nil || *summary.PrivateClusterEnabled {
		t.Fatalf("aksClusterSummary().PrivateClusterEnabled = %v, want false", summary.PrivateClusterEnabled)
	}
	if summary.PublicFQDNEnabled == nil || *summary.PublicFQDNEnabled {
		t.Fatalf("aksClusterSummary().PublicFQDNEnabled = %v, want false", summary.PublicFQDNEnabled)
	}
}
