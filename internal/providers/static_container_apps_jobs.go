package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) ContainerAppsJobs(_ context.Context, tenant string, subscription string) (ContainerAppsJobsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	acaEnvProdID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/managedEnvironments/aca-env-prod"
	acaEnvInternalID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/managedEnvironments/aca-env-internal"
	uaContainerJobsID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-container-jobs"
	scheduleTrigger := "Schedule"
	eventTrigger := "Event"
	systemAssigned := "SystemAssigned"
	systemAndUserAssigned := "SystemAssigned, UserAssigned"
	parallelismOne := 1
	parallelismTwo := 2
	completionOne := 1
	completionTwo := 2
	retryThree := 3
	timeout1800 := 1800
	timeout3600 := 3600
	secretCountTwo := 2
	secretCountOne := 1
	keyVaultSecretCountOne := 1
	keyVaultSecretCountZero := 0
	registryIdentityOne := 1
	registryIdentityZero := 0
	registryPasswordRefOne := 1
	registryPasswordRefZero := 0

	return ContainerAppsJobsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		ContainerAppsJobs: []models.ContainerAppsJobAsset{
			{
				Command:         []string{"reconcile: /app/reconcile --tenant prod"},
				ContainerImages: []string{"ghcr.io/harrierops/jobs/reconcile:1.4.2"},
				EnvironmentID:   &acaEnvProdID,
				EventRules:      []models.ContainerAppsJobEventRule{},
				ID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/" +
					"Microsoft.App/jobs/nightly-reconcile",
				KeyVaultSecretCount:      &keyVaultSecretCountOne,
				Location:                 "eastus",
				Name:                     "nightly-reconcile",
				Parallelism:              &parallelismOne,
				RegistryIdentityCount:    &registryIdentityOne,
				RegistryPasswordRefCount: &registryPasswordRefZero,
				RegistryServers:          []string{"ghcr.io"},
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/jobs/nightly-reconcile",
					acaEnvProdID,
					"abab3333-3333-3333-3333-333333333333",
				},
				ReplicaCompletionCount: &completionOne,
				ReplicaRetryLimit:      &retryThree,
				ReplicaTimeout:         &timeout1800,
				ResourceGroup:          "rg-containers",
				ScheduleExpression:     models.StringPtr("0 3 * * *"),
				SecretCount:            &secretCountTwo,
				Summary:                "Container Apps job 'nightly-reconcile' uses Schedule trigger with schedule '0 3 * * *', stores 1 container image clue(s), and uses managed identity (SystemAssigned). Safe posture: secrets 2, Key Vault-backed secrets 1, registry servers 1, registry identity refs 1.",
				TriggerType:            &scheduleTrigger,
				WorkloadClientID:       models.StringPtr("cdcd3333-3333-3333-3333-333333333333"),
				WorkloadIdentityIDs:    []string{},
				WorkloadIdentityType:   &systemAssigned,
				WorkloadPrincipalID:    models.StringPtr("abab3333-3333-3333-3333-333333333333"),
			},
			{
				Command:       []string{"worker: /app/drain --queue orders"},
				EnvironmentID: &acaEnvInternalID,
				EventRules: []models.ContainerAppsJobEventRule{
					{
						AuthSecretRefs: []string{"queue-connection"},
						Identity:       models.StringPtr(uaContainerJobsID),
						Name:           "orders-queue",
						Type:           "azure-queue",
					},
				},
				ContainerImages: []string{"contoso.azurecr.io/jobs/queue-drain:2026.04"},
				ID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/" +
					"Microsoft.App/jobs/queue-drain",
				KeyVaultSecretCount:      &keyVaultSecretCountZero,
				Location:                 "eastus",
				Name:                     "queue-drain",
				Parallelism:              &parallelismTwo,
				RegistryIdentityCount:    &registryIdentityZero,
				RegistryPasswordRefCount: &registryPasswordRefOne,
				RegistryServers:          []string{"contoso.azurecr.io"},
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/jobs/queue-drain",
					acaEnvInternalID,
					"abab2222-2222-2222-2222-222222222222",
					uaContainerJobsID,
				},
				ReplicaCompletionCount: &completionTwo,
				ReplicaRetryLimit:      &retryThree,
				ReplicaTimeout:         &timeout3600,
				ResourceGroup:          "rg-containers",
				SecretCount:            &secretCountOne,
				Summary:                "Container Apps job 'queue-drain' uses Event trigger with 1 event scale rule(s), stores 1 container image clue(s), and uses managed identity (SystemAssigned, UserAssigned). Safe posture: secrets 1, registry servers 1.",
				TriggerType:            &eventTrigger,
				WorkloadClientID:       models.StringPtr("cdcd2222-2222-2222-2222-222222222222"),
				WorkloadIdentityIDs: []string{
					uaContainerJobsID,
				},
				WorkloadIdentityType: &systemAndUserAssigned,
				WorkloadPrincipalID:  models.StringPtr("abab2222-2222-2222-2222-222222222222"),
			},
		},
		Issues: []models.Issue{},
	}, nil
}
