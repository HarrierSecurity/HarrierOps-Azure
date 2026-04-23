package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) WebJobs(_ context.Context, tenant string, subscription string) (WebJobsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	systemAssigned := "SystemAssigned"
	running := "Running"
	success := "Success"
	schedule := "Schedule"
	nightlySchedule := "0 0 * * * *"

	return WebJobsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		WebJobs: []models.WebJobAsset{
			{
				DetailedStatus:     models.StringPtr("Polling queue messages"),
				ID:                 "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api/continuouswebjobs/queue-worker",
				JobType:            models.StringPtr("Continuous"),
				Location:           "eastus",
				Mode:               "continuous",
				Name:               "queue-worker",
				ParentAppID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
				ParentAppName:      "app-public-api",
				ParentHostname:     models.StringPtr("app-public-api.azurewebsites.net"),
				ParentIdentityIDs:  []string{},
				ParentIdentityType: &systemAssigned,
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api/continuouswebjobs/queue-worker",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
					"aaaa1111-1111-1111-1111-111111111111",
				},
				ResourceGroup: "rg-apps",
				RunCommand:    models.StringPtr("node /home/site/wwwroot/app_data/jobs/continuous/queue-worker/index.js"),
				Status:        &running,
				Summary:       "WebJob 'queue-worker' is a continuous WebJob under App Service 'app-public-api'; status Running, run command 'node /home/site/wwwroot/app_data/jobs/continuous/queue-worker/index.js'; parent App Service 'app-public-api' publishes hostname 'app-public-api.azurewebsites.net'; the parent app uses managed identity (SystemAssigned).",
			},
			{
				ID:                 "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api/triggeredwebjobs/nightly-reconcile",
				JobType:            models.StringPtr("Triggered"),
				LatestRunStatus:    &success,
				LatestRunTrigger:   &schedule,
				Location:           "eastus",
				Mode:               "scheduled",
				Name:               "nightly-reconcile",
				ParentAppID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
				ParentAppName:      "app-public-api",
				ParentHostname:     models.StringPtr("app-public-api.azurewebsites.net"),
				ParentIdentityIDs:  []string{},
				ParentIdentityType: &systemAssigned,
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api/triggeredwebjobs/nightly-reconcile",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
					"aaaa1111-1111-1111-1111-111111111111",
				},
				ResourceGroup:    "rg-apps",
				RunCommand:       models.StringPtr("python /home/site/wwwroot/app_data/jobs/triggered/nightly-reconcile/run.py"),
				ScheduleExpression: &nightlySchedule,
				SchedulerLogsURL: models.StringPtr("https://app-public-api.scm.azurewebsites.net/api/triggeredwebjobs/nightly-reconcile/history"),
				Status:           &success,
				Summary:          "WebJob 'nightly-reconcile' is a scheduled WebJob under App Service 'app-public-api'; status Success, latest visible trigger Schedule, schedule '0 0 * * * *', run command 'python /home/site/wwwroot/app_data/jobs/triggered/nightly-reconcile/run.py'; parent App Service 'app-public-api' publishes hostname 'app-public-api.azurewebsites.net'; the parent app uses managed identity (SystemAssigned).",
			},
		},
		Issues: []models.Issue{},
	}, nil
}
