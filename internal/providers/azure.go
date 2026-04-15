package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"

	"harrierops-azure/internal/models"
)

const managementScope = "https://management.azure.com/.default"

type AzureProvider struct{}

func NewAzureProvider() AzureProvider {
	return AzureProvider{}
}

type azureSession struct {
	claims        map[string]string
	credential    azcore.TokenCredential
	tokenSource   string
	authMode      string
	tenantID      string
	subscription  models.SubscriptionRef
	clientFactory *armresources.ClientFactory
}

func (provider AzureProvider) WhoAmI(ctx context.Context, tenant string, subscription string) (WhoAmIFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return WhoAmIFacts{}, err
	}

	principalID := firstNonEmpty(
		session.claims["oid"],
		session.claims["appid"],
		session.claims["sub"],
	)
	displayName := firstNonEmpty(
		session.claims["name"],
		session.claims["preferred_username"],
		session.claims["upn"],
	)

	return WhoAmIFacts{
		TenantID:     session.tenantID,
		Subscription: session.subscription,
		Principal: models.Principal{
			DisplayName:   displayName,
			ID:            principalID,
			PrincipalType: principalTypeFromClaims(session.claims),
			TenantID:      session.tenantID,
		},
		EffectiveScopes: []models.ScopeRef{
			{
				DisplayName: session.subscription.DisplayName,
				ID:          "/subscriptions/" + session.subscription.ID,
				ScopeType:   "subscription",
			},
		},
		TokenSource: session.tokenSource,
		AuthMode:    session.authMode,
		Issues:      []models.Issue{},
	}, nil
}

func (provider AzureProvider) Inventory(ctx context.Context, tenant string, subscription string) (InventoryFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return InventoryFacts{}, err
	}

	resourceGroupsClient := session.clientFactory.NewResourceGroupsClient()
	resourceClient := session.clientFactory.NewClient()

	resourceGroupCount := 0
	resourceCount := 0
	topTypes := models.TopResourceTypes{}
	issues := []models.Issue{}

	resourceGroupPager := resourceGroupsClient.NewListPager(nil)
	for resourceGroupPager.More() {
		page, err := resourceGroupPager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("inventory.resource_groups", err))
			break
		}
		resourceGroupCount += len(page.Value)
	}

	resourcePager := resourceClient.NewListPager(nil)
	for resourcePager.More() {
		page, err := resourcePager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("inventory.resources", err))
			break
		}
		resourceCount += len(page.Value)
		for _, resource := range page.Value {
			resourceType := stringValue(resource.Type)
			if resourceType == "" {
				continue
			}
			topTypes[resourceType]++
		}
	}

	return InventoryFacts{
		TenantID:           session.tenantID,
		Subscription:       session.subscription,
		ResourceGroupCount: resourceGroupCount,
		ResourceCount:      resourceCount,
		TopResourceTypes:   topTypes,
		Issues:             issues,
	}, nil
}

func (provider AzureProvider) ArmDeployments(ctx context.Context, tenant string, subscription string) (ArmDeploymentsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return ArmDeploymentsFacts{}, err
	}

	deploymentsClient := session.clientFactory.NewDeploymentsClient()
	resourceGroupsClient := session.clientFactory.NewResourceGroupsClient()

	deployments := []models.ArmDeploymentSummary{}
	issues := []models.Issue{}

	subscriptionPager := deploymentsClient.NewListAtSubscriptionScopePager(nil)
	for subscriptionPager.More() {
		page, err := subscriptionPager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("arm_deployments.subscription", err))
			break
		}
		for _, deployment := range page.Value {
			deployments = append(deployments, deploymentSummary(
				deployment,
				"/subscriptions/"+session.subscription.ID,
				"subscription",
				nil,
			))
		}
	}

	resourceGroupPager := resourceGroupsClient.NewListPager(nil)
	for resourceGroupPager.More() {
		page, err := resourceGroupPager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("arm_deployments.resource_groups", err))
			break
		}
		for _, resourceGroup := range page.Value {
			resourceGroupName := stringValue(resourceGroup.Name)
			if resourceGroupName == "" {
				continue
			}
			groupPager := deploymentsClient.NewListByResourceGroupPager(resourceGroupName, nil)
			for groupPager.More() {
				groupPage, err := groupPager.NextPage(ctx)
				if err != nil {
					issues = append(issues, issueFromError("arm_deployments.resource_groups["+resourceGroupName+"]", err))
					break
				}
				for _, deployment := range groupPage.Value {
					deployments = append(deployments, deploymentSummary(
						deployment,
						"/subscriptions/"+session.subscription.ID+"/resourceGroups/"+resourceGroupName,
						"resource_group",
						models.StringPtr(resourceGroupName),
					))
				}
			}
		}
	}

	return ArmDeploymentsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		Deployments:    dedupeDeployments(deployments),
		Issues:         issues,
	}, nil
}

func (provider AzureProvider) AppServices(ctx context.Context, tenant string, subscription string) (AppServicesFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return AppServicesFacts{}, err
	}

	webAppsClient, err := armappservice.NewWebAppsClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return AppServicesFacts{}, fmt.Errorf("build web apps client: %w", err)
	}

	rows := []models.AppServiceAsset{}
	issues := []models.Issue{}
	pager := webAppsClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("app_services.web_apps", err))
			break
		}
		for _, app := range page.Value {
			appMap := map[string]any{}
			decodeJSONInto(app, &appMap)
			if webAssetKind(mapStringValue(appMap, "kind")) != "AppService" {
				continue
			}

			configMap := map[string]any{}
			resourceGroup := resourceGroupFromID(mapStringValue(appMap, "id"))
			appName := mapStringValue(appMap, "name")
			if resourceGroup != "" && appName != "" {
				config, err := webAppsClient.GetConfiguration(ctx, resourceGroup, appName, nil)
				if err != nil {
					issues = append(issues, issueFromError("app_services["+resourceGroup+"/"+appName+"].configuration", err))
				} else {
					decodeJSONInto(config.SiteConfigResource, &configMap)
				}
			}

			rows = append(rows, appServiceSummary(appMap, configMap))
		}
	}

	return AppServicesFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		AppServices:    rows,
		Issues:         issues,
	}, nil
}

