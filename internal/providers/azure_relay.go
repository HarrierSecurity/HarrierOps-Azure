package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

const armRelayAPIVersion = "2021-11-01"

func (provider AzureProvider) Relay(ctx context.Context, tenant string, subscription string) (RelayFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return RelayFacts{}, err
	}

	namespaces, err := armListObjects(ctx, session.credential, "/subscriptions/"+session.subscription.ID+"/providers/Microsoft.Relay/namespaces", armRelayAPIVersion)
	if err != nil {
		return RelayFacts{}, err
	}

	rows := []models.RelayNamespaceAsset{}
	issues := []models.Issue{}
	attachments := relayAppServiceHybridConnectionAttachments(ctx, session, &issues)
	for _, namespace := range namespaces {
		namespaceID := mapStringValue(namespace, "id")
		resourceGroup, namespaceName := resourceGroupAndNameFromID(namespaceID)
		hybridConnections := []models.RelayHybridConnectionAsset{}
		authRules := []map[string]any{}
		if resourceGroup != "" && namespaceName != "" {
			hybridPath := namespaceID + "/hybridConnections"
			items, listErr := armListObjects(ctx, session.credential, hybridPath, armRelayAPIVersion)
			if listErr != nil {
				issues = append(issues, issueFromError("relay["+resourceGroup+"/"+namespaceName+"].hybrid_connections", listErr))
			} else {
				for _, item := range items {
					hybridConnections = append(hybridConnections, relayHybridConnectionAsset(namespaceName, item, attachments))
				}
			}
			rules, rulesErr := armListObjects(ctx, session.credential, namespaceID+"/authorizationRules", armRelayAPIVersion)
			if rulesErr != nil {
				issues = append(issues, issueFromError("relay["+resourceGroup+"/"+namespaceName+"].authorization_rules", rulesErr))
			} else {
				authRules = rules
			}
		}
		rows = append(rows, relayNamespaceAsset(namespace, hybridConnections, authRules))
	}

	return RelayFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		Namespaces:     rows,
		Issues:         issues,
	}, nil
}

func relayNamespaceAsset(namespace map[string]any, hybridConnections []models.RelayHybridConnectionAsset, authRules []map[string]any) models.RelayNamespaceAsset {
	id := mapStringValue(namespace, "id")
	properties := mapValue(namespace, "properties")
	sku := mapValue(namespace, "sku")
	hybridCount := len(hybridConnections)
	authRuleCount := len(authRules)
	name := firstNonEmpty(mapStringValue(namespace, "name"), resourceNameFromID(id), "unknown")
	return models.RelayNamespaceAsset{
		ID:                     id,
		Name:                   name,
		ResourceGroup:          resourceGroupFromID(id),
		Location:               stringPtr(mapStringValue(namespace, "location")),
		SKUName:                stringPtr(mapStringValue(sku, "name")),
		ProvisioningState:      stringPtr(mapStringValue(properties, "provisioningState", "provisioning_state")),
		ServiceBusEndpoint:     stringPtr(mapStringValue(properties, "serviceBusEndpoint", "service_bus_endpoint")),
		MetricID:               stringPtr(mapStringValue(properties, "metricId", "metric_id")),
		HybridConnectionCount:  &hybridCount,
		AuthorizationRuleCount: &authRuleCount,
		HybridConnections:      append([]models.RelayHybridConnectionAsset{}, hybridConnections...),
		Summary:                relayNamespaceSummary(name, hybridCount, authRuleCount),
		RelatedIDs:             relayRelatedIDs(id, hybridConnections),
	}
}

func relayHybridConnectionAsset(namespaceName string, item map[string]any, attachments []relayAppServiceHybridConnectionAttachment) models.RelayHybridConnectionAsset {
	id := mapStringValue(item, "id")
	properties := mapValue(item, "properties")
	name := firstNonEmpty(mapStringValue(item, "name"), resourceNameFromID(id), "unknown")
	apps, relatedIDs := relayAppServiceAttachmentNames(namespaceName, name, attachments)
	return models.RelayHybridConnectionAsset{
		ID:                          id,
		Name:                        name,
		RequiresClientAuthorization: optionalBoolPtr(properties, "requiresClientAuthorization", "requires_client_authorization"),
		UserMetadata:                stringPtr(mapStringValue(properties, "userMetadata", "user_metadata")),
		ListenerCount:               optionalIntPtr(properties, "listenerCount", "listener_count"),
		AppServiceAttachments:       apps,
		Summary:                     relayHybridConnectionSummary(namespaceName, name, apps),
		RelatedIDs:                  dedupeStrings(append([]string{id}, relatedIDs...)),
	}
}

