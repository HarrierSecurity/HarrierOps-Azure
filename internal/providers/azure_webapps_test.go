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

	summary := appServiceSummary(app, map[string]any{}, map[string]any{}, map[string]any{}, map[string]any{})
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

func TestAppServiceSummaryIncludesDeploymentAndConfigSignals(t *testing.T) {
	app := map[string]any{
		"id":       "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
		"name":     "app-public-api",
		"location": "eastus",
		"identity": map[string]any{
			"type": "SystemAssigned",
		},
		"properties": map[string]any{
			"defaultHostName": "app-public-api.azurewebsites.net",
		},
	}
	settings := map[string]any{
		"properties": map[string]any{
			"API_TOKEN":                "@Microsoft.KeyVault(SecretUri=https://kv-prod.vault.azure.net/secrets/api-token)",
			"WEBSITE_RUN_FROM_PACKAGE": "1",
			"LOG_LEVEL":                "info",
		},
	}
	connectionStrings := map[string]any{
		"properties": map[string]any{
			"PrimaryDb": map[string]any{
				"type":  "SQLAzure",
				"value": "@Microsoft.KeyVault(SecretUri=https://kv-prod.vault.azure.net/secrets/primary-db)",
			},
			"RedisCache": map[string]any{
				"type":  "Custom",
				"value": "Endpoint=cache.redis.cache.windows.net:6380",
			},
		},
	}
	sourceControl := map[string]any{
		"properties": map[string]any{
			"branch":              "main",
			"isGitHubAction":      true,
			"isManualIntegration": false,
			"repoUrl":             "https://github.com/contoso/customer-portal",
		},
	}

	summary := appServiceSummary(app, map[string]any{}, settings, connectionStrings, sourceControl)
	if summary.Deployment == nil || !strings.Contains(*summary.Deployment, "repo github.com/contoso/customer-portal") {
		t.Fatalf("appServiceSummary().Deployment = %v, want repo summary", summary.Deployment)
	}
	if summary.DeploymentBranch == nil || *summary.DeploymentBranch != "main" {
		t.Fatalf("appServiceSummary().DeploymentBranch = %v, want main", summary.DeploymentBranch)
	}
	if summary.DeploymentIsGitHubAction == nil || !*summary.DeploymentIsGitHubAction {
		t.Fatalf("appServiceSummary().DeploymentIsGitHubAction = %v, want true", summary.DeploymentIsGitHubAction)
	}
	if summary.DeploymentManualIntegration == nil || *summary.DeploymentManualIntegration {
		t.Fatalf("appServiceSummary().DeploymentManualIntegration = %v, want false", summary.DeploymentManualIntegration)
	}
	if summary.RunFromPackage == nil || !*summary.RunFromPackage {
		t.Fatalf("appServiceSummary().RunFromPackage = %v, want true", summary.RunFromPackage)
	}
	if summary.AppSettingsCount == nil || *summary.AppSettingsCount != 3 {
		t.Fatalf("appServiceSummary().AppSettingsCount = %v, want 3", summary.AppSettingsCount)
	}
	if summary.KeyVaultReferenceCount == nil || *summary.KeyVaultReferenceCount != 1 {
		t.Fatalf("appServiceSummary().KeyVaultReferenceCount = %v, want 1", summary.KeyVaultReferenceCount)
	}
	if summary.SensitiveSettingCount == nil || *summary.SensitiveSettingCount != 1 {
		t.Fatalf("appServiceSummary().SensitiveSettingCount = %v, want 1", summary.SensitiveSettingCount)
	}
	if summary.ConnectionStringCount == nil || *summary.ConnectionStringCount != 2 {
		t.Fatalf("appServiceSummary().ConnectionStringCount = %v, want 2", summary.ConnectionStringCount)
	}
	if summary.KeyVaultConnectionStringCount == nil || *summary.KeyVaultConnectionStringCount != 1 {
		t.Fatalf("appServiceSummary().KeyVaultConnectionStringCount = %v, want 1", summary.KeyVaultConnectionStringCount)
	}
	if len(summary.ConnectionStringTypes) != 2 || summary.ConnectionStringTypes[0] != "Custom" || summary.ConnectionStringTypes[1] != "SQLAzure" {
		t.Fatalf("appServiceSummary().ConnectionStringTypes = %v, want Custom/SQLAzure", summary.ConnectionStringTypes)
	}
	if !strings.Contains(summary.Summary, "Deployment signals: repo github.com/contoso/customer-portal, branch main, GitHub Actions, continuous integration, run-from-package enabled.") {
		t.Fatalf("appServiceSummary().Summary = %q, want deployment phrase", summary.Summary)
	}
	if !strings.Contains(summary.Summary, "Visible config: 3 app setting(s), 1 Key Vault-backed setting(s), 1 sensitive-looking setting name(s), 2 connection string(s), 1 Key Vault-backed connection string(s), connection types Custom, SQLAzure.") {
		t.Fatalf("appServiceSummary().Summary = %q, want config phrase", summary.Summary)
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

	summary := functionAppSummary(app, map[string]any{}, map[string]any{}, nil)
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

func TestFunctionAppSummaryPreservesUserAssignedIdentityPrincipalDetails(t *testing.T) {
	app := map[string]any{
		"id":       "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
		"name":     "func-orders",
		"location": "eastus",
		"identity": map[string]any{
			"type": "SystemAssigned, UserAssigned",
			"userAssignedIdentities": map[string]any{
				"/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders": map[string]any{
					"principalId": "cece2222-2222-2222-2222-222222222222",
					"clientId":    "dfdf2222-2222-2222-2222-222222222222",
				},
			},
		},
	}

	summary := functionAppSummary(app, map[string]any{}, map[string]any{}, nil)
	if len(summary.UserAssignedIdentities) != 1 {
		t.Fatalf("functionAppSummary().UserAssignedIdentities = %#v, want 1 identity", summary.UserAssignedIdentities)
	}
	identity := summary.UserAssignedIdentities[0]
	if identity.Name != "ua-orders" {
		t.Fatalf("functionAppSummary().UserAssignedIdentities[0].Name = %q, want ua-orders", identity.Name)
	}
	if identity.PrincipalID == nil || *identity.PrincipalID != "cece2222-2222-2222-2222-222222222222" {
		t.Fatalf("functionAppSummary().UserAssignedIdentities[0].PrincipalID = %v, want principal id", identity.PrincipalID)
	}
	if identity.ClientID == nil || *identity.ClientID != "dfdf2222-2222-2222-2222-222222222222" {
		t.Fatalf("functionAppSummary().UserAssignedIdentities[0].ClientID = %v, want client id", identity.ClientID)
	}
	if !containsStringValue(summary.RelatedIDs, "cece2222-2222-2222-2222-222222222222") {
		t.Fatalf("functionAppSummary().RelatedIDs = %v, want user-assigned principal id present", summary.RelatedIDs)
	}
}

func TestFunctionChildAssetFromMapParsesTriggerAndBindings(t *testing.T) {
	function := map[string]any{
		"id":   "/subscriptions/sub/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders/functions/OrdersWebhook",
		"name": "OrdersWebhook",
		"properties": map[string]any{
			"invoke_url_template": "https://func-orders.azurewebsites.net/api/orders/webhook",
			"isDisabled":          false,
			"language":            "Python",
			"config": map[string]any{
				"bindings": []any{
					map[string]any{
						"authLevel": "function",
						"direction": "in",
						"name":      "req",
						"route":     "orders/webhook",
						"type":      "httpTrigger",
					},
					map[string]any{
						"direction": "out",
						"name":      "$return",
						"type":      "http",
					},
				},
			},
		},
	}

	child := functionChildAssetFromMap(function)
	if child.TriggerType == nil || *child.TriggerType != "HTTP" {
		t.Fatalf("functionChildAssetFromMap().TriggerType = %v, want HTTP", child.TriggerType)
	}
	if child.InvokeURLTemplate == nil || *child.InvokeURLTemplate != "https://func-orders.azurewebsites.net/api/orders/webhook" {
		t.Fatalf("functionChildAssetFromMap().InvokeURLTemplate = %v, want invoke URL", child.InvokeURLTemplate)
	}
	if len(child.Bindings) != 2 {
		t.Fatalf("functionChildAssetFromMap().Bindings = %v, want 2 bindings", child.Bindings)
	}
	if len(child.BindingTypes) != 2 || child.BindingTypes[0] != "httpTrigger" || child.BindingTypes[1] != "http" {
		t.Fatalf("functionChildAssetFromMap().BindingTypes = %v, want httpTrigger/http", child.BindingTypes)
	}
}

func TestEnvVarSummaryPreservesKeyVaultReferenceIdentityFromProperties(t *testing.T) {
	app := map[string]any{
		"id":       "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
		"name":     "func-orders",
		"location": "eastus",
		"identity": map[string]any{
			"type": "SystemAssigned, UserAssigned",
			"userAssignedIdentities": map[string]any{
				"/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders": map[string]any{},
			},
		},
		"properties": map[string]any{
			"keyVaultReferenceIdentity": "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders",
		},
	}

	summary := envVarSummary(app, "FunctionApp", "PAYMENT_API_KEY", "@Microsoft.KeyVault(SecretUri=https://kvlabopen01.vault.azure.net/secrets/payment-api-key)")
	if summary.KeyVaultReferenceIdentity == nil {
		t.Fatal("envVarSummary().KeyVaultReferenceIdentity = nil, want identity")
	}
	if *summary.KeyVaultReferenceIdentity != "/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders" {
		t.Fatalf("envVarSummary().KeyVaultReferenceIdentity = %q, want user-assigned identity id", *summary.KeyVaultReferenceIdentity)
	}
	if !strings.Contains(summary.Summary, "via /subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders") {
		t.Fatalf("envVarSummary().Summary = %q, want identity phrase", summary.Summary)
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