func (provider AzureProvider) Functions(ctx context.Context, tenant string, subscription string) (FunctionsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return FunctionsFacts{}, err
	}

	webAppsClient, err := armappservice.NewWebAppsClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return FunctionsFacts{}, fmt.Errorf("build web apps client: %w", err)
	}

	rows := []models.FunctionAppAsset{}
	issues := []models.Issue{}
	pager := webAppsClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("functions.web_apps", err))
			break
		}
		for _, app := range page.Value {
			appMap := map[string]any{}
			decodeJSONInto(app, &appMap)
			if webAssetKind(mapStringValue(appMap, "kind")) != "FunctionApp" {
				continue
			}

			resourceGroup := resourceGroupFromID(mapStringValue(appMap, "id"))
			appName := mapStringValue(appMap, "name")
			configMap := map[string]any{}
			settingsMap := map[string]any{}
			if resourceGroup != "" && appName != "" {
				config, err := webAppsClient.GetConfiguration(ctx, resourceGroup, appName, nil)
				if err != nil {
					issues = append(issues, issueFromError("functions["+resourceGroup+"/"+appName+"].configuration", err))
				} else {
					decodeJSONInto(config.SiteConfigResource, &configMap)
				}
				settings, err := webAppsClient.ListApplicationSettings(ctx, resourceGroup, appName, nil)
				if err != nil {
					issues = append(issues, issueFromError("functions["+resourceGroup+"/"+appName+"].app_settings", err))
				} else {
					decodeJSONInto(settings.StringDictionary, &settingsMap)
				}
			}

			rows = append(rows, functionAppSummary(
				appMap,
				configMap,
				mapValue(settingsMap, "properties"),
			))
		}
	}

	return FunctionsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		FunctionApps:   rows,
		Issues:         issues,
	}, nil
}

func (provider AzureProvider) ContainerApps(ctx context.Context, tenant string, subscription string) (ContainerAppsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return ContainerAppsFacts{}, err
	}

	resourcesClient := session.clientFactory.NewClient()
	rows, issues := collectResourceSummaries(
		ctx,
		resourcesClient,
		"Microsoft.App/containerApps",
		"2024-03-01",
		containerAppSummary,
		"container_apps.resources",
		"container_apps.hydrate",
	)

	return ContainerAppsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		ContainerApps:  rows,
		Issues:         issues,
	}, nil
}

func (provider AzureProvider) ContainerInstances(ctx context.Context, tenant string, subscription string) (ContainerInstancesFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return ContainerInstancesFacts{}, err
	}

	resourcesClient := session.clientFactory.NewClient()
	rows, issues := collectResourceSummaries(
		ctx,
		resourcesClient,
		"Microsoft.ContainerInstance/containerGroups",
		"2023-05-01",
		containerInstanceSummary,
		"container_instances.resources",
		"container_instances.hydrate",
	)

	return ContainerInstancesFacts{
		TenantID:           session.tenantID,
		SubscriptionID:     session.subscription.ID,
		ContainerInstances: rows,
		Issues:             issues,
	}, nil
}

func (provider AzureProvider) EnvVars(ctx context.Context, tenant string, subscription string) (EnvVarsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return EnvVarsFacts{}, err
	}

	webAppsClient, err := armappservice.NewWebAppsClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return EnvVarsFacts{}, fmt.Errorf("build web apps client: %w", err)
	}

	rows := []models.EnvVarSummary{}
	issues := []models.Issue{}
	pager := webAppsClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("env_vars.web_apps", err))
			break
		}
		for _, app := range page.Value {
			appMap := map[string]any{}
			decodeJSONInto(app, &appMap)
			assetKind := webAssetKind(mapStringValue(appMap, "kind"))
			if assetKind == "" {
				continue
			}

			resourceGroup := resourceGroupFromID(mapStringValue(appMap, "id"))
			appName := mapStringValue(appMap, "name")
			if resourceGroup == "" || appName == "" {
				continue
			}

			settings, err := webAppsClient.ListApplicationSettings(ctx, resourceGroup, appName, nil)
			if err != nil {
				issues = append(issues, issueFromError("env_vars["+resourceGroup+"/"+appName+"]", err))
				continue
			}

			settingsMap := map[string]any{}
			decodeJSONInto(settings.StringDictionary, &settingsMap)
			for settingName, settingValue := range mapValue(settingsMap, "properties") {
				rows = append(rows, envVarSummary(appMap, assetKind, settingName, settingValue))
			}
		}
	}

	return EnvVarsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		EnvVars:        rows,
		Issues:         issues,
	}, nil
}

func (provider AzureProvider) session(ctx context.Context, tenant string, subscription string) (azureSession, error) {
	credential, tokenSource, authMode, claims, tenantID, err := newAzureCredential(ctx, tenant)
	if err != nil {
		return azureSession{}, err
	}

	subscriptionsClient, err := armsubscriptions.NewClient(credential, nil)
	if err != nil {
		return azureSession{}, fmt.Errorf("build subscriptions client: %w", err)
	}

	subscriptionRef, err := resolveSubscription(ctx, subscriptionsClient, subscription)
	if err != nil {
		return azureSession{}, err
	}

	clientFactory, err := armresources.NewClientFactory(subscriptionRef.ID, credential, nil)
	if err != nil {
		return azureSession{}, fmt.Errorf("build resource client factory: %w", err)
	}

	return azureSession{
		claims:        claims,
		credential:    credential,
		tokenSource:   tokenSource,
		authMode:      authMode,
		tenantID:      tenantID,
		subscription:  subscriptionRef,
		clientFactory: clientFactory,
	}, nil
}

func newAzureCredential(ctx context.Context, tenant string) (azcore.TokenCredential, string, string, map[string]string, string, error) {
	cliOptions := &azidentity.AzureCLICredentialOptions{}
	if strings.TrimSpace(tenant) != "" {
		cliOptions.TenantID = tenant
	}
	cliFailure := ""
	cliCredential, err := azidentity.NewAzureCLICredential(cliOptions)
	if err == nil {
		token, tokenErr := cliCredential.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: []string{managementScope},
		})
		if tokenErr == nil {
			claims := decodeJWTPayload(token.Token)
			return cliCredential, "azure_cli", "azure_cli", claims, firstNonEmpty(claims["tid"], tenant), nil
		}
		cliFailure = tokenErr.Error()
	} else {
		cliFailure = err.Error()
	}

	envCredential, envErr := azidentity.NewEnvironmentCredential(nil)
	if envErr != nil {
		if cliFailure != "" {
			return nil, "", "", nil, "", fmt.Errorf("azure cli auth failed: %s; environment credential unavailable: %w", cliFailure, envErr)
		}
		return nil, "", "", nil, "", fmt.Errorf("build environment credential: %w", envErr)
	}
	token, err := envCredential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{managementScope},
	})
	if err != nil {
		if cliFailure != "" {
			return nil, "", "", nil, "", fmt.Errorf("azure cli auth failed: %s; environment auth failed: %w", cliFailure, err)
		}
		return nil, "", "", nil, "", fmt.Errorf("authenticate with environment credential: %w", err)
	}

	claims := decodeJWTPayload(token.Token)
	return envCredential, "environment", "environment", claims, firstNonEmpty(claims["tid"], tenant), nil
}

func resolveSubscription(ctx context.Context, client *armsubscriptions.Client, requested string) (models.SubscriptionRef, error) {
	pager := client.NewListPager(nil)
	first := models.SubscriptionRef{}
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return models.SubscriptionRef{}, fmt.Errorf("list subscriptions: %w", err)
		}
		for _, subscription := range page.Value {
			ref := models.SubscriptionRef{
				DisplayName: stringValue(subscription.DisplayName),
				ID:          stringValue(subscription.SubscriptionID),
				State:       stringValue(subscription.State),
			}
			if ref.ID == "" {
				continue
			}
			if first.ID == "" {
				first = ref
			}
			if requested != "" && ref.ID == requested {
				return ref, nil
			}
		}
	}

	if requested != "" {
		return models.SubscriptionRef{}, fmt.Errorf("requested subscription %q not visible to current credential", requested)
	}
	if first.ID == "" {
		return models.SubscriptionRef{}, fmt.Errorf("no subscriptions found for current credential")
	}
	return first, nil
}

