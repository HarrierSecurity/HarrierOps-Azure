package providers

import (
	"strings"
	"testing"
)

func TestAcrRegistryNeedsHydrationWhenAnonymousPullEnabledMissing(t *testing.T) {
	registry := map[string]any{
		"identity": map[string]any{
			"type": "systemAssigned",
		},
		"properties": map[string]any{
			"publicNetworkAccess": "Enabled",
		},
	}

	if !acrRegistryNeedsHydration(registry) {
		t.Fatal("acrRegistryNeedsHydration() = false, want true when anonymous pull field is still missing")
	}
}

func TestAcrRegistrySummaryPreservesAnonymousPullEnabledFromProperties(t *testing.T) {
	summary := acrRegistrySummary(
		map[string]any{
			"id":       "/subscriptions/sub/resourceGroups/rg-ops/providers/Microsoft.ContainerRegistry/registries/acrtest",
			"name":     "acrtest",
			"location": "eastus",
			"sku": map[string]any{
				"name": "Basic",
			},
			"properties": map[string]any{
				"adminUserEnabled":         false,
				"anonymousPullEnabled":     false,
				"loginServer":              "acrtest.azurecr.io",
				"provisioningState":        "Succeeded",
				"publicNetworkAccess":      "Enabled",
				"networkRuleBypassOptions": "AzureServices",
			},
		},
		nil,
		nil,
	)

	if summary.AnonymousPullEnabled == nil || *summary.AnonymousPullEnabled {
		t.Fatalf("acrRegistrySummary().AnonymousPullEnabled = %v, want false", summary.AnonymousPullEnabled)
	}
	if summary.PublicNetworkAccess == nil || *summary.PublicNetworkAccess != "Enabled" {
		t.Fatalf("acrRegistrySummary().PublicNetworkAccess = %v, want Enabled", summary.PublicNetworkAccess)
	}
	if summary.AdminUserEnabled == nil || *summary.AdminUserEnabled {
		t.Fatalf("acrRegistrySummary().AdminUserEnabled = %v, want false", summary.AdminUserEnabled)
	}
	if !strings.Contains(summary.Summary, "anonymous pull disabled") {
		t.Fatalf("acrRegistrySummary().Summary = %q, want anonymous pull wording", summary.Summary)
	}
}
