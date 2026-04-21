package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	liptable "github.com/charmbracelet/lipgloss/table"

	"harrierops-azure/internal/models"
)

const findingNoteWrapWidth = 100

var commandNarration = map[string]string{
	"whoami":             "Checking caller context and active subscription scope.",
	"inventory":          "Scoping the visible Azure resource footprint.",
	"automation":         "Reviewing Azure Automation accounts for identity, execution, webhook, worker, and secure-asset posture.",
	"app-credentials":    "Reviewing application and service-principal authentication material, federated trust, and visible current-identity control paths.",
	"devops":             "Reviewing Azure DevOps build definitions for trusted source inputs, visible injection surfaces, and Azure-facing change paths.",
	"app-services":       "Reviewing App Service runtime, hostname, identity, and ingress cues that change follow-on paths.",
	"acr":                "Reviewing Azure Container Registry login, auth, network, and registry automation/trust cues.",
	"databases":          "Reviewing relational database server posture across Azure SQL, PostgreSQL Flexible, and MySQL Flexible.",
	"dns":                "Reviewing public and private DNS zone inventory and namespace boundaries.",
	"aks":                "Reviewing AKS control-plane endpoint, identity, auth posture, and Azure-side federation and addon cues.",
	"api-mgmt":           "Reviewing API Management gateway hostnames, identity, subscription, backend, and secret posture.",
	"functions":          "Reviewing Function App runtime, storage binding, identity, and deployment posture.",
	"azure-ml":           "Reviewing Azure ML runtime, scheduling, endpoint, identity, and storage-linked workspace posture.",
	"event-grid":         "Reviewing Event Grid trigger routes, destination types, and visible execution-capable follow-on paths.",
	"logic-apps":         "Reviewing Logic Apps trigger posture, identity context, and safe downstream action relationships.",
	"arm-deployments":    "Reviewing ARM deployment history for config exposure and linked content.",
	"endpoints":          "Mapping reachable IP and hostname surfaces from compute and web workloads.",
	"network-effective":  "Prioritizing likely public-IP reachability by combining visible endpoint and NSG evidence.",
	"env-vars":           "Reviewing App Service and Function App settings for exposed config paths and likely credential or secret follow-on.",
	"network-ports":      "Tracing likely inbound port exposure from visible NIC and subnet NSG rules.",
	"tokens-credentials": "Correlating token-minting workloads, credential-bearing metadata paths, and the next likely follow-on.",
	"rbac":               "Collecting raw RBAC assignments across the current subscription.",
	"principals":         "Mapping visible principals, identity footholds, and follow-on candidates.",
	"permissions":        "Ranking principals by high-impact RBAC exposure and the next likely follow-on.",
	"privesc":            "Triage likely privilege-escalation and workload identity abuse paths.",
	"role-trusts":        "Reviewing high-signal identity trust edges and the clearest next review without implying delegated or admin consent.",
	"cross-tenant":       "Reviewing outside-tenant trust, delegated management, and tenant policy cues that most change control or pivot paths.",
	"lighthouse":         "Reviewing Azure Lighthouse delegations for cross-tenant management scope and high-impact access cues.",
	"auth-policies":      "Reviewing tenant auth controls that widen guest, consent, app-creation, or sign-in abuse paths.",
	"managed-identities": "Mapping workload-linked managed identities and their visible privilege cues.",
	"keyvault":           "Reviewing Key Vault exposure, access-model weakness, and destructive leverage cues.",
	"resource-trusts":    "Correlating resource trust surfaces across public network and private-link paths.",
	"storage":            "Checking storage exposure and network posture for likely data targets.",
	"snapshots-disks":    "Reviewing managed disks and snapshots for offline-copy, sharing/export, and encryption posture with highest-value targets first.",
	"nics":               "Enumerating NIC attachments, IP context, and network boundary references.",
	"workloads":          "Joining workload assets with identity context and visible ingress paths.",
	"vms":                "Summarizing reachable compute assets and identity-bearing workloads.",
	"vmss":               "Reviewing Virtual Machine Scale Sets (VMSS) for fleet posture, identity, and frontend network cues.",
	"chains":             "Correlating grouped chain evidence with conservative cross-command joins.",
	"persistence":        "Walking the current identity through Azure-native persistence surfaces one service at a time.",
}

var commandCompactIntroHint = map[string]string{
	"aks":                "table view is compact by design; the JSON artifact keeps the fuller visible field set",
	"acr":                "table view is compact by design; the JSON artifact keeps the fuller visible field set",
	"api-mgmt":           "table view is compact by design; the JSON artifact keeps the fuller visible field set",
	"env-vars":           "table view is compact by design; the JSON artifact keeps the fuller visible field set",
	"functions":          "table view is compact by design; the JSON artifact keeps the fuller visible field set",
	"azure-ml":           "table view is compact by design; the JSON artifact keeps the fuller visible field set",
	"event-grid":         "table view is compact by design; the JSON artifact keeps the fuller visible field set",
	"logic-apps":         "table view is compact by design; the JSON artifact keeps the fuller visible field set",
	"tokens-credentials": "table view is compact by design; the JSON artifact keeps the fuller visible field set",
}

var chainsFamilyTableRenderers = map[string]func(models.ChainsOutput) string{
	"compute-control": chainsComputeControlTable,
	"credential-path": chainsCredentialPathTable,
	"deployment-path": chainsDeploymentPathTable,
	"escalation-path": chainsEscalationPathTable,
}

func chainsTableRenderer(payload any) (string, error) {
	switch out := payload.(type) {
	case models.ChainsOverviewOutput:
		return chainsOverviewTable(out), nil
	case models.ChainsOutput:
		return chainsFamilyTable(out), nil
	default:
		return "", fmt.Errorf("unexpected payload type for chains: %T", payload)
	}
}

func Table(command string, payload any, context models.RenderContext) (string, error) {
	entry, err := renderRegistryEntry(command)
	if err != nil {
		return "", err
	}
	if entry.table == nil {
		return "", fmt.Errorf("table rendering is not implemented for command %q", command)
	}
	body, err := entry.table(payload)
	if err != nil {
		return "", err
	}
	if !commandSuppressesBottomFindings(command) {
		body = appendPayloadFindingsSection(body, payload)
	}
	body = appendPayloadIssuesSection(body, payload)
	return renderTableDocument(command, context, body), nil
}

func commandSuppressesBottomFindings(command string) bool {
	switch command {
	case "tokens-credentials":
		return true
	default:
		return false
	}
}

func renderStructuredTable(title string, headers []string, rows [][]string) string {
	return renderStructuredTableWithTitle(title, headers, rows, true)
}

func renderStructuredTableWithTitle(title string, headers []string, rows [][]string, includeTitle bool) string {
	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := cellStyle.Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true)

	table := liptable.New().
		Headers(headers...).
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == liptable.HeaderRow {
				return headerStyle
			}
			return cellStyle
		})

	body := strings.TrimRight(table.String(), "\n") + "\n"
	if !includeTitle {
		return body
	}
	return titleStyle.Render(title) + "\n\n" + body
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

func renderTableDocument(command string, context models.RenderContext, body string) string {
	prelude := renderTablePrelude(command, context)
	if prelude == "" {
		return body
	}
	return prelude + body
}

func renderTablePrelude(command string, context models.RenderContext) string {
	narration := commandNarration[command]
	if narration == "" {
		narration = "Running command."
	}
	lines := []string{
		"HO-Azure :: attack-path-focused Azure recon",
		fmt.Sprintf(
			"context :: tenant=%s subscription=%s output=table",
			tableContextValue(context.Tenant),
			tableContextValue(context.Subscription),
		),
		"",
		fmt.Sprintf("[%s] %s", command, narration),
	}
	if compactHint := commandCompactIntroHint[command]; compactHint != "" {
		lines = append(lines, compactHint)
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func tableContextValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "auto"
	}
	return value
}

func rbacTable(payload models.RbacOutput) string {
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
	return renderListTable(
		"ho-azure rbac",
		[]string{"id", "principal id", "principal type", "role definition id", "role name", "scope id"},
		rows,
		[]string{"no records", "", "", "", "", ""},
		rbacTakeaway(payload),
	)
}