func collectResourceSummaries[T any](
	ctx context.Context,
	client *armresources.Client,
	resourceType string,
	apiVersion string,
	summaryFn func(map[string]any) T,
	listIssueScope string,
	hydrateIssueScope string,
) ([]T, []models.Issue) {
	rows := []T{}
	issues := []models.Issue{}

	pager := client.NewListPager(&armresources.ClientListOptions{
		Filter: toPtr("resourceType eq '" + resourceType + "'"),
	})
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError(listIssueScope, err))
			break
		}
		for _, resource := range page.Value {
			resourceID := stringValue(resource.ID)
			hydrated := genericResourceExpandedToMap(*resource)
			if resourceID != "" {
				response, err := client.GetByID(ctx, resourceID, apiVersion, nil)
				if err != nil {
					issues = append(issues, issueFromError(hydrateIssueScope+"["+resourceID+"]", err))
				} else {
					hydrated = genericResourceToMap(&response.GenericResource)
				}
			}
			rows = append(rows, summaryFn(hydrated))
		}
	}

	return rows, issues
}

func deploymentSummary(
	deployment *armresources.DeploymentExtended,
	scope string,
	scopeType string,
	resourceGroup *string,
) models.ArmDeploymentSummary {
	if deployment == nil {
		return models.ArmDeploymentSummary{
			ID:        scope + "/providers/Microsoft.Resources/deployments/unknown",
			Name:      "unknown",
			Scope:     scope,
			ScopeType: scopeType,
			Summary:   scopeType + " deployment 'unknown' is unknown with no outputs recorded; no providers recorded.",
		}
	}

	name := stringValue(deployment.Name)
	if name == "" {
		name = "unknown"
	}
	deploymentID := stringValue(deployment.ID)
	if deploymentID == "" {
		deploymentID = scope + "/providers/Microsoft.Resources/deployments/" + name
	}

	state := ""
	mode := ""
	timestamp := ""
	duration := ""
	outputsCount := 0
	outputResourceCount := 0
	templateLink := (*string)(nil)
	parametersLink := (*string)(nil)
	providersList := []string{}

	if deployment.Properties != nil {
		state = stringValue(deployment.Properties.ProvisioningState)
		mode = stringValue(deployment.Properties.Mode)
		timestamp = timeValue(deployment.Properties.Timestamp)
		duration = stringValue(deployment.Properties.Duration)
		if deployment.Properties.Outputs != nil {
			outputMap := map[string]any{}
			if decodeJSONInto(deployment.Properties.Outputs, &outputMap) {
				outputsCount = len(outputMap)
			}
		}
		outputResourceCount = len(deployment.Properties.OutputResources)
		if deployment.Properties.TemplateLink != nil {
			templateLink = stringPtrValueOrNil(deployment.Properties.TemplateLink.URI)
		}
		if deployment.Properties.ParametersLink != nil {
			parametersLink = stringPtrValueOrNil(deployment.Properties.ParametersLink.URI)
		}
		for _, provider := range deployment.Properties.Providers {
			namespace := ""
			if provider != nil {
				namespace = stringValue(provider.Namespace)
			}
			if namespace != "" && !slices.Contains(providersList, namespace) {
				providersList = append(providersList, namespace)
			}
		}
	}

	providerSummary := "no providers recorded"
	if len(providersList) > 0 {
		providerSummary = strconv.Itoa(len(providersList)) + " providers"
	}
	outputSummary := "no outputs recorded"
	if outputsCount > 0 {
		outputSummary = strconv.Itoa(outputsCount) + " outputs"
	}

	return models.ArmDeploymentSummary{
		Duration:            duration,
		ID:                  deploymentID,
		Mode:                mode,
		Name:                name,
		OutputResourceCount: outputResourceCount,
		OutputsCount:        outputsCount,
		ParametersLink:      parametersLink,
		Providers:           providersList,
		ProvisioningState:   state,
		RelatedIDs:          []string{deploymentID},
		ResourceGroup:       resourceGroup,
		Scope:               scope,
		ScopeType:           scopeType,
		Summary:             strings.ReplaceAll(scopeType, "_", " ") + " deployment '" + name + "' is " + valueOrUnknown(state) + " with " + outputSummary + "; " + providerSummary + ".",
		TemplateLink:        templateLink,
		Timestamp:           timestamp,
	}
}

func containerAppSummary(resource map[string]any) models.ContainerAppAsset {
	resourceID := stringMapValue(resource, "id")
	name := stringMapValue(resource, "name")
	if name == "" {
		name = "unknown"
	}

	identity := mapValue(resource, "identity")
	properties := mapValue(resource, "properties")
	configuration := mapValue(properties, "configuration")
	ingress := mapValue(configuration, "ingress")

	workloadIdentityIDs := sortedKeys(mapValue(identity, "userAssignedIdentities"), "user_assigned_identities")
	workloadIdentityType := stringPtr(mapStringValue(identity, "type"))
	workloadPrincipalID := stringPtr(mapStringValue(identity, "principalId", "principal_id"))
	workloadClientID := stringPtr(mapStringValue(identity, "clientId", "client_id"))
	defaultHostname := stringPtr(mapStringValue(ingress, "fqdn"))
	externalIngressEnabled := boolPtr(mapBoolValue(ingress, "external"))
	ingressTargetPort := intPtr(mapIntValue(ingress, "targetPort", "target_port"))
	ingressTransport := stringPtr(mapStringValue(ingress, "transport"))
	revisionMode := stringPtr(mapStringValue(configuration, "activeRevisionsMode", "active_revisions_mode"))
	latestRevisionName := stringPtr(mapStringValue(properties, "latestRevisionName", "latest_revision_name"))
	latestReadyRevisionName := stringPtr(mapStringValue(properties, "latestReadyRevisionName", "latest_ready_revision_name"))
	environmentID := stringPtr(mapStringValue(properties, "managedEnvironmentId", "managed_environment_id"))

	ingressParts := []string{}
	if externalIngressEnabled != nil {
		if *externalIngressEnabled {
			ingressParts = append(ingressParts, "external ingress enabled")
		} else {
			ingressParts = append(ingressParts, "internal ingress only")
		}
	}
	if ingressTargetPort != nil {
		ingressParts = append(ingressParts, "target port "+strconv.Itoa(*ingressTargetPort))
	}
	if ingressTransport != nil && *ingressTransport != "" {
		ingressParts = append(ingressParts, "transport "+*ingressTransport)
	}

	revisionParts := []string{}
	if revisionMode != nil && *revisionMode != "" {
		revisionParts = append(revisionParts, "revision mode "+*revisionMode)
	}
	if latestReadyRevisionName != nil && *latestReadyRevisionName != "" {
		revisionParts = append(revisionParts, "latest ready revision "+*latestReadyRevisionName)
	} else if latestRevisionName != nil && *latestRevisionName != "" {
		revisionParts = append(revisionParts, "latest revision "+*latestRevisionName)
	}

	endpointPhrase := "has no visible hostname from the current read path"
	if defaultHostname != nil && *defaultHostname != "" {
		endpointPhrase = "publishes hostname '" + *defaultHostname + "'"
	}
	identityPhrase := "has no managed identity visible from the current read path"
	if workloadIdentityType != nil && *workloadIdentityType != "" {
		identityPhrase = "uses managed identity (" + *workloadIdentityType + ")"
	}
	postureParts := append(append([]string{}, ingressParts...), revisionParts...)
	posturePhrase := ""
	if len(postureParts) > 0 {
		posturePhrase = " Visible posture: " + strings.Join(postureParts, ", ") + "."
	}

	return models.ContainerAppAsset{
		DefaultHostname:         defaultHostname,
		EnvironmentID:           environmentID,
		ExternalIngressEnabled:  externalIngressEnabled,
		ID:                      firstNonEmpty(resourceID, "/unknown/"+name),
		IngressTargetPort:       ingressTargetPort,
		IngressTransport:        ingressTransport,
		LatestReadyRevisionName: latestReadyRevisionName,
		LatestRevisionName:      latestRevisionName,
		Location:                stringMapValue(resource, "location"),
		Name:                    name,
		RelatedIDs: dedupeStrings(
			append([]string{
				resourceID,
				stringPtrValue(environmentID),
				stringPtrValue(workloadPrincipalID),
			}, workloadIdentityIDs...),
		),
		ResourceGroup:        resourceGroupFromID(resourceID),
		RevisionMode:         revisionMode,
		Summary:              "Container App '" + name + "' " + endpointPhrase + " and " + identityPhrase + "." + posturePhrase,
		WorkloadClientID:     workloadClientID,
		WorkloadIdentityIDs:  workloadIdentityIDs,
		WorkloadIdentityType: workloadIdentityType,
		WorkloadPrincipalID:  workloadPrincipalID,
	}
}

