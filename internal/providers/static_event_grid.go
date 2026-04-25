package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) EventGrid(_ context.Context, tenant string, subscription string) (EventGridFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return EventGridFacts{
		ArtifactIdentityFacts: staticArtifactIdentityFacts(session),
		TenantID:              session.TenantID,
		SubscriptionID:        subscriptionID,
		Routes: []models.EventGridRouteAsset{
			{
				ID:                  "/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stlanding/providers/Microsoft.EventGrid/eventSubscriptions/to-function",
				Name:                "to-function",
				DestinationType:     "AzureFunction",
				Classification:      "execution-capable",
				SourceID:            "/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stlanding",
				SourceType:          "Microsoft.Storage/storageAccounts",
				DestinationTargetID: models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-app/providers/Microsoft.Web/sites/fa-ingest/functions/BlobCreated"),
				ProvisioningState:   models.StringPtr("Succeeded"),
				EventDeliverySchema: models.StringPtr("EventGridSchema"),
				IncludedEventTypes:  []string{"Microsoft.Storage.BlobCreated"},
				Summary:             "Storage account events are visibly routed into an Azure Function destination, so this path already looks execution-capable from the current control-plane read path.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stlanding",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-app/providers/Microsoft.Web/sites/fa-ingest/functions/BlobCreated",
				},
			},
			{
				ID:                  "/subscriptions/" + subscriptionID + "/resourceGroups/rg-integration/providers/Microsoft.EventGrid/topics/ops-topic/providers/Microsoft.EventGrid/eventSubscriptions/to-webhook",
				Name:                "to-webhook",
				DestinationType:     "WebHook",
				Classification:      "external-callback",
				SourceID:            "/subscriptions/" + subscriptionID + "/resourceGroups/rg-integration/providers/Microsoft.EventGrid/topics/ops-topic",
				SourceType:          "Microsoft.EventGrid/topics",
				ExternalDelivery:    true,
				ProvisioningState:   models.StringPtr("Succeeded"),
				EventDeliverySchema: models.StringPtr("CloudEventSchemaV1_0"),
				IncludedEventTypes:  []string{"All"},
				Summary:             "Custom topic events are visibly delivered to a webhook destination. The base command keeps the callback target redacted, but this route already crosses the normal Azure service boundary.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-integration/providers/Microsoft.EventGrid/topics/ops-topic",
				},
			},
			{
				ID:                  "/subscriptions/" + subscriptionID + "/providers/Microsoft.EventGrid/eventSubscriptions/subscription-to-queue",
				Name:                "subscription-to-queue",
				DestinationType:     "StorageQueue",
				Classification:      "supporting-context",
				SourceID:            "/subscriptions/" + subscriptionID,
				SourceType:          "subscription",
				DestinationTargetID: models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-queue/providers/Microsoft.Storage/storageAccounts/stqueue/queueServices/default/queues/incoming-events"),
				ProvisioningState:   models.StringPtr("Succeeded"),
				EventDeliverySchema: models.StringPtr("EventGridSchema"),
				IncludedEventTypes:  []string{"Microsoft.Resources.ResourceWriteSuccess"},
				Summary:             "Subscription-scoped resource events are visibly buffered into Storage Queue. This is useful trigger context, but the current read path does not yet show the consumer that would turn the queue into execution.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID,
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-queue/providers/Microsoft.Storage/storageAccounts/stqueue/queueServices/default/queues/incoming-events",
				},
			},
		},
		Issues: []models.Issue{},
	}, nil
}
