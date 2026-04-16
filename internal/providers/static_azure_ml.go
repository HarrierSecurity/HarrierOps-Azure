package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) AzureML(_ context.Context, tenant string, subscription string) (AzureMLFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return AzureMLFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Workspaces: []models.AzureMLWorkspaceAsset{
			{
				ID:                   "/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub",
				Name:                 "ml-ops-hub",
				Classification:       "execution-capable",
				ResourceGroup:        "rg-ml",
				Location:             models.StringPtr("eastus"),
				WorkspaceKind:        models.StringPtr("Default"),
				State:                models.StringPtr("Succeeded"),
				PublicNetworkAccess:  models.StringPtr("Enabled"),
				IdentityType:         models.StringPtr("SystemAssigned,UserAssigned"),
				PrincipalID:          models.StringPtr("56565656-5656-5656-5656-565656565656"),
				IdentityIDs:          []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub/identities/system", "/subscriptions/" + subscriptionID + "/resourceGroups/rg-id/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-ml-ops"},
				ComputeCount:         2,
				ComputeTypes:         []string{"ComputeCluster", "ComputeInstance"},
				JobCount:             2,
				JobTypes:             []string{"Command", "Pipeline"},
				ScheduleCount:        1,
				ScheduleTriggerTypes: []string{"Cron"},
				EndpointCount:        1,
				EndpointAuthModes:    []string{"AADToken"},
				EndpointPublicAccess: []string{"Enabled"},
				DatastoreCount:       2,
				DatastoreTypes:       []string{"AzureBlob", "AzureDataLakeGen2"},
				StorageAccountID:     models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stamlops"),
				KeyVaultID:           models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-sec/providers/Microsoft.KeyVault/vaults/kv-amlops"),
				ContainerRegistryID:  models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-app/providers/Microsoft.ContainerRegistry/registries/cramlops"),
				Summary:              "Visible Azure ML workspace already shows execution-capable runtime surfaces through compute, jobs, and an online endpoint. Cron-backed scheduling is also visible, and the workspace carries managed identity plus linked datastore and storage context for follow-up.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub/identities/system",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-id/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-ml-ops",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stamlops",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-sec/providers/Microsoft.KeyVault/vaults/kv-amlops",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-app/providers/Microsoft.ContainerRegistry/registries/cramlops",
				},
			},
			{
				ID:                   "/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-nightly-train",
				Name:                 "ml-nightly-train",
				Classification:       "supporting-persistence-context",
				ResourceGroup:        "rg-ml",
				Location:             models.StringPtr("centralus"),
				WorkspaceKind:        models.StringPtr("Default"),
				State:                models.StringPtr("Succeeded"),
				PublicNetworkAccess:  models.StringPtr("Disabled"),
				ScheduleCount:        1,
				ScheduleTriggerTypes: []string{"Recurrence"},
				DatastoreCount:       1,
				DatastoreTypes:       []string{"AzureBlob"},
				StorageAccountID:     models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stmltrain"),
				Summary:              "Visible recurrence-backed scheduling makes this workspace relevant as persistence-adjacent ML context. The current control-plane read path does not yet prove a stronger compute, job, or serving surface behind that schedule.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-nightly-train",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stmltrain",
				},
			},
			{
				ID:                    "/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-catalog",
				Name:                  "ml-catalog",
				Classification:        "supporting-context",
				ResourceGroup:         "rg-ml",
				Location:              models.StringPtr("westus2"),
				WorkspaceKind:         models.StringPtr("Default"),
				State:                 models.StringPtr("Succeeded"),
				PublicNetworkAccess:   models.StringPtr("Enabled"),
				DatastoreCount:        2,
				DatastoreTypes:        []string{"AzureBlob", "AzureDataLakeGen2"},
				StorageAccountID:      models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stmlcatalog"),
				ApplicationInsightsID: models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.Insights/components/ai-mlcatalog"),
				Summary:               "Visible Azure ML workspace currently reads more like supporting context than an active execution surface. Storage-linked and datastore relationships are visible, but no compute, job, schedule, or online endpoint is confirmed from the current read path.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-catalog",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stmlcatalog",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.Insights/components/ai-mlcatalog",
				},
			},
		},
		Issues: []models.Issue{},
	}, nil
}
