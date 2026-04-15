package providers

import (
	"context"
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

const armManagedClustersAPIVersion = "2024-10-01"

func (provider AzureProvider) AKS(ctx context.Context, tenant string, subscription string) (AksFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return AksFacts{}, err
	}

	clusterMaps, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.ContainerService/managedClusters",
		armManagedClustersAPIVersion,
	)

	clusters := []models.AksClusterAsset{}
	issues := []models.Issue{}
	if err != nil {
		issues = append(issues, issueFromError("aks.managed_clusters", err))
		return AksFacts{
			TenantID:       session.tenantID,
			SubscriptionID: session.subscription.ID,
			AksClusters:    clusters,
			Issues:         issues,
		}, nil
	}

	for _, clusterMap := range clusterMaps {
		hydrated := clusterMap
		clusterID := mapStringValue(clusterMap, "id")
		clusterName := firstNonEmpty(mapStringValue(clusterMap, "name"), resourceNameFromID(clusterID), "unknown")
		resourceGroup := resourceGroupFromID(clusterID)

		if clusterID != "" && aksClusterNeedsHydration(clusterMap) {
			fullCluster, err := armGetObject(ctx, session.credential, clusterID, armManagedClustersAPIVersion)
			if err != nil {
				issues = append(issues, issueFromError(aksClusterIssueScope(resourceGroup, clusterName), err))
			} else {
				hydrated = fullCluster
			}
		}

		clusters = append(clusters, aksClusterSummary(hydrated))
	}

	return AksFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		AksClusters:    clusters,
		Issues:         issues,
	}, nil
}

func aksClusterIssueScope(resourceGroup string, clusterName string) string {
	if strings.TrimSpace(resourceGroup) != "" && strings.TrimSpace(clusterName) != "" {
		return fmt.Sprintf("aks[%s/%s].cluster", resourceGroup, clusterName)
	}
	return "aks.cluster"
}

func aksClusterNeedsHydration(cluster map[string]any) bool {
	properties := mapValue(cluster, "properties")
	apiServerAccessProfile := mapValue(properties, "apiServerAccessProfile", "api_server_access_profile")
	securityProfile := mapValue(properties, "securityProfile", "security_profile")
	workloadIdentity := mapValue(securityProfile, "workloadIdentity", "workload_identity")
	ingressProfile := mapValue(properties, "ingressProfile", "ingress_profile")
	webAppRouting := mapValue(ingressProfile, "webAppRouting", "web_app_routing")

	return optionalBoolPtr(apiServerAccessProfile, "enablePrivateCluster", "enable_private_cluster") == nil ||
		optionalBoolPtr(workloadIdentity, "enabled") == nil ||
		optionalBoolPtr(webAppRouting, "enabled") == nil
}

