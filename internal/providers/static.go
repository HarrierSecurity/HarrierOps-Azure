package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

type Provider interface {
	AKS(context.Context, string, string) (AksFacts, error)
	Acr(context.Context, string, string) (AcrFacts, error)
	AppCredentials(context.Context, string, string) (AppCredentialsFacts, error)
	Automation(context.Context, string, string) (AutomationFacts, error)
	Devops(context.Context, string, string, string) (DevopsFacts, error)
	ApplicationGateway(context.Context, string, string) (ApplicationGatewayFacts, error)
	ApiMgmt(context.Context, string, string) (ApiMgmtFacts, error)
	ArmDeployments(context.Context, string, string) (ArmDeploymentsFacts, error)
	AppServices(context.Context, string, string) (AppServicesFacts, error)
	ContainerApps(context.Context, string, string) (ContainerAppsFacts, error)
	ContainerInstances(context.Context, string, string) (ContainerInstancesFacts, error)
	Databases(context.Context, string, string) (DatabasesFacts, error)
	Endpoints(context.Context, string, string) (EndpointsFacts, error)
	EnvVars(context.Context, string, string) (EnvVarsFacts, error)
	AzureML(context.Context, string, string) (AzureMLFacts, error)
	EventGrid(context.Context, string, string) (EventGridFacts, error)
	Functions(context.Context, string, string) (FunctionsFacts, error)
	Inventory(context.Context, string, string) (InventoryFacts, error)
	KeyVault(context.Context, string, string) (KeyVaultFacts, error)
	LogicApps(context.Context, string, string) (LogicAppsFacts, error)
	ManagedIdentities(context.Context, string, string) (ManagedIdentitiesFacts, error)
	DNS(context.Context, string, string) (DNSFacts, error)
	NetworkEffective(context.Context, string, string) (NetworkEffectiveFacts, error)
	NetworkPorts(context.Context, string, string) (NetworkPortsFacts, error)
	NICs(context.Context, string, string) (NICsFacts, error)
	Permissions(context.Context, string, string) (PermissionsFacts, error)
	Principals(context.Context, string, string) (PrincipalsFacts, error)
	Privesc(context.Context, string, string) (PrivescFacts, error)
	Lighthouse(context.Context, string, string) (LighthouseFacts, error)
	CrossTenant(context.Context, string, string) (CrossTenantFacts, error)
	AuthPolicies(context.Context, string, string) (AuthPoliciesFacts, error)
	ResourceTrusts(context.Context, string, string) (ResourceTrustsFacts, error)
	RBAC(context.Context, string, string) (RBACFacts, error)
	RoleTrusts(context.Context, string, string, models.RoleTrustsMode) (RoleTrustsFacts, error)
	Storage(context.Context, string, string) (StorageFacts, error)
	SnapshotsDisks(context.Context, string, string) (SnapshotsDisksFacts, error)
	TokensCredentials(context.Context, string, string) (TokensCredentialsFacts, error)
	VMs(context.Context, string, string) (VMsFacts, error)
	VMSS(context.Context, string, string) (VMSSFacts, error)
	WebJobs(context.Context, string, string) (WebJobsFacts, error)
	Workloads(context.Context, string, string) (WorkloadsFacts, error)
	WhoAmI(context.Context, string, string) (WhoAmIFacts, error)
}

type WhoAmIFacts struct {
	TenantID        string
	Subscription    models.SubscriptionRef
	Principal       models.Principal
	EffectiveScopes []models.ScopeRef
	TokenSource     string
	AuthMode        string
	Issues          []models.Issue
}

type InventoryFacts struct {
	TenantID           string
	Subscription       models.SubscriptionRef
	ResourceGroupCount int
	ResourceCount      int
	TopResourceTypes   models.TopResourceTypes
	Issues             []models.Issue
}

type ArmDeploymentsFacts struct {
	TenantID       string
	SubscriptionID string
	Deployments    []models.ArmDeploymentSummary
	Issues         []models.Issue
}

type AutomationFacts struct {
	TenantID           string
	SubscriptionID     string
	AutomationAccounts []models.AutomationAccountAsset
	Issues             []models.Issue
}

type DevopsFacts struct {
	TenantID           string
	SubscriptionID     string
	DevOpsOrganization string
	AuthMode           string
	TokenSource        string
	Pipelines          []models.DevopsPipelineAsset
	Issues             []models.Issue
}

type AksFacts struct {
	TenantID       string
	SubscriptionID string
	AksClusters    []models.AksClusterAsset
	Issues         []models.Issue
}

type AcrFacts struct {
	TenantID       string
	SubscriptionID string
	Registries     []models.AcrRegistryAsset
	Issues         []models.Issue
}

type ApplicationGatewayFacts struct {
	TenantID            string
	SubscriptionID      string
	ApplicationGateways []models.ApplicationGatewayAsset
	Issues              []models.Issue
}

type ApiMgmtFacts struct {
	TenantID              string
	SubscriptionID        string
	ApiManagementServices []models.ApiMgmtServiceAsset
	Issues                []models.Issue
}

type DatabasesFacts struct {
	TenantID        string
	SubscriptionID  string
	DatabaseServers []models.DatabaseServerAsset
	Issues          []models.Issue
}

type KeyVaultFacts struct {
	TenantID       string
	SubscriptionID string
	KeyVaults      []models.KeyVaultAsset
	Issues         []models.Issue
}

type StorageFacts struct {
	TenantID       string
	SubscriptionID string
	StorageAssets  []models.StorageAsset
	Issues         []models.Issue
}

type ResourceTrustsFacts struct {
	TenantID       string
	SubscriptionID string
	StorageAssets  []models.StorageAsset
	KeyVaults      []models.KeyVaultAsset
	Issues         []models.Issue
}

type LighthouseFacts struct {
	TenantID              string
	SubscriptionID        string
	LighthouseDelegations []models.LighthouseDelegationAsset
	Issues                []models.Issue
}

type CrossTenantFacts struct {
	TenantID         string
	SubscriptionID   string
	CrossTenantPaths []models.CrossTenantPathSummary
	Issues           []models.Issue
}

type SnapshotsDisksFacts struct {
	TenantID           string
	SubscriptionID     string
	SnapshotDiskAssets []models.SnapshotDiskAsset
	Issues             []models.Issue
}

type DNSFacts struct {
	TenantID       string
	SubscriptionID string
	DNSZones       []models.DnsZoneAsset
	Issues         []models.Issue
}

type AppServicesFacts struct {
	TenantID       string
	SubscriptionID string
	AppServices    []models.AppServiceAsset
	Issues         []models.Issue
}

type FunctionsFacts struct {
	TenantID       string
	SubscriptionID string
	FunctionApps   []models.FunctionAppAsset
	Issues         []models.Issue
}

type WebJobsFacts struct {
	TenantID       string
	SubscriptionID string
	WebJobs        []models.WebJobAsset
	Issues         []models.Issue
}

type AzureMLFacts struct {
	TenantID       string
	SubscriptionID string
	Workspaces     []models.AzureMLWorkspaceAsset
	Issues         []models.Issue
}

type EventGridFacts struct {
	TenantID       string
	SubscriptionID string
	Routes         []models.EventGridRouteAsset
	Issues         []models.Issue
}

type LogicAppsFacts struct {
	TenantID       string
	SubscriptionID string
	Workflows      []models.LogicAppWorkflowAsset
	Issues         []models.Issue
}

type ContainerAppsFacts struct {
	TenantID       string
	SubscriptionID string
	ContainerApps  []models.ContainerAppAsset
	Issues         []models.Issue
}

type ContainerInstancesFacts struct {
	TenantID           string
	SubscriptionID     string
	ContainerInstances []models.ContainerInstanceAsset
	Issues             []models.Issue
}

type EndpointsFacts struct {
	TenantID       string
	SubscriptionID string
	Endpoints      []models.EndpointSummary
	Issues         []models.Issue
}

type NetworkPortsFacts struct {
	TenantID       string
	SubscriptionID string
	NetworkPorts   []models.NetworkPortSummary
	Issues         []models.Issue
}

type NetworkEffectiveFacts struct {
	TenantID           string
	SubscriptionID     string
	EffectiveExposures []models.NetworkEffectiveSummary
	Issues             []models.Issue
}

type NICsFacts struct {
	TenantID       string
	SubscriptionID string
	NICAssets      []models.NicAsset
	Issues         []models.Issue
}

type VMsFacts struct {
	TenantID       string
	SubscriptionID string
	VMAssets       []models.VmAsset
	Issues         []models.Issue
}

type VMSSFacts struct {
	TenantID       string
	SubscriptionID string
	VMSSAssets     []models.VmssAsset
	Issues         []models.Issue
}

type WorkloadsFacts struct {
	TenantID       string
	SubscriptionID string
	Workloads      []models.WorkloadSummary
	Issues         []models.Issue
}

type RBACFacts struct {
	TenantID        string
	Principals      []models.Principal
	Scopes          []models.ScopeRef
	RoleAssignments []models.RoleAssignment
	Issues          []models.Issue
}

type PermissionsFacts struct {
	TenantID       string
	SubscriptionID string
	Permissions    []PermissionFact
	Principals     []PermissionPrincipalFact
	Issues         []models.Issue
}

type PrincipalsFacts struct {
	TenantID       string
	SubscriptionID string
	Principals     []models.PrincipalSummary
	Issues         []models.Issue
}

type PrivescFacts struct {
	TenantID       string
	SubscriptionID string
	Paths          []models.PrivescPathSummary
	Issues         []models.Issue
}

type AuthPoliciesFacts struct {
	TenantID       string
	SubscriptionID string
	AuthPolicies   []models.AuthPolicySummary
	Issues         []models.Issue
}

type AppCredentialsFacts struct {
	TenantID       string
	SubscriptionID string
	AppCredentials []models.AppCredentialSummary
	Issues         []models.Issue
}

type PermissionFact struct {
	PrincipalID         string
	DisplayName         string
	PrincipalType       string
	HighImpactRoles     []string
	AllRoleNames        []string
	RoleAssignmentCount int
	ScopeCount          int
	ScopeIDs            []string
	Privileged          bool
	IsCurrentIdentity   bool
}

type PermissionPrincipalFact struct {
	ID            string
	Sources       []string
	IdentityNames []string
	AttachedTo    []string
}

type RoleTrustsFacts struct {
	TenantID       string
	SubscriptionID string
	Mode           models.RoleTrustsMode
	Trusts         []models.RoleTrustSummary
	Issues         []models.Issue
}

type ManagedIdentitiesFacts struct {
	TenantID        string
	SubscriptionID  string
	Identities      []models.ManagedIdentity
	RoleAssignments []models.ManagedIdentityRoleAssignment
	Findings        []models.ManagedIdentityFinding
	Issues          []models.Issue
}

type EnvVarsFacts struct {
	TenantID       string
	SubscriptionID string
	EnvVars        []models.EnvVarSummary
	Issues         []models.Issue
}

type TokensCredentialsFacts struct {
	TenantID       string
	SubscriptionID string
	Surfaces       []models.TokenCredentialSurfaceSummary
	Issues         []models.Issue
}

type StaticProvider struct{}

func NewStaticProvider() StaticProvider {
	return StaticProvider{}
}

func (StaticProvider) WhoAmI(_ context.Context, tenant string, subscription string) (WhoAmIFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	return WhoAmIFacts{
		TenantID:        session.TenantID,
		Subscription:    session.Subscription,
		Principal:       session.Principal,
		EffectiveScopes: session.EffectiveScopes,
		TokenSource:     "fixture",
		AuthMode:        "fixture",
		Issues:          []models.Issue{},
	}, nil
}