func containerInstanceSummary(resource map[string]any) models.ContainerInstanceAsset {
	resourceID := stringMapValue(resource, "id")
	name := stringMapValue(resource, "name")
	if name == "" {
		name = "unknown"
	}

	identity := mapValue(resource, "identity")
	properties := mapValue(resource, "properties")
	ipAddress := mapValue(properties, "ipAddress", "ip_address")
	containers := listValue(properties, "containers")

	workloadIdentityIDs := sortedKeys(mapValue(identity, "userAssignedIdentities"), "user_assigned_identities")
	workloadIdentityType := stringPtr(mapStringValue(identity, "type"))
	workloadPrincipalID := stringPtr(mapStringValue(identity, "principalId", "principal_id"))
	workloadClientID := stringPtr(mapStringValue(identity, "clientId", "client_id"))
	fqdn := stringPtr(mapStringValue(ipAddress, "fqdn"))
	publicIPAddress := stringPtr(mapStringValue(ipAddress, "ip"))
	exposedPorts := []int{}
	for _, entry := range listValue(ipAddress, "ports") {
		if port := mapIntValue(entry, "port"); port != 0 {
			exposedPorts = append(exposedPorts, port)
		}
	}
	sort.Ints(exposedPorts)
	exposedPorts = slices.Compact(exposedPorts)

	subnetIDs := []string{}
	for _, entry := range listValue(properties, "subnetIds", "subnet_ids") {
		value := mapStringValue(entry, "id")
		if value != "" {
			subnetIDs = append(subnetIDs, value)
		}
	}
	subnetIDs = dedupeStrings(subnetIDs)

	containerImages := []string{}
	for _, entry := range containers {
		image := mapStringValue(mapValue(entry, "properties"), "image")
		if image != "" {
			containerImages = append(containerImages, image)
		}
	}
	containerImages = dedupeStrings(containerImages)

	restartPolicy := stringPtr(mapStringValue(properties, "restartPolicy", "restart_policy"))
	osType := stringPtr(mapStringValue(properties, "osType", "os_type"))
	provisioningState := stringPtr(mapStringValue(properties, "provisioningState", "provisioning_state"))
	containerCount := intPtr(len(containers))

	endpointParts := []string{}
	if fqdn != nil && *fqdn != "" {
		endpointParts = append(endpointParts, "publishes FQDN '"+*fqdn+"'")
	}
	if publicIPAddress != nil && *publicIPAddress != "" {
		endpointParts = append(endpointParts, "uses public IP "+*publicIPAddress)
	}
	if len(endpointParts) == 0 {
		endpointParts = append(endpointParts, "has no public endpoint visible from the current read path")
	}

	postureParts := []string{}
	if osType != nil && *osType != "" {
		postureParts = append(postureParts, "os "+*osType)
	}
	if restartPolicy != nil && *restartPolicy != "" {
		postureParts = append(postureParts, "restart "+*restartPolicy)
	}
	if len(exposedPorts) > 0 {
		portStrings := make([]string, 0, len(exposedPorts))
		for _, port := range exposedPorts {
			portStrings = append(portStrings, strconv.Itoa(port))
		}
		postureParts = append(postureParts, "ports "+strings.Join(portStrings, ", "))
	}
	if len(subnetIDs) > 0 {
		postureParts = append(postureParts, "subnets "+strconv.Itoa(len(subnetIDs)))
	}
	if len(containers) > 0 {
		postureParts = append(postureParts, strconv.Itoa(len(containers))+" container(s)")
	}
	identityPhrase := "has no managed identity visible from the current read path"
	if workloadIdentityType != nil && *workloadIdentityType != "" {
		identityPhrase = "uses managed identity (" + *workloadIdentityType + ")"
	}
	posturePhrase := ""
	if len(postureParts) > 0 {
		posturePhrase = " Visible posture: " + strings.Join(postureParts, ", ") + "."
	}

	return models.ContainerInstanceAsset{
		ContainerCount:    containerCount,
		ContainerImages:   containerImages,
		ExposedPorts:      exposedPorts,
		FQDN:              fqdn,
		ID:                firstNonEmpty(resourceID, "/unknown/"+name),
		Location:          stringMapValue(resource, "location"),
		Name:              name,
		OSType:            osType,
		ProvisioningState: provisioningState,
		PublicIPAddress:   publicIPAddress,
		RelatedIDs: dedupeStrings(
			append([]string{
				resourceID,
				stringPtrValue(workloadPrincipalID),
			}, append(workloadIdentityIDs, subnetIDs...)...),
		),
		ResourceGroup:        resourceGroupFromID(resourceID),
		RestartPolicy:        restartPolicy,
		SubnetIDs:            subnetIDs,
		Summary:              "Container Instance '" + name + "' " + strings.Join(endpointParts, " and ") + " and " + identityPhrase + "." + posturePhrase,
		WorkloadClientID:     workloadClientID,
		WorkloadIdentityIDs:  workloadIdentityIDs,
		WorkloadIdentityType: workloadIdentityType,
		WorkloadPrincipalID:  workloadPrincipalID,
	}
}

