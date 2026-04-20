package providers

import (
	"strings"
	"testing"
)

func TestWebAppDefaultHostnameFallsBackToProperties(t *testing.T) {
	app := map[string]any{
		"name": "app-public-api",
		"properties": map[string]any{
			"defaultHostName": "app-public-api.azurewebsites.net",
		},
	}

	got := webAppDefaultHostname(app)
	if got == nil {
		t.Fatal("webAppDefaultHostname() = nil, want hostname")
	}
	if *got != "app-public-api.azurewebsites.net" {
		t.Fatalf("webAppDefaultHostname() = %q, want %q", *got, "app-public-api.azurewebsites.net")
	}
}

func TestAppServiceSummaryPreservesListResponseWebAppFields(t *testing.T) {
	app := map[string]any{
		"id":                  "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
		"name":                "app-public-api",
		"location":            "eastus",
		"publicNetworkAccess": "Enabled",
		"properties": map[string]any{
			"clientCertEnabled": true,
			"defaultHostName":   "app-public-api.azurewebsites.net",
			"httpsOnly":         true,
			"serverFarmId":      "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.Web/serverfarms/asp-public-api",
			"state":             "Running",
		},
	}

	summary := appServiceSummary(app, map[string]any{})
	if summary.DefaultHostname == nil {
		t.Fatal("appServiceSummary().DefaultHostname = nil, want hostname")
	}
	if *summary.DefaultHostname != "app-public-api.azurewebsites.net" {
		t.Fatalf("appServiceSummary().DefaultHostname = %q, want %q", *summary.DefaultHostname, "app-public-api.azurewebsites.net")
	}
	if summary.PublicNetworkAccess == nil || *summary.PublicNetworkAccess != "Enabled" {
		t.Fatalf("appServiceSummary().PublicNetworkAccess = %v, want Enabled", summary.PublicNetworkAccess)
	}
	if !summary.HTTPSOnly {
		t.Fatal("appServiceSummary().HTTPSOnly = false, want true")
	}
	if !summary.ClientCertEnabled {
		t.Fatal("appServiceSummary().ClientCertEnabled = false, want true")
	}
	if summary.AppServicePlanID == nil || *summary.AppServicePlanID != "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.Web/serverfarms/asp-public-api" {
		t.Fatalf("appServiceSummary().AppServicePlanID = %v, want app service plan id", summary.AppServicePlanID)
	}
	if summary.State == nil || *summary.State != "Running" {
		t.Fatalf("appServiceSummary().State = %v, want Running", summary.State)
	}
	if !containsStringValue(summary.RelatedIDs, "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.Web/serverfarms/asp-public-api") {
		t.Fatalf("appServiceSummary().RelatedIDs = %v, want plan id present", summary.RelatedIDs)
	}
	if !strings.Contains(summary.Summary, "publishes hostname 'app-public-api.azurewebsites.net'") {
		t.Fatalf("appServiceSummary().Summary = %q, want hostname phrase", summary.Summary)
	}
	if !strings.Contains(summary.Summary, "public network access Enabled") {
		t.Fatalf("appServiceSummary().Summary = %q, want public network access phrase", summary.Summary)
	}
	if !strings.Contains(summary.Summary, "HTTPS-only enabled") {
		t.Fatalf("appServiceSummary().Summary = %q, want HTTPS-only phrase", summary.Summary)
	}
}

func TestFunctionAppSummaryPreservesListResponseWebAppFields(t *testing.T) {
	app := map[string]any{
		"id":                  "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
		"name":                "func-orders",
		"location":            "eastus",
		"publicNetworkAccess": "Enabled",
		"properties": map[string]any{
			"clientCertEnabled": true,
			"defaultHostName":   "func-orders.azurewebsites.net",
			"httpsOnly":         true,
			"serverFarmId":      "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.Web/serverfarms/asp-functions",
			"state":             "Running",
		},
	}

	summary := functionAppSummary(app, map[string]any{}, map[string]any{})
	if summary.DefaultHostname == nil || *summary.DefaultHostname != "func-orders.azurewebsites.net" {
		t.Fatalf("functionAppSummary().DefaultHostname = %v, want func-orders.azurewebsites.net", summary.DefaultHostname)
	}
	if summary.PublicNetworkAccess == nil || *summary.PublicNetworkAccess != "Enabled" {
		t.Fatalf("functionAppSummary().PublicNetworkAccess = %v, want Enabled", summary.PublicNetworkAccess)
	}
	if !summary.HTTPSOnly {
		t.Fatal("functionAppSummary().HTTPSOnly = false, want true")
	}
	if !summary.ClientCertEnabled {
		t.Fatal("functionAppSummary().ClientCertEnabled = false, want true")
	}
	if summary.AppServicePlanID == nil || *summary.AppServicePlanID != "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.Web/serverfarms/asp-functions" {
		t.Fatalf("functionAppSummary().AppServicePlanID = %v, want app service plan id", summary.AppServicePlanID)
	}
	if summary.State == nil || *summary.State != "Running" {
		t.Fatalf("functionAppSummary().State = %v, want Running", summary.State)
	}
	if !containsStringValue(summary.RelatedIDs, "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.Web/serverfarms/asp-functions") {
		t.Fatalf("functionAppSummary().RelatedIDs = %v, want plan id present", summary.RelatedIDs)
	}
	if !strings.Contains(summary.Summary, "publishes hostname 'func-orders.azurewebsites.net'") {
		t.Fatalf("functionAppSummary().Summary = %q, want hostname phrase", summary.Summary)
	}
	if !strings.Contains(summary.Summary, "public network access Enabled") {
		t.Fatalf("functionAppSummary().Summary = %q, want public network access phrase", summary.Summary)
	}
	if !strings.Contains(summary.Summary, "HTTPS-only enabled") {
		t.Fatalf("functionAppSummary().Summary = %q, want HTTPS-only phrase", summary.Summary)
	}
}

func containsStringValue(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
