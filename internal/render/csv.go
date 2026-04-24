package render

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

var chainsFamilyCSVRenderers = map[string]func(models.ChainsOutput) (string, error){
	"compute-control": chainsComputeControlCSV,
	"credential-path": chainsCredentialPathCSV,
	"deployment-path": chainsDeploymentPathCSV,
	"escalation-path": chainsEscalationPathCSV,
}

func CSV(command string, payload any) (string, error) {
	entry, err := renderRegistryEntry(command)
	if err != nil {
		return "", err
	}
	if entry.csv == nil {
		return "", fmt.Errorf("csv rendering is not implemented for command %q", command)
	}
	return entry.csv(payload)
}

func chainsCSVRenderer(payload any) (string, error) {
	switch out := payload.(type) {
	case models.ChainsOverviewOutput:
		return chainsOverviewCSV(out)
	case models.ChainsOutput:
		return chainsFamilyCSV(out)
	default:
		return "", fmt.Errorf("unexpected payload type for chains: %T", payload)
	}
}

func persistenceCSVRenderer(payload any) (string, error) {
	switch out := payload.(type) {
	case models.PersistenceOverviewOutput:
		return persistenceOverviewCSV(out)
	case models.PersistenceAutomationOutput:
		return persistenceAutomationCSV(out)
	case models.PersistenceAppServiceOutput:
		return persistenceAppServiceCSV(out)
	case models.PersistenceWebJobsOutput:
		return persistenceWebJobsCSV(out)
	case models.PersistenceContainerAppsJobsOutput:
		return persistenceContainerAppsJobsCSV(out)
	case models.PersistenceVMExtensionsOutput:
		return persistenceVMExtensionsCSV(out)
	case models.PersistenceAzureMLOutput:
		return persistenceAzureMLCSV(out)
	case models.PersistenceFunctionsOutput:
		return persistenceFunctionsCSV(out)
	case models.PersistenceLogicAppsOutput:
		return persistenceLogicAppsCSV(out)
	default:
		return "", fmt.Errorf("unexpected payload type for persistence: %T", payload)
	}
}

func evasionCSVRenderer(payload any) (string, error) {
	switch out := payload.(type) {
	case models.EvasionOverviewOutput:
		return evasionOverviewCSV(out)
	case models.EvasionDCROutput:
		return evasionDCRCSV(out)
	case models.EvasionDiagnosticSettingsOutput:
		return evasionDiagnosticSettingsCSV(out)
	case models.EvasionAppInsightsOutput:
		return evasionAppInsightsCSV(out)
	default:
		return "", fmt.Errorf("unexpected payload type for evasion: %T", payload)
	}
}

func resourceHijackingCSVRenderer(payload any) (string, error) {
	switch out := payload.(type) {
	case models.ResourceHijackingOverviewOutput:
		return resourceHijackingOverviewCSV(out)
	case models.ResourceHijackingAPIMOutput:
		return resourceHijackingAPIMCSV(out)
	case models.ResourceHijackingAutomationOutput:
		return resourceHijackingAutomationCSV(out)
	case models.ResourceHijackingLogicAppsOutput:
		return resourceHijackingLogicAppsCSV(out)
	default:
		return "", fmt.Errorf("unexpected payload type for resourcehijacking: %T", payload)
	}
}

func pathMaskingCSVRenderer(payload any) (string, error) {
	switch out := payload.(type) {
	case models.PathMaskingOverviewOutput:
		return pathMaskingOverviewCSV(out)
	case models.PathMaskingAPIMOutput:
		return pathMaskingAPIMCSV(out)
	case models.PathMaskingLogicAppsOutput:
		return pathMaskingLogicAppsCSV(out)
	case models.PathMaskingRelayOutput:
		return pathMaskingRelayCSV(out)
	default:
		return "", fmt.Errorf("unexpected payload type for pathmasking: %T", payload)
	}
}

