package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) LogicApps(_ context.Context, tenant string, subscription string) (LogicAppsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return LogicAppsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Workflows: []models.LogicAppWorkflowAsset{
			{
				ID:                               "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workflow/providers/Microsoft.Logic/workflows/la-inbound-redeploy",
				Name:                             "la-inbound-redeploy",
				Classification:                   "persistence-capable",
				ResourceGroup:                    "rg-workflow",
				Location:                         models.StringPtr("centralus"),
				Platform:                         models.StringPtr("Consumption"),
				State:                            models.StringPtr("Enabled"),
				IdentityType:                     models.StringPtr("SystemAssigned"),
				PrincipalID:                      models.StringPtr("56565656-5656-5656-5656-565656565656"),
				ClientID:                         models.StringPtr("78787878-7878-7878-7878-787878787878"),
				IdentityIDs:                      []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workflow/providers/Microsoft.Logic/workflows/la-inbound-redeploy/identities/system"},
				TriggerTypes:                     []string{"request"},
				ExternallyCallableRequestTrigger: true,
				DownstreamActionKinds:            []string{"automation", "external-http"},
				Summary:                          "Request trigger is visible from workflow definition, so this Logic App already looks like a callable re-entry path. Workflow uses managed identity (SystemAssigned), and visible actions touch Azure Automation and external HTTP destinations.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workflow/providers/Microsoft.Logic/workflows/la-inbound-redeploy",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workflow/providers/Microsoft.Logic/workflows/la-inbound-redeploy/identities/system",
				},
			},
			{
				ID:                    "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workflow/providers/Microsoft.Logic/workflows/la-nightly-sync",
				Name:                  "la-nightly-sync",
				Classification:        "persistence-capable",
				ResourceGroup:         "rg-workflow",
				Location:              models.StringPtr("centralus"),
				Platform:              models.StringPtr("Consumption"),
				State:                 models.StringPtr("Enabled"),
				TriggerTypes:          []string{"recurrence"},
				RecurrenceSummary:     models.StringPtr("Day/1"),
				DownstreamActionKinds: []string{"storage", "connector"},
				Summary:               "Recurrence is visible from workflow definition (Day/1), so Azure already has a durable schedule for this workflow. Visible downstream actions touch storage and connector-backed service paths, but no workflow identity is exposed from the current read path.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workflow/providers/Microsoft.Logic/workflows/la-nightly-sync",
				},
			},
			{
				ID:                    "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workflow/providers/Microsoft.Logic/workflows/la-event-router",
				Name:                  "la-event-router",
				Classification:        "execution-capable-only",
				ResourceGroup:         "rg-workflow",
				Location:              models.StringPtr("eastus"),
				Platform:              models.StringPtr("Consumption"),
				State:                 models.StringPtr("Enabled"),
				IdentityType:          models.StringPtr("UserAssigned"),
				PrincipalID:           models.StringPtr("90909090-9090-9090-9090-909090909090"),
				ClientID:              models.StringPtr("abababab-abab-abab-abab-abababababab"),
				IdentityIDs:           []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identity/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-workflow-router"},
				TriggerTypes:          []string{"api-connection"},
				DownstreamActionKinds: []string{"function", "messaging"},
				Summary:               "Visible trigger and action posture suggest workflow-driven execution, but the current definition does not yet show a durable request or recurrence trigger. Workflow uses a user-assigned managed identity and visibly reaches Azure Functions and messaging paths.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workflow/providers/Microsoft.Logic/workflows/la-event-router",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identity/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-workflow-router",
				},
			},
		},
		Issues: []models.Issue{},
	}, nil
}
