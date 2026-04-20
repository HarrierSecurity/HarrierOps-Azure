package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"

	"harrierops-azure/internal/models"
)

const armAcrRegistriesAPIVersion = "2025-04-01"

func (provider AzureProvider) Acr(ctx context.Context, tenant string, subscription string) (AcrFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return AcrFacts{}, err
	}

	clientFactory, err := armcontainerregistry.NewClientFactory(session.subscription.ID, session.credential, nil)
	if err != nil {
		return AcrFacts{}, fmt.Errorf("build container registry client factory: %w", err)
	}

	registriesClient := clientFactory.NewRegistriesClient()
	webhooksClient := clientFactory.NewWebhooksClient()
	replicationsClient := clientFactory.NewReplicationsClient()

	acrRegistries := []models.AcrRegistryAsset{}
	issues := []models.Issue{}
	pager := registriesClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("acr.registries", err))
			break
		}
		for _, registry := range page.Value {
			if registry == nil {
				continue
			}

			registryMap := map[string]any{}
			decodeJSONInto(registry, &registryMap)
			registryID := firstNonEmpty(mapStringValue(registryMap, "id"), stringValue(registry.ID))
			resourceGroup, registryName := resourceGroupAndNameFromID(registryID)
			if registryName == "" {
				registryName = stringValue(registry.Name)
			}
			hydrated := registryMap
			if registryID != "" && acrRegistryNeedsHydration(registryMap) {
				fullRegistry, getErr := armGetObject(ctx, session.credential, registryID, armAcrRegistriesAPIVersion)
				if getErr != nil {
					issues = append(issues, issueFromError(acrScope(resourceGroup, registryName, "registry"), getErr))
				} else {
					hydrated = fullRegistry
				}
			}

			var webhooks []map[string]any
			var replications []map[string]any
			if resourceGroup != "" && registryName != "" {
				webhooks, issues = acrWebhookList(ctx, webhooksClient, resourceGroup, registryName, acrScope(resourceGroup, registryName, "webhooks"), issues)
				replications, issues = acrReplicationList(ctx, replicationsClient, resourceGroup, registryName, acrScope(resourceGroup, registryName, "replications"), issues)
			}

			acrRegistries = append(acrRegistries, acrRegistrySummary(hydrated, webhooks, replications))
		}
	}

	return AcrFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		Registries:     acrRegistries,
		Issues:         issues,
	}, nil
}

func acrWebhookList(
	ctx context.Context,
	client *armcontainerregistry.WebhooksClient,
	resourceGroup string,
	registryName string,
	scope string,
	issues []models.Issue,
) ([]map[string]any, []models.Issue) {
	rows := []map[string]any{}
	pager := client.NewListPager(resourceGroup, registryName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError(scope, err))
			return nil, issues
		}
		rows = appendDecodedMaps(rows, page.Value)
	}
	return rows, issues
}

func acrReplicationList(
	ctx context.Context,
	client *armcontainerregistry.ReplicationsClient,
	resourceGroup string,
	registryName string,
	scope string,
	issues []models.Issue,
) ([]map[string]any, []models.Issue) {
	rows := []map[string]any{}
	pager := client.NewListPager(resourceGroup, registryName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError(scope, err))
			return nil, issues
		}
		rows = appendDecodedMaps(rows, page.Value)
	}
	return rows, issues
}

func acrScope(resourceGroup string, registryName string, suffix string) string {
	if resourceGroup != "" && registryName != "" {
		return "acr[" + resourceGroup + "/" + registryName + "]." + suffix
	}
	return "acr." + suffix
}

func acrRegistryNeedsHydration(registry map[string]any) bool {
	properties := mapValue(registry, "properties")
	identity := mapValue(registry, "identity")
	return mapStringValue(properties, "publicNetworkAccess", "public_network_access") == "" ||
		optionalBoolPtr(properties, "anonymousPullEnabled", "anonymous_pull_enabled") == nil ||
		mapStringValue(identity, "type") == ""
}

