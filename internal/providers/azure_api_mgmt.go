package providers

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"

	"harrierops-azure/internal/models"
)

const armApiManagementAPIVersion = "2024-05-01"

func (provider AzureProvider) ApiMgmt(ctx context.Context, tenant string, subscription string) (ApiMgmtFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return ApiMgmtFacts{}, err
	}

	clientFactory, err := armapimanagement.NewClientFactory(session.subscription.ID, session.credential, nil)
	if err != nil {
		return ApiMgmtFacts{}, fmt.Errorf("build api management client factory: %w", err)
	}

	serviceClient := clientFactory.NewServiceClient()
	apiClient := clientFactory.NewAPIClient()
	subscriptionClient := clientFactory.NewSubscriptionClient()
	backendClient := clientFactory.NewBackendClient()
	namedValueClient := clientFactory.NewNamedValueClient()

	apiMgmtServices := []models.ApiMgmtServiceAsset{}
	issues := []models.Issue{}
	pager := serviceClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("api_mgmt.services", err))
			break
		}
		for _, service := range page.Value {
			if service == nil {
				continue
			}

			serviceMap := map[string]any{}
			decodeJSONInto(service, &serviceMap)
			serviceID := firstNonEmpty(mapStringValue(serviceMap, "id"), stringValue(service.ID))
			resourceGroup, serviceName := resourceGroupAndNameFromID(serviceID)
			if serviceName == "" {
				serviceName = stringValue(service.Name)
			}
			hydrated := serviceMap
			if resourceGroup != "" && serviceName != "" {
				fullService, getErr := serviceClient.Get(ctx, resourceGroup, serviceName, nil)
				if getErr != nil {
					issues = append(issues, issueFromError("api_mgmt["+resourceGroup+"/"+serviceName+"].service", getErr))
				} else {
					hydrated = map[string]any{}
					decodeJSONInto(fullService.ServiceResource, &hydrated)
				}
			}

			var apis []map[string]any
			var subscriptions []map[string]any
			var backends []map[string]any
			var namedValues []map[string]any
			var policies []map[string]any

			if resourceGroup != "" && serviceName != "" {
				scopePrefix := "api_mgmt[" + resourceGroup + "/" + serviceName + "]."
				apis, issues = apiMgmtAPIList(ctx, apiClient, resourceGroup, serviceName, scopePrefix+"apis", issues)
				subscriptions, issues = apiMgmtSubscriptionList(ctx, subscriptionClient, resourceGroup, serviceName, scopePrefix+"subscriptions", issues)
				backends, issues = apiMgmtBackendList(ctx, backendClient, resourceGroup, serviceName, scopePrefix+"backends", issues)
				namedValues, issues = apiMgmtNamedValueList(ctx, namedValueClient, resourceGroup, serviceName, scopePrefix+"named_values", issues)
				policies, issues = apiMgmtPolicyList(ctx, session, serviceID, apis, scopePrefix+"policies", issues)
			}

			apiMgmtServices = append(apiMgmtServices, apiMgmtServiceSummary(hydrated, apis, subscriptions, backends, namedValues, policies))
		}
	}

	return ApiMgmtFacts{
		TenantID:              session.tenantID,
		SubscriptionID:        session.subscription.ID,
		ApiManagementServices: apiMgmtServices,
		Issues:                issues,
	}, nil
}

