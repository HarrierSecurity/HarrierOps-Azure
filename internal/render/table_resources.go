package render

import (
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

func aksTable(payload models.AksOutput) string {
	if len(payload.AksClusters) == 0 {
		return renderListTable(
			"ho-azure aks",
			[]string{"cluster", "endpoint", "identity", "auth"},
			nil,
			[]string{"No visible AKS clusters were confirmed from current scope.", "", "", ""},
			aksTakeaway(payload),
		)
	}

	sections := make([]string, 0, len(payload.AksClusters))
	for _, cluster := range payload.AksClusters {
		row := renderStructuredTableWithTitle(
			"",
			[]string{"cluster", "endpoint", "identity", "auth"},
			[][]string{{
				cluster.Name,
				aksEndpointContext(cluster),
				aksIdentityContext(cluster),
				aksAuthContext(cluster),
			}},
			false,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if signal := aksOperatorSignal(cluster); signal != "" {
			parts = append(parts, renderWrappedDetailTableWithWidth("operator signal", signal, rowWidth))
		}
		if note := aksNote(cluster); note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}

	return joinRenderedBlocks(sections) + "\n\nTakeaway: " + aksTakeaway(payload) + "\n"
}

func acrTable(payload models.AcrOutput) string {
	if len(payload.Registries) == 0 {
		return renderListTable(
			"ho-azure acr",
			[]string{"registry", "login server", "identity", "auth", "exposure"},
			nil,
			[]string{"No visible container registries were confirmed from current scope.", "", "", "", ""},
			acrTakeaway(payload),
		)
	}

	sections := make([]string, 0, len(payload.Registries))
	for _, registry := range payload.Registries {
		row := renderStructuredTableWithTitle(
			"",
			[]string{"registry", "login server", "identity", "auth", "exposure"},
			[][]string{{
				registry.Name,
				valueOrFallback(registry.LoginServer, "-"),
				resourceIdentityContext(registry.WorkloadIdentityType, registry.WorkloadIdentityIDs),
				acrAuthContext(registry),
				acrExposureContext(registry),
			}},
			false,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if signal := acrOperatorSignal(registry); signal != "" {
			parts = append(parts, renderWrappedDetailTableWithWidth("operator signal", signal, rowWidth))
		}
		if note := acrNote(registry); note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}

	return joinRenderedBlocks(sections) + "\n\nTakeaway: " + acrTakeaway(payload) + "\n"
}

func automationTable(payload models.AutomationOutput) string {
	rows := make([][]string, 0, len(payload.AutomationAccounts))
	for _, account := range payload.AutomationAccounts {
		rows = append(rows, []string{
			account.Name,
			automationIdentityContext(account),
			automationExecutionContext(account),
			automationTriggerContext(account),
			automationWorkerContext(account),
			automationAssetContext(account),
			account.Summary,
		})
	}

	return renderListTable(
		"ho-azure automation",
		[]string{"automation account", "identity", "execution", "triggers", "workers", "assets", "why it matters"},
		rows,
		[]string{"No visible Automation accounts were confirmed from current scope.", "", "", "", "", "", ""},
		automationTakeaway(payload),
	)
}

func azureMLTable(payload models.AzureMLOutput) string {
	if len(payload.Workspaces) == 0 {
		return renderListTable(
			"ho-azure azure-ml",
			[]string{"workspace", "runtime", "serving", "identity", "classification"},
			nil,
			[]string{"No visible Azure ML workspaces were confirmed from current scope.", "", "", "", ""},
			azureMLTakeaway(payload),
		)
	}

	sections := make([]string, 0, len(payload.Workspaces))
	for index, workspace := range payload.Workspaces {
		row := renderStructuredTableWithTitle(
			"ho-azure azure-ml",
			[]string{"workspace", "runtime", "serving", "identity", "classification"},
			[][]string{{
				workspace.Name,
				valueOrFallback(workspace.Runtime, "-"),
				valueOrFallback(workspace.Serving, "-"),
				valueOrFallback(workspace.Identity, "-"),
				workspace.Classification,
			}},
			index == 0,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		note := workspace.Summary
		if storage := azureMLStorageContext(workspace); storage != "" {
			note += " Storage context: " + storage + "."
		}
		if note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}

	return joinRenderedBlocks(sections) + "\n\nTakeaway: " + azureMLTakeaway(payload) + "\n"
}

func azureMLTakeaway(payload models.AzureMLOutput) string {
	if len(payload.Workspaces) == 0 {
		return "No visible Azure ML workspaces were confirmed from current scope."
	}

	executionCapable := 0
	persistenceContext := 0
	for _, workspace := range payload.Workspaces {
		switch workspace.Classification {
		case "execution-capable":
			executionCapable++
		case "supporting-persistence-context":
			persistenceContext++
		}
	}

	return fmt.Sprintf(
		"%d workspaces surfaced; %d execution-capable and %d persistence-adjacent from the current control-plane read path.",
		len(payload.Workspaces),
		executionCapable,
		persistenceContext,
	)
}

func azureMLStorageContext(workspace models.AzureMLWorkspaceAsset) string {
	parts := []string{}
	if workspace.DatastoreCount > 0 {
		parts = append(parts, fmt.Sprintf("%d visible datastore relationships", workspace.DatastoreCount))
	}
	if workspace.StorageAccountID != nil && *workspace.StorageAccountID != "" {
		parts = append(parts, "linked storage account")
	}
	if workspace.KeyVaultID != nil && *workspace.KeyVaultID != "" {
		parts = append(parts, "linked key vault")
	}
	if workspace.ContainerRegistryID != nil && *workspace.ContainerRegistryID != "" {
		parts = append(parts, "linked container registry")
	}
	if workspace.ApplicationInsightsID != nil && *workspace.ApplicationInsightsID != "" {
		parts = append(parts, "linked app insights")
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ")
}

func eventGridTable(payload models.EventGridOutput) string {
	if len(payload.Routes) == 0 {
		return renderListTable(
			"ho-azure event-grid",
			[]string{"source", "destination", "destination type", "classification"},
			nil,
			[]string{"No visible Event Grid routes were confirmed from current scope.", "", "", ""},
			eventGridTakeaway(payload),
		)
	}

	sections := make([]string, 0, len(payload.Routes))
	for index, route := range payload.Routes {
		row := renderStructuredTableWithTitle(
			"ho-azure event-grid",
			[]string{"source", "destination", "destination type", "classification"},
			[][]string{{
				valueOrFallback(route.Source, "-"),
				valueOrFallback(route.Destination, "-"),
				route.DestinationType,
				route.Classification,
			}},
			index == 0,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if route.Summary != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(route.Summary, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}

	return joinRenderedBlocks(sections) + "\n\nTakeaway: " + eventGridTakeaway(payload) + "\n"
}

func eventGridTakeaway(payload models.EventGridOutput) string {
	if len(payload.Routes) == 0 {
		return "No visible Event Grid routes were confirmed from current scope."
	}

	executionCapable := 0
	externalCallbacks := 0
	for _, route := range payload.Routes {
		switch route.Classification {
		case "execution-capable":
			executionCapable++
		case "external-callback":
			externalCallbacks++
		}
	}

	return fmt.Sprintf(
		"%d routes surfaced; %d execution-capable and %d external-callback from the current control-plane read path.",
		len(payload.Routes),
		executionCapable,
		externalCallbacks,
	)
}

func logicAppsTable(payload models.LogicAppsOutput) string {
	if len(payload.Workflows) == 0 {
		return renderListTable(
			"ho-azure logic-apps",
			[]string{"logic app", "trigger", "identity", "downstream", "classification"},
			nil,
			[]string{"No visible Logic Apps workflows were confirmed from current scope.", "", "", "", ""},
			logicAppsTakeaway(payload),
		)
	}

	sections := make([]string, 0, len(payload.Workflows))
	for index, workflow := range payload.Workflows {
		row := renderStructuredTableWithTitle(
			"ho-azure logic-apps",
			[]string{"logic app", "trigger", "identity", "downstream", "classification"},
			[][]string{{
				workflow.Name,
				valueOrFallback(workflow.Trigger, "-"),
				valueOrFallback(workflow.Identity, "-"),
				valueOrFallback(workflow.Downstream, "-"),
				workflow.Classification,
			}},
			index == 0,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if note := workflow.Summary; note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}

	return joinRenderedBlocks(sections) + "\n\nTakeaway: " + logicAppsTakeaway(payload) + "\n"
}

func logicAppsTakeaway(payload models.LogicAppsOutput) string {
	if len(payload.Workflows) == 0 {
		return "No visible Logic Apps workflows were confirmed from current scope."
	}

	persistenceCapable := 0
	executionOnly := 0
	for _, workflow := range payload.Workflows {
		switch workflow.Classification {
		case "persistence-capable":
			persistenceCapable++
		case "execution-capable-only":
			executionOnly++
		}
	}

	return fmt.Sprintf(
		"%d workflows surfaced; %d persistence-capable and %d execution-capable-only from the current control-plane read path.",
		len(payload.Workflows),
		persistenceCapable,
		executionOnly,
	)
}

func devopsTable(payload models.DevopsOutput) string {
	if len(payload.Pipelines) == 0 {
		return renderListTable(
			"ho-azure devops",
			[]string{"info"},
			nil,
			[]string{"No visible Azure DevOps build definitions were confirmed from current scope."},
			"",
		)
	}

	sections := make([]string, 0, len(payload.Pipelines))
	for index, pipeline := range payload.Pipelines {
		row := renderStructuredTableWithTitle(
			"ho-azure devops",
			[]string{"project", "pipeline", "source", "start path", "injection", "impact point"},
			[][]string{{
				pipeline.ProjectName,
				pipeline.Name,
				devopsRepositoryContext(pipeline),
				devopsTriggerContext(pipeline),
				devopsInjectionContext(pipeline),
				devopsTargetContext(pipeline),
			}},
			index == 0,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if operatorSignal := devopsOperatorSignal(pipeline); operatorSignal != "" {
			parts = append(parts, renderWrappedDetailTableWithWidth("operator signal", operatorSignal, rowWidth))
		}
		if note := devopsNote(pipeline); note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}

	return joinRenderedBlocks(sections) + "\n\nTakeaway: " + devopsTakeaway(payload) + "\n"
}

func apiMgmtTable(payload models.ApiMgmtOutput) string {
	if len(payload.ApiManagementServices) == 0 {
		return renderListTable(
			"ho-azure api-mgmt",
			[]string{"service", "gateway", "identity", "exposure"},
			nil,
			[]string{"No visible API Management services were confirmed from current scope.", "", "", ""},
			apiMgmtTakeaway(payload),
		)
	}

	sections := make([]string, 0, len(payload.ApiManagementServices))
	for _, service := range payload.ApiManagementServices {
		row := renderStructuredTableWithTitle(
			"",
			[]string{"service", "gateway", "identity", "exposure"},
			[][]string{{
				service.Name,
				apiMgmtGatewayLabel(service),
				resourceIdentityContext(service.WorkloadIdentityType, service.WorkloadIdentityIDs),
				apiMgmtExposureContext(service),
			}},
			false,
		)
		rowWidth := renderedTableCellWidth(row)
		parts := []string{row}
		if signal := apiMgmtOperatorSignal(service); signal != "" {
			parts = append(parts, renderWrappedDetailTableWithWidth("operator signal", signal, rowWidth))
		}
		if note := apiMgmtNote(service); note != "" {
			parts = append(parts, renderWrappedNoteTableWithWidth(note, rowWidth))
		}
		sections = append(sections, joinRenderedSections(parts...))
	}

	return joinRenderedBlocks(sections) + "\n\nTakeaway: " + apiMgmtTakeaway(payload) + "\n"
}

func databasesTable(payload models.DatabasesOutput) string {
	rows := make([][]string, 0, len(payload.DatabaseServers))
	for _, server := range payload.DatabaseServers {
		rows = append(rows, []string{
			server.Name,
			server.Engine,
			valueOrEmpty(server.FullyQualifiedDomainName),
			resourceIdentityContext(server.WorkloadIdentityType, server.WorkloadIdentityIDs),
			databasesInventoryContext(server),
			databasesExposureContext(server),
			databasesPostureContext(server),
			server.Summary,
		})
	}

	return renderListTable(
		"ho-azure databases",
		[]string{"server", "engine", "endpoint", "identity", "inventory", "exposure", "posture", "why it matters"},
		rows,
		[]string{"No visible relational database servers were confirmed from current scope.", "", "", "", "", "", "", ""},
		databasesTakeaway(payload),
	)
}

func keyVaultTable(payload models.KeyVaultOutput) string {
	rows := make([][]string, 0, len(payload.KeyVaults))
	for _, vault := range payload.KeyVaults {
		rows = append(rows, []string{
			vault.Name,
			vault.ResourceGroup,
			valueOrFallback(vault.PublicNetworkAccess, "-"),
			keyVaultTableDefaultAction(vault),
			yesNo(vault.PrivateEndpointEnabled),
			yesNo(vault.PurgeProtectionEnabled),
			yesNo(vault.EnableRBACAuthorization),
		})
	}

	output := renderListTable(
		"ho-azure keyvault",
		[]string{"vault", "resource group", "public network", "default action", "private endpoint", "purge protection", "rbac mode"},
		rows,
		[]string{"No visible Key Vault assets were confirmed from current scope.", "", "", "", "", "", ""},
		keyVaultTakeaway(payload),
	)
	return appendCustomFindingsSection(output, payload.Findings,
		func(f models.KeyVaultFinding) string { return f.Severity },
		func(f models.KeyVaultFinding) string { return f.Title },
		func(f models.KeyVaultFinding) string { return f.Description },
	)
}

func storageTable(payload models.StorageOutput) string {
	rows := make([][]string, 0, len(payload.StorageAssets))
	for _, asset := range payload.StorageAssets {
		rows = append(rows, []string{
			asset.Name,
			asset.ResourceGroup,
			storageExposureContext(asset),
			storageAuthContext(asset),
			storageProtocolContext(asset),
			storageInventoryContext(asset),
		})
	}

	output := renderListTable(
		"ho-azure storage",
		[]string{"account", "resource group", "exposure", "auth / transport", "protocols", "inventory"},
		rows,
		[]string{"No visible storage accounts were confirmed from current scope.", "", "", "", "", ""},
		storageTakeaway(payload),
	)
	return appendCustomFindingsSection(output, payload.Findings,
		func(f models.StorageFinding) string { return f.Severity },
		func(f models.StorageFinding) string { return f.Title },
		func(f models.StorageFinding) string { return f.Description },
	)
}

func snapshotsDisksTable(payload models.SnapshotsDisksOutput) string {
	rows := make([][]string, 0, len(payload.SnapshotDiskAssets))
	for _, asset := range payload.SnapshotDiskAssets {
		rows = append(rows, []string{
			asset.Name,
			asset.AssetKind,
			snapshotDiskPriorityContext(asset),
			snapshotDiskAttachmentContext(asset),
			snapshotDiskSharingContext(asset),
			snapshotDiskEncryptionContext(asset),
			asset.Summary,
		})
	}

	return renderListTable(
		"ho-azure snapshots-disks",
		[]string{"asset", "kind", "priority", "attachment / source", "sharing / export", "encryption", "why it matters"},
		rows,
		[]string{"No visible managed disks or snapshots were confirmed from current scope.", "", "", "", "", "", ""},
		snapshotsDisksTakeaway(payload),
	)
}

func applicationGatewayTable(payload models.ApplicationGatewayOutput) string {
	rows := make([][]string, 0, len(payload.ApplicationGateways))
	for _, gateway := range payload.ApplicationGateways {
		rows = append(rows, []string{
			gateway.Name,
			applicationGatewayExposureContext(gateway),
			applicationGatewayRoutingContext(gateway),
			applicationGatewayBackendContext(gateway),
			applicationGatewayWAFContext(gateway),
			gateway.Summary,
		})
	}

	return renderListTable(
		"ho-azure application-gateway",
		[]string{"gateway", "exposure", "routing", "backends", "waf", "why it matters"},
		rows,
		[]string{"No visible Application Gateways were confirmed from current scope.", "", "", "", "", ""},
		applicationGatewayTakeaway(payload),
	)
}

func dnsTable(payload models.DnsOutput) string {
	rows := make([][]string, 0, len(payload.DNSZones))
	for _, zone := range payload.DNSZones {
		rows = append(rows, []string{
			zone.Name,
			zone.ZoneKind,
			dnsInventoryContext(zone),
			dnsNamespaceContext(zone),
			zone.Summary,
		})
	}

	return renderListTable(
		"ho-azure dns",
		[]string{"zone", "kind", "inventory", "namespace", "why it matters"},
		rows,
		[]string{"No visible DNS zones were confirmed from current scope.", "", "", "", ""},
		dnsTakeaway(payload),
	)
}