func appServiceSummary(app map[string]any, config map[string]any) models.AppServiceAsset {
	appID := mapStringValue(app, "id")
	appName := mapStringValue(app, "name")
	if appName == "" {
		appName = "unknown"
	}

	identity := mapValue(app, "identity")
	publicNetworkAccess := stringPtr(mapStringValue(app, "publicNetworkAccess", "public_network_access"))
	runtimeStack := appServiceRuntimeStack(config)
	minTLSVersion := stringPtr(mapStringValue(config, "minTlsVersion", "min_tls_version"))
	ftpsState := stringPtr(mapStringValue(config, "ftpsState", "ftps_state"))
	workloadIdentityType := stringPtr(mapStringValue(identity, "type"))
	workloadPrincipalID := stringPtr(mapStringValue(identity, "principalId", "principal_id"))
	workloadClientID := stringPtr(mapStringValue(identity, "clientId", "client_id"))
	workloadIdentityIDs := sortedKeys(mapValue(identity, "userAssignedIdentities"), "user_assigned_identities")

	postureParts := []string{
		"public network access " + valueOrUnknown(stringPtrValue(publicNetworkAccess)),
		"HTTPS-only " + disabledEnabled(mapBoolValue(app, "httpsOnly", "https_only")),
	}
	if minTLSVersion != nil && *minTLSVersion != "" {
		postureParts = append(postureParts, "TLS "+*minTLSVersion)
	}
	if ftpsState != nil && *ftpsState != "" {
		postureParts = append(postureParts, "FTPS "+*ftpsState)
	}

	hostnamePhrase := "has no default hostname visible from the current read path"
	if hostname := stringPtr(mapStringValue(app, "defaultHostName", "default_host_name")); hostname != nil {
		hostnamePhrase = "publishes hostname '" + *hostname + "'"
	}
	runtimePhrase := "does not expose a readable runtime summary from the current read path"
	if runtimeStack != nil && *runtimeStack != "" {
		runtimePhrase = "runs runtime '" + *runtimeStack + "'"
	}
	identityPhrase := "has no managed identity visible from the current read path"
	if workloadIdentityType != nil && *workloadIdentityType != "" {
		identityPhrase = "uses managed identity (" + *workloadIdentityType + ")"
	}

	return models.AppServiceAsset{
		AppServicePlanID:    stringPtr(mapStringValue(app, "serverFarmId", "server_farm_id")),
		ClientCertEnabled:   mapBoolValue(app, "clientCertEnabled", "client_cert_enabled"),
		DefaultHostname:     stringPtr(mapStringValue(app, "defaultHostName", "default_host_name")),
		FTPSState:           ftpsState,
		HTTPSOnly:           mapBoolValue(app, "httpsOnly", "https_only"),
		ID:                  firstNonEmpty(appID, "/unknown/"+appName),
		Location:            mapStringValue(app, "location"),
		MinTLSVersion:       minTLSVersion,
		Name:                appName,
		PublicNetworkAccess: publicNetworkAccess,
		RelatedIDs: dedupeStrings(
			append([]string{
				appID,
				stringPtrValue(workloadPrincipalID),
				mapStringValue(app, "serverFarmId", "server_farm_id"),
			}, workloadIdentityIDs...),
		),
		ResourceGroup:        resourceGroupFromID(appID),
		RuntimeStack:         runtimeStack,
		State:                stringPtr(mapStringValue(app, "state")),
		Summary:              "App Service '" + appName + "' " + hostnamePhrase + ", " + runtimePhrase + ", and " + identityPhrase + ". Visible posture: " + strings.Join(postureParts, ", ") + ".",
		WorkloadClientID:     workloadClientID,
		WorkloadIdentityIDs:  workloadIdentityIDs,
		WorkloadIdentityType: workloadIdentityType,
		WorkloadPrincipalID:  workloadPrincipalID,
	}
}