func encodeCSV(headers []string, rows [][]string) (string, error) {
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)

	if err := writer.Write(headers); err != nil {
		return "", err
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

type csvColumn[T any] struct {
	header string
	value  func(T) string
}

func encodeCSVColumns[T any](columns []csvColumn[T], values []T) (string, error) {
	headers := make([]string, 0, len(columns))
	for _, column := range columns {
		headers = append(headers, column.header)
	}
	rows := make([][]string, 0, len(values))
	for _, value := range values {
		row := make([]string, 0, len(columns))
		for _, column := range columns {
			row = append(row, column.value(value))
		}
		rows = append(rows, row)
	}
	return encodeCSV(headers, rows)
}

func rbacCSV(payload models.RbacOutput) (string, error) {
	headers := []string{
		"id",
		"principal_id",
		"principal_type",
		"role_definition_id",
		"role_name",
		"scope_id",
	}
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
	return encodeCSV(headers, rows)
}

func automationCSV(payload models.AutomationOutput) (string, error) {
	rows := make([][]string, 0, len(payload.AutomationAccounts))
	for _, account := range payload.AutomationAccounts {
		rows = append(rows, []string{
			account.ID,
			account.Name,
			account.ResourceGroup,
			valueOrEmpty(account.Location),
			valueOrEmpty(account.State),
			valueOrEmpty(account.SKUName),
			valueOrEmpty(account.IdentityType),
			valueOrEmpty(account.PrincipalID),
			valueOrEmpty(account.ClientID),
			jsonStringSlice(account.IdentityIDs),
			intPtrString(account.RunbookCount),
			intPtrString(account.PublishedRunbookCount),
			jsonStringSlice(account.PublishedRunbookNames),
			jsonStringSlice(account.RunbookTypes),
			jsonStringSlice(account.RunbookCommandClues),
			jsonStringSlice(account.RunbookResourceClues),
			intPtrString(account.ScheduleCount),
			jsonStringSlice(account.ScheduleDefinitions),
			intPtrString(account.JobScheduleCount),
			intPtrString(account.WebhookCount),
			intPtrString(account.HybridWorkerGroupCount),
			intPtrString(account.CredentialCount),
			intPtrString(account.CertificateCount),
			intPtrString(account.ConnectionCount),
			intPtrString(account.VariableCount),
			intPtrString(account.EncryptedVariableCount),
			jsonStringSlice(account.StartModes),
			valueOrEmpty(account.PrimaryStartMode),
			valueOrEmpty(account.PrimaryRunbookName),
			jsonStringSlice(account.ScheduleRunbookNames),
			jsonStringSlice(account.WebhookRunbookNames),
			jsonStringSlice(account.HybridWorkerGroupIDs),
			jsonStringSlice(account.TriggerJoinIDs),
			jsonStringSlice(account.IdentityJoinIDs),
			jsonStringSlice(account.SecretSupportTypes),
			jsonStringSlice(account.SecretDependencyIDs),
			jsonStringSlice(account.ConsequenceTypes),
			boolString(account.MissingExecutionPath),
			boolString(account.MissingTargetMapping),
			account.Summary,
			jsonStringSlice(account.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"name",
		"resource_group",
		"location",
		"state",
		"sku_name",
		"identity_type",
		"principal_id",
		"client_id",
		"identity_ids",
		"runbook_count",
		"published_runbook_count",
		"published_runbook_names",
		"runbook_types",
		"runbook_command_clues",
		"runbook_resource_clues",
		"schedule_count",
		"schedule_definitions",
		"job_schedule_count",
		"webhook_count",
		"hybrid_worker_group_count",
		"credential_count",
		"certificate_count",
		"connection_count",
		"variable_count",
		"encrypted_variable_count",
		"start_modes",
		"primary_start_mode",
		"primary_runbook_name",
		"schedule_runbook_names",
		"webhook_runbook_names",
		"hybrid_worker_group_ids",
		"trigger_join_ids",
		"identity_join_ids",
		"secret_support_types",
		"secret_dependency_ids",
		"consequence_types",
		"missing_execution_path",
		"missing_target_mapping",
		"summary",
		"related_ids",
	}, rows)
}

func dcrCSV(payload models.DCROutput) (string, error) {
	rows := make([][]string, 0, len(payload.DCRs))
	for _, dcr := range payload.DCRs {
		rows = append(rows, []string{
			dcr.ID,
			dcr.Name,
			dcr.ResourceGroup,
			dcr.Location,
			valueOrEmpty(dcr.Kind),
			valueOrEmpty(dcr.Description),
			valueOrEmpty(dcr.DataCollectionEndpointID),
			jsonStringSlice(dcr.DataSourceTypes),
			jsonStringSlice(dcr.Streams),
			jsonStringSlice(dcr.HighSignalStreams),
			jsonStringSlice(dcr.DestinationTypes),
			intString(dcr.TransformationCount),
			intString(dcr.AssociationCount),
			jsonValue(dcr.DataSources),
			jsonValue(dcr.DataFlows),
			jsonValue(dcr.Destinations),
			jsonValue(dcr.Associations),
			dcr.Summary,
			jsonStringSlice(dcr.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"name",
		"resource_group",
		"location",
		"kind",
		"description",
		"data_collection_endpoint_id",
		"data_source_types",
		"streams",
		"high_signal_streams",
		"destination_types",
		"transformation_count",
		"association_count",
		"data_sources",
		"data_flows",
		"destinations",
		"associations",
		"summary",
		"related_ids",
	}, rows)
}

func diagnosticSettingsCSV(payload models.DiagnosticSettingsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Sources))
	for _, source := range payload.Sources {
		rows = append(rows, []string{
			source.ID,
			source.Name,
			source.Type,
			source.ResourceGroup,
			source.Location,
			intString(source.DiagnosticSettingCount),
			boolString(source.HasDiagnosticSettings),
			boolString(source.HasPartialLogPosture),
			boolString(source.HasHighSignalLogPosture),
			boolString(source.HasNonWorkspaceDestination),
			jsonStringSlice(source.EnabledCategories),
			jsonStringSlice(source.DisabledCategories),
			jsonStringSlice(source.SupportedCategories),
			jsonStringSlice(source.NotExportedSupported),
			boolString(source.SupportedCategoryCatalog),
			jsonStringSlice(source.CategoryGroups),
			jsonStringSlice(source.HighSignalCategories),
			jsonStringSlice(source.DestinationTypes),
			jsonValue(source.DiagnosticSettings),
			source.Summary,
			jsonStringSlice(source.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"name",
		"type",
		"resource_group",
		"location",
		"diagnostic_setting_count",
		"has_diagnostic_settings",
		"has_partial_log_posture",
		"has_high_signal_log_posture",
		"has_non_workspace_destination",
		"enabled_categories",
		"disabled_categories",
		"supported_categories",
		"not_exported_supported_categories",
		"supported_category_catalog",
		"category_groups",
		"high_signal_categories",
		"destination_types",
		"diagnostic_settings",
		"summary",
		"related_ids",
	}, rows)
}

func monitoringSinksCSV(payload models.MonitoringSinksOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Sinks))
	for _, sink := range payload.Sinks {
		rows = append(rows, []string{
			sink.ID,
			sink.Name,
			sink.Kind,
			sink.ResourceType,
			sink.ResourceGroup,
			sink.Location,
			sink.VisibilitySource,
			boolPtrString(sink.SentinelEnabled),
			intString(sink.ReferenceCount),
			jsonValue(sink.References),
			sink.Summary,
			jsonStringSlice(sink.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"name",
		"kind",
		"resource_type",
		"resource_group",
		"location",
		"visibility_source",
		"sentinel_enabled",
		"reference_count",
		"references",
		"summary",
		"related_ids",
	}, rows)
}

func appInsightsCSV(payload models.AppInsightsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Targets))
	for _, target := range payload.Targets {
		rows = append(rows, []string{
			target.ID,
			target.Name,
			target.Kind,
			target.ResourceGroup,
			target.Location,
			jsonStringSlice(target.InstrumentationClues),
			jsonStringSlice(target.SamplingClues),
			jsonStringSlice(target.FilteringClues),
			jsonStringSlice(target.LoggingLevelClues),
			jsonStringSlice(target.VisibleTelemetryTypes),
			target.Summary,
			jsonStringSlice(target.RelatedIDs),
		})
	}
	return encodeCSV([]string{
		"id",
		"name",
		"kind",
		"resource_group",
		"location",
		"instrumentation_clues",
		"sampling_clues",
		"filtering_clues",
		"logging_level_clues",
		"visible_telemetry_types",
		"summary",
		"related_ids",
	}, rows)
}

func relayCSV(payload models.RelayOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Namespaces))
	for _, namespace := range payload.Namespaces {
		rows = append(rows, []string{
			namespace.ID,
			namespace.Name,
			namespace.ResourceGroup,
			valueOrEmpty(namespace.Location),
			valueOrEmpty(namespace.SKUName),
			valueOrEmpty(namespace.ProvisioningState),
			valueOrEmpty(namespace.ServiceBusEndpoint),
			intPtrString(namespace.HybridConnectionCount),
			intPtrString(namespace.AuthorizationRuleCount),
			relayHybridConnectionNames(namespace.HybridConnections),
			relayListenerSummary(namespace),
			relayAppServiceAttachmentNames(namespace.HybridConnections),
			namespace.Summary,
			jsonStringSlice(namespace.RelatedIDs),
		})
	}
	return encodeCSV([]string{
		"id",
		"namespace",
		"resource_group",
		"location",
		"sku_name",
		"provisioning_state",
		"service_bus_endpoint",
		"hybrid_connection_count",
		"authorization_rule_count",
		"hybrid_connections",
		"listeners",
		"app_service_attachments",
		"summary",
		"related_ids",
	}, rows)
}

func relayHybridConnectionNames(connections []models.RelayHybridConnectionAsset) string {
	values := make([]string, 0, len(connections))
	for _, connection := range connections {
		values = append(values, connection.Name)
	}
	return jsonStringSlice(values)
}

func relayAppServiceAttachmentNames(connections []models.RelayHybridConnectionAsset) string {
	values := make([]string, 0, len(connections))
	for _, connection := range connections {
		for _, app := range connection.AppServiceAttachments {
			values = append(values, connection.Name+"->"+app)
		}
	}
	return jsonStringSlice(values)
}

func evasionOverviewCSV(payload models.EvasionOverviewOutput) (string, error) {
	return familyOverviewCSV(payload.Surfaces)
}

func familyOverviewCSV(surfaces []models.FamilySurfaceDescriptor) (string, error) {
	return encodeCSVColumns([]csvColumn[models.FamilySurfaceDescriptor]{
		{header: "surface", value: func(surface models.FamilySurfaceDescriptor) string { return surface.Surface }},
		{header: "state", value: func(surface models.FamilySurfaceDescriptor) string { return surface.State }},
		{header: "summary", value: func(surface models.FamilySurfaceDescriptor) string { return surface.Summary }},
		{header: "operator_question", value: func(surface models.FamilySurfaceDescriptor) string { return surface.OperatorQuestion }},
		{header: "backing_commands", value: func(surface models.FamilySurfaceDescriptor) string { return jsonStringSlice(surface.BackingCommands) }},
	}, surfaces)
}

func evasionDCRCSV(payload models.EvasionDCROutput) (string, error) {
	return encodeCSVColumns([]csvColumn[models.EvasionDCR]{
		{"id", func(dcr models.EvasionDCR) string { return dcr.ID }},
		{"dcr", func(dcr models.EvasionDCR) string { return dcr.Name }},
		{"resource_group", func(dcr models.EvasionDCR) string { return dcr.ResourceGroup }},
		{"location", func(dcr models.EvasionDCR) string { return dcr.Location }},
		{"disruption_rank", func(dcr models.EvasionDCR) string { return intString(dcr.DisruptionRank) }},
		{"disruption_reason", func(dcr models.EvasionDCR) string { return dcr.DisruptionReason }},
		{"capability_steps", func(dcr models.EvasionDCR) string { return jsonValue(dcr.CapabilitySteps) }},
		{"current_identity_summary", func(dcr models.EvasionDCR) string {
			return familyRoleSummary(dcr.CurrentIdentityContext)
		}},
		{"current_state", func(dcr models.EvasionDCR) string { return jsonValue(dcr.CurrentState) }},
		{"not_collected_by_default", func(dcr models.EvasionDCR) string {
			return jsonValue(dcr.NotCollectedByDefault)
		}},
		{"summary", func(dcr models.EvasionDCR) string { return dcr.Summary }},
		{"related_ids", func(dcr models.EvasionDCR) string { return jsonStringSlice(dcr.RelatedIDs) }},
	}, payload.DCRs)
}

func evasionDiagnosticSettingsCSV(payload models.EvasionDiagnosticSettingsOutput) (string, error) {
	return encodeCSVColumns([]csvColumn[models.EvasionDiagnosticSettingsSource]{
		{"id", func(source models.EvasionDiagnosticSettingsSource) string { return source.ID }},
		{"source", func(source models.EvasionDiagnosticSettingsSource) string { return source.Name }},
		{"resource_group", func(source models.EvasionDiagnosticSettingsSource) string { return source.ResourceGroup }},
		{"location", func(source models.EvasionDiagnosticSettingsSource) string { return source.Location }},
		{"disruption_rank", func(source models.EvasionDiagnosticSettingsSource) string {
			return intString(source.DisruptionRank)
		}},
		{"disruption_reason", func(source models.EvasionDiagnosticSettingsSource) string {
			return source.DisruptionReason
		}},
		{"capability_steps", func(source models.EvasionDiagnosticSettingsSource) string {
			return jsonValue(source.CapabilitySteps)
		}},
		{"current_identity_summary", func(source models.EvasionDiagnosticSettingsSource) string {
			return familyRoleSummary(source.CurrentIdentityContext)
		}},
		{"current_state", func(source models.EvasionDiagnosticSettingsSource) string { return jsonValue(source.CurrentState) }},
		{"not_collected_by_default", func(source models.EvasionDiagnosticSettingsSource) string {
			return jsonValue(source.NotCollectedByDefault)
		}},
		{"summary", func(source models.EvasionDiagnosticSettingsSource) string { return source.Summary }},
		{"related_ids", func(source models.EvasionDiagnosticSettingsSource) string {
			return jsonStringSlice(source.RelatedIDs)
		}},
	}, payload.Sources)
}

func evasionAppInsightsCSV(payload models.EvasionAppInsightsOutput) (string, error) {
	return encodeCSVColumns([]csvColumn[models.EvasionAppInsightsTarget]{
		{"id", func(target models.EvasionAppInsightsTarget) string { return target.ID }},
		{"target", func(target models.EvasionAppInsightsTarget) string { return target.Name }},
		{"resource_group", func(target models.EvasionAppInsightsTarget) string { return target.ResourceGroup }},
		{"location", func(target models.EvasionAppInsightsTarget) string { return target.Location }},
		{"disruption_rank", func(target models.EvasionAppInsightsTarget) string {
			return intString(target.DisruptionRank)
		}},
		{"disruption_reason", func(target models.EvasionAppInsightsTarget) string { return target.DisruptionReason }},
		{"capability_steps", func(target models.EvasionAppInsightsTarget) string { return jsonValue(target.CapabilitySteps) }},
		{"current_identity_summary", func(target models.EvasionAppInsightsTarget) string {
			return familyRoleSummary(target.CurrentIdentityContext)
		}},
		{"current_state", func(target models.EvasionAppInsightsTarget) string { return jsonValue(target.CurrentState) }},
		{"not_collected_by_default", func(target models.EvasionAppInsightsTarget) string {
			return jsonValue(target.NotCollectedByDefault)
		}},
		{"summary", func(target models.EvasionAppInsightsTarget) string { return target.Summary }},
		{"related_ids", func(target models.EvasionAppInsightsTarget) string { return jsonStringSlice(target.RelatedIDs) }},
	}, payload.Targets)
}

func resourceHijackingOverviewCSV(payload models.ResourceHijackingOverviewOutput) (string, error) {
	return familyOverviewCSV(payload.Surfaces)
}

func resourceHijackingAPIMCSV(payload models.ResourceHijackingAPIMOutput) (string, error) {
	return encodeCSVColumns([]csvColumn[models.ResourceHijackingAPIMTarget]{
		{"id", func(target models.ResourceHijackingAPIMTarget) string { return target.ID }},
		{"api_management_service", func(target models.ResourceHijackingAPIMTarget) string { return target.Name }},
		{"resource_group", func(target models.ResourceHijackingAPIMTarget) string { return target.ResourceGroup }},
		{"location", func(target models.ResourceHijackingAPIMTarget) string { return valueOrEmpty(target.Location) }},
		{"takeover_rank", func(target models.ResourceHijackingAPIMTarget) string {
			return fmt.Sprintf("%d", target.TakeoverRank)
		}},
		{"takeover_reason", func(target models.ResourceHijackingAPIMTarget) string { return target.TakeoverReason }},
		{"gateway_hostnames", func(target models.ResourceHijackingAPIMTarget) string {
			return jsonStringSlice(target.CurrentState.GatewayHostnames)
		}},
		{"backend_hostnames", func(target models.ResourceHijackingAPIMTarget) string {
			return jsonStringSlice(target.CurrentState.BackendHostnames)
		}},
		{"api_count", func(target models.ResourceHijackingAPIMTarget) string {
			return intPtrString(target.CurrentState.APICount)
		}},
		{"active_subscription_count", func(target models.ResourceHijackingAPIMTarget) string {
			return intPtrString(target.CurrentState.ActiveSubscriptionCount)
		}},
		{"current_identity", func(target models.ResourceHijackingAPIMTarget) string {
			return familyRoleControlLabel(target.CurrentIdentityContext)
		}},
		{"summary", func(target models.ResourceHijackingAPIMTarget) string { return target.Summary }},
		{"related_ids", func(target models.ResourceHijackingAPIMTarget) string { return jsonStringSlice(target.RelatedIDs) }},
	}, payload.Targets)
}

func resourceHijackingAutomationCSV(payload models.ResourceHijackingAutomationOutput) (string, error) {
	return encodeCSVColumns([]csvColumn[models.ResourceHijackingAutomationTarget]{
		{"id", func(target models.ResourceHijackingAutomationTarget) string { return target.ID }},
		{"automation_account", func(target models.ResourceHijackingAutomationTarget) string { return target.Name }},
		{"resource_group", func(target models.ResourceHijackingAutomationTarget) string { return target.ResourceGroup }},
		{"location", func(target models.ResourceHijackingAutomationTarget) string { return valueOrEmpty(target.Location) }},
		{"takeover_rank", func(target models.ResourceHijackingAutomationTarget) string {
			return fmt.Sprintf("%d", target.TakeoverRank)
		}},
		{"takeover_reason", func(target models.ResourceHijackingAutomationTarget) string { return target.TakeoverReason }},
		{"published_runbook_count", func(target models.ResourceHijackingAutomationTarget) string {
			return intPtrString(target.CurrentState.PublishedRunbookCount)
		}},
		{"published_runbook_names", func(target models.ResourceHijackingAutomationTarget) string {
			return jsonStringSlice(target.CurrentState.PublishedRunbookNames)
		}},
		{"job_schedule_count", func(target models.ResourceHijackingAutomationTarget) string {
			return intPtrString(target.CurrentState.JobScheduleCount)
		}},
		{"webhook_count", func(target models.ResourceHijackingAutomationTarget) string {
			return intPtrString(target.CurrentState.WebhookCount)
		}},
		{"hybrid_worker_group_count", func(target models.ResourceHijackingAutomationTarget) string {
			return intPtrString(target.CurrentState.HybridWorkerGroupCount)
		}},
		{"identity_type", func(target models.ResourceHijackingAutomationTarget) string {
			return valueOrEmpty(target.CurrentState.IdentityType)
		}},
		{"current_identity", func(target models.ResourceHijackingAutomationTarget) string {
			return familyRoleControlLabel(target.CurrentIdentityContext)
		}},
		{"summary", func(target models.ResourceHijackingAutomationTarget) string { return target.Summary }},
		{"related_ids", func(target models.ResourceHijackingAutomationTarget) string {
			return jsonStringSlice(target.RelatedIDs)
		}},
	}, payload.Targets)
}

func resourceHijackingLogicAppsCSV(payload models.ResourceHijackingLogicAppsOutput) (string, error) {
	return encodeCSVColumns([]csvColumn[models.ResourceHijackingLogicAppTarget]{
		{"id", func(target models.ResourceHijackingLogicAppTarget) string { return target.ID }},
		{"logic_app", func(target models.ResourceHijackingLogicAppTarget) string { return target.Name }},
		{"resource_group", func(target models.ResourceHijackingLogicAppTarget) string { return target.ResourceGroup }},
		{"location", func(target models.ResourceHijackingLogicAppTarget) string { return valueOrEmpty(target.Location) }},
		{"takeover_rank", func(target models.ResourceHijackingLogicAppTarget) string {
			return fmt.Sprintf("%d", target.TakeoverRank)
		}},
		{"takeover_reason", func(target models.ResourceHijackingLogicAppTarget) string { return target.TakeoverReason }},
		{"trigger_types", func(target models.ResourceHijackingLogicAppTarget) string {
			return jsonStringSlice(target.CurrentState.TriggerTypes)
		}},
		{"externally_callable_request_trigger", func(target models.ResourceHijackingLogicAppTarget) string {
			return boolString(target.CurrentState.ExternallyCallableRequestTrigger)
		}},
		{"recurrence_summary", func(target models.ResourceHijackingLogicAppTarget) string {
			return valueOrEmpty(target.CurrentState.RecurrenceSummary)
		}},
		{"downstream_action_kinds", func(target models.ResourceHijackingLogicAppTarget) string {
			return jsonStringSlice(target.CurrentState.DownstreamActionKinds)
		}},
		{"identity_type", func(target models.ResourceHijackingLogicAppTarget) string {
			return valueOrEmpty(target.CurrentState.IdentityType)
		}},
		{"current_identity", func(target models.ResourceHijackingLogicAppTarget) string {
			return familyRoleControlLabel(target.CurrentIdentityContext)
		}},
		{"summary", func(target models.ResourceHijackingLogicAppTarget) string { return target.Summary }},
		{"related_ids", func(target models.ResourceHijackingLogicAppTarget) string {
			return jsonStringSlice(target.RelatedIDs)
		}},
	}, payload.Targets)
}

func pathMaskingOverviewCSV(payload models.PathMaskingOverviewOutput) (string, error) {
	return familyOverviewCSV(payload.Surfaces)
}

func pathMaskingAPIMCSV(payload models.PathMaskingAPIMOutput) (string, error) {
	return encodeCSVColumns([]csvColumn[models.PathMaskingAPIMTarget]{
		{"id", func(target models.PathMaskingAPIMTarget) string { return target.ID }},
		{"api_management_service", func(target models.PathMaskingAPIMTarget) string { return target.Name }},
		{"resource_group", func(target models.PathMaskingAPIMTarget) string { return target.ResourceGroup }},
		{"location", func(target models.PathMaskingAPIMTarget) string { return valueOrEmpty(target.Location) }},
		{"masking_rank", func(target models.PathMaskingAPIMTarget) string {
			return fmt.Sprintf("%d", target.MaskingRank)
		}},
		{"masking_reason", func(target models.PathMaskingAPIMTarget) string { return target.MaskingReason }},
		{"gateway_hostnames", func(target models.PathMaskingAPIMTarget) string {
			return jsonStringSlice(target.CurrentState.GatewayHostnames)
		}},
		{"backend_hostnames", func(target models.PathMaskingAPIMTarget) string {
			return jsonStringSlice(target.CurrentState.BackendHostnames)
		}},
		{"api_count", func(target models.PathMaskingAPIMTarget) string {
			return intPtrString(target.CurrentState.APICount)
		}},
		{"subscription_count", func(target models.PathMaskingAPIMTarget) string {
			return intPtrString(target.CurrentState.SubscriptionCount)
		}},
		{"current_identity", func(target models.PathMaskingAPIMTarget) string {
			return familyRoleControlLabel(target.CurrentIdentityContext)
		}},
		{"summary", func(target models.PathMaskingAPIMTarget) string { return target.Summary }},
		{"related_ids", func(target models.PathMaskingAPIMTarget) string { return jsonStringSlice(target.RelatedIDs) }},
	}, payload.Targets)
}

func pathMaskingLogicAppsCSV(payload models.PathMaskingLogicAppsOutput) (string, error) {
	return encodeCSVColumns([]csvColumn[models.PathMaskingLogicAppTarget]{
		{"id", func(target models.PathMaskingLogicAppTarget) string { return target.ID }},
		{"logic_app", func(target models.PathMaskingLogicAppTarget) string { return target.Name }},
		{"resource_group", func(target models.PathMaskingLogicAppTarget) string { return target.ResourceGroup }},
		{"location", func(target models.PathMaskingLogicAppTarget) string { return valueOrEmpty(target.Location) }},
		{"masking_rank", func(target models.PathMaskingLogicAppTarget) string {
			return fmt.Sprintf("%d", target.MaskingRank)
		}},
		{"masking_reason", func(target models.PathMaskingLogicAppTarget) string { return target.MaskingReason }},
		{"trigger_types", func(target models.PathMaskingLogicAppTarget) string {
			return jsonStringSlice(target.CurrentState.TriggerTypes)
		}},
		{"externally_callable_request_trigger", func(target models.PathMaskingLogicAppTarget) string {
			return boolString(target.CurrentState.ExternallyCallableRequestTrigger)
		}},
		{"recurrence_summary", func(target models.PathMaskingLogicAppTarget) string {
			return valueOrEmpty(target.CurrentState.RecurrenceSummary)
		}},
		{"downstream_action_kinds", func(target models.PathMaskingLogicAppTarget) string {
			return jsonStringSlice(target.CurrentState.DownstreamActionKinds)
		}},
		{"identity_type", func(target models.PathMaskingLogicAppTarget) string {
			return valueOrEmpty(target.CurrentState.IdentityType)
		}},
		{"current_identity", func(target models.PathMaskingLogicAppTarget) string {
			return familyRoleControlLabel(target.CurrentIdentityContext)
		}},
		{"summary", func(target models.PathMaskingLogicAppTarget) string { return target.Summary }},
		{"related_ids", func(target models.PathMaskingLogicAppTarget) string {
			return jsonStringSlice(target.RelatedIDs)
		}},
	}, payload.Targets)
}

func pathMaskingRelayCSV(payload models.PathMaskingRelayOutput) (string, error) {
	return encodeCSVColumns([]csvColumn[models.PathMaskingRelayTarget]{
		{"id", func(target models.PathMaskingRelayTarget) string { return target.ID }},
		{"relay_namespace", func(target models.PathMaskingRelayTarget) string { return target.Name }},
		{"resource_group", func(target models.PathMaskingRelayTarget) string { return target.ResourceGroup }},
		{"location", func(target models.PathMaskingRelayTarget) string { return valueOrEmpty(target.Location) }},
		{"masking_rank", func(target models.PathMaskingRelayTarget) string {
			return fmt.Sprintf("%d", target.MaskingRank)
		}},
		{"masking_reason", func(target models.PathMaskingRelayTarget) string { return target.MaskingReason }},
		{"service_bus_endpoint", func(target models.PathMaskingRelayTarget) string {
			return valueOrEmpty(target.CurrentState.ServiceBusEndpoint)
		}},
		{"hybrid_connection_count", func(target models.PathMaskingRelayTarget) string {
			return intPtrString(target.CurrentState.HybridConnectionCount)
		}},
		{"hybrid_connection_names", func(target models.PathMaskingRelayTarget) string {
			return jsonStringSlice(target.CurrentState.HybridConnectionNames)
		}},
		{"authorization_rule_count", func(target models.PathMaskingRelayTarget) string {
			return intPtrString(target.CurrentState.AuthorizationRuleCount)
		}},
		{"listener_summary", func(target models.PathMaskingRelayTarget) string {
			return target.CurrentState.ListenerSummary
		}},
		{"app_service_attachments", func(target models.PathMaskingRelayTarget) string {
			return jsonStringSlice(target.CurrentState.AppServiceAttachments)
		}},
		{"current_identity", func(target models.PathMaskingRelayTarget) string {
			return familyRoleControlLabel(target.CurrentIdentityContext)
		}},
		{"summary", func(target models.PathMaskingRelayTarget) string { return target.Summary }},
		{"related_ids", func(target models.PathMaskingRelayTarget) string { return jsonStringSlice(target.RelatedIDs) }},
	}, payload.Targets)
}

func eventGridCSV(payload models.EventGridOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Routes))
	for _, route := range payload.Routes {
		rows = append(rows, []string{
			route.ID,
			route.Name,
			valueOrEmpty(route.Source),
			valueOrEmpty(route.Destination),
			route.DestinationType,
			route.Classification,
			route.SourceID,
			route.SourceType,
			valueOrEmpty(route.DestinationTargetID),
			boolString(route.ExternalDelivery),
			valueOrEmpty(route.ProvisioningState),
			valueOrEmpty(route.IdentityType),
			valueOrEmpty(route.IdentityID),
			valueOrEmpty(route.EventDeliverySchema),
			jsonStringSlice(route.IncludedEventTypes),
			route.Summary,
			jsonStringSlice(route.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"route",
		"source",
		"destination",
		"destination_type",
		"classification",
		"source_id",
		"source_type",
		"destination_target_id",
		"external_delivery",
		"provisioning_state",
		"identity_type",
		"identity_id",
		"event_delivery_schema",
		"included_event_types",
		"summary",
		"related_ids",
	}, rows)
}

func azureMLCSV(payload models.AzureMLOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Workspaces))
	for _, workspace := range payload.Workspaces {
		rows = append(rows, []string{
			workspace.ID,
			workspace.Name,
			valueOrEmpty(workspace.Runtime),
			valueOrEmpty(workspace.Serving),
			valueOrEmpty(workspace.Identity),
			valueOrEmpty(workspace.Storage),
			workspace.Classification,
			workspace.ResourceGroup,
			valueOrEmpty(workspace.Location),
			valueOrEmpty(workspace.WorkspaceKind),
			valueOrEmpty(workspace.State),
			valueOrEmpty(workspace.PublicNetworkAccess),
			valueOrEmpty(workspace.IdentityType),
			valueOrEmpty(workspace.PrincipalID),
			jsonStringSlice(workspace.IdentityIDs),
			intString(workspace.ComputeCount),
			jsonStringSlice(workspace.ComputeTypes),
			intString(workspace.JobCount),
			jsonStringSlice(workspace.JobTypes),
			intString(workspace.ScheduleCount),
			jsonStringSlice(workspace.ScheduleTriggerTypes),
			intString(workspace.EndpointCount),
			jsonStringSlice(workspace.EndpointAuthModes),
			jsonStringSlice(workspace.EndpointPublicAccess),
			intString(workspace.DatastoreCount),
			jsonStringSlice(workspace.DatastoreTypes),
			valueOrEmpty(workspace.StorageAccountID),
			valueOrEmpty(workspace.KeyVaultID),
			valueOrEmpty(workspace.ContainerRegistryID),
			valueOrEmpty(workspace.ApplicationInsightsID),
			workspace.Summary,
			jsonStringSlice(workspace.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"workspace",
		"runtime",
		"serving",
		"identity",
		"storage",
		"classification",
		"resource_group",
		"location",
		"workspace_kind",
		"state",
		"public_network_access",
		"identity_type",
		"principal_id",
		"identity_ids",
		"compute_count",
		"compute_types",
		"job_count",
		"job_types",
		"schedule_count",
		"schedule_trigger_types",
		"endpoint_count",
		"endpoint_auth_modes",
		"endpoint_public_access",
		"datastore_count",
		"datastore_types",
		"storage_account_id",
		"key_vault_id",
		"container_registry_id",
		"application_insights_id",
		"summary",
		"related_ids",
	}, rows)
}

func logicAppsCSV(payload models.LogicAppsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Workflows))
	for _, workflow := range payload.Workflows {
		rows = append(rows, []string{
			workflow.ID,
			workflow.Name,
			valueOrEmpty(workflow.Trigger),
			valueOrEmpty(workflow.Identity),
			valueOrEmpty(workflow.Downstream),
			workflow.Classification,
			workflow.ResourceGroup,
			valueOrEmpty(workflow.Location),
			valueOrEmpty(workflow.Platform),
			valueOrEmpty(workflow.WorkflowKind),
			valueOrEmpty(workflow.State),
			valueOrEmpty(workflow.IdentityType),
			valueOrEmpty(workflow.PrincipalID),
			valueOrEmpty(workflow.ClientID),
			jsonStringSlice(workflow.IdentityIDs),
			intString(workflow.TriggerCount),
			intString(workflow.ActionCount),
			intString(workflow.BranchCount),
			jsonStringSlice(workflow.TriggerTypes),
			boolString(workflow.ExternallyCallableRequestTrigger),
			valueOrEmpty(workflow.RecurrenceSummary),
			jsonStringSlice(workflow.DownstreamActionKinds),
			jsonStringSlice(workflow.ConnectorReferences),
			jsonStringSlice(workflow.ParameterNames),
			jsonStringSlice(workflow.DownstreamResourceReferences),
			workflow.Summary,
			jsonStringSlice(workflow.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"logic_app",
		"trigger",
		"identity",
		"downstream",
		"classification",
		"resource_group",
		"location",
		"platform",
		"workflow_kind",
		"state",
		"identity_type",
		"principal_id",
		"client_id",
		"identity_ids",
		"trigger_count",
		"action_count",
		"branch_count",
		"trigger_types",
		"externally_callable_request_trigger",
		"recurrence_summary",
		"downstream_action_kinds",
		"connector_references",
		"parameter_names",
		"downstream_resource_references",
		"summary",
		"related_ids",
	}, rows)
}

func persistenceOverviewCSV(payload models.PersistenceOverviewOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Surfaces))
	for _, surface := range payload.Surfaces {
		rows = append(rows, []string{
			surface.Surface,
			surface.State,
			surface.Summary,
			surface.OperatorQuestion,
			jsonStringSlice(surface.BackingCommands),
		})
	}
	return encodeCSV(
		[]string{"surface", "state", "summary", "operator_question", "backing_commands"},
		rows,
	)
}

