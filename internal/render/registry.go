package render

import "fmt"

type rendererEntry struct {
	table func(any) (string, error)
	csv   func(any) (string, error)
}

func wrapTableRenderer[T any](command string, render func(T) string) func(any) (string, error) {
	return func(payload any) (string, error) {
		out, ok := payload.(T)
		if !ok {
			return "", fmt.Errorf("unexpected payload type for %s: %T", command, payload)
		}
		return render(out), nil
	}
}

func wrapCSVRenderer[T any](command string, render func(T) (string, error)) func(any) (string, error) {
	return func(payload any) (string, error) {
		out, ok := payload.(T)
		if !ok {
			return "", fmt.Errorf("unexpected payload type for %s: %T", command, payload)
		}
		return render(out)
	}
}

func renderRegistryEntry(command string) (rendererEntry, error) {
	switch command {
	case "automation":
		return rendererEntry{table: wrapTableRenderer("automation", automationTable), csv: wrapCSVRenderer("automation", automationCSV)}, nil
	case "devops":
		return rendererEntry{table: wrapTableRenderer("devops", devopsTable), csv: wrapCSVRenderer("devops", devopsCSV)}, nil
	case "acr":
		return rendererEntry{table: wrapTableRenderer("acr", acrTable), csv: wrapCSVRenderer("acr", acrCSV)}, nil
	case "databases":
		return rendererEntry{table: wrapTableRenderer("databases", databasesTable), csv: wrapCSVRenderer("databases", databasesCSV)}, nil
	case "storage":
		return rendererEntry{table: wrapTableRenderer("storage", storageTable), csv: wrapCSVRenderer("storage", storageCSV)}, nil
	case "snapshots-disks":
		return rendererEntry{table: wrapTableRenderer("snapshots-disks", snapshotsDisksTable), csv: wrapCSVRenderer("snapshots-disks", snapshotsDisksCSV)}, nil
	case "keyvault":
		return rendererEntry{table: wrapTableRenderer("keyvault", keyVaultTable), csv: wrapCSVRenderer("keyvault", keyVaultCSV)}, nil
	case "application-gateway":
		return rendererEntry{table: wrapTableRenderer("application-gateway", applicationGatewayTable), csv: wrapCSVRenderer("application-gateway", applicationGatewayCSV)}, nil
	case "dns":
		return rendererEntry{table: wrapTableRenderer("dns", dnsTable), csv: wrapCSVRenderer("dns", dnsCSV)}, nil
	case "aks":
		return rendererEntry{table: wrapTableRenderer("aks", aksTable), csv: wrapCSVRenderer("aks", aksCSV)}, nil
	case "api-mgmt":
		return rendererEntry{table: wrapTableRenderer("api-mgmt", apiMgmtTable), csv: wrapCSVRenderer("api-mgmt", apiMgmtCSV)}, nil
	case "app-credentials":
		return rendererEntry{table: wrapTableRenderer("app-credentials", appCredentialsTable), csv: wrapCSVRenderer("app-credentials", appCredentialsCSV)}, nil
	case "app-services":
		return rendererEntry{table: wrapTableRenderer("app-services", appServicesTable), csv: wrapCSVRenderer("app-services", appServicesCSV)}, nil
	case "functions":
		return rendererEntry{table: wrapTableRenderer("functions", functionsTable), csv: wrapCSVRenderer("functions", functionsCSV)}, nil
	case "webjobs":
		return rendererEntry{table: wrapTableRenderer("webjobs", webJobsTable), csv: wrapCSVRenderer("webjobs", webJobsCSV)}, nil
	case "azure-ml":
		return rendererEntry{table: wrapTableRenderer("azure-ml", azureMLTable), csv: wrapCSVRenderer("azure-ml", azureMLCSV)}, nil
	case "event-grid":
		return rendererEntry{table: wrapTableRenderer("event-grid", eventGridTable), csv: wrapCSVRenderer("event-grid", eventGridCSV)}, nil
	case "logic-apps":
		return rendererEntry{table: wrapTableRenderer("logic-apps", logicAppsTable), csv: wrapCSVRenderer("logic-apps", logicAppsCSV)}, nil
	case "container-apps":
		return rendererEntry{table: wrapTableRenderer("container-apps", containerAppsTable), csv: wrapCSVRenderer("container-apps", containerAppsCSV)}, nil
	case "container-instances":
		return rendererEntry{table: wrapTableRenderer("container-instances", containerInstancesTable), csv: wrapCSVRenderer("container-instances", containerInstancesCSV)}, nil
	case "arm-deployments":
		return rendererEntry{table: wrapTableRenderer("arm-deployments", armDeploymentsTable), csv: wrapCSVRenderer("arm-deployments", armDeploymentsCSV)}, nil
	case "endpoints":
		return rendererEntry{table: wrapTableRenderer("endpoints", endpointsTable), csv: wrapCSVRenderer("endpoints", endpointsCSV)}, nil
	case "network-ports":
		return rendererEntry{table: wrapTableRenderer("network-ports", networkPortsTable), csv: wrapCSVRenderer("network-ports", networkPortsCSV)}, nil
	case "network-effective":
		return rendererEntry{table: wrapTableRenderer("network-effective", networkEffectiveTable), csv: wrapCSVRenderer("network-effective", networkEffectiveCSV)}, nil
	case "nics":
		return rendererEntry{table: wrapTableRenderer("nics", nicsTable), csv: wrapCSVRenderer("nics", nicsCSV)}, nil
	case "vms":
		return rendererEntry{table: wrapTableRenderer("vms", vmsTable), csv: wrapCSVRenderer("vms", vmsCSV)}, nil
	case "vmss":
		return rendererEntry{table: wrapTableRenderer("vmss", vmssTable), csv: wrapCSVRenderer("vmss", vmssCSV)}, nil
	case "workloads":
		return rendererEntry{table: wrapTableRenderer("workloads", workloadsTable), csv: wrapCSVRenderer("workloads", workloadsCSV)}, nil
	case "principals":
		return rendererEntry{table: wrapTableRenderer("principals", principalsTable), csv: wrapCSVRenderer("principals", principalsCSV)}, nil
	case "permissions":
		return rendererEntry{table: wrapTableRenderer("permissions", permissionsTable), csv: wrapCSVRenderer("permissions", permissionsCSV)}, nil
	case "privesc":
		return rendererEntry{table: wrapTableRenderer("privesc", privescTable), csv: wrapCSVRenderer("privesc", privescCSV)}, nil
	case "lighthouse":
		return rendererEntry{table: wrapTableRenderer("lighthouse", lighthouseTable), csv: wrapCSVRenderer("lighthouse", lighthouseCSV)}, nil
	case "cross-tenant":
		return rendererEntry{table: wrapTableRenderer("cross-tenant", crossTenantTable), csv: wrapCSVRenderer("cross-tenant", crossTenantCSV)}, nil
	case "auth-policies":
		return rendererEntry{table: wrapTableRenderer("auth-policies", authPoliciesTable), csv: wrapCSVRenderer("auth-policies", authPoliciesCSV)}, nil
	case "resource-trusts":
		return rendererEntry{table: wrapTableRenderer("resource-trusts", resourceTrustsTable), csv: wrapCSVRenderer("resource-trusts", resourceTrustsCSV)}, nil
	case "rbac":
		return rendererEntry{table: wrapTableRenderer("rbac", rbacTable), csv: wrapCSVRenderer("rbac", rbacCSV)}, nil
	case "managed-identities":
		return rendererEntry{table: wrapTableRenderer("managed-identities", managedIdentitiesTable), csv: wrapCSVRenderer("managed-identities", managedIdentitiesCSV)}, nil
	case "env-vars":
		return rendererEntry{table: wrapTableRenderer("env-vars", envVarsTable), csv: wrapCSVRenderer("env-vars", envVarsCSV)}, nil
	case "tokens-credentials":
		return rendererEntry{table: wrapTableRenderer("tokens-credentials", tokensCredentialsTable), csv: wrapCSVRenderer("tokens-credentials", tokensCredentialsCSV)}, nil
	case "chains":
		return rendererEntry{table: chainsTableRenderer, csv: chainsCSVRenderer}, nil
	case "persistence":
		return rendererEntry{table: persistenceTableRenderer, csv: persistenceCSVRenderer}, nil
	case "role-trusts":
		return rendererEntry{table: wrapTableRenderer("role-trusts", roleTrustsTable), csv: wrapCSVRenderer("role-trusts", roleTrustsCSV)}, nil
	case "inventory":
		return rendererEntry{table: wrapTableRenderer("inventory", inventoryTable), csv: wrapCSVRenderer("inventory", inventoryCSV)}, nil
	case "whoami":
		return rendererEntry{table: wrapTableRenderer("whoami", whoAmITable), csv: wrapCSVRenderer("whoami", whoAmICSV)}, nil
	default:
		return rendererEntry{}, fmt.Errorf("rendering is not implemented for command %q", command)
	}
}