func functionAppSummary(app map[string]any, config map[string]any, settings map[string]any) models.FunctionAppAsset {
	appID := mapStringValue(app, "id")
	appName := mapStringValue(app, "name")
	if appName == "" {
		appName = "unknown"
	}

	identity := mapValue(app, "identity")
	publicNetworkAccess := stringPtr(mapStringValue(app, "publicNetworkAccess", "public_network_access"))
	runtimeStack := appServiceRuntimeStack(config)
	minTLSVersion := stringPtr(mapStringValue(config, "minTlsVersion", "min_tls_version"))
	ftpsState := stringPtr(mapStringValue(config, "ftpsState", "ftps_state"))
	functionsExtensionVersion := stringPtr(mapStringValue(config, "functionsExtensionVersion", "functions_extension_version"))
	alwaysOn := optionalBoolPtr(config, "alwaysOn", "always_on")
	workloadIdentityType := stringPtr(mapStringValue(identity, "type"))
	workloadPrincipalID := stringPtr(mapStringValue(identity, "principalId", "principal_id"))
	workloadClientID := stringPtr(mapStringValue(identity, "clientId", "client_id"))
	workloadIdentityIDs := sortedKeys(mapValue(identity, "userAssignedIdentities"), "user_assigned_identities")
	azureWebJobsStorageValue, hasAzureWebJobsStorage := settings["AzureWebJobsStorage"]
	azureWebJobsStorageValueType := stringPtr(envVarValueType(azureWebJobsStorageValue, hasAzureWebJobsStorage))
	azureWebJobsStorageReferenceTarget := stringPtr(envVarReferenceTarget(azureWebJobsStorageValue))
	runFromPackage := runFromPackageSignal(settings)
	keyVaultReferenceCount := keyVaultReferenceCount(settings)

	hostnamePhrase := "has no default hostname visible from the current read path"
	if hostname := stringPtr(mapStringValue(app, "defaultHostName", "default_host_name")); hostname != nil {
		hostnamePhrase = "publishes hostname '" + *hostname + "'"
	}
	runtimePhrase := "does not expose a readable runtime summary from the current read path"
	if runtimeStack != nil && *runtimeStack != "" {
		runtimePhrase = "runs runtime '" + *runtimeStack + "'"
	}
	functionsPhrase := "does not expose a readable Functions runtime version from the current read path"
	if functionsExtensionVersion != nil && *functionsExtensionVersion != "" {
		functionsPhrase = "targets Functions runtime '" + *functionsExtensionVersion + "'"
	}
	identityPhrase := "has no managed identity visible from the current read path"
	if workloadIdentityType != nil && *workloadIdentityType != "" {
		identityPhrase = "uses managed identity (" + *workloadIdentityType + ")"
	}

	deploymentParts := []string{}
	switch stringPtrValue(azureWebJobsStorageValueType) {
	case "keyvault-ref":
		target := ""
		if azureWebJobsStorageReferenceTarget != nil && *azureWebJobsStorageReferenceTarget != "" {
			target = " (" + *azureWebJobsStorageReferenceTarget + ")"
		}
		deploymentParts = append(deploymentParts, "AzureWebJobsStorage via Key Vault reference"+target)
	case "plain-text":
		deploymentParts = append(deploymentParts, "AzureWebJobsStorage as plain-text app setting")
	case "empty":
		deploymentParts = append(deploymentParts, "AzureWebJobsStorage visible but empty")
	case "missing":
		deploymentParts = append(deploymentParts, "no AzureWebJobsStorage setting visible")
	}
	if runFromPackage != nil {
		if *runFromPackage {
			deploymentParts = append(deploymentParts, "run-from-package enabled")
		} else {
			deploymentParts = append(deploymentParts, "run-from-package disabled")
		}
	}
	if keyVaultReferenceCount != nil {
		deploymentParts = append(deploymentParts, strconv.Itoa(*keyVaultReferenceCount)+" Key Vault-backed setting(s)")
	}

	postureParts := []string{
		"public network access " + valueOrUnknown(stringPtrValue(publicNetworkAccess)),
		"HTTPS-only " + disabledEnabled(mapBoolValue(app, "httpsOnly", "https_only")),
	}
	if minTLSVersion != nil && *minTLSVersion != "" {
		postureParts = append(postureParts, "TLS "+*minTLSVersion)
	}
	if ftpsState != nil && *ftpsState != "" {
		postureParts = append(postureParts, "FTPS "+*ftpsState)
	}
	if alwaysOn != nil {
		postureParts = append(postureParts, "Always On "+disabledEnabled(*alwaysOn))
	}
	deploymentPhrase := "Deployment signals are not readable from the current read path."
	if len(deploymentParts) > 0 {
		deploymentPhrase = "Deployment signals: " + strings.Join(deploymentParts, ", ") + "."
	}

	return models.FunctionAppAsset{
		AlwaysOn:                           alwaysOn,
		AppServicePlanID:                   stringPtr(mapStringValue(app, "serverFarmId", "server_farm_id")),
		AzureWebJobsStorageReferenceTarget: azureWebJobsStorageReferenceTarget,
		AzureWebJobsStorageValueType:       azureWebJobsStorageValueType,
		ClientCertEnabled:                  mapBoolValue(app, "clientCertEnabled", "client_cert_enabled"),
		DefaultHostname:                    stringPtr(mapStringValue(app, "defaultHostName", "default_host_name")),
		FTPSState:                          ftpsState,
		FunctionsExtensionVersion:          functionsExtensionVersion,
		HTTPSOnly:                          mapBoolValue(app, "httpsOnly", "https_only"),
		ID:                                 firstNonEmpty(appID, "/unknown/"+appName),
		KeyVaultReferenceCount:             keyVaultReferenceCount,
		Location:                           mapStringValue(app, "location"),
		MinTLSVersion:                      minTLSVersion,
		Name:                               appName,
		PublicNetworkAccess:                publicNetworkAccess,
		RelatedIDs: dedupeStrings(
			append([]string{
				appID,
				stringPtrValue(workloadPrincipalID),
				mapStringValue(app, "serverFarmId", "server_farm_id"),
			}, workloadIdentityIDs...),
		),
		ResourceGroup:        resourceGroupFromID(appID),
		RunFromPackage:       runFromPackage,
		RuntimeStack:         runtimeStack,
		State:                stringPtr(mapStringValue(app, "state")),
		Summary:              "Function App '" + appName + "' " + hostnamePhrase + ", " + runtimePhrase + ", " + functionsPhrase + ", and " + identityPhrase + ". " + deploymentPhrase + " Visible posture: " + strings.Join(postureParts, ", ") + ".",
		WorkloadClientID:     workloadClientID,
		WorkloadIdentityIDs:  workloadIdentityIDs,
		WorkloadIdentityType: workloadIdentityType,
		WorkloadPrincipalID:  workloadPrincipalID,
	}
}

func envVarSummary(app map[string]any, assetKind string, settingName string, settingValue any) models.EnvVarSummary {
	appID := mapStringValue(app, "id")
	appName := mapStringValue(app, "name")
	if appName == "" {
		appName = "unknown"
	}

	identity := mapValue(app, "identity")
	valueType := envVarValueType(settingValue, true)
	looksSensitive := looksSensitiveSettingName(settingName)
	referenceTarget := stringPtr(envVarReferenceTarget(settingValue))
	workloadIdentityType := stringPtr(mapStringValue(identity, "type"))
	workloadPrincipalID := stringPtr(mapStringValue(identity, "principalId", "principal_id"))
	workloadClientID := stringPtr(mapStringValue(identity, "clientId", "client_id"))
	workloadIdentityIDs := sortedKeys(mapValue(identity, "userAssignedIdentities"), "user_assigned_identities")
	keyVaultReferenceIdentity := stringPtr(mapStringValue(app, "keyVaultReferenceIdentity", "key_vault_reference_identity"))

	summary := assetKind + " '" + appName + "' exposes setting '" + settingName + "' through management-plane app settings (" + valueType + ")."
	if valueType == "keyvault-ref" {
		summary = assetKind + " '" + appName + "' maps setting '" + settingName + "' to Key Vault-backed configuration"
		if referenceTarget != nil && *referenceTarget != "" {
			summary += " (" + *referenceTarget + ")"
		}
		if keyVaultReferenceIdentity != nil && *keyVaultReferenceIdentity != "" {
			summary += " via " + *keyVaultReferenceIdentity
		}
		summary += "."
	} else if looksSensitive && valueType == "plain-text" {
		summary = assetKind + " '" + appName + "' stores sensitive-looking setting '" + settingName + "' as plain-text app configuration."
	}
	summary = strings.TrimSuffix(summary, ".") + " " + envVarNextReviewHint(settingName, valueType, looksSensitive, stringPtrValue(referenceTarget), stringPtrValue(workloadIdentityType))

	return models.EnvVarSummary{
		AssetID:                   firstNonEmpty(appID, "/unknown/"+appName),
		AssetKind:                 assetKind,
		AssetName:                 appName,
		KeyVaultReferenceIdentity: keyVaultReferenceIdentity,
		Location:                  mapStringValue(app, "location"),
		LooksSensitive:            looksSensitive,
		ReferenceTarget:           referenceTarget,
		RelatedIDs:                dedupeStrings([]string{appID}),
		ResourceGroup:             resourceGroupFromID(appID),
		SettingName:               settingName,
		Summary:                   summary,
		ValueType:                 valueType,
		WorkloadClientID:          workloadClientID,
		WorkloadIdentityIDs:       workloadIdentityIDs,
		WorkloadIdentityType:      workloadIdentityType,
		WorkloadPrincipalID:       workloadPrincipalID,
		TargetServices:            envVarTargetServices(settingName),
	}
}

func genericResourceToMap(resource *armresources.GenericResource) map[string]any {
	if resource == nil {
		return map[string]any{}
	}
	data := map[string]any{}
	decodeJSONInto(resource, &data)
	return data
}