func persistenceAutomationCSV(payload models.PersistenceAutomationOutput) (string, error) {
	rows := make([][]string, 0, len(payload.AutomationAccounts))
	for _, account := range payload.AutomationAccounts {
		rows = append(rows, []string{
			account.ID,
			account.Name,
			account.ResourceGroup,
			valueOrEmpty(account.Location),
			persistenceCSVStepStatus(account.CapabilitySteps, "create or modify account"),
			persistenceCSVStepStatus(account.CapabilitySteps, "add or edit runbook"),
			persistenceCSVStepStatus(account.CapabilitySteps, "upload or replace code"),
			persistenceCSVStepStatus(account.CapabilitySteps, "publish runbook"),
			persistenceCSVStepStatus(account.CapabilitySteps, "attach or reuse exec ctx"),
			persistenceCSVStepStatus(account.CapabilitySteps, "create schedule"),
			persistenceCSVStepStatus(account.CapabilitySteps, "link schedule to runbook"),
			persistenceCSVStepStatus(account.CapabilitySteps, "create webhook"),
			persistenceCSVRoleSummary(account.CurrentIdentityContext),
			jsonStringSlice(account.ExecutionContextOptions),
			persistenceCSVRoleSummary(account.CurrentState.StrongestVisibleExecutionContext),
			intPtrString(account.CurrentState.RunbookCount),
			intPtrString(account.CurrentState.PublishedRunbookCount),
			jsonStringSlice(account.CurrentState.PublishedRunbookNames),
			intPtrString(account.CurrentState.ScheduleCount),
			jsonStringSlice(account.CurrentState.ScheduleDefinitions),
			intPtrString(account.CurrentState.JobScheduleCount),
			intPtrString(account.CurrentState.WebhookCount),
			intPtrString(account.CurrentState.HybridWorkerGroupCount),
			intPtrString(account.CurrentState.CredentialCount),
			intPtrString(account.CurrentState.CertificateCount),
			intPtrString(account.CurrentState.ConnectionCount),
			intPtrString(account.CurrentState.VariableCount),
			intPtrString(account.CurrentState.EncryptedVariableCount),
			valueOrEmpty(account.CurrentState.PrimaryStartMode),
			valueOrEmpty(account.CurrentState.PrimaryRunbookName),
			valueOrEmpty(account.CurrentState.IdentityType),
			jsonStringSlice(account.CurrentState.NearbyThematicNames),
			boolString(account.CurrentState.MissingTargetMapping),
			jsonStringSlice(account.StillUnmapped),
			account.Summary,
			jsonStringSlice(account.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"automation_account",
		"resource_group",
		"location",
		"create_or_modify_account",
		"add_or_edit_runbook",
		"upload_or_replace_code",
		"publish_runbook",
		"attach_or_reuse_exec_ctx",
		"create_schedule",
		"link_schedule_to_runbook",
		"create_webhook",
		"current_identity_context",
		"execution_context_options",
		"strongest_visible_execution_context",
		"runbook_count",
		"published_runbook_count",
		"published_runbook_names",
		"schedule_count",
		"schedule_definitions",
		"job_schedule_count",
		"webhook_count",
		"hybrid_worker_group_count",
		"credential_count",
		"certificate_count",
		"connection_count",
		"variable_count",
		"encrypted_variable_count",
		"primary_start_mode",
		"primary_runbook_name",
		"identity_type",
		"nearby_thematic_names",
		"missing_target_mapping",
		"still_unmapped",
		"summary",
		"related_ids",
	}, rows)
}

func persistenceAppServiceCSV(payload models.PersistenceAppServiceOutput) (string, error) {
	rows := make([][]string, 0, len(payload.AppServices))
	for _, app := range payload.AppServices {
		rows = append(rows, []string{
			app.ID,
			app.Name,
			app.ResourceGroup,
			app.Location,
			jsonStringSlice(persistenceCapabilityStepsCSV(app.CapabilitySteps)),
			jsonStringSlice(persistenceRoleContextCSV(app.CurrentIdentityContext)),
			jsonStringSlice(app.ExecutionContextOptions),
			valueOrEmpty(app.CurrentState.State),
			valueOrEmpty(app.CurrentState.Hostname),
			valueOrEmpty(app.CurrentState.PublicNetworkAccess),
			valueOrEmpty(app.CurrentState.Runtime),
			valueOrEmpty(app.CurrentState.Deployment),
			valueOrEmpty(app.CurrentState.DeploymentRepoURL),
			valueOrEmpty(app.CurrentState.DeploymentBranch),
			boolPtrString(app.CurrentState.DeploymentIsGitHubAction),
			boolPtrString(app.CurrentState.DeploymentManualIntegration),
			valueOrEmpty(app.CurrentState.IdentityType),
			intPtrString(app.CurrentState.AppSettingsCount),
			intPtrString(app.CurrentState.KeyVaultReferenceCount),
			intPtrString(app.CurrentState.SensitiveSettingCount),
			intPtrString(app.CurrentState.ConnectionStringCount),
			intPtrString(app.CurrentState.KeyVaultConnectionStringCount),
			jsonStringSlice(app.CurrentState.ConnectionStringTypes),
			boolPtrString(app.CurrentState.RunFromPackage),
			boolPtrString(app.CurrentState.HTTPSOnly),
			valueOrEmpty(app.CurrentState.MinTLSVersion),
			valueOrEmpty(app.CurrentState.FTPSState),
			jsonStringSlice(app.CurrentState.VisibleSensitiveSettingNames),
			jsonStringSlice(persistenceRoleContextCSV(app.CurrentState.StrongestVisibleExecutionContext)),
			jsonStringSlice(app.CurrentState.NearbyThematicNames),
			jsonStringSlice(app.StillUnmapped),
			app.Summary,
			jsonStringSlice(app.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"app_service",
		"resource_group",
		"location",
		"capability_steps",
		"current_identity_context",
		"execution_context_options",
		"state",
		"hostname",
		"public_network_access",
		"runtime",
		"deployment",
		"deployment_repo_url",
		"deployment_branch",
		"deployment_is_github_action",
		"deployment_manual_integration",
		"identity_type",
		"app_settings_count",
		"key_vault_reference_count",
		"sensitive_setting_count",
		"connection_string_count",
		"key_vault_connection_string_count",
		"connection_string_types",
		"run_from_package",
		"https_only",
		"min_tls_version",
		"ftps_state",
		"visible_sensitive_setting_names",
		"strongest_visible_execution_context",
		"nearby_thematic_names",
		"still_unmapped",
		"summary",
		"related_ids",
	}, rows)
}

func persistenceWebJobsCSV(payload models.PersistenceWebJobsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.WebJobs))
	for _, job := range payload.WebJobs {
		rows = append(rows, []string{
			job.ID,
			job.Name,
			job.ResourceGroup,
			job.Location,
			jsonStringSlice(persistenceCapabilityStepsCSV(job.CapabilitySteps)),
			jsonStringSlice(persistenceRoleContextCSV(job.CurrentIdentityContext)),
			jsonStringSlice(job.ExecutionContextOptions),
			job.CurrentState.Mode,
			valueOrEmpty(job.CurrentState.JobType),
			valueOrEmpty(job.CurrentState.Status),
			valueOrEmpty(job.CurrentState.DetailedStatus),
			valueOrEmpty(job.CurrentState.LatestRunStatus),
			valueOrEmpty(job.CurrentState.LatestRunTrigger),
			valueOrEmpty(job.CurrentState.RunCommand),
			valueOrEmpty(job.CurrentState.ScheduleExpression),
			valueOrEmpty(job.CurrentState.SchedulerLogsURL),
			job.CurrentState.ParentAppName,
			valueOrEmpty(job.CurrentState.ParentHostname),
			valueOrEmpty(job.CurrentState.ParentRuntime),
			valueOrEmpty(job.CurrentState.ParentPublicNetworkAccess),
			valueOrEmpty(job.CurrentState.ParentIdentityType),
			intPtrString(job.CurrentState.ParentAppSettingsCount),
			intPtrString(job.CurrentState.ParentKeyVaultReferenceCount),
			intPtrString(job.CurrentState.ParentConnectionStringCount),
			jsonStringSlice(persistenceRoleContextCSV(job.CurrentState.StrongestVisibleExecutionContext)),
			jsonStringSlice(job.CurrentState.NearbyThematicNames),
			jsonStringSlice(job.StillUnmapped),
			job.Summary,
			jsonStringSlice(job.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"webjob",
		"resource_group",
		"location",
		"capability_steps",
		"current_identity_context",
		"execution_context_options",
		"mode",
		"job_type",
		"status",
		"detailed_status",
		"latest_run_status",
		"latest_run_trigger",
		"run_command",
		"schedule_expression",
		"scheduler_logs_url",
		"parent_app_name",
		"parent_hostname",
		"parent_runtime",
		"parent_public_network_access",
		"parent_identity_type",
		"parent_app_settings_count",
		"parent_key_vault_reference_count",
		"parent_connection_string_count",
		"strongest_visible_execution_context",
		"nearby_thematic_names",
		"still_unmapped",
		"summary",
		"related_ids",
	}, rows)
}

func persistenceContainerAppsJobsCSV(payload models.PersistenceContainerAppsJobsOutput) (string, error) {
	return encodeCSVColumns(persistenceContainerAppsJobsCSVColumns(), payload.ContainerAppsJobs)
}

func persistenceContainerAppsJobsCSVColumns() []csvColumn[models.PersistenceContainerAppsJob] {
	return []csvColumn[models.PersistenceContainerAppsJob]{
		{"id", func(job models.PersistenceContainerAppsJob) string { return job.ID }},
		{"container_apps_job", func(job models.PersistenceContainerAppsJob) string { return job.Name }},
		{"resource_group", func(job models.PersistenceContainerAppsJob) string { return job.ResourceGroup }},
		{"location", func(job models.PersistenceContainerAppsJob) string { return job.Location }},
		{"capability_steps", func(job models.PersistenceContainerAppsJob) string {
			return jsonStringSlice(persistenceCapabilityStepsCSV(job.CapabilitySteps))
		}},
		{"current_identity_context", func(job models.PersistenceContainerAppsJob) string {
			return jsonStringSlice(persistenceRoleContextCSV(job.CurrentIdentityContext))
		}},
		{"execution_context_options", func(job models.PersistenceContainerAppsJob) string {
			return jsonStringSlice(job.ExecutionContextOptions)
		}},
		{"environment_id", func(job models.PersistenceContainerAppsJob) string {
			return valueOrEmpty(job.CurrentState.EnvironmentID)
		}},
		{"trigger_type", func(job models.PersistenceContainerAppsJob) string { return valueOrEmpty(job.CurrentState.TriggerType) }},
		{"schedule_expression", func(job models.PersistenceContainerAppsJob) string {
			return valueOrEmpty(job.CurrentState.ScheduleExpression)
		}},
		{"event_rules", func(job models.PersistenceContainerAppsJob) string {
			return containerAppsJobEventRulesCSV(job.CurrentState.EventRules)
		}},
		{"container_images", func(job models.PersistenceContainerAppsJob) string {
			return jsonStringSlice(job.CurrentState.ContainerImages)
		}},
		{"command", func(job models.PersistenceContainerAppsJob) string { return jsonStringSlice(job.CurrentState.Command) }},
		{"parallelism", func(job models.PersistenceContainerAppsJob) string { return intPtrString(job.CurrentState.Parallelism) }},
		{"replica_completion_count", func(job models.PersistenceContainerAppsJob) string {
			return intPtrString(job.CurrentState.ReplicaCompletionCount)
		}},
		{"replica_retry_limit", func(job models.PersistenceContainerAppsJob) string {
			return intPtrString(job.CurrentState.ReplicaRetryLimit)
		}},
		{"replica_timeout", func(job models.PersistenceContainerAppsJob) string {
			return intPtrString(job.CurrentState.ReplicaTimeout)
		}},
		{"identity_type", func(job models.PersistenceContainerAppsJob) string {
			return valueOrEmpty(job.CurrentState.IdentityType)
		}},
		{"workload_principal_id", func(job models.PersistenceContainerAppsJob) string {
			return valueOrEmpty(job.CurrentState.WorkloadPrincipalID)
		}},
		{"workload_client_id", func(job models.PersistenceContainerAppsJob) string {
			return valueOrEmpty(job.CurrentState.WorkloadClientID)
		}},
		{"workload_identity_ids", func(job models.PersistenceContainerAppsJob) string {
			return jsonStringSlice(job.CurrentState.WorkloadIdentityIDs)
		}},
		{"secret_count", func(job models.PersistenceContainerAppsJob) string { return intPtrString(job.CurrentState.SecretCount) }},
		{"key_vault_secret_count", func(job models.PersistenceContainerAppsJob) string {
			return intPtrString(job.CurrentState.KeyVaultSecretCount)
		}},
		{"registry_servers", func(job models.PersistenceContainerAppsJob) string {
			return jsonStringSlice(job.CurrentState.RegistryServers)
		}},
		{"registry_identity_count", func(job models.PersistenceContainerAppsJob) string {
			return intPtrString(job.CurrentState.RegistryIdentityCount)
		}},
		{"registry_password_ref_count", func(job models.PersistenceContainerAppsJob) string {
			return intPtrString(job.CurrentState.RegistryPasswordRefCount)
		}},
		{"strongest_visible_execution_context", func(job models.PersistenceContainerAppsJob) string {
			return jsonStringSlice(persistenceRoleContextCSV(job.CurrentState.StrongestVisibleExecutionContext))
		}},
		{"nearby_thematic_names", func(job models.PersistenceContainerAppsJob) string {
			return jsonStringSlice(job.CurrentState.NearbyThematicNames)
		}},
		{"still_unmapped", func(job models.PersistenceContainerAppsJob) string { return jsonStringSlice(job.StillUnmapped) }},
		{"summary", func(job models.PersistenceContainerAppsJob) string { return job.Summary }},
		{"related_ids", func(job models.PersistenceContainerAppsJob) string { return jsonStringSlice(job.RelatedIDs) }},
	}
}

func persistenceVMExtensionsCSV(payload models.PersistenceVMExtensionsOutput) (string, error) {
	return encodeCSVColumns(persistenceVMExtensionsCSVColumns(), payload.VMExtensions)
}

func persistenceVMExtensionsCSVColumns() []csvColumn[models.PersistenceVMExtension] {
	return []csvColumn[models.PersistenceVMExtension]{
		{"id", func(extension models.PersistenceVMExtension) string { return extension.ID }},
		{"vm_extension", func(extension models.PersistenceVMExtension) string { return extension.Name }},
		{"resource_group", func(extension models.PersistenceVMExtension) string { return extension.ResourceGroup }},
		{"location", func(extension models.PersistenceVMExtension) string { return extension.Location }},
		{"capability_steps", func(extension models.PersistenceVMExtension) string {
			return jsonStringSlice(persistenceCapabilityStepsCSV(extension.CapabilitySteps))
		}},
		{"current_identity_context", func(extension models.PersistenceVMExtension) string {
			return jsonStringSlice(persistenceRoleContextCSV(extension.CurrentIdentityContext))
		}},
		{"execution_context_options", func(extension models.PersistenceVMExtension) string {
			return jsonStringSlice(extension.ExecutionContextOptions)
		}},
		{"target_kind", func(extension models.PersistenceVMExtension) string { return extension.CurrentState.TargetKind }},
		{"target_name", func(extension models.PersistenceVMExtension) string { return extension.CurrentState.TargetName }},
		{"target_id", func(extension models.PersistenceVMExtension) string { return extension.CurrentState.TargetID }},
		{"publisher", func(extension models.PersistenceVMExtension) string {
			return valueOrEmpty(extension.CurrentState.Publisher)
		}},
		{"extension_type", func(extension models.PersistenceVMExtension) string {
			return valueOrEmpty(extension.CurrentState.ExtensionType)
		}},
		{"type_handler_version", func(extension models.PersistenceVMExtension) string {
			return valueOrEmpty(extension.CurrentState.TypeHandlerVersion)
		}},
		{"auto_upgrade_minor_version", func(extension models.PersistenceVMExtension) string {
			return boolPtrString(extension.CurrentState.AutoUpgradeMinorVersion)
		}},
		{"enable_automatic_upgrade", func(extension models.PersistenceVMExtension) string {
			return boolPtrString(extension.CurrentState.EnableAutomaticUpgrade)
		}},
		{"file_uri_hosts", func(extension models.PersistenceVMExtension) string {
			return jsonStringSlice(extension.CurrentState.FileURIHosts)
		}},
		{"file_uri_count", func(extension models.PersistenceVMExtension) string {
			return intPtrString(extension.CurrentState.FileURICount)
		}},
		{"command_clue", func(extension models.PersistenceVMExtension) string {
			return valueOrEmpty(extension.CurrentState.CommandClue)
		}},
		{"public_setting_keys", func(extension models.PersistenceVMExtension) string {
			return jsonStringSlice(extension.CurrentState.PublicSettingKeys)
		}},
		{"protected_settings_present", func(extension models.PersistenceVMExtension) string {
			return boolPtrString(extension.CurrentState.ProtectedSettingsPresent)
		}},
		{"key_vault_protected_settings", func(extension models.PersistenceVMExtension) string {
			return boolPtrString(extension.CurrentState.KeyVaultProtectedSettings)
		}},
		{"suppress_failures", func(extension models.PersistenceVMExtension) string {
			return boolPtrString(extension.CurrentState.SuppressFailures)
		}},
		{"force_update_tag", func(extension models.PersistenceVMExtension) string {
			return valueOrEmpty(extension.CurrentState.ForceUpdateTag)
		}},
		{"rerun_clues", func(extension models.PersistenceVMExtension) string {
			return jsonStringSlice(extension.CurrentState.RerunClues)
		}},
		{"provision_after_extensions", func(extension models.PersistenceVMExtension) string {
			return jsonStringSlice(extension.CurrentState.ProvisionAfterExtensions)
		}},
		{"provisioning_state", func(extension models.PersistenceVMExtension) string {
			return valueOrEmpty(extension.CurrentState.ProvisioningState)
		}},
		{"instance_view_statuses", func(extension models.PersistenceVMExtension) string {
			return jsonStringSlice(extension.CurrentState.InstanceViewStatuses)
		}},
		{"target_identity_ids", func(extension models.PersistenceVMExtension) string {
			return jsonStringSlice(extension.CurrentState.TargetIdentityIDs)
		}},
		{"strongest_visible_execution_context", func(extension models.PersistenceVMExtension) string {
			return jsonStringSlice(persistenceRoleContextCSV(extension.CurrentState.StrongestVisibleExecutionContext))
		}},
		{"vmss_orchestration_mode", func(extension models.PersistenceVMExtension) string {
			return valueOrEmpty(extension.CurrentState.VMSSOrchestrationMode)
		}},
		{"vmss_upgrade_mode", func(extension models.PersistenceVMExtension) string {
			return valueOrEmpty(extension.CurrentState.VMSSUpgradeMode)
		}},
		{"nearby_thematic_names", func(extension models.PersistenceVMExtension) string {
			return jsonStringSlice(extension.CurrentState.NearbyThematicNames)
		}},
		{"still_unmapped", func(extension models.PersistenceVMExtension) string { return jsonStringSlice(extension.StillUnmapped) }},
		{"summary", func(extension models.PersistenceVMExtension) string { return extension.Summary }},
		{"related_ids", func(extension models.PersistenceVMExtension) string { return jsonStringSlice(extension.RelatedIDs) }},
	}
}

func persistenceAzureMLCSV(payload models.PersistenceAzureMLOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Workspaces))
	for _, workspace := range payload.Workspaces {
		rows = append(rows, []string{
			workspace.ID,
			workspace.Name,
			workspace.ResourceGroup,
			valueOrEmpty(workspace.Location),
			persistenceCSVStepStatus(workspace.CapabilitySteps, "create or modify workspace"),
			persistenceCSVStepStatus(workspace.CapabilitySteps, "attach or reuse compute"),
			persistenceCSVStepStatus(workspace.CapabilitySteps, "add or modify jobs or pipelines"),
			persistenceCSVStepStatus(workspace.CapabilitySteps, "create or modify schedule"),
			persistenceCSVStepStatus(workspace.CapabilitySteps, "attach or reuse exec ctx"),
			persistenceCSVStepStatus(workspace.CapabilitySteps, "expose or reuse endpoint"),
			persistenceCSVRoleSummary(workspace.CurrentIdentityContext),
			jsonStringSlice(workspace.ExecutionContextOptions),
			workspace.CurrentState.Classification,
			valueOrEmpty(workspace.CurrentState.State),
			valueOrEmpty(workspace.CurrentState.PublicNetworkAccess),
			valueOrEmpty(workspace.CurrentState.IdentityType),
			intPtrString(workspace.CurrentState.ComputeCount),
			jsonStringSlice(workspace.CurrentState.ComputeTypes),
			intPtrString(workspace.CurrentState.JobCount),
			jsonStringSlice(workspace.CurrentState.JobTypes),
			intPtrString(workspace.CurrentState.ScheduleCount),
			jsonStringSlice(workspace.CurrentState.ScheduleTriggerTypes),
			intPtrString(workspace.CurrentState.EndpointCount),
			jsonStringSlice(workspace.CurrentState.EndpointAuthModes),
			jsonStringSlice(workspace.CurrentState.EndpointPublicAccess),
			intPtrString(workspace.CurrentState.DatastoreCount),
			jsonStringSlice(workspace.CurrentState.DatastoreTypes),
			persistenceCSVRoleSummary(workspace.CurrentState.StrongestVisibleExecutionContext),
			jsonStringSlice(workspace.CurrentState.NearbyThematicNames),
			jsonStringSlice(workspace.StillUnmapped),
			workspace.Summary,
			jsonStringSlice(workspace.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"workspace",
		"resource_group",
		"location",
		"create_or_modify_workspace",
		"attach_or_reuse_compute",
		"add_or_modify_jobs_or_pipelines",
		"create_or_modify_schedule",
		"attach_or_reuse_exec_ctx",
		"expose_or_reuse_endpoint",
		"current_identity_context",
		"execution_context_options",
		"classification",
		"state",
		"public_network_access",
		"identity_type",
		"compute_count",
		"compute_types",
		"job_count",
		"job_types",
		"schedule_count",
		"schedule_trigger_types",
		"endpoint_count",
		"endpoint_auth_modes",
		"endpoint_public_access",
		"datastore_count",
		"datastore_types",
		"strongest_visible_execution_context",
		"nearby_thematic_names",
		"still_unmapped",
		"summary",
		"related_ids",
	}, rows)
}

func persistenceLogicAppsCSV(payload models.PersistenceLogicAppsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Workflows))
	for _, workflow := range payload.Workflows {
		rows = append(rows, []string{
			workflow.ID,
			workflow.Name,
			workflow.ResourceGroup,
			valueOrEmpty(workflow.Location),
			persistenceCSVStepStatus(workflow.CapabilitySteps, "create or modify workflow"),
			persistenceCSVStepStatus(workflow.CapabilitySteps, "edit workflow definition"),
			persistenceCSVStepStatus(workflow.CapabilitySteps, "attach or reuse exec ctx"),
			persistenceCSVStepStatus(workflow.CapabilitySteps, "define or modify trigger"),
			persistenceCSVStepStatus(workflow.CapabilitySteps, "enable workflow"),
			persistenceCSVStepStatus(workflow.CapabilitySteps, "add or repurpose downstream actions"),
			persistenceCSVRoleSummary(workflow.CurrentIdentityContext),
			jsonStringSlice(workflow.ExecutionContextOptions),
			workflow.CurrentState.Classification,
			valueOrEmpty(workflow.CurrentState.Platform),
			valueOrEmpty(workflow.CurrentState.WorkflowKind),
			valueOrEmpty(workflow.CurrentState.State),
			jsonStringSlice(workflow.CurrentState.TriggerTypes),
			boolString(workflow.CurrentState.ExternallyCallableRequestTrigger),
			valueOrEmpty(workflow.CurrentState.RecurrenceSummary),
			valueOrEmpty(workflow.CurrentState.IdentityType),
			persistenceCSVRoleSummary(workflow.CurrentState.StrongestVisibleExecutionContext),
			jsonStringSlice(workflow.CurrentState.DownstreamActionKinds),
			jsonStringSlice(workflow.StillUnmapped),
			workflow.Summary,
			jsonStringSlice(workflow.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"logic_app",
		"resource_group",
		"location",
		"create_or_modify_workflow",
		"edit_workflow_definition",
		"attach_or_reuse_exec_ctx",
		"define_or_modify_trigger",
		"enable_workflow",
		"add_or_repurpose_downstream_actions",
		"current_identity_context",
		"execution_context_options",
		"classification",
		"platform",
		"workflow_kind",
		"state",
		"trigger_types",
		"externally_callable_request_trigger",
		"recurrence_summary",
		"identity_type",
		"strongest_visible_execution_context",
		"downstream_action_kinds",
		"still_unmapped",
		"summary",
		"related_ids",
	}, rows)
}

func persistenceFunctionsCSV(payload models.PersistenceFunctionsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.FunctionApps))
	for _, app := range payload.FunctionApps {
		rows = append(rows, []string{
			app.ID,
			app.Name,
			app.ResourceGroup,
			app.Location,
			jsonStringSlice(persistenceCapabilityStepsCSV(app.CapabilitySteps)),
			jsonStringSlice(persistenceRoleContextCSV(app.CurrentIdentityContext)),
			jsonStringSlice(app.ExecutionContextOptions),
			valueOrEmpty(app.CurrentState.State),
			valueOrEmpty(app.CurrentState.Hostname),
			valueOrEmpty(app.CurrentState.PublicNetworkAccess),
			valueOrEmpty(app.CurrentState.Runtime),
			valueOrEmpty(app.CurrentState.Deployment),
			valueOrEmpty(app.CurrentState.IdentityType),
			boolPtrString(app.CurrentState.AlwaysOn),
			valueOrEmpty(app.CurrentState.AzureWebJobsStorageValueType),
			intPtrString(app.CurrentState.KeyVaultReferenceCount),
			boolPtrString(app.CurrentState.RunFromPackage),
			jsonStringSlice(app.CurrentState.TriggerTypes),
			jsonStringSlice(functionNamesCSV(app.CurrentState.VisibleFunctions)),
			jsonStringSlice(persistenceRoleContextCSV(app.CurrentState.StrongestVisibleExecutionContext)),
			jsonStringSlice(app.CurrentState.NearbyThematicNames),
			jsonStringSlice(app.StillUnmapped),
			app.Summary,
			jsonStringSlice(app.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"function_app",
		"resource_group",
		"location",
		"capability_steps",
		"current_identity_context",
		"execution_context_options",
		"state",
		"hostname",
		"public_network_access",
		"runtime",
		"deployment",
		"identity_type",
		"always_on",
		"azure_webjobs_storage_value_type",
		"key_vault_reference_count",
		"run_from_package",
		"trigger_types",
		"visible_function_names",
		"strongest_visible_execution_context",
		"nearby_thematic_names",
		"still_unmapped",
		"summary",
		"related_ids",
	}, rows)
}

func persistenceCapabilityStepsCSV(steps []models.PersistenceCapabilityStep) []string {
	rows := make([]string, 0, len(steps))
	for _, step := range steps {
		rows = append(rows, step.Action+" | "+step.APISurface+" | "+step.Status)
	}
	return rows
}

func persistenceRoleContextCSV(context *models.PersistenceRoleContext) []string {
	if context == nil {
		return nil
	}
	rows := []string{
		"name=" + context.Name,
		"kind=" + context.Kind,
		"principal_id=" + valueOrEmpty(context.PrincipalID),
		"identity_type=" + valueOrEmpty(context.IdentityType),
		"role_names=" + strings.Join(context.RoleNames, ";"),
		"scope_ids=" + strings.Join(context.ScopeIDs, ";"),
		"summary=" + context.Summary,
	}
	return rows
}

func devopsCSV(payload models.DevopsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Pipelines))
	for _, pipeline := range payload.Pipelines {
		rows = append(rows, []string{
			pipeline.ID,
			pipeline.DefinitionID,
			pipeline.Name,
			pipeline.ProjectID,
			pipeline.ProjectName,
			pipeline.Path,
			valueOrEmpty(pipeline.RepositoryID),
			pipeline.RepositoryName,
			pipeline.RepositoryType,
			pipeline.RepositoryURL,
			pipeline.RepositoryHostType,
			pipeline.SourceVisibilityState,
			pipeline.DefaultBranch,
			jsonStringSlice(pipeline.TriggerTypes),
			jsonStringSlice(pipeline.VariableGroupNames),
			intString(pipeline.SecretVariableCount),
			jsonStringSlice(pipeline.SecretVariableNames),
			jsonStringSlice(pipeline.KeyVaultGroupNames),
			jsonStringSlice(pipeline.KeyVaultNames),
			jsonStringSlice(pipeline.AzureServiceConnectionNames),
			jsonStringSlice(pipeline.AzureServiceConnectionTypes),
			jsonStringSlice(pipeline.AzureServiceConnectionAuthSchemes),
			jsonStringSlice(pipeline.AzureServiceConnectionIDs),
			jsonStringSlice(pipeline.AzureServiceConnectionPrincipalIDs),
			jsonStringSlice(pipeline.AzureServiceConnectionClientIDs),
			jsonStringSlice(pipeline.AzureServiceConnectionTenantIDs),
			jsonStringSlice(pipeline.AzureServiceConnectionSubscriptionIDs),
			jsonStringSlice(pipeline.TargetClues),
			jsonStringSlice(pipeline.RiskCues),
			jsonStringSlice(pipeline.ExecutionModes),
			jsonStringSlice(pipeline.UpstreamSources),
			jsonValue(pipeline.TrustedInputs),
			jsonStringSlice(pipeline.TrustedInputTypes),
			jsonStringSlice(pipeline.TrustedInputRefs),
			jsonStringSlice(pipeline.TrustedInputJoinIDs),
			pipeline.PrimaryInjectionSurface,
			pipeline.PrimaryTrustedInputRef,
			jsonStringSlice(pipeline.SourceJoinIDs),
			jsonStringSlice(pipeline.TriggerJoinIDs),
			jsonStringSlice(pipeline.IdentityJoinIDs),
			jsonStringSlice(pipeline.SecretSupportTypes),
			jsonStringSlice(pipeline.SecretDependencyIDs),
			jsonStringSlice(pipeline.InjectionSurfaceTypes),
			jsonStringSlice(pipeline.CurrentOperatorInjectionSurfaceTypes),
			pipeline.EditPathState,
			pipeline.QueuePathState,
			pipeline.RerunPathState,
			pipeline.ApprovalPathState,
			boolPtrString(pipeline.CurrentOperatorCanViewDefinition),
			boolPtrString(pipeline.CurrentOperatorCanQueue),
			boolPtrString(pipeline.CurrentOperatorCanEdit),
			boolPtrString(pipeline.CurrentOperatorCanViewSource),
			boolPtrString(pipeline.CurrentOperatorCanContributeSource),
			jsonStringSlice(pipeline.ConsequenceTypes),
			boolString(pipeline.MissingExecutionPath),
			boolString(pipeline.MissingInjectionPoint),
			boolString(pipeline.MissingTargetMapping),
			boolString(pipeline.PartialRead),
			pipeline.Summary,
			jsonStringSlice(pipeline.RelatedIDs),
		})
	}

	return encodeCSV([]string{
		"id",
		"definition_id",
		"name",
		"project_id",
		"project_name",
		"path",
		"repository_id",
		"repository_name",
		"repository_type",
		"repository_url",
		"repository_host_type",
		"source_visibility_state",
		"default_branch",
		"trigger_types",
		"variable_group_names",
		"secret_variable_count",
		"secret_variable_names",
		"key_vault_group_names",
		"key_vault_names",
		"azure_service_connection_names",
		"azure_service_connection_types",
		"azure_service_connection_auth_schemes",
		"azure_service_connection_ids",
		"azure_service_connection_principal_ids",
		"azure_service_connection_client_ids",
		"azure_service_connection_tenant_ids",
		"azure_service_connection_subscription_ids",
		"target_clues",
		"risk_cues",
		"execution_modes",
		"upstream_sources",
		"trusted_inputs",
		"trusted_input_types",
		"trusted_input_refs",
		"trusted_input_join_ids",
		"primary_injection_surface",
		"primary_trusted_input_ref",
		"source_join_ids",
		"trigger_join_ids",
		"identity_join_ids",
		"secret_support_types",
		"secret_dependency_ids",
		"injection_surface_types",
		"current_operator_injection_surface_types",
		"edit_path_state",
		"queue_path_state",
		"rerun_path_state",
		"approval_path_state",
		"current_operator_can_view_definition",
		"current_operator_can_queue",
		"current_operator_can_edit",
		"current_operator_can_view_source",
		"current_operator_can_contribute_source",
		"consequence_types",
		"missing_execution_path",
		"missing_injection_point",
		"missing_target_mapping",
		"partial_read",
		"summary",
		"related_ids",
	}, rows)
}

func chainsOverviewCSV(payload models.ChainsOverviewOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Families))
	for _, family := range payload.Families {
		rows = append(rows, []string{
			family.Family,
			family.State,
			family.Summary,
			strings.Join(family.BestCurrentExamples, ", "),
			chainsBackingCommands(family.SourceCommands),
			family.AllowedClaim,
			family.CurrentGap,
		})
	}
	return encodeCSV(
		[]string{"family", "state", "summary", "examples", "backing_commands", "allowed_claim", "current_gap"},
		rows,
	)
}