func inventoryTable(payload models.InventoryOutput) string {
	topType := "none visible"
	for _, key := range sortedResourceTypeKeys(payload.TopResourceTypes) {
		topType = key
		break
	}

	rows := [][]string{
		{
			fmt.Sprintf("%d", payload.ResourceGroupCount),
			fmt.Sprintf("%d", payload.ResourceCount),
			topType,
			fmt.Sprintf("%d", len(payload.Issues)),
		},
	}
	return renderListTable(
		"ho-azure inventory",
		[]string{"resource groups", "resources", "top type", "issues"},
		rows,
		[]string{"0", "0", "none visible", "0"},
		inventoryTakeaway(payload),
	)
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
	return renderListTable("ho-azure app-services", []string{
		"app service", "hostname", "runtime", "identity", "exposure", "posture", "why it matters",
	}, rows, []string{"no App Service apps visible", "", "", "", "", "", ""}, appServicesTakeaway(payload))
}

func functionsTable(payload models.FunctionsOutput) string {
	if len(payload.FunctionApps) == 0 {
		return renderListTable(
			"ho-azure functions",
			[]string{"function app", "hostname", "runtime", "identity", "deployment"},
			nil,
			[]string{"no Function Apps visible", "", "", "", ""},
			functionsTakeaway(payload),
		)
	}

	sections := make([]string, 0, len(payload.FunctionApps))
	for _, app := range payload.FunctionApps {
		row := renderStructuredTableWithTitle(
			"",
			[]string{"function app", "hostname", "runtime", "identity", "deployment"},
			[][]string{{
				app.Name,
				valueOrEmpty(app.DefaultHostname),
				functionRuntimeContext(app),
				resourceIdentityContext(app.WorkloadIdentityType, app.WorkloadIdentityIDs),
				functionDeploymentContext(app),
			}},
			false,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if note := strings.TrimSpace(app.Summary); note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}

	return joinRenderedBlocks(sections) + "\n\nTakeaway: " + functionsTakeaway(payload) + "\n"
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
	return renderListTable("ho-azure container-apps", []string{
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
	return renderListTable("ho-azure container-instances", []string{
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

	output := renderStructuredTable("ho-azure arm-deployments", []string{
		"deployment", "scope", "state", "outputs", "linked refs", "why it matters",
	}, rows)
	output = appendCustomFindingsSection(output, payload.Findings,
		func(f models.ArmDeploymentFinding) string { return f.Severity },
		func(f models.ArmDeploymentFinding) string { return f.Title },
		func(f models.ArmDeploymentFinding) string { return f.Description },
	)
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

	return renderStructuredTable("ho-azure endpoints", []string{
		"endpoint", "asset", "kind", "family", "ingress", "why it matters",
	}, rows) + "\nTakeaway: " + endpointsTakeaway(payload) + "\n"
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

	return renderStructuredTable("ho-azure network-ports", []string{
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

	return renderStructuredTable("ho-azure network-effective", []string{
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
			stringOrFallback(nicDisplayResourceName(nic.AttachedAssetID), "-"),
			join(nic.PrivateIPs, ", "),
			stringOrFallback(nicDisplayResourceRefs(nic.PublicIPIDs), "-"),
			nicNetworkScopeSummary(nic),
			stringOrFallback(nicDisplayResourceName(nic.NetworkSecurityGroupID), "-"),
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"no nic assets visible", "", "", "", "", ""})
	}

	return renderStructuredTable("ho-azure nics", []string{
		"nic", "attached asset", "private ips", "public ip refs", "subnet / vnet", "nsg",
	}, rows) + "\nTakeaway: " + nicsTakeaway(payload) + "\n"
}

func vmsTable(payload models.VmsOutput) string {
	rows := make([][]string, 0, len(payload.VMAssets))
	for _, vm := range payload.VMAssets {
		rows = append(rows, []string{
			vm.Name,
			vm.VMType,
			join(vm.PublicIPs, ", "),
			join(vm.PrivateIPs, ", "),
			join(vm.IdentityIDs, ", "),
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"no compute assets visible", "", "", "", ""})
	}

	output := renderStructuredTable("ho-azure vms", []string{
		"asset", "type", "public ips", "private ips", "identities",
	}, rows)
	output = appendCustomFindingsSection(output, payload.Findings,
		func(f models.VmsFinding) string { return f.Severity },
		func(f models.VmsFinding) string { return f.Title },
		func(f models.VmsFinding) string { return f.Description },
	)
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

	return renderStructuredTable("ho-azure vmss", []string{
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

	return renderStructuredTable("ho-azure workloads", []string{
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

func aksOperatorSignal(cluster models.AksClusterAsset) string {
	parts := []string{}
	if version := aksVersionContext(cluster); version != "-" {
		parts = append(parts, version)
	}
	if boolPtrValue(cluster.PrivateClusterEnabled) && valueOrEmpty(cluster.PrivateFQDN) != "" {
		parts = append(parts, "api-host="+valueOrEmpty(cluster.PrivateFQDN))
		if boolPtrValue(cluster.PublicFQDNEnabled) && valueOrEmpty(cluster.FQDN) != "" {
			parts = append(parts, "public-fqdn="+valueOrEmpty(cluster.FQDN))
		}
	} else if valueOrEmpty(cluster.FQDN) != "" {
		parts = append(parts, "api-host="+valueOrEmpty(cluster.FQDN))
	}
	if cluster.NetworkPlugin != nil {
		parts = append(parts, "plugin="+*cluster.NetworkPlugin)
	}
	if cluster.NetworkPolicy != nil {
		parts = append(parts, "policy="+*cluster.NetworkPolicy)
	}
	if cluster.OutboundType != nil {
		parts = append(parts, "outbound="+*cluster.OutboundType)
	}
	if len(cluster.AddonNames) == 1 {
		parts = append(parts, "addons="+cluster.AddonNames[0])
	} else if len(cluster.AddonNames) > 1 {
		parts = append(parts, fmt.Sprintf("addons=%s (+%d more)", cluster.AddonNames[0], len(cluster.AddonNames)-1))
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
	if boolPtrValue(cluster.OIDCIssuerEnabled) && valueOrEmpty(cluster.OIDCIssuerURL) != "" {
		parts = append(parts, "oidc-url="+valueOrEmpty(cluster.OIDCIssuerURL))
	}
	return strings.Join(parts, "; ")
}

func aksNote(cluster models.AksClusterAsset) string {
	if cluster.PrivateClusterEnabled != nil && !*cluster.PrivateClusterEnabled && valueOrEmpty(cluster.FQDN) != "" {
		parts := []string{"public API stays visible"}
		if valueOrEmpty(cluster.ClusterIdentityType) == "ServicePrincipal" {
			parts = append(parts, "service-principal-backed credentials are in use")
		}
		if cluster.AzureRBACEnabled != nil && !*cluster.AzureRBACEnabled {
			parts = append(parts, "Azure RBAC is disabled")
		}
		if cluster.LocalAccountsDisabled != nil && !*cluster.LocalAccountsDisabled {
			parts = append(parts, "local accounts remain enabled")
		}
		return "This cluster is the broadest visible AKS surface in the current set: " + strings.Join(parts, ", ") + "."
	}

	parts := []string{}
	if boolPtrValue(cluster.PrivateClusterEnabled) {
		parts = append(parts, "private API is visible")
	}
	switch valueOrEmpty(cluster.ClusterIdentityType) {
	case "SystemAssigned":
		parts = append(parts, "system-assigned identity is visible")
	case "UserAssigned":
		parts = append(parts, "user-assigned identity is visible")
	case "SystemAssigned, UserAssigned":
		parts = append(parts, "system- and user-assigned identity is visible")
	}
	if boolPtrValue(cluster.AzureRBACEnabled) {
		parts = append(parts, "Azure RBAC is enabled")
	}
	if boolPtrValue(cluster.WorkloadIdentityEnabled) {
		parts = append(parts, "workload identity is enabled")
	}
	if len(parts) == 0 {
		return "This cluster has narrower visible edge exposure in the current set."
	}
	return "This cluster is less exposed at the edge, but it still shows " + strings.Join(parts, ", ") + "."
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
		parts = append(parts, "sku="+*registry.SKUName)
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

func acrOperatorSignal(registry models.AcrRegistryAsset) string {
	parts := []string{}
	if posture := acrPostureContext(registry); posture != "-" {
		parts = append(parts, posture)
	}
	if depth := acrDepthContext(registry); depth != "-" {
		parts = append(parts, depth)
	}
	return strings.Join(parts, "; ")
}

func acrNote(registry models.AcrRegistryAsset) string {
	if strings.EqualFold(valueOrEmpty(registry.PublicNetworkAccess), "Enabled") ||
		boolPtrValue(registry.AdminUserEnabled) ||
		boolPtrValue(registry.AnonymousPullEnabled) {
		parts := []string{}
		if strings.EqualFold(valueOrEmpty(registry.PublicNetworkAccess), "Enabled") {
			parts = append(parts, "public network access")
		}
		if boolPtrValue(registry.AdminUserEnabled) {
			parts = append(parts, "admin-user authentication")
		}
		if boolPtrValue(registry.AnonymousPullEnabled) {
			parts = append(parts, "anonymous pull")
		}
		return "This registry is the broadest visible ACR surface in the current set: " + strings.Join(parts, ", ") + " are all enabled."
	}

	parts := []string{}
	if strings.EqualFold(valueOrEmpty(registry.PublicNetworkAccess), "Disabled") {
		parts = append(parts, "public network access is disabled")
	}
	if intPtrValue(registry.PrivateEndpointConnectionCount) > 0 {
		parts = append(parts, "private endpoint coverage is visible")
	}
	if valueOrEmpty(registry.WorkloadIdentityType) != "" {
		parts = append(parts, "managed identity is visible")
	}
	if intPtrValue(registry.ReplicationCount) > 0 {
		parts = append(parts, "replication is visible")
	}
	if strings.EqualFold(valueOrEmpty(registry.TrustPolicyStatus), "enabled") {
		parts = append(parts, "content trust is enabled")
	}
	if len(parts) == 0 {
		return "This registry has narrower visible exposure in the current set."
	}
	return "This registry is less exposed at the edge, but it still shows " + strings.Join(parts, ", ") + "."
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
		parts = append(parts, "sku="+*service.SKUName)
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

func apiMgmtGatewayLabel(service models.ApiMgmtServiceAsset) string {
	lines := []string{}
	for _, hostname := range service.GatewayHostnames {
		lines = append(lines, "gateway: "+hostname)
	}
	for _, hostname := range service.ManagementHostnames {
		lines = append(lines, "management: "+hostname)
	}
	for _, hostname := range service.PortalHostnames {
		lines = append(lines, "portal: "+hostname)
	}
	if len(lines) == 0 {
		return "-"
	}
	return strings.Join(lines, "\n")
}

func apiMgmtOperatorSignal(service models.ApiMgmtServiceAsset) string {
	parts := []string{}
	if inventory := apiMgmtInventoryContext(service); inventory != "-" {
		parts = append(parts, inventory)
	}
	if posture := apiMgmtPostureContext(service); posture != "-" {
		parts = append(parts, posture)
	}
	return strings.Join(parts, "; ")
}

func apiMgmtNote(service models.ApiMgmtServiceAsset) string {
	if strings.EqualFold(valueOrEmpty(service.PublicNetworkAccess), "Enabled") && len(service.GatewayHostnames) > 0 {
		parts := []string{"public gateway hostnames are visible"}
		if len(service.PortalHostnames) > 0 {
			parts = append(parts, "portal hostnames are visible")
		}
		if valueOrEmpty(service.WorkloadIdentityType) != "" {
			parts = append(parts, "managed identity is visible")
		}
		if intPtrValue(service.BackendCount) > 0 {
			parts = append(parts, "backend depth is visible")
		}
		if intPtrValue(service.NamedValueSecretCount) > 0 || intPtrValue(service.NamedValueKeyVaultCount) > 0 {
			parts = append(parts, "named-value secret depth is visible")
		}
		return "This service is the broadest visible API Management surface in the current set: " + strings.Join(parts, ", ") + "."
	}

	parts := []string{}
	if strings.EqualFold(valueOrEmpty(service.PublicNetworkAccess), "Disabled") {
		parts = append(parts, "public network access is disabled")
	}
	if valueOrEmpty(service.VirtualNetworkType) != "" {
		parts = append(parts, "virtual network type "+valueOrEmpty(service.VirtualNetworkType)+" is visible")
	}
	if valueOrEmpty(service.WorkloadIdentityType) != "" {
		parts = append(parts, "managed identity is visible")
	}
	if intPtrValue(service.BackendCount) > 0 {
		parts = append(parts, "backend depth is visible")
	}
	if intPtrValue(service.NamedValueSecretCount) > 0 || intPtrValue(service.NamedValueKeyVaultCount) > 0 {
		parts = append(parts, "named-value secret depth is visible")
	}
	if len(parts) == 0 {
		return "This service has narrower visible API gateway exposure in the current set."
	}
	return "This service is less exposed at the edge, but it still shows " + strings.Join(parts, ", ") + "."
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

func stringOrFallback(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
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

func devopsOperatorSignal(item models.DevopsPipelineAsset) string {
	parts := []string{}
	if access := strings.TrimSpace(devopsAccessContext(item)); access != "" && access != "-" {
		parts = append(parts, access)
	}
	if secret := strings.TrimSpace(devopsSecretContext(item)); secret != "" && secret != "none visible" {
		parts = append(parts, secret)
	}
	if item.PartialRead {
		parts = append(parts, "read-path coverage still partial")
	}
	return strings.Join(parts, "; ")
}

func devopsNote(item models.DevopsPipelineAsset) string {
	statement := ""
	switch {
	case item.PartialRead:
		statement = "Trusted input is visible here, but Azure DevOps read coverage is still partial."
	case len(item.CurrentOperatorInjectionSurfaceTypes) > 0:
		statement = "Current Azure DevOps evidence shows this pipeline trusts an input the operator can poison here."
	case devopsPrimaryAccessState(item) == "read":
		statement = "Trusted repository input is visible here, but current Azure DevOps evidence only proves read access, not poisoning."
	case devopsPrimaryAccessState(item) == "exists-only":
		statement = "This path trusts a source reference, but current Azure DevOps evidence only proves that the source exists."
	case boolPtrValue(item.CurrentOperatorCanQueue):
		statement = "Current credentials can start this pipeline, but Azure DevOps evidence does not yet prove a poisonable trusted input."
	default:
		statement = firstSentence(item.Summary)
	}

	hint := devopsHint(item)
	switch {
	case statement == "":
		return hint
	case hint == "":
		return statement
	default:
		return statement + " " + hint
	}
}

func devopsHint(item models.DevopsPipelineAsset) string {
	targetHint := ""
	for _, clue := range item.TargetClues {
		switch clue {
		case "AKS/Kubernetes":
			targetHint = "Start with aks."
		case "App Service":
			targetHint = "Start with app-services."
		case "Functions":
			targetHint = "Start with functions."
		case "ARM/Bicep/Terraform":
			targetHint = "Start with arm-deployments."
		case "ACR/Containers":
			targetHint = "Start with acr."
		}
		if targetHint != "" {
			break
		}
	}

	controlHint := ""
	if len(item.AzureServiceConnectionNames) > 0 {
		controlHint = "Then validate Azure control behind " + item.AzureServiceConnectionNames[0] + "."
	}

	switch {
	case targetHint != "" && controlHint != "":
		return targetHint + " " + controlHint
	case targetHint != "":
		return targetHint
	default:
		return controlHint
	}
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

func firstSentence(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	for index, r := range text {
		if r != '.' && r != '!' && r != '?' {
			continue
		}
		next := index + len(string(r))
		if next < len(text) && text[next] != ' ' {
			continue
		}
		return strings.TrimSpace(text[:next])
	}
	return text
}

func boolPtrValue(value *bool) bool {
	return value != nil && *value
}

func intPtrValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
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
	parts := []string{}
	if subnet := join(subnets, ", "); subnet != "" {
		parts = append(parts, "subnet="+subnet)
	}
	if vnet := join(vnets, ", "); vnet != "" {
		parts = append(parts, "vnet="+vnet)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func whoAmITable(payload models.WhoAmIOutput) string {
	rows := [][]string{
		{
			stringOrFallback(payload.Subscription.DisplayName, payload.Subscription.ID),
			stringOrFallback(payload.Principal.DisplayName, payload.Principal.ID),
			payload.Principal.PrincipalType,
			authModeLabel(payload.Metadata.AuthMode),
			stringOrFallback(valueOrEmpty(payload.Metadata.TokenSource), "unknown"),
			whoAmIScopeContext(payload.EffectiveScopes),
		},
	}
	return renderListTable(
		fmt.Sprintf("ho-azure %s", payload.Metadata.Command),
		[]string{"subscription", "principal", "type", "auth", "token", "scope"},
		rows,
		[]string{"unknown", "unknown", "unknown", "unknown", "unknown", "unknown"},
		whoAmITakeaway(payload),
	)
}

func whoAmIScopeContext(scopes []models.ScopeRef) string {
	if len(scopes) == 0 {
		return "unknown"
	}
	scope := scopes[0]
	return stringOrFallback(scope.DisplayName, scope.ID)
}

func whoAmITakeaway(payload models.WhoAmIOutput) string {
	return fmt.Sprintf(
		"Operating as %s (%s) in %s via %s.",
		stringOrFallback(payload.Principal.DisplayName, payload.Principal.ID),
		stringOrFallback(payload.Principal.PrincipalType, "unknown"),
		stringOrFallback(payload.Subscription.DisplayName, payload.Subscription.ID),
		authModeLabel(payload.Metadata.AuthMode),
	)
}

func authModeLabel(value *string) string {
	normalized := strings.TrimSpace(valueOrEmpty(value))
	if normalized == "" {
		return "unknown"
	}

	switch normalized {
	case "azure_cli":
		return "Azure CLI"
	case "azure_cli_user":
		return "Azure CLI user"
	case "azure_cli_service_principal":
		return "Azure CLI service principal"
	case "azure_cli_managed_identity":
		return "Azure CLI managed identity"
	case "environment":
		return "Environment credential"
	case "environment_client_secret":
		return "Environment client secret"
	case "environment_client_certificate":
		return "Environment client certificate"
	case "fixture":
		return "Fixture"
	default:
		return strings.ReplaceAll(normalized, "_", " ")
	}
}

func permissionsTable(payload models.PermissionsOutput) string {
	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := cellStyle.Bold(true)
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
			return cellStyle
		})

	return titleStyle.Render("ho-azure permissions") + "\n\n" + strings.TrimRight(table.String(), "\n") + "\n" +
		"\nTakeaway: " + permissionsTakeaway(payload) + "\n"
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
		"ho-azure principals",
		[]string{"principal", "type", "roles", "assignments", "identity context", "sources", "current"},
		rows,
		[]string{"No visible principals were confirmed from current scope.", "", "", "", "", "", ""},
		principalsTakeaway(payload),
	)
}

func privescTable(payload models.PrivescOutput) string {
	if len(payload.Paths) == 0 {
		return renderListTable(
			"ho-azure privesc",
			[]string{"priority", "starting foothold", "path type", "target"},
			nil,
			[]string{"No visible privilege-escalation paths were confirmed from current scope.", "", "", ""},
			privescTakeaway(payload),
		)
	}

	sections := make([]string, 0, len(payload.Paths))
	for _, path := range payload.Paths {
		row := renderStructuredTableWithTitle(
			"",
			[]string{"priority", "starting foothold", "path type", "target"},
			[][]string{{
				path.Priority,
				path.StartingFoothold,
				privescPathLabel(path.PathType),
				privescTarget(path),
			}},
			false,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if note := privescNote(path); note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}

	return joinRenderedBlocks(sections) + "\n\nTakeaway: " + privescTakeaway(payload) + "\n"
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
		"ho-azure lighthouse",
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
		"ho-azure cross-tenant",
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

	output := renderListTable(
		"ho-azure auth-policies",
		[]string{"policy", "state", "scope", "operator signal"},
		rows,
		[]string{"No visible auth policy rows were confirmed from current scope.", "", "", ""},
		authPoliciesTakeaway(payload),
	)
	return appendCustomFindingsSection(output, payload.Findings,
		func(f models.AuthPolicyFinding) string { return f.Severity },
		func(f models.AuthPolicyFinding) string { return f.Title },
		func(f models.AuthPolicyFinding) string { return f.Description },
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

	output := renderListTable(
		"ho-azure resource-trusts",
		[]string{"resource", "type", "trust", "target", "exposure", "why it matters"},
		rows,
		[]string{"No visible resource trust surfaces were confirmed from current scope.", "", "", "", "", ""},
		resourceTrustsTakeaway(payload),
	)
	return appendCustomFindingsSection(output, payload.Findings,
		func(f models.ResourceTrustFinding) string { return f.Severity },
		func(f models.ResourceTrustFinding) string { return f.Title },
		func(f models.ResourceTrustFinding) string { return f.Description },
	)
}

func appCredentialsTable(payload models.AppCredentialsOutput) string {
	if len(payload.AppCredentials) == 0 {
		return renderListTable(
			"ho-azure app-credentials",
			[]string{"target", "row", "auth surface"},
			nil,
			[]string{"No visible high-signal application or service-principal credential rows were confirmed from current scope.", "", ""},
			appCredentialsTakeaway(payload),
		)
	}

	sections := make([]string, 0, len(payload.AppCredentials))
	for index, item := range payload.AppCredentials {
		row := renderStructuredTableWithTitle(
			"ho-azure app-credentials",
			[]string{"target", "row", "auth surface"},
			[][]string{{
				appCredentialTargetLabel(item),
				appCredentialRowClassLabel(item.RowClass),
				appCredentialSurfaceLabel(item.CredentialType),
			}},
			index == 0,
		)
		note := renderWrappedNoteTableWithWidth(appCredentialEvidenceLabel(item), renderedTableCellWidth(row))
		sections = append(sections, joinRenderedSections(row, note))
	}

	return joinRenderedBlocks(sections) + "\n\nTakeaway: " + appCredentialsTakeaway(payload) + "\n"
}

func roleTrustsTable(payload models.RoleTrustsOutput) string {
	if len(payload.Trusts) == 0 {
		return renderListTable(
			"ho-azure role-trusts",
			[]string{"info"},
			nil,
			[]string{"No visible role-trusts were confirmed from current scope."},
			"",
		)
	}

	sections := make([]string, 0, len(payload.Trusts))
	for index, trust := range payload.Trusts {
		sections = append(sections, renderStructuredTableWithTitle(
			"ho-azure role-trusts",
			[]string{"trust", "source", "target", "confidence", "operator signal", "next review"},
			[][]string{{
				trust.TrustType,
				stringOrFallback(valueOrEmpty(trust.SourceName), trust.SourceObjectID),
				stringOrFallback(valueOrEmpty(trust.TargetName), trust.TargetObjectID),
				trust.Confidence,
				valueOrEmpty(trust.OperatorSignal),
				valueOrEmpty(trust.NextReview),
			}},
			index == 0,
		))
		if visibleTransform := roleTrustVisibleTransform(trust); visibleTransform != "" {
			sections = append(sections, renderStructuredTableWithTitle(
				"",
				[]string{"how control could widen"},
				[][]string{{visibleTransform}},
				false,
			))
		}
	}

	return strings.Join(sections, "\n") + "\nTakeaway: " + roleTrustsTakeaway(payload) + "\n"
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

func appCredentialsTakeaway(payload models.AppCredentialsOutput) string {
	if len(payload.AppCredentials) == 0 {
		return "0 app-credential rows visible; no high-signal credential or federated-trust evidence surfaced from the current foothold."
	}

	counts := map[string]int{}
	for _, item := range payload.AppCredentials {
		counts[item.RowClass]++
	}

	parts := []string{}
	if counts["directly_addable_federated_trust"] > 0 {
		parts = append(parts, fmt.Sprintf("%d direct federated-trust control path(s)", counts["directly_addable_federated_trust"]))
	}
	if counts["directly_addable"] > 0 {
		parts = append(parts, fmt.Sprintf("%d direct credential-control path(s)", counts["directly_addable"]))
	}
	if counts["federated_trust_present"] > 0 {
		parts = append(parts, fmt.Sprintf("%d existing federated trust row(s)", counts["federated_trust_present"]))
	}
	if counts["existing_credential"] > 0 {
		parts = append(parts, fmt.Sprintf("%d existing credential row(s)", counts["existing_credential"]))
	}
	if counts["control_context_only"] > 0 {
		parts = append(parts, fmt.Sprintf("%d control-context-only row(s)", counts["control_context_only"]))
	}
	if len(parts) == 0 {
		parts = append(parts, "no classified app-credential evidence")
	}

	return fmt.Sprintf("%d app-credential row(s) visible; %s.", len(payload.AppCredentials), strings.Join(parts, ", "))
}

func inventoryTakeaway(payload models.InventoryOutput) string {
	return fmt.Sprintf("%d resources across %d resource groups.", payload.ResourceCount, payload.ResourceGroupCount)
}

func rbacTakeaway(payload models.RbacOutput) string {
	return fmt.Sprintf("%d RBAC assignments across %d principals.", len(payload.RoleAssignments), len(payload.Principals))
}

func endpointsTakeaway(payload models.EndpointsOutput) string {
	families := map[string]int{}
	for _, endpoint := range payload.Endpoints {
		key := endpoint.ExposureFamily
		if key == "" {
			key = "unknown"
		}
		families[key]++
	}

	ordered := []string{"public-ip", "managed-web-hostname"}
	seen := map[string]bool{}
	parts := []string{}
	for _, key := range ordered {
		if families[key] > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", families[key], key))
			seen[key] = true
		}
	}
	for _, key := range sortedFamilyKeys(families) {
		if seen[key] {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d %s", families[key], key))
	}

	counts := "no reachable surfaces visible"
	if len(parts) > 0 {
		counts = strings.Join(parts, ", ")
	}
	return fmt.Sprintf("%d reachable surfaces visible; %s.", len(payload.Endpoints), counts)
}

func nicsTakeaway(payload models.NicsOutput) string {
	attached := 0
	publicRefs := 0
	for _, nic := range payload.NicAssets {
		if nic.AttachedAssetID != nil && strings.TrimSpace(*nic.AttachedAssetID) != "" {
			attached++
		}
		publicRefs += len(nic.PublicIPIDs)
	}
	return fmt.Sprintf("%d NICs visible; %d attached to visible assets and %d reference public IP resources.", len(payload.NicAssets), attached, publicRefs)
}

func permissionsTakeaway(payload models.PermissionsOutput) string {
	privileged := 0
	workloadPivots := 0
	trustFollowOns := 0
	for _, permission := range payload.Permissions {
		if permission.Privileged {
			privileged++
		}
		signal := strings.ToLower(permission.OperatorSignal)
		if strings.Contains(signal, "workload pivot visible") {
			workloadPivots++
		}
		if strings.Contains(signal, "trust expansion follow-on") {
			trustFollowOns++
		}
	}
	return fmt.Sprintf("%d of %d principals hold high-impact RBAC roles; %d workload-pivot follow-ons and %d trust-expansion follow-ons.", privileged, len(payload.Permissions), workloadPivots, trustFollowOns)
}

func appendFindingsSection(output string, findings []models.Finding) string {
	return appendCustomFindingsSection(output, findings,
		func(f models.Finding) string { return f.Severity },
		func(f models.Finding) string { return f.Title },
		func(f models.Finding) string { return f.Description },
	)
}

func appendCustomFindingsSection[T any](output string, findings []T, severity func(T) string, title func(T) string, description func(T) string) string {
	if len(findings) == 0 {
		return output
	}

	findingsBlock := "\nFindings:\n"
	for _, finding := range findings {
		findingsBlock += fmt.Sprintf("- %s: %s\n", strings.ToUpper(severity(finding)), title(finding))
		findingsBlock += fmt.Sprintf("  %s\n", description(finding))
	}

	if marker := strings.Index(output, "\nTakeaway: "); marker >= 0 {
		return output[:marker] + findingsBlock + output[marker:]
	}
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	return output + findingsBlock
}

func appendPayloadFindingsSection(output string, payload any) string {
	if strings.Contains(output, "\nFindings:\n") {
		return output
	}

	findings := payloadFindings(payload)
	if len(findings) == 0 {
		return output
	}

	findingsBlock := "\nFindings:\n"
	limit := minInt(len(findings), 5)
	for _, finding := range findings[:limit] {
		findingsBlock += fmt.Sprintf("- %s: %s\n", strings.ToUpper(finding.Severity), finding.Title)
		findingsBlock += fmt.Sprintf("  %s\n", finding.Description)
	}
	if remaining := len(findings) - limit; remaining > 0 {
		findingsBlock += fmt.Sprintf("- ... plus %d more findings in JSON artifacts.\n", remaining)
	}
	return insertBeforeTakeaway(output, findingsBlock)
}

func appendPayloadIssuesSection(output string, payload any) string {
	if strings.Contains(output, "\nCurrent-scope issues:\n") {
		return output
	}

	issues := payloadIssues(payload)
	if len(issues) == 0 {
		return output
	}

	issuesBlock := "\nCurrent-scope issues:\n"
	limit := minInt(len(issues), 5)
	for _, issue := range issues[:limit] {
		issuesBlock += fmt.Sprintf("- %s: %s\n", issue.Kind, issue.Message)
	}
	if remaining := len(issues) - limit; remaining > 0 {
		issuesBlock += fmt.Sprintf("- ... plus %d more current-scope issues in JSON artifacts.\n", remaining)
	}
	return insertBeforeTakeaway(output, issuesBlock)
}

func insertBeforeTakeaway(output string, block string) string {
	if marker := strings.Index(output, "\nTakeaway: "); marker >= 0 {
		return output[:marker] + block + output[marker:]
	}
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	return output + block
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
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
		return "current foothold (" + privescOperatorPrincipalType(path.PrincipalType) + ")"
	}
	principal := strings.TrimSpace(path.Principal)
	asset := strings.TrimSpace(valueOrEmpty(path.Asset))
	principalType := privescOperatorPrincipalType(path.PrincipalType)
	if principal != "" && asset != "" {
		return principalType + " " + principal + " via " + asset
	}
	if principal != "" {
		return principalType + " " + principal
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

	countSummary := fmt.Sprintf("%d privilege-escalation paths surfaced; %d current-identity-rooted, %d %s, %s.", len(payload.Paths), rooted, visibleOnly, visibleLabel, counts)
	return countSummary
}

func privescPreferredTakeaway(paths []models.PrivescPathSummary) string {
	for _, path := range paths {
		if path.Preferred && strings.TrimSpace(path.PreferredReason) != "" {
			return path.PreferredReason
		}
	}
	return ""
}

func privescNote(path models.PrivescPathSummary) string {
	parts := []string{}
	if path.Preferred && strings.TrimSpace(path.PreferredReason) != "" {
		parts = append(parts, strings.TrimSpace(path.PreferredReason))
	}
	if boundary := strings.TrimSpace(privescProofBoundary(path)); boundary != "" {
		parts = append(parts, boundary)
	}
	if review := privescCompactNextReview(path); review != "" {
		parts = append(parts, review)
	}
	return strings.Join(parts, " ")
}

func privescCompactNextReview(path models.PrivescPathSummary) string {
	switch {
	case path.CurrentIdentity:
		return "Next review: rbac assignment evidence and scope."
	case strings.TrimSpace(path.PathType) == "ingress-backed-workload-identity":
		return "Next review: managed-identities workload-to-identity anchor."
	case strings.TrimSpace(path.NextReview) != "":
		return "Next review: role-trusts influence paths."
	default:
		return ""
	}
}

func privescDisplayPrincipalType(principalType string) string {
	switch strings.TrimSpace(principalType) {
	case "ManagedIdentity":
		return "ManagedIdentity"
	case "ServicePrincipal":
		return "ServicePrincipal"
	case "User":
		return "User"
	default:
		normalized := strings.TrimSpace(principalType)
		if normalized == "" {
			return "unknown"
		}
		return normalized
	}
}

func privescOperatorPrincipalType(principalType string) string {
	switch strings.TrimSpace(principalType) {
	case "ManagedIdentity":
		return "managed identity"
	case "ServicePrincipal":
		return "service principal"
	case "User":
		return "user"
	default:
		normalized := strings.TrimSpace(principalType)
		if normalized == "" {
			return "unknown"
		}
		return strings.ToLower(normalized)
	}
}

func authPoliciesTakeaway(payload models.AuthPoliciesOutput) string {
	return fmt.Sprintf("%d policy rows, %d findings, and %d current-scope issues.", len(payload.AuthPolicies), len(payload.Findings), len(payload.Issues))
}

func roleTrustVisibleTransform(trust models.RoleTrustSummary) string {
	if value := strings.TrimSpace(valueOrEmpty(trust.UsableIdentityResult)); value != "" {
		return value
	}
	return strings.TrimSpace(valueOrEmpty(trust.EscalationMechanism))
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
	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := cellStyle.Bold(true)
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
			return cellStyle
		})

	output := titleStyle.Render("ho-azure managed-identities") + "\n\n" + strings.TrimRight(table.String(), "\n") + "\n"
	output = appendCustomFindingsSection(output, payload.Findings,
		func(f models.ManagedIdentityFinding) string { return f.Severity },
		func(f models.ManagedIdentityFinding) string { return f.Title },
		func(f models.ManagedIdentityFinding) string { return f.Description },
	)
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
	if len(payload.EnvVars) == 0 {
		output := renderStructuredTableWithTitle(
			"",
			[]string{"workload", "kind", "setting", "value type", "identity", "reference"},
			[][]string{{"no environment variable rows visible", "", "", "", "", ""}},
			false,
		)
		output = appendCustomFindingsSection(output, payload.Findings,
			func(f models.EnvVarFinding) string { return f.Severity },
			func(f models.EnvVarFinding) string { return f.Title },
			func(f models.EnvVarFinding) string { return f.Description },
		)
		return output + "\nTakeaway: " + envVarsTakeaway(payload) + "\n"
	}

	sections := make([]string, 0, len(payload.EnvVars))
	for _, envVar := range payload.EnvVars {
		row := renderStructuredTableWithTitle(
			"",
			[]string{"workload", "kind", "setting", "value type", "identity", "reference"},
			[][]string{{
				envVar.AssetName,
				envVar.AssetKind,
				envVar.SettingName,
				envVar.ValueType,
				envVarWorkloadIdentityLabel(envVar),
				envVarReferenceLabel(envVar),
			}},
			false,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if note := envVarNote(envVar); note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}

	output := joinRenderedBlocks(sections) + "\n"
	output = appendCustomFindingsSection(output, payload.Findings,
		func(f models.EnvVarFinding) string { return f.Severity },
		func(f models.EnvVarFinding) string { return f.Title },
		func(f models.EnvVarFinding) string { return f.Description },
	)
	return output + "\nTakeaway: " + envVarsTakeaway(payload) + "\n"
}

func tokensCredentialsTable(payload models.TokensCredentialsOutput) string {
	if len(payload.Surfaces) == 0 {
		output := tokenCredentialMainRowTable(
			[]string{"priority", "asset", "kind", "surface", "access path", "operator signal"},
			[][]string{{"", "no token or credential surfaces visible", "", "", "", ""}},
		)
		output = appendCustomFindingsSection(output, payload.Findings,
			func(f models.TokenCredentialFinding) string { return f.Severity },
			func(f models.TokenCredentialFinding) string { return f.Title },
			func(f models.TokenCredentialFinding) string { return f.Description },
		)
		return output + "\nTakeaway: " + tokensCredentialsTakeaway(payload) + "\n"
	}

	sections := make([]string, 0, len(payload.Surfaces))
	for _, surface := range payload.Surfaces {
		row := tokenCredentialMainRowTable(
			[]string{"priority", "asset", "kind", "surface", "access path", "operator signal"},
			[][]string{{
				surface.Priority,
				surface.AssetName,
				surface.AssetKind,
				string(surface.SurfaceType),
				surface.AccessPath,
				surface.OperatorSignal,
			}},
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if note := tokenCredentialNote(surface); note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}

	output := joinRenderedBlocks(sections) + "\n"
	return output + "\nTakeaway: " + tokensCredentialsTakeaway(payload) + "\n"
}

func tokenCredentialNote(surface models.TokenCredentialSurfaceSummary) string {
	return strings.TrimSpace(surface.Summary)
}

func tokenCredentialMainRowTable(headers []string, rows [][]string) string {
	return strings.TrimRight(
		liptable.New().
			Headers(headers...).
			Rows(rows...).
			StyleFunc(func(row int, col int) lipgloss.Style {
				style := lipgloss.NewStyle().Padding(0, 1)
				if row == liptable.HeaderRow {
					return style.Bold(true)
				}
				return style
			}).
			String(),
		"\n",
	) + "\n"
}

func chainsOverviewTable(payload models.ChainsOverviewOutput) string {
	rows := make([][]string, 0, len(payload.Families))
	for _, family := range payload.Families {
		rows = append(rows, []string{
			family.Family,
			family.State,
			family.Summary,
			strings.Join(family.BestCurrentExamples, ", "),
			chainsBackingCommands(family.SourceCommands),
		})
	}

	output := renderListTable(
		"ho-azure chains",
		[]string{"family", "state", "summary", "best examples", "backing commands"},
		rows,
		[]string{"no chain families are currently registered", "", "", "", ""},
		"",
	)

	boundaryRows := make([][]string, 0, len(payload.Families))
	for _, family := range payload.Families {
		boundaryRows = append(boundaryRows, []string{
			family.Family,
			family.AllowedClaim,
			family.CurrentGap,
		})
	}

	return output + "\n" + renderStructuredTable(
		"chains claim boundaries",
		[]string{"family", "allowed claim", "current gap"},
		boundaryRows,
	)
}

func chainsFamilyTable(payload models.ChainsOutput) string {
	if renderer, ok := chainsFamilyTableRenderers[payload.Family]; ok {
		return renderer(payload)
	}

	rows := make([][]string, 0, len(payload.Paths))
	for _, path := range payload.Paths {
		rows = append(rows, []string{
			path.Priority,
			path.AssetName,
			valueOrEmpty(path.SettingName),
			path.TargetService,
			path.TargetResolution,
			strings.Join(path.TargetNames, ","),
			path.NextReview,
			credentialPathNote(path),
		})
	}

	return renderListTable(
		"ho-azure chains",
		[]string{"priority", "asset", "setting", "target", "target resolution", "visible targets", "next review", "note"},
		rows,
		[]string{"no visible credential paths were confirmed from current scope", "", "", "", "", "", "", ""},
		chainsCredentialPathTakeaway(payload.Paths),
	)
}

func chainsCredentialPathTable(payload models.ChainsOutput) string {
	sections := []string{}

	if len(payload.Paths) == 0 {
		sections = append(sections, renderListTable(
			"ho-azure chains",
			[]string{"priority", "urgency", "asset", "setting", "target", "target resolution", "visible targets", "next review", "confidence boundary"},
			nil,
			[]string{"no visible credential paths were confirmed from current scope", "", "", "", "", "", "", "", ""},
			"",
		))
		return strings.Join(sections, "\n\n") + "\n"
	}

	for _, path := range payload.Paths {
		sections = append(sections, renderStructuredTableWithTitle(
			"ho-azure chains",
			[]string{"priority", "urgency", "asset", "setting", "target", "target resolution", "visible targets", "next review", "confidence boundary"},
			[][]string{{
				path.Priority,
				valueOrEmpty(path.Urgency),
				path.AssetName,
				valueOrEmpty(path.SettingName),
				path.TargetService,
				path.TargetResolution,
				chainsTargetContext(path),
				path.NextReview,
				credentialPathConfidenceBoundary(path),
			}},
			len(sections) == 0,
		))
	}

	return strings.Join(sections, "\n\n") + "\n"
}

func chainsComputeControlTable(payload models.ChainsOutput) string {
	sections := []string{}

	if len(payload.Paths) == 0 {
		sections = append(sections, renderListTable(
			"ho-azure chains",
			[]string{"priority", "when", "reach from here", "compute foothold", "token path", "identity", "Azure access", "proof status"},
			nil,
			[]string{"no visible compute-control paths were confirmed from current scope", "", "", "", "", "", "", ""},
			"",
		))
		return strings.Join(sections, "\n\n") + "\n"
	}

	for _, path := range payload.Paths {
		row := renderStructuredTableWithTitle(
			"ho-azure chains",
			[]string{"priority", "when", "reach from here", "compute foothold", "token path", "identity", "Azure access", "proof status"},
			[][]string{{
				path.Priority,
				computeControlWhenLabel(valueOrEmpty(path.Urgency)),
				computeControlReachFromHereLabel(valueOrEmpty(path.InsertionPoint)),
				path.AssetName,
				computeControlTokenPathLabel(valueOrEmpty(path.InsertionPoint)),
				computeControlIdentityLabel(path.TargetNames),
				firstNonEmptyText(valueOrEmpty(path.StrongerOutcome), valueOrEmpty(path.LikelyImpact), "-"),
				computeControlProofStatusLabel(path.TargetResolution),
			}},
			len(sections) == 0,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if note := firstNonEmptyText(valueOrEmpty(path.Note), valueOrEmpty(path.WhyCare)); note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		if missingConfirmation := strings.TrimSpace(path.MissingConfirmation); missingConfirmation != "" {
			parts = append(parts, renderWrappedDetailTableWithWidth("what is still missing", missingConfirmation, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}
	return joinRenderedBlocks(sections) + "\n"
}

func chainsDeploymentPathTable(payload models.ChainsOutput) string {
	sections := []string{}

	if len(payload.Paths) == 0 {
		sections = append(sections, renderListTable(
			"ho-azure chains",
			[]string{"priority", "urgency", "source", "actionability", "insertion point", "likely azure impact", "what's missing", "next review"},
			nil,
			[]string{"no visible deployment paths were confirmed from current scope", "", "", "", "", "", "", ""},
			"",
		))
		if payload.ClaimBoundary != "" {
			sections = append(sections, "Claim boundary: "+payload.ClaimBoundary)
		}
		return strings.Join(sections, "\n\n") + "\n"
	}

	for _, path := range payload.Paths {
		row := renderStructuredTableWithTitle(
			"ho-azure chains",
			[]string{"priority", "urgency", "source", "actionability", "insertion point", "likely azure impact", "what's missing", "next review"},
			[][]string{{
				path.Priority,
				valueOrEmpty(path.Urgency),
				valueOrEmpty(path.Source),
				firstNonEmpty(path.Actionability, path.ActionabilityState),
				valueOrEmpty(path.InsertionPointLabel),
				firstNonEmpty(path.LikelyAzureImpact, path.LikelyImpact),
				firstNonEmpty(path.WhatsMissing, path.ConfidenceBoundary),
				path.NextReview,
			}},
			len(sections) == 0,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if note := firstNonEmpty(path.Note, path.WhyCare); note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}
	if payload.ClaimBoundary != "" {
		sections = append(sections, "Claim boundary: "+payload.ClaimBoundary)
	}
	return joinRenderedBlocks(sections) + "\n"
}

func chainsEscalationPathTable(payload models.ChainsOutput) string {
	sections := []string{}

	if len(payload.Paths) == 0 {
		sections = append(sections, renderListTable(
			"ho-azure chains",
			[]string{"priority", "urgency", "starting foothold", "path type", "stronger outcome"},
			nil,
			[]string{"no visible escalation paths were confirmed from current scope", "", "", "", ""},
			"",
		))
		if payload.ClaimBoundary != "" {
			sections = append(sections, "Claim boundary: "+payload.ClaimBoundary)
		}
		return strings.Join(sections, "\n\n") + "\n"
	}

	for _, path := range payload.Paths {
		row := renderStructuredTableWithTitle(
			"ho-azure chains",
			[]string{"priority", "urgency", "starting foothold", "path type", "stronger outcome"},
			[][]string{{
				path.Priority,
				valueOrEmpty(path.Urgency),
				firstNonEmptyText(valueOrEmpty(path.StartingFoothold), path.AssetName),
				firstNonEmptyText(valueOrEmpty(path.PathType), chainsEscalationPathTypeLabel(valueOrEmpty(path.PathConcept))),
				firstNonEmptyText(valueOrEmpty(path.StrongerOutcome), valueOrEmpty(path.LikelyImpact), "Potential stronger Azure control; exact privilege not yet confirmed"),
			}},
			len(sections) == 0,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if note := firstNonEmptyText(valueOrEmpty(path.Note), valueOrEmpty(path.WhyCare)); note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}
	if payload.ClaimBoundary != "" {
		sections = append(sections, "Claim boundary: "+payload.ClaimBoundary)
	}
	return joinRenderedBlocks(sections) + "\n"
}

func chainsTargetContext(path models.ChainPathRecord) string {
	if path.TargetVisibility != nil && strings.TrimSpace(*path.TargetVisibility) != "" {
		return *path.TargetVisibility
	}
	if len(path.TargetNames) == 0 {
		if path.TargetCount > 0 {
			return fmt.Sprintf("%d visible target(s)", path.TargetCount)
		}
		return "none visible"
	}
	if len(path.TargetNames) == 1 {
		return path.TargetNames[0]
	}
	return strings.Join(path.TargetNames, "\n")
}

func credentialPathConfidenceBoundary(path models.ChainPathRecord) string {
	if path.ConfidenceBoundary != nil && strings.TrimSpace(*path.ConfidenceBoundary) != "" {
		return *path.ConfidenceBoundary
	}
	if note := credentialPathNote(path); strings.TrimSpace(note) != "" {
		return note
	}
	return path.Summary
}

func chainsEscalationPathTypeLabel(pathConcept string) string {
	switch pathConcept {
	case "current-foothold-direct-control":
		return "current foothold direct control"
	case "app-permission-reach":
		return "app-permission reach"
	case "trust-expansion":
		return "trust expansion"
	default:
		return firstNonEmptyText(pathConcept, "escalation path")
	}
}

func computeControlWhenLabel(urgency string) string {
	switch urgency {
	case "pivot-now":
		return "act now"
	case "review-soon":
		return "review soon"
	case "bookmark":
		return "keep in view"
	default:
		return firstNonEmptyText(urgency, "-")
	}
}

func computeControlTokenPathLabel(insertionPoint string) string {
	switch insertionPoint {
	case "reachable service token request path":
		return "service token request"
	case "public IMDS token path":
		return "public VM metadata token"
	case "IMDS token path":
		return "VM metadata token"
	default:
		return firstNonEmptyText(insertionPoint, "-")
	}
}

func computeControlReachFromHereLabel(insertionPoint string) string {
	switch insertionPoint {
	case "reachable service token request path", "public IMDS token path":
		return "public exposure visible; exploitation not proved"
	default:
		return "current access does not show the start"
	}
}

func computeControlIdentityLabel(targetNames []string) string {
	names := make([]string, 0, len(targetNames))
	for _, name := range targetNames {
		if strings.TrimSpace(name) != "" {
			names = append(names, name)
		}
	}
	switch len(names) {
	case 0:
		return "not visible"
	case 1:
		return names[0]
	default:
		return "multiple possible: " + strings.Join(names, ", ")
	}
}

func computeControlProofStatusLabel(targetResolution string) string {
	switch targetResolution {
	case "path-confirmed":
		return "confirmed"
	case "identity-choice-corroborated":
		return "best current match"
	case "narrowed candidates":
		return "multiple identities possible"
	case "visibility blocked":
		return "limited visibility"
	case "tenant-wide candidates":
		return "broad match only"
	case "service hint only":
		return "early signal only"
	case "named target not visible":
		return "named identity not visible"
	default:
		return "bounded"
	}
}

func firstNonEmptyText(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func appCredentialTargetLabel(item models.AppCredentialSummary) string {
	target := firstNonEmptyText(item.TargetObjectName, item.TargetObjectID, "unknown target")
	if item.TargetObjectType == "" {
		return target
	}
	return fmt.Sprintf("%s (%s)", target, item.TargetObjectType)
}

func appCredentialRowClassLabel(rowClass string) string {
	switch rowClass {
	case "directly_addable_federated_trust":
		return "direct federated trust control"
	case "directly_addable":
		return "direct credential control"
	case "federated_trust_present":
		return "federated trust present"
	case "existing_credential":
		return "credential material present"
	case "control_context_only":
		return "control context only"
	default:
		return firstNonEmptyText(rowClass, "unknown")
	}
}

func appCredentialSurfaceLabel(kind *string) string {
	switch valueOrEmpty(kind) {
	case "password":
		return "password credential"
	case "key":
		return "key credential"
	case "federated":
		return "federated trust"
	case "password-or-key":
		return "password or key"
	default:
		return "-"
	}
}

func appCredentialEvidenceLabel(item models.AppCredentialSummary) string {
	evidence := firstNonEmptyText(item.CurrentEvidence, item.Summary, "-")
	roleContext := strings.TrimSpace(item.RoleContext)
	if roleContext == "" {
		return evidence
	}
	switch {
	case strings.Contains(strings.ToLower(roleContext), "high-impact azure roles"):
		return evidence + " " + appCredentialRoleContextLabel(roleContext)
	case strings.Contains(strings.ToLower(roleContext), "visible azure roles"):
		return evidence + " " + appCredentialRoleContextLabel(roleContext)
	default:
		return evidence
	}
}

func appCredentialRoleContextLabel(roleContext string) string {
	lower := strings.ToLower(strings.TrimSpace(roleContext))
	switch {
	case strings.Contains(lower, "high-impact azure roles"):
		return strings.Replace(roleContext, "Backed identity", "Backing identity", 1)
	case strings.Contains(lower, "visible azure roles"):
		return strings.Replace(roleContext, "Backed identity", "Backing identity", 1)
	default:
		return roleContext
	}
}

func wrapTableNote(text string, width int) string {
	text = strings.TrimSpace(text)
	if text == "" || width <= 0 {
		return text
	}

	paragraphs := strings.Split(text, "\n")
	lines := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		for _, sentence := range splitTableNoteSentences(paragraph) {
			lines = append(lines, wrapTableNoteLine(sentence, width)...)
		}
	}
	return strings.Join(lines, "\n")
}

func renderWrappedNoteTable(note string) string {
	return renderWrappedNoteTableWithWidth(note, findingNoteWrapWidth)
}

func renderWrappedNoteTableWithWidth(note string, width int) string {
	return renderWrappedDetailTableWithWidth("note", note, width)
}

func renderWrappedDetailTableWithWidth(title string, text string, width int) string {
	if width <= 0 {
		width = findingNoteWrapWidth
	}
	return renderStructuredTableWithTitle(
		"",
		[]string{title},
		[][]string{{wrapTableNote(text, width)}},
		false,
	)
}

func renderedTableCellWidth(rendered string) int {
	maxWidth := 0
	for _, line := range strings.Split(strings.TrimRight(rendered, "\n"), "\n") {
		if width := lipgloss.Width(line); width > maxWidth {
			maxWidth = width
		}
	}
	if maxWidth <= 2 {
		return findingNoteWrapWidth
	}
	return maxWidth - 2
}

func joinRenderedSections(sections ...string) string {
	normalized := make([]string, 0, len(sections))
	for _, section := range sections {
		section = strings.TrimRight(section, "\n")
		if section == "" {
			continue
		}
		normalized = append(normalized, section)
	}
	return strings.Join(normalized, "\n")
}

func joinRenderedBlocks(blocks []string) string {
	normalized := make([]string, 0, len(blocks))
	for _, block := range blocks {
		block = strings.TrimRight(block, "\n")
		if block == "" {
			continue
		}
		normalized = append(normalized, block)
	}
	return strings.Join(normalized, "\n\n")
}

func splitTableNoteSentences(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	sentences := []string{}
	start := 0
	for index, r := range text {
		if r != '.' && r != '!' && r != '?' {
			continue
		}
		next := index + len(string(r))
		if next < len(text) && text[next] != ' ' {
			continue
		}
		sentence := strings.TrimSpace(text[start:next])
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
		start = next
	}

	remainder := strings.TrimSpace(text[start:])
	if remainder != "" {
		sentences = append(sentences, remainder)
	}
	if len(sentences) == 0 {
		return []string{text}
	}
	return sentences
}

func wrapTableNoteLine(text string, width int) []string {
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return nil
	}

	lines := []string{}
	current := words[0]
	for _, word := range words[1:] {
		if len(current)+1+len(word) <= width {
			current += " " + word
			continue
		}
		lines = append(lines, current)
		current = word
	}
	lines = append(lines, current)
	return lines
}

func chainsBackingCommands(sources []models.ChainSourceDescriptor) string {
	commands := make([]string, 0, len(sources))
	for _, source := range sources {
		if source.Command != "" {
			commands = append(commands, source.Command)
		}
	}
	return strings.Join(commands, ", ")
}

func credentialPathNote(path models.ChainPathRecord) string {
	switch path.TargetResolution {
	case "named match":
		return "Named target matched visible inventory."
	case "visibility blocked":
		return path.TargetService + " visibility is blocked; do not infer a target."
	case "narrowed candidates":
		return "Secret-shaped clue suggests a " + path.TargetService + " path; exact target unconfirmed."
	case "tenant-wide candidates":
		return "This app likely reaches " + path.TargetService + ", but the target set is still broad."
	case "service hint only":
		return "HO-Azure sees a likely " + path.TargetService + " path, but no target inventory yet."
	case "named target not visible":
		return "This app names a " + path.TargetService + " target HO-Azure cannot see in current inventory."
	default:
		return path.Summary
	}
}

func chainsCredentialPathTakeaway(paths []models.ChainPathRecord) string {
	if len(paths) == 0 {
		return "No visible credential paths were confirmed from current scope."
	}

	priorityCounts := map[string]int{}
	targetCounts := map[string]int{}
	for _, path := range paths {
		priorityCounts[path.Priority]++
		targetCounts[path.TargetService]++
	}

	parts := []string{
		fmt.Sprintf("%d visible credential paths", len(paths)),
	}
	for _, priority := range []string{"high", "medium", "low"} {
		if count := priorityCounts[priority]; count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", count, priority))
		}
	}
	targets := make([]string, 0, len(targetCounts))
	for target := range targetCounts {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	for _, target := range targets {
		parts = append(parts, fmt.Sprintf("%d %s", targetCounts[target], target))
	}

	return strings.Join(parts, ", ") + "."
}

func firstNonEmpty(values ...*string) string {
	for _, value := range values {
		if value != nil && strings.TrimSpace(*value) != "" {
			return *value
		}
	}
	return ""
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

func envVarWorkloadIdentityLabel(envVar models.EnvVarSummary) string {
	identityType := valueOrEmpty(envVar.WorkloadIdentityType)
	switch {
	case identityType == "" && len(envVar.WorkloadIdentityIDs) == 0:
		return "-"
	case len(envVar.WorkloadIdentityIDs) == 0:
		return identityType
	case identityType == "":
		return fmt.Sprintf("%d user-assigned", len(envVar.WorkloadIdentityIDs))
	default:
		return fmt.Sprintf("%s (%d user-assigned)", identityType, len(envVar.WorkloadIdentityIDs))
	}
}

func envVarReferenceLabel(envVar models.EnvVarSummary) string {
	target := valueOrEmpty(envVar.ReferenceTarget)
	refIdentity := valueOrEmpty(envVar.KeyVaultReferenceIdentity)
	switch {
	case target == "":
		return "-"
	case refIdentity == "":
		return target
	default:
		return target + "\nreference identity: " + refIdentity
	}
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

func envVarNote(envVar models.EnvVarSummary) string {
	if envVar.ValueType == "keyvault-ref" {
		return "This setting maps to Key Vault-backed configuration."
	}

	if envVar.LooksSensitive && envVar.ValueType == "plain-text" {
		return "This setting looks sensitive and remains exposed as plain-text management-plane app configuration."
	}

	if envVar.ValueType == "plain-text" {
		return "This setting remains visible through management-plane app configuration."
	}

	return strings.TrimSpace(envVar.Summary)
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
