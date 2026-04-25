package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) Relay(_ context.Context, tenant string, subscription string) (RelayFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID
	namespaceID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-integration/providers/Microsoft.Relay/namespaces/relay-hybrid-prod"
	hybridID := namespaceID + "/hybridConnections/onprem-orders"
	appServiceID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api"
	appServiceHybridID := appServiceID + "/hybridConnectionNamespaces/relay-hybrid-prod/relays/onprem-orders"
	location := "eastus"
	standard := "Standard"
	succeeded := "Succeeded"
	endpoint := "https://relay-hybrid-prod.servicebus.windows.net:443/"
	metricID := namespaceID
	requiresAuth := true
	listenerCount := 1
	hybridCount := 1
	authRules := 2

	return RelayFacts{
		ArtifactIdentityFacts: staticArtifactIdentityFacts(session),
		TenantID:              session.TenantID,
		SubscriptionID:        subscriptionID,
		Namespaces: []models.RelayNamespaceAsset{
			{
				ID:                     namespaceID,
				Name:                   "relay-hybrid-prod",
				ResourceGroup:          "rg-integration",
				Location:               &location,
				SKUName:                &standard,
				ProvisioningState:      &succeeded,
				ServiceBusEndpoint:     &endpoint,
				MetricID:               &metricID,
				HybridConnectionCount:  &hybridCount,
				AuthorizationRuleCount: &authRules,
				HybridConnections: []models.RelayHybridConnectionAsset{
					{
						ID:                          hybridID,
						Name:                        "onprem-orders",
						RequiresClientAuthorization: &requiresAuth,
						ListenerCount:               &listenerCount,
						AppServiceAttachments:       []string{"app-public-api"},
						Summary:                     "Hybrid Connection \"onprem-orders\" is visible under relay namespace \"relay-hybrid-prod\" with App Service attachment(s): app-public-api.",
						RelatedIDs:                  []string{hybridID, appServiceHybridID, appServiceID},
					},
				},
				Summary:    "Relay namespace \"relay-hybrid-prod\" exposes 1 hybrid connection and 2 authorization rule(s), giving Azure a visible cloud rendezvous point for private-path communication.",
				RelatedIDs: []string{namespaceID, hybridID, appServiceHybridID, appServiceID},
			},
		},
		Issues: []models.Issue{},
	}, nil
}