func (StaticProvider) Inventory(_ context.Context, tenant string, subscription string) (InventoryFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	return InventoryFacts{
		TenantID:           session.TenantID,
		Subscription:       session.Subscription,
		ResourceGroupCount: 4,
		ResourceCount:      30,
		TopResourceTypes: models.TopResourceTypes{
			"Microsoft.Compute/virtualMachines":          3,
			"Microsoft.ContainerService/managedClusters": 2,
			"Microsoft.Network/networkInterfaces":        6,
			"Microsoft.Storage/storageAccounts":          2,
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) ArmDeployments(_ context.Context, tenant string, subscription string) (ArmDeploymentsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return ArmDeploymentsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Deployments: []models.ArmDeploymentSummary{
			{
				Duration:            "PT1M12S",
				ID:                  "/subscriptions/" + subscriptionID + "/resourceGroups/rg-app/providers/Microsoft.Resources/deployments/app-failed",
				Mode:                "Incremental",
				Name:                "app-failed",
				OutputResourceCount: 0,
				OutputsCount:        0,
				ParametersLink:      nil,
				Providers:           []string{"Microsoft.Web"},
				ProvisioningState:   "Failed",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-app/providers/Microsoft.Resources/deployments/app-failed",
				},
				ResourceGroup: models.StringPtr("rg-app"),
				Scope:         "/subscriptions/" + subscriptionID + "/resourceGroups/rg-app",
				ScopeType:     "resource_group",
				Summary:       "resource group deployment 'app-failed' is Failed with no outputs recorded; 1 providers.",
				TemplateLink:  nil,
				Timestamp:     "2026-03-30T19:03:00Z",
			},
			{
				Duration:            "PT42S",
				ID:                  "/subscriptions/" + subscriptionID + "/providers/Microsoft.Resources/deployments/sub-foundation",
				Mode:                "Incremental",
				Name:                "sub-foundation",
				OutputResourceCount: 3,
				OutputsCount:        2,
				ParametersLink:      nil,
				Providers:           []string{"Microsoft.Network", "Microsoft.Storage"},
				ProvisioningState:   "Succeeded",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/providers/Microsoft.Resources/deployments/sub-foundation",
				},
				ResourceGroup: nil,
				Scope:         "/subscriptions/" + subscriptionID,
				ScopeType:     "subscription",
				Summary:       "subscription deployment 'sub-foundation' is Succeeded with 2 outputs; 2 providers.",
				TemplateLink:  models.StringPtr("https://example.blob.core.windows.net/templates/sub-foundation.json"),
				Timestamp:     "2026-03-30T18:42:00Z",
			},
			{
				Duration:            "PT18S",
				ID:                  "/subscriptions/" + subscriptionID + "/resourceGroups/rg-secrets/providers/Microsoft.Resources/deployments/kv-secrets",
				Mode:                "Incremental",
				Name:                "kv-secrets",
				OutputResourceCount: 1,
				OutputsCount:        1,
				ParametersLink:      models.StringPtr("https://example.blob.core.windows.net/parameters/kv-secrets.parameters.json"),
				Providers:           []string{"Microsoft.KeyVault"},
				ProvisioningState:   "Succeeded",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-secrets/providers/Microsoft.Resources/deployments/kv-secrets",
				},
				ResourceGroup: models.StringPtr("rg-secrets"),
				Scope:         "/subscriptions/" + subscriptionID + "/resourceGroups/rg-secrets",
				ScopeType:     "resource_group",
				Summary:       "resource group deployment 'kv-secrets' is Succeeded with 1 outputs; 1 providers.",
				TemplateLink:  nil,
				Timestamp:     "2026-03-30T18:50:00Z",
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) AppServices(_ context.Context, tenant string, subscription string) (AppServicesFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	aspLinuxID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/serverfarms/asp-linux"
	enabled := "Enabled"
	allAllowed := "AllAllowed"
	dotnet := "DOTNETCORE|8.0"
	running := "Running"
	systemAssigned := "SystemAssigned"
	tls10 := "1.0"
	disabled := "Disabled"
	node20 := "NODE|20-lts"
	tls12 := "1.2"
	sqlAzure := "SQLAzure"
	custom := "Custom"
	packageDisabled := false
	packageEnabled := true
	emptyMIDeployment := "run-from-package disabled"
	publicAPIDeployment := "repo github.com/contoso/customer-portal, branch main, GitHub Actions, continuous integration, run-from-package enabled"
	mainBranch := "main"
	trueValue := true
	falseValue := false
	repoURL := "https://github.com/contoso/customer-portal"
	emptyMIAppSettingsCount := 2
	emptyMIConnectionStringCount := 1
	emptyMIKeyVaultConnectionStringCount := 0
	emptyMIKeyVaultReferenceCount := 0
	emptyMISensitiveSettingCount := 1
	publicAPIAppSettingsCount := 4
	publicAPIConnectionStringCount := 2
	publicAPIKeyVaultConnectionStringCount := 1
	publicAPIKeyVaultReferenceCount := 2
	publicAPISensitiveSettingCount := 1

	return AppServicesFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		AppServices: []models.AppServiceAsset{
			{
				AppSettingsCount:              &emptyMIAppSettingsCount,
				AppServicePlanID:              &aspLinuxID,
				ClientCertEnabled:             false,
				ConnectionStringCount:         &emptyMIConnectionStringCount,
				ConnectionStringTypes:         []string{sqlAzure},
				DefaultHostname:               models.StringPtr("app-empty-mi.azurewebsites.net"),
				Deployment:                    &emptyMIDeployment,
				FTPSState:                     &allAllowed,
				HTTPSOnly:                     false,
				ID:                            "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-empty-mi",
				KeyVaultConnectionStringCount: &emptyMIKeyVaultConnectionStringCount,
				KeyVaultReferenceCount:        &emptyMIKeyVaultReferenceCount,
				Location:                      "eastus",
				MinTLSVersion:                 &tls10,
				Name:                          "app-empty-mi",
				PublicNetworkAccess:           &enabled,
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-empty-mi",
					"eeee3333-3333-3333-3333-333333333333",
					aspLinuxID,
				},
				ResourceGroup:         "rg-apps",
				RunFromPackage:        &packageDisabled,
				RuntimeStack:          &dotnet,
				SensitiveSettingCount: &emptyMISensitiveSettingCount,
				State:                 &running,
				Summary:               "App Service 'app-empty-mi' publishes hostname 'app-empty-mi.azurewebsites.net', runs runtime 'DOTNETCORE|8.0', and uses managed identity (SystemAssigned). Deployment signals: run-from-package disabled. Visible config: 2 app setting(s), 1 sensitive-looking setting name(s), 1 connection string(s), connection types SQLAzure. Visible posture: public network access Enabled, HTTPS-only disabled, TLS 1.0, FTPS AllAllowed.",
				WorkloadClientID:      models.StringPtr("ffff3333-3333-3333-3333-333333333333"),
				WorkloadIdentityIDs:   []string{},
				WorkloadIdentityType:  &systemAssigned,
				WorkloadPrincipalID:   models.StringPtr("eeee3333-3333-3333-3333-333333333333"),
			},
			{
				AppSettingsCount:              &publicAPIAppSettingsCount,
				AppServicePlanID:              &aspLinuxID,
				ClientCertEnabled:             true,
				ConnectionStringCount:         &publicAPIConnectionStringCount,
				ConnectionStringTypes:         []string{custom, sqlAzure},
				DefaultHostname:               models.StringPtr("app-public-api.azurewebsites.net"),
				Deployment:                    &publicAPIDeployment,
				DeploymentBranch:              &mainBranch,
				DeploymentIsGitHubAction:      &trueValue,
				DeploymentManualIntegration:   &falseValue,
				DeploymentRepoURL:             &repoURL,
				FTPSState:                     &disabled,
				HTTPSOnly:                     true,
				ID:                            "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
				KeyVaultConnectionStringCount: &publicAPIKeyVaultConnectionStringCount,
				KeyVaultReferenceCount:        &publicAPIKeyVaultReferenceCount,
				Location:                      "eastus",
				MinTLSVersion:                 &tls12,
				Name:                          "app-public-api",
				PublicNetworkAccess:           &enabled,
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
					"aaaa1111-1111-1111-1111-111111111111",
					aspLinuxID,
				},
				ResourceGroup:         "rg-apps",
				RunFromPackage:        &packageEnabled,
				RuntimeStack:          &node20,
				SensitiveSettingCount: &publicAPISensitiveSettingCount,
				State:                 &running,
				Summary:               "App Service 'app-public-api' publishes hostname 'app-public-api.azurewebsites.net', runs runtime 'NODE|20-lts', and uses managed identity (SystemAssigned). Deployment signals: repo github.com/contoso/customer-portal, branch main, GitHub Actions, continuous integration, run-from-package enabled. Visible config: 4 app setting(s), 2 Key Vault-backed setting(s), 1 sensitive-looking setting name(s), 2 connection string(s), 1 Key Vault-backed connection string(s), connection types Custom, SQLAzure. Visible posture: public network access Enabled, HTTPS-only enabled, TLS 1.2, FTPS Disabled.",
				WorkloadClientID:      models.StringPtr("bbbb1111-1111-1111-1111-111111111111"),
				WorkloadIdentityIDs:   []string{},
				WorkloadIdentityType:  &systemAssigned,
				WorkloadPrincipalID:   models.StringPtr("aaaa1111-1111-1111-1111-111111111111"),
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) Functions(_ context.Context, tenant string, subscription string) (FunctionsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	aspFunctionsID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/serverfarms/asp-functions"
	enabled := "Enabled"
	disabled := "Disabled"
	python311 := "PYTHON|3.11"
	functions4 := "~4"
	running := "Running"
	systemAndUserAssigned := "SystemAssigned, UserAssigned"
	plainText := "plain-text"
	tls12 := "1.2"
	alwaysOn := true
	keyVaultRefs := 1
	httpTrigger := "HTTP"
	timerTrigger := "timer"
	serviceBusTrigger := "Service Bus"
	uaOrdersPrincipalID := "cece2222-2222-2222-2222-222222222222"
	uaOrdersClientID := "dfdf2222-2222-2222-2222-222222222222"
	falseValue := false
	trueValue := true

	return FunctionsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		FunctionApps: []models.FunctionAppAsset{
			{
				AlwaysOn:                           &alwaysOn,
				AppServicePlanID:                   &aspFunctionsID,
				AzureWebJobsStorageReferenceTarget: nil,
				AzureWebJobsStorageValueType:       &plainText,
				ClientCertEnabled:                  false,
				DefaultHostname:                    models.StringPtr("func-orders.azurewebsites.net"),
				FTPSState:                          &disabled,
				FunctionsExtensionVersion:          &functions4,
				HTTPSOnly:                          true,
				ID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/" +
					"Microsoft.Web/sites/func-orders",
				KeyVaultReferenceCount: &keyVaultRefs,
				Location:               "eastus",
				MinTLSVersion:          &tls12,
				Name:                   "func-orders",
				PublicNetworkAccess:    &enabled,
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
					"cccc2222-2222-2222-2222-222222222222",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders",
					aspFunctionsID,
				},
				ResourceGroup:  "rg-apps",
				RunFromPackage: nil,
				RuntimeStack:   &python311,
				State:          &running,
				Summary:        "Function App 'func-orders' publishes hostname 'func-orders.azurewebsites.net', runs runtime 'PYTHON|3.11', targets Functions runtime '~4', and uses managed identity (SystemAssigned, UserAssigned). Deployment signals: AzureWebJobsStorage as plain-text app setting, 1 Key Vault-backed setting(s). Visible posture: public network access Enabled, HTTPS-only enabled, TLS 1.2, FTPS Disabled, Always On enabled.",
				TriggerTypes:   []string{httpTrigger, timerTrigger, serviceBusTrigger},
				VisibleFunctions: []models.FunctionChildAsset{
					{
						BindingTypes: []string{"httpTrigger", "http"},
						Bindings: []models.FunctionBinding{
							{Direction: "in", Name: "req", Type: "httpTrigger"},
							{Direction: "out", Name: "$return", Type: "http"},
						},
						Config: map[string]any{
							"bindings": []any{
								map[string]any{
									"authLevel": "function",
									"direction": "in",
									"methods":   []any{"post"},
									"name":      "req",
									"route":     "orders/webhook",
									"type":      "httpTrigger",
								},
								map[string]any{
									"direction": "out",
									"name":      "$return",
									"type":      "http",
								},
							},
						},
						ID:                "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders/functions/OrdersWebhook",
						InvokeURLTemplate: models.StringPtr("https://func-orders.azurewebsites.net/api/orders/webhook"),
						IsDisabled:        &falseValue,
						Language:          models.StringPtr("Python"),
						Name:              "OrdersWebhook",
						TriggerType:       &httpTrigger,
					},
					{
						BindingTypes: []string{"timerTrigger"},
						Bindings: []models.FunctionBinding{
							{Direction: "in", Name: "timer", Type: "timerTrigger"},
						},
						Config: map[string]any{
							"bindings": []any{
								map[string]any{
									"direction": "in",
									"name":      "timer",
									"schedule":  "0 0 * * * *",
									"type":      "timerTrigger",
								},
							},
						},
						ID:          "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders/functions/NightlyReconcile",
						IsDisabled:  &falseValue,
						Language:    models.StringPtr("Python"),
						Name:        "NightlyReconcile",
						TriggerType: &timerTrigger,
					},
					{
						BindingTypes: []string{"serviceBusTrigger", "queue"},
						Bindings: []models.FunctionBinding{
							{Direction: "in", Name: "message", Type: "serviceBusTrigger"},
							{Direction: "out", Name: "outQueue", Type: "queue"},
						},
						Config: map[string]any{
							"bindings": []any{
								map[string]any{
									"connection": "ServiceBusConnection",
									"direction":  "in",
									"name":       "message",
									"queueName":  "orders-incoming",
									"type":       "serviceBusTrigger",
								},
								map[string]any{
									"connection": "StorageConnection",
									"direction":  "out",
									"name":       "outQueue",
									"queueName":  "orders-processed",
									"type":       "queue",
								},
							},
						},
						ID:          "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders/functions/BusReconcile",
						IsDisabled:  &trueValue,
						Language:    models.StringPtr("Python"),
						Name:        "BusReconcile",
						TriggerType: &serviceBusTrigger,
					},
				},
				WorkloadClientID: models.StringPtr("dddd2222-2222-2222-2222-222222222222"),
				UserAssignedIdentities: []models.FunctionAttachedIdentity{
					{
						ID:          "/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders",
						Name:        "ua-orders",
						PrincipalID: models.StringPtr(uaOrdersPrincipalID),
						ClientID:    models.StringPtr(uaOrdersClientID),
					},
				},
				WorkloadIdentityIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders",
				},
				WorkloadIdentityType: &systemAndUserAssigned,
				WorkloadPrincipalID:  models.StringPtr("cccc2222-2222-2222-2222-222222222222"),
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) ContainerApps(_ context.Context, tenant string, subscription string) (ContainerAppsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	acaEnvProdID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/managedEnvironments/aca-env-prod"
	acaEnvInternalID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/managedEnvironments/aca-env-internal"
	trueValue := true
	falseValue := false
	ingress8080 := 8080
	autoTransport := "auto"
	httpTransport := "http"
	single := "Single"
	multiple := "Multiple"
	systemAssigned := "SystemAssigned"
	systemAndUserAssigned := "SystemAssigned, UserAssigned"

	return ContainerAppsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		ContainerApps: []models.ContainerAppAsset{
			{
				DefaultHostname:        models.StringPtr("aca-orders.wittyfield.eastus.azurecontainerapps.io"),
				EnvironmentID:          &acaEnvProdID,
				ExternalIngressEnabled: &trueValue,
				ID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/" +
					"Microsoft.App/containerApps/aca-orders",
				IngressTargetPort:       &ingress8080,
				IngressTransport:        &autoTransport,
				LatestReadyRevisionName: models.StringPtr("aca-orders--x1"),
				LatestRevisionName:      models.StringPtr("aca-orders--x1"),
				Location:                "eastus",
				Name:                    "aca-orders",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/containerApps/aca-orders",
					acaEnvProdID,
					"abab1111-1111-1111-1111-111111111111",
				},
				ResourceGroup:        "rg-containers",
				RevisionMode:         &single,
				Summary:              "Container App 'aca-orders' publishes hostname 'aca-orders.wittyfield.eastus.azurecontainerapps.io' and uses managed identity (SystemAssigned). Visible posture: external ingress enabled, target port 8080, transport auto, revision mode Single, latest ready revision aca-orders--x1.",
				WorkloadClientID:     models.StringPtr("cdcd1111-1111-1111-1111-111111111111"),
				WorkloadIdentityIDs:  []string{},
				WorkloadIdentityType: &systemAssigned,
				WorkloadPrincipalID:  models.StringPtr("abab1111-1111-1111-1111-111111111111"),
			},
			{
				DefaultHostname:        nil,
				EnvironmentID:          &acaEnvInternalID,
				ExternalIngressEnabled: &falseValue,
				ID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/" +
					"Microsoft.App/containerApps/aca-internal-jobs",
				IngressTargetPort:       &ingress8080,
				IngressTransport:        &httpTransport,
				LatestReadyRevisionName: models.StringPtr("aca-internal-jobs--x2"),
				LatestRevisionName:      models.StringPtr("aca-internal-jobs--x3"),
				Location:                "eastus",
				Name:                    "aca-internal-jobs",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/containerApps/aca-internal-jobs",
					acaEnvInternalID,
					"abab2222-2222-2222-2222-222222222222",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-container-jobs",
				},
				ResourceGroup:    "rg-containers",
				RevisionMode:     &multiple,
				Summary:          "Container App 'aca-internal-jobs' has no visible hostname from the current read path and uses managed identity (SystemAssigned, UserAssigned). Visible posture: internal ingress only, target port 8080, transport http, revision mode Multiple, latest ready revision aca-internal-jobs--x2.",
				WorkloadClientID: models.StringPtr("cdcd2222-2222-2222-2222-222222222222"),
				WorkloadIdentityIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-container-jobs",
				},
				WorkloadIdentityType: &systemAndUserAssigned,
				WorkloadPrincipalID:  models.StringPtr("abab2222-2222-2222-2222-222222222222"),
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) ContainerInstances(_ context.Context, tenant string, subscription string) (ContainerInstancesFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID
	aciSubscriptionID := "00000000-0000-0000-0000-000000000000"
	linux := "Linux"
	always := "Always"
	onFailure := "OnFailure"
	succeeded := "Succeeded"
	systemAssigned := "SystemAssigned"
	systemAndUserAssigned := "SystemAssigned, UserAssigned"
	containerCountTwo := 2
	containerCountOne := 1

	return ContainerInstancesFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		ContainerInstances: []models.ContainerInstanceAsset{
			{
				ContainerCount:    &containerCountTwo,
				ContainerImages:   []string{"mcr.microsoft.com/azuredocs/aci-helloworld:latest", "ghcr.io/harrierops/metrics-sidecar:1.0"},
				ExposedPorts:      []int{80, 443},
				FQDN:              models.StringPtr("aci-public-api.eastus.azurecontainer.io"),
				ID:                "/subscriptions/" + aciSubscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-public-api",
				Location:          "eastus",
				Name:              "aci-public-api",
				OSType:            &linux,
				ProvisioningState: &succeeded,
				PublicIPAddress:   models.StringPtr("52.160.10.30"),
				RelatedIDs: []string{
					"/subscriptions/" + aciSubscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-public-api",
					"52.160.10.30",
					"acac1111-1111-1111-1111-111111111111",
				},
				ResourceGroup:        "rg-apps",
				RestartPolicy:        &always,
				SubnetIDs:            []string{},
				Summary:              "Container group 'aci-public-api' publishes FQDN 'aci-public-api.eastus.azurecontainer.io' and uses public IP 52.160.10.30 and uses managed identity (SystemAssigned). Visible posture: os Linux, restart Always, ports 80, 443, containers 2.",
				WorkloadClientID:     models.StringPtr("acacaaaa-1111-1111-1111-111111111111"),
				WorkloadIdentityIDs:  []string{},
				WorkloadIdentityType: &systemAssigned,
				WorkloadPrincipalID:  models.StringPtr("acac1111-1111-1111-1111-111111111111"),
			},
			{
				ContainerCount:    &containerCountOne,
				ContainerImages:   []string{"ghcr.io/harrierops/internal-worker:2.0"},
				ExposedPorts:      []int{},
				FQDN:              nil,
				ID:                "/subscriptions/" + aciSubscriptionID + "/resourceGroups/rg-jobs/providers/Microsoft.ContainerInstance/containerGroups/aci-internal-worker",
				Location:          "eastus",
				Name:              "aci-internal-worker",
				OSType:            &linux,
				ProvisioningState: &succeeded,
				PublicIPAddress:   nil,
				RelatedIDs: []string{
					"/subscriptions/" + aciSubscriptionID + "/resourceGroups/rg-jobs/providers/Microsoft.ContainerInstance/containerGroups/aci-internal-worker",
					"acac2222-2222-2222-2222-222222222222",
					"/subscriptions/" + aciSubscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-aci-jobs",
					"/subscriptions/" + aciSubscriptionID + "/resourceGroups/rg-net/providers/Microsoft.Network/virtualNetworks/vnet-shared/subnets/jobs",
				},
				ResourceGroup: "rg-jobs",
				RestartPolicy: &onFailure,
				SubnetIDs: []string{
					"/subscriptions/" + aciSubscriptionID + "/resourceGroups/rg-net/providers/Microsoft.Network/virtualNetworks/vnet-shared/subnets/jobs",
				},
				Summary:          "Container group 'aci-internal-worker' has no public endpoint visible from the current read path and uses managed identity (SystemAssigned, UserAssigned). Visible posture: os Linux, restart OnFailure, subnets 1, containers 1.",
				WorkloadClientID: models.StringPtr("acacbbbb-2222-2222-2222-222222222222"),
				WorkloadIdentityIDs: []string{
					"/subscriptions/" + aciSubscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-aci-jobs",
				},
				WorkloadIdentityType: &systemAndUserAssigned,
				WorkloadPrincipalID:  models.StringPtr("acac2222-2222-2222-2222-222222222222"),
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) Endpoints(_ context.Context, tenant string, subscription string) (EndpointsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return EndpointsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Endpoints: []models.EndpointSummary{
			{
				Endpoint:       "52.160.10.30",
				EndpointType:   "ip",
				ExposureFamily: "public-ip",
				IngressPath:    "azure-container-instances-public-ip",
				RelatedIDs: []string{
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-public-api",
					"acac1111-1111-1111-1111-111111111111",
				},
				SourceAssetID:   "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-public-api",
				SourceAssetKind: "ContainerInstance",
				SourceAssetName: "aci-public-api",
				Summary:         "ContainerInstance 'aci-public-api' exposes public IP 52.160.10.30. Review the visible ingress path, ports, and runtime posture together.",
			},
			{
				Endpoint:       "52.160.10.20",
				EndpointType:   "ip",
				ExposureFamily: "public-ip",
				IngressPath:    "direct-vm-ip",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/networkInterfaces/nic-web-01",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
				},
				SourceAssetID:   "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
				SourceAssetKind: "VM",
				SourceAssetName: "vm-web-01",
				Summary:         "VM 'vm-web-01' exposes public IP 52.160.10.20. Review direct ingress path alongside NIC and NSG context.",
			},
			{
				Endpoint:       "aca-orders.wittyfield.eastus.azurecontainerapps.io",
				EndpointType:   "hostname",
				ExposureFamily: "managed-web-hostname",
				IngressPath:    "azure-container-apps-default-hostname",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/containerApps/aca-orders",
					"abab1111-1111-1111-1111-111111111111",
				},
				SourceAssetID:   "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/containerApps/aca-orders",
				SourceAssetKind: "ContainerApp",
				SourceAssetName: "aca-orders",
				Summary:         "ContainerApp 'aca-orders' publishes Azure-managed hostname 'aca-orders.wittyfield.eastus.azurecontainerapps.io'. Validate whether that ingress path is intended and how it is constrained.",
			},
			{
				Endpoint:       "aci-public-api.eastus.azurecontainer.io",
				EndpointType:   "hostname",
				ExposureFamily: "managed-container-fqdn",
				IngressPath:    "azure-container-instances-fqdn",
				RelatedIDs: []string{
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-public-api",
					"acac1111-1111-1111-1111-111111111111",
				},
				SourceAssetID:   "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-public-api",
				SourceAssetKind: "ContainerInstance",
				SourceAssetName: "aci-public-api",
				Summary:         "ContainerInstance 'aci-public-api' publishes hostname 'aci-public-api.eastus.azurecontainer.io'. Validate whether that ingress path is intended and how it is constrained.",
			},
			{
				Endpoint:       "app-empty-mi.azurewebsites.net",
				EndpointType:   "hostname",
				ExposureFamily: "managed-web-hostname",
				IngressPath:    "azurewebsites-default-hostname",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-empty-mi",
					"eeee3333-3333-3333-3333-333333333333",
				},
				SourceAssetID:   "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-empty-mi",
				SourceAssetKind: "AppService",
				SourceAssetName: "app-empty-mi",
				Summary:         "AppService 'app-empty-mi' publishes Azure-managed hostname 'app-empty-mi.azurewebsites.net'. Validate whether that ingress path is intended and how it is constrained.",
			},
			{
				Endpoint:       "app-public-api.azurewebsites.net",
				EndpointType:   "hostname",
				ExposureFamily: "managed-web-hostname",
				IngressPath:    "azurewebsites-default-hostname",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
					"aaaa1111-1111-1111-1111-111111111111",
				},
				SourceAssetID:   "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
				SourceAssetKind: "AppService",
				SourceAssetName: "app-public-api",
				Summary:         "AppService 'app-public-api' publishes Azure-managed hostname 'app-public-api.azurewebsites.net'. Validate whether that ingress path is intended and how it is constrained.",
			},
			{
				Endpoint:       "func-orders.azurewebsites.net",
				EndpointType:   "hostname",
				ExposureFamily: "managed-web-hostname",
				IngressPath:    "azure-functions-default-hostname",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
					"cccc2222-2222-2222-2222-222222222222",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders",
				},
				SourceAssetID:   "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
				SourceAssetKind: "FunctionApp",
				SourceAssetName: "func-orders",
				Summary:         "FunctionApp 'func-orders' publishes Azure-managed hostname 'func-orders.azurewebsites.net'. Validate whether that ingress path is intended and how it is constrained.",
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) NetworkPorts(_ context.Context, tenant string, subscription string) (NetworkPortsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID
	vmAssetID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01"
	nicID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/networkInterfaces/nic-web-01"
	nicNSGID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/networkSecurityGroups/nsg-web"
	subnetNSGID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/networkSecurityGroups/nsg-vnet-app"
	uaAppID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app"

	return NetworkPortsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		NetworkPorts: []models.NetworkPortSummary{
			{
				AllowSourceSummary: "Any via nic-nsg:rg-workload/nsg-web/allow-ssh-internet",
				AssetID:            vmAssetID,
				AssetName:          "vm-web-01",
				Endpoint:           "52.160.10.20",
				ExposureConfidence: "high",
				Port:               "22",
				Protocol:           "TCP",
				RelatedIDs:         []string{vmAssetID, nicID, nicNSGID, uaAppID},
				Summary:            "Asset 'vm-web-01' has inbound TCP 22 allow evidence for endpoint 52.160.10.20 from Any via nic-nsg:rg-workload/nsg-web/allow-ssh-internet.",
			},
			{
				AllowSourceSummary: "AzureLoadBalancer via subnet-nsg:rg-workload/nsg-vnet-app/allow-https-lb",
				AssetID:            vmAssetID,
				AssetName:          "vm-web-01",
				Endpoint:           "52.160.10.20",
				ExposureConfidence: "medium",
				Port:               "443",
				Protocol:           "TCP",
				RelatedIDs:         []string{vmAssetID, nicID, subnetNSGID, uaAppID},
				Summary:            "Asset 'vm-web-01' has inbound TCP 443 allow evidence for endpoint 52.160.10.20 from AzureLoadBalancer via subnet-nsg:rg-workload/nsg-vnet-app/allow-https-lb.",
			},
			{
				AllowSourceSummary: "10.20.0.0/16 via subnet-nsg:rg-workload/nsg-vnet-app/allow-app-private",
				AssetID:            vmAssetID,
				AssetName:          "vm-web-01",
				Endpoint:           "52.160.10.20",
				ExposureConfidence: "low",
				Port:               "8080",
				Protocol:           "TCP",
				RelatedIDs:         []string{vmAssetID, nicID, subnetNSGID, uaAppID},
				Summary:            "Asset 'vm-web-01' has inbound TCP 8080 allow evidence for endpoint 52.160.10.20 from 10.20.0.0/16 via subnet-nsg:rg-workload/nsg-vnet-app/allow-app-private.",
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) NetworkEffective(_ context.Context, tenant string, subscription string) (NetworkEffectiveFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return NetworkEffectiveFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		EffectiveExposures: []models.NetworkEffectiveSummary{
			{
				AssetID:              "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
				AssetName:            "vm-web-01",
				ConstrainedPorts:     []string{"TCP/443", "TCP/8080"},
				EffectiveExposure:    "high",
				Endpoint:             "52.160.10.20",
				EndpointType:         "ip",
				InternetExposedPorts: []string{"TCP/22"},
				ObservedPaths: []string{
					"Any via nic-nsg:rg-workload/nsg-web/allow-ssh-internet",
					"AzureLoadBalancer via subnet-nsg:rg-workload/nsg-vnet-app/allow-https-lb",
					"10.20.0.0/16 via subnet-nsg:rg-workload/nsg-vnet-app/allow-app-private",
				},
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/networkInterfaces/nic-web-01",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/networkSecurityGroups/nsg-web",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/networkSecurityGroups/nsg-vnet-app",
				},
				Summary: "Asset 'vm-web-01' endpoint 52.160.10.20 has internet-facing allow evidence on TCP/22 and narrower allow evidence on TCP/443, TCP/8080. Treat this as visible Azure network triage signal, not proof of full effective reachability.",
			},
			{
				AssetID:              "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-public-api",
				AssetName:            "aci-public-api",
				ConstrainedPorts:     []string{},
				EffectiveExposure:    "low",
				Endpoint:             "52.160.10.30",
				EndpointType:         "ip",
				InternetExposedPorts: []string{},
				ObservedPaths:        []string{},
				RelatedIDs: []string{
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-public-api",
					"acac1111-1111-1111-1111-111111111111",
				},
				Summary: "Asset 'aci-public-api' endpoint 52.160.10.30 is visible as a public IP path, but no inbound-rule evidence was surfaced from the current read path. Treat this as a low-confidence triage clue rather than proof of exposure.",
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) NICs(_ context.Context, tenant string, subscription string) (NICsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return NICsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		NICAssets: []models.NicAsset{
			{
				AttachedAssetID:        models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01"),
				AttachedAssetName:      models.StringPtr("vm-web-01"),
				ID:                     "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/networkInterfaces/nic-web-01",
				Name:                   "nic-web-01",
				NetworkSecurityGroupID: models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/networkSecurityGroups/nsg-web"),
				PrivateIPs:             []string{"10.0.1.4"},
				PublicIPIDs:            []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/publicIPAddresses/pip-web-01"},
				SubnetIDs:              []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/virtualNetworks/vnet-workload/subnets/vnet-app"},
				VnetIDs:                []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/virtualNetworks/vnet-workload"},
			},
			{
				AttachedAssetID:        nil,
				AttachedAssetName:      nil,
				ID:                     "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/networkInterfaces/nic-db-01",
				Name:                   "nic-db-01",
				NetworkSecurityGroupID: nil,
				PrivateIPs:             []string{"10.0.2.5", "10.0.2.6"},
				PublicIPIDs:            []string{},
				SubnetIDs:              []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/virtualNetworks/vnet-workload/subnets/vnet-db"},
				VnetIDs:                []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/virtualNetworks/vnet-workload"},
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) VMs(_ context.Context, tenant string, subscription string) (VMsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return VMsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		VMAssets: []models.VmAsset{
			{
				ID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/" +
					"Microsoft.Compute/virtualMachines/vm-web-01",
				IdentityIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/" +
						"Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
				},
				Location: "eastus",
				Name:     "vm-web-01",
				NICIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/" +
						"Microsoft.Network/networkInterfaces/nic-web-01",
				},
				PowerState:    "running",
				PrivateIPs:    []string{"10.0.1.4"},
				PublicIPs:     []string{"52.160.10.20"},
				ResourceGroup: "rg-workload",
				VMType:        "vm",
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) VMSS(_ context.Context, tenant string, subscription string) (VMSSFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	standardD4sV5 := "Standard_D4s_v5"
	uniform := "Uniform"
	rolling := "Rolling"
	systemAssigned := "SystemAssigned"
	trueValue := true
	falseValue := false
	six := 6
	standardD2sV5 := "Standard_D2s_v5"
	flexible := "Flexible"
	manual := "Manual"
	two := 2

	return VMSSFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		VMSSAssets: []models.VmssAsset{
			{
				ApplicationGatewayBackendPoolCount: 0,
				ClientID:                           models.StringPtr("77770000-0000-0000-0000-000000000002"),
				ID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/" +
					"Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01",
				IdentityIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/" +
						"Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01/identities/system",
				},
				IdentityType:                 &systemAssigned,
				InboundNATPoolCount:          1,
				InstanceCount:                &six,
				LoadBalancerBackendPoolCount: 1,
				Location:                     "eastus",
				Name:                         "vmss-edge-01",
				NICConfigurationCount:        1,
				OrchestrationMode:            &uniform,
				Overprovision:                &trueValue,
				PrincipalID:                  models.StringPtr("77770000-0000-0000-0000-000000000001"),
				PublicIPConfigurationCount:   1,
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01",
					"77770000-0000-0000-0000-000000000001",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01/identities/system",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-network/providers/Microsoft.Network/virtualNetworks/vnet-edge/subnets/app",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-network/providers/Microsoft.Network/loadBalancers/lb-edge/backendAddressPools/web",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-network/providers/Microsoft.Network/loadBalancers/lb-edge/inboundNatPools/ssh",
				},
				ResourceGroup:        "rg-workload",
				SinglePlacementGroup: &falseValue,
				SKUName:              &standardD4sV5,
				SubnetIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-network/providers/Microsoft.Network/virtualNetworks/vnet-edge/subnets/app",
				},
				Summary:     "Virtual Machine Scale Sets (VMSS) asset 'vmss-edge-01' carries SKU Standard_D4s_v5, 6 configured instance(s) and uses managed identity (SystemAssigned). Visible frontend or network cues: 1 public IP config(s), 1 inbound NAT pool ref(s), 1 LB backend pool ref(s), 1 NIC config(s), 1 subnet ref(s). Visible posture: orchestration Uniform, upgrade Rolling, single-placement-group no, overprovision yes, zone-balance yes, zones 1,2.",
				UpgradeMode: &rolling,
				ZoneBalance: &trueValue,
				Zones:       []string{"1", "2"},
			},
			{
				ApplicationGatewayBackendPoolCount: 0,
				ClientID:                           nil,
				ID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-batch/providers/" +
					"Microsoft.Compute/virtualMachineScaleSets/vmss-batch-01",
				IdentityIDs:                  []string{},
				IdentityType:                 nil,
				InboundNATPoolCount:          0,
				InstanceCount:                &two,
				LoadBalancerBackendPoolCount: 0,
				Location:                     "centralus",
				Name:                         "vmss-batch-01",
				NICConfigurationCount:        1,
				OrchestrationMode:            &flexible,
				Overprovision:                &falseValue,
				PrincipalID:                  nil,
				PublicIPConfigurationCount:   0,
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-batch/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-batch-01",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-network/providers/Microsoft.Network/virtualNetworks/vnet-batch/subnets/workers",
				},
				ResourceGroup:        "rg-batch",
				SinglePlacementGroup: &trueValue,
				SKUName:              &standardD2sV5,
				SubnetIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-network/providers/Microsoft.Network/virtualNetworks/vnet-batch/subnets/workers",
				},
				Summary:     "Virtual Machine Scale Sets (VMSS) asset 'vmss-batch-01' carries SKU Standard_D2s_v5, 2 configured instance(s) and has no managed identity visible from the current read path. Visible frontend or network cues: 1 NIC config(s), 1 subnet ref(s). Visible posture: orchestration Flexible, upgrade Manual, single-placement-group yes, overprovision no, zone-balance no.",
				UpgradeMode: &manual,
				ZoneBalance: &falseValue,
				Zones:       []string{},
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) Workloads(_ context.Context, tenant string, subscription string) (WorkloadsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	systemAssigned := "SystemAssigned"
	userAssigned := "UserAssigned"
	systemAndUserAssigned := "SystemAssigned, UserAssigned"

	return WorkloadsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Workloads: []models.WorkloadSummary{
			{
				AssetID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/" +
					"Microsoft.Compute/virtualMachines/vm-web-01",
				AssetKind:        "VM",
				AssetName:        "vm-web-01",
				Endpoints:        []string{"52.160.10.20"},
				ExposureFamilies: []string{"public-ip"},
				IdentityClientID: nil,
				IdentityIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/" +
						"Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
				},
				IdentityPrincipalID: nil,
				IdentityType:        &userAssigned,
				IngressPaths:        []string{"direct-vm-ip"},
				Location:            "eastus",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Network/networkInterfaces/nic-web-01",
				},
				ResourceGroup: "rg-workload",
				Summary:       "VM 'vm-web-01' exposes reachable endpoint '52.160.10.20' and carries managed identity context (UserAssigned). Visible signals: public-ip=1, private-ip=1, nic=1. Use this as a quick workload census pivot before deeper service-specific review.",
			},
			{
				AssetID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/" +
					"Microsoft.Web/sites/app-empty-mi",
				AssetKind:        "AppService",
				AssetName:        "app-empty-mi",
				Endpoints:        []string{"app-empty-mi.azurewebsites.net"},
				ExposureFamilies: []string{"managed-web-hostname"},
				IdentityClientID: models.StringPtr("ffff3333-3333-3333-3333-333333333333"),
				IdentityIDs:      []string{},
				IdentityPrincipalID: models.StringPtr(
					"eeee3333-3333-3333-3333-333333333333",
				),
				IdentityType: &systemAssigned,
				IngressPaths: []string{"azurewebsites-default-hostname"},
				Location:     "eastus",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-empty-mi",
					"eeee3333-3333-3333-3333-333333333333",
				},
				ResourceGroup: "rg-apps",
				Summary:       "AppService 'app-empty-mi' publishes visible endpoint hostname 'app-empty-mi.azurewebsites.net' and carries managed identity context (SystemAssigned). Visible signals: default-hostname. Use this as a quick workload census pivot before deeper service-specific review.",
			},
			{
				AssetID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/" +
					"Microsoft.Web/sites/app-public-api",
				AssetKind:        "AppService",
				AssetName:        "app-public-api",
				Endpoints:        []string{"app-public-api.azurewebsites.net"},
				ExposureFamilies: []string{"managed-web-hostname"},
				IdentityClientID: models.StringPtr("bbbb1111-1111-1111-1111-111111111111"),
				IdentityIDs:      []string{},
				IdentityPrincipalID: models.StringPtr(
					"aaaa1111-1111-1111-1111-111111111111",
				),
				IdentityType: &systemAssigned,
				IngressPaths: []string{"azurewebsites-default-hostname"},
				Location:     "eastus",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
					"aaaa1111-1111-1111-1111-111111111111",
				},
				ResourceGroup: "rg-apps",
				Summary:       "AppService 'app-public-api' publishes visible endpoint hostname 'app-public-api.azurewebsites.net' and carries managed identity context (SystemAssigned). Visible signals: default-hostname. Use this as a quick workload census pivot before deeper service-specific review.",
			},
			{
				AssetID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/" +
					"Microsoft.Web/sites/func-orders",
				AssetKind:        "FunctionApp",
				AssetName:        "func-orders",
				Endpoints:        []string{"func-orders.azurewebsites.net"},
				ExposureFamilies: []string{"managed-web-hostname"},
				IdentityClientID: models.StringPtr("dddd2222-2222-2222-2222-222222222222"),
				IdentityIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/" +
						"Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders",
				},
				IdentityPrincipalID: models.StringPtr("cccc2222-2222-2222-2222-222222222222"),
				IdentityType:        &systemAndUserAssigned,
				IngressPaths:        []string{"azure-functions-default-hostname"},
				Location:            "eastus",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
					"cccc2222-2222-2222-2222-222222222222",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders",
				},
				ResourceGroup: "rg-apps",
				Summary:       "FunctionApp 'func-orders' publishes visible endpoint hostname 'func-orders.azurewebsites.net' and carries managed identity context (SystemAssigned, UserAssigned). Visible signals: default-hostname, user-assigned=1. Use this as a quick workload census pivot before deeper service-specific review.",
			},
			{
				AssetID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/" +
					"Microsoft.App/containerApps/aca-orders",
				AssetKind:        "ContainerApp",
				AssetName:        "aca-orders",
				Endpoints:        []string{"aca-orders.wittyfield.eastus.azurecontainerapps.io"},
				ExposureFamilies: []string{"managed-web-hostname"},
				IdentityClientID: models.StringPtr("cdcd1111-1111-1111-1111-111111111111"),
				IdentityIDs:      []string{},
				IdentityPrincipalID: models.StringPtr(
					"abab1111-1111-1111-1111-111111111111",
				),
				IdentityType: &systemAssigned,
				IngressPaths: []string{"azure-container-apps-default-hostname"},
				Location:     "eastus",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/containerApps/aca-orders",
					"abab1111-1111-1111-1111-111111111111",
				},
				ResourceGroup: "rg-containers",
				Summary:       "ContainerApp 'aca-orders' publishes visible endpoint hostname 'aca-orders.wittyfield.eastus.azurecontainerapps.io' and carries managed identity context (SystemAssigned). Visible signals: default-hostname, external-ingress. Use this as a quick workload census pivot before deeper service-specific review.",
			},
			{
				AssetID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-apps/providers/" +
					"Microsoft.ContainerInstance/containerGroups/aci-public-api",
				AssetKind:        "ContainerInstance",
				AssetName:        "aci-public-api",
				Endpoints:        []string{"52.160.10.30", "aci-public-api.eastus.azurecontainer.io"},
				ExposureFamilies: []string{"public-ip", "managed-container-fqdn"},
				IdentityClientID: models.StringPtr("acacaaaa-1111-1111-1111-111111111111"),
				IdentityIDs:      []string{},
				IdentityPrincipalID: models.StringPtr(
					"acac1111-1111-1111-1111-111111111111",
				),
				IdentityType: &systemAssigned,
				IngressPaths: []string{
					"azure-container-instances-public-ip",
					"azure-container-instances-fqdn",
				},
				Location: "eastus",
				RelatedIDs: []string{
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-public-api",
					"acac1111-1111-1111-1111-111111111111",
				},
				ResourceGroup: "rg-apps",
				Summary:       "ContainerInstance 'aci-public-api' publishes 2 visible endpoint paths (52.160.10.30, aci-public-api.eastus.azurecontainer.io) and carries managed identity context (SystemAssigned). Visible signals: public-ip, fqdn, ports=2, containers=2. Use this as a quick workload census pivot before deeper service-specific review.",
			},
			{
				AssetID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/" +
					"Microsoft.App/containerApps/aca-internal-jobs",
				AssetKind:        "ContainerApp",
				AssetName:        "aca-internal-jobs",
				Endpoints:        []string{},
				ExposureFamilies: []string{},
				IdentityClientID: models.StringPtr("cdcd2222-2222-2222-2222-222222222222"),
				IdentityIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/" +
						"Microsoft.ManagedIdentity/userAssignedIdentities/ua-container-jobs",
				},
				IdentityPrincipalID: models.StringPtr("abab2222-2222-2222-2222-222222222222"),
				IdentityType:        &systemAndUserAssigned,
				IngressPaths:        []string{},
				Location:            "eastus",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/containerApps/aca-internal-jobs",
					"abab2222-2222-2222-2222-222222222222",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-container-jobs",
				},
				ResourceGroup: "rg-containers",
				Summary:       "ContainerApp 'aca-internal-jobs' has no visible endpoint path from the current read path and carries managed identity context (SystemAssigned, UserAssigned). Visible signals: internal-only, user-assigned=1. Use this as a quick workload census pivot before deeper service-specific review.",
			},
			{
				AssetID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-jobs/providers/" +
					"Microsoft.ContainerInstance/containerGroups/aci-internal-worker",
				AssetKind:        "ContainerInstance",
				AssetName:        "aci-internal-worker",
				Endpoints:        []string{},
				ExposureFamilies: []string{},
				IdentityClientID: models.StringPtr("acacbbbb-2222-2222-2222-222222222222"),
				IdentityIDs: []string{
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-aci-jobs",
				},
				IdentityPrincipalID: models.StringPtr("acac2222-2222-2222-2222-222222222222"),
				IdentityType:        &systemAndUserAssigned,
				IngressPaths:        []string{},
				Location:            "eastus",
				RelatedIDs: []string{
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-jobs/providers/Microsoft.ContainerInstance/containerGroups/aci-internal-worker",
					"acac2222-2222-2222-2222-222222222222",
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-aci-jobs",
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-net/providers/Microsoft.Network/virtualNetworks/vnet-shared/subnets/jobs",
				},
				ResourceGroup: "rg-jobs",
				Summary:       "ContainerInstance 'aci-internal-worker' has no visible endpoint path from the current read path and carries managed identity context (SystemAssigned, UserAssigned). Visible signals: subnets=1, containers=1, user-assigned=1. Use this as a quick workload census pivot before deeper service-specific review.",
			},
			{
				AssetID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/" +
					"Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01",
				AssetKind:        "VMSS",
				AssetName:        "vmss-edge-01",
				Endpoints:        []string{},
				ExposureFamilies: []string{},
				IdentityClientID: nil,
				IdentityIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/" +
						"Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01/identities/system",
				},
				IdentityPrincipalID: nil,
				IdentityType:        &systemAssigned,
				IngressPaths:        []string{},
				Location:            "eastus",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01/identities/system",
				},
				ResourceGroup: "rg-workload",
				Summary:       "VMSS 'vmss-edge-01' has no visible endpoint path from the current read path and carries managed identity context (SystemAssigned). Use this as a quick workload census pivot before deeper service-specific review.",
			},
			{
				AssetID: "/subscriptions/" + subscriptionID + "/resourceGroups/rg-batch/providers/" +
					"Microsoft.Compute/virtualMachineScaleSets/vmss-batch-01",
				AssetKind:           "VMSS",
				AssetName:           "vmss-batch-01",
				Endpoints:           []string{},
				ExposureFamilies:    []string{},
				IdentityClientID:    nil,
				IdentityIDs:         []string{},
				IdentityPrincipalID: nil,
				IdentityType:        nil,
				IngressPaths:        []string{},
				Location:            "centralus",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-batch/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-batch-01",
				},
				ResourceGroup: "rg-batch",
				Summary:       "VMSS 'vmss-batch-01' has no visible endpoint path from the current read path and has no managed identity context visible from the current read path. Use this as a quick workload census pivot before deeper service-specific review.",
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) RBAC(_ context.Context, tenant string, subscription string) (RBACFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	return RBACFacts{
		TenantID: session.TenantID,
		Principals: append([]models.Principal{
			session.Principal,
			{
				DisplayName:   "operator@lab.local",
				ID:            "44444444-4444-4444-4444-444444444444",
				PrincipalType: "User",
				TenantID:      session.TenantID,
			},
		}, append(staticLogicAppRBACPrincipals(session.TenantID), staticAzureMLRBACPrincipals(session.TenantID)...)...),
		Scopes: []models.ScopeRef{
			{
				DisplayName: session.Subscription.DisplayName,
				ID:          "/subscriptions/" + session.Subscription.ID,
				ScopeType:   "subscription",
			},
		},
		RoleAssignments: append([]models.RoleAssignment{
			{
				ID:               "ra-1",
				PrincipalID:      session.Principal.ID,
				PrincipalType:    session.Principal.PrincipalType,
				RoleDefinitionID: "rd-owner",
				RoleName:         "Owner",
				ScopeID:          "/subscriptions/" + session.Subscription.ID,
			},
			{
				ID:               "ra-2",
				PrincipalID:      "44444444-4444-4444-4444-444444444444",
				PrincipalType:    "User",
				RoleDefinitionID: "rd-reader",
				RoleName:         "Reader",
				ScopeID:          "/subscriptions/" + session.Subscription.ID,
			},
		}, append(staticLogicAppRBACRoleAssignments(session.Subscription.ID), staticAzureMLRBACRoleAssignments(session.Subscription.ID)...)...),
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) Permissions(_ context.Context, tenant string, subscription string) (PermissionsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	return PermissionsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: session.Subscription.ID,
		Permissions: append([]PermissionFact{
			{
				PrincipalID:         session.Principal.ID,
				DisplayName:         session.Principal.DisplayName,
				PrincipalType:       session.Principal.PrincipalType,
				HighImpactRoles:     []string{"Owner"},
				AllRoleNames:        []string{"Owner"},
				RoleAssignmentCount: 1,
				ScopeCount:          1,
				ScopeIDs:            []string{"/subscriptions/" + session.Subscription.ID},
				Privileged:          true,
				IsCurrentIdentity:   true,
			},
			{
				PrincipalID:         "44444444-4444-4444-4444-444444444444",
				DisplayName:         "operator@lab.local",
				PrincipalType:       "User",
				HighImpactRoles:     []string{},
				AllRoleNames:        []string{"Reader"},
				RoleAssignmentCount: 1,
				ScopeCount:          1,
				ScopeIDs:            []string{"/subscriptions/" + session.Subscription.ID},
				Privileged:          false,
				IsCurrentIdentity:   false,
			},
			{
				PrincipalID:         "66666666-6666-6666-6666-666666666666",
				DisplayName:         "build-sp",
				PrincipalType:       "ServicePrincipal",
				HighImpactRoles:     []string{"Owner"},
				AllRoleNames:        []string{"Owner"},
				RoleAssignmentCount: 2,
				ScopeCount:          2,
				ScopeIDs: []string{
					"/subscriptions/99999999-9999-9999-9999-999999999999/resourceGroups/rg-build-dr",
					"/subscriptions/99999999-9999-9999-9999-999999999999/resourceGroups/rg-identity",
				},
				Privileged:        true,
				IsCurrentIdentity: false,
			},
			{
				PrincipalID:         "12121212-1212-1212-1212-121212121212",
				DisplayName:         "aa-hybrid-prod-mi",
				PrincipalType:       "ServicePrincipal",
				HighImpactRoles:     []string{"Contributor"},
				AllRoleNames:        []string{"Contributor"},
				RoleAssignmentCount: 1,
				ScopeCount:          1,
				ScopeIDs:            []string{"/subscriptions/" + session.Subscription.ID + "/resourceGroups/rg-ops"},
				Privileged:          true,
				IsCurrentIdentity:   false,
			},
			{
				PrincipalID:         "cccc2222-2222-2222-2222-222222222222",
				DisplayName:         "func-orders-system",
				PrincipalType:       "ServicePrincipal",
				HighImpactRoles:     []string{"Contributor"},
				AllRoleNames:        []string{"Contributor"},
				RoleAssignmentCount: 1,
				ScopeCount:          1,
				ScopeIDs:            []string{"/subscriptions/" + session.Subscription.ID + "/resourceGroups/rg-apps"},
				Privileged:          true,
				IsCurrentIdentity:   false,
			},
			{
				PrincipalID:         "cece2222-2222-2222-2222-222222222222",
				DisplayName:         "ua-orders",
				PrincipalType:       "ServicePrincipal",
				HighImpactRoles:     []string{"Owner"},
				AllRoleNames:        []string{"Owner"},
				RoleAssignmentCount: 1,
				ScopeCount:          1,
				ScopeIDs:            []string{"/subscriptions/" + session.Subscription.ID},
				Privileged:          true,
				IsCurrentIdentity:   false,
			},
			{
				PrincipalID:         "eeee3333-3333-3333-3333-333333333333",
				DisplayName:         "app-empty-mi-system",
				PrincipalType:       "ServicePrincipal",
				HighImpactRoles:     []string{"Contributor"},
				AllRoleNames:        []string{"Contributor"},
				RoleAssignmentCount: 1,
				ScopeCount:          1,
				ScopeIDs:            []string{"/subscriptions/" + session.Subscription.ID + "/resourceGroups/rg-apps"},
				Privileged:          true,
				IsCurrentIdentity:   false,
			},
			{
				PrincipalID:         "abab1111-1111-1111-1111-111111111111",
				DisplayName:         "aca-orders-system",
				PrincipalType:       "ServicePrincipal",
				HighImpactRoles:     []string{"Contributor"},
				AllRoleNames:        []string{"Contributor"},
				RoleAssignmentCount: 1,
				ScopeCount:          1,
				ScopeIDs:            []string{"/subscriptions/" + session.Subscription.ID + "/resourceGroups/rg-containers"},
				Privileged:          true,
				IsCurrentIdentity:   false,
			},
			{
				PrincipalID:         "acac1111-1111-1111-1111-111111111111",
				DisplayName:         "aci-public-api-system",
				PrincipalType:       "ServicePrincipal",
				HighImpactRoles:     []string{"Contributor"},
				AllRoleNames:        []string{"Contributor"},
				RoleAssignmentCount: 1,
				ScopeCount:          1,
				ScopeIDs:            []string{"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-apps"},
				Privileged:          true,
				IsCurrentIdentity:   false,
			},
			{
				PrincipalID:         "77770000-0000-0000-0000-000000000001",
				DisplayName:         "vmss-edge-01-system",
				PrincipalType:       "ServicePrincipal",
				HighImpactRoles:     []string{"Contributor"},
				AllRoleNames:        []string{"Contributor"},
				RoleAssignmentCount: 1,
				ScopeCount:          1,
				ScopeIDs:            []string{"/subscriptions/" + session.Subscription.ID + "/resourceGroups/rg-workload"},
				Privileged:          true,
				IsCurrentIdentity:   false,
			},
		}, append(staticLogicAppPermissionFacts(session.Subscription.ID), staticAzureMLPermissionFacts(session.Subscription.ID)...)...),
		Principals: append([]PermissionPrincipalFact{
			{
				ID:            session.Principal.ID,
				Sources:       []string{"rbac", "whoami", "managed-identities"},
				IdentityNames: []string{"ua-app"},
				AttachedTo: []string{
					"/subscriptions/" + session.Subscription.ID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
				},
			},
			{
				ID:            "cccc2222-2222-2222-2222-222222222222",
				Sources:       []string{"managed-identities"},
				IdentityNames: []string{"func-orders-system"},
				AttachedTo: []string{
					"/subscriptions/" + session.Subscription.ID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
				},
			},
			{
				ID:            "cece2222-2222-2222-2222-222222222222",
				Sources:       []string{"managed-identities"},
				IdentityNames: []string{"ua-orders"},
				AttachedTo: []string{
					"/subscriptions/" + session.Subscription.ID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
				},
			},
			{
				ID:            "eeee3333-3333-3333-3333-333333333333",
				Sources:       []string{"managed-identities"},
				IdentityNames: []string{"app-empty-mi-system"},
				AttachedTo: []string{
					"/subscriptions/" + session.Subscription.ID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-empty-mi",
				},
			},
			{
				ID:            "77770000-0000-0000-0000-000000000001",
				Sources:       []string{"managed-identities"},
				IdentityNames: []string{"vmss-edge-01-system"},
				AttachedTo: []string{
					"/subscriptions/" + session.Subscription.ID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01",
				},
			},
		}, append(staticLogicAppPermissionPrincipals(session.Subscription.ID), staticAzureMLPermissionPrincipals(session.Subscription.ID)...)...),
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) RoleTrusts(_ context.Context, tenant string, subscription string, mode models.RoleTrustsMode) (RoleTrustsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	if !mode.Valid() {
		mode = models.RoleTrustsModeFast
	}
	mode = mode.Semantic()

	trusts := []models.RoleTrustSummary{
		{
			TrustType:            "federated-credential",
			SourceObjectID:       "55555555-5555-5555-5555-555555555555",
			SourceName:           models.StringPtr("build-app"),
			SourceType:           "Application",
			TargetObjectID:       "66666666-6666-6666-6666-666666666666",
			TargetName:           models.StringPtr("build-sp"),
			TargetType:           "ServicePrincipal",
			EvidenceType:         "graph-federated-credential",
			Confidence:           "confirmed",
			ControlPrimitive:     models.StringPtr("existing-federated-credential"),
			ControlledObjectType: models.StringPtr("Application"),
			ControlledObjectName: models.StringPtr("build-app"),
			EscalationMechanism:  models.StringPtr("Application 'build-app' already has federated trust that can yield service principal 'build-sp' access."),
			UsableIdentityResult: models.StringPtr("Federated sign-in can yield service principal 'build-sp' access."),
			DefenderCutPoint:     models.StringPtr("Remove or tighten the federated credential on application 'build-app'."),
			OperatorSignal:       models.StringPtr("Trust expansion visible; privilege confirmation next."),
			NextReview:           models.StringPtr("Check permissions for Azure control on service principal 'build-sp'."),
			Summary:              "Application 'build-app' trusts federated subject 'repo:TacoRocket/AzureFox:ref:refs/heads/main' from issuer 'https://token.actions.githubusercontent.com'. This row shows trust expansion into the target identity rather than direct Azure privilege by itself. Check permissions for Azure control on service principal 'build-sp'.",
			RelatedIDs:           []string{"55555555-5555-5555-5555-555555555555", "fic-build-main", "66666666-6666-6666-6666-666666666666"},
			FollowOnKind:         models.RoleTrustFollowOnPrivilegeConfirmation,
		},
		{
			TrustType:            "service-principal-owner",
			SourceObjectID:       "12121212-1212-1212-1212-121212121212",
			SourceName:           models.StringPtr("aa-hybrid-prod"),
			SourceType:           "ServicePrincipal",
			TargetObjectID:       "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			TargetName:           models.StringPtr("ops-deploy-sp"),
			TargetType:           "ServicePrincipal",
			EvidenceType:         "graph-owner",
			Confidence:           "confirmed",
			ControlPrimitive:     models.StringPtr("owner-control"),
			ControlledObjectType: models.StringPtr("ServicePrincipal"),
			ControlledObjectName: models.StringPtr("ops-deploy-sp"),
			EscalationMechanism:  models.StringPtr("Owner-level control over service principal 'ops-deploy-sp' could add or replace authentication material Azure accepts for service principal 'ops-deploy-sp'."),
			UsableIdentityResult: models.StringPtr("That could make service principal 'ops-deploy-sp' usable."),
			DefenderCutPoint:     models.StringPtr("Remove the owner-level control path over service principal 'ops-deploy-sp'."),
			OperatorSignal:       models.StringPtr("Trust expansion visible; privilege confirmation next."),
			NextReview:           models.StringPtr("Check permissions for Azure control on service principal 'ops-deploy-sp'."),
			Summary:              "Owner 'aa-hybrid-prod' can modify service principal 'ops-deploy-sp'. This row shows a service-principal takeover path rather than direct Azure privilege by itself. Check permissions for Azure control on service principal 'ops-deploy-sp'.",
			RelatedIDs:           []string{"12121212-1212-1212-1212-121212121212", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"},
			FollowOnKind:         models.RoleTrustFollowOnPrivilegeConfirmation,
		},
		{
			TrustType:            "service-principal-owner",
			SourceObjectID:       "88888888-8888-8888-8888-888888888888",
			SourceName:           models.StringPtr("automation-runner"),
			SourceType:           "ServicePrincipal",
			TargetObjectID:       "66666666-6666-6666-6666-666666666666",
			TargetName:           models.StringPtr("build-sp"),
			TargetType:           "ServicePrincipal",
			EvidenceType:         "graph-owner",
			Confidence:           "confirmed",
			ControlPrimitive:     models.StringPtr("owner-control"),
			ControlledObjectType: models.StringPtr("ServicePrincipal"),
			ControlledObjectName: models.StringPtr("build-sp"),
			EscalationMechanism:  models.StringPtr("Owner-level control over service principal 'build-sp' could add or replace authentication material Azure accepts for service principal 'build-sp'."),
			UsableIdentityResult: models.StringPtr("That could make service principal 'build-sp' usable."),
			DefenderCutPoint:     models.StringPtr("Remove the owner-level control path over service principal 'build-sp'."),
			OperatorSignal:       models.StringPtr("Trust expansion visible; privilege confirmation next."),
			NextReview:           models.StringPtr("Check permissions for Azure control on service principal 'build-sp'."),
			Summary:              "Owner 'automation-runner' can modify service principal 'build-sp'. This row shows a service-principal takeover path rather than direct Azure privilege by itself. Check permissions for Azure control on service principal 'build-sp'.",
			RelatedIDs:           []string{"88888888-8888-8888-8888-888888888888", "66666666-6666-6666-6666-666666666666"},
			FollowOnKind:         models.RoleTrustFollowOnPrivilegeConfirmation,
		},
		{
			TrustType:            "service-principal-owner",
			SourceObjectID:       session.Principal.ID,
			SourceName:           models.StringPtr(session.Principal.DisplayName),
			SourceType:           "ServicePrincipal",
			TargetObjectID:       "66666666-6666-6666-6666-666666666666",
			TargetName:           models.StringPtr("build-sp"),
			TargetType:           "ServicePrincipal",
			EvidenceType:         "graph-owner",
			Confidence:           "confirmed",
			ControlPrimitive:     models.StringPtr("owner-control"),
			ControlledObjectType: models.StringPtr("ServicePrincipal"),
			ControlledObjectName: models.StringPtr("build-sp"),
			EscalationMechanism:  models.StringPtr("Owner-level control over service principal 'build-sp' could add or replace authentication material Azure accepts for service principal 'build-sp'."),
			UsableIdentityResult: models.StringPtr("That could make service principal 'build-sp' usable."),
			DefenderCutPoint:     models.StringPtr("Remove the owner-level control path over service principal 'build-sp'."),
			OperatorSignal:       models.StringPtr("Trust expansion visible; privilege confirmation next."),
			NextReview:           models.StringPtr("Check permissions for Azure control on service principal 'build-sp'."),
			Summary:              "Owner '" + session.Principal.DisplayName + "' can modify service principal 'build-sp'. This row shows a service-principal takeover path rather than direct Azure privilege by itself. Check permissions for Azure control on service principal 'build-sp'.",
			RelatedIDs:           []string{session.Principal.ID, "66666666-6666-6666-6666-666666666666"},
			FollowOnKind:         models.RoleTrustFollowOnPrivilegeConfirmation,
		},
		{
			TrustType:                   "app-owner",
			SourceObjectID:              session.Principal.ID,
			SourceName:                  models.StringPtr(session.Principal.DisplayName),
			SourceType:                  "ServicePrincipal",
			TargetObjectID:              "55555555-5555-5555-5555-555555555555",
			TargetName:                  models.StringPtr("build-app"),
			TargetType:                  "Application",
			EvidenceType:                "graph-owner",
			Confidence:                  "confirmed",
			ControlPrimitive:            models.StringPtr("change-auth-material"),
			ControlledObjectType:        models.StringPtr("Application"),
			ControlledObjectName:        models.StringPtr("build-app"),
			BackingServicePrincipalID:   models.StringPtr("66666666-6666-6666-6666-666666666666"),
			BackingServicePrincipalName: models.StringPtr("build-sp"),
			EscalationMechanism:         models.StringPtr("Control of application 'build-app' could change authentication material that makes service principal 'build-sp' usable."),
			UsableIdentityResult:        models.StringPtr("Control of application 'build-app' could make service principal 'build-sp' usable."),
			DefenderCutPoint:            models.StringPtr("Remove the ownership path that lets the source control application 'build-app'."),
			OperatorSignal:              models.StringPtr("Indirect control visible; ownership review next."),
			NextReview:                  models.StringPtr("Review ownership around application 'build-app'; if it backs an Azure-facing identity, confirm that identity in permissions."),
			Summary:                     "Owner '" + session.Principal.DisplayName + "' can modify application 'build-app'. This is an indirect-control row: ownership is the visible trust path, not direct Azure privilege by itself. Review ownership around application 'build-app'; if it backs an Azure-facing identity, confirm that identity in permissions.",
			RelatedIDs:                  []string{session.Principal.ID, "55555555-5555-5555-5555-555555555555"},
			FollowOnKind:                models.RoleTrustFollowOnOwnershipReview,
		},
		{
			TrustType:                   "app-owner",
			SourceObjectID:              "77777777-7777-7777-7777-777777777777",
			SourceName:                  models.StringPtr("ci-admin@lab.local"),
			SourceType:                  "User",
			TargetObjectID:              "55555555-5555-5555-5555-555555555555",
			TargetName:                  models.StringPtr("build-app"),
			TargetType:                  "Application",
			EvidenceType:                "graph-owner",
			Confidence:                  "confirmed",
			ControlPrimitive:            models.StringPtr("change-auth-material"),
			ControlledObjectType:        models.StringPtr("Application"),
			ControlledObjectName:        models.StringPtr("build-app"),
			BackingServicePrincipalID:   models.StringPtr("66666666-6666-6666-6666-666666666666"),
			BackingServicePrincipalName: models.StringPtr("build-sp"),
			EscalationMechanism:         models.StringPtr("Control of application 'build-app' could change authentication material that makes service principal 'build-sp' usable."),
			UsableIdentityResult:        models.StringPtr("Control of application 'build-app' could make service principal 'build-sp' usable."),
			DefenderCutPoint:            models.StringPtr("Remove the ownership path that lets the source control application 'build-app'."),
			OperatorSignal:              models.StringPtr("Indirect control visible; ownership review next."),
			NextReview:                  models.StringPtr("Review ownership around application 'build-app'; if it backs an Azure-facing identity, confirm that identity in permissions."),
			Summary:                     "Owner 'ci-admin@lab.local' can modify application 'build-app'. This is an indirect-control row: ownership is the visible trust path, not direct Azure privilege by itself. Review ownership around application 'build-app'; if it backs an Azure-facing identity, confirm that identity in permissions.",
			RelatedIDs:                  []string{"77777777-7777-7777-7777-777777777777", "55555555-5555-5555-5555-555555555555"},
			FollowOnKind:                models.RoleTrustFollowOnOwnershipReview,
		},
		{
			TrustType:            "app-to-service-principal",
			SourceObjectID:       "99999999-9999-9999-9999-999999999999",
			SourceName:           models.StringPtr("reporting-sp"),
			SourceType:           "ServicePrincipal",
			TargetObjectID:       "00000003-0000-0000-c000-000000000000",
			TargetName:           models.StringPtr("Microsoft Graph"),
			TargetType:           "ServicePrincipal",
			EvidenceType:         "graph-app-role-assignment",
			Confidence:           "confirmed",
			ControlPrimitive:     models.StringPtr("existing-app-role-assignment"),
			ControlledObjectType: models.StringPtr("ServicePrincipal"),
			ControlledObjectName: models.StringPtr("Microsoft Graph"),
			EscalationMechanism:  models.StringPtr("Service principal 'reporting-sp' already holds an application-permission path into service principal 'Microsoft Graph'."),
			UsableIdentityResult: models.StringPtr("Service principal 'reporting-sp' already has application-permission reach to 'Microsoft Graph'."),
			DefenderCutPoint:     models.StringPtr("Remove the app-role assignment path from service principal 'reporting-sp' to 'Microsoft Graph'."),
			OperatorSignal:       models.StringPtr("Trust expansion visible; privilege confirmation next."),
			NextReview:           models.StringPtr("Check permissions for Azure control on service principal 'reporting-sp'."),
			Summary:              "Service principal 'reporting-sp' holds an application permission or app-role assignment to 'Microsoft Graph'. This row is a trust-edge and application-permission cue; confirm whether the same identity also holds Azure control. Check permissions for Azure control on service principal 'reporting-sp'.",
			RelatedIDs:           []string{"99999999-9999-9999-9999-999999999999", "app-role-graph-1", "00000003-0000-0000-c000-000000000000"},
			FollowOnKind:         models.RoleTrustFollowOnPrivilegeConfirmation,
		},
	}

	if mode == models.RoleTrustsModeFull {
		trusts = append(trusts,
			models.RoleTrustSummary{
				TrustType:            "federated-credential",
				SourceObjectID:       "12121212-1212-1212-1212-121212121212",
				SourceName:           models.StringPtr("orphan-build-app"),
				SourceType:           "Application",
				TargetObjectID:       "12121212-1212-1212-1212-121212121212",
				TargetName:           models.StringPtr("orphan-build-app"),
				TargetType:           "Application",
				EvidenceType:         "graph-federated-credential",
				Confidence:           "confirmed",
				ControlPrimitive:     models.StringPtr("existing-federated-credential"),
				ControlledObjectType: models.StringPtr("Application"),
				ControlledObjectName: models.StringPtr("orphan-build-app"),
				EscalationMechanism:  models.StringPtr("Application 'orphan-build-app' already has a federated trust path."),
				DefenderCutPoint:     models.StringPtr("Remove or tighten the federated credential on application 'orphan-build-app'."),
				OperatorSignal:       models.StringPtr("Trust expansion visible; privilege confirmation next."),
				NextReview:           models.StringPtr("Check permissions for the backing identity behind application 'orphan-build-app'."),
				Summary:              "Application 'orphan-build-app' trusts federated subject 'repo:TacoRocket/legacy-ci:environment:prod' from issuer 'https://token.actions.githubusercontent.com'. This row shows trust expansion into the target identity rather than direct Azure privilege by itself. Check permissions for the backing identity behind application 'orphan-build-app'.",
				RelatedIDs:           []string{"12121212-1212-1212-1212-121212121212", "fic-orphan-prod"},
				FollowOnKind:         models.RoleTrustFollowOnPrivilegeConfirmation,
			},
			models.RoleTrustSummary{
				TrustType:            "app-owner",
				SourceObjectID:       "13131313-1313-1313-1313-131313131313",
				SourceName:           models.StringPtr("ops-admin@lab.local"),
				SourceType:           "User",
				TargetObjectID:       "12121212-1212-1212-1212-121212121212",
				TargetName:           models.StringPtr("orphan-build-app"),
				TargetType:           "Application",
				EvidenceType:         "graph-owner",
				Confidence:           "confirmed",
				ControlPrimitive:     models.StringPtr("change-auth-material"),
				ControlledObjectType: models.StringPtr("Application"),
				ControlledObjectName: models.StringPtr("orphan-build-app"),
				EscalationMechanism:  models.StringPtr("Control of application 'orphan-build-app' could change authentication material Azure accepts for identities backed by that application."),
				DefenderCutPoint:     models.StringPtr("Remove the ownership path that lets the source control application 'orphan-build-app'."),
				OperatorSignal:       models.StringPtr("Indirect control visible; ownership review next."),
				NextReview:           models.StringPtr("Review ownership around application 'orphan-build-app'; if it backs an Azure-facing identity, confirm that identity in permissions."),
				Summary:              "Owner 'ops-admin@lab.local' can modify application 'orphan-build-app'. This is an indirect-control row: ownership is the visible trust path, not direct Azure privilege by itself. Review ownership around application 'orphan-build-app'; if it backs an Azure-facing identity, confirm that identity in permissions.",
				RelatedIDs:           []string{"13131313-1313-1313-1313-131313131313", "12121212-1212-1212-1212-121212121212"},
				FollowOnKind:         models.RoleTrustFollowOnOwnershipReview,
			},
		)
	}

	return RoleTrustsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: session.Subscription.ID,
		Mode:           mode,
		Trusts:         trusts,
		Issues:         []models.Issue{},
	}, nil
}

func (StaticProvider) ManagedIdentities(_ context.Context, tenant string, subscription string) (ManagedIdentitiesFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return ManagedIdentitiesFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Identities: append([]models.ManagedIdentity{
			{
				ID:           "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
				Name:         "ua-app",
				IdentityType: "userAssigned",
				PrincipalID:  models.StringPtr("33333333-3333-3333-3333-333333333333"),
				ClientID:     models.StringPtr("55555555-5555-5555-5555-555555555555"),
				AttachedTo: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
				},
				ScopeIDs:             []string{"/subscriptions/" + subscriptionID},
				OperatorSignal:       models.StringPtr("Public VM workload pivot; direct control visible."),
				NextReview:           models.StringPtr("Check permissions for direct control on this identity, then vms for the host context behind the workload pivot."),
				Summary:              models.StringPtr("VM 'vm-web-01' gives a public workload pivot into managed identity 'ua-app'. Current scope already shows direct control through high-impact roles (Owner). Check permissions for direct control on this identity, then vms for the host context behind the workload pivot."),
				WorkloadExposure:     models.WorkloadExposurePublic,
				DirectControlVisible: true,
			},
			{
				ID:           "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01/identities/system",
				Name:         "vmss-edge-01-system",
				IdentityType: "systemAssigned",
				PrincipalID:  models.StringPtr("77770000-0000-0000-0000-000000000001"),
				ClientID:     models.StringPtr("77770000-0000-0000-0000-000000000002"),
				AttachedTo: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01",
				},
				ScopeIDs:             []string{"/subscriptions/" + subscriptionID},
				OperatorSignal:       models.StringPtr("Exposed VMSS workload pivot; direct control not confirmed."),
				NextReview:           models.StringPtr("Check vmss for the fleet context behind this workload pivot, then permissions to confirm direct control."),
				Summary:              models.StringPtr("VMSS 'vmss-edge-01' gives a public workload pivot into managed identity 'vmss-edge-01-system'. Current scope does not confirm direct control. Check vmss for the fleet context behind this workload pivot, then permissions to confirm direct control."),
				WorkloadExposure:     models.WorkloadExposureExposed,
				DirectControlVisible: false,
			},
			{
				ID:           "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders/identities/system",
				Name:         "func-orders-system",
				IdentityType: "systemAssigned",
				PrincipalID:  models.StringPtr("cccc2222-2222-2222-2222-222222222222"),
				ClientID:     models.StringPtr("dddd2222-2222-2222-2222-222222222222"),
				AttachedTo: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
				},
				ScopeIDs:             []string{"/subscriptions/" + subscriptionID},
				OperatorSignal:       models.StringPtr("Public Function App workload pivot; direct control not confirmed."),
				NextReview:           models.StringPtr("Check env-vars for secret-bearing config on this workload, then permissions to confirm direct control."),
				Summary:              models.StringPtr("Function App 'func-orders' gives a public workload pivot into managed identity 'func-orders-system'. Current scope does not confirm direct control. Check env-vars for secret-bearing config on this workload, then permissions to confirm direct control."),
				WorkloadExposure:     models.WorkloadExposurePublic,
				DirectControlVisible: false,
			},
			{
				ID:           "/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders",
				Name:         "ua-orders",
				IdentityType: "userAssigned",
				PrincipalID:  nil,
				ClientID:     nil,
				AttachedTo: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
				},
				ScopeIDs:             []string{"/subscriptions/" + subscriptionID},
				OperatorSignal:       models.StringPtr("Public Function App workload pivot; visibility blocked."),
				NextReview:           models.StringPtr("Check env-vars for the backing workload context; current scope does not yet show direct control on this identity."),
				Summary:              models.StringPtr("Function App 'func-orders' gives a public workload pivot into managed identity 'ua-orders', but current scope does not show the backing principal cleanly. Check env-vars for the backing workload context; current scope does not yet show direct control on this identity."),
				WorkloadExposure:     models.WorkloadExposurePublic,
				DirectControlVisible: false,
			},
			{
				ID:           "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-empty-mi/identities/system",
				Name:         "app-empty-mi-system",
				IdentityType: "systemAssigned",
				PrincipalID:  models.StringPtr("eeee3333-3333-3333-3333-333333333333"),
				ClientID:     models.StringPtr("ffff3333-3333-3333-3333-333333333333"),
				AttachedTo: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-empty-mi",
				},
				ScopeIDs:             []string{"/subscriptions/" + subscriptionID},
				OperatorSignal:       models.StringPtr("Public App Service workload pivot; direct control not confirmed."),
				NextReview:           models.StringPtr("Check env-vars for secret-bearing config on this workload, then permissions to confirm direct control."),
				Summary:              models.StringPtr("App Service 'app-empty-mi' gives a public workload pivot into managed identity 'app-empty-mi-system'. Current scope does not confirm direct control. Check env-vars for secret-bearing config on this workload, then permissions to confirm direct control."),
				WorkloadExposure:     models.WorkloadExposurePublic,
				DirectControlVisible: false,
			},
			{
				ID:           "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api/identities/system",
				Name:         "app-public-api-system",
				IdentityType: "systemAssigned",
				PrincipalID:  models.StringPtr("aaaa1111-1111-1111-1111-111111111111"),
				ClientID:     models.StringPtr("bbbb1111-1111-1111-1111-111111111111"),
				AttachedTo: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
				},
				ScopeIDs:             []string{"/subscriptions/" + subscriptionID},
				OperatorSignal:       models.StringPtr("Public App Service workload pivot; direct control not confirmed."),
				NextReview:           models.StringPtr("Check env-vars for secret-bearing config on this workload, then permissions to confirm direct control."),
				Summary:              models.StringPtr("App Service 'app-public-api' gives a public workload pivot into managed identity 'app-public-api-system'. Current scope does not confirm direct control. Check env-vars for secret-bearing config on this workload, then permissions to confirm direct control."),
				WorkloadExposure:     models.WorkloadExposurePublic,
				DirectControlVisible: false,
			},
		}, append(staticLogicAppManagedIdentities(subscriptionID), staticAzureMLManagedIdentities(subscriptionID)...)...),
		RoleAssignments: append([]models.ManagedIdentityRoleAssignment{
			{
				ID:               "ra-1",
				ScopeID:          "/subscriptions/" + subscriptionID,
				PrincipalID:      "33333333-3333-3333-3333-333333333333",
				PrincipalType:    "ServicePrincipal",
				RoleDefinitionID: "rd-owner",
				RoleName:         "Owner",
			},
		}, append(staticLogicAppManagedIdentityRoleAssignments(subscriptionID), staticAzureMLManagedIdentityRoleAssignments(subscriptionID)...)...),
		Findings: append([]models.ManagedIdentityFinding{
			{
				ID:          "identity-privileged-/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
				Severity:    "high",
				Title:       "Managed identity has elevated role assignment",
				Description: "Identity 'ua-app' is assigned one or more high-impact roles (Owner).",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
					"ra-1",
				},
			},
		}, append(staticLogicAppManagedIdentityFindings(subscriptionID), staticAzureMLManagedIdentityFindings(subscriptionID)...)...),
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) EnvVars(_ context.Context, tenant string, subscription string) (EnvVarsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	appPublicAPIID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api"
	funcOrdersID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders"
	uaOrdersID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders"

	return EnvVarsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		EnvVars: []models.EnvVarSummary{
			{
				AssetID:                   appPublicAPIID,
				AssetKind:                 "AppService",
				AssetName:                 "app-public-api",
				KeyVaultReferenceIdentity: nil,
				Location:                  "eastus",
				LooksSensitive:            false,
				ReferenceTarget:           nil,
				RelatedIDs:                []string{appPublicAPIID},
				ResourceGroup:             "rg-apps",
				SettingName:               "API_BASE_URL",
				Summary:                   "AppService 'app-public-api' exposes setting 'API_BASE_URL' through management-plane app settings (plain-text). Check managed-identities for the workload token path behind this setting.",
				ValueType:                 "plain-text",
				WorkloadClientID:          models.StringPtr("bbbb1111-1111-1111-1111-111111111111"),
				WorkloadIdentityIDs:       []string{},
				WorkloadIdentityType:      models.StringPtr("SystemAssigned"),
				WorkloadPrincipalID:       models.StringPtr("aaaa1111-1111-1111-1111-111111111111"),
				TargetServices:            []models.EnvVarTargetService{},
			},
			{
				AssetID:                   appPublicAPIID,
				AssetKind:                 "AppService",
				AssetName:                 "app-public-api",
				KeyVaultReferenceIdentity: nil,
				Location:                  "eastus",
				LooksSensitive:            true,
				ReferenceTarget:           nil,
				RelatedIDs:                []string{appPublicAPIID},
				ResourceGroup:             "rg-apps",
				SettingName:               "DB_PASSWORD",
				Summary:                   "AppService 'app-public-api' stores sensitive-looking setting 'DB_PASSWORD' as plain-text app configuration. Check tokens-credentials first; this likely feeds a database credential path.",
				ValueType:                 "plain-text",
				WorkloadClientID:          models.StringPtr("bbbb1111-1111-1111-1111-111111111111"),
				WorkloadIdentityIDs:       []string{},
				WorkloadIdentityType:      models.StringPtr("SystemAssigned"),
				WorkloadPrincipalID:       models.StringPtr("aaaa1111-1111-1111-1111-111111111111"),
				TargetServices:            []models.EnvVarTargetService{models.EnvVarTargetServiceDatabase},
			},
			{
				AssetID:                   funcOrdersID,
				AssetKind:                 "FunctionApp",
				AssetName:                 "func-orders",
				KeyVaultReferenceIdentity: models.StringPtr("SystemAssigned"),
				Location:                  "eastus",
				LooksSensitive:            false,
				ReferenceTarget:           nil,
				RelatedIDs:                []string{funcOrdersID},
				ResourceGroup:             "rg-apps",
				SettingName:               "AzureWebJobsStorage",
				Summary:                   "FunctionApp 'func-orders' exposes setting 'AzureWebJobsStorage' through management-plane app settings (plain-text). Check tokens-credentials for the config-backed access path, then managed-identities for the workload token path.",
				ValueType:                 "plain-text",
				WorkloadClientID:          models.StringPtr("dddd2222-2222-2222-2222-222222222222"),
				WorkloadIdentityIDs:       []string{uaOrdersID},
				WorkloadIdentityType:      models.StringPtr("SystemAssigned, UserAssigned"),
				WorkloadPrincipalID:       models.StringPtr("cccc2222-2222-2222-2222-222222222222"),
				TargetServices:            []models.EnvVarTargetService{models.EnvVarTargetServiceStorage},
			},
			{
				AssetID:                   funcOrdersID,
				AssetKind:                 "FunctionApp",
				AssetName:                 "func-orders",
				KeyVaultReferenceIdentity: models.StringPtr("SystemAssigned"),
				Location:                  "eastus",
				LooksSensitive:            true,
				ReferenceTarget:           models.StringPtr("kvlabopen01.vault.azure.net/secrets/payment-api-key"),
				RelatedIDs:                []string{funcOrdersID},
				ResourceGroup:             "rg-apps",
				SettingName:               "PAYMENT_API_KEY",
				Summary:                   "FunctionApp 'func-orders' maps setting 'PAYMENT_API_KEY' to Key Vault-backed configuration (kvlabopen01.vault.azure.net/secrets/payment-api-key) via SystemAssigned identity. Check keyvault for the referenced secret path; review managed-identities for the workload token path.",
				ValueType:                 "keyvault-ref",
				WorkloadClientID:          models.StringPtr("dddd2222-2222-2222-2222-222222222222"),
				WorkloadIdentityIDs:       []string{uaOrdersID},
				WorkloadIdentityType:      models.StringPtr("SystemAssigned, UserAssigned"),
				WorkloadPrincipalID:       models.StringPtr("cccc2222-2222-2222-2222-222222222222"),
				TargetServices:            []models.EnvVarTargetService{},
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (StaticProvider) TokensCredentials(_ context.Context, tenant string, subscription string) (TokensCredentialsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return TokensCredentialsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Surfaces: []models.TokenCredentialSurfaceSummary{
			{
				AccessPath:     "app-setting",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
				AssetKind:      "AppService",
				AssetName:      "app-public-api",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "setting=DB_PASSWORD",
				Priority:       "high",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
					"aaaa1111-1111-1111-1111-111111111111",
				},
				ResourceGroup:  models.StringPtr("rg-apps"),
				Summary:        "AppService 'app-public-api' exposes credential-like setting 'DB_PASSWORD' as plain-text management-plane app configuration. Check env-vars for the exact setting context behind this credential clue.",
				SurfaceType:    models.TokenCredentialSurfacePlainTextSecret,
				NextReviewKind: models.TokenCredentialReviewEnvVarsSettingContext,
			},
			{
				AccessPath:     "app-setting",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
				AssetKind:      "FunctionApp",
				AssetName:      "func-orders",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "setting=AzureWebJobsStorage",
				Priority:       "high",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders",
					"cccc2222-2222-2222-2222-222222222222",
				},
				ResourceGroup:  models.StringPtr("rg-apps"),
				Summary:        "FunctionApp 'func-orders' exposes credential-like setting 'AzureWebJobsStorage' as plain-text management-plane app configuration. Check env-vars for the exact setting context behind this credential clue.",
				SurfaceType:    models.TokenCredentialSurfacePlainTextSecret,
				NextReviewKind: models.TokenCredentialReviewEnvVarsSettingContext,
			},
			{
				AccessPath:     "imds",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
				AssetKind:      "VM",
				AssetName:      "vm-web-01",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "public-ip=52.160.10.20; identities=1",
				Priority:       "high",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
				},
				ResourceGroup:     models.StringPtr("rg-workload"),
				Summary:           "VM 'vm-web-01' is publicly reachable and exposes a token minting path through IMDS for its attached managed identity. Check endpoints for the ingress path, then managed-identities and permissions for Azure control.",
				SurfaceType:       models.TokenCredentialSurfaceManagedIdentityToken,
				NextReviewKind:    models.TokenCredentialReviewEndpointsIngressAndControl,
				PubliclyReachable: true,
			},
			{
				AccessPath:     "workload-identity",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/containerApps/aca-internal-jobs",
				AssetKind:      "ContainerApp",
				AssetName:      "aca-internal-jobs",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "SystemAssigned, UserAssigned; user-assigned=1",
				Priority:       "medium",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/containerApps/aca-internal-jobs",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-container-jobs",
					"abab2222-2222-2222-2222-222222222222",
				},
				ResourceGroup:  models.StringPtr("rg-containers"),
				Summary:        "ContainerApp 'aca-internal-jobs' can request tokens through attached managed identity (SystemAssigned, UserAssigned). Check managed-identities for the identity path, then permissions for Azure control.",
				SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,
				NextReviewKind: models.TokenCredentialReviewManagedIdentityAndPermissions,
			},
			{
				AccessPath:     "workload-identity",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/containerApps/aca-orders",
				AssetKind:      "ContainerApp",
				AssetName:      "aca-orders",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "SystemAssigned",
				Priority:       "medium",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-containers/providers/Microsoft.App/containerApps/aca-orders",
					"abab1111-1111-1111-1111-111111111111",
				},
				ResourceGroup:  models.StringPtr("rg-containers"),
				Summary:        "ContainerApp 'aca-orders' can request tokens through attached managed identity (SystemAssigned). Check managed-identities for the identity path, then permissions for Azure control.",
				SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,
				NextReviewKind: models.TokenCredentialReviewManagedIdentityAndPermissions,
			},
			{
				AccessPath:     "workload-identity",
				AssetID:        "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-jobs/providers/Microsoft.ContainerInstance/containerGroups/aci-internal-worker",
				AssetKind:      "ContainerInstance",
				AssetName:      "aci-internal-worker",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "SystemAssigned, UserAssigned; user-assigned=1",
				Priority:       "medium",
				RelatedIDs: []string{
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-jobs/providers/Microsoft.ContainerInstance/containerGroups/aci-internal-worker",
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-aci-jobs",
					"acac2222-2222-2222-2222-222222222222",
				},
				ResourceGroup:  models.StringPtr("rg-jobs"),
				Summary:        "ContainerInstance 'aci-internal-worker' can request tokens through attached managed identity (SystemAssigned, UserAssigned). Check managed-identities for the identity path, then permissions for Azure control.",
				SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,
				NextReviewKind: models.TokenCredentialReviewManagedIdentityAndPermissions,
			},
			{
				AccessPath:     "workload-identity",
				AssetID:        "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-public-api",
				AssetKind:      "ContainerInstance",
				AssetName:      "aci-public-api",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "SystemAssigned",
				Priority:       "medium",
				RelatedIDs: []string{
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-apps/providers/Microsoft.ContainerInstance/containerGroups/aci-public-api",
					"acac1111-1111-1111-1111-111111111111",
				},
				ResourceGroup:  models.StringPtr("rg-apps"),
				Summary:        "ContainerInstance 'aci-public-api' can request tokens through attached managed identity (SystemAssigned). Check managed-identities for the identity path, then permissions for Azure control.",
				SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,
				NextReviewKind: models.TokenCredentialReviewManagedIdentityAndPermissions,
			},
			{
				AccessPath:     "workload-identity",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-empty-mi",
				AssetKind:      "AppService",
				AssetName:      "app-empty-mi",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "SystemAssigned",
				Priority:       "medium",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-empty-mi",
					"eeee3333-3333-3333-3333-333333333333",
				},
				ResourceGroup:  models.StringPtr("rg-apps"),
				Summary:        "AppService 'app-empty-mi' can request tokens through attached managed identity (SystemAssigned). Check managed-identities for the identity path, then permissions for Azure control.",
				SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,
				NextReviewKind: models.TokenCredentialReviewManagedIdentityAndPermissions,
			},
			{
				AccessPath:     "workload-identity",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
				AssetKind:      "AppService",
				AssetName:      "app-public-api",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "SystemAssigned",
				Priority:       "medium",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
					"aaaa1111-1111-1111-1111-111111111111",
				},
				ResourceGroup:  models.StringPtr("rg-apps"),
				Summary:        "AppService 'app-public-api' can request tokens through attached managed identity (SystemAssigned). Check managed-identities for the identity path, then permissions for Azure control.",
				SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,
				NextReviewKind: models.TokenCredentialReviewManagedIdentityAndPermissions,
			},
			{
				AccessPath:     "app-setting",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
				AssetKind:      "FunctionApp",
				AssetName:      "func-orders",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "target=kvlabopen01.vault.azure.net/secrets/payment-api-key; identity=SystemAssigned",
				Priority:       "medium",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders",
					"cccc2222-2222-2222-2222-222222222222",
				},
				ResourceGroup:  models.StringPtr("rg-apps"),
				Summary:        "FunctionApp 'func-orders' uses setting 'PAYMENT_API_KEY' to reach Key Vault-backed secret material (kvlabopen01.vault.azure.net/secrets/payment-api-key) via SystemAssigned. Check keyvault for the referenced secret boundary, then managed-identities for the backing workload identity.",
				SurfaceType:    models.TokenCredentialSurfaceKeyVaultReference,
				NextReviewKind: models.TokenCredentialReviewKeyVaultAndManagedIdentity,
			},
			{
				AccessPath:     "workload-identity",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
				AssetKind:      "FunctionApp",
				AssetName:      "func-orders",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "SystemAssigned, UserAssigned; user-assigned=1",
				Priority:       "medium",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-identities/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-orders",
					"cccc2222-2222-2222-2222-222222222222",
				},
				ResourceGroup:  models.StringPtr("rg-apps"),
				Summary:        "FunctionApp 'func-orders' can request tokens through attached managed identity (SystemAssigned, UserAssigned). Check managed-identities for the identity path, then permissions for Azure control.",
				SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,
				NextReviewKind: models.TokenCredentialReviewManagedIdentityAndPermissions,
			},
			{
				AccessPath:     "deployment-history",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-secrets/providers/Microsoft.Resources/deployments/kv-secrets",
				AssetKind:      "ArmDeployment",
				AssetName:      "kv-secrets",
				Location:       nil,
				OperatorSignal: "outputs=1; providers=1",
				Priority:       "medium",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-secrets/providers/Microsoft.Resources/deployments/kv-secrets",
				},
				ResourceGroup:  models.StringPtr("rg-secrets"),
				Summary:        "Deployment 'kv-secrets' recorded 1 output values in deployment history. Check arm-deployments for the exact output context behind this credential clue.",
				SurfaceType:    models.TokenCredentialSurfaceDeploymentOutput,
				NextReviewKind: models.TokenCredentialReviewARMDeploymentOutputs,
			},
			{
				AccessPath:     "deployment-history",
				AssetID:        "/subscriptions/" + subscriptionID + "/providers/Microsoft.Resources/deployments/sub-foundation",
				AssetKind:      "ArmDeployment",
				AssetName:      "sub-foundation",
				Location:       nil,
				OperatorSignal: "outputs=2; providers=2",
				Priority:       "medium",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/providers/Microsoft.Resources/deployments/sub-foundation",
				},
				ResourceGroup:  nil,
				Summary:        "Deployment 'sub-foundation' recorded 2 output values in deployment history. Check arm-deployments for the exact output context behind this credential clue.",
				SurfaceType:    models.TokenCredentialSurfaceDeploymentOutput,
				NextReviewKind: models.TokenCredentialReviewARMDeploymentOutputs,
			},
			{
				AccessPath:     "imds",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01",
				AssetKind:      "VMSS",
				AssetName:      "vmss-edge-01",
				Location:       models.StringPtr("eastus"),
				OperatorSignal: "public-ip=none; identities=1",
				Priority:       "medium",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01/identities/system",
				},
				ResourceGroup:  models.StringPtr("rg-workload"),
				Summary:        "VMSS 'vmss-edge-01' exposes a token minting path through IMDS for its attached managed identity. Check managed-identities for the identity path, then permissions for Azure control.",
				SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,
				NextReviewKind: models.TokenCredentialReviewManagedIdentityAndPermissions,
			},
			{
				AccessPath:     "deployment-history",
				AssetID:        "/subscriptions/" + subscriptionID + "/resourceGroups/rg-secrets/providers/Microsoft.Resources/deployments/kv-secrets",
				AssetKind:      "ArmDeployment",
				AssetName:      "kv-secrets",
				Location:       nil,
				OperatorSignal: "parameters=example.blob.core.windows.net/parameters/kv-secrets.parameters.json",
				Priority:       "low",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-secrets/providers/Microsoft.Resources/deployments/kv-secrets",
				},
				ResourceGroup:  models.StringPtr("rg-secrets"),
				Summary:        "Deployment 'kv-secrets' references remote template or parameter content that may expose reusable configuration or credential context. Check arm-deployments for the linked template or parameter path behind this credential clue.",
				SurfaceType:    models.TokenCredentialSurfaceLinkedDeploymentAsset,
				NextReviewKind: models.TokenCredentialReviewARMDeploymentLinks,
			},
			{
				AccessPath:     "deployment-history",
				AssetID:        "/subscriptions/" + subscriptionID + "/providers/Microsoft.Resources/deployments/sub-foundation",
				AssetKind:      "ArmDeployment",
				AssetName:      "sub-foundation",
				Location:       nil,
				OperatorSignal: "template=example.blob.core.windows.net/templates/sub-foundation.json",
				Priority:       "low",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/providers/Microsoft.Resources/deployments/sub-foundation",
				},
				ResourceGroup:  nil,
				Summary:        "Deployment 'sub-foundation' references remote template or parameter content that may expose reusable configuration or credential context. Check arm-deployments for the linked template or parameter path behind this credential clue.",
				SurfaceType:    models.TokenCredentialSurfaceLinkedDeploymentAsset,
				NextReviewKind: models.TokenCredentialReviewARMDeploymentLinks,
			},
		},
		Issues: []models.Issue{},
	}, nil
}

type fixtureSession struct {
	TenantID        string
	Subscription    models.SubscriptionRef
	Principal       models.Principal
	EffectiveScopes []models.ScopeRef
}

func staticFixtureSession(tenant string, subscription string) fixtureSession {
	if tenant == "" {
		tenant = "11111111-1111-1111-1111-111111111111"
	}
	if subscription == "" {
		subscription = "22222222-2222-2222-2222-222222222222"
	}

	subscriptionRef := models.SubscriptionRef{
		ID:          subscription,
		DisplayName: "azurefox-lab-sub",
		State:       "Enabled",
	}

	return fixtureSession{
		TenantID:     tenant,
		Subscription: subscriptionRef,
		Principal: models.Principal{
			ID:            "33333333-3333-3333-3333-333333333333",
			PrincipalType: "ServicePrincipal",
			DisplayName:   "azurefox-lab-sp",
			TenantID:      tenant,
		},
		EffectiveScopes: []models.ScopeRef{
			{
				ID:          "/subscriptions/" + subscription,
				ScopeType:   "subscription",
				DisplayName: subscriptionRef.DisplayName,
			},
		},
	}
}