func aksClusterSummary(cluster map[string]any) models.AksClusterAsset {
	properties := mapValue(cluster, "properties")
	identity := mapValue(cluster, "identity")
	servicePrincipalProfile := mapValue(properties, "servicePrincipalProfile", "service_principal_profile")
	aadProfile := mapValue(properties, "aadProfile", "aad_profile")
	apiServerAccessProfile := mapValue(properties, "apiServerAccessProfile", "api_server_access_profile")
	networkProfile := mapValue(properties, "networkProfile", "network_profile")
	oidcIssuerProfile := mapValue(properties, "oidcIssuerProfile", "oidc_issuer_profile")
	securityProfile := mapValue(properties, "securityProfile", "security_profile")
	workloadIdentity := mapValue(securityProfile, "workloadIdentity", "workload_identity")
	ingressProfile := mapValue(properties, "ingressProfile", "ingress_profile")
	webAppRouting := mapValue(ingressProfile, "webAppRouting", "web_app_routing")
	sku := mapValue(cluster, "sku")

	clusterID := mapStringValue(cluster, "id")
	clusterName := firstNonEmpty(mapStringValue(cluster, "name"), resourceNameFromID(clusterID), "unknown")
	clusterIdentityIDs := sortedKeys(mapValue(identity, "userAssignedIdentities", "user_assigned_identities"))
	clusterIdentityType := stringPtr(mapStringValue(identity, "type"))
	clusterPrincipalID := stringPtr(mapStringValue(identity, "principalId", "principal_id"))
	clusterClientID := stringPtr(mapStringValue(identity, "clientId", "client_id"))

	if clusterIdentityType == nil {
		if servicePrincipalClientID := mapStringValue(servicePrincipalProfile, "clientId", "client_id"); servicePrincipalClientID != "" {
			clusterIdentityType = stringPtr("ServicePrincipal")
			clusterClientID = stringPtr(servicePrincipalClientID)
		}
	}

	privateClusterEnabled := optionalBoolPtr(apiServerAccessProfile, "enablePrivateCluster", "enable_private_cluster")
	publicFQDNEnabled := optionalBoolPtr(apiServerAccessProfile, "enablePrivateClusterPublicFQDN", "enable_private_cluster_public_fqdn")
	aadManaged := optionalBoolPtr(aadProfile, "managed")
	azureRBACEnabled := optionalBoolPtr(aadProfile, "enableAzureRBAC", "enable_azure_rbac")
	localAccountsDisabled := optionalBoolPtr(properties, "disableLocalAccounts", "disable_local_accounts")
	oidcIssuerEnabled := optionalBoolPtr(oidcIssuerProfile, "enabled")
	workloadIdentityEnabled := optionalBoolPtr(workloadIdentity, "enabled")
	webAppRoutingEnabled := optionalBoolPtr(webAppRouting, "enabled")

	addonProfiles := mapValue(properties, "addonProfiles", "addon_profiles")
	addonNames := make([]string, 0, len(addonProfiles))
	for _, addonName := range sortedKeys(addonProfiles) {
		if mapBoolValue(addonProfiles[addonName], "enabled") {
			addonNames = append(addonNames, addonName)
		}
	}

	agentPoolCount := aksCollectionCount(properties, "agentPoolProfiles", "agent_pool_profiles")
	webAppRoutingDNSZoneCount := aksCollectionCount(webAppRouting, "dnsZoneResourceIds", "dns_zone_resource_ids")

	return models.AksClusterAsset{
		ID:                        firstNonEmpty(clusterID, "/unknown/"+clusterName),
		Name:                      clusterName,
		ResourceGroup:             resourceGroupFromID(clusterID),
		Location:                  stringPtr(mapStringValue(cluster, "location")),
		ProvisioningState:         stringPtr(mapStringValue(properties, "provisioningState", "provisioning_state")),
		KubernetesVersion:         stringPtr(mapStringValue(properties, "kubernetesVersion", "kubernetes_version")),
		SKUTier:                   stringPtr(mapStringValue(sku, "tier")),
		NodeResourceGroup:         stringPtr(mapStringValue(properties, "nodeResourceGroup", "node_resource_group")),
		FQDN:                      stringPtr(mapStringValue(properties, "fqdn")),
		PrivateFQDN:               stringPtr(mapStringValue(properties, "privateFqdn", "private_fqdn")),
		PrivateClusterEnabled:     privateClusterEnabled,
		PublicFQDNEnabled:         publicFQDNEnabled,
		ClusterIdentityType:       clusterIdentityType,
		ClusterPrincipalID:        clusterPrincipalID,
		ClusterClientID:           clusterClientID,
		ClusterIdentityIDs:        clusterIdentityIDs,
		AADManaged:                aadManaged,
		AzureRBACEnabled:          azureRBACEnabled,
		LocalAccountsDisabled:     localAccountsDisabled,
		NetworkPlugin:             stringPtr(mapStringValue(networkProfile, "networkPlugin", "network_plugin")),
		NetworkPolicy:             stringPtr(mapStringValue(networkProfile, "networkPolicy", "network_policy")),
		OutboundType:              stringPtr(mapStringValue(networkProfile, "outboundType", "outbound_type")),
		AgentPoolCount:            agentPoolCount,
		OIDCIssuerEnabled:         oidcIssuerEnabled,
		OIDCIssuerURL:             stringPtr(mapStringValue(oidcIssuerProfile, "issuerURL", "issuerUrl", "issuer_url")),
		WorkloadIdentityEnabled:   workloadIdentityEnabled,
		AddonNames:                addonNames,
		WebAppRoutingEnabled:      webAppRoutingEnabled,
		WebAppRoutingDNSZoneCount: webAppRoutingDNSZoneCount,
		Summary: aksOperatorSummary(
			clusterName,
			stringPtr(mapStringValue(properties, "kubernetesVersion", "kubernetes_version")),
			stringPtr(mapStringValue(properties, "fqdn")),
			stringPtr(mapStringValue(properties, "privateFqdn", "private_fqdn")),
			privateClusterEnabled,
			publicFQDNEnabled,
			clusterIdentityType,
			clusterClientID,
			aadManaged,
			azureRBACEnabled,
			localAccountsDisabled,
			stringPtr(mapStringValue(networkProfile, "networkPlugin", "network_plugin")),
			stringPtr(mapStringValue(networkProfile, "networkPolicy", "network_policy")),
			stringPtr(mapStringValue(networkProfile, "outboundType", "outbound_type")),
			agentPoolCount,
			oidcIssuerEnabled,
			stringPtr(mapStringValue(oidcIssuerProfile, "issuerURL", "issuerUrl", "issuer_url")),
			workloadIdentityEnabled,
			addonNames,
			webAppRoutingEnabled,
			webAppRoutingDNSZoneCount,
		),
		RelatedIDs: dedupeStrings(append([]string{clusterID, stringPtrValue(clusterPrincipalID)}, clusterIdentityIDs...)),
	}
}

