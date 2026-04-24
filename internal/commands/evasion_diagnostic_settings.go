package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

var evasionDiagnosticSettingsSteps = []familyStepDefinition{
	{
		Action:           "create or modify diagnostic setting",
		APISurface:       "Microsoft.Insights/diagnosticSettings/write",
		NeedsWrite:       true,
		DownstreamEffect: "Sets the Azure Monitor export object on the source resource.",
		Boundary:         "Does not prove a source event occurred or was collected.",
	},
	{
		Action:           "pick source resource",
		APISurface:       "source resource ARM scope",
		NeedsWrite:       true,
		DownstreamEffect: "Chooses which resource's logs, metrics, or activity export posture is shaped.",
		Boundary:         "Source value is ranked from visible type and current settings, not from future activity.",
	},
	{
		Action:           "choose exported categories",
		APISurface:       "logs, metrics, category groups",
		NeedsWrite:       true,
		DownstreamEffect: "Controls which visible categories or metrics are exported to the configured sink.",
		Boundary:         "Default output names categories present in visible settings; supported-but-absent categories need a catalog pass.",
	},
	{
		Action:           "choose destination sink",
		APISurface:       "workspaceId, storageAccountId, eventHubAuthorizationRuleId, marketplacePartnerId",
		NeedsWrite:       true,
		DownstreamEffect: "Moves selected telemetry toward Log Analytics, Storage, Event Hubs, or a partner destination.",
		Boundary:         "Does not call a destination wrong without an expected SOC sink baseline.",
	},
	{
		Action:           "save or edit setting",
		APISurface:       "diagnosticSettings/write",
		NeedsWrite:       true,
		DownstreamEffect: "Makes the category and destination posture durable as Azure configuration.",
		Boundary:         "Persistence here means stored management-plane posture, not proof of sink delivery.",
	},
	{
		Action:           "shape visibility",
		APISurface:       "selected categories and destination IDs",
		NeedsWrite:       true,
		DownstreamEffect: "Can leave monitoring objects present while selected evidence is not exported by the visible setting or is routed elsewhere.",
		Boundary:         "Does not claim detector failure or malicious intent from posture alone.",
	},
	{
		Action:           "reuse later",
		APISurface:       "stored diagnostic setting",
		NeedsWrite:       true,
		DownstreamEffect: "The export posture remains until another actor or automation changes it.",
		Boundary:         "Change author and timing require activity-log history.",
	},
	{
		Action:           "blend as monitoring admin",
		APISurface:       "diagnostic setting metadata",
		DownstreamEffect: "Common cover stories include onboarding, migration, cost control, archive routing, and category cleanup.",
		Boundary:         "Cover story is an administrative explanation, not a claim of benign or malicious intent.",
	},
}

func buildEvasionDiagnosticSettingsOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.EvasionSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	settingsFuture := runGroupedCommandOutput[models.DiagnosticSettingsOutput](group, ctx, request, diagnosticSettingsHandler(provider, now), "diagnostic-settings")
	evidenceFutures := runFamilyEvidence(group, ctx, request, provider, now)

	settings, err := settingsFuture.wait()
	if err != nil {
		return nil, err
	}
	evidence, err := evidenceFutures.wait()
	if err != nil {
		return nil, err
	}

	sinks := providers.MonitoringSinksFromDiagnosticReferences(settings.Sources)
	sources := make([]models.EvasionDiagnosticSettingsSource, 0, len(settings.Sources))
	for _, source := range settings.Sources {
		control, controlOK := evasionDiagnosticSettingsControl(source.ID, evidence.principal.currentIdentityAssignments)
		currentContext := evasionDiagnosticSettingsIdentityContext(evidence.principal.currentIdentity, control, controlOK)
		rank, reason := evasionDiagnosticSettingsDisruptionRank(source, controlOK)
		sources = append(sources, models.EvasionDiagnosticSettingsSource{
			ID:                     source.ID,
			Name:                   source.Name,
			ResourceGroup:          source.ResourceGroup,
			Location:               source.Location,
			DisruptionRank:         rank,
			DisruptionReason:       reason,
			CapabilitySteps:        evasionDiagnosticSettingsCapabilitySteps(controlOK),
			CurrentIdentityContext: currentContext,
			CurrentState:           evasionDiagnosticSettingsState(source),
			NotCollectedByDefault:  evasionDiagnosticSettingsNotCollectedByDefault(),
			Summary:                evasionDiagnosticSettingsSummary(source, rank, controlOK),
			RelatedIDs:             mergeRelatedIDs(source.RelatedIDs),
		})
	}
	sort.SliceStable(sources, func(i, j int) bool {
		if sources[i].DisruptionRank != sources[j].DisruptionRank {
			return sources[i].DisruptionRank > sources[j].DisruptionRank
		}
		return sources[i].Name < sources[j].Name
	})

	issues := familyIssues(settings.Issues, evidence)

	return models.EvasionDiagnosticSettingsOutput{
		Metadata: scopedMetadata(
			now,
			request,
			firstNonEmpty(request.Tenant, stringPtrValue(settings.Metadata.TenantID), stringPtrValue(evidence.permissions.Metadata.TenantID)),
			firstNonEmpty(request.Subscription, stringPtrValue(settings.Metadata.SubscriptionID), stringPtrValue(evidence.permissions.Metadata.SubscriptionID)),
			"evasion",
		),
		GroupedCommandName: "evasion",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		MonitoringSinks:    sinks,
		Sources:            sources,
		Issues:             issues,
	}, nil
}

func evasionDiagnosticSettingsControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	return evasionDCRBestControl(resourceID, assignments, "Microsoft.Insights/diagnosticSettings/write")
}

func evasionDiagnosticSettingsIdentityContext(
	currentIdentity models.PermissionRow,
	control persistenceCurrentIdentityControl,
	controlOK bool,
) *models.EvasionRoleContext {
	if strings.TrimSpace(currentIdentity.DisplayName) == "" && !controlOK {
		return nil
	}
	name := firstNonEmpty(currentIdentity.DisplayName, "current identity")
	roleNames := append([]string{}, currentIdentity.HighImpactRoles...)
	if len(roleNames) == 0 {
		roleNames = append(roleNames, currentIdentity.AllRoleNames...)
	}
	scopeIDs := append([]string{}, currentIdentity.ScopeIDs...)
	summary := "Current foothold identity is visible, but diagnostic settings write control is not proven here."
	controlLabel := "not proven"
	if controlOK {
		summary = fmt.Sprintf("Current foothold `%s` has visible diagnostic settings write control.", name)
		roleNames = []string{evasionControlRoleName(control)}
		scopeIDs = []string{control.ScopeID}
		controlLabel = "diagnostic settings write"
	}
	return &models.EvasionRoleContext{
		Name:         name,
		Kind:         "current-foothold",
		PrincipalID:  stringPtrIf(currentIdentity.PrincipalID),
		RoleNames:    dedupeStrings(roleNames),
		ScopeIDs:     dedupeStrings(scopeIDs),
		ControlLabel: controlLabel,
		Summary:      summary,
	}
}

func evasionDiagnosticSettingsCapabilitySteps(controlOK bool) []models.EvasionCapabilityStep {
	return familyCapabilitySteps(evasionDiagnosticSettingsSteps, controlOK)
}

func evasionDiagnosticSettingsState(source models.DiagnosticSettingsSource) models.EvasionDiagnosticSettingsState {
	return models.EvasionDiagnosticSettingsState{
		SourceType:             source.Type,
		DiagnosticSettingCount: source.DiagnosticSettingCount,
		EnabledCategories:      append([]string{}, source.EnabledCategories...),
		NotExportedCategories:  evasionDiagnosticSettingsNotExported(source),
		SupportedCategories:    append([]string{}, source.SupportedCategories...),
		SupportedCategoryProof: source.SupportedCategoryCatalog,
		CategoryGroups:         append([]string{}, source.CategoryGroups...),
		HighSignalCategories:   append([]string{}, source.HighSignalCategories...),
		DestinationTypes:       append([]string{}, source.DestinationTypes...),
		HasNonWorkspaceSink:    source.HasNonWorkspaceDestination,
		ExportPosture:          evasionDiagnosticSettingsExportPosture(source),
		DestinationPosture:     evasionDiagnosticSettingsDestinationPosture(source),
	}
}

func evasionDiagnosticSettingsNotExported(source models.DiagnosticSettingsSource) []string {
	if source.SupportedCategoryCatalog {
		return append([]string{}, source.NotExportedSupported...)
	}
	return append([]string{}, source.DisabledCategories...)
}