func chainsFamilyCSV(payload models.ChainsOutput) (string, error) {
	if renderer, ok := chainsFamilyCSVRenderers[payload.Family]; ok {
		return renderer(payload)
	}
	return chainsCredentialPathCSV(payload)
}

func chainsComputeControlCSV(payload models.ChainsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Paths))
	for _, path := range payload.Paths {
		rows = append(rows, []string{
			path.AssetID,
			path.AssetKind,
			path.AssetName,
			path.ChainID,
			path.ClueType,
			valueOrEmpty(path.ConfidenceBoundary),
			valueOrEmpty(path.ConfirmationBasis),
			jsonStringSlice(path.EvidenceCommands),
			computeControlIdentityLabel(path.TargetNames),
			valueOrEmpty(path.InsertionPoint),
			jsonStringSlice(path.JoinedSurfaceTypes),
			firstNonEmptyString(path.LikelyImpact, path.StrongerOutcome),
			valueOrEmpty(path.Location),
			path.MissingConfirmation,
			path.NextReview,
			firstNonEmptyString(path.Note, path.WhyCare),
			valueOrEmpty(path.PathConcept),
			path.Priority,
			computeControlProofStatusLabel(path.TargetResolution),
			computeControlReachFromHereLabel(valueOrEmpty(path.InsertionPoint)),
			jsonStringSlice(path.RelatedIDs),
			valueOrEmpty(path.SourceCommand),
			valueOrEmpty(path.SourceContext),
			firstNonEmptyString(path.StrongerOutcome, path.LikelyImpact),
			path.Summary,
			intString(path.TargetCount),
			jsonStringSlice(path.TargetIDs),
			jsonStringSlice(path.TargetNames),
			path.TargetResolution,
			path.TargetService,
			computeControlTokenPathLabel(valueOrEmpty(path.InsertionPoint)),
			valueOrEmpty(path.Urgency),
			path.VisiblePath,
			computeControlWhenLabel(valueOrEmpty(path.Urgency)),
		})
	}
	return encodeCSV([]string{
		"asset_id",
		"asset_kind",
		"asset_name",
		"chain_id",
		"clue_type",
		"confidence_boundary",
		"confirmation_basis",
		"evidence_commands",
		"identity",
		"insertion_point",
		"joined_surface_types",
		"likely_impact",
		"location",
		"missing_confirmation",
		"next_review",
		"note",
		"path_concept",
		"priority",
		"proof_status",
		"reach_from_here",
		"related_ids",
		"source_command",
		"source_context",
		"stronger_outcome",
		"summary",
		"target_count",
		"target_ids",
		"target_names",
		"target_resolution",
		"target_service",
		"token_path",
		"urgency",
		"visible_path",
		"when",
	}, rows)
}

