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

var renderRegistry = map[string]rendererEntry{
	"acr":                 {table: wrapTableRenderer("acr", acrTable), csv: wrapCSVRenderer("acr", acrCSV)},
	"aks":                 {table: wrapTableRenderer("aks", aksTable), csv: wrapCSVRenderer("aks", aksCSV)},
	"api-mgmt":            {table: wrapTableRenderer("api-mgmt", apiMgmtTable), csv: wrapCSVRenderer("api-mgmt", apiMgmtCSV)},
	"app-credentials":     {table: wrapTableRenderer("app-credentials", appCredentialsTable), csv: wrapCSVRenderer("app-credentials", appCredentialsCSV)},
	"app-services":        {table: wrapTableRenderer("app-services", appServicesTable), csv: wrapCSVRenderer("app-services", appServicesCSV)},
	"appinsights":         {table: wrapTableRenderer("appinsights", appInsightsTable), csv: wrapCSVRenderer("appinsights", appInsightsCSV)},
	"application-gateway": {table: wrapTableRenderer("application-gateway", applicationGatewayTable), csv: wrapCSVRenderer("application-gateway", applicationGatewayCSV)},
	"arm-deployments":     {table: wrapTableRenderer("arm-deployments", armDeploymentsTable), csv: wrapCSVRenderer("arm-deployments", armDeploymentsCSV)},
	"auth-policies":       {table: wrapTableRenderer("auth-policies", authPoliciesTable), csv: wrapCSVRenderer("auth-policies", authPoliciesCSV)},
	"automation":          {table: wrapTableRenderer("automation", automationTable), csv: wrapCSVRenderer("automation", automationCSV)},
	"azure-ml":            {table: wrapTableRenderer("azure-ml", azureMLTable), csv: wrapCSVRenderer("azure-ml", azureMLCSV)},
	"chains":              {table: chainsTableRenderer, csv: chainsCSVRenderer},
	"container-apps":      {table: wrapTableRenderer("container-apps", containerAppsTable), csv: wrapCSVRenderer("container-apps", containerAppsCSV)},
	"container-apps-jobs": {table: wrapTableRenderer("container-apps-jobs", containerAppsJobsTable), csv: wrapCSVRenderer("container-apps-jobs", containerAppsJobsCSV)},
	"container-instances": {table: wrapTableRenderer("container-instances", containerInstancesTable), csv: wrapCSVRenderer("container-instances", containerInstancesCSV)},
	"cross-tenant":        {table: wrapTableRenderer("cross-tenant", crossTenantTable), csv: wrapCSVRenderer("cross-tenant", crossTenantCSV)},
	"databases":           {table: wrapTableRenderer("databases", databasesTable), csv: wrapCSVRenderer("databases", databasesCSV)},
	"dcr":                 {table: wrapTableRenderer("dcr", dcrTable), csv: wrapCSVRenderer("dcr", dcrCSV)},
	"devops":              {table: wrapTableRenderer("devops", devopsTable), csv: wrapCSVRenderer("devops", devopsCSV)},
	"diagnostic-settings": {table: wrapTableRenderer("diagnostic-settings", diagnosticSettingsTable), csv: wrapCSVRenderer("diagnostic-settings", diagnosticSettingsCSV)},
	"dns":                 {table: wrapTableRenderer("dns", dnsTable), csv: wrapCSVRenderer("dns", dnsCSV)},
	"endpoints":           {table: wrapTableRenderer("endpoints", endpointsTable), csv: wrapCSVRenderer("endpoints", endpointsCSV)},
	"env-vars":            {table: wrapTableRenderer("env-vars", envVarsTable), csv: wrapCSVRenderer("env-vars", envVarsCSV)},
	"event-grid":          {table: wrapTableRenderer("event-grid", eventGridTable), csv: wrapCSVRenderer("event-grid", eventGridCSV)},
	"evasion":             {table: evasionTableRenderer, csv: evasionCSVRenderer},
	"functions":           {table: wrapTableRenderer("functions", functionsTable), csv: wrapCSVRenderer("functions", functionsCSV)},
	"inventory":           {table: wrapTableRenderer("inventory", inventoryTable), csv: wrapCSVRenderer("inventory", inventoryCSV)},
	"keyvault":            {table: wrapTableRenderer("keyvault", keyVaultTable), csv: wrapCSVRenderer("keyvault", keyVaultCSV)},
	"lighthouse":          {table: wrapTableRenderer("lighthouse", lighthouseTable), csv: wrapCSVRenderer("lighthouse", lighthouseCSV)},
	"logic-apps":          {table: wrapTableRenderer("logic-apps", logicAppsTable), csv: wrapCSVRenderer("logic-apps", logicAppsCSV)},
	"managed-identities":  {table: wrapTableRenderer("managed-identities", managedIdentitiesTable), csv: wrapCSVRenderer("managed-identities", managedIdentitiesCSV)},
	"monitoring-sinks":    {table: wrapTableRenderer("monitoring-sinks", monitoringSinksTable), csv: wrapCSVRenderer("monitoring-sinks", monitoringSinksCSV)},
	"network-effective":   {table: wrapTableRenderer("network-effective", networkEffectiveTable), csv: wrapCSVRenderer("network-effective", networkEffectiveCSV)},
	"network-ports":       {table: wrapTableRenderer("network-ports", networkPortsTable), csv: wrapCSVRenderer("network-ports", networkPortsCSV)},
	"nics":                {table: wrapTableRenderer("nics", nicsTable), csv: wrapCSVRenderer("nics", nicsCSV)},
	"pathmasking":         {table: pathMaskingTableRenderer, csv: pathMaskingCSVRenderer},
	"permissions":         {table: wrapTableRenderer("permissions", permissionsTable), csv: wrapCSVRenderer("permissions", permissionsCSV)},
	"persistence":         {table: persistenceTableRenderer, csv: persistenceCSVRenderer},
	"principals":          {table: wrapTableRenderer("principals", principalsTable), csv: wrapCSVRenderer("principals", principalsCSV)},
	"privesc":             {table: wrapTableRenderer("privesc", privescTable), csv: wrapCSVRenderer("privesc", privescCSV)},
	"rbac":                {table: wrapTableRenderer("rbac", rbacTable), csv: wrapCSVRenderer("rbac", rbacCSV)},
	"relay":               {table: wrapTableRenderer("relay", relayTable), csv: wrapCSVRenderer("relay", relayCSV)},
	"resource-trusts":     {table: wrapTableRenderer("resource-trusts", resourceTrustsTable), csv: wrapCSVRenderer("resource-trusts", resourceTrustsCSV)},
	"resourcehijacking":   {table: resourceHijackingTableRenderer, csv: resourceHijackingCSVRenderer},
	"role-trusts":         {table: wrapTableRenderer("role-trusts", roleTrustsTable), csv: wrapCSVRenderer("role-trusts", roleTrustsCSV)},
	"snapshots-disks":     {table: wrapTableRenderer("snapshots-disks", snapshotsDisksTable), csv: wrapCSVRenderer("snapshots-disks", snapshotsDisksCSV)},
	"storage":             {table: wrapTableRenderer("storage", storageTable), csv: wrapCSVRenderer("storage", storageCSV)},
	"tokens-credentials":  {table: wrapTableRenderer("tokens-credentials", tokensCredentialsTable), csv: wrapCSVRenderer("tokens-credentials", tokensCredentialsCSV)},
	"vm-extensions":       {table: wrapTableRenderer("vm-extensions", vmExtensionsTable), csv: wrapCSVRenderer("vm-extensions", vmExtensionsCSV)},
	"vms":                 {table: wrapTableRenderer("vms", vmsTable), csv: wrapCSVRenderer("vms", vmsCSV)},
	"vmss":                {table: wrapTableRenderer("vmss", vmssTable), csv: wrapCSVRenderer("vmss", vmssCSV)},
	"webjobs":             {table: wrapTableRenderer("webjobs", webJobsTable), csv: wrapCSVRenderer("webjobs", webJobsCSV)},
	"whoami":              {table: wrapTableRenderer("whoami", whoAmITable), csv: wrapCSVRenderer("whoami", whoAmICSV)},
	"workloads":           {table: wrapTableRenderer("workloads", workloadsTable), csv: wrapCSVRenderer("workloads", workloadsCSV)},
}

func renderRegistryEntry(command string) (rendererEntry, error) {
	entry, ok := renderRegistry[command]
	if !ok {
		return rendererEntry{}, fmt.Errorf("rendering is not implemented for command %q", command)
	}
	return entry, nil
}