func apiMgmtAPIList(
	ctx context.Context,
	client *armapimanagement.APIClient,
	resourceGroup string,
	serviceName string,
	scope string,
	issues []models.Issue,
) ([]map[string]any, []models.Issue) {
	rows := []map[string]any{}
	pager := client.NewListByServicePager(resourceGroup, serviceName, nil)
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

func apiMgmtSubscriptionList(
	ctx context.Context,
	client *armapimanagement.SubscriptionClient,
	resourceGroup string,
	serviceName string,
	scope string,
	issues []models.Issue,
) ([]map[string]any, []models.Issue) {
	rows := []map[string]any{}
	pager := client.NewListPager(resourceGroup, serviceName, nil)
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

func apiMgmtBackendList(
	ctx context.Context,
	client *armapimanagement.BackendClient,
	resourceGroup string,
	serviceName string,
	scope string,
	issues []models.Issue,
) ([]map[string]any, []models.Issue) {
	rows := []map[string]any{}
	pager := client.NewListByServicePager(resourceGroup, serviceName, nil)
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

func apiMgmtNamedValueList(
	ctx context.Context,
	client *armapimanagement.NamedValueClient,
	resourceGroup string,
	serviceName string,
	scope string,
	issues []models.Issue,
) ([]map[string]any, []models.Issue) {
	rows := []map[string]any{}
	pager := client.NewListByServicePager(resourceGroup, serviceName, nil)
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

func apiMgmtPolicyList(
	ctx context.Context,
	session azureSession,
	serviceID string,
	apis []map[string]any,
	scope string,
	issues []models.Issue,
) ([]map[string]any, []models.Issue) {
	policies := []map[string]any{}
	servicePolicies, err := armListObjects(ctx, session.credential, strings.TrimRight(serviceID, "/")+"/policies", armApiManagementAPIVersion)
	if err != nil {
		issues = append(issues, issueFromError(scope+".service", err))
	} else {
		policies = append(policies, servicePolicies...)
	}
	for _, api := range apis {
		apiID := mapStringValue(api, "id")
		if apiID == "" {
			continue
		}
		apiPolicies, err := armListObjects(ctx, session.credential, strings.TrimRight(apiID, "/")+"/policies", armApiManagementAPIVersion)
		if err != nil {
			issues = append(issues, issueFromError(scope+"["+firstNonEmpty(mapStringValue(api, "name"), resourceNameFromID(apiID), "api")+"].api", err))
			continue
		}
		policies = append(policies, apiPolicies...)
	}
	return policies, issues
}

func apiMgmtServiceSummary(
	service map[string]any,
	apis []map[string]any,
	subscriptions []map[string]any,
	backends []map[string]any,
	namedValues []map[string]any,
	policies []map[string]any,
) models.ApiMgmtServiceAsset {
	properties := mapValue(service, "properties")
	identity := mapValue(service, "identity")
	sku := mapValue(service, "sku")
	hostnameConfigurations := listValue(properties, "hostnameConfigurations", "hostname_configurations")
	gatewayHostnames, managementHostnames, portalHostnames := apiMgmtHostnames(service, hostnameConfigurations)
	publicIPAddressID := stringPtr(mapStringValue(properties, "publicIPAddressesId", "publicIpAddressId", "public_ip_address_id"))
	publicIPAddresses := dedupeStrings(apiMgmtStringList(properties, "publicIPAddresses", "public_ip_addresses"))
	privateIPAddresses := dedupeStrings(apiMgmtStringList(properties, "privateIPAddresses", "private_ip_addresses"))
	workloadIdentityIDs := sortedKeys(mapValue(identity, "userAssignedIdentities", "user_assigned_identities"))

	return models.ApiMgmtServiceAsset{
		ID:                           firstNonEmpty(mapStringValue(service, "id"), "/unknown/"+firstNonEmpty(mapStringValue(service, "name"), "unknown")),
		Name:                         firstNonEmpty(mapStringValue(service, "name"), "unknown"),
		ResourceGroup:                resourceGroupFromID(mapStringValue(service, "id")),
		Location:                     stringPtr(mapStringValue(service, "location")),
		State:                        stringPtr(mapStringValue(properties, "provisioningState", "provisioning_state")),
		SKUName:                      stringPtr(mapStringValue(sku, "name")),
		SKUCapacity:                  apiMgmtIntPtr(sku, "capacity"),
		PublicNetworkAccess:          stringPtr(mapStringValue(properties, "publicNetworkAccess", "public_network_access")),
		VirtualNetworkType:           stringPtr(mapStringValue(properties, "virtualNetworkType", "virtual_network_type")),
		PublicIPAddressID:            publicIPAddressID,
		PublicIPAddresses:            publicIPAddresses,
		PrivateIPAddresses:           privateIPAddresses,
		GatewayHostnames:             gatewayHostnames,
		ManagementHostnames:          managementHostnames,
		PortalHostnames:              portalHostnames,
		WorkloadIdentityType:         stringPtr(mapStringValue(identity, "type")),
		WorkloadPrincipalID:          stringPtr(mapStringValue(identity, "principalId", "principal_id")),
		WorkloadClientID:             stringPtr(mapStringValue(identity, "clientId", "client_id")),
		WorkloadIdentityIDs:          workloadIdentityIDs,
		GatewayEnabled:               apiMgmtGatewayEnabled(properties),
		DeveloperPortalStatus:        stringPtr(mapStringValue(properties, "developerPortalStatus", "developer_portal_status")),
		LegacyPortalStatus:           stringPtr(mapStringValue(properties, "legacyPortalStatus", "legacy_portal_status")),
		APICount:                     apiMgmtCountPtr(apis),
		APISubscriptionRequiredCount: apiMgmtSubscriptionRequiredCount(apis),
		SubscriptionCount:            apiMgmtCountPtr(subscriptions),
		ActiveSubscriptionCount:      apiMgmtActiveSubscriptionCount(subscriptions),
		BackendCount:                 apiMgmtCountPtr(backends),
		BackendHostnames:             apiMgmtBackendHostnames(backends),
		PolicyCount:                  apiMgmtCountPtr(policies),
		PolicyControlTypes:           apiMgmtPolicyControlTypes(policies),
		NamedValueCount:              apiMgmtCountPtr(namedValues),
		NamedValueSecretCount:        apiMgmtNamedValueSecretCount(namedValues),
		NamedValueKeyVaultCount:      apiMgmtNamedValueKeyVaultCount(namedValues),
		Summary: apiMgmtOperatorSummary(
			firstNonEmpty(mapStringValue(service, "name"), "unknown"),
			gatewayHostnames,
			managementHostnames,
			portalHostnames,
			stringPtr(mapStringValue(properties, "publicNetworkAccess", "public_network_access")),
			stringPtr(mapStringValue(properties, "virtualNetworkType", "virtual_network_type")),
			stringPtr(mapStringValue(sku, "name")),
			stringPtr(mapStringValue(identity, "type")),
			apiMgmtCountPtr(apis),
			apiMgmtSubscriptionRequiredCount(apis),
			apiMgmtCountPtr(subscriptions),
			apiMgmtActiveSubscriptionCount(subscriptions),
			apiMgmtCountPtr(backends),
			apiMgmtBackendHostnames(backends),
			apiMgmtCountPtr(policies),
			apiMgmtPolicyControlTypes(policies),
			apiMgmtCountPtr(namedValues),
			apiMgmtNamedValueSecretCount(namedValues),
			apiMgmtNamedValueKeyVaultCount(namedValues),
			apiMgmtGatewayEnabled(properties),
			stringPtr(mapStringValue(properties, "developerPortalStatus", "developer_portal_status")),
		),
		RelatedIDs: dedupeStrings(append([]string{
			mapStringValue(service, "id"),
			mapStringValue(identity, "principalId", "principal_id"),
			stringPtrValue(publicIPAddressID),
		}, workloadIdentityIDs...)),
	}
}

func apiMgmtHostnames(service map[string]any, hostnameConfigurations []any) ([]string, []string, []string) {
	properties := mapValue(service, "properties")
	gateway := dedupeStrings(append(
		apiMgmtHostnameConfigs(hostnameConfigurations, "proxy"),
		hostnameFromURL(mapStringValue(properties, "gatewayUrl", "gateway_url")),
	))
	management := dedupeStrings(append(
		apiMgmtHostnameConfigs(hostnameConfigurations, "management"),
		hostnameFromURL(mapStringValue(properties, "managementApiUrl", "management_api_url")),
	))
	portal := dedupeStrings(append(
		append(
			apiMgmtHostnameConfigs(hostnameConfigurations, "portal", "developerportal"),
			hostnameFromURL(mapStringValue(properties, "portalUrl", "portal_url")),
		),
		hostnameFromURL(mapStringValue(properties, "developerPortalUrl", "developer_portal_url")),
	))
	return gateway, management, portal
}

func apiMgmtHostnameConfigs(configs []any, types ...string) []string {
	typeSet := map[string]struct{}{}
	for _, value := range types {
		typeSet[strings.ToLower(value)] = struct{}{}
	}
	values := []string{}
	for _, raw := range configs {
		config, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if _, exists := typeSet[strings.ToLower(mapStringValue(config, "type"))]; !exists {
			continue
		}
		values = append(values, mapStringValue(config, "hostName", "host_name"))
	}
	return dedupeStrings(values)
}

func hostnameFromURL(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err == nil && parsed.Hostname() != "" {
		return parsed.Hostname()
	}
	return value
}

func apiMgmtStringList(input map[string]any, keys ...string) []string {
	values := []string{}
	for _, item := range listValue(input, keys...) {
		text := stringValue(item)
		if strings.TrimSpace(text) != "" {
			values = append(values, text)
		}
	}
	return values
}

func apiMgmtCountPtr(items []map[string]any) *int {
	if items == nil {
		return nil
	}
	count := len(items)
	return &count
}

func apiMgmtSubscriptionRequiredCount(apis []map[string]any) *int {
	if apis == nil {
		return nil
	}
	count := 0
	for _, api := range apis {
		if mapBoolValue(mapValue(api, "properties"), "subscriptionRequired", "subscription_required") {
			count++
		}
	}
	return &count
}

func apiMgmtActiveSubscriptionCount(subscriptions []map[string]any) *int {
	if subscriptions == nil {
		return nil
	}
	count := 0
	for _, subscription := range subscriptions {
		if strings.EqualFold(strings.TrimSpace(mapStringValue(mapValue(subscription, "properties"), "state")), "active") {
			count++
		}
	}
	return &count
}

func apiMgmtBackendHostnames(backends []map[string]any) []string {
	if backends == nil {
		return []string{}
	}
	values := []string{}
	for _, backend := range backends {
		properties := mapValue(backend, "properties")
		values = append(values, hostnameFromURL(firstNonEmpty(
			mapStringValue(backend, "url"),
			mapStringValue(properties, "url"),
		)))
	}
	return dedupeStrings(values)
}

func apiMgmtPolicyControlTypes(policies []map[string]any) []string {
	if policies == nil {
		return []string{}
	}
	values := []string{}
	for _, policy := range policies {
		text := strings.ToLower(firstNonEmpty(
			mapStringValue(policy, "value"),
			mapStringValue(mapValue(policy, "properties"), "value", "policyContent", "policy_content"),
		))
		switch {
		case strings.Contains(text, "<set-backend-service") || strings.Contains(text, "<forward-request"):
			values = append(values, "backend-routing")
		}
		if strings.Contains(text, "<rewrite-uri") || strings.Contains(text, "<set-query-parameter") {
			values = append(values, "request-rewrite")
		}
		if strings.Contains(text, "<set-header") || strings.Contains(text, "<authentication-") || strings.Contains(text, "<validate-jwt") {
			values = append(values, "header-auth")
		}
		if strings.Contains(text, "<choose") || strings.Contains(text, "<when ") {
			values = append(values, "conditional-routing")
		}
		if strings.Contains(text, "<send-request") || strings.Contains(text, "<send-one-way-request") {
			values = append(values, "side-request")
		}
	}
	return dedupeStrings(values)
}

func apiMgmtNamedValueSecretCount(namedValues []map[string]any) *int {
	if namedValues == nil {
		return nil
	}
	count := 0
	for _, namedValue := range namedValues {
		if mapBoolValue(mapValue(namedValue, "properties"), "secret") {
			count++
		}
	}
	return &count
}

func apiMgmtNamedValueKeyVaultCount(namedValues []map[string]any) *int {
	if namedValues == nil {
		return nil
	}
	count := 0
	for _, namedValue := range namedValues {
		if mapStringValue(mapValue(mapValue(namedValue, "properties"), "keyVault", "key_vault"), "secretIdentifier", "secret_identifier") != "" {
			count++
		}
	}
	return &count
}

func apiMgmtGatewayEnabled(properties map[string]any) *bool {
	if value := optionalBoolPtr(properties, "disableGateway", "disable_gateway"); value != nil {
		enabled := !*value
		return &enabled
	}
	return nil
}

func apiMgmtIntPtr(input map[string]any, keys ...string) *int {
	for _, key := range keys {
		if value, exists := input[key]; exists {
			switch typed := value.(type) {
			case int:
				return &typed
			case int32:
				converted := int(typed)
				return &converted
			case int64:
				converted := int(typed)
				return &converted
			case float64:
				converted := int(typed)
				return &converted
			}
		}
	}
	return nil
}

func apiMgmtOperatorSummary(
	serviceName string,
	gatewayHostnames []string,
	managementHostnames []string,
	portalHostnames []string,
	publicNetworkAccess *string,
	virtualNetworkType *string,
	skuName *string,
	workloadIdentityType *string,
	apiCount *int,
	apiSubscriptionRequiredCount *int,
	subscriptionCount *int,
	activeSubscriptionCount *int,
	backendCount *int,
	backendHostnames []string,
	policyCount *int,
	policyControlTypes []string,
	namedValueCount *int,
	namedValueSecretCount *int,
	namedValueKeyVaultCount *int,
	gatewayEnabled *bool,
	developerPortalStatus *string,
) string {
	hostParts := []string{}
	if len(gatewayHostnames) > 0 {
		hostParts = append(hostParts, "gateway hostnames "+strings.Join(gatewayHostnames, ", "))
	}
	if len(managementHostnames) > 0 {
		hostParts = append(hostParts, "management hostnames "+strings.Join(managementHostnames, ", "))
	}
	if len(portalHostnames) > 0 {
		hostParts = append(hostParts, "portal hostnames "+strings.Join(portalHostnames, ", "))
	}

	hostPhrase := "does not expose readable gateway or portal hostnames from the current read path"
	if len(hostParts) > 0 {
		hostPhrase = "publishes " + strings.Join(hostParts, "; ")
	}

	inventoryParts := []string{}
	if apiCount != nil {
		inventoryParts = append(inventoryParts, intText(*apiCount)+" APIs")
	}
	if apiSubscriptionRequiredCount != nil {
		if apiCount != nil {
			inventoryParts = append(inventoryParts, intText(*apiSubscriptionRequiredCount)+" require subscriptions")
		} else {
			inventoryParts = append(inventoryParts, intText(*apiSubscriptionRequiredCount)+" APIs require subscriptions")
		}
	}
	if subscriptionCount != nil {
		if activeSubscriptionCount != nil {
			inventoryParts = append(inventoryParts, intText(*subscriptionCount)+" subscriptions ("+intText(*activeSubscriptionCount)+" active)")
		} else {
			inventoryParts = append(inventoryParts, intText(*subscriptionCount)+" subscriptions")
		}
	}
	if backendCount != nil {
		inventoryParts = append(inventoryParts, intText(*backendCount)+" backends")
	}
	if policyCount != nil {
		inventoryParts = append(inventoryParts, intText(*policyCount)+" policy scope(s)")
	}
	if namedValueCount != nil {
		inventoryParts = append(inventoryParts, intText(*namedValueCount)+" named values")
	}

	inventoryPhrase := "Inventory counts are not fully readable from the current read path."
	if len(inventoryParts) > 0 {
		inventoryPhrase = "Visible inventory: " + strings.Join(inventoryParts, ", ") + "."
	}

	depthParts := []string{}
	if namedValueSecretCount != nil {
		depthParts = append(depthParts, intText(*namedValueSecretCount)+" named values marked secret")
	}
	if namedValueKeyVaultCount != nil {
		depthParts = append(depthParts, intText(*namedValueKeyVaultCount)+" Key Vault-backed named values")
	}
	if len(backendHostnames) > 0 {
		depthParts = append(depthParts, "backend hosts "+strings.Join(backendHostnames, ", "))
	}
	if len(policyControlTypes) > 0 {
		depthParts = append(depthParts, "policy controls "+strings.Join(policyControlTypes, ", "))
	}

	depthPhrase := ""
	if len(depthParts) > 0 {
		depthPhrase = " Depth cues: " + strings.Join(depthParts, ", ") + "."
	}

	postureParts := []string{
		"public network access " + firstNonEmpty(stringPtrValue(publicNetworkAccess), "unknown"),
		"virtual network type " + firstNonEmpty(stringPtrValue(virtualNetworkType), "none"),
	}
	if stringPtrValue(skuName) != "" {
		postureParts = append(postureParts, "SKU "+stringPtrValue(skuName))
	}
	if gatewayEnabled != nil {
		if *gatewayEnabled {
			postureParts = append(postureParts, "gateway enabled")
		} else {
			postureParts = append(postureParts, "gateway disabled")
		}
	}
	if stringPtrValue(developerPortalStatus) != "" {
		postureParts = append(postureParts, "developer portal "+stringPtrValue(developerPortalStatus))
	}

	identityPhrase := "has no managed identity visible from the current read path"
	if stringPtrValue(workloadIdentityType) != "" {
		identityPhrase = "uses managed identity (" + stringPtrValue(workloadIdentityType) + ")"
	}

	return "API Management service '" + serviceName + "' " + hostPhrase + " and " + identityPhrase + ". " +
		inventoryPhrase + depthPhrase + " Visible posture: " + strings.Join(postureParts, ", ") + "."
}

func intText(value int) string {
	return stringValue(value)
}
