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
	Tenant             string
	Subscription       string
	DevOpsOrganization string
	ChainFamily        string
	PersistenceSurface string
	Output             models.OutputMode
	RoleTrustsMode     models.RoleTrustsMode
	OutDir             string
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

type Registry struct {
	definitions map[string]Definition
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
	switch name {
	case "whoami":
		return whoAmIHandler(provider, now)
	case "inventory":
		return inventoryHandler(provider, now)
	case "automation":
		return automationHandler(provider, now)
	case "devops":
		return devopsHandler(provider, now)
	case "acr":
		return acrHandler(provider, now)
	case "databases":
		return databasesHandler(provider, now)
	case "storage":
		return storageHandler(provider, now)
	case "snapshots-disks":
		return snapshotsDisksHandler(provider, now)
	case "keyvault":
		return keyVaultHandler(provider, now)
	case "application-gateway":
		return applicationGatewayHandler(provider, now)
	case "dns":
		return dnsHandler(provider, now)
	case "aks":
		return aksHandler(provider, now)
	case "api-mgmt":
		return apiMgmtHandler(provider, now)
	case "app-credentials":
		return appCredentialsHandler(provider, now)
	case "app-services":
		return appServicesHandler(provider, now)
	case "functions":
		return functionsHandler(provider, now)
	case "webjobs":
		return webJobsHandler(provider, now)
	case "azure-ml":
		return azureMLHandler(provider, now)
	case "event-grid":
		return eventGridHandler(provider, now)
	case "logic-apps":
		return logicAppsHandler(provider, now)
	case "container-apps":
		return containerAppsHandler(provider, now)
	case "container-instances":
		return containerInstancesHandler(provider, now)
	case "arm-deployments":
		return armDeploymentsHandler(provider, now)
	case "endpoints":
		return endpointsHandler(provider, now)
	case "network-ports":
		return networkPortsHandler(provider, now)
	case "network-effective":
		return networkEffectiveHandler(provider, now)
	case "nics":
		return nicsHandler(provider, now)
	case "vms":
		return vmsHandler(provider, now)
	case "vmss":
		return vmssHandler(provider, now)
	case "workloads":
		return workloadsHandler(provider, now)
	case "rbac":
		return rbacHandler(provider, now)
	case "principals":
		return principalsHandler(provider, now)
	case "permissions":
		return permissionsHandler(provider, now)
	case "privesc":
		return privescHandler(provider, now)
	case "lighthouse":
		return lighthouseHandler(provider, now)
	case "cross-tenant":
		return crossTenantHandler(provider, now)
	case "role-trusts":
		return roleTrustsHandler(provider, now)
	case "auth-policies":
		return authPoliciesHandler(provider, now)
	case "resource-trusts":
		return resourceTrustsHandler(provider, now)
	case "managed-identities":
		return managedIdentitiesHandler(provider, now)
	case "env-vars":
		return envVarsHandler(provider, now)
	case "tokens-credentials":
		return tokensCredentialsHandler(provider, now)
	case "chains":
		return chainsHandler(provider, now)
	case "persistence":
		return persistenceHandler(provider, now)
	default:
		return nil
	}
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