func chainsEscalationPathCSV(payload models.ChainsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Paths))
	for _, path := range payload.Paths {
		rows = append(rows, []string{
			path.AssetID,
			path.AssetKind,
			path.AssetName,
			valueOrEmpty(path.StartingFoothold),
			path.ChainID,
			path.ClueType,
			valueOrEmpty(path.ConfidenceBoundary),
			valueOrEmpty(path.ConfirmationBasis),
			jsonStringSlice(path.EvidenceCommands),
			jsonStringSlice(path.JoinedSurfaceTypes),
			firstNonEmptyString(path.LikelyImpact, path.StrongerOutcome),
			firstNonEmptyString(path.StrongerOutcome, path.LikelyImpact),
			path.MissingConfirmation,
			path.NextReview,
			firstNonEmptyString(path.Note, path.WhyCare),
			valueOrEmpty(path.PathConcept),
			valueOrEmpty(path.PathType),
			path.Priority,
			jsonStringSlice(path.RelatedIDs),
			valueOrEmpty(path.SourceCommand),
			valueOrEmpty(path.SourceContext),
			path.Summary,
			intString(path.TargetCount),
			jsonStringSlice(path.TargetIDs),
			jsonStringSlice(path.TargetNames),
			path.TargetResolution,
			path.TargetService,
			valueOrEmpty(path.Urgency),
			path.VisiblePath,
		})
	}
	return encodeCSV([]string{
		"asset_id",
		"asset_kind",
		"asset_name",
		"starting_foothold",
		"chain_id",
		"clue_type",
		"confidence_boundary",
		"confirmation_basis",
		"evidence_commands",
		"joined_surface_types",
		"likely_impact",
		"stronger_outcome",
		"missing_confirmation",
		"next_review",
		"note",
		"path_concept",
		"path_type",
		"priority",
		"related_ids",
		"source_command",
		"source_context",
		"summary",
		"target_count",
		"target_ids",
		"target_names",
		"target_resolution",
		"target_service",
		"urgency",
		"visible_path",
	}, rows)
}

