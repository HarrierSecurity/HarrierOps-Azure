package providers

import (
	"context"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

const armApplicationGatewayAPIVersion = "2024-05-01"

func (provider AzureProvider) ApplicationGateway(ctx context.Context, tenant string, subscription string) (ApplicationGatewayFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return ApplicationGatewayFacts{}, err
	}

	publicIPLookup := map[string]string{}
	issues := []models.Issue{}

	publicIPs, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.Network/publicIPAddresses",
		armApplicationGatewayAPIVersion,
	)
	if err != nil {
		issues = append(issues, issueFromError("application_gateway.public_ip_addresses", err))
	} else {
		for _, publicIP := range publicIPs {
			publicIPID := strings.ToLower(strings.TrimSpace(mapStringValue(publicIP, "id")))
			publicIPAddress := mapStringValue(mapValue(publicIP, "properties"), "ipAddress", "ip_address")
			if publicIPID != "" && publicIPAddress != "" {
				publicIPLookup[publicIPID] = publicIPAddress
			}
		}
	}

	gateways, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.Network/applicationGateways",
		armApplicationGatewayAPIVersion,
	)
	if err != nil {
		issues = append(issues, issueFromError("application_gateway.gateways", err))
		return ApplicationGatewayFacts{
			TenantID:            session.tenantID,
			SubscriptionID:      session.subscription.ID,
			ApplicationGateways: []models.ApplicationGatewayAsset{},
			Issues:              issues,
		}, nil
	}

	rows := make([]models.ApplicationGatewayAsset, 0, len(gateways))
	for _, gateway := range gateways {
		rows = append(rows, applicationGatewaySummary(gateway, publicIPLookup))
	}

	return ApplicationGatewayFacts{
		TenantID:            session.tenantID,
		SubscriptionID:      session.subscription.ID,
		ApplicationGateways: rows,
		Issues:              issues,
	}, nil
}

func applicationGatewaySummary(gateway map[string]any, publicIPLookup map[string]string) models.ApplicationGatewayAsset {
	properties := mapValue(gateway, "properties")
	sku := mapValue(gateway, "sku")
	firewallPolicyID := stringPtr(mapStringValue(mapValue(properties, "firewallPolicy", "firewall_policy"), "id"))

	publicIPAddressIDs := []string{}
	publicIPAddresses := []string{}
	privateFrontendIPs := []string{}
	subnetIDs := []string{}

	for _, raw := range listValue(properties, "frontendIPConfigurations", "frontend_ip_configurations") {
		frontend := mapValue(raw)
		frontendProperties := mapValue(frontend, "properties")
		publicIPID := mapStringValue(mapValue(frontendProperties, "publicIPAddress", "public_ip_address"), "id")
		if publicIPID != "" {
			publicIPAddressIDs = append(publicIPAddressIDs, publicIPID)
			if ip, ok := publicIPLookup[strings.ToLower(strings.TrimSpace(publicIPID))]; ok && ip != "" {
				publicIPAddresses = append(publicIPAddresses, ip)
			}
		}
		privateIP := mapStringValue(frontendProperties, "privateIPAddress", "private_ip_address")
		if privateIP != "" {
			privateFrontendIPs = append(privateFrontendIPs, privateIP)
		}
		subnetID := mapStringValue(mapValue(frontendProperties, "subnet"), "id")
		if subnetID != "" {
			subnetIDs = append(subnetIDs, subnetID)
		}
	}

	backendPools := listValue(properties, "backendAddressPools", "backend_address_pools")
	publicIPAddressIDs = dedupeStrings(publicIPAddressIDs)
	publicIPAddresses = dedupeStrings(publicIPAddresses)
	privateFrontendIPs = dedupeStrings(privateFrontendIPs)
	subnetIDs = dedupeStrings(subnetIDs)
	sort.Strings(publicIPAddresses)

	wafConfiguration := mapValue(properties, "webApplicationFirewallConfiguration", "web_application_firewall_configuration")
	wafEnabled := optionalBoolPtr(wafConfiguration, "enabled")
	wafMode := stringPtr(mapStringValue(wafConfiguration, "firewallMode", "firewall_mode"))

	return models.ApplicationGatewayAsset{
		BackendPoolCount:           len(backendPools),
		BackendTargetCount:         applicationGatewayBackendTargetCount(backendPools),
		FirewallPolicyID:           firewallPolicyID,
		ID:                         firstNonEmpty(mapStringValue(gateway, "id"), "/unknown/"+firstNonEmpty(mapStringValue(gateway, "name"), "unknown")),
		ListenerCount:              len(listValue(properties, "httpListeners", "http_listeners")),
		Location:                   stringPtr(mapStringValue(gateway, "location")),
		Name:                       firstNonEmpty(mapStringValue(gateway, "name"), "unknown"),
		PrivateFrontendCount:       len(privateFrontendIPs),
		PrivateFrontendIPs:         privateFrontendIPs,
		PublicFrontendCount:        len(publicIPAddressIDs),
		PublicIPAddressIDs:         publicIPAddressIDs,
		PublicIPAddresses:          publicIPAddresses,
		RedirectConfigurationCount: len(listValue(properties, "redirectConfigurations", "redirect_configurations")),
		RelatedIDs:                 dedupeStrings(append([]string{mapStringValue(gateway, "id")}, append(append(publicIPAddressIDs, subnetIDs...), stringPtrValue(firewallPolicyID))...)),
		RequestRoutingRuleCount:    len(listValue(properties, "requestRoutingRules", "request_routing_rules")),
		ResourceGroup:              resourceGroupFromID(mapStringValue(gateway, "id")),
		RewriteRuleSetCount:        len(listValue(properties, "rewriteRuleSets", "rewrite_rule_sets")),
		SKUName:                    stringPtr(mapStringValue(sku, "name")),
		SKUTier:                    stringPtr(mapStringValue(sku, "tier")),
		State:                      stringPtr(firstNonEmpty(mapStringValue(gateway, "operationalState", "operational_state"), mapStringValue(properties, "operationalState", "operational_state"))),
		SubnetIDs:                  subnetIDs,
		Summary: applicationGatewayOperatorSummary(
			firstNonEmpty(mapStringValue(gateway, "name"), "unknown"),
			len(publicIPAddressIDs),
			len(privateFrontendIPs),
			publicIPAddresses,
			len(listValue(properties, "httpListeners", "http_listeners")),
			len(listValue(properties, "requestRoutingRules", "request_routing_rules")),
			len(backendPools),
			applicationGatewayBackendTargetCount(backendPools),
			wafEnabled,
			wafMode,
			firewallPolicyID,
		),
		URLPathMapCount: len(listValue(properties, "urlPathMaps", "url_path_maps")),
		WAFEnabled:      wafEnabled,
		WAFMode:         wafMode,
	}
}

