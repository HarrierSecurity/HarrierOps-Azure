package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	liptable "github.com/charmbracelet/lipgloss/table"

	"harrierops-azure/internal/models"
)

func Table(command string, payload any) (string, error) {
	entry, err := renderRegistryEntry(command)
	if err != nil {
		return "", err
	}
	if entry.table == nil {
		return "", fmt.Errorf("table rendering is not implemented for command %q", command)
	}
	return entry.table(payload)
}

func renderStructuredTable(title string, headers []string, rows [][]string) string {
	headerStyle := lipgloss.NewStyle().Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true)

	table := liptable.New().
		Headers(headers...).
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == liptable.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	return titleStyle.Render(title) + "\n\n" + strings.TrimRight(table.String(), "\n") + "\n"
}

func renderListTable(title string, headers []string, rows [][]string, emptyRow []string, takeaway string) string {
	if len(rows) == 0 {
		rows = append(rows, emptyRow)
	}

	output := renderStructuredTable(title, headers, rows)
	if takeaway != "" {
		output += "\nTakeaway: " + takeaway + "\n"
	}
	return output
}

func rbacTable(payload models.RbacOutput) string {
	headerStyle := lipgloss.NewStyle().Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true)

	rows := make([][]string, 0, len(payload.RoleAssignments))
	for _, assignment := range payload.RoleAssignments {
		rows = append(rows, []string{
			assignment.ID,
			assignment.PrincipalID,
			assignment.PrincipalType,
			assignment.RoleDefinitionID,
			assignment.RoleName,
			assignment.ScopeID,
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"", "", "", "", "no records", ""})
	}

	table := liptable.New().
		Headers("id", "principal_id", "principal_type", "role_definition_id", "role_name", "scope_id").
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == liptable.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	return titleStyle.Render("azurefox rbac") + "\n\n" + strings.TrimRight(table.String(), "\n") + "\n"
}

func inventoryTable(payload models.InventoryOutput) string {
	headerStyle := lipgloss.NewStyle().Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true)
	topType, _ := topResourceType(payload.TopResourceTypes)

	rows := [][]string{
		{"resource_groups", fmt.Sprintf("%d", payload.ResourceGroupCount)},
		{"resources", fmt.Sprintf("%d", payload.ResourceCount)},
		{"top_type", topType},
		{"issues", fmt.Sprintf("%d", len(payload.Issues))},
	}

	table := liptable.New().
		Headers("field", "value").
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == liptable.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	return titleStyle.Render("azurefox inventory") + "\n\n" + strings.TrimRight(table.String(), "\n") + "\n"
}

func appServicesTable(payload models.AppServicesOutput) string {
	rows := make([][]string, 0, len(payload.AppServices))
	for _, app := range payload.AppServices {
		rows = append(rows, []string{
			app.Name,
			valueOrEmpty(app.DefaultHostname),
			valueOrFallback(app.RuntimeStack, "-"),
			resourceIdentityContext(app.WorkloadIdentityType, app.WorkloadIdentityIDs),
			appServiceExposureContext(app),
			appServicePostureContext(app),
			app.Summary,
		})
	}
	return renderListTable("azurefox app-services", []string{
		"app service", "hostname", "runtime", "identity", "exposure", "posture", "why it matters",
	}, rows, []string{"no App Service apps visible", "", "", "", "", "", ""}, appServicesTakeaway(payload))
}

func functionsTable(payload models.FunctionsOutput) string {
	rows := make([][]string, 0, len(payload.FunctionApps))
	for _, app := range payload.FunctionApps {
		rows = append(rows, []string{
			app.Name,
			valueOrEmpty(app.DefaultHostname),
			functionRuntimeContext(app),
			resourceIdentityContext(app.WorkloadIdentityType, app.WorkloadIdentityIDs),
			functionDeploymentContext(app),
			functionPostureContext(app),
			app.Summary,
		})
	}
	return renderListTable("azurefox functions", []string{
		"function app", "hostname", "runtime", "identity", "deployment", "posture", "why it matters",
	}, rows, []string{"no Function Apps visible", "", "", "", "", "", ""}, functionsTakeaway(payload))
}

func containerAppsTable(payload models.ContainerAppsOutput) string {
	rows := make([][]string, 0, len(payload.ContainerApps))
	for _, app := range payload.ContainerApps {
		rows = append(rows, []string{
			app.Name,
			containerAppEnvironmentContext(app),
			valueOrFallback(app.DefaultHostname, "-"),
			containerAppIngressContext(app),
			resourceIdentityContext(app.WorkloadIdentityType, app.WorkloadIdentityIDs),
			containerAppRevisionContext(app),
			app.Summary,
		})
	}
	return renderListTable("azurefox container-apps", []string{
		"container app", "environment", "hostname", "ingress", "identity", "revisions", "why it matters",
	}, rows, []string{"no Container Apps visible", "", "", "", "", "", ""}, containerAppsTakeaway(payload))
}

func containerInstancesTable(payload models.ContainerInstancesOutput) string {
	rows := make([][]string, 0, len(payload.ContainerInstances))
	for _, item := range payload.ContainerInstances {
		rows = append(rows, []string{
			item.Name,
			containerInstanceEndpointContext(item),
			containerInstanceNetworkContext(item),
			resourceIdentityContext(item.WorkloadIdentityType, item.WorkloadIdentityIDs),
			containerInstanceRuntimeContext(item),
			containerInstanceImagesContext(item),
			item.Summary,
		})
	}
	return renderListTable("azurefox container-instances", []string{
		"container group", "endpoint", "network", "identity", "runtime", "images", "why it matters",
	}, rows, []string{"no Container Instances visible", "", "", "", "", "", ""}, containerInstancesTakeaway(payload))
}

func armDeploymentsTable(payload models.ArmDeploymentsOutput) string {
	rows := make([][]string, 0, len(payload.Deployments))
	for _, deployment := range payload.Deployments {
		rows = append(rows, []string{
			deployment.Name,
			armDeploymentScopeLabel(deployment),
			deployment.ProvisioningState,
			fmt.Sprintf("%d", deployment.OutputsCount),
			armDeploymentLinkedReferenceSummary(deployment),
			deployment.Summary,
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"", "", "no deployment history visible", "", "", ""})
	}

	output := renderStructuredTable("azurefox arm-deployments", []string{
		"deployment", "scope", "state", "outputs", "linked refs", "why it matters",
	}, rows)
	if len(payload.Findings) > 0 {
		output += "\nFindings:\n"
		for _, finding := range payload.Findings {
			output += fmt.Sprintf("- %s: %s\n", strings.ToUpper(finding.Severity), finding.Title)
			output += fmt.Sprintf("  %s\n", finding.Description)
		}
	}

	return output + "\nTakeaway: " + armDeploymentsTakeaway(payload) + "\n"
}

func appServicesTakeaway(payload models.AppServicesOutput) string {
	httpsOnly := 0
	publicNetwork := 0
	identities := 0

	for _, app := range payload.AppServices {
		if app.HTTPSOnly {
			httpsOnly++
		}
		if strings.EqualFold(valueOrEmpty(app.PublicNetworkAccess), "enabled") {
			publicNetwork++
		}
		if app.WorkloadIdentityType != nil && *app.WorkloadIdentityType != "" {
			identities++
		}
	}

	return fmt.Sprintf(
		"%d App Service apps visible; %d keep public network access enabled, %d enforce HTTPS-only, and %d carry managed identity context.",
		len(payload.AppServices),
		publicNetwork,
		httpsOnly,
		identities,
	)
}

func functionsTakeaway(payload models.FunctionsOutput) string {
	identities := 0
	runFromPackage := 0
	keyVaultBacked := 0

	for _, app := range payload.FunctionApps {
		if app.WorkloadIdentityType != nil && *app.WorkloadIdentityType != "" {
			identities++
		}
		if app.RunFromPackage != nil && *app.RunFromPackage {
			runFromPackage++
		}
		if app.KeyVaultReferenceCount != nil && *app.KeyVaultReferenceCount > 0 {
			keyVaultBacked++
		}
	}

	return fmt.Sprintf(
		"%d Function Apps visible; %d carry managed identity context, %d show run-from-package deployment, and %d include Key Vault-backed settings.",
		len(payload.FunctionApps),
		identities,
		runFromPackage,
		keyVaultBacked,
	)
}

func containerAppsTakeaway(payload models.ContainerAppsOutput) string {
	external := 0
	hostnames := 0
	identities := 0

	for _, app := range payload.ContainerApps {
		if app.ExternalIngressEnabled != nil && *app.ExternalIngressEnabled {
			external++
		}
		if app.DefaultHostname != nil && *app.DefaultHostname != "" {
			hostnames++
		}
		if app.WorkloadIdentityType != nil && *app.WorkloadIdentityType != "" {
			identities++
		}
	}

	return fmt.Sprintf(
		"%d Container Apps visible; %d expose external ingress, %d publish visible hostnames, and %d carry managed identity context.",
		len(payload.ContainerApps),
		external,
		hostnames,
		identities,
	)
}

func containerInstancesTakeaway(payload models.ContainerInstancesOutput) string {
	publicEndpoints := 0
	identities := 0
	subnets := 0

	for _, item := range payload.ContainerInstances {
		if valueOrEmpty(item.PublicIPAddress) != "" || valueOrEmpty(item.FQDN) != "" {
			publicEndpoints++
		}
		if valueOrEmpty(item.WorkloadIdentityType) != "" {
			identities++
		}
		if len(item.SubnetIDs) > 0 {
			subnets++
		}
	}

	return fmt.Sprintf(
		"%d Container Instances visible; %d publish public endpoint cues, %d carry managed identity context, and %d show subnet placement.",
		len(payload.ContainerInstances),
		publicEndpoints,
		identities,
		subnets,
	)
}

func endpointsTable(payload models.EndpointsOutput) string {
	rows := make([][]string, 0, len(payload.Endpoints))
	for _, endpoint := range payload.Endpoints {
		rows = append(rows, []string{
			endpoint.Endpoint,
			endpoint.SourceAssetName,
			endpoint.SourceAssetKind,
			endpoint.ExposureFamily,
			endpoint.IngressPath,
			endpoint.Summary,
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"", "no endpoints visible", "", "", "", ""})
	}

	return renderStructuredTable("azurefox endpoints", []string{
		"endpoint", "asset", "kind", "family", "ingress", "why it matters",
	}, rows)
}

func networkPortsTable(payload models.NetworkPortsOutput) string {
	rows := make([][]string, 0, len(payload.NetworkPorts))
	for _, networkPort := range payload.NetworkPorts {
		rows = append(rows, []string{
			networkPort.AssetName,
			networkPort.Endpoint,
			networkPort.Protocol,
			networkPort.Port,
			networkPort.AllowSourceSummary,
			networkPort.ExposureConfidence,
			networkPort.Summary,
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"no NIC-backed public port rows visible", "", "", "", "", "", ""})
	}

	return renderStructuredTable("azurefox network-ports", []string{
		"asset", "endpoint", "protocol", "port", "allow source", "confidence", "why it matters",
	}, rows) + "\nTakeaway: " + networkPortsTakeaway(payload) + "\n"
}

func networkPortsTakeaway(payload models.NetworkPortsOutput) string {
	confidenceCounts := map[string]int{}
	for _, row := range payload.NetworkPorts {
		key := row.ExposureConfidence
		if key == "" {
			key = "unknown"
		}
		confidenceCounts[key]++
	}

	keys := make([]string, 0, len(confidenceCounts))
	for key := range confidenceCounts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%d %s", confidenceCounts[key], key))
	}

	counts := "no port exposure rows visible"
	if len(parts) > 0 {
		counts = strings.Join(parts, ", ")
	}
	return fmt.Sprintf("%d port exposure rows visible; %s.", len(payload.NetworkPorts), counts)
}

func networkEffectiveTable(payload models.NetworkEffectiveOutput) string {
	rows := make([][]string, 0, len(payload.EffectiveExposures))
	for _, exposure := range payload.EffectiveExposures {
		rows = append(rows, []string{
			exposure.AssetName,
			exposure.Endpoint,
			exposure.EffectiveExposure,
			join(exposure.InternetExposedPorts, ", "),
			join(exposure.ConstrainedPorts, ", "),
			exposure.Summary,
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"no public-IP exposure summaries visible", "", "", "", "", ""})
	}

	return renderStructuredTable("azurefox network-effective", []string{
		"asset", "endpoint", "priority", "internet ports", "narrower ports", "why it matters",
	}, rows) + "\nTakeaway: " + networkEffectiveTakeaway(payload) + "\n"
}

