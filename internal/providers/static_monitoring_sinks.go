package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (provider StaticProvider) MonitoringSinks(ctx context.Context, tenant string, subscription string) (MonitoringSinksFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID
	workspaceID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.OperationalInsights/workspaces/law-soc-prod"
	eventHubRuleID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.EventHub/namespaces/eh-monitor/authorizationRules/send"
	storageID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-data/providers/Microsoft.Storage/storageAccounts/stdataprod"

	dcrFacts, _ := provider.DCR(ctx, tenant, subscription)
	diagnosticFacts, _ := provider.DiagnosticSettings(ctx, tenant, subscription)

	sinks := []models.MonitoringSinkAsset{
		{
			ID:               workspaceID,
			Name:             "law-soc-prod",
			Kind:             "sentinel",
			ResourceType:     "Microsoft.OperationalInsights/workspaces",
			ResourceGroup:    "rg-monitor",
			Location:         "eastus",
			VisibilitySource: "resource inventory",
			SentinelEnabled:  boolPtr(true),
			RelatedIDs:       []string{workspaceID},
		},
		{
			ID:               eventHubRuleID,
			Name:             "send",
			Kind:             "eventHubs",
			ResourceType:     "Microsoft.EventHub/namespaces/authorizationRules",
			ResourceGroup:    "rg-monitor",
			Location:         "eastus",
			VisibilitySource: "declared destination",
			RelatedIDs:       []string{eventHubRuleID},
		},
		{
			ID:               storageID,
			Name:             "stdataprod",
			Kind:             "storage",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			ResourceGroup:    "rg-data",
			Location:         "eastus",
			VisibilitySource: "resource inventory",
			RelatedIDs:       []string{storageID},
		},
	}
	monitoringSinksAttachDCRReferences(sinks, dcrFacts.DCRs)
	monitoringSinksAttachDiagnosticReferences(sinks, diagnosticFacts.Sources)
	sinks = monitoringSinksFinalize(sinks)

	return MonitoringSinksFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Sinks:          sinks,
		Issues:         []models.Issue{},
	}, nil
}