func acrRegistrySummary(
	registry map[string]any,
	webhooks []map[string]any,
	replications []map[string]any,
) models.AcrRegistryAsset {
	properties := mapValue(registry, "properties")
	identity := mapValue(registry, "identity")
	sku := mapValue(registry, "sku")
	policies := mapValue(properties, "policies")
	networkRuleSet := mapValue(properties, "networkRuleSet", "network_rule_set")
	privateEndpoints := listValue(properties, "privateEndpointConnections", "private_endpoint_connections")
	workloadIdentityIDs := sortedKeys(mapValue(identity, "userAssignedIdentities", "user_assigned_identities"))

	loginServer := stringPtr(mapStringValue(properties, "loginServer", "login_server"))
	workloadIdentityType := stringPtr(mapStringValue(identity, "type"))
	publicNetworkAccess := stringPtr(mapStringValue(properties, "publicNetworkAccess", "public_network_access"))
	networkRuleDefaultAction := stringPtr(mapStringValue(networkRuleSet, "defaultAction", "default_action"))
	networkRuleBypassOptions := stringPtr(mapStringValue(properties, "networkRuleBypassOptions", "network_rule_bypass_options"))
	adminUserEnabled := optionalBoolPtr(properties, "adminUserEnabled", "admin_user_enabled")
	anonymousPullEnabled := optionalBoolPtr(registry, "anonymousPullEnabled", "anonymous_pull_enabled")
	if anonymousPullEnabled == nil {
		anonymousPullEnabled = optionalBoolPtr(properties, "anonymousPullEnabled", "anonymous_pull_enabled")
	}
	dataEndpointEnabled := optionalBoolPtr(properties, "dataEndpointEnabled", "data_endpoint_enabled")
	privateEndpointConnectionCount := acrCountPtr(privateEndpoints)
	webhookCount := acrCountPtrMaps(webhooks)
	enabledWebhookCount := acrEnabledWebhookCount(webhooks)
	webhookActionTypes := acrWebhookActionTypes(webhooks)
	broadWebhookScopeCount := acrBroadWebhookScopeCount(webhooks)
	replicationCount := acrCountPtrMaps(replications)
	replicationRegions := acrReplicationRegions(replications)
	quarantinePolicyStatus := acrNormalizedEnumPtr(mapStringValue(mapValue(policies, "quarantinePolicy", "quarantine_policy"), "status"))
	retentionPolicyStatus := acrNormalizedEnumPtr(mapStringValue(mapValue(policies, "retentionPolicy", "retention_policy"), "status"))
	retentionPolicyDays := acrIntPtr(mapValue(policies, "retentionPolicy", "retention_policy"), "days")
	trustPolicyStatus := acrNormalizedEnumPtr(mapStringValue(mapValue(policies, "trustPolicy", "trust_policy"), "status"))
	trustPolicyType := acrNormalizedEnumPtr(mapStringValue(mapValue(policies, "trustPolicy", "trust_policy"), "type"))

	return models.AcrRegistryAsset{
		ID:                             firstNonEmpty(mapStringValue(registry, "id"), "/unknown/"+firstNonEmpty(mapStringValue(registry, "name"), "unknown")),
		Name:                           firstNonEmpty(mapStringValue(registry, "name"), "unknown"),
		ResourceGroup:                  resourceGroupFromID(mapStringValue(registry, "id")),
		Location:                       stringPtr(mapStringValue(registry, "location")),
		State:                          stringPtr(mapStringValue(properties, "provisioningState", "provisioning_state")),
		LoginServer:                    loginServer,
		SKUName:                        stringPtr(mapStringValue(sku, "name")),
		PublicNetworkAccess:            publicNetworkAccess,
		NetworkRuleDefaultAction:       networkRuleDefaultAction,
		NetworkRuleBypassOptions:       networkRuleBypassOptions,
		AdminUserEnabled:               adminUserEnabled,
		AnonymousPullEnabled:           anonymousPullEnabled,
		DataEndpointEnabled:            dataEndpointEnabled,
		PrivateEndpointConnectionCount: privateEndpointConnectionCount,
		WebhookCount:                   webhookCount,
		EnabledWebhookCount:            enabledWebhookCount,
		WebhookActionTypes:             webhookActionTypes,
		BroadWebhookScopeCount:         broadWebhookScopeCount,
		ReplicationCount:               replicationCount,
		ReplicationRegions:             replicationRegions,
		QuarantinePolicyStatus:         quarantinePolicyStatus,
		RetentionPolicyStatus:          retentionPolicyStatus,
		RetentionPolicyDays:            retentionPolicyDays,
		TrustPolicyStatus:              trustPolicyStatus,
		TrustPolicyType:                trustPolicyType,
		WorkloadIdentityType:           workloadIdentityType,
		WorkloadPrincipalID:            stringPtr(mapStringValue(identity, "principalId", "principal_id")),
		WorkloadClientID:               stringPtr(mapStringValue(identity, "clientId", "client_id")),
		WorkloadIdentityIDs:            workloadIdentityIDs,
		Summary: acrOperatorSummary(
			firstNonEmpty(mapStringValue(registry, "name"), "unknown"),
			loginServer,
			workloadIdentityType,
			publicNetworkAccess,
			networkRuleDefaultAction,
			networkRuleBypassOptions,
			adminUserEnabled,
			anonymousPullEnabled,
			dataEndpointEnabled,
			privateEndpointConnectionCount,
			stringPtr(mapStringValue(sku, "name")),
			webhookCount,
			enabledWebhookCount,
			webhookActionTypes,
			broadWebhookScopeCount,
			replicationCount,
			replicationRegions,
			quarantinePolicyStatus,
			retentionPolicyStatus,
			retentionPolicyDays,
			trustPolicyStatus,
			trustPolicyType,
		),
		RelatedIDs: dedupeStrings(append([]string{
			mapStringValue(registry, "id"),
			mapStringValue(identity, "principalId", "principal_id"),
		}, append(workloadIdentityIDs, append(acrChildIDs(webhooks), acrChildIDs(replications)...)...)...)),
	}
}