func chainsDeploymentPathCSV(payload models.ChainsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Paths))
	for _, path := range payload.Paths {
		rows = append(rows, []string{
			valueOrEmpty(path.Actionability),
			valueOrEmpty(path.ActionabilityState),
			path.AssetID,
			path.AssetKind,
			path.AssetName,
			path.ChainID,
			path.ClueType,
			valueOrEmpty(path.ConfidenceBoundary),
			valueOrEmpty(path.ConfirmationBasis),
			jsonStringSlice(path.EvidenceCommands),
			valueOrEmpty(path.InsertionPoint),
			valueOrEmpty(path.InsertionPointLabel),
			jsonStringSlice(path.JoinedSurfaceTypes),
			firstNonEmptyString(path.LikelyAzureImpact, path.LikelyImpact),
			firstNonEmptyString(path.LikelyImpact, path.LikelyAzureImpact),
			valueOrEmpty(path.Location),
			path.MissingConfirmation,
			path.NextReview,
			firstNonEmptyString(path.Note, path.WhyCare),
			valueOrEmpty(path.PathConcept),
			valueOrEmpty(path.PrimarySurface),
			valueOrEmpty(path.PrimaryInputRef),
			path.Priority,
			jsonStringSlice(path.RelatedIDs),
			valueOrEmpty(path.SettingName),
			valueOrEmpty(path.Source),
			valueOrEmpty(path.SourceCommand),
			valueOrEmpty(path.SourceContext),
			valueOrEmpty(path.StrongerOutcome),
			path.Summary,
			intString(path.TargetCount),
			jsonStringSlice(path.TargetIDs),
			jsonStringSlice(path.TargetNames),
			path.TargetResolution,
			path.TargetService,
			valueOrEmpty(path.TargetVisibility),
			valueOrEmpty(path.Urgency),
			path.VisiblePath,
			firstNonEmptyString(path.WhatsMissing, path.ConfidenceBoundary),
			firstNonEmptyString(path.WhyCare, path.Note),
		})
	}
	return encodeCSV([]string{
		"actionability",
		"actionability_state",
		"asset_id",
		"asset_kind",
		"asset_name",
		"chain_id",
		"clue_type",
		"confidence_boundary",
		"confirmation_basis",
		"evidence_commands",
		"insertion_point",
		"insertion_point_display",
		"joined_surface_types",
		"likely_azure_impact",
		"likely_impact",
		"location",
		"missing_confirmation",
		"next_review",
		"note",
		"path_concept",
		"primary_injection_surface",
		"primary_trusted_input_ref",
		"priority",
		"related_ids",
		"setting_name",
		"source",
		"source_command",
		"source_context",
		"stronger_outcome",
		"summary",
		"target_count",
		"target_ids",
		"target_names",
		"target_resolution",
		"target_service",
		"target_visibility_issue",
		"urgency",
		"visible_path",
		"whats_missing",
		"why_care",
	}, rows)
}

func chainsCredentialPathCSV(payload models.ChainsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Paths))
	for _, path := range payload.Paths {
		rows = append(rows, []string{
			path.ChainID,
			path.AssetID,
			path.AssetKind,
			path.AssetName,
			valueOrEmpty(path.Location),
			valueOrEmpty(path.SettingName),
			path.ClueType,
			path.Priority,
			valueOrEmpty(path.Urgency),
			path.VisiblePath,
			path.TargetService,
			path.TargetResolution,
			jsonStringSlice(path.EvidenceCommands),
			jsonStringSlice(path.JoinedSurfaceTypes),
			intString(path.TargetCount),
			jsonStringSlice(path.TargetIDs),
			jsonStringSlice(path.TargetNames),
			valueOrEmpty(path.TargetVisibility),
			path.NextReview,
			valueOrEmpty(path.ConfidenceBoundary),
			path.Summary,
			path.MissingConfirmation,
			jsonStringSlice(path.RelatedIDs),
		})
	}
	return encodeCSV([]string{
		"chain_id",
		"asset_id",
		"asset_kind",
		"asset_name",
		"location",
		"setting_name",
		"clue_type",
		"priority",
		"urgency",
		"visible_path",
		"target_service",
		"target_resolution",
		"evidence_commands",
		"joined_surface_types",
		"target_count",
		"target_ids",
		"target_names",
		"target_visibility_issue",
		"next_review",
		"confidence_boundary",
		"summary",
		"missing_confirmation",
		"related_ids",
	}, rows)
}

func firstNonEmptyString(values ...*string) string {
	for _, value := range values {
		if value != nil && strings.TrimSpace(*value) != "" {
			return *value
		}
	}
	return ""
}

func appServicesCSV(payload models.AppServicesOutput) (string, error) {
	rows := make([][]string, 0, len(payload.AppServices))
	for _, app := range payload.AppServices {
		rows = append(rows, []string{
			intPtrString(app.AppSettingsCount),
			valueOrEmpty(app.AppServicePlanID),
			fmt.Sprintf("%t", app.ClientCertEnabled),
			intPtrString(app.ConnectionStringCount),
			jsonStringSlice(app.ConnectionStringTypes),
			valueOrEmpty(app.DefaultHostname),
			valueOrEmpty(app.Deployment),
			valueOrEmpty(app.DeploymentBranch),
			boolPtrString(app.DeploymentIsGitHubAction),
			boolPtrString(app.DeploymentManualIntegration),
			valueOrEmpty(app.DeploymentRepoURL),
			valueOrEmpty(app.FTPSState),
			fmt.Sprintf("%t", app.HTTPSOnly),
			app.ID,
			intPtrString(app.KeyVaultConnectionStringCount),
			intPtrString(app.KeyVaultReferenceCount),
			app.Location,
			valueOrEmpty(app.MinTLSVersion),
			app.Name,
			valueOrEmpty(app.PublicNetworkAccess),
			join(app.RelatedIDs, ";"),
			app.ResourceGroup,
			boolPtrString(app.RunFromPackage),
			valueOrEmpty(app.RuntimeStack),
			intPtrString(app.SensitiveSettingCount),
			valueOrEmpty(app.State),
			app.Summary,
			valueOrEmpty(app.WorkloadClientID),
			join(app.WorkloadIdentityIDs, ";"),
			valueOrEmpty(app.WorkloadIdentityType),
			valueOrEmpty(app.WorkloadPrincipalID),
		})
	}
	return encodeCSV([]string{
		"app_settings_count",
		"app_service_plan_id",
		"client_cert_enabled",
		"connection_string_count",
		"connection_string_types",
		"default_hostname",
		"deployment",
		"deployment_branch",
		"deployment_is_github_action",
		"deployment_manual_integration",
		"deployment_repo_url",
		"ftps_state",
		"https_only",
		"id",
		"key_vault_connection_string_count",
		"key_vault_reference_count",
		"location",
		"min_tls_version",
		"name",
		"public_network_access",
		"related_ids",
		"resource_group",
		"run_from_package",
		"runtime_stack",
		"sensitive_setting_count",
		"state",
		"summary",
		"workload_client_id",
		"workload_identity_ids",
		"workload_identity_type",
		"workload_principal_id",
	}, rows)
}

func functionsCSV(payload models.FunctionsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.FunctionApps))
	for _, app := range payload.FunctionApps {
		rows = append(rows, []string{
			boolPtrString(app.AlwaysOn),
			valueOrEmpty(app.AppServicePlanID),
			valueOrEmpty(app.AzureWebJobsStorageReferenceTarget),
			valueOrEmpty(app.AzureWebJobsStorageValueType),
			fmt.Sprintf("%t", app.ClientCertEnabled),
			valueOrEmpty(app.DefaultHostname),
			valueOrEmpty(app.Deployment),
			valueOrEmpty(app.FTPSState),
			valueOrEmpty(app.FunctionsExtensionVersion),
			fmt.Sprintf("%t", app.HTTPSOnly),
			app.ID,
			valueOrEmpty(app.Identity),
			intPtrString(app.KeyVaultReferenceCount),
			app.Location,
			valueOrEmpty(app.MinTLSVersion),
			app.Name,
			fmt.Sprintf("%d", len(app.VisibleFunctions)),
			valueOrEmpty(app.PublicNetworkAccess),
			join(app.RelatedIDs, ";"),
			app.ResourceGroup,
			boolPtrString(app.RunFromPackage),
			valueOrEmpty(app.Runtime),
			valueOrEmpty(app.RuntimeStack),
			valueOrEmpty(app.State),
			app.Summary,
			jsonStringSlice(app.TriggerTypes),
			jsonStringSlice(functionNamesCSV(app.VisibleFunctions)),
			jsonValue(app.VisibleFunctions),
			valueOrEmpty(app.WorkloadClientID),
			join(app.WorkloadIdentityIDs, ";"),
			valueOrEmpty(app.WorkloadIdentityType),
			valueOrEmpty(app.WorkloadPrincipalID),
		})
	}
	return encodeCSV([]string{
		"always_on",
		"app_service_plan_id",
		"azure_webjobs_storage_reference_target",
		"azure_webjobs_storage_value_type",
		"client_cert_enabled",
		"hostname",
		"deployment",
		"ftps_state",
		"functions_extension_version",
		"https_only",
		"id",
		"identity",
		"key_vault_reference_count",
		"location",
		"min_tls_version",
		"function_app",
		"visible_function_count",
		"public_network_access",
		"related_ids",
		"resource_group",
		"run_from_package",
		"runtime",
		"runtime_stack",
		"state",
		"summary",
		"trigger_types",
		"visible_function_names",
		"visible_functions",
		"workload_client_id",
		"workload_identity_ids",
		"workload_identity_type",
		"workload_principal_id",
	}, rows)
}

func webJobsCSV(payload models.WebJobsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.WebJobs))
	for _, job := range payload.WebJobs {
		rows = append(rows, []string{
			valueOrEmpty(job.DetailedStatus),
			job.ID,
			valueOrEmpty(job.JobType),
			valueOrEmpty(job.LatestRunStatus),
			valueOrEmpty(job.LatestRunTrigger),
			job.Location,
			job.Mode,
			job.Name,
			job.ParentAppID,
			job.ParentAppName,
			valueOrEmpty(job.ParentHostname),
			join(job.ParentIdentityIDs, ";"),
			valueOrEmpty(job.ParentIdentityType),
			join(job.RelatedIDs, ";"),
			job.ResourceGroup,
			valueOrEmpty(job.RunCommand),
			valueOrEmpty(job.ScheduleExpression),
			valueOrEmpty(job.SchedulerLogsURL),
			valueOrEmpty(job.Status),
			job.Summary,
		})
	}
	return encodeCSV([]string{
		"detailed_status",
		"id",
		"job_type",
		"latest_run_status",
		"latest_run_trigger",
		"location",
		"mode",
		"name",
		"parent_app_id",
		"parent_app_name",
		"parent_hostname",
		"parent_identity_ids",
		"parent_identity_type",
		"related_ids",
		"resource_group",
		"run_command",
		"schedule_expression",
		"scheduler_logs_url",
		"status",
		"summary",
	}, rows)
}

func functionNamesCSV(values []models.FunctionChildAsset) []string {
	names := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value.Name) == "" {
			continue
		}
		names = append(names, value.Name)
	}
	return names
}

func containerAppsCSV(payload models.ContainerAppsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.ContainerApps))
	for _, app := range payload.ContainerApps {
		rows = append(rows, []string{
			valueOrEmpty(app.DefaultHostname),
			valueOrEmpty(app.EnvironmentID),
			boolPtrString(app.ExternalIngressEnabled),
			app.ID,
			intPtrString(app.IngressTargetPort),
			valueOrEmpty(app.IngressTransport),
			valueOrEmpty(app.LatestReadyRevisionName),
			valueOrEmpty(app.LatestRevisionName),
			app.Location,
			app.Name,
			join(app.RelatedIDs, ";"),
			app.ResourceGroup,
			valueOrEmpty(app.RevisionMode),
			app.Summary,
			valueOrEmpty(app.WorkloadClientID),
			join(app.WorkloadIdentityIDs, ";"),
			valueOrEmpty(app.WorkloadIdentityType),
			valueOrEmpty(app.WorkloadPrincipalID),
		})
	}
	return encodeCSV([]string{
		"default_hostname",
		"environment_id",
		"external_ingress_enabled",
		"id",
		"ingress_target_port",
		"ingress_transport",
		"latest_ready_revision_name",
		"latest_revision_name",
		"location",
		"name",
		"related_ids",
		"resource_group",
		"revision_mode",
		"summary",
		"workload_client_id",
		"workload_identity_ids",
		"workload_identity_type",
		"workload_principal_id",
	}, rows)
}

func containerAppsJobsCSV(payload models.ContainerAppsJobsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.ContainerAppsJobs))
	for _, job := range payload.ContainerAppsJobs {
		rows = append(rows, []string{
			join(job.Command, ";"),
			join(job.ContainerImages, ";"),
			valueOrEmpty(job.EnvironmentID),
			containerAppsJobEventRulesCSV(job.EventRules),
			job.ID,
			intPtrString(job.KeyVaultSecretCount),
			job.Location,
			job.Name,
			intPtrString(job.Parallelism),
			intPtrString(job.RegistryIdentityCount),
			intPtrString(job.RegistryPasswordRefCount),
			join(job.RegistryServers, ";"),
			join(job.RelatedIDs, ";"),
			intPtrString(job.ReplicaCompletionCount),
			intPtrString(job.ReplicaRetryLimit),
			intPtrString(job.ReplicaTimeout),
			job.ResourceGroup,
			valueOrEmpty(job.ScheduleExpression),
			intPtrString(job.SecretCount),
			job.Summary,
			valueOrEmpty(job.TriggerType),
			valueOrEmpty(job.WorkloadClientID),
			join(job.WorkloadIdentityIDs, ";"),
			valueOrEmpty(job.WorkloadIdentityType),
			valueOrEmpty(job.WorkloadPrincipalID),
		})
	}
	return encodeCSV([]string{
		"command",
		"container_images",
		"environment_id",
		"event_rules",
		"id",
		"key_vault_secret_count",
		"location",
		"name",
		"parallelism",
		"registry_identity_count",
		"registry_password_ref_count",
		"registry_servers",
		"related_ids",
		"replica_completion_count",
		"replica_retry_limit",
		"replica_timeout",
		"resource_group",
		"schedule_expression",
		"secret_count",
		"summary",
		"trigger_type",
		"workload_client_id",
		"workload_identity_ids",
		"workload_identity_type",
		"workload_principal_id",
	}, rows)
}