func networkEffectiveTakeaway(payload models.NetworkEffectiveOutput) string {
	byConfidence := map[string]int{}
	internetExposed := 0
	for _, exposure := range payload.EffectiveExposures {
		key := strings.ToLower(exposure.EffectiveExposure)
		if key == "" {
			key = "unknown"
		}
		byConfidence[key]++
		if len(exposure.InternetExposedPorts) > 0 {
			internetExposed++
		}
	}
	return fmt.Sprintf(
		"%d public-IP exposure summaries visible; %d high, %d medium, %d low, and %d show broad internet-facing allow evidence.",
		len(payload.EffectiveExposures),
		byConfidence["high"],
		byConfidence["medium"],
		byConfidence["low"],
		internetExposed,
	)
}

func nicsTable(payload models.NicsOutput) string {
	rows := make([][]string, 0, len(payload.NicAssets))
	for _, nic := range payload.NicAssets {
		rows = append(rows, []string{
			nic.Name,
			nicDisplayResourceName(nic.AttachedAssetID),
			join(nic.PrivateIPs, ", "),
			nicDisplayResourceRefs(nic.PublicIPIDs),
			nicNetworkScopeSummary(nic),
			nicDisplayResourceName(nic.NetworkSecurityGroupID),
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"no nic assets visible", "", "", "", "", ""})
	}

	return renderStructuredTable("azurefox nics", []string{
		"nic", "attached asset", "private ips", "public ip refs", "subnet / vnet", "nsg",
	}, rows)
}

func vmsTable(payload models.VmsOutput) string {
	rows := make([][]string, 0, len(payload.VMAssets))
	for _, vm := range payload.VMAssets {
		rows = append(rows, []string{
			vm.Name,
			vm.VMType,
			join(vm.PublicIPs, ", "),
			join(vm.PrivateIPs, ", "),
			fmt.Sprintf("%d", len(vm.IdentityIDs)),
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"no compute assets visible", "", "", "", ""})
	}

	output := renderStructuredTable("azurefox vms", []string{
		"asset", "type", "public ips", "private ips", "identities",
	}, rows)
	if len(payload.Findings) > 0 {
		output += "\nFindings:\n"
		for _, finding := range payload.Findings {
			output += fmt.Sprintf("- %s: %s\n", strings.ToUpper(finding.Severity), finding.Title)
			output += fmt.Sprintf("  %s\n", finding.Description)
		}
	}

	return output + "\nTakeaway: " + vmsTakeaway(payload) + "\n"
}

func vmsTakeaway(payload models.VmsOutput) string {
	publicAssets := 0
	for _, vm := range payload.VMAssets {
		if len(vm.PublicIPs) > 0 {
			publicAssets++
		}
	}
	return fmt.Sprintf("%d compute assets visible; %d have public IP exposure.", len(payload.VMAssets), publicAssets)
}

func vmssTable(payload models.VmssOutput) string {
	rows := make([][]string, 0, len(payload.VmssAssets))
	for _, vmss := range payload.VmssAssets {
		rows = append(rows, []string{
			vmss.Name,
			vmss.Location,
			vmssCapacityContext(vmss),
			vmssRolloutContext(vmss),
			vmssIdentityContext(vmss),
			vmssFrontendContext(vmss),
			vmssNetworkContext(vmss),
			vmss.Summary,
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"no scale sets visible", "", "", "", "", "", "", ""})
	}

	return renderStructuredTable("azurefox vmss", []string{
		"scale set", "location", "sku / capacity", "orchestration", "identity", "frontend", "network", "why it matters",
	}, rows) + "\nTakeaway: " + vmssTakeaway(payload) + "\n"
}

func vmssTakeaway(payload models.VmssOutput) string {
	identityAssets := 0
	publicFrontendAssets := 0
	configuredInstances := 0

	for _, asset := range payload.VmssAssets {
		if asset.IdentityType != nil {
			identityAssets++
		}
		if asset.PublicIPConfigurationCount > 0 {
			publicFrontendAssets++
		}
		if asset.InstanceCount != nil {
			configuredInstances += *asset.InstanceCount
		}
	}

	return fmt.Sprintf(
		"%d VM scale sets visible; %d show public frontend cues, %d carry managed identity context, and %d configured instances are visible.",
		len(payload.VmssAssets),
		publicFrontendAssets,
		identityAssets,
		configuredInstances,
	)
}

func workloadsTable(payload models.WorkloadsOutput) string {
	rows := make([][]string, 0, len(payload.Workloads))
	for _, workload := range payload.Workloads {
		rows = append(rows, []string{
			workload.AssetName,
			workload.AssetKind,
			workloadIdentityContext(workload),
			join(workload.Endpoints, ", "),
			join(workload.IngressPaths, ", "),
			workload.Summary,
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"no workloads visible", "", "", "", "", ""})
	}

	return renderStructuredTable("azurefox workloads", []string{
		"workload", "kind", "identity", "endpoints", "ingress", "why it matters",
	}, rows) + "\nTakeaway: " + workloadsTakeaway(payload) + "\n"
}