func genericResourceExpandedToMap(resource armresources.GenericResourceExpanded) map[string]any {
	data := map[string]any{}
	decodeJSONInto(resource, &data)
	return data
}

func appendDecodedMaps[T any](rows []map[string]any, values []*T) []map[string]any {
	for _, value := range values {
		if value == nil {
			continue
		}
		data := map[string]any{}
		if decodeJSONInto(value, &data) {
			rows = append(rows, data)
		}
	}
	return rows
}

func decodeJSONInto(input any, target any) bool {
	encoded, err := json.Marshal(input)
	if err != nil {
		return false
	}
	if err := json.Unmarshal(encoded, target); err != nil {
		return false
	}
	return true
}

func decodeJWTPayload(token string) map[string]string {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return map[string]string{}
	}

	payload := parts[1]
	if padding := len(payload) % 4; padding != 0 {
		payload += strings.Repeat("=", 4-padding)
	}

	raw, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		raw, err = base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return map[string]string{}
		}
	}

	decoded := map[string]any{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return map[string]string{}
	}

	claims := map[string]string{}
	for key, value := range decoded {
		switch typed := value.(type) {
		case string:
			claims[key] = typed
		case float64:
			claims[key] = strconv.FormatFloat(typed, 'f', -1, 64)
		case bool:
			claims[key] = strconv.FormatBool(typed)
		}
	}
	return claims
}

func principalTypeFromClaims(claims map[string]string) string {
	if claims["xms_mirid"] != "" {
		return "ServicePrincipal"
	}
	if looksLikeUserClaims(claims) {
		return "User"
	}
	if strings.EqualFold(claims["idtyp"], "app") || claims["appid"] != "" {
		return "ServicePrincipal"
	}
	if claims["oid"] != "" {
		return "User"
	}
	return "Unknown"
}

func looksLikeUserClaims(claims map[string]string) bool {
	if strings.EqualFold(claims["idtyp"], "app") {
		return false
	}
	return claims["upn"] != "" ||
		claims["preferred_username"] != "" ||
		claims["unique_name"] != "" ||
		claims["scp"] != "" ||
		(claims["oid"] != "" && claims["appid"] == "")
}

func liveCollectionNotImplemented(command string) error {
	return fmt.Errorf("azure provider does not implement live %s collection yet", command)
}

func issueFromError(scope string, err error) models.Issue {
	return models.Issue{
		Kind:    "collection_error",
		Message: scope + ": " + err.Error(),
		Scope:   scope,
		Context: map[string]string{"collector": scope},
	}
}

func resourceGroupFromID(resourceID string) string {
	parts := strings.Split(strings.Trim(resourceID, "/"), "/")
	for index := 0; index < len(parts)-1; index++ {
		if strings.EqualFold(parts[index], "resourceGroups") {
			return parts[index+1]
		}
	}
	return ""
}

func timeValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func disabledEnabled(value bool) string {
	if value {
		return "enabled"
	}
	return "disabled"
}

func toPtr(value string) *string {
	return &value
}

func valueOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case *string:
		if typed == nil {
			return ""
		}
		return *typed
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		reflected := reflect.ValueOf(value)
		for reflected.IsValid() && reflected.Kind() == reflect.Pointer {
			if reflected.IsNil() {
				return ""
			}
			reflected = reflected.Elem()
		}
		if !reflected.IsValid() {
			return ""
		}
		if reflected.CanInterface() {
			if stringer, ok := reflected.Interface().(fmt.Stringer); ok {
				return stringer.String()
			}
			if reflected.Kind() == reflect.String {
				return reflected.String()
			}
			return fmt.Sprint(reflected.Interface())
		}
		return fmt.Sprint(typed)
	}
}

func stringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringPtrValueOrNil(value any) *string {
	return stringPtr(stringValue(value))
}