func aksCollectionCount(input map[string]any, keys ...string) *int {
	for _, key := range keys {
		raw, ok := input[key]
		if !ok {
			continue
		}
		if raw == nil {
			return nil
		}
		count := len(listValue(raw))
		return &count
	}
	return nil
}

func aksOperatorSummary(
	clusterName string,
	kubernetesVersion *string,
	fqdn *string,
	privateFQDN *string,
	privateClusterEnabled *bool,
	publicFQDNEnabled *bool,
	clusterIdentityType *string,
	clusterClientID *string,
	aadManaged *bool,
	azureRBACEnabled *bool,
	localAccountsDisabled *bool,
	networkPlugin *string,
	networkPolicy *string,
	outboundType *string,
	agentPoolCount *int,
	oidcIssuerEnabled *bool,
	oidcIssuerURL *string,
	workloadIdentityEnabled *bool,
	addonNames []string,
	webAppRoutingEnabled *bool,
	webAppRoutingDNSZoneCount *int,
) string {
	endpointPhrase := "does not expose a readable API endpoint from the current read path"
	if boolPtrIsTrue(privateClusterEnabled) && stringPtrValue(privateFQDN) != "" && boolPtrIsTrue(publicFQDNEnabled) && stringPtrValue(fqdn) != "" {
		endpointPhrase = "uses private API endpoint '" + stringPtrValue(privateFQDN) + "' and keeps public FQDN '" + stringPtrValue(fqdn) + "' enabled"
	} else if boolPtrIsTrue(privateClusterEnabled) && stringPtrValue(privateFQDN) != "" {
		endpointPhrase = "uses private API endpoint '" + stringPtrValue(privateFQDN) + "'"
	} else if stringPtrValue(fqdn) != "" {
		endpointPhrase = "publishes API endpoint '" + stringPtrValue(fqdn) + "'"
	}

	identityPhrase := "has no cluster identity context visible from the current read path"
	switch stringPtrValue(clusterIdentityType) {
	case "ServicePrincipal":
		if stringPtrValue(clusterClientID) != "" {
			identityPhrase = "uses service principal client '" + stringPtrValue(clusterClientID) + "'"
		} else {
			identityPhrase = "uses service principal-backed cluster credentials"
		}
	case "":
	default:
		identityPhrase = "uses cluster identity (" + stringPtrValue(clusterIdentityType) + ")"
	}

	authParts := []string{}
	switch {
	case boolPtrIsTrue(aadManaged):
		authParts = append(authParts, "AAD-managed auth")
	case boolPtrIsFalse(aadManaged):
		authParts = append(authParts, "AAD profile not managed")
	}
	switch {
	case boolPtrIsTrue(azureRBACEnabled):
		authParts = append(authParts, "Azure RBAC enabled")
	case boolPtrIsFalse(azureRBACEnabled):
		authParts = append(authParts, "Azure RBAC disabled")
	}
	switch {
	case boolPtrIsTrue(localAccountsDisabled):
		authParts = append(authParts, "local accounts disabled")
	case boolPtrIsFalse(localAccountsDisabled):
		authParts = append(authParts, "local accounts enabled")
	}
	switch {
	case boolPtrIsTrue(oidcIssuerEnabled):
		authParts = append(authParts, "OIDC issuer enabled")
	case boolPtrIsFalse(oidcIssuerEnabled):
		authParts = append(authParts, "OIDC issuer disabled")
	}
	switch {
	case boolPtrIsTrue(workloadIdentityEnabled):
		authParts = append(authParts, "workload identity enabled")
	case boolPtrIsFalse(workloadIdentityEnabled):
		authParts = append(authParts, "workload identity disabled")
	}

	networkParts := []string{}
	switch {
	case boolPtrIsTrue(privateClusterEnabled):
		networkParts = append(networkParts, "private cluster enabled")
	case boolPtrIsFalse(privateClusterEnabled):
		networkParts = append(networkParts, "private cluster disabled")
	}
	if stringPtrValue(networkPlugin) != "" {
		networkParts = append(networkParts, "network plugin "+stringPtrValue(networkPlugin))
	}
	if stringPtrValue(networkPolicy) != "" {
		networkParts = append(networkParts, "network policy "+stringPtrValue(networkPolicy))
	}
	if stringPtrValue(outboundType) != "" {
		networkParts = append(networkParts, "outbound "+stringPtrValue(outboundType))
	}
	switch {
	case boolPtrIsTrue(webAppRoutingEnabled):
		if webAppRoutingDNSZoneCount != nil {
			networkParts = append(networkParts, fmt.Sprintf("web app routing enabled (%d DNS zone links)", *webAppRoutingDNSZoneCount))
		} else {
			networkParts = append(networkParts, "web app routing enabled")
		}
	case boolPtrIsFalse(webAppRoutingEnabled):
		networkParts = append(networkParts, "web app routing disabled")
	}

	inventoryParts := []string{}
	if stringPtrValue(kubernetesVersion) != "" {
		inventoryParts = append(inventoryParts, "Kubernetes "+stringPtrValue(kubernetesVersion))
	}
	if agentPoolCount != nil {
		inventoryParts = append(inventoryParts, fmt.Sprintf("%d agent pool(s)", *agentPoolCount))
	}
	if len(addonNames) > 0 {
		inventoryParts = append(inventoryParts, "addons "+strings.Join(addonNames, ", "))
	}

	depthParts := []string{}
	if boolPtrIsTrue(oidcIssuerEnabled) && stringPtrValue(oidcIssuerURL) != "" {
		depthParts = append(depthParts, "OIDC issuer "+stringPtrValue(oidcIssuerURL))
	} else if boolPtrIsTrue(oidcIssuerEnabled) {
		depthParts = append(depthParts, "OIDC issuer enabled")
	}
	if boolPtrIsTrue(workloadIdentityEnabled) {
		depthParts = append(depthParts, "workload identity enabled")
	}
	if len(addonNames) > 0 {
		depthParts = append(depthParts, "enabled addons "+strings.Join(addonNames, ", "))
	}

	depthPhrase := ""
	if len(depthParts) > 0 {
		depthPhrase = " Depth cues: " + strings.Join(depthParts, ", ") + "."
	}

	authPhrase := "Auth posture is not fully readable from the current read path."
	if len(authParts) > 0 {
		authPhrase = "Visible auth posture: " + strings.Join(authParts, ", ") + "."
	}

	networkPhrase := "Network shape is not fully readable from the current read path."
	if len(networkParts) > 0 {
		networkPhrase = "Visible network shape: " + strings.Join(networkParts, ", ") + "."
	}

	inventoryPhrase := "Cluster version and pool counts are not fully readable from the current read path."
	if len(inventoryParts) > 0 {
		inventoryPhrase = "Visible inventory: " + strings.Join(inventoryParts, ", ") + "."
	}

	return "AKS cluster '" + clusterName + "' " + endpointPhrase + " and " + identityPhrase + ". " +
		inventoryPhrase + depthPhrase + " " + authPhrase + " " + networkPhrase
}

func boolPtrIsTrue(value *bool) bool {
	return value != nil && *value
}

func boolPtrIsFalse(value *bool) bool {
	return value != nil && !*value
}
