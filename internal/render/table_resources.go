package render

import "harrierops-azure/internal/models"

func aksTable(payload models.AksOutput) string {
	rows := make([][]string, 0, len(payload.AksClusters))
	for _, cluster := range payload.AksClusters {
		rows = append(rows, []string{
			cluster.Name,
			aksVersionContext(cluster),
			aksEndpointContext(cluster),
			aksIdentityContext(cluster),
			aksAuthContext(cluster),
			aksNetworkContext(cluster),
			cluster.Summary,
		})
	}

	return renderListTable(
		"azurefox aks",
		[]string{"cluster", "version", "endpoint", "identity", "auth", "network", "why it matters"},
		rows,
		[]string{"No visible AKS clusters were confirmed from current scope.", "", "", "", "", "", ""},
		aksTakeaway(payload),
	)
}

func acrTable(payload models.AcrOutput) string {
	rows := make([][]string, 0, len(payload.Registries))
	for _, registry := range payload.Registries {
		rows = append(rows, []string{
			registry.Name,
			valueOrFallback(registry.LoginServer, "-"),
			resourceIdentityContext(registry.WorkloadIdentityType, registry.WorkloadIdentityIDs),
			acrAuthContext(registry),
			acrExposureContext(registry),
			acrDepthContext(registry),
			acrPostureContext(registry),
			registry.Summary,
		})
	}

	return renderListTable(
		"azurefox acr",
		[]string{"registry", "login server", "identity", "auth", "exposure", "depth", "posture", "why it matters"},
		rows,
		[]string{"No visible container registries were confirmed from current scope.", "", "", "", "", "", "", ""},
		acrTakeaway(payload),
	)
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
		"azurefox automation",
		[]string{"automation account", "identity", "execution", "triggers", "workers", "assets", "why it matters"},
		rows,
		[]string{"No visible Automation accounts were confirmed from current scope.", "", "", "", "", "", ""},
		automationTakeaway(payload),
	)
}

func devopsTable(payload models.DevopsOutput) string {
	rows := make([][]string, 0, len(payload.Pipelines))
	for _, pipeline := range payload.Pipelines {
		rows = append(rows, []string{
			pipeline.ProjectName,
			pipeline.Name,
			devopsRepositoryContext(pipeline),
			devopsTriggerContext(pipeline),
			devopsInjectionContext(pipeline),
			devopsAccessContext(pipeline),
			devopsSecretContext(pipeline),
			devopsTargetContext(pipeline),
			devopsNextReview(pipeline),
			pipeline.Summary,
		})
	}

	return renderListTable(
		"azurefox devops",
		[]string{"project", "pipeline", "source", "execution path", "injection", "control path", "secret support", "impact point", "next review", "why it matters"},
		rows,
		[]string{"No visible Azure DevOps build definitions were confirmed from current scope.", "", "", "", "", "", "", "", "", ""},
		devopsTakeaway(payload),
	)
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
		"azurefox databases",
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

	return renderListTable(
		"azurefox keyvault",
		[]string{"vault", "resource group", "public network", "default action", "private endpoint", "purge protection", "rbac mode"},
		rows,
		[]string{"No visible Key Vault assets were confirmed from current scope.", "", "", "", "", "", ""},
		keyVaultTakeaway(payload),
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

	return renderListTable(
		"azurefox storage",
		[]string{"account", "resource group", "exposure", "auth / transport", "protocols", "inventory"},
		rows,
		[]string{"No visible storage accounts were confirmed from current scope.", "", "", "", "", ""},
		storageTakeaway(payload),
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
		"azurefox snapshots-disks",
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
		"azurefox application-gateway",
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
		"azurefox dns",
		[]string{"zone", "kind", "inventory", "namespace", "why it matters"},
		rows,
		[]string{"No visible DNS zones were confirmed from current scope.", "", "", "", ""},
		dnsTakeaway(payload),
	)
}

func apiMgmtTable(payload models.ApiMgmtOutput) string {
	rows := make([][]string, 0, len(payload.ApiManagementServices))
	for _, service := range payload.ApiManagementServices {
		rows = append(rows, []string{
			service.Name,
			join(service.GatewayHostnames, ", "),
			resourceIdentityContext(service.WorkloadIdentityType, service.WorkloadIdentityIDs),
			apiMgmtInventoryContext(service),
			apiMgmtExposureContext(service),
			apiMgmtPostureContext(service),
			service.Summary,
		})
	}

	return renderListTable(
		"azurefox api-mgmt",
		[]string{"service", "gateway", "identity", "inventory", "exposure", "posture", "why it matters"},
		rows,
		[]string{"No visible API Management services were confirmed from current scope.", "", "", "", "", "", ""},
		apiMgmtTakeaway(payload),
	)
}