func containerInstancesCSV(payload models.ContainerInstancesOutput) (string, error) {
	rows := make([][]string, 0, len(payload.ContainerInstances))
	for _, item := range payload.ContainerInstances {
		rows = append(rows, []string{
			intPtrString(item.ContainerCount),
			join(item.ContainerImages, ";"),
			intJoin(item.ExposedPorts, ";"),
			valueOrEmpty(item.FQDN),
			item.ID,
			item.Location,
			item.Name,
			valueOrEmpty(item.OSType),
			valueOrEmpty(item.ProvisioningState),
			valueOrEmpty(item.PublicIPAddress),
			join(item.RelatedIDs, ";"),
			item.ResourceGroup,
			valueOrEmpty(item.RestartPolicy),
			join(item.SubnetIDs, ";"),
			item.Summary,
			valueOrEmpty(item.WorkloadClientID),
			join(item.WorkloadIdentityIDs, ";"),
			valueOrEmpty(item.WorkloadIdentityType),
			valueOrEmpty(item.WorkloadPrincipalID),
		})
	}
	return encodeCSV([]string{
		"container_count",
		"container_images",
		"exposed_ports",
		"fqdn",
		"id",
		"location",
		"name",
		"os_type",
		"provisioning_state",
		"public_ip_address",
		"related_ids",
		"resource_group",
		"restart_policy",
		"subnet_ids",
		"summary",
		"workload_client_id",
		"workload_identity_ids",
		"workload_identity_type",
		"workload_principal_id",
	}, rows)
}

func armDeploymentsCSV(payload models.ArmDeploymentsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Deployments))
	for _, deployment := range payload.Deployments {
		rows = append(rows, []string{
			deployment.Duration,
			deployment.ID,
			deployment.Mode,
			deployment.Name,
			fmt.Sprintf("%d", deployment.OutputResourceCount),
			fmt.Sprintf("%d", deployment.OutputsCount),
			valueOrEmpty(deployment.ParametersLink),
			join(deployment.Providers, ";"),
			deployment.ProvisioningState,
			join(deployment.RelatedIDs, ";"),
			valueOrEmpty(deployment.ResourceGroup),
			deployment.Scope,
			deployment.ScopeType,
			deployment.Summary,
			valueOrEmpty(deployment.TemplateLink),
			deployment.Timestamp,
		})
	}
	return encodeCSV([]string{
		"duration",
		"id",
		"mode",
		"name",
		"output_resource_count",
		"outputs_count",
		"parameters_link",
		"providers",
		"provisioning_state",
		"related_ids",
		"resource_group",
		"scope",
		"scope_type",
		"summary",
		"template_link",
		"timestamp",
	}, rows)
}

func endpointsCSV(payload models.EndpointsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Endpoints))
	for _, endpoint := range payload.Endpoints {
		rows = append(rows, []string{
			endpoint.Endpoint,
			endpoint.EndpointType,
			endpoint.ExposureFamily,
			endpoint.IngressPath,
			join(endpoint.RelatedIDs, ";"),
			endpoint.SourceAssetID,
			endpoint.SourceAssetKind,
			endpoint.SourceAssetName,
			endpoint.Summary,
		})
	}
	return encodeCSV([]string{
		"endpoint",
		"endpoint_type",
		"exposure_family",
		"ingress_path",
		"related_ids",
		"source_asset_id",
		"source_asset_kind",
		"source_asset_name",
		"summary",
	}, rows)
}

func networkPortsCSV(payload models.NetworkPortsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.NetworkPorts))
	for _, networkPort := range payload.NetworkPorts {
		rows = append(rows, []string{
			networkPort.AllowSourceSummary,
			networkPort.AssetID,
			networkPort.AssetName,
			networkPort.Endpoint,
			networkPort.ExposureConfidence,
			networkPort.Port,
			networkPort.Protocol,
			join(networkPort.RelatedIDs, ";"),
			networkPort.Summary,
		})
	}
	return encodeCSV([]string{
		"allow_source_summary",
		"asset_id",
		"asset_name",
		"endpoint",
		"exposure_confidence",
		"port",
		"protocol",
		"related_ids",
		"summary",
	}, rows)
}

func networkEffectiveCSV(payload models.NetworkEffectiveOutput) (string, error) {
	rows := make([][]string, 0, len(payload.EffectiveExposures))
	for _, exposure := range payload.EffectiveExposures {
		rows = append(rows, []string{
			exposure.AssetID,
			exposure.AssetName,
			join(exposure.ConstrainedPorts, ";"),
			exposure.EffectiveExposure,
			exposure.Endpoint,
			exposure.EndpointType,
			join(exposure.InternetExposedPorts, ";"),
			join(exposure.ObservedPaths, ";"),
			join(exposure.RelatedIDs, ";"),
			exposure.Summary,
		})
	}
	return encodeCSV([]string{
		"asset_id",
		"asset_name",
		"constrained_ports",
		"effective_exposure",
		"endpoint",
		"endpoint_type",
		"internet_exposed_ports",
		"observed_paths",
		"related_ids",
		"summary",
	}, rows)
}

func nicsCSV(payload models.NicsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.NicAssets))
	for _, nic := range payload.NicAssets {
		rows = append(rows, []string{
			valueOrEmpty(nic.AttachedAssetID),
			valueOrEmpty(nic.AttachedAssetName),
			nic.ID,
			nic.Name,
			valueOrEmpty(nic.NetworkSecurityGroupID),
			join(nic.PrivateIPs, ";"),
			join(nic.PublicIPIDs, ";"),
			join(nic.SubnetIDs, ";"),
			join(nic.VnetIDs, ";"),
		})
	}
	return encodeCSV([]string{
		"attached_asset_id",
		"attached_asset_name",
		"id",
		"name",
		"network_security_group_id",
		"private_ips",
		"public_ip_ids",
		"subnet_ids",
		"vnet_ids",
	}, rows)
}

func vmsCSV(payload models.VmsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.VMAssets))
	for _, vm := range payload.VMAssets {
		rows = append(rows, []string{
			vm.ID,
			join(vm.IdentityIDs, ";"),
			vm.Location,
			vm.Name,
			join(vm.NICIDs, ";"),
			vm.PowerState,
			join(vm.PrivateIPs, ";"),
			join(vm.PublicIPs, ";"),
			vm.ResourceGroup,
			vm.VMType,
		})
	}
	return encodeCSV([]string{
		"id",
		"identity_ids",
		"location",
		"name",
		"nic_ids",
		"power_state",
		"private_ips",
		"public_ips",
		"resource_group",
		"vm_type",
	}, rows)
}

func vmExtensionsCSV(payload models.VMExtensionsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.VMExtensions))
	for _, extension := range payload.VMExtensions {
		rows = append(rows, []string{
			boolPtrString(extension.AutoUpgradeMinorVersion),
			valueOrEmpty(extension.CommandClue),
			boolPtrString(extension.EnableAutomaticUpgrade),
			valueOrEmpty(extension.ExtensionType),
			join(extension.FileURIHosts, ";"),
			intPtrString(extension.FileURICount),
			valueOrEmpty(extension.ForceUpdateTag),
			extension.ID,
			join(extension.InstanceViewStatuses, ";"),
			boolPtrString(extension.KeyVaultProtectedSettings),
			extension.Location,
			extension.Name,
			boolPtrString(extension.ProtectedSettingsPresent),
			join(extension.ProvisionAfterExtensions, ";"),
			valueOrEmpty(extension.ProvisioningState),
			valueOrEmpty(extension.Publisher),
			join(extension.PublicSettingKeys, ";"),
			join(extension.RelatedIDs, ";"),
			extension.ResourceGroup,
			join(extension.RerunClues, ";"),
			join(extension.SourceClues, ";"),
			extension.Summary,
			boolPtrString(extension.SuppressFailures),
			extension.TargetID,
			join(extension.TargetIdentityIDs, ";"),
			extension.TargetKind,
			extension.TargetName,
			valueOrEmpty(extension.TypeHandlerVersion),
			valueOrEmpty(extension.VMSSOrchestrationMode),
			valueOrEmpty(extension.VMSSUpgradeMode),
		})
	}
	return encodeCSV([]string{
		"auto_upgrade_minor_version",
		"command_clue",
		"enable_automatic_upgrade",
		"extension_type",
		"file_uri_hosts",
		"file_uri_count",
		"force_update_tag",
		"id",
		"instance_view_statuses",
		"key_vault_protected_settings",
		"location",
		"name",
		"protected_settings_present",
		"provision_after_extensions",
		"provisioning_state",
		"publisher",
		"public_setting_keys",
		"related_ids",
		"resource_group",
		"rerun_clues",
		"source_clues",
		"summary",
		"suppress_failures",
		"target_id",
		"target_identity_ids",
		"target_kind",
		"target_name",
		"type_handler_version",
		"vmss_orchestration_mode",
		"vmss_upgrade_mode",
	}, rows)
}

func vmssCSV(payload models.VmssOutput) (string, error) {
	rows := make([][]string, 0, len(payload.VmssAssets))
	for _, vmss := range payload.VmssAssets {
		rows = append(rows, []string{
			intString(vmss.ApplicationGatewayBackendPoolCount),
			valueOrEmpty(vmss.ClientID),
			vmss.ID,
			join(vmss.IdentityIDs, ";"),
			valueOrEmpty(vmss.IdentityType),
			intString(vmss.InboundNATPoolCount),
			intPtrString(vmss.InstanceCount),
			intString(vmss.LoadBalancerBackendPoolCount),
			vmss.Location,
			vmss.Name,
			intString(vmss.NICConfigurationCount),
			valueOrEmpty(vmss.OrchestrationMode),
			boolPtrString(vmss.Overprovision),
			valueOrEmpty(vmss.PrincipalID),
			intString(vmss.PublicIPConfigurationCount),
			join(vmss.RelatedIDs, ";"),
			vmss.ResourceGroup,
			boolPtrString(vmss.SinglePlacementGroup),
			valueOrEmpty(vmss.SKUName),
			join(vmss.SubnetIDs, ";"),
			vmss.Summary,
			valueOrEmpty(vmss.UpgradeMode),
			boolPtrString(vmss.ZoneBalance),
			join(vmss.Zones, ";"),
		})
	}
	return encodeCSV([]string{
		"application_gateway_backend_pool_count",
		"client_id",
		"id",
		"identity_ids",
		"identity_type",
		"inbound_nat_pool_count",
		"instance_count",
		"load_balancer_backend_pool_count",
		"location",
		"name",
		"nic_configuration_count",
		"orchestration_mode",
		"overprovision",
		"principal_id",
		"public_ip_configuration_count",
		"related_ids",
		"resource_group",
		"single_placement_group",
		"sku_name",
		"subnet_ids",
		"summary",
		"upgrade_mode",
		"zone_balance",
		"zones",
	}, rows)
}

func workloadsCSV(payload models.WorkloadsOutput) (string, error) {
	rows := make([][]string, 0, len(payload.Workloads))
	for _, workload := range payload.Workloads {
		rows = append(rows, []string{
			workload.AssetID,
			workload.AssetKind,
			workload.AssetName,
			join(workload.Endpoints, ";"),
			join(workload.ExposureFamilies, ";"),
			valueOrEmpty(workload.IdentityClientID),
			join(workload.IdentityIDs, ";"),
			valueOrEmpty(workload.IdentityPrincipalID),
			valueOrEmpty(workload.IdentityType),
			join(workload.IngressPaths, ";"),
			workload.Location,
			join(workload.RelatedIDs, ";"),
			workload.ResourceGroup,
			workload.Summary,
		})
	}
	return encodeCSV([]string{
		"asset_id",
		"asset_kind",
		"asset_name",
		"endpoints",
		"exposure_families",
		"identity_client_id",
		"identity_ids",
		"identity_principal_id",
		"identity_type",
		"ingress_paths",
		"location",
		"related_ids",
		"resource_group",
		"summary",
	}, rows)
}

func permissionsCSV(payload models.PermissionsOutput) (string, error) {
	headers := []string{
		"principal_id",
		"display_name",
		"principal_type",
		"priority",
		"high_impact_roles",
		"all_role_names",
		"role_assignment_count",
		"scope_count",
		"scope_ids",
		"privileged",
		"is_current_identity",
		"operator_signal",
		"next_review",
		"summary",
	}
	rows := make([][]string, 0, len(payload.Permissions))
	for _, permission := range payload.Permissions {
		rows = append(rows, []string{
			permission.PrincipalID,
			permission.DisplayName,
			permission.PrincipalType,
			permission.Priority,
			join(permission.HighImpactRoles, ";"),
			join(permission.AllRoleNames, ";"),
			fmt.Sprintf("%d", permission.RoleAssignmentCount),
			fmt.Sprintf("%d", permission.ScopeCount),
			join(permission.ScopeIDs, ";"),
			fmt.Sprintf("%t", permission.Privileged),
			fmt.Sprintf("%t", permission.IsCurrentIdentity),
			permission.OperatorSignal,
			permission.NextReview,
			permission.Summary,
		})
	}
	return encodeCSV(headers, rows)
}

func principalsCSV(payload models.PrincipalsOutput) (string, error) {
	headers := []string{
		"attached_to",
		"display_name",
		"id",
		"identity_names",
		"identity_types",
		"is_current_identity",
		"principal_type",
		"role_assignment_count",
		"role_names",
		"scope_ids",
		"sources",
		"tenant_id",
	}
	rows := make([][]string, 0, len(payload.Principals))
	for _, principal := range payload.Principals {
		rows = append(rows, []string{
			join(principal.AttachedTo, ";"),
			valueOrEmpty(principal.DisplayName),
			principal.ID,
			join(principal.IdentityNames, ";"),
			join(principal.IdentityTypes, ";"),
			fmt.Sprintf("%t", principal.IsCurrentIdentity),
			principal.PrincipalType,
			fmt.Sprintf("%d", principal.RoleAssignmentCount),
			join(principal.RoleNames, ";"),
			join(principal.ScopeIDs, ";"),
			join(principal.Sources, ";"),
			valueOrEmpty(principal.TenantID),
		})
	}
	return encodeCSV(headers, rows)
}

func privescCSV(payload models.PrivescOutput) (string, error) {
	headers := []string{
		"asset",
		"current_identity",
		"impact_roles",
		"starting_foothold",
		"missing_proof",
		"next_review",
		"operator_signal",
		"path_type",
		"target",
		"preferred",
		"preferred_reason",
		"priority",
		"principal",
		"principal_id",
		"principal_type",
		"proven_path",
		"related_ids",
		"summary",
	}
	rows := make([][]string, 0, len(payload.Paths))
	for _, path := range payload.Paths {
		rows = append(rows, []string{
			valueOrEmpty(path.Asset),
			fmt.Sprintf("%t", path.CurrentIdentity),
			join(path.ImpactRoles, ";"),
			path.StartingFoothold,
			path.MissingProof,
			path.NextReview,
			path.OperatorSignal,
			path.PathType,
			path.Target,
			fmt.Sprintf("%t", path.Preferred),
			path.PreferredReason,
			path.Priority,
			path.Principal,
			path.PrincipalID,
			path.PrincipalType,
			path.ProvenPath,
			join(path.RelatedIDs, ";"),
			path.Summary,
		})
	}
	return encodeCSV(headers, rows)
}