func acrCountPtr(items []any) *int {
	count := len(items)
	return &count
}

func acrCountPtrMaps(items []map[string]any) *int {
	if items == nil {
		return nil
	}
	count := len(items)
	return &count
}

func acrEnabledWebhookCount(webhooks []map[string]any) *int {
	if webhooks == nil {
		return nil
	}
	count := 0
	for _, webhook := range webhooks {
		if strings.EqualFold(mapStringValue(mapValue(webhook, "properties"), "status"), "enabled") {
			count++
		}
	}
	return &count
}

func acrWebhookActionTypes(webhooks []map[string]any) []string {
	if webhooks == nil {
		return []string{}
	}
	seen := map[string]struct{}{}
	values := []string{}
	for _, webhook := range webhooks {
		for _, action := range listValue(mapValue(webhook, "properties"), "actions") {
			normalized := strings.TrimSpace(strings.ToLower(stringValue(action)))
			if normalized == "" {
				continue
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			values = append(values, normalized)
		}
	}
	values = dedupeStrings(values)
	sort.Strings(values)
	return values
}

func acrBroadWebhookScopeCount(webhooks []map[string]any) *int {
	if webhooks == nil {
		return nil
	}
	count := 0
	for _, webhook := range webhooks {
		scope := strings.TrimSpace(mapStringValue(mapValue(webhook, "properties"), "scope"))
		if scope == "" || strings.Contains(scope, "*") {
			count++
		}
	}
	return &count
}

func acrReplicationRegions(replications []map[string]any) []string {
	if replications == nil {
		return []string{}
	}
	values := []string{}
	for _, replication := range replications {
		if region := strings.TrimSpace(mapStringValue(replication, "location")); region != "" {
			values = append(values, region)
		}
	}
	values = dedupeStrings(values)
	sort.Strings(values)
	return values
}

func acrChildIDs(items []map[string]any) []string {
	values := []string{}
	for _, item := range items {
		values = append(values, mapStringValue(item, "id"))
	}
	return dedupeStrings(values)
}

func acrIntPtr(input map[string]any, keys ...string) *int {
	for _, key := range keys {
		if _, exists := input[key]; exists {
			value := mapIntValue(input, key)
			return &value
		}
	}
	return nil
}

func acrNormalizedEnumPtr(value string) *string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return nil
	}
	return &value
}

