package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) DiagnosticSettings(_ context.Context, tenant string, subscription string) (DiagnosticSettingsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID
	workspaceID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.OperationalInsights/workspaces/law-soc-prod"
	eventHubRuleID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.EventHub/namespaces/eh-monitor/authorizationRules/send"
	keyVaultID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-sec/providers/Microsoft.KeyVault/vaults/kv-prod"
	storageID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.Storage/storageAccounts/stdataprod"
	appID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-app/providers/Microsoft.Web/sites/app-prod"

	keyVaultSetting := models.DiagnosticSettingAsset{
		ID:               keyVaultID + "/providers/Microsoft.Insights/diagnosticSettings/send-audit",
		Name:             "send-audit",
		SourceResourceID: keyVaultID,
		Destinations: []models.DiagnosticSettingsDestination{
			{Type: "logAnalytics", ResourceID: &workspaceID, Detail: models.StringPtr("Dedicated")},
		},
		Logs: []models.DiagnosticSettingsCategory{
			{Name: "AuditEvent", Type: "log", Enabled: false},
		},
		Metrics: []models.DiagnosticSettingsCategory{
			{Name: "AllMetrics", Type: "metric", Enabled: true},
		},
		EnabledCategories:    []string{"AllMetrics"},
		DisabledCategories:   []string{"AuditEvent"},
		CategoryGroups:       []string{"AuditEvent"},
		HighSignalCategories: []string{"AuditEvent"},
		DestinationTypes:     []string{"logAnalytics"},
		RelatedIDs:           []string{keyVaultID, keyVaultID + "/providers/Microsoft.Insights/diagnosticSettings/send-audit", workspaceID},
	}
	keyVaultSetting.Summary = diagnosticSettingSummary(keyVaultSetting)
	storageSetting := models.DiagnosticSettingAsset{
		ID:               storageID + "/providers/Microsoft.Insights/diagnosticSettings/archive-storage",
		Name:             "archive-storage",
		SourceResourceID: storageID,
		Destinations: []models.DiagnosticSettingsDestination{
			{Type: "eventHubs", ResourceID: &eventHubRuleID, Detail: models.StringPtr("monitoring-migration")},
		},
		Logs: []models.DiagnosticSettingsCategory{
			{Name: "StorageRead", Type: "log", Enabled: true},
			{Name: "StorageWrite", Type: "log", Enabled: true},
		},
		Metrics:              []models.DiagnosticSettingsCategory{},
		EnabledCategories:    []string{"StorageRead", "StorageWrite"},
		HighSignalCategories: []string{},
		DestinationTypes:     []string{"eventHubs"},
		RelatedIDs:           []string{eventHubRuleID, storageID, storageID + "/providers/Microsoft.Insights/diagnosticSettings/archive-storage"},
	}
	storageSetting.Summary = diagnosticSettingSummary(storageSetting)

	facts := DiagnosticSettingsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Sources: []models.DiagnosticSettingsSource{
			{
				ID:                         keyVaultID,
				Name:                       "kv-prod",
				Type:                       "Microsoft.KeyVault/vaults",
				ResourceGroup:              "rg-sec",
				Location:                   "eastus",
				DiagnosticSettings:         []models.DiagnosticSettingAsset{keyVaultSetting},
				DiagnosticSettingCount:     1,
				EnabledCategories:          []string{"AllMetrics"},
				DisabledCategories:         []string{"AuditEvent"},
				SupportedCategories:        []string{"AllMetrics", "AuditEvent"},
				NotExportedSupported:       []string{"AuditEvent"},
				SupportedCategoryCatalog:   true,
				CategoryGroups:             []string{"AuditEvent"},
				HighSignalCategories:       []string{"AuditEvent"},
				DestinationTypes:           []string{"logAnalytics"},
				HasDiagnosticSettings:      true,
				HasPartialLogPosture:       true,
				HasHighSignalLogPosture:    true,
				HasNonWorkspaceDestination: false,
				RelatedIDs:                 []string{keyVaultID, keyVaultSetting.ID, workspaceID},
			},
			{
				ID:                         storageID,
				Name:                       "stdataprod",
				Type:                       "Microsoft.Storage/storageAccounts",
				ResourceGroup:              "rg-data",
				Location:                   "eastus",
				DiagnosticSettings:         []models.DiagnosticSettingAsset{storageSetting},
				DiagnosticSettingCount:     1,
				EnabledCategories:          []string{"StorageRead", "StorageWrite"},
				SupportedCategories:        []string{"StorageDelete", "StorageRead", "StorageWrite"},
				NotExportedSupported:       []string{"StorageDelete"},
				SupportedCategoryCatalog:   true,
				DestinationTypes:           []string{"eventHubs"},
				HasDiagnosticSettings:      true,
				HasPartialLogPosture:       true,
				HasHighSignalLogPosture:    true,
				HasNonWorkspaceDestination: true,
				RelatedIDs:                 []string{eventHubRuleID, storageID, storageSetting.ID},
			},
			{
				ID:                       appID,
				Name:                     "app-prod",
				Type:                     "Microsoft.Web/sites",
				ResourceGroup:            "rg-app",
				Location:                 "eastus",
				DiagnosticSettings:       []models.DiagnosticSettingAsset{},
				DiagnosticSettingCount:   0,
				SupportedCategories:      []string{"AppServiceHTTPLogs", "AppServiceConsoleLogs"},
				NotExportedSupported:     []string{"AppServiceConsoleLogs", "AppServiceHTTPLogs"},
				SupportedCategoryCatalog: true,
				HasDiagnosticSettings:    false,
				HasHighSignalLogPosture:  true,
				RelatedIDs:               []string{appID},
			},
		},
		Issues: []models.Issue{},
	}
	for index := range facts.Sources {
		facts.Sources[index].Summary = diagnosticSettingsSourceSummary(facts.Sources[index])
	}
	return facts, nil
}