func resourceIdentityContext(identityType *string, identityIDs []string) string {
	parts := make([]string, 0, 2)
	if identityType != nil && *identityType != "" {
		parts = append(parts, *identityType)
	}
	if len(identityIDs) > 0 {
		parts = append(parts, fmt.Sprintf("user-assigned=%d", len(identityIDs)))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func appServiceExposureContext(app models.AppServiceAsset) string {
	parts := make([]string, 0, 2)
	if app.DefaultHostname != nil && *app.DefaultHostname != "" {
		parts = append(parts, "hostname")
	}
	if app.PublicNetworkAccess != nil && *app.PublicNetworkAccess != "" {
		parts = append(parts, "public="+*app.PublicNetworkAccess)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func appServicePostureContext(app models.AppServiceAsset) string {
	parts := []string{fmt.Sprintf("https=%s", yesNo(app.HTTPSOnly))}
	if app.MinTLSVersion != nil && *app.MinTLSVersion != "" {
		parts = append(parts, "tls="+*app.MinTLSVersion)
	}
	if app.FTPSState != nil && *app.FTPSState != "" {
		parts = append(parts, "ftps="+*app.FTPSState)
	}
	if app.ClientCertEnabled {
		parts = append(parts, "client-cert=yes")
	}
	return strings.Join(parts, "; ")
}

func functionRuntimeContext(app models.FunctionAppAsset) string {
	parts := make([]string, 0, 2)
	if app.RuntimeStack != nil && *app.RuntimeStack != "" {
		parts = append(parts, *app.RuntimeStack)
	}
	if app.FunctionsExtensionVersion != nil && *app.FunctionsExtensionVersion != "" {
		parts = append(parts, "functions="+*app.FunctionsExtensionVersion)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func functionDeploymentContext(app models.FunctionAppAsset) string {
	parts := make([]string, 0, 3)
	if app.AzureWebJobsStorageValueType != nil && *app.AzureWebJobsStorageValueType != "" {
		parts = append(parts, "storage="+*app.AzureWebJobsStorageValueType)
	}
	if app.RunFromPackage != nil {
		if *app.RunFromPackage {
			parts = append(parts, "package=yes")
		} else {
			parts = append(parts, "package=disabled")
		}
	}
	if app.KeyVaultReferenceCount != nil {
		parts = append(parts, fmt.Sprintf("kv-refs=%d", *app.KeyVaultReferenceCount))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func functionPostureContext(app models.FunctionAppAsset) string {
	parts := []string{fmt.Sprintf("https=%s", yesNo(app.HTTPSOnly))}
	if app.MinTLSVersion != nil && *app.MinTLSVersion != "" {
		parts = append(parts, "tls="+*app.MinTLSVersion)
	}
	if app.FTPSState != nil && *app.FTPSState != "" {
		parts = append(parts, "ftps="+*app.FTPSState)
	}
	if app.AlwaysOn != nil {
		if *app.AlwaysOn {
			parts = append(parts, "always-on=yes")
		} else {
			parts = append(parts, "always-on=no")
		}
	}
	return strings.Join(parts, "; ")
}

func containerAppEnvironmentContext(app models.ContainerAppAsset) string {
	if app.EnvironmentID == nil || *app.EnvironmentID == "" {
		return "-"
	}
	parts := strings.Split(strings.TrimRight(*app.EnvironmentID, "/"), "/")
	return parts[len(parts)-1]
}

func containerAppIngressContext(app models.ContainerAppAsset) string {
	parts := make([]string, 0, 3)
	if app.ExternalIngressEnabled != nil {
		if *app.ExternalIngressEnabled {
			parts = append(parts, "external")
		} else {
			parts = append(parts, "internal")
		}
	}
	if app.IngressTargetPort != nil {
		parts = append(parts, fmt.Sprintf("port %d", *app.IngressTargetPort))
	}
	if app.IngressTransport != nil && *app.IngressTransport != "" {
		parts = append(parts, *app.IngressTransport)
	}
	if len(parts) == 0 {
		return "not visible"
	}
	return strings.Join(parts, "; ")
}

func containerAppRevisionContext(app models.ContainerAppAsset) string {
	parts := make([]string, 0, 2)
	if app.RevisionMode != nil && *app.RevisionMode != "" {
		parts = append(parts, *app.RevisionMode)
	}
	if app.LatestReadyRevisionName != nil && *app.LatestReadyRevisionName != "" {
		parts = append(parts, "ready "+*app.LatestReadyRevisionName)
	} else if app.LatestRevisionName != nil && *app.LatestRevisionName != "" {
		parts = append(parts, "latest "+*app.LatestRevisionName)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func containerInstanceEndpointContext(item models.ContainerInstanceAsset) string {
	parts := []string{}
	if value := valueOrEmpty(item.FQDN); value != "" {
		parts = append(parts, value)
	}
	if value := valueOrEmpty(item.PublicIPAddress); value != "" {
		parts = append(parts, value)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func containerInstanceNetworkContext(item models.ContainerInstanceAsset) string {
	parts := []string{}
	if len(item.ExposedPorts) > 0 {
		ports := make([]string, 0, len(item.ExposedPorts))
		for _, port := range item.ExposedPorts {
			ports = append(ports, fmt.Sprintf("%d", port))
		}
		portsText := strings.Join(ports, ", ")
		if len(ports) > 5 {
			portsText = strings.Join(ports[:5], ", ") + "..."
		}
		parts = append(parts, "ports "+portsText)
	}
	if len(item.SubnetIDs) > 0 {
		parts = append(parts, fmt.Sprintf("subnets=%d", len(item.SubnetIDs)))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func containerInstanceRuntimeContext(item models.ContainerInstanceAsset) string {
	parts := []string{}
	if value := valueOrEmpty(item.OSType); value != "" {
		parts = append(parts, "os="+value)
	}
	if value := valueOrEmpty(item.RestartPolicy); value != "" {
		parts = append(parts, "restart="+value)
	}
	if item.ContainerCount != nil {
		parts = append(parts, fmt.Sprintf("containers=%d", *item.ContainerCount))
	}
	if value := valueOrEmpty(item.ProvisioningState); value != "" {
		parts = append(parts, "state="+value)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func containerInstanceImagesContext(item models.ContainerInstanceAsset) string {
	if len(item.ContainerImages) == 0 {
		return "-"
	}
	if len(item.ContainerImages) == 1 {
		return item.ContainerImages[0]
	}
	return fmt.Sprintf("%s (+%d more)", item.ContainerImages[0], len(item.ContainerImages)-1)
}

func aksVersionContext(cluster models.AksClusterAsset) string {
	parts := []string{}
	if cluster.KubernetesVersion != nil {
		parts = append(parts, "k8s="+*cluster.KubernetesVersion)
	}
	if cluster.AgentPoolCount != nil {
		parts = append(parts, fmt.Sprintf("pools=%d", *cluster.AgentPoolCount))
	}
	if cluster.SKUTier != nil {
		parts = append(parts, "tier="+*cluster.SKUTier)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func aksIdentityContext(cluster models.AksClusterAsset) string {
	parts := []string{}
	if cluster.ClusterIdentityType != nil {
		parts = append(parts, *cluster.ClusterIdentityType)
	}
	if len(cluster.ClusterIdentityIDs) > 0 {
		parts = append(parts, fmt.Sprintf("user-assigned=%d", len(cluster.ClusterIdentityIDs)))
	}
	if valueOrEmpty(cluster.ClusterIdentityType) == "ServicePrincipal" && cluster.ClusterClientID != nil {
		parts = append(parts, "client-id=yes")
	}
	if cluster.WorkloadIdentityEnabled != nil {
		if *cluster.WorkloadIdentityEnabled {
			parts = append(parts, "workload-id=yes")
		} else {
			parts = append(parts, "workload-id=no")
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func aksEndpointContext(cluster models.AksClusterAsset) string {
	parts := []string{}
	if cluster.PrivateClusterEnabled != nil {
		if *cluster.PrivateClusterEnabled {
			parts = append(parts, "private-api=yes")
		} else {
			parts = append(parts, "private-api=no")
		}
	}
	if cluster.FQDN != nil {
		parts = append(parts, "fqdn")
	}
	if cluster.PrivateFQDN != nil {
		parts = append(parts, "private-fqdn")
	}
	if cluster.PublicFQDNEnabled != nil {
		if *cluster.PublicFQDNEnabled {
			parts = append(parts, "public-fqdn=yes")
		} else if cluster.PrivateClusterEnabled != nil && *cluster.PrivateClusterEnabled {
			parts = append(parts, "public-fqdn=no")
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func aksAuthContext(cluster models.AksClusterAsset) string {
	parts := []string{}
	if cluster.AADManaged != nil {
		if *cluster.AADManaged {
			parts = append(parts, "aad=yes")
		} else {
			parts = append(parts, "aad=no")
		}
	}
	if cluster.AzureRBACEnabled != nil {
		if *cluster.AzureRBACEnabled {
			parts = append(parts, "azure-rbac=yes")
		} else {
			parts = append(parts, "azure-rbac=no")
		}
	}
	if cluster.LocalAccountsDisabled != nil {
		if *cluster.LocalAccountsDisabled {
			parts = append(parts, "local-accounts=disabled")
		} else {
			parts = append(parts, "local-accounts=enabled")
		}
	}
	if cluster.OIDCIssuerEnabled != nil {
		if *cluster.OIDCIssuerEnabled {
			parts = append(parts, "oidc=yes")
		} else {
			parts = append(parts, "oidc=no")
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func aksNetworkContext(cluster models.AksClusterAsset) string {
	parts := []string{}
	if cluster.NetworkPlugin != nil {
		parts = append(parts, "plugin="+*cluster.NetworkPlugin)
	}
	if cluster.NetworkPolicy != nil {
		parts = append(parts, "policy="+*cluster.NetworkPolicy)
	}
	if cluster.OutboundType != nil {
		parts = append(parts, "outbound="+*cluster.OutboundType)
	}
	if len(cluster.AddonNames) > 0 {
		parts = append(parts, fmt.Sprintf("addons=%d", len(cluster.AddonNames)))
	}
	if cluster.WebAppRoutingEnabled != nil {
		if *cluster.WebAppRoutingEnabled {
			parts = append(parts, "webapp-routing=yes")
		} else {
			parts = append(parts, "webapp-routing=no")
		}
	}
	if cluster.NodeResourceGroup != nil {
		parts = append(parts, "node-rg="+*cluster.NodeResourceGroup)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func acrAuthContext(registry models.AcrRegistryAsset) string {
	parts := []string{}
	if registry.AdminUserEnabled != nil {
		parts = append(parts, "admin="+yesNo(*registry.AdminUserEnabled))
	}
	if registry.AnonymousPullEnabled != nil {
		parts = append(parts, "anon-pull="+yesNo(*registry.AnonymousPullEnabled))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func acrExposureContext(registry models.AcrRegistryAsset) string {
	parts := []string{}
	if registry.PublicNetworkAccess != nil {
		parts = append(parts, "public="+*registry.PublicNetworkAccess)
	}
	if registry.NetworkRuleDefaultAction != nil {
		parts = append(parts, "default="+*registry.NetworkRuleDefaultAction)
	}
	if registry.PrivateEndpointConnectionCount != nil {
		parts = append(parts, "pe="+intText(*registry.PrivateEndpointConnectionCount))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func acrDepthContext(registry models.AcrRegistryAsset) string {
	parts := []string{}
	if registry.WebhookCount != nil {
		parts = append(parts, "webhooks="+intText(*registry.WebhookCount))
	}
	if registry.EnabledWebhookCount != nil {
		parts = append(parts, "enabled="+intText(*registry.EnabledWebhookCount))
	}
	if registry.BroadWebhookScopeCount != nil && *registry.BroadWebhookScopeCount > 0 {
		parts = append(parts, "wide-scopes="+intText(*registry.BroadWebhookScopeCount))
	}
	if len(registry.WebhookActionTypes) > 0 {
		parts = append(parts, "actions="+strings.Join(registry.WebhookActionTypes, ","))
	}
	if registry.ReplicationCount != nil {
		parts = append(parts, "replications="+intText(*registry.ReplicationCount))
	}
	if len(registry.ReplicationRegions) > 0 {
		parts = append(parts, "regions="+strings.Join(registry.ReplicationRegions, ","))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func acrPostureContext(registry models.AcrRegistryAsset) string {
	parts := []string{}
	if registry.SKUName != nil {
		parts = append(parts, *registry.SKUName)
	}
	if registry.NetworkRuleBypassOptions != nil {
		parts = append(parts, "bypass="+*registry.NetworkRuleBypassOptions)
	}
	if registry.DataEndpointEnabled != nil {
		parts = append(parts, "data-endpoint="+yesNo(*registry.DataEndpointEnabled))
	}
	if registry.QuarantinePolicyStatus != nil {
		parts = append(parts, "quarantine="+*registry.QuarantinePolicyStatus)
	}
	if registry.RetentionPolicyStatus != nil {
		if strings.EqualFold(*registry.RetentionPolicyStatus, "enabled") && registry.RetentionPolicyDays != nil {
			parts = append(parts, "retention="+intText(*registry.RetentionPolicyDays)+"d")
		} else {
			parts = append(parts, "retention="+*registry.RetentionPolicyStatus)
		}
	}
	if registry.TrustPolicyStatus != nil {
		if strings.EqualFold(*registry.TrustPolicyStatus, "enabled") && registry.TrustPolicyType != nil {
			parts = append(parts, "trust="+*registry.TrustPolicyType)
		} else {
			parts = append(parts, "trust="+*registry.TrustPolicyStatus)
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func databasesInventoryContext(server models.DatabaseServerAsset) string {
	parts := []string{}
	if server.DatabaseCount != nil {
		parts = append(parts, "dbs="+intText(*server.DatabaseCount))
	}
	if len(server.UserDatabaseNames) > 0 {
		parts = append(parts, strings.Join(server.UserDatabaseNames, ","))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func databasesExposureContext(server models.DatabaseServerAsset) string {
	parts := []string{}
	if server.FullyQualifiedDomainName != nil && *server.FullyQualifiedDomainName != "" {
		parts = append(parts, "fqdn")
	}
	if server.PublicNetworkAccess != nil && *server.PublicNetworkAccess != "" {
		parts = append(parts, "public="+*server.PublicNetworkAccess)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func databasesPostureContext(server models.DatabaseServerAsset) string {
	parts := []string{}
	if server.MinimalTLSVersion != nil && *server.MinimalTLSVersion != "" {
		parts = append(parts, "tls="+*server.MinimalTLSVersion)
	}
	if server.ServerVersion != nil && *server.ServerVersion != "" {
		parts = append(parts, "version="+*server.ServerVersion)
	}
	if server.HighAvailabilityMode != nil && *server.HighAvailabilityMode != "" {
		parts = append(parts, "ha="+*server.HighAvailabilityMode)
	}
	if server.DelegatedSubnetResourceID != nil && *server.DelegatedSubnetResourceID != "" {
		parts = append(parts, "delegated-subnet=yes")
	}
	if server.PrivateDNSZoneResourceID != nil && *server.PrivateDNSZoneResourceID != "" {
		parts = append(parts, "private-dns=yes")
	}
	if server.State != nil && *server.State != "" {
		parts = append(parts, "state="+*server.State)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func applicationGatewayExposureContext(gateway models.ApplicationGatewayAsset) string {
	parts := []string{}
	if gateway.PublicFrontendCount > 0 {
		publicPhrase := "public=" + intText(gateway.PublicFrontendCount)
		if len(gateway.PublicIPAddresses) > 0 {
			publicPhrase += " (" + strings.Join(gateway.PublicIPAddresses, ", ") + ")"
		}
		parts = append(parts, publicPhrase)
	}
	if gateway.PrivateFrontendCount > 0 {
		parts = append(parts, "private="+intText(gateway.PrivateFrontendCount))
	}
	if len(gateway.SubnetIDs) > 0 {
		parts = append(parts, "subnets="+intText(len(gateway.SubnetIDs)))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func applicationGatewayRoutingContext(gateway models.ApplicationGatewayAsset) string {
	parts := []string{"listeners=" + intText(gateway.ListenerCount), "rules=" + intText(gateway.RequestRoutingRuleCount)}
	if gateway.URLPathMapCount > 0 {
		parts = append(parts, "path-maps="+intText(gateway.URLPathMapCount))
	}
	if gateway.RedirectConfigurationCount > 0 {
		parts = append(parts, "redirects="+intText(gateway.RedirectConfigurationCount))
	}
	if gateway.RewriteRuleSetCount > 0 {
		parts = append(parts, "rewrites="+intText(gateway.RewriteRuleSetCount))
	}
	return strings.Join(parts, "; ")
}

func applicationGatewayBackendContext(gateway models.ApplicationGatewayAsset) string {
	return "pools=" + intText(gateway.BackendPoolCount) + "; targets=" + intText(gateway.BackendTargetCount)
}

func applicationGatewayWAFContext(gateway models.ApplicationGatewayAsset) string {
	if gateway.FirewallPolicyID != nil && *gateway.FirewallPolicyID != "" && gateway.WAFMode != nil && *gateway.WAFMode != "" {
		return "policy; " + strings.ToLower(*gateway.WAFMode)
	}
	if gateway.FirewallPolicyID != nil && *gateway.FirewallPolicyID != "" {
		return "policy attached"
	}
	if gateway.WAFEnabled != nil && *gateway.WAFEnabled && gateway.WAFMode != nil && *gateway.WAFMode != "" {
		return "enabled; " + strings.ToLower(*gateway.WAFMode)
	}
	if gateway.WAFEnabled != nil && *gateway.WAFEnabled {
		return "enabled"
	}
	if gateway.WAFEnabled != nil && !*gateway.WAFEnabled {
		return "disabled"
	}
	return "not visible"
}

func dnsInventoryContext(zone models.DnsZoneAsset) string {
	if zone.RecordSetCount == nil && zone.MaxRecordSetCount == nil {
		return "-"
	}
	if zone.RecordSetCount == nil {
		return "records=?/" + intText(*zone.MaxRecordSetCount)
	}
	if zone.MaxRecordSetCount == nil {
		return "records=" + intText(*zone.RecordSetCount)
	}
	return "records=" + intText(*zone.RecordSetCount) + "/" + intText(*zone.MaxRecordSetCount)
}

func dnsNamespaceContext(zone models.DnsZoneAsset) string {
	if zone.ZoneKind == "public" {
		if len(zone.NameServers) == 0 {
			return "-"
		}
		return "ns=" + intText(len(zone.NameServers))
	}

	parts := []string{}
	if zone.LinkedVirtualNetworkCount != nil {
		parts = append(parts, "vnet-links="+intText(*zone.LinkedVirtualNetworkCount))
	}
	if zone.RegistrationVirtualNetworkCount != nil {
		parts = append(parts, "reg-links="+intText(*zone.RegistrationVirtualNetworkCount))
	}
	if zone.PrivateEndpointReferenceCount != nil {
		parts = append(parts, "pe-refs="+intText(*zone.PrivateEndpointReferenceCount))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func apiMgmtInventoryContext(service models.ApiMgmtServiceAsset) string {
	parts := []string{}
	if service.APICount != nil {
		parts = append(parts, "apis="+intText(*service.APICount))
	}
	if service.APISubscriptionRequiredCount != nil {
		if service.APICount != nil {
			parts = append(parts, "sub-required="+intText(*service.APISubscriptionRequiredCount)+"/"+intText(*service.APICount))
		} else {
			parts = append(parts, "sub-required="+intText(*service.APISubscriptionRequiredCount))
		}
	}
	if service.SubscriptionCount != nil {
		parts = append(parts, "subs="+intText(*service.SubscriptionCount))
	}
	if service.ActiveSubscriptionCount != nil {
		parts = append(parts, "active-subs="+intText(*service.ActiveSubscriptionCount))
	}
	if service.BackendCount != nil {
		parts = append(parts, "backends="+intText(*service.BackendCount))
	}
	if len(service.BackendHostnames) > 0 {
		parts = append(parts, "backend-hosts="+intText(len(service.BackendHostnames)))
	}
	if service.NamedValueCount != nil {
		parts = append(parts, "named-values="+intText(*service.NamedValueCount))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func apiMgmtExposureContext(service models.ApiMgmtServiceAsset) string {
	parts := []string{}
	if len(service.GatewayHostnames) > 0 {
		parts = append(parts, "gateway="+intText(len(service.GatewayHostnames)))
	}
	if len(service.ManagementHostnames) > 0 {
		parts = append(parts, "management="+intText(len(service.ManagementHostnames)))
	}
	if len(service.PortalHostnames) > 0 {
		parts = append(parts, "portal="+intText(len(service.PortalHostnames)))
	}
	if service.PublicNetworkAccess != nil {
		parts = append(parts, "public="+*service.PublicNetworkAccess)
	}
	if len(service.PublicIPAddresses) > 0 {
		parts = append(parts, "public-ip="+intText(len(service.PublicIPAddresses)))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func apiMgmtPostureContext(service models.ApiMgmtServiceAsset) string {
	parts := []string{}
	if service.SKUName != nil {
		parts = append(parts, *service.SKUName)
	}
	if service.VirtualNetworkType != nil {
		parts = append(parts, "vnet="+*service.VirtualNetworkType)
	}
	if service.GatewayEnabled != nil {
		if *service.GatewayEnabled {
			parts = append(parts, "gateway=yes")
		} else {
			parts = append(parts, "gateway=no")
		}
	}
	if service.DeveloperPortalStatus != nil {
		parts = append(parts, "devportal="+*service.DeveloperPortalStatus)
	}
	if service.NamedValueSecretCount != nil {
		parts = append(parts, "named-secrets="+intText(*service.NamedValueSecretCount))
	}
	if service.NamedValueKeyVaultCount != nil {
		parts = append(parts, "kv-backed="+intText(*service.NamedValueKeyVaultCount))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func intText(value int) string {
	return fmt.Sprintf("%d", value)
}

func valueOrFallback(value *string, fallback string) string {
	if value == nil || *value == "" {
		return fallback
	}
	return *value
}

func keyVaultTableDefaultAction(vault models.KeyVaultAsset) string {
	if strings.EqualFold(valueOrFallback(vault.PublicNetworkAccess, ""), "Enabled") && valueOrFallback(vault.NetworkDefaultAction, "") == "" {
		return "implicit allow (ACL omitted)"
	}
	return valueOrFallback(vault.NetworkDefaultAction, "-")
}

func automationIdentityContext(item models.AutomationAccountAsset) string {
	if item.IdentityType == nil || *item.IdentityType == "" {
		return "none"
	}
	return *item.IdentityType
}

func automationExecutionContext(item models.AutomationAccountAsset) string {
	return "published=" + intOrUnknown(item.PublishedRunbookCount) + "/" + intOrUnknown(item.RunbookCount) +
		"; job-schedules=" + intOrUnknown(item.JobScheduleCount)
}

func automationTriggerContext(item models.AutomationAccountAsset) string {
	return "schedules=" + intOrUnknown(item.ScheduleCount) + "; webhooks=" + intOrUnknown(item.WebhookCount)
}

func automationWorkerContext(item models.AutomationAccountAsset) string {
	if item.HybridWorkerGroupCount == nil {
		return "groups=?"
	}
	return "groups=" + intOrUnknown(item.HybridWorkerGroupCount)
}

func automationAssetContext(item models.AutomationAccountAsset) string {
	return "cred=" + intOrUnknown(item.CredentialCount) +
		"; cert=" + intOrUnknown(item.CertificateCount) +
		"; conn=" + intOrUnknown(item.ConnectionCount) +
		"; vars=" + intOrUnknown(item.VariableCount) +
		" (" + intOrUnknown(item.EncryptedVariableCount) + " enc)"
}

func automationTakeaway(payload models.AutomationOutput) string {
	identityAccounts := 0
	webhookAccounts := 0
	workerAccounts := 0
	publishedRunbooks := 0
	for _, account := range payload.AutomationAccounts {
		if account.IdentityType != nil && *account.IdentityType != "" {
			identityAccounts++
		}
		if account.WebhookCount != nil && *account.WebhookCount > 0 {
			webhookAccounts++
		}
		if account.HybridWorkerGroupCount != nil && *account.HybridWorkerGroupCount > 0 {
			workerAccounts++
		}
		if account.PublishedRunbookCount != nil {
			publishedRunbooks += *account.PublishedRunbookCount
		}
	}
	return intString(len(payload.AutomationAccounts)) + " Automation account(s) visible; " +
		intString(identityAccounts) + " carry managed identity context, " +
		intString(webhookAccounts) + " expose webhook start paths, " +
		intString(workerAccounts) + " show Hybrid Runbook Worker reach, and " +
		intString(publishedRunbooks) + " published runbooks are visible."
}

func devopsRepositoryContext(item models.DevopsPipelineAsset) string {
	parts := []string{}
	if item.RepositoryHostType != "" || item.RepositoryName != "" {
		parts = append(parts, strings.TrimPrefix(strings.Join([]string{item.RepositoryHostType, item.RepositoryName}, ":"), ":"))
	}
	if item.SourceVisibilityState != "" {
		parts = append(parts, "visibility="+item.SourceVisibilityState)
	}
	if item.DefaultBranch != "" {
		parts = append(parts, item.DefaultBranch)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func devopsTriggerContext(item models.DevopsPipelineAsset) string {
	parts := []string{}
	if len(item.ExecutionModes) > 0 {
		parts = append(parts, strings.Join(item.ExecutionModes, ", "))
	}
	if len(item.TriggerTypes) > 0 {
		parts = append(parts, "triggers="+strings.Join(item.TriggerTypes, ", "))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func devopsInjectionContext(item models.DevopsPipelineAsset) string {
	if len(item.CurrentOperatorInjectionSurfaceTypes) > 0 {
		return strings.Join(item.CurrentOperatorInjectionSurfaceTypes, ", ")
	}
	if item.PrimaryInjectionSurface != "" {
		return item.PrimaryInjectionSurface + " (" + devopsPrimaryAccessState(item) + ")"
	}
	if item.MissingInjectionPoint {
		return "no proven poisonable input"
	}
	return "-"
}

func devopsAccessContext(item models.DevopsPipelineAsset) string {
	parts := []string{}
	if len(item.AzureServiceConnectionNames) > 0 {
		parts = append(parts, "connections="+strings.Join(item.AzureServiceConnectionNames, ", "))
	}
	if len(item.AzureServiceConnectionAuthSchemes) > 0 {
		parts = append(parts, "auth="+strings.Join(item.AzureServiceConnectionAuthSchemes, ", "))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func devopsSecretContext(item models.DevopsPipelineAsset) string {
	parts := []string{}
	if item.SecretVariableCount > 0 {
		parts = append(parts, "secret-vars="+intString(item.SecretVariableCount))
	}
	if len(item.VariableGroupNames) > 0 {
		parts = append(parts, "groups="+intString(len(item.VariableGroupNames)))
	}
	if len(item.KeyVaultGroupNames) > 0 {
		parts = append(parts, "keyvault="+strings.Join(item.KeyVaultGroupNames, ", "))
	}
	if len(parts) == 0 {
		return "none visible"
	}
	return strings.Join(parts, "; ")
}

func devopsTargetContext(item models.DevopsPipelineAsset) string {
	if len(item.TargetClues) == 0 {
		return "review definition directly"
	}
	return strings.Join(item.TargetClues, ", ")
}

func devopsNextReview(item models.DevopsPipelineAsset) string {
	if item.PartialRead {
		return "Restore missing Azure DevOps read paths before deciding the next Azure follow-up."
	}
	primaryTarget := "-"
	for _, clue := range item.TargetClues {
		switch clue {
		case "AKS/Kubernetes":
			primaryTarget = "Check aks for the named deployment target."
		case "App Service":
			primaryTarget = "Check app-services for the named deployment target."
		case "Functions":
			primaryTarget = "Check functions for the named deployment target."
		case "ARM/Bicep/Terraform":
			primaryTarget = "Check arm-deployments for the named deployment target."
		case "ACR/Containers":
			primaryTarget = "Check acr for the named deployment target."
		}
		if primaryTarget != "-" {
			break
		}
	}
	if primaryTarget == "-" && len(item.AzureServiceConnectionNames) > 0 {
		return "Review permissions and role-trusts for the Azure control path behind this service connection."
	}
	if primaryTarget == "-" {
		return "Review the definition directly to confirm the next Azure target."
	}
	if len(item.AzureServiceConnectionNames) > 0 && len(item.KeyVaultNames) > 0 {
		return primaryTarget + " Review permissions and role-trusts for Azure control; review keyvault for vault-backed support."
	}
	if len(item.AzureServiceConnectionNames) > 0 {
		return primaryTarget + " Review permissions and role-trusts for Azure control."
	}
	return primaryTarget
}

func devopsTakeaway(payload models.DevopsOutput) string {
	provenInjection := 0
	queueOnly := 0
	nonRepoTrust := 0
	azurePaths := 0
	visibleAzureRepos := 0
	externalSources := 0
	for _, pipeline := range payload.Pipelines {
		if len(pipeline.CurrentOperatorInjectionSurfaceTypes) > 0 {
			provenInjection++
		}
		if boolPtrValue(pipeline.CurrentOperatorCanQueue) && len(pipeline.CurrentOperatorInjectionSurfaceTypes) == 0 {
			queueOnly++
		}
		if len(pipeline.TrustedInputTypes) > 0 {
			for _, inputType := range pipeline.TrustedInputTypes {
				if inputType != "repository" {
					nonRepoTrust++
					break
				}
			}
		}
		if len(pipeline.AzureServiceConnectionNames) > 0 {
			azurePaths++
		}
		if pipeline.RepositoryHostType == "azure-repos" && pipeline.SourceVisibilityState == "visible" {
			visibleAzureRepos++
		}
		if pipeline.SourceVisibilityState == "external-reference" {
			externalSources++
		}
	}
	return intString(len(payload.Pipelines)) + " Azure DevOps build definition(s) surfaced; " +
		intString(provenInjection) + " expose a proven current-credential injection point, " +
		intString(queueOnly) + " add queue-only support without poisoning proof, " +
		intString(visibleAzureRepos) + " point to visible Azure Repos sources, " +
		intString(externalSources) + " point to external sources, " +
		intString(nonRepoTrust) + " trust non-repo inputs, and " +
		intString(azurePaths) + " show Azure-facing service connections."
}

func devopsPrimaryAccessState(item models.DevopsPipelineAsset) string {
	if len(item.TrustedInputs) == 0 {
		return "unknown"
	}
	return item.TrustedInputs[0].CurrentOperatorAccessState
}

func boolPtrValue(value *bool) bool {
	return value != nil && *value
}

func intOrUnknown(value *int) string {
	if value == nil {
		return "?"
	}
	return intString(*value)
}

func storageExposureContext(asset models.StorageAsset) string {
	parts := []string{"blob-public=" + yesNo(asset.PublicAccess)}
	if asset.PublicNetworkAccess != nil && *asset.PublicNetworkAccess != "" {
		parts = append(parts, "public-net="+strings.ToLower(*asset.PublicNetworkAccess))
	}
	if asset.NetworkDefaultAction != nil && *asset.NetworkDefaultAction != "" {
		parts = append(parts, "default="+*asset.NetworkDefaultAction)
	}
	parts = append(parts, "private-endpoint="+yesNo(asset.PrivateEndpointEnabled))
	return strings.Join(parts, "; ")
}

func storageAuthContext(asset models.StorageAsset) string {
	parts := []string{}
	if asset.AllowSharedKeyAccess != nil {
		parts = append(parts, "shared-key="+yesNo(*asset.AllowSharedKeyAccess))
	}
	if asset.MinimumTLSVersion != nil && *asset.MinimumTLSVersion != "" {
		parts = append(parts, "tls="+*asset.MinimumTLSVersion)
	}
	if asset.HTTPSTrafficOnlyEnabled != nil {
		parts = append(parts, "https-only="+yesNo(*asset.HTTPSTrafficOnlyEnabled))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func storageProtocolContext(asset models.StorageAsset) string {
	parts := []string{}
	if asset.IsHNSEnabled != nil {
		parts = append(parts, "hns="+yesNo(*asset.IsHNSEnabled))
	}
	if asset.IsSFTPEnabled != nil {
		parts = append(parts, "sftp="+yesNo(*asset.IsSFTPEnabled))
	}
	if asset.NFSV3Enabled != nil {
		parts = append(parts, "nfs="+yesNo(*asset.NFSV3Enabled))
	}
	if asset.DNSEndpointType != nil && *asset.DNSEndpointType != "" {
		parts = append(parts, "dns="+strings.ToLower(*asset.DNSEndpointType))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func storageInventoryContext(asset models.StorageAsset) string {
	parts := []string{}
	for _, item := range []struct {
		label string
		value *int
	}{
		{label: "blob", value: asset.ContainerCount},
		{label: "file", value: asset.FileShareCount},
		{label: "queue", value: asset.QueueCount},
		{label: "table", value: asset.TableCount},
	} {
		if item.value != nil {
			parts = append(parts, item.label+"="+intText(*item.value))
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func snapshotDiskPriorityContext(asset models.SnapshotDiskAsset) string {
	parts := []string{}
	if asset.AttachmentState == "detached" {
		parts = append(parts, "detached")
	}
	if asset.AssetKind == "snapshot" {
		parts = append(parts, "offline-copy")
	}
	if strings.EqualFold(valueOrFallback(asset.PublicNetworkAccess, ""), "Enabled") {
		parts = append(parts, "public-net")
	}
	if strings.EqualFold(valueOrFallback(asset.NetworkAccessPolicy, ""), "AllowAll") {
		parts = append(parts, "allow-all")
	}
	if asset.MaxShares != nil && *asset.MaxShares != 1 {
		parts = append(parts, "shared="+intText(*asset.MaxShares))
	}
	if asset.DiskAccessID != nil {
		parts = append(parts, "disk-access")
	}
	if len(parts) == 0 {
		return "baseline"
	}
	return strings.Join(parts, ", ")
}

func snapshotDiskAttachmentContext(asset models.SnapshotDiskAsset) string {
	parts := []string{}
	switch {
	case asset.AttachmentState == "snapshot":
		parts = append(parts, "source="+valueOrFallback(asset.SourceResourceName, "-"))
		if asset.Incremental != nil && *asset.Incremental {
			parts = append(parts, "incremental=yes")
		}
	case asset.AttachedToName != nil:
		parts = append(parts, "attached="+*asset.AttachedToName)
		if asset.DiskRole != nil {
			parts = append(parts, "role="+*asset.DiskRole)
		}
	default:
		parts = append(parts, "detached")
	}
	return strings.Join(parts, "; ")
}

func snapshotDiskSharingContext(asset models.SnapshotDiskAsset) string {
	parts := []string{}
	if asset.NetworkAccessPolicy != nil {
		parts = append(parts, "policy="+*asset.NetworkAccessPolicy)
	}
	if asset.PublicNetworkAccess != nil {
		parts = append(parts, "public="+*asset.PublicNetworkAccess)
	}
	if asset.MaxShares != nil {
		parts = append(parts, "max-shares="+intText(*asset.MaxShares))
	}
	if asset.DiskAccessID != nil {
		parts = append(parts, "disk-access=yes")
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func snapshotDiskEncryptionContext(asset models.SnapshotDiskAsset) string {
	parts := []string{}
	if asset.EncryptionType != nil {
		parts = append(parts, "type="+*asset.EncryptionType)
	}
	if asset.DiskEncryptionSetID != nil {
		parts = append(parts, "des=yes")
	} else {
		parts = append(parts, "des=no")
	}
	if asset.OSType != nil {
		parts = append(parts, "os="+*asset.OSType)
	}
	if asset.SizeGB != nil {
		parts = append(parts, "size="+intText(*asset.SizeGB)+"g")
	}
	return strings.Join(parts, "; ")
}

func workloadsTakeaway(payload models.WorkloadsOutput) string {
	exposed := 0
	identityBearing := 0
	computeAssets := 0

	for _, workload := range payload.Workloads {
		if len(workload.Endpoints) > 0 {
			exposed++
		}
		if workload.IdentityType != nil && *workload.IdentityType != "" {
			identityBearing++
		}
		switch workload.AssetKind {
		case "VM", "VMSS":
			computeAssets++
		}
	}

	return fmt.Sprintf(
		"%d workloads visible; %d with visible endpoint paths, %d with identity context, across %d compute and %d web assets.",
		len(payload.Workloads),
		exposed,
		identityBearing,
		computeAssets,
		len(payload.Workloads)-computeAssets,
	)
}

func acrTakeaway(payload models.AcrOutput) string {
	webhookCounts := []*int{}
	readableWebhookCount := 0
	publicNetwork := 0
	adminAuth := 0
	replicated := 0
	replicationCounts := []*int{}

	for _, registry := range payload.Registries {
		if strings.EqualFold(valueOrFallback(registry.PublicNetworkAccess, ""), "Enabled") {
			publicNetwork++
		}
		if registry.AdminUserEnabled != nil && *registry.AdminUserEnabled {
			adminAuth++
		}
		webhookCounts = append(webhookCounts, registry.WebhookCount)
		if registry.WebhookCount != nil {
			readableWebhookCount += *registry.WebhookCount
		}
		replicationCounts = append(replicationCounts, registry.ReplicationCount)
		if registry.ReplicationCount != nil && len(registry.ReplicationRegions) > 0 {
			replicated++
		}
	}

	webhookPhrase := intText(readableWebhookCount) + " webhooks are visible"
	if anyNilIntPtr(webhookCounts) {
		if readableWebhookCount > 0 {
			webhookPhrase = "at least " + intText(readableWebhookCount) + " webhooks are visible, with some registries outside current credential visibility"
		} else {
			webhookPhrase = "current credentials do not show webhook visibility on at least one visible registry"
		}
	}

	replicationPhrase := intText(replicated) + " registries replicate content into additional regions"
	if anyNilIntPtr(replicationCounts) {
		if replicated > 0 {
			replicationPhrase = "at least " + intText(replicated) + " registries show replicated regions, with some registries outside current credential visibility"
		} else {
			replicationPhrase = "current credentials do not show replication visibility on at least one visible registry"
		}
	} else if replicated == 1 {
		replicationPhrase = "1 registry replicates content into additional regions"
	}

	return fmt.Sprintf(
		"%d registries visible; %d keep public network access enabled, %d allow admin-user auth, %s, and %s.",
		len(payload.Registries),
		publicNetwork,
		adminAuth,
		webhookPhrase,
		replicationPhrase,
	)
}

func databasesTakeaway(payload models.DatabasesOutput) string {
	publicServers := 0
	identities := 0
	engineFamilies := map[string]struct{}{}
	databaseCounts := []*int{}
	readableDatabases := 0

	for _, server := range payload.DatabaseServers {
		if strings.EqualFold(valueOrFallback(server.PublicNetworkAccess, ""), "Enabled") {
			publicServers++
		}
		if server.WorkloadIdentityType != nil && *server.WorkloadIdentityType != "" {
			identities++
		}
		if server.Engine != "" {
			engineFamilies[server.Engine] = struct{}{}
		}
		databaseCounts = append(databaseCounts, server.DatabaseCount)
		if server.DatabaseCount != nil {
			readableDatabases += *server.DatabaseCount
		}
	}

	databasePhrase := intText(readableDatabases) + " user databases are visible"
	if anyNilIntPtr(databaseCounts) {
		if readableDatabases > 0 {
			databasePhrase = "at least " + intText(readableDatabases) + " user databases are visible, with some servers outside current credential visibility"
		} else {
			databasePhrase = "current credentials do not show database visibility on at least one visible server"
		}
	}

	return fmt.Sprintf(
		"%d relational database servers visible across %d engine families; %d keep public network access enabled, %d carry managed identity context, and %s.",
		len(payload.DatabaseServers),
		len(engineFamilies),
		publicServers,
		identities,
		databasePhrase,
	)
}

func keyVaultTakeaway(payload models.KeyVaultOutput) string {
	return fmt.Sprintf(
		"%d Key Vault assets visible; %d exposure or recovery findings.",
		len(payload.KeyVaults),
		len(payload.Findings),
	)
}

func storageTakeaway(payload models.StorageOutput) string {
	publicAssets := 0
	publicNetworkAssets := 0
	sharedKeyAssets := 0
	publicNetworkUnreadable := 0
	sharedKeyUnreadable := 0

	for _, asset := range payload.StorageAssets {
		if asset.PublicAccess {
			publicAssets++
		}
		if strings.EqualFold(valueOrFallback(asset.PublicNetworkAccess, ""), "Enabled") {
			publicNetworkAssets++
		}
		if asset.PublicNetworkAccess == nil {
			publicNetworkUnreadable++
		}
		if asset.AllowSharedKeyAccess != nil && *asset.AllowSharedKeyAccess {
			sharedKeyAssets++
		}
		if asset.AllowSharedKeyAccess == nil {
			sharedKeyUnreadable++
		}
	}

	parts := []string{
		fmt.Sprintf("%d allow public blob access", publicAssets),
		fmt.Sprintf("%d keep public network access enabled", publicNetworkAssets),
	}
	if publicNetworkUnreadable > 0 {
		parts = append(parts, fmt.Sprintf("%d have unreadable public-network posture", publicNetworkUnreadable))
	}
	parts = append(parts, fmt.Sprintf("%d allow shared-key access", sharedKeyAssets))
	if sharedKeyUnreadable > 0 {
		parts = append(parts, fmt.Sprintf("%d have unreadable shared-key posture", sharedKeyUnreadable))
	}

	return fmt.Sprintf("%d storage accounts visible; %s.", len(payload.StorageAssets), strings.Join(parts, ", "))
}

func snapshotsDisksTakeaway(payload models.SnapshotsDisksOutput) string {
	snapshots := 0
	detached := 0
	broadAccess := 0

	for _, asset := range payload.SnapshotDiskAssets {
		if asset.AssetKind == "snapshot" {
			snapshots++
		}
		if asset.AttachmentState == "detached" {
			detached++
		}
		if strings.EqualFold(valueOrFallback(asset.PublicNetworkAccess, ""), "Enabled") ||
			strings.EqualFold(valueOrFallback(asset.NetworkAccessPolicy, ""), "AllowAll") ||
			(asset.MaxShares != nil && *asset.MaxShares != 1) ||
			asset.DiskAccessID != nil {
			broadAccess++
		}
	}

	detachedLabel := "disks"
	if detached == 1 {
		detachedLabel = "disk"
	}

	return fmt.Sprintf(
		"%d disk-backed assets visible; %d snapshots, %d detached %s, and %d show broader sharing or export posture.",
		len(payload.SnapshotDiskAssets),
		snapshots,
		detached,
		detachedLabel,
		broadAccess,
	)
}

func dnsTakeaway(payload models.DnsOutput) string {
	publicZones := 0
	privateZones := 0
	privateEndpointLinked := 0
	recordCounts := []*int{}
	readableRecords := 0

	for _, zone := range payload.DNSZones {
		if zone.ZoneKind == "public" {
			publicZones++
		} else if zone.ZoneKind == "private" {
			privateZones++
		}
		if zone.PrivateEndpointReferenceCount != nil && *zone.PrivateEndpointReferenceCount > 0 {
			privateEndpointLinked++
		}
		recordCounts = append(recordCounts, zone.RecordSetCount)
		if zone.RecordSetCount != nil {
			readableRecords += *zone.RecordSetCount
		}
	}

	recordPhrase := intText(readableRecords) + " record sets are visible"
	if anyNilIntPtr(recordCounts) {
		if readableRecords > 0 {
			recordPhrase = "at least " + intText(readableRecords) + " record sets are visible, with some zones outside current credential visibility"
		} else {
			recordPhrase = "current credentials do not show record-set totals on at least one visible zone"
		}
	}

	return fmt.Sprintf(
		"%d DNS zones visible; %d public, %d private, %d private zone(s) show visible private endpoint references, and %s.",
		len(payload.DNSZones),
		publicZones,
		privateZones,
		privateEndpointLinked,
		recordPhrase,
	)
}

func applicationGatewayTakeaway(payload models.ApplicationGatewayOutput) string {
	publicGateways := 0
	sharedPublicGateways := 0
	weakPublicGateways := 0

	for _, gateway := range payload.ApplicationGateways {
		if gateway.PublicFrontendCount > 0 {
			publicGateways++
			if applicationGatewayHasSharedBreadthTable(gateway) {
				sharedPublicGateways++
			}
			if applicationGatewayWAFRankTable(gateway) < 3 {
				weakPublicGateways++
			}
		}
	}

	return fmt.Sprintf(
		"%d Application Gateways visible; %d have public frontends, %d look like shared public front doors, and %d public gateway(s) lack strong visible WAF coverage. Treat weak shared edge layers as clues that the apps behind them may deserve review next.",
		len(payload.ApplicationGateways),
		publicGateways,
		sharedPublicGateways,
		weakPublicGateways,
	)
}

func applicationGatewayHasSharedBreadthTable(gateway models.ApplicationGatewayAsset) bool {
	return gateway.ListenerCount > 1 ||
		gateway.RequestRoutingRuleCount > 1 ||
		gateway.BackendPoolCount > 1 ||
		gateway.BackendTargetCount > 1
}

func applicationGatewayWAFRankTable(gateway models.ApplicationGatewayAsset) int {
	if gateway.FirewallPolicyID != nil && *gateway.FirewallPolicyID != "" {
		switch strings.ToLower(valueOrFallback(gateway.WAFMode, "")) {
		case "prevention":
			return 3
		case "detection":
			return 1
		default:
			return 2
		}
	}
	if gateway.WAFEnabled != nil && !*gateway.WAFEnabled {
		return 0
	}
	switch strings.ToLower(valueOrFallback(gateway.WAFMode, "")) {
	case "prevention":
		return 3
	case "detection":
		return 1
	default:
		if gateway.WAFEnabled != nil && *gateway.WAFEnabled {
			return 2
		}
		return 4
	}
}

func anyNilIntPtr(values []*int) bool {
	for _, value := range values {
		if value == nil {
			return true
		}
	}
	return false
}

func aksTakeaway(payload models.AksOutput) string {
	privateClusters := 0
	identityClusters := 0
	azureRBACClusters := 0
	federationClusters := 0

	for _, cluster := range payload.AksClusters {
		if cluster.PrivateClusterEnabled != nil && *cluster.PrivateClusterEnabled {
			privateClusters++
		}
		if cluster.ClusterIdentityType != nil {
			identityClusters++
		}
		if cluster.AzureRBACEnabled != nil && *cluster.AzureRBACEnabled {
			azureRBACClusters++
		}
		if (cluster.OIDCIssuerEnabled != nil && *cluster.OIDCIssuerEnabled) ||
			(cluster.WorkloadIdentityEnabled != nil && *cluster.WorkloadIdentityEnabled) {
			federationClusters++
		}
	}

	return fmt.Sprintf(
		"%d AKS clusters visible; %d use private API endpoints, %d expose cluster identity context, %d enable Azure RBAC, and %d show Azure-side federation cues.",
		len(payload.AksClusters),
		privateClusters,
		identityClusters,
		azureRBACClusters,
		federationClusters,
	)
}

func apiMgmtTakeaway(payload models.ApiMgmtOutput) string {
	publicNetwork := 0
	identities := 0
	namedValueCounts := []*int{}
	readableNamedValues := 0
	secretCounts := []*int{}
	readableSecretNamedValues := 0

	for _, service := range payload.ApiManagementServices {
		if strings.EqualFold(valueOrFallback(service.PublicNetworkAccess, ""), "Enabled") {
			publicNetwork++
		}
		if service.WorkloadIdentityType != nil {
			identities++
		}
		namedValueCounts = append(namedValueCounts, service.NamedValueCount)
		if service.NamedValueCount != nil {
			readableNamedValues += *service.NamedValueCount
		}
		secretCounts = append(secretCounts, service.NamedValueSecretCount)
		if service.NamedValueSecretCount != nil {
			readableSecretNamedValues += *service.NamedValueSecretCount
		}
	}

	namedValuePhrase := intText(readableNamedValues) + " named values are visible"
	hasUnknownNamedValues := false
	for _, count := range namedValueCounts {
		if count == nil {
			hasUnknownNamedValues = true
			break
		}
	}
	if hasUnknownNamedValues {
		if readableNamedValues > 0 {
			namedValuePhrase = "at least " + intText(readableNamedValues) + " named values are visible, with some services outside current credential visibility"
		} else {
			namedValuePhrase = "current credentials do not show named values on at least one visible service"
		}
	}

	secretPhrase := ""
	allSecretCountsReadable := len(secretCounts) > 0
	for _, count := range secretCounts {
		if count == nil {
			allSecretCountsReadable = false
			break
		}
	}
	if allSecretCountsReadable {
		secretPhrase = ", including " + intText(readableSecretNamedValues) + " marked secret"
	}

	return intText(len(payload.ApiManagementServices)) + " API Management services visible; " +
		intText(publicNetwork) + " keep public network access enabled, " +
		intText(identities) + " carry managed identity context, and " +
		namedValuePhrase + secretPhrase + "."
}

func workloadIdentityContext(workload models.WorkloadSummary) string {
	parts := make([]string, 0, 2)
	if workload.IdentityType != nil && *workload.IdentityType != "" {
		parts = append(parts, *workload.IdentityType)
	}
	if len(workload.IdentityIDs) > 0 {
		parts = append(parts, fmt.Sprintf("ids=%d", len(workload.IdentityIDs)))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func vmssCapacityContext(asset models.VmssAsset) string {
	parts := make([]string, 0, 3)
	if asset.SKUName != nil && *asset.SKUName != "" {
		parts = append(parts, *asset.SKUName)
	}
	if asset.InstanceCount != nil {
		parts = append(parts, fmt.Sprintf("instances=%d", *asset.InstanceCount))
	}
	if len(asset.Zones) > 0 {
		parts = append(parts, fmt.Sprintf("zones=%d", len(asset.Zones)))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func vmssRolloutContext(asset models.VmssAsset) string {
	parts := make([]string, 0, 4)
	if asset.OrchestrationMode != nil && *asset.OrchestrationMode != "" {
		parts = append(parts, *asset.OrchestrationMode)
	}
	if asset.UpgradeMode != nil && *asset.UpgradeMode != "" {
		parts = append(parts, "upgrade="+*asset.UpgradeMode)
	}
	if asset.SinglePlacementGroup != nil {
		if *asset.SinglePlacementGroup {
			parts = append(parts, "spg=yes")
		} else {
			parts = append(parts, "spg=no")
		}
	}
	if asset.Overprovision != nil {
		if *asset.Overprovision {
			parts = append(parts, "overprov=yes")
		} else {
			parts = append(parts, "overprov=no")
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func vmssIdentityContext(asset models.VmssAsset) string {
	parts := make([]string, 0, 2)
	if asset.IdentityType != nil && *asset.IdentityType != "" {
		parts = append(parts, *asset.IdentityType)
	}
	if len(asset.IdentityIDs) > 0 {
		parts = append(parts, fmt.Sprintf("ids=%d", len(asset.IdentityIDs)))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func vmssFrontendContext(asset models.VmssAsset) string {
	parts := make([]string, 0, 4)
	if asset.PublicIPConfigurationCount > 0 {
		parts = append(parts, fmt.Sprintf("public-ip=%d", asset.PublicIPConfigurationCount))
	}
	if asset.InboundNATPoolCount > 0 {
		parts = append(parts, fmt.Sprintf("nat-pools=%d", asset.InboundNATPoolCount))
	}
	if asset.LoadBalancerBackendPoolCount > 0 {
		parts = append(parts, fmt.Sprintf("lb-backends=%d", asset.LoadBalancerBackendPoolCount))
	}
	if asset.ApplicationGatewayBackendPoolCount > 0 {
		parts = append(parts, fmt.Sprintf("appgw=%d", asset.ApplicationGatewayBackendPoolCount))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func vmssNetworkContext(asset models.VmssAsset) string {
	parts := make([]string, 0, 3)
	if asset.NICConfigurationCount > 0 {
		parts = append(parts, fmt.Sprintf("nic-configs=%d", asset.NICConfigurationCount))
	}
	subnetNames := make([]string, 0, len(asset.SubnetIDs))
	for _, subnetID := range asset.SubnetIDs {
		if subnetID == "" {
			continue
		}
		value := subnetID
		subnetNames = append(subnetNames, nicDisplayResourceName(&value))
	}
	if len(subnetNames) > 0 {
		parts = append(parts, "subnet="+strings.Join(subnetNames, ","))
	}
	if asset.ZoneBalance != nil {
		if *asset.ZoneBalance {
			parts = append(parts, "zone-balance=yes")
		} else {
			parts = append(parts, "zone-balance=no")
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func nicDisplayResourceName(resourceID *string) string {
	if resourceID == nil || *resourceID == "" {
		return ""
	}
	parts := strings.Split(strings.TrimRight(*resourceID, "/"), "/")
	if len(parts) == 0 {
		return *resourceID
	}
	return parts[len(parts)-1]
}

func nicDisplayResourceRefs(resourceIDs []string) string {
	names := make([]string, 0, len(resourceIDs))
	for _, resourceID := range resourceIDs {
		if resourceID == "" {
			continue
		}
		value := resourceID
		names = append(names, nicDisplayResourceName(&value))
	}
	return join(names, ", ")
}

func nicNetworkScopeSummary(nic models.NicAsset) string {
	subnets := make([]string, 0, len(nic.SubnetIDs))
	for _, subnetID := range nic.SubnetIDs {
		if subnetID == "" {
			continue
		}
		value := subnetID
		subnets = append(subnets, nicDisplayResourceName(&value))
	}
	vnets := make([]string, 0, len(nic.VnetIDs))
	for _, vnetID := range nic.VnetIDs {
		if vnetID == "" {
			continue
		}
		value := vnetID
		vnets = append(vnets, nicDisplayResourceName(&value))
	}
	return join(subnets, ", ") + " / " + join(vnets, ", ")
}

func whoAmITable(payload models.WhoAmIOutput) string {
	headerStyle := lipgloss.NewStyle().Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true)

	rows := [][]string{
		{"tenant_id", payload.TenantID},
		{"subscription_id", payload.Subscription.ID},
		{"subscription_display_name", payload.Subscription.DisplayName},
		{"subscription_state", payload.Subscription.State},
		{"principal_id", payload.Principal.ID},
		{"principal_type", payload.Principal.PrincipalType},
		{"principal_display_name", payload.Principal.DisplayName},
		{"effective_scope_ids", joinScopes(payload.EffectiveScopes, "id")},
		{"effective_scope_display_names", joinScopes(payload.EffectiveScopes, "display_name")},
		{"token_source", valueOrEmpty(payload.Metadata.TokenSource)},
		{"auth_mode", valueOrEmpty(payload.Metadata.AuthMode)},
		{"schema_version", payload.Metadata.SchemaVersion},
	}

	table := liptable.New().
		Headers("field", "value").
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == liptable.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	return titleStyle.Render(fmt.Sprintf("azurefox %s", payload.Metadata.Command)) + "\n\n" + strings.TrimRight(table.String(), "\n") + "\n"
}

func permissionsTable(payload models.PermissionsOutput) string {
	headerStyle := lipgloss.NewStyle().Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true)

	rows := make([][]string, 0, len(payload.Permissions))
	for _, permission := range payload.Permissions {
		rows = append(rows, []string{
			permission.Priority,
			permission.DisplayName,
			permission.PrincipalType,
			join(permission.HighImpactRoles, ";"),
			fmt.Sprintf("%d", permission.ScopeCount),
			permission.OperatorSignal,
			permission.NextReview,
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"", "no records", "", "", "", "", ""})
	}

	table := liptable.New().
		Headers("priority", "principal", "type", "high-impact roles", "scopes", "operator signal", "next review").
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == liptable.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	return titleStyle.Render("azurefox permissions") + "\n\n" + strings.TrimRight(table.String(), "\n") + "\n"
}

func principalsTable(payload models.PrincipalsOutput) string {
	rows := make([][]string, 0, len(payload.Principals))
	for _, principal := range payload.Principals {
		rows = append(rows, []string{
			valueOrFallback(principal.DisplayName, principal.ID),
			principal.PrincipalType,
			join(principal.RoleNames, ";"),
			intString(principal.RoleAssignmentCount),
			principalsIdentityContext(principal),
			join(principal.Sources, ";"),
			yesNo(principal.IsCurrentIdentity),
		})
	}

	return renderListTable(
		"azurefox principals",
		[]string{"principal", "type", "roles", "assignments", "identity context", "sources", "current"},
		rows,
		[]string{"No visible principals were confirmed from current scope.", "", "", "", "", "", ""},
		principalsTakeaway(payload),
	)
}

func privescTable(payload models.PrivescOutput) string {
	rows := make([][]string, 0, len(payload.Paths))
	for _, path := range payload.Paths {
		rows = append(rows, []string{
			path.Priority,
			path.StartingFoothold,
			privescPathLabel(path.PathType),
			privescTarget(path),
			path.OperatorSignal,
			privescProofBoundary(path),
			path.NextReview,
		})
	}

	return renderListTable(
		"azurefox privesc",
		[]string{"priority", "starting foothold", "path type", "target", "operator signal", "note", "next review"},
		rows,
		[]string{"No visible privilege-escalation paths were confirmed from current scope.", "", "", "", "", "", ""},
		privescTakeaway(payload),
	)
}

func lighthouseTable(payload models.LighthouseOutput) string {
	rows := make([][]string, 0, len(payload.LighthouseDelegations))
	for _, delegation := range payload.LighthouseDelegations {
		rows = append(rows, []string{
			lighthouseScopeContext(delegation),
			valueOrFallback(delegation.ManagedByTenantName, valueOrFallback(delegation.ManagedByTenantID, "-")),
			valueOrFallback(delegation.ManageeTenantName, valueOrFallback(delegation.ManageeTenantID, "-")),
			lighthouseAccessContext(delegation),
			lighthouseStateContext(delegation),
			delegation.Summary,
		})
	}

	return renderListTable(
		"azurefox lighthouse",
		[]string{"scope", "managing tenant", "managed tenant", "access", "state", "why it matters"},
		rows,
		[]string{"No visible Azure Lighthouse delegations were confirmed from current scope.", "", "", "", "", ""},
		lighthouseTakeaway(payload),
	)
}

func crossTenantTable(payload models.CrossTenantOutput) string {
	rows := make([][]string, 0, len(payload.CrossTenantPaths))
	for _, path := range payload.CrossTenantPaths {
		rows = append(rows, []string{
			path.Name,
			path.SignalType,
			crossTenantTenantContext(path),
			valueOrFallback(path.Scope, "-"),
			crossTenantPostureContext(path),
			crossTenantAttackPathContext(path),
			path.Summary,
		})
	}

	return renderListTable(
		"azurefox cross-tenant",
		[]string{"signal", "type", "tenant", "scope", "posture", "attack path", "why it matters"},
		rows,
		[]string{"No visible cross-tenant signals were confirmed from current scope.", "", "", "", "", "", ""},
		crossTenantTakeaway(payload),
	)
}

func authPoliciesTable(payload models.AuthPoliciesOutput) string {
	rows := make([][]string, 0, len(payload.AuthPolicies))
	for _, policy := range payload.AuthPolicies {
		rows = append(rows, []string{
			policy.Name,
			policy.State,
			valueOrFallback(policy.Scope, "-"),
			policy.Summary,
		})
	}

	return renderListTable(
		"azurefox auth-policies",
		[]string{"policy", "state", "scope", "operator signal"},
		rows,
		[]string{"No visible auth policy rows were confirmed from current scope.", "", "", ""},
		authPoliciesTakeaway(payload),
	)
}

func resourceTrustsTable(payload models.ResourceTrustsOutput) string {
	rows := make([][]string, 0, len(payload.ResourceTrusts))
	for _, trust := range payload.ResourceTrusts {
		rows = append(rows, []string{
			resourceTrustDisplayName(trust),
			trust.ResourceType,
			trust.TrustType,
			trust.Target,
			trust.Exposure,
			trust.Summary,
		})
	}

	return renderListTable(
		"azurefox resource-trusts",
		[]string{"resource", "type", "trust", "target", "exposure", "why it matters"},
		rows,
		[]string{"No visible resource trust surfaces were confirmed from current scope.", "", "", "", "", ""},
		resourceTrustsTakeaway(payload),
	)
}

func roleTrustsTable(payload models.RoleTrustsOutput) string {
	headerStyle := lipgloss.NewStyle().Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true)

	rows := make([][]string, 0, len(payload.Trusts))
	for _, trust := range payload.Trusts {
		rows = append(rows, []string{
			trust.TrustType,
			valueOrEmpty(trust.SourceName),
			valueOrEmpty(trust.TargetName),
			trust.Confidence,
			valueOrEmpty(trust.OperatorSignal),
			valueOrEmpty(trust.NextReview),
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"", "no trust edges visible", "", "", "", ""})
	}

	table := liptable.New().
		Headers("trust", "source", "target", "confidence", "operator signal", "next review").
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == liptable.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	body := titleStyle.Render("azurefox role-trusts") + "\n\n" + strings.TrimRight(table.String(), "\n") + "\n"
	if len(payload.Trusts) == 0 {
		return body
	}
	return body + "\nTakeaway: " + roleTrustsTakeaway(payload) + "\n"
}

func resourceTrustsTakeaway(payload models.ResourceTrustsOutput) string {
	exposures := map[string]int{}
	for _, trust := range payload.ResourceTrusts {
		exposure := trust.Exposure
		if exposure == "" {
			exposure = "unknown"
		}
		exposures[exposure]++
	}
	if len(exposures) == 0 {
		return "0 resource trust surfaces visible; no resource trust surfaces visible."
	}

	names := make([]string, 0, len(exposures))
	for name := range exposures {
		names = append(names, name)
	}
	sort.Strings(names)

	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%d %s", exposures[name], name))
	}
	return fmt.Sprintf(
		"%d resource trust surfaces visible; %s.",
		len(payload.ResourceTrusts),
		strings.Join(parts, ", "),
	)
}

func resourceTrustDisplayName(trust models.ResourceTrustSummary) string {
	if trust.ResourceName != "" {
		return trust.ResourceName
	}
	return trust.ResourceID
}

func principalsIdentityContext(principal models.PrincipalSummary) string {
	parts := []string{}
	if len(principal.IdentityNames) > 0 {
		parts = append(parts, "names="+join(principal.IdentityNames, ";"))
	}
	if len(principal.IdentityTypes) > 0 {
		parts = append(parts, "types="+join(principal.IdentityTypes, ";"))
	}
	if len(principal.AttachedTo) > 0 {
		parts = append(parts, "attached="+intString(len(principal.AttachedTo)))
	}
	if len(parts) == 0 {
		return "none"
	}
	return join(parts, "; ")
}

func lighthouseTakeaway(payload models.LighthouseOutput) string {
	delegations := payload.LighthouseDelegations
	subscriptionScope := 0
	eligible := 0
	broadRoles := 0
	for _, delegation := range delegations {
		if delegation.ScopeType == "subscription" {
			subscriptionScope++
		}
		if delegation.EligibleAuthorizationCount > 0 {
			eligible++
		}
		if delegation.HasOwnerRole || delegation.HasUserAccessAdministrator {
			broadRoles++
		}
	}
	return fmt.Sprintf(
		"%d Azure Lighthouse delegation(s) visible; %d are subscription-scoped, %d grant Owner or User Access Administrator, and %d include eligible access.",
		len(delegations),
		subscriptionScope,
		broadRoles,
		eligible,
	)
}

func crossTenantTakeaway(payload models.CrossTenantOutput) string {
	high := 0
	lighthouse := 0
	externalSP := 0
	policy := 0
	for _, path := range payload.CrossTenantPaths {
		if strings.EqualFold(strings.TrimSpace(path.Priority), "high") {
			high++
		}
		switch strings.ToLower(strings.TrimSpace(path.SignalType)) {
		case "lighthouse":
			lighthouse++
		case "external-sp":
			externalSP++
		case "policy":
			policy++
		}
	}
	return fmt.Sprintf(
		"%d cross-tenant signal(s) visible; %d high priority, %d delegated management, %d externally owned service principal, and %d tenant policy cue.",
		len(payload.CrossTenantPaths),
		high,
		lighthouse,
		externalSP,
		policy,
	)
}

func lighthouseScopeContext(item models.LighthouseDelegationAsset) string {
	scopeLabel := valueOrFallback(item.ScopeDisplayName, displayResourceRef(item.ScopeID))
	if item.ScopeType == "resource_group" {
		return "resource-group::" + scopeLabel
	}
	return "subscription::" + scopeLabel
}

func lighthouseAccessContext(item models.LighthouseDelegationAsset) string {
	parts := []string{}
	if strongestRole := valueOrEmpty(item.StrongestRoleName); strongestRole != "" {
		parts = append(parts, "strongest="+strongestRole)
	}
	parts = append(parts, "auth="+intString(item.AuthorizationCount))
	parts = append(parts, "eligible="+intString(item.EligibleAuthorizationCount))
	if item.HasDelegatedRoleAssignments {
		parts = append(parts, "delegated-role-assign=yes")
	}
	if planName := valueOrEmpty(item.PlanName); planName != "" {
		parts = append(parts, "plan="+planName)
	}
	if len(parts) == 0 {
		return "-"
	}
	return join(parts, "; ")
}

func crossTenantTenantContext(item models.CrossTenantPathSummary) string {
	tenantName := valueOrEmpty(item.TenantName)
	tenantID := valueOrEmpty(item.TenantID)
	switch {
	case tenantName != "" && tenantID != "":
		return tenantName + " (" + tenantID + ")"
	case tenantName != "":
		return tenantName
	case tenantID != "":
		return tenantID
	default:
		return "-"
	}
}

func crossTenantPostureContext(item models.CrossTenantPathSummary) string {
	parts := []string{}
	if item.Priority != "" {
		parts = append(parts, "priority="+item.Priority)
	}
	if posture := valueOrEmpty(item.Posture); posture != "" {
		parts = append(parts, posture)
	}
	if len(parts) == 0 {
		return "-"
	}
	return join(parts, "; ")
}

func crossTenantAttackPathContext(item models.CrossTenantPathSummary) string {
	attackPath := item.AttackPath
	signalType := item.SignalType
	switch {
	case attackPath != "" && signalType != "":
		return attackPath + " via " + signalType
	case attackPath != "":
		return attackPath
	case signalType != "":
		return signalType
	default:
		return "-"
	}
}

func lighthouseStateContext(item models.LighthouseDelegationAsset) string {
	parts := []string{}
	assignmentState := valueOrEmpty(item.ProvisioningState)
	if assignmentState != "" {
		parts = append(parts, "assignment="+assignmentState)
	}
	definitionState := valueOrEmpty(item.DefinitionProvisioningState)
	if definitionState != "" && definitionState != assignmentState {
		parts = append(parts, "definition="+definitionState)
	}
	if len(parts) == 0 {
		return "-"
	}
	return join(parts, "; ")
}

func principalsTakeaway(payload models.PrincipalsOutput) string {
	current := 0
	privileged := 0
	for _, principal := range payload.Principals {
		if principal.IsCurrentIdentity {
			current++
		}
		for _, role := range principal.RoleNames {
			if strings.EqualFold(role, "Owner") {
				privileged++
				break
			}
		}
	}
	return intString(len(payload.Principals)) + " principals visible; " + intString(privileged) + " hold Owner and " + intString(current) + " match the current identity."
}

func privescProofBoundary(path models.PrivescPathSummary) string {
	proven := strings.TrimSpace(path.ProvenPath)
	missing := strings.TrimSpace(path.MissingProof)
	if proven != "" && missing != "" {
		return proven + "\n\n" + missing
	}
	if proven != "" {
		return proven
	}
	return missing
}

func privescTarget(path models.PrivescPathSummary) string {
	if path.CurrentIdentity {
		return "current foothold"
	}
	principal := strings.TrimSpace(path.Principal)
	asset := strings.TrimSpace(valueOrEmpty(path.Asset))
	if principal != "" && asset != "" {
		return principal + " via " + asset
	}
	if principal != "" {
		return principal
	}
	if asset != "" {
		return asset
	}
	return "-"
}

func privescPathLabel(pathType string) string {
	switch strings.TrimSpace(pathType) {
	case "current-foothold-direct-control":
		return "current foothold direct control"
	case "visible-privileged-lead":
		return "visible privileged lead"
	case "ingress-backed-workload-identity":
		return "ingress-backed workload identity"
	default:
		normalized := strings.TrimSpace(pathType)
		if normalized == "" {
			return "-"
		}
		return normalized
	}
}

func privescTakeaway(payload models.PrivescOutput) string {
	rooted := 0
	priorityCounts := map[string]int{}
	for _, path := range payload.Paths {
		if path.CurrentIdentity {
			rooted++
		}
		priorityCounts[path.Priority]++
	}

	visibleOnly := len(payload.Paths) - rooted
	countParts := []string{}
	for _, priority := range []string{"high", "medium", "low", "unknown"} {
		if priorityCounts[priority] > 0 {
			countParts = append(countParts, fmt.Sprintf("%d %s", priorityCounts[priority], priority))
		}
	}

	counts := "no meaningful paths"
	if len(countParts) > 0 {
		counts = strings.Join(countParts, ", ")
	}

	visibleLabel := "visible-only leads"
	if visibleOnly == 1 {
		visibleLabel = "visible-only lead"
	}

	return fmt.Sprintf("%d privilege-escalation paths surfaced; %d current-identity-rooted, %d %s, %s.", len(payload.Paths), rooted, visibleOnly, visibleLabel, counts)
}

func authPoliciesTakeaway(payload models.AuthPoliciesOutput) string {
	return fmt.Sprintf("%d policy rows, %d findings, and %d current-scope issues.", len(payload.AuthPolicies), len(payload.Findings), len(payload.Issues))
}

func roleTrustsTakeaway(payload models.RoleTrustsOutput) string {
	families := map[string]int{}
	for _, trust := range payload.Trusts {
		families[trust.TrustType]++
	}

	parts := make([]string, 0, len(families))
	for _, name := range sortedFamilyKeys(families) {
		parts = append(parts, fmt.Sprintf("%d %s", families[name], name))
	}

	privilegeFollowOns := 0
	ownershipFollowOns := 0
	outsideFollowOns := 0
	for _, trust := range payload.Trusts {
		switch trust.FollowOnKind {
		case models.RoleTrustFollowOnPrivilegeConfirmation:
			privilegeFollowOns++
		case models.RoleTrustFollowOnOwnershipReview:
			ownershipFollowOns++
		case models.RoleTrustFollowOnOutsideTenant:
			outsideFollowOns++
		}
	}

	counts := "no trust edges visible"
	if len(parts) > 0 {
		counts = strings.Join(parts, ", ")
	}

	return fmt.Sprintf(
		"%d trust edges surfaced in %s mode; %s. %d privilege-confirmation follow-ons, %d ownership-review follow-ons, and %d outside-tenant follow-ons. Delegated and admin consent grants are out of scope for this command.",
		len(payload.Trusts),
		payload.Mode,
		counts,
		privilegeFollowOns,
		ownershipFollowOns,
		outsideFollowOns,
	)
}

func managedIdentitiesTable(payload models.ManagedIdentitiesOutput) string {
	headerStyle := lipgloss.NewStyle().Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true)

	rows := make([][]string, 0, len(payload.Identities))
	for _, identity := range payload.Identities {
		rows = append(rows, []string{
			identity.Name,
			identity.IdentityType,
			join(displayResourceRefs(identity.AttachedTo), ", "),
			valueOrEmpty(identity.OperatorSignal),
			valueOrEmpty(identity.NextReview),
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"", "", "no managed identities visible", "", ""})
	}

	table := liptable.New().
		Headers("identity", "type", "attached to", "operator signal", "next review").
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == liptable.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	output := titleStyle.Render("azurefox managed-identities") + "\n\n" + strings.TrimRight(table.String(), "\n") + "\n"
	if len(payload.Findings) > 0 {
		output += "\nFindings:\n"
		for _, finding := range payload.Findings {
			output += fmt.Sprintf("- %s: %s\n", strings.ToUpper(finding.Severity), finding.Title)
			output += fmt.Sprintf("  %s\n", finding.Description)
		}
	}

	return output + "\nTakeaway: " + managedIdentitiesTakeaway(payload) + "\n"
}

func managedIdentitiesTakeaway(payload models.ManagedIdentitiesOutput) string {
	exposed := 0
	directControl := 0
	for _, identity := range payload.Identities {
		if identity.WorkloadExposure == models.WorkloadExposurePublic || identity.WorkloadExposure == models.WorkloadExposureExposed {
			exposed++
		}
		if identity.DirectControlVisible {
			directControl++
		}
	}

	return fmt.Sprintf(
		"%d managed identities visible; %d exposed workload pivots and %d direct-control cues from current scope.",
		len(payload.Identities),
		exposed,
		directControl,
	)
}

func envVarsTable(payload models.EnvVarsOutput) string {
	headerStyle := lipgloss.NewStyle().Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true)

	rows := make([][]string, 0, len(payload.EnvVars))
	for _, envVar := range payload.EnvVars {
		rows = append(rows, []string{
			envVar.AssetName,
			envVar.AssetKind,
			envVarIdentityContext(envVar),
			envVar.SettingName,
			envVar.ValueType,
			envVarSignal(envVar),
			envVarNextReview(envVar),
			envVar.Summary,
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"", "", "no environment variable rows visible", "", "", "", "", ""})
	}

	table := liptable.New().
		Headers("workload", "kind", "identity", "setting", "value type", "signal", "next review", "why it matters").
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == liptable.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	output := titleStyle.Render("azurefox env-vars") + "\n\n" + strings.TrimRight(table.String(), "\n") + "\n"
	if len(payload.Findings) > 0 {
		output += "\nFindings:\n"
		for _, finding := range payload.Findings {
			output += fmt.Sprintf("- %s: %s\n", strings.ToUpper(finding.Severity), finding.Title)
			output += fmt.Sprintf("  %s\n", finding.Description)
		}
	}

	return output + "\nTakeaway: " + envVarsTakeaway(payload) + "\n"
}

func tokensCredentialsTable(payload models.TokensCredentialsOutput) string {
	headerStyle := lipgloss.NewStyle().Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true)

	rows := make([][]string, 0, len(payload.Surfaces))
	for _, surface := range payload.Surfaces {
		rows = append(rows, []string{
			surface.AssetName,
			surface.AssetKind,
			string(surface.SurfaceType),
			surface.AccessPath,
			surface.Priority,
			surface.OperatorSignal,
			tokenCredentialNextReview(surface),
			surface.Summary,
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"", "", "no token or credential surfaces visible", "", "", "", "", ""})
	}

	table := liptable.New().
		Headers("asset", "kind", "surface", "access path", "priority", "operator signal", "next review", "why it matters").
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == liptable.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	output := titleStyle.Render("azurefox tokens-credentials") + "\n\n" + strings.TrimRight(table.String(), "\n") + "\n"
	if len(payload.Findings) > 0 {
		output += "\nFindings:\n"
		for _, finding := range payload.Findings {
			output += fmt.Sprintf("- %s: %s\n", strings.ToUpper(finding.Severity), finding.Title)
			output += fmt.Sprintf("  %s\n", finding.Description)
		}
	}

	return output + "\nTakeaway: " + tokensCredentialsTakeaway(payload) + "\n"
}

func envVarsTakeaway(payload models.EnvVarsOutput) string {
	workloads := map[string]struct{}{}
	plainSensitive := 0
	keyVaultRefs := 0
	for _, envVar := range payload.EnvVars {
		if envVar.AssetID != "" {
			workloads[envVar.AssetID] = struct{}{}
		}
		if envVar.LooksSensitive && envVar.ValueType == "plain-text" {
			plainSensitive++
		}
		if envVar.ValueType == "keyvault-ref" {
			keyVaultRefs++
		}
	}

	return fmt.Sprintf(
		"%d settings across %d workloads; %d plain-text sensitive settings, %d Key Vault references, and %d findings.",
		len(payload.EnvVars),
		len(workloads),
		plainSensitive,
		keyVaultRefs,
		len(payload.Findings),
	)
}

func envVarIdentityContext(envVar models.EnvVarSummary) string {
	parts := make([]string, 0, 3)
	if envVar.WorkloadIdentityType != nil && *envVar.WorkloadIdentityType != "" {
		parts = append(parts, *envVar.WorkloadIdentityType)
	}
	if len(envVar.WorkloadIdentityIDs) > 0 {
		parts = append(parts, fmt.Sprintf("user-assigned=%d", len(envVar.WorkloadIdentityIDs)))
	}
	if envVar.KeyVaultReferenceIdentity != nil && *envVar.KeyVaultReferenceIdentity != "" {
		parts = append(parts, displayReferenceIdentity(*envVar.KeyVaultReferenceIdentity))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func envVarSignal(envVar models.EnvVarSummary) string {
	parts := make([]string, 0, 3)
	if envVar.LooksSensitive {
		parts = append(parts, "sensitive-name")
	}
	if envVar.ValueType == "keyvault-ref" {
		parts = append(parts, "keyvault-ref")
	}
	if envVar.ReferenceTarget != nil && *envVar.ReferenceTarget != "" {
		parts = append(parts, *envVar.ReferenceTarget)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func envVarNextReview(envVar models.EnvVarSummary) string {
	return envVarNextReviewHint(
		envVar.ValueType,
		envVar.LooksSensitive,
		valueOrEmpty(envVar.ReferenceTarget),
		valueOrEmpty(envVar.WorkloadIdentityType),
		envVar.TargetServices,
	)
}

func envVarNextReviewHint(valueType string, looksSensitive bool, referenceTarget string, workloadIdentityType string, targetServices []models.EnvVarTargetService) string {
	hasIdentity := workloadIdentityType != ""

	if valueType == "keyvault-ref" {
		if hasIdentity {
			return "Check keyvault for the referenced secret path; review managed-identities for the workload token path."
		}
		return "Check keyvault for the referenced secret path."
	}

	if looksSensitive && valueType == "plain-text" {
		switch firstEnvVarTargetService(targetServices) {
		case models.EnvVarTargetServiceStorage:
			return "Check tokens-credentials first; this likely feeds a storage credential path."
		case models.EnvVarTargetServiceDatabase:
			return "Check tokens-credentials first; this likely feeds a database credential path."
		default:
			return "Check tokens-credentials for the workload credential surface."
		}
	}

	switch firstEnvVarTargetService(targetServices) {
	case models.EnvVarTargetServiceStorage:
		if hasIdentity {
			return "Check tokens-credentials for the config-backed access path, then managed-identities for the workload token path."
		}
		return "Check tokens-credentials for the config-backed storage access path."
	case models.EnvVarTargetServiceDatabase:
		return "Check tokens-credentials for the config-backed database access path."
	}

	if referenceTarget != "" && strings.Contains(strings.ToLower(referenceTarget), "vault") {
		return "Check keyvault for the referenced secret path."
	}
	if hasIdentity {
		return "Check managed-identities for the workload token path behind this setting."
	}
	return "Review the workload config directly before deeper follow-up."
}

func tokenCredentialNextReview(surface models.TokenCredentialSurfaceSummary) string {
	switch surface.NextReviewKind {
	case models.TokenCredentialReviewEnvVarsSettingContext:
		return "Check env-vars for the exact setting context behind this credential clue."
	case models.TokenCredentialReviewEndpointsIngressAndControl:
		return "Check endpoints for the ingress path, then managed-identities and permissions for Azure control."
	case models.TokenCredentialReviewManagedIdentityAndPermissions:
		return "Check managed-identities for the identity path, then permissions for Azure control."
	case models.TokenCredentialReviewKeyVaultAndManagedIdentity:
		return "Check keyvault for the referenced secret boundary, then managed-identities for the backing workload identity."
	case models.TokenCredentialReviewKeyVaultBoundary:
		return "Check keyvault for the referenced secret boundary."
	case models.TokenCredentialReviewARMDeploymentOutputs:
		return "Check arm-deployments for the exact output context behind this credential clue."
	case models.TokenCredentialReviewARMDeploymentLinks:
		return "Check arm-deployments for the linked template or parameter path behind this credential clue."
	default:
		return "Review the surfaced workload context before deeper follow-up."
	}
}

func tokensCredentialsTakeaway(payload models.TokensCredentialsOutput) string {
	assets := map[string]struct{}{}
	families := map[string]int{}
	for _, surface := range payload.Surfaces {
		if surface.AssetID != "" {
			assets[surface.AssetID] = struct{}{}
		}
		families[string(surface.SurfaceType)]++
	}

	parts := make([]string, 0, len(families))
	for _, name := range sortedFamilyKeys(families) {
		parts = append(parts, fmt.Sprintf("%d %s", families[name], name))
	}

	counts := "no surfaces visible"
	if len(parts) > 0 {
		counts = strings.Join(parts, ", ")
	}

	return fmt.Sprintf(
		"%d token or credential surfaces across %d assets; %s and %d findings.",
		len(payload.Surfaces),
		len(assets),
		counts,
		len(payload.Findings),
	)
}

func firstEnvVarTargetService(targetServices []models.EnvVarTargetService) models.EnvVarTargetService {
	if len(targetServices) == 0 {
		return models.EnvVarTargetServiceNone
	}
	return targetServices[0]
}

func displayReferenceIdentity(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return "-"
	}
	if strings.EqualFold(text, "SystemAssigned") {
		return "kv-ref=SystemAssigned"
	}
	parts := strings.Split(text, "/")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	if len(filtered) > 0 {
		return "kv-ref=" + filtered[len(filtered)-1]
	}
	return "kv-ref=" + text
}

func displayResourceRefs(ids []string) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, displayResourceRef(id))
	}
	return values
}

func displayResourceRef(id string) string {
	trimmed := strings.TrimRight(id, "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	return parts[len(parts)-1]
}

func sortedFamilyKeys(families map[string]int) []string {
	keys := make([]string, 0, len(families))
	for key := range families {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func topResourceType(resourceTypes models.TopResourceTypes) (string, int) {
	if len(resourceTypes) == 0 {
		return "none visible", 0
	}
	bestKey := ""
	bestValue := -1
	for _, key := range sortedResourceTypeKeys(resourceTypes) {
		value := resourceTypes[key]
		if value > bestValue {
			bestKey = key
			bestValue = value
		}
	}
	return bestKey, bestValue
}

func sortedResourceTypeKeys(resourceTypes models.TopResourceTypes) []string {
	keys := make([]string, 0, len(resourceTypes))
	for key := range resourceTypes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func armDeploymentsTakeaway(payload models.ArmDeploymentsOutput) string {
	subscriptionScope := 0
	for _, deployment := range payload.Deployments {
		if deployment.ScopeType == "subscription" {
			subscriptionScope++
		}
	}

	return fmt.Sprintf(
		"%d deployments visible; %d at subscription scope and %d findings.",
		len(payload.Deployments),
		subscriptionScope,
		len(payload.Findings),
	)
}

func armDeploymentScopeLabel(deployment models.ArmDeploymentSummary) string {
	if deployment.ResourceGroup != nil && *deployment.ResourceGroup != "" {
		return "rg:" + *deployment.ResourceGroup
	}
	scope := strings.TrimRight(deployment.Scope, "/")
	if strings.Contains(scope, "/subscriptions/") {
		subscriptionID := strings.SplitN(strings.SplitN(scope, "/subscriptions/", 2)[1], "/", 2)[0]
		if subscriptionID != "" {
			return "sub:" + subscriptionID
		}
	}
	if scope != "" {
		return scope
	}
	if deployment.ScopeType != "" {
		return deployment.ScopeType
	}
	return "-"
}

func armDeploymentLinkedReferenceSummary(deployment models.ArmDeploymentSummary) string {
	parts := make([]string, 0, 2)
	if deployment.TemplateLink != nil && *deployment.TemplateLink != "" {
		parts = append(parts, "template="+displayLink(*deployment.TemplateLink))
	}
	if deployment.ParametersLink != nil && *deployment.ParametersLink != "" {
		parts = append(parts, "parameters="+displayLink(*deployment.ParametersLink))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ", ")
}

func displayLink(value string) string {
	text := strings.TrimSpace(value)
	text = strings.TrimPrefix(text, "https://")
	text = strings.TrimPrefix(text, "http://")
	return text
}

func join(values []string, separator string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, separator)
}