func applicationGatewayBackendTargetCount(backendPools []any) int {
	targets := []string{}
	for _, raw := range backendPools {
		pool := mapValue(raw)
		properties := mapValue(pool, "properties")
		for _, rawAddress := range listValue(properties, "backendAddresses", "backend_addresses") {
			address := mapValue(rawAddress)
			if fqdn := mapStringValue(address, "fqdn"); fqdn != "" {
				targets = append(targets, "fqdn:"+fqdn)
			} else if ip := mapStringValue(address, "ipAddress", "ip_address"); ip != "" {
				targets = append(targets, "ip:"+ip)
			}
		}
		for _, rawConfig := range listValue(properties, "backendIPConfigurations", "backend_ip_configurations") {
			id := mapStringValue(mapValue(rawConfig), "id")
			if id != "" {
				targets = append(targets, "id:"+id)
			}
		}
	}
	return len(dedupeStrings(targets))
}

func applicationGatewayOperatorSummary(
	gatewayName string,
	publicFrontendCount int,
	privateFrontendCount int,
	publicIPAddresses []string,
	listenerCount int,
	requestRoutingRuleCount int,
	backendPoolCount int,
	backendTargetCount int,
	wafEnabled *bool,
	wafMode *string,
	firewallPolicyID *string,
) string {
	exposurePhrase := "does not expose readable frontend IP posture from the current read path"
	if publicFrontendCount > 0 {
		exposurePhrase = "publishes " + stringValue(publicFrontendCount) + " public frontend(s)"
		if len(publicIPAddresses) > 0 {
			exposurePhrase += " (" + strings.Join(publicIPAddresses, ", ") + ")"
		}
	} else if privateFrontendCount > 0 {
		exposurePhrase = "is private-only from the current read path (" + stringValue(privateFrontendCount) + " private frontend(s))"
	}

	routingParts := []string{}
	if listenerCount > 0 {
		routingParts = append(routingParts, stringValue(listenerCount)+" listener(s)")
	}
	if requestRoutingRuleCount > 0 {
		routingParts = append(routingParts, stringValue(requestRoutingRuleCount)+" routing rule(s)")
	}
	if backendPoolCount > 0 {
		routingParts = append(routingParts, stringValue(backendPoolCount)+" backend pool(s)")
	}
	if backendTargetCount > 0 {
		routingParts = append(routingParts, stringValue(backendTargetCount)+" backend target(s)")
	}
	routingPhrase := "Routing breadth is not fully readable from the current read path."
	if len(routingParts) > 0 {
		routingPhrase = "Visible routing breadth: " + strings.Join(routingParts, ", ") + "."
	}

	wafPhrase := "No visible WAF protection is configured from the current read path."
	if firewallPolicyID != nil && *firewallPolicyID != "" && wafMode != nil && *wafMode != "" {
		wafPhrase = "WAF policy is attached and running in " + *wafMode + " mode."
	} else if firewallPolicyID != nil && *firewallPolicyID != "" {
		wafPhrase = "WAF policy is attached."
	} else if wafEnabled != nil && *wafEnabled && wafMode != nil && *wafMode != "" {
		wafPhrase = "Gateway-level WAF is enabled in " + *wafMode + " mode."
	} else if wafEnabled != nil && *wafEnabled {
		wafPhrase = "Gateway-level WAF is enabled."
	} else if wafEnabled != nil && !*wafEnabled {
		wafPhrase = "Visible WAF protection is disabled."
	}

	whyPhrase := "This is still useful shared-ingress context, but it is not an obvious internet-first path."
	if publicFrontendCount > 0 && applicationGatewayHasSharedBreadth(listenerCount, requestRoutingRuleCount, backendPoolCount, backendTargetCount) {
		whyPhrase = "This is a shared front door, so if the edge is weak the apps behind it may deserve review next."
	} else if publicFrontendCount > 0 {
		whyPhrase = "Because this gateway is public, weak edge controls here would make the backend path worth checking next."
	}

	return "Application Gateway '" + gatewayName + "' " + exposurePhrase + ". " + routingPhrase + " " + wafPhrase + " " + whyPhrase
}

func applicationGatewayHasSharedBreadth(listenerCount int, requestRoutingRuleCount int, backendPoolCount int, backendTargetCount int) bool {
	return listenerCount > 1 || requestRoutingRuleCount > 1 || backendPoolCount > 1 || backendTargetCount > 1
}