func evasionDiagnosticSettingsExportPosture(source models.DiagnosticSettingsSource) string {
	if !source.HasDiagnosticSettings {
		return "no visible diagnostic settings on this source"
	}
	if len(evasionDiagnosticSettingsNotExported(source)) > 0 && source.SupportedCategoryCatalog {
		return "supported categories are not exported by visible settings"
	}
	if len(evasionDiagnosticSettingsNotExported(source)) > 0 {
		return "some categories present in visible settings are not exported"
	}
	return "visible settings export selected categories or metrics"
}

func evasionDiagnosticSettingsDestinationPosture(source models.DiagnosticSettingsSource) string {
	if len(source.DestinationTypes) == 0 {
		return "no destination visible"
	}
	return "operator-selected destinations visible: " + strings.Join(source.DestinationTypes, ", ")
}

func evasionDiagnosticSettingsDisruptionRank(source models.DiagnosticSettingsSource, controlOK bool) (int, string) {
	rank := 1
	reasons := []string{}
	notExported := evasionDiagnosticSettingsNotExported(source)
	if len(notExported) > 0 && source.HasHighSignalLogPosture {
		rank = 5
		if source.SupportedCategoryCatalog {
			reasons = append(reasons, "supported high-signal categories are not exported by visible settings")
		} else {
			reasons = append(reasons, "high-signal categories are present in visible settings but not exported")
		}
	} else if source.HasNonWorkspaceDestination && source.HasHighSignalLogPosture {
		rank = 4
		reasons = append(reasons, "high-value telemetry is routed to a non-Log Analytics sink")
	} else if source.HasNonWorkspaceDestination {
		rank = 3
		reasons = append(reasons, "selected telemetry is routed outside Log Analytics")
	} else if !source.HasDiagnosticSettings && source.HasHighSignalLogPosture {
		rank = 2
		reasons = append(reasons, "high-value source has no visible diagnostic setting, but supported-category proof is not collected")
	}
	if controlOK {
		reasons = append(reasons, "current identity has visible diagnostic settings write control")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "visible posture does not support a stronger dynamic disruption ranking")
	}
	return rank, strings.Join(reasons, "; ")
}

func evasionDiagnosticSettingsNotCollectedByDefault() []models.EvasionBoundaryNote {
	return []models.EvasionBoundaryNote{
		{
			Name:           "supported category catalog",
			Classification: "API/noise",
			Reason:         "Default output attempts the supported-category catalog, but Azure can reject category reads for some source types; those are reported as collection issues rather than treated as unsupported categories.",
		},
		{
			Name:           "activity-log change history",
			Classification: "API/noise",
			Reason:         "Actor, timing, quick revert, and maintenance-window proof require history collection outside the default posture view.",
		},
		{
			Name:           "sink contents",
			Classification: "proof boundary",
			Reason:         "Log Analytics, Storage, Event Hub, or partner sink contents are data/content evidence, not management-plane posture.",
		},
		{
			Name:           "detector wiring",
			Classification: "proof boundary",
			Reason:         "Sentinel and defender rule dependencies are not inspected, so the command cannot claim a detection failed.",
		},
		{
			Name:           "expected SOC destination baseline",
			Classification: "scope/sequencing",
			Reason:         "Destination drift needs a defended expected sink model before the tool can call the current sink wrong.",
		},
	}
}

func evasionDiagnosticSettingsSummary(source models.DiagnosticSettingsSource, rank int, controlOK bool) string {
	parts := []string{fmt.Sprintf("source %q ranks %d/5 for diagnostic-settings truth-disruption posture", source.Name, rank)}
	if notExported := evasionDiagnosticSettingsNotExported(source); len(notExported) > 0 {
		label := "not exported by visible setting"
		if source.SupportedCategoryCatalog {
			label = "supported but not exported"
		}
		parts = append(parts, label+": "+strings.Join(notExported, ", "))
	}
	if len(source.DestinationTypes) > 0 {
		parts = append(parts, "destinations: "+strings.Join(source.DestinationTypes, ", "))
	}
	if controlOK {
		parts = append(parts, "current identity can modify diagnostic settings from visible RBAC")
	} else {
		parts = append(parts, "current identity write control is not proven")
	}
	return strings.Join(parts, "; ") + "."
}