func relayHybridConnectionSummary(namespaceName string, connectionName string, attachments []string) string {
	summary := fmt.Sprintf("Hybrid Connection %q is visible under relay namespace %q", connectionName, namespaceName)
	if len(attachments) > 0 {
		summary += fmt.Sprintf(" with App Service attachment(s): %s", strings.Join(attachments, ", "))
	}
	return summary + "."
}

func relayNamespaceSummary(name string, hybridCount int, authRuleCount int) string {
	parts := []string{fmt.Sprintf("Relay namespace %q exposes %d hybrid connection(s)", name, hybridCount)}
	if authRuleCount > 0 {
		parts = append(parts, fmt.Sprintf("%d authorization rule(s)", authRuleCount))
	}
	return strings.Join(parts, " and ") + "."
}

func relayRelatedIDs(namespaceID string, hybridConnections []models.RelayHybridConnectionAsset) []string {
	values := []string{namespaceID}
	for _, connection := range hybridConnections {
		values = append(values, connection.ID)
		values = append(values, connection.RelatedIDs...)
	}
	return dedupeStrings(values)
}

type relayAppServiceHybridConnectionAttachment struct {
	ID               string
	AppServiceName   string
	AppServiceID     string
	RelayNamespace   string
	HybridConnection string
}

func relayAppServiceHybridConnectionAttachments(ctx context.Context, session azureSession, issues *[]models.Issue) []relayAppServiceHybridConnectionAttachment {
	resources, err := armListObjects(ctx, session.credential, "/subscriptions/"+session.subscription.ID+"/resources", armResourcesAPIVersion)
	if err != nil {
		*issues = append(*issues, issueFromError("relay.app_service_hybrid_connections", err))
		return []relayAppServiceHybridConnectionAttachment{}
	}
	attachments := []relayAppServiceHybridConnectionAttachment{}
	for _, resource := range resources {
		if !strings.EqualFold(mapStringValue(resource, "type"), "Microsoft.Web/sites/hybridConnectionNamespaces/relays") {
			continue
		}
		if attachment := relayAppServiceHybridConnectionAttachmentFromResource(resource); attachment.ID != "" {
			attachments = append(attachments, attachment)
		}
	}
	return attachments
}

func relayAppServiceHybridConnectionAttachmentFromResource(resource map[string]any) relayAppServiceHybridConnectionAttachment {
	id := mapStringValue(resource, "id")
	nameParts := strings.Split(mapStringValue(resource, "name"), "/")
	properties := mapValue(resource, "properties")
	appName := relayAppServiceNameFromHybridConnectionID(id)
	namespaceName := firstNonEmpty(mapStringValue(properties, "serviceBusNamespace"), relayAppServiceNamespaceFromID(id))
	connectionName := firstNonEmpty(mapStringValue(properties, "relayName"), resourceNameFromID(id))
	if len(nameParts) >= 3 {
		appName = firstNonEmpty(appName, nameParts[0])
		namespaceName = firstNonEmpty(namespaceName, nameParts[1])
		connectionName = firstNonEmpty(connectionName, nameParts[2])
	}
	return relayAppServiceHybridConnectionAttachment{
		ID:               id,
		AppServiceName:   appName,
		AppServiceID:     relayAppServiceIDFromHybridConnectionID(id),
		RelayNamespace:   namespaceName,
		HybridConnection: connectionName,
	}
}

func relayAppServiceAttachmentNames(namespaceName string, connectionName string, attachments []relayAppServiceHybridConnectionAttachment) ([]string, []string) {
	names := []string{}
	relatedIDs := []string{}
	for _, attachment := range attachments {
		if !strings.EqualFold(attachment.RelayNamespace, namespaceName) || !strings.EqualFold(attachment.HybridConnection, connectionName) {
			continue
		}
		names = append(names, attachment.AppServiceName)
		relatedIDs = append(relatedIDs, attachment.ID, attachment.AppServiceID)
	}
	sort.Strings(names)
	return dedupeStrings(names), dedupeStrings(relatedIDs)
}

func relayAppServiceIDFromHybridConnectionID(id string) string {
	marker := "/hybridConnectionNamespaces/"
	if index := strings.Index(strings.ToLower(id), strings.ToLower(marker)); index > 0 {
		return id[:index]
	}
	return ""
}

func relayAppServiceNameFromHybridConnectionID(id string) string {
	return resourceNameFromID(relayAppServiceIDFromHybridConnectionID(id))
}

func relayAppServiceNamespaceFromID(id string) string {
	parts := strings.Split(id, "/")
	for index, part := range parts {
		if strings.EqualFold(part, "hybridConnectionNamespaces") && index+1 < len(parts) {
			return parts[index+1]
		}
	}
	return ""
}