func lighthouseCSV(payload models.LighthouseOutput) (string, error) {
	headers := []string{
		"authorization_count",
		"definition_provisioning_state",
		"description",
		"eligible_authorization_count",
		"eligible_principal_count",
		"eligible_role_names",
		"has_delegated_role_assignments",
		"has_owner_role",
		"has_user_access_administrator",
		"id",
		"managed_by_tenant_id",
		"managed_by_tenant_name",
		"managee_tenant_id",
		"managee_tenant_name",
		"name",
		"plan_name",
		"plan_product",
		"plan_publisher",
		"principal_count",
		"provisioning_state",
		"registration_definition_id",
		"registration_definition_name",
		"related_ids",
		"resource_group",
		"role_names",
		"scope_display_name",
		"scope_id",
		"scope_type",
		"strongest_role_name",
		"summary",
	}
	rows := make([][]string, 0, len(payload.LighthouseDelegations))
	for _, delegation := range payload.LighthouseDelegations {
		rows = append(rows, []string{
			intString(delegation.AuthorizationCount),
			valueOrEmpty(delegation.DefinitionProvisioningState),
			valueOrEmpty(delegation.Description),
			intString(delegation.EligibleAuthorizationCount),
			intString(delegation.EligiblePrincipalCount),
			join(delegation.EligibleRoleNames, ";"),
			boolString(delegation.HasDelegatedRoleAssignments),
			boolString(delegation.HasOwnerRole),
			boolString(delegation.HasUserAccessAdministrator),
			delegation.ID,
			valueOrEmpty(delegation.ManagedByTenantID),
			valueOrEmpty(delegation.ManagedByTenantName),
			valueOrEmpty(delegation.ManageeTenantID),
			valueOrEmpty(delegation.ManageeTenantName),
			delegation.Name,
			valueOrEmpty(delegation.PlanName),
			valueOrEmpty(delegation.PlanProduct),
			valueOrEmpty(delegation.PlanPublisher),
			intString(delegation.PrincipalCount),
			valueOrEmpty(delegation.ProvisioningState),
			valueOrEmpty(delegation.RegistrationDefinitionID),
			valueOrEmpty(delegation.RegistrationDefinitionName),
			join(delegation.RelatedIDs, ";"),
			valueOrEmpty(delegation.ResourceGroup),
			join(delegation.RoleNames, ";"),
			valueOrEmpty(delegation.ScopeDisplayName),
			delegation.ScopeID,
			delegation.ScopeType,
			valueOrEmpty(delegation.StrongestRoleName),
			delegation.Summary,
		})
	}
	return encodeCSV(headers, rows)
}

func crossTenantCSV(payload models.CrossTenantOutput) (string, error) {
	headers := []string{
		"attack_path",
		"id",
		"name",
		"posture",
		"priority",
		"related_ids",
		"scope",
		"signal_type",
		"summary",
		"tenant_id",
		"tenant_name",
	}
	rows := make([][]string, 0, len(payload.CrossTenantPaths))
	for _, path := range payload.CrossTenantPaths {
		rows = append(rows, []string{
			path.AttackPath,
			path.ID,
			path.Name,
			valueOrEmpty(path.Posture),
			path.Priority,
			join(path.RelatedIDs, ";"),
			valueOrEmpty(path.Scope),
			path.SignalType,
			path.Summary,
			valueOrEmpty(path.TenantID),
			valueOrEmpty(path.TenantName),
		})
	}
	return encodeCSV(headers, rows)
}

func authPoliciesCSV(payload models.AuthPoliciesOutput) (string, error) {
	headers := []string{
		"controls",
		"name",
		"policy_type",
		"related_ids",
		"scope",
		"state",
		"summary",
	}
	rows := make([][]string, 0, len(payload.AuthPolicies))
	for _, policy := range payload.AuthPolicies {
		rows = append(rows, []string{
			join(policy.Controls, ";"),
			policy.Name,
			policy.PolicyType,
			join(policy.RelatedIDs, ";"),
			valueOrEmpty(policy.Scope),
			policy.State,
			policy.Summary,
		})
	}
	return encodeCSV(headers, rows)
}

func resourceTrustsCSV(payload models.ResourceTrustsOutput) (string, error) {
	headers := []string{
		"confidence",
		"exposure",
		"related_ids",
		"resource_id",
		"resource_name",
		"resource_type",
		"summary",
		"target",
		"trust_type",
	}
	rows := make([][]string, 0, len(payload.ResourceTrusts))
	for _, trust := range payload.ResourceTrusts {
		rows = append(rows, []string{
			trust.Confidence,
			trust.Exposure,
			join(trust.RelatedIDs, ";"),
			trust.ResourceID,
			trust.ResourceName,
			trust.ResourceType,
			trust.Summary,
			trust.Target,
			trust.TrustType,
		})
	}
	return encodeCSV(headers, rows)
}

func roleTrustsCSV(payload models.RoleTrustsOutput) (string, error) {
	headers := []string{
		"trust_type",
		"source_object_id",
		"source_name",
		"source_type",
		"target_object_id",
		"target_name",
		"target_type",
		"evidence_type",
		"confidence",
		"control_primitive",
		"controlled_object_type",
		"controlled_object_name",
		"backing_service_principal_id",
		"backing_service_principal_name",
		"escalation_mechanism",
		"usable_identity_result",
		"defender_cut_point",
		"operator_signal",
		"next_review",
		"summary",
		"related_ids",
	}
	rows := make([][]string, 0, len(payload.Trusts))
	for _, trust := range payload.Trusts {
		rows = append(rows, []string{
			trust.TrustType,
			trust.SourceObjectID,
			valueOrEmpty(trust.SourceName),
			trust.SourceType,
			trust.TargetObjectID,
			valueOrEmpty(trust.TargetName),
			trust.TargetType,
			trust.EvidenceType,
			trust.Confidence,
			valueOrEmpty(trust.ControlPrimitive),
			valueOrEmpty(trust.ControlledObjectType),
			valueOrEmpty(trust.ControlledObjectName),
			valueOrEmpty(trust.BackingServicePrincipalID),
			valueOrEmpty(trust.BackingServicePrincipalName),
			valueOrEmpty(trust.EscalationMechanism),
			valueOrEmpty(trust.UsableIdentityResult),
			valueOrEmpty(trust.DefenderCutPoint),
			valueOrEmpty(trust.OperatorSignal),
			valueOrEmpty(trust.NextReview),
			trust.Summary,
			join(trust.RelatedIDs, ";"),
		})
	}
	return encodeCSV(headers, rows)
}

func appCredentialsCSV(payload models.AppCredentialsOutput) (string, error) {
	headers := []string{
		"row_class",
		"target_object_type",
		"target_object_id",
		"target_object_name",
		"backing_service_principal_id",
		"backing_service_principal_name",
		"credential_type",
		"control_path",
		"role_context",
		"tenant_context",
		"current_evidence",
		"missing_proof",
		"operator_actionability",
		"recommended_fix_focus",
		"summary",
		"related_ids",
	}
	rows := make([][]string, 0, len(payload.AppCredentials))
	for _, item := range payload.AppCredentials {
		rows = append(rows, []string{
			item.RowClass,
			item.TargetObjectType,
			item.TargetObjectID,
			item.TargetObjectName,
			valueOrEmpty(item.BackingServicePrincipalID),
			valueOrEmpty(item.BackingServicePrincipalName),
			valueOrEmpty(item.CredentialType),
			item.ControlPath,
			item.RoleContext,
			item.TenantContext,
			item.CurrentEvidence,
			item.MissingProof,
			item.OperatorActionability,
			item.RecommendedFixFocus,
			item.Summary,
			join(item.RelatedIDs, ";"),
		})
	}
	return encodeCSV(headers, rows)
}

func managedIdentitiesCSV(payload models.ManagedIdentitiesOutput) (string, error) {
	headers := []string{
		"id",
		"name",
		"identity_type",
		"principal_id",
		"client_id",
		"attached_to",
		"scope_ids",
		"operator_signal",
		"next_review",
		"summary",
	}
	rows := make([][]string, 0, len(payload.Identities))
	for _, identity := range payload.Identities {
		rows = append(rows, []string{
			identity.ID,
			identity.Name,
			identity.IdentityType,
			valueOrEmpty(identity.PrincipalID),
			valueOrEmpty(identity.ClientID),
			join(identity.AttachedTo, ";"),
			join(identity.ScopeIDs, ";"),
			valueOrEmpty(identity.OperatorSignal),
			valueOrEmpty(identity.NextReview),
			valueOrEmpty(identity.Summary),
		})
	}
	return encodeCSV(headers, rows)
}

func envVarsCSV(payload models.EnvVarsOutput) (string, error) {
	headers := []string{
		"asset_id",
		"asset_kind",
		"asset_name",
		"key_vault_reference_identity",
		"location",
		"looks_sensitive",
		"reference_target",
		"related_ids",
		"resource_group",
		"setting_name",
		"summary",
		"value_type",
		"workload_client_id",
		"workload_identity_ids",
		"workload_identity_type",
		"workload_principal_id",
	}
	rows := make([][]string, 0, len(payload.EnvVars))
	for _, envVar := range payload.EnvVars {
		rows = append(rows, []string{
			envVar.AssetID,
			envVar.AssetKind,
			envVar.AssetName,
			valueOrEmpty(envVar.KeyVaultReferenceIdentity),
			envVar.Location,
			fmt.Sprintf("%t", envVar.LooksSensitive),
			valueOrEmpty(envVar.ReferenceTarget),
			join(envVar.RelatedIDs, ";"),
			envVar.ResourceGroup,
			envVar.SettingName,
			envVar.Summary,
			envVar.ValueType,
			valueOrEmpty(envVar.WorkloadClientID),
			join(envVar.WorkloadIdentityIDs, ";"),
			valueOrEmpty(envVar.WorkloadIdentityType),
			valueOrEmpty(envVar.WorkloadPrincipalID),
		})
	}
	return encodeCSV(headers, rows)
}

func tokensCredentialsCSV(payload models.TokensCredentialsOutput) (string, error) {
	headers := []string{
		"access_path",
		"asset_id",
		"kind",
		"asset",
		"location",
		"operator_signal",
		"priority",
		"related_ids",
		"resource_group",
		"summary",
		"surface",
	}
	rows := make([][]string, 0, len(payload.Surfaces))
	for _, surface := range payload.Surfaces {
		rows = append(rows, []string{
			surface.AccessPath,
			surface.AssetID,
			surface.AssetKind,
			surface.AssetName,
			valueOrEmpty(surface.Location),
			surface.OperatorSignal,
			surface.Priority,
			join(surface.RelatedIDs, ";"),
			valueOrEmpty(surface.ResourceGroup),
			surface.Summary,
			string(surface.SurfaceType),
		})
	}
	return encodeCSV(headers, rows)
}

func inventoryCSV(payload models.InventoryOutput) (string, error) {
	headers := []string{
		"subscription_id",
		"subscription_display_name",
		"subscription_state",
		"resource_group_count",
		"resource_count",
		"top_type",
		"top_type_count",
		"top_resource_types",
		"issue_count",
		"metadata_command",
		"metadata_generated_at",
		"metadata_schema_version",
	}

	topType, topCount := topResourceType(payload.TopResourceTypes)
	row := []string{
		payload.Subscription.ID,
		payload.Subscription.DisplayName,
		payload.Subscription.State,
		fmt.Sprintf("%d", payload.ResourceGroupCount),
		fmt.Sprintf("%d", payload.ResourceCount),
		topType,
		fmt.Sprintf("%d", topCount),
		flattenResourceTypes(payload.TopResourceTypes),
		fmt.Sprintf("%d", len(payload.Issues)),
		payload.Metadata.Command,
		payload.Metadata.GeneratedAt,
		payload.Metadata.SchemaVersion,
	}
	return encodeCSV(headers, [][]string{row})
}

func whoAmICSV(payload models.WhoAmIOutput) (string, error) {
	row := []string{
		payload.TenantID,
		payload.Subscription.ID,
		payload.Subscription.DisplayName,
		payload.Subscription.State,
		payload.Principal.ID,
		payload.Principal.PrincipalType,
		payload.Principal.DisplayName,
		payload.Principal.TenantID,
		joinScopes(payload.EffectiveScopes, "id"),
		joinScopes(payload.EffectiveScopes, "display_name"),
		payload.Metadata.Command,
		payload.Metadata.GeneratedAt,
		payload.Metadata.SchemaVersion,
		valueOrEmpty(payload.Metadata.TokenSource),
		valueOrEmpty(payload.Metadata.AuthMode),
	}
	headers := []string{
		"tenant_id",
		"subscription_id",
		"subscription_display_name",
		"subscription_state",
		"principal_id",
		"principal_type",
		"principal_display_name",
		"principal_tenant_id",
		"effective_scope_ids",
		"effective_scope_display_names",
		"metadata_command",
		"metadata_generated_at",
		"metadata_schema_version",
		"metadata_token_source",
		"metadata_auth_mode",
	}
	return encodeCSV(headers, [][]string{row})
}

func joinScopes(scopes []models.ScopeRef, field string) string {
	values := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		switch field {
		case "id":
			values = append(values, scope.ID)
		case "display_name":
			values = append(values, scope.DisplayName)
		}
	}
	return join(values, ";")
}

func flattenResourceTypes(resourceTypes models.TopResourceTypes) string {
	if len(resourceTypes) == 0 {
		return ""
	}
	keys := sortedResourceTypeKeys(resourceTypes)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, resourceTypes[key]))
	}
	return join(parts, ";")
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intString(value int) string {
	return fmt.Sprintf("%d", value)
}

func intPtrString(value *int) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%d", *value)
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func persistenceCSVStepStatus(steps []models.PersistenceCapabilityStep, action string) string {
	for _, step := range steps {
		if step.Action == action {
			return step.Status
		}
	}
	return ""
}

func persistenceCSVRoleSummary(context *models.PersistenceRoleContext) string {
	if context == nil {
		return ""
	}
	return context.Summary
}

func intJoin(values []int, separator string) string {
	if len(values) == 0 {
		return ""
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%d", value))
	}
	return join(parts, separator)
}

func containerAppsJobEventRulesCSV(rules []models.ContainerAppsJobEventRule) string {
	if len(rules) == 0 {
		return ""
	}
	parts := make([]string, 0, len(rules))
	for _, rule := range rules {
		section := rule.Name
		if rule.Type != "" {
			section += ":" + rule.Type
		}
		if len(rule.AuthSecretRefs) > 0 {
			section += ":auth=" + join(rule.AuthSecretRefs, "|")
		}
		if rule.Identity != nil && *rule.Identity != "" {
			section += ":identity=" + *rule.Identity
		}
		parts = append(parts, section)
	}
	return join(parts, ";")
}

func boolPtrString(value *bool) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%t", *value)
}

func jsonStringSlice(values []string) string {
	encoded, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func jsonValue(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}
