package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) AppInsights(_ context.Context, tenant string, subscription string) (AppInsightsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID
	componentID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.Insights/components/ai-public-api"
	appID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api"
	functionID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders"

	facts := AppInsightsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Components: []models.AppInsightsComponent{
			{
				ID:                  componentID,
				Name:                "ai-public-api",
				ResourceGroup:       "rg-monitor",
				Location:            "eastus",
				Kind:                models.StringPtr("web"),
				ApplicationType:     models.StringPtr("web"),
				WorkspaceResourceID: models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.OperationalInsights/workspaces/law-soc-prod"),
				IngestionMode:       models.StringPtr("LogAnalytics"),
				Summary:             "Application Insights component \"ai-public-api\" is visible in eastus.",
				RelatedIDs:          []string{componentID},
			},
		},
		Targets: []models.AppInsightsAppTarget{
			{
				ID:                    appID,
				Name:                  "app-public-api",
				Kind:                  "AppService",
				ResourceGroup:         "rg-apps",
				Location:              "eastus",
				InstrumentationClues:  []string{"APPLICATIONINSIGHTS_CONNECTION_STRING"},
				SamplingClues:         []string{"ApplicationInsights__Sampling__Percentage=25"},
				FilteringClues:        []string{"ApplicationInsights__TelemetryProcessor__HealthCheckFilter"},
				LoggingLevelClues:     []string{"Logging__ApplicationInsights__LogLevel__Default=Warning"},
				VisibleTelemetryTypes: []string{"traces"},
				RelatedIDs:            []string{appID},
			},
			{
				ID:                    functionID,
				Name:                  "func-orders",
				Kind:                  "FunctionApp",
				ResourceGroup:         "rg-apps",
				Location:              "eastus",
				InstrumentationClues:  []string{"APPINSIGHTS_INSTRUMENTATIONKEY"},
				SamplingClues:         []string{"AzureFunctionsJobHost__logging__applicationInsights__samplingSettings__isEnabled=true"},
				FilteringClues:        []string{},
				LoggingLevelClues:     []string{},
				VisibleTelemetryTypes: []string{},
				RelatedIDs:            []string{functionID},
			},
		},
		Issues: []models.Issue{},
	}
	for index := range facts.Targets {
		facts.Targets[index].Summary = appInsightsTargetSummary(facts.Targets[index])
	}
	return facts, nil
}
