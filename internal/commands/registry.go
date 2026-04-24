package commands

import (
	"context"
	"fmt"
	"sort"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

type Request struct {
	Tenant                   string
	Subscription             string
	DevOpsOrganization       string
	ChainFamily              string
	PersistenceSurface       string
	EvasionSurface           string
	ResourceHijackingSurface string
	PathMaskingSurface       string
	Output                   models.OutputMode
	RoleTrustsMode           models.RoleTrustsMode
	OutDir                   string
}

type Response struct {
	Command  string
	Contract contracts.CommandContract
	Payload  any
}

type Handler func(context.Context, Request) (any, error)

type Definition struct {
	Contract contracts.CommandContract
	Handler  Handler
}

type handlerFactory func(providers.Provider, func() time.Time) Handler

type Registry struct {
	definitions map[string]Definition
}

var commandHandlers = map[string]handlerFactory{
	"acr":                 acrHandler,
	"aks":                 aksHandler,
	"api-mgmt":            apiMgmtHandler,
	"app-credentials":     appCredentialsHandler,
	"app-services":        appServicesHandler,
	"appinsights":         appInsightsHandler,
	"application-gateway": applicationGatewayHandler,
	"arm-deployments":     armDeploymentsHandler,
	"auth-policies":       authPoliciesHandler,
	"automation":          automationHandler,
	"azure-ml":            azureMLHandler,
	"chains":              chainsHandler,
	"container-apps":      containerAppsHandler,
	"container-apps-jobs": containerAppsJobsHandler,
	"container-instances": containerInstancesHandler,
	"cross-tenant":        crossTenantHandler,
	"databases":           databasesHandler,
	"dcr":                 dcrHandler,
	"devops":              devopsHandler,
	"diagnostic-settings": diagnosticSettingsHandler,
	"dns":                 dnsHandler,
	"endpoints":           endpointsHandler,
	"env-vars":            envVarsHandler,
	"event-grid":          eventGridHandler,
	"evasion":             evasionHandler,
	"functions":           functionsHandler,
	"inventory":           inventoryHandler,
	"keyvault":            keyVaultHandler,
	"lighthouse":          lighthouseHandler,
	"logic-apps":          logicAppsHandler,
	"managed-identities":  managedIdentitiesHandler,
	"monitoring-sinks":    monitoringSinksHandler,
	"network-effective":   networkEffectiveHandler,
	"network-ports":       networkPortsHandler,
	"nics":                nicsHandler,
	"pathmasking":         pathMaskingHandler,
	"permissions":         permissionsHandler,
	"persistence":         persistenceHandler,
	"principals":          principalsHandler,
	"privesc":             privescHandler,
	"rbac":                rbacHandler,
	"relay":               relayHandler,
	"resource-trusts":     resourceTrustsHandler,
	"resourcehijacking":   resourceHijackingHandler,
	"role-trusts":         roleTrustsHandler,
	"snapshots-disks":     snapshotsDisksHandler,
	"storage":             storageHandler,
	"tokens-credentials":  tokensCredentialsHandler,
	"vm-extensions":       vmExtensionsHandler,
	"vms":                 vmsHandler,
	"vmss":                vmssHandler,
	"webjobs":             webJobsHandler,
	"whoami":              whoAmIHandler,
	"workloads":           workloadsHandler,
}

func NewRegistry(provider providers.Provider, now func() time.Time) *Registry {
	definitions := map[string]Definition{}
	for _, name := range contracts.CommandNames() {
		contract, _ := contracts.Command(name)
		definitions[name] = Definition{
			Contract: contract,
			Handler:  handlerFor(name, provider, now),
		}
	}

	return &Registry{definitions: definitions}
}

func handlerFor(name string, provider providers.Provider, now func() time.Time) Handler {
	factory, ok := commandHandlers[name]
	if !ok {
		return nil
	}
	return factory(provider, now)
}

func (registry *Registry) Run(ctx context.Context, name string, request Request) (Response, error) {
	definition, ok := registry.definitions[name]
	if !ok {
		return Response{}, fmt.Errorf("unknown command %q", name)
	}
	if definition.Contract.Status != contracts.StatusImplemented {
		return Response{}, fmt.Errorf("command %q is not implemented yet; scaffold contract is in place for migration", name)
	}
	if definition.Handler == nil {
		return Response{}, fmt.Errorf("command %q has no handler registered", name)
	}

	payload, err := definition.Handler(ctx, request)
	if err != nil {
		return Response{}, err
	}

	return Response{
		Command:  name,
		Contract: definition.Contract,
		Payload:  payload,
	}, nil
}

func (registry *Registry) Commands() []contracts.CommandContract {
	contractsByName := make([]contracts.CommandContract, 0, len(registry.definitions))
	for _, name := range registry.CommandNames() {
		contractsByName = append(contractsByName, registry.definitions[name].Contract)
	}
	return contractsByName
}

func (registry *Registry) CommandNames() []string {
	names := make([]string, 0, len(registry.definitions))
	for name := range registry.definitions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