func intPtr(value int) *int {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func optionalBoolPtr(input any, keys ...string) *bool {
	if mapped, ok := input.(map[string]any); ok {
		for _, key := range keys {
			if value, exists := mapped[key]; exists {
				if typed, ok := value.(bool); ok {
					return &typed
				}
			}
		}
	}
	return nil
}

func mapValue(input any, keys ...string) map[string]any {
	switch typed := input.(type) {
	case map[string]any:
		if len(keys) == 0 {
			return typed
		}
		for _, key := range keys {
			if next, ok := typed[key]; ok {
				if nested, ok := next.(map[string]any); ok {
					return nested
				}
			}
		}
	case nil:
	}
	return map[string]any{}
}

func listValue(input any, keys ...string) []any {
	if len(keys) > 0 {
		if mapped, ok := input.(map[string]any); ok {
			for _, key := range keys {
				if value, exists := mapped[key]; exists {
					return listValue(value)
				}
			}
			return []any{}
		}
	}
	switch typed := input.(type) {
	case []any:
		return typed
	case nil:
		return []any{}
	default:
		return []any{}
	}
}

func mapStringValue(input any, keys ...string) string {
	if mapped, ok := input.(map[string]any); ok {
		for _, key := range keys {
			if value, exists := mapped[key]; exists {
				return stringValue(value)
			}
		}
	}
	return ""
}

func stringMapValue(input map[string]any, key string) string {
	return mapStringValue(input, key)
}

func mapIntValue(input any, keys ...string) int {
	if mapped, ok := input.(map[string]any); ok {
		for _, key := range keys {
			if value, exists := mapped[key]; exists {
				switch typed := value.(type) {
				case int:
					return typed
				case int32:
					return int(typed)
				case int64:
					return int(typed)
				case float64:
					return int(typed)
				case json.Number:
					intValue, _ := typed.Int64()
					return int(intValue)
				}
			}
		}
	}
	return 0
}

func mapBoolValue(input any, keys ...string) bool {
	if mapped, ok := input.(map[string]any); ok {
		for _, key := range keys {
			if value, exists := mapped[key]; exists {
				if typed, ok := value.(bool); ok {
					return typed
				}
			}
		}
	}
	return false
}

func sortedKeys(input map[string]any, aliases ...string) []string {
	values := input
	if len(values) == 0 {
		for _, alias := range aliases {
			if candidate, ok := input[alias].(map[string]any); ok {
				values = candidate
				break
			}
		}
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	deduped := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		deduped = append(deduped, value)
	}
	return deduped
}

func dedupeDeployments(values []models.ArmDeploymentSummary) []models.ArmDeploymentSummary {
	seen := map[string]struct{}{}
	deduped := make([]models.ArmDeploymentSummary, 0, len(values))
	for _, value := range values {
		if value.ID == "" {
			deduped = append(deduped, value)
			continue
		}
		if _, exists := seen[value.ID]; exists {
			continue
		}
		seen[value.ID] = struct{}{}
		deduped = append(deduped, value)
	}
	return deduped
}

func webAssetKind(kind string) string {
	value := strings.ToLower(kind)
	if strings.Contains(value, "workflowapp") {
		return ""
	}
	if strings.Contains(value, "functionapp") {
		return "FunctionApp"
	}
	if value == "" || strings.Contains(value, "app") {
		return "AppService"
	}
	return ""
}

func appServiceRuntimeStack(config map[string]any) *string {
	if len(config) == 0 {
		return nil
	}
	for _, key := range []string{"linuxFxVersion", "linux_fx_version"} {
		if value := mapStringValue(config, key); value != "" {
			return &value
		}
	}
	for _, key := range []string{"windowsFxVersion", "windows_fx_version"} {
		if value := mapStringValue(config, key); value != "" {
			return &value
		}
	}

	runtimeParts := []string{}
	for _, pair := range []struct {
		key   string
		label string
	}{
		{key: "pythonVersion", label: "python"},
		{key: "python_version", label: "python"},
		{key: "nodeVersion", label: "node"},
		{key: "node_version", label: "node"},
		{key: "powerShellVersion", label: "powershell"},
		{key: "power_shell_version", label: "powershell"},
		{key: "javaVersion", label: "java"},
		{key: "java_version", label: "java"},
		{key: "phpVersion", label: "php"},
		{key: "php_version", label: "php"},
		{key: "netFrameworkVersion", label: ".net"},
		{key: "net_framework_version", label: ".net"},
	} {
		if value := mapStringValue(config, pair.key); value != "" {
			runtimeParts = append(runtimeParts, pair.label+"="+value)
		}
	}
	if len(runtimeParts) == 0 {
		return nil
	}
	stack := strings.Join(runtimeParts, "; ")
	return &stack
}

func runFromPackageSignal(settings map[string]any) *bool {
	value, exists := settings["WEBSITE_RUN_FROM_PACKAGE"]
	if !exists {
		return nil
	}
	normalized := strings.ToLower(strings.TrimSpace(stringValue(value)))
	if normalized == "" {
		return nil
	}
	disabled := map[string]struct{}{"0": {}, "false": {}, "no": {}, "off": {}, "disabled": {}}
	if _, found := disabled[normalized]; found {
		result := false
		return &result
	}
	result := true
	return &result
}

func envVarValueType(value any, exists bool) string {
	if !exists {
		return "missing"
	}
	text := strings.TrimSpace(stringValue(value))
	if text == "" {
		return "empty"
	}
	if strings.HasPrefix(text, "@Microsoft.KeyVault(") {
		return "keyvault-ref"
	}
	return "plain-text"
}

func keyVaultReferenceCount(settings map[string]any) *int {
	if settings == nil {
		return nil
	}
	count := 0
	for _, value := range settings {
		if envVarValueType(value, true) == "keyvault-ref" {
			count++
		}
	}
	return &count
}

func looksSensitiveSettingName(settingName string) bool {
	normalized := strings.ToLower(settingName)
	replacer := strings.NewReplacer("-", "", "_", "", ".", "", " ", "")
	normalized = replacer.Replace(normalized)
	for _, token := range []string{
		"key",
		"secret",
		"token",
		"password",
		"passwd",
		"connectionstring",
		"connection_string",
		"connstr",
		"clientsecret",
	} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func envVarReferenceTarget(value any) string {
	text := strings.TrimSpace(stringValue(value))
	if !strings.HasPrefix(text, "@Microsoft.KeyVault(") {
		return ""
	}
	if match := regexpMustCompile(`SecretUri=([^)]+)`).FindStringSubmatch(text); len(match) == 2 {
		return compactLink(match[1])
	}
	vaultName := keyVaultReferencePart(text, "VaultName")
	secretName := keyVaultReferencePart(text, "SecretName")
	secretVersion := keyVaultReferencePart(text, "SecretVersion")
	if vaultName == "" || secretName == "" {
		return ""
	}
	target := vaultName + ".vault.azure.net/secrets/" + secretName
	if secretVersion != "" {
		target += "/" + secretVersion
	}
	return target
}

func compactLink(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	if parsed, err := url.Parse(text); err == nil && parsed.Host != "" && parsed.Path != "" {
		return parsed.Host + parsed.Path
	}
	return text
}

func keyVaultReferenceIdentitySummary(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	if strings.EqualFold(text, "systemassigned") {
		return "SystemAssigned"
	}
	parts := strings.Split(strings.Trim(text, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return text
}

func keyVaultReferencePart(text string, key string) string {
	match := regexpMustCompile(key + `=([^;)]*)`).FindStringSubmatch(text)
	if len(match) != 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func envVarNextReviewHint(settingName string, valueType string, looksSensitive bool, referenceTarget string, workloadIdentityType string) string {
	targetServices := envVarTargetServices(settingName)
	hasIdentity := workloadIdentityType != ""

	if valueType == "keyvault-ref" {
		if hasIdentity {
			return "Check keyvault for the referenced secret path; review managed-identities for the workload token path."
		}
		return "Check keyvault for the referenced secret path."
	}
	if looksSensitive && valueType == "plain-text" {
		if slices.Contains(targetServices, models.EnvVarTargetServiceStorage) {
			return "Check tokens-credentials first; this likely feeds a storage credential path."
		}
		if slices.Contains(targetServices, models.EnvVarTargetServiceDatabase) {
			return "Check tokens-credentials first; this likely feeds a database credential path."
		}
		return "Check tokens-credentials for the workload credential surface."
	}
	if slices.Contains(targetServices, models.EnvVarTargetServiceStorage) {
		if hasIdentity {
			return "Check tokens-credentials for the config-backed access path, then managed-identities for the workload token path."
		}
		return "Check tokens-credentials for the config-backed storage access path."
	}
	if slices.Contains(targetServices, models.EnvVarTargetServiceDatabase) {
		return "Check tokens-credentials for the config-backed database access path."
	}
	if strings.Contains(strings.ToLower(referenceTarget), "vault") {
		return "Check keyvault for the referenced secret path."
	}
	if hasIdentity {
		return "Check managed-identities for the workload token path behind this setting."
	}
	return "Review the workload config directly before deeper follow-up."
}

func envVarTargetServices(settingName string) []models.EnvVarTargetService {
	lowered := strings.ToLower(settingName)
	services := []models.EnvVarTargetService{}
	if lowered == "azurewebjobsstorage" {
		services = append(services, models.EnvVarTargetServiceStorage)
	}
	for _, token := range []string{"storage", "blob", "queue", "table", "share", "file", "container"} {
		if strings.Contains(lowered, token) {
			services = append(services, models.EnvVarTargetServiceStorage)
			break
		}
	}
	for _, token := range []string{"db", "database", "sql", "mysql", "postgres"} {
		if strings.Contains(lowered, token) {
			services = append(services, models.EnvVarTargetServiceDatabase)
			break
		}
	}
	return slices.Compact(services)
}

var regexpCache = map[string]*regexp.Regexp{}

func regexpMustCompile(pattern string) *regexp.Regexp {
	if compiled, ok := regexpCache[pattern]; ok {
		return compiled
	}
	compiled := regexp.MustCompile(pattern)
	regexpCache[pattern] = compiled
	return compiled
}