func acrOperatorSummary(
	registryName string,
	loginServer *string,
	workloadIdentityType *string,
	publicNetworkAccess *string,
	networkRuleDefaultAction *string,
	networkRuleBypassOptions *string,
	adminUserEnabled *bool,
	anonymousPullEnabled *bool,
	dataEndpointEnabled *bool,
	privateEndpointConnectionCount *int,
	skuName *string,
	webhookCount *int,
	enabledWebhookCount *int,
	webhookActionTypes []string,
	broadWebhookScopeCount *int,
	replicationCount *int,
	replicationRegions []string,
	quarantinePolicyStatus *string,
	retentionPolicyStatus *string,
	retentionPolicyDays *int,
	trustPolicyStatus *string,
	trustPolicyType *string,
) string {
	loginPhrase := "does not expose a readable login server from the current read path"
	if loginServer != nil && *loginServer != "" {
		loginPhrase = "publishes login server '" + *loginServer + "'"
	}

	identityPhrase := "has no managed identity visible from the current read path"
	if workloadIdentityType != nil && *workloadIdentityType != "" {
		identityPhrase = "uses managed identity (" + *workloadIdentityType + ")"
	}

	authParts := []string{}
	if adminUserEnabled != nil {
		if *adminUserEnabled {
			authParts = append(authParts, "admin user enabled")
		} else {
			authParts = append(authParts, "admin user disabled")
		}
	}
	if anonymousPullEnabled != nil {
		if *anonymousPullEnabled {
			authParts = append(authParts, "anonymous pull enabled")
		} else {
			authParts = append(authParts, "anonymous pull disabled")
		}
	}

	networkParts := []string{"public network access " + valueOrString(publicNetworkAccess)}
	if networkRuleDefaultAction != nil && *networkRuleDefaultAction != "" {
		networkParts = append(networkParts, "default action "+*networkRuleDefaultAction)
	}
	if networkRuleBypassOptions != nil && *networkRuleBypassOptions != "" {
		networkParts = append(networkParts, "bypass "+*networkRuleBypassOptions)
	}
	if privateEndpointConnectionCount != nil && *privateEndpointConnectionCount > 0 {
		networkParts = append(networkParts, acrIntText(*privateEndpointConnectionCount)+" private endpoint(s)")
	} else {
		networkParts = append(networkParts, "no private endpoints visible")
	}

	serviceParts := []string{}
	if skuName != nil && *skuName != "" {
		serviceParts = append(serviceParts, "SKU "+*skuName)
	}
	if dataEndpointEnabled != nil {
		if *dataEndpointEnabled {
			serviceParts = append(serviceParts, "data endpoint enabled")
		} else {
			serviceParts = append(serviceParts, "data endpoint disabled")
		}
	}

	authPhrase := "Auth posture is not fully readable from the current read path."
	if len(authParts) > 0 {
		authPhrase = "Visible auth posture: " + strings.Join(authParts, ", ") + "."
	}
	servicePhrase := "Service shape is not fully readable from the current read path."
	if len(serviceParts) > 0 {
		servicePhrase = "Visible service shape: " + strings.Join(serviceParts, ", ") + "."
	}

	depthParts := []string{}
	if webhookCount != nil {
		webhookPhrase := acrIntText(*webhookCount) + " webhooks"
		if enabledWebhookCount != nil {
			webhookPhrase += " (" + acrIntText(*enabledWebhookCount) + " enabled)"
		}
		depthParts = append(depthParts, webhookPhrase)
	}
	if broadWebhookScopeCount != nil && *broadWebhookScopeCount > 0 {
		depthParts = append(depthParts, acrIntText(*broadWebhookScopeCount)+" broad webhook scope(s)")
	}
	if len(webhookActionTypes) > 0 {
		depthParts = append(depthParts, "webhook actions "+strings.Join(webhookActionTypes, ", "))
	}
	if replicationCount != nil && len(replicationRegions) > 0 {
		depthParts = append(depthParts, acrIntText(*replicationCount)+" replications across "+strings.Join(replicationRegions, ", "))
	} else if replicationCount != nil {
		depthParts = append(depthParts, acrIntText(*replicationCount)+" replications")
	}
	if quarantinePolicyStatus != nil && *quarantinePolicyStatus != "" {
		depthParts = append(depthParts, "quarantine "+*quarantinePolicyStatus)
	}
	if retentionPolicyStatus != nil && *retentionPolicyStatus != "" {
		if strings.EqualFold(*retentionPolicyStatus, "enabled") && retentionPolicyDays != nil {
			depthParts = append(depthParts, "retention enabled ("+acrIntText(*retentionPolicyDays)+"d)")
		} else {
			depthParts = append(depthParts, "retention "+*retentionPolicyStatus)
		}
	}
	if trustPolicyStatus != nil && *trustPolicyStatus != "" {
		if strings.EqualFold(*trustPolicyStatus, "enabled") && trustPolicyType != nil && *trustPolicyType != "" {
			depthParts = append(depthParts, "content trust enabled ("+*trustPolicyType+")")
		} else {
			depthParts = append(depthParts, "content trust "+*trustPolicyStatus)
		}
	}

	depthPhrase := ""
	if len(depthParts) > 0 {
		depthPhrase = " Depth cues: " + strings.Join(depthParts, ", ") + "."
	}

	return "Container Registry '" + registryName + "' " + loginPhrase + " and " + identityPhrase + ". " +
		authPhrase + " Visible network posture: " + strings.Join(networkParts, ", ") + ". " +
		servicePhrase + depthPhrase
}

func valueOrString(value *string) string {
	if value == nil {
		return "unknown"
	}
	return *value
}

func acrIntText(value int) string {
	return stringValue(value)
}
