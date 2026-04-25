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

var evasionAppInsightsSteps = []familyStepDefinition{
	{Action: "change instrumentation posture", APISurface: "Microsoft.Web/sites/config/write", NeedsWrite: true, DownstreamEffect: "Changes app settings that can control Application Insights instrumentation behavior.", Boundary: "Does not read code-level instrumentation bodies."},
	{Action: "choose telemetry target", APISurface: "Application Insights component and app settings", NeedsWrite: true, DownstreamEffect: "Selects the instrumented app or function where telemetry is shaped.", Boundary: "A visible setting name is a posture clue, not proof of emitted telemetry."},
	{Action: "configure sampling", APISurface: "sampling app settings or SDK/OpenTelemetry config clues", NeedsWrite: true, DownstreamEffect: "Can reduce retained request, dependency, trace, or exception examples while dashboards remain alive.", Boundary: "Does not prove true unsampled event count."},
	{Action: "configure filtering or logging level", APISurface: "filter, processor, or logging-level setting clues", NeedsWrite: true, DownstreamEffect: "Can narrow selected telemetry types before investigators query Application Insights.", Boundary: "Does not prove filtered events occurred."},
	{Action: "preserve app-side config", APISurface: "stored app settings", NeedsWrite: true, DownstreamEffect: "The instrumentation posture remains as normal application configuration until changed.", Boundary: "Change timing and author require history."},
	{Action: "blend as observability tuning", APISurface: "app setting names and component posture", DownstreamEffect: "Common cover stories include cost control, health-check filtering, privacy, and performance tuning.", Boundary: "Cover story is not an intent claim."},
}

func buildEvasionAppInsightsOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.EvasionSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	expected := helperArtifactExpectedSessions(ctx, request, provider, now, "appinsights", "permissions", "rbac")
	appInsightsFuture := runHelperOutput[models.AppInsightsOutput](group, ctx, request, appInsightsHandler(provider, now), "appinsights", expected)
	evidenceFutures := runFamilyEvidenceWithExpected(group, ctx, request, provider, now, expected)

	appInsights, appInsightsSource, err := appInsightsFuture.waitWithSource()
	if err != nil {
		return nil, err
	}
	evidence, err := evidenceFutures.wait()
	if err != nil {
		return nil, err
	}

	targets := make([]models.EvasionAppInsightsTarget, 0, len(appInsights.Targets))
	for _, target := range appInsights.Targets {
		control, controlOK := evasionAppInsightsControl(target.ID, evidence.principal.currentIdentityAssignments)
		rank, reason := evasionAppInsightsDisruptionRank(target, controlOK)
		targets = append(targets, models.EvasionAppInsightsTarget{
			ID:                     target.ID,
			Name:                   target.Name,
			ResourceGroup:          target.ResourceGroup,
			Location:               target.Location,
			DisruptionRank:         rank,
			DisruptionReason:       reason,
			CapabilitySteps:        evasionAppInsightsCapabilitySteps(controlOK),
			CurrentIdentityContext: evasionAppInsightsIdentityContext(evidence.principal.currentIdentity, control, controlOK),
			CurrentState:           evasionAppInsightsState(target),
			NotCollectedByDefault:  evasionAppInsightsNotCollectedByDefault(),
			Summary:                evasionAppInsightsSummary(target, rank, controlOK),
			RelatedIDs:             mergeRelatedIDs(target.RelatedIDs),
		})
	}
	sort.SliceStable(targets, func(i, j int) bool {
		if targets[i].DisruptionRank != targets[j].DisruptionRank {
			return targets[i].DisruptionRank > targets[j].DisruptionRank
		}
		return targets[i].Name < targets[j].Name
	})

	issues := familyIssues(appInsights.Issues, evidence)

	return models.EvasionAppInsightsOutput{
		Metadata: withSessionArtifacts(
			scopedMetadata(
				now,
				request,
				firstNonEmpty(request.Tenant, stringPtrValue(appInsights.Metadata.TenantID), stringPtrValue(evidence.permissions.Metadata.TenantID)),
				firstNonEmpty(request.Subscription, stringPtrValue(appInsights.Metadata.SubscriptionID), stringPtrValue(evidence.permissions.Metadata.SubscriptionID)),
				"evasion",
			),
			appendSessionArtifact(evidence.sessionArtifacts, appInsightsSource),
		),
		GroupedCommandName: "evasion",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		Targets:            targets,
		Components:         append([]models.AppInsightsComponent{}, appInsights.Components...),
		Issues:             issues,
	}, nil
}

func evasionAppInsightsControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	return evasionDCRBestControl(resourceID, assignments, "Microsoft.Web/sites/config/write")
}

func evasionAppInsightsIdentityContext(
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
	summary := "Current foothold identity is visible, but app configuration write control is not proven here."
	controlLabel := "not proven"
	if controlOK {
		summary = fmt.Sprintf("Current foothold `%s` has visible app configuration write control.", name)
		roleNames = []string{evasionControlRoleName(control)}
		scopeIDs = []string{control.ScopeID}
		controlLabel = "app config write"
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

func evasionAppInsightsCapabilitySteps(controlOK bool) []models.EvasionCapabilityStep {
	return familyCapabilitySteps(evasionAppInsightsSteps, controlOK)
}

func evasionAppInsightsState(target models.AppInsightsAppTarget) models.EvasionAppInsightsState {
	return models.EvasionAppInsightsState{
		Kind:                  target.Kind,
		InstrumentationClues:  append([]string{}, target.InstrumentationClues...),
		SamplingClues:         append([]string{}, target.SamplingClues...),
		FilteringClues:        append([]string{}, target.FilteringClues...),
		LoggingLevelClues:     append([]string{}, target.LoggingLevelClues...),
		VisibleTelemetryTypes: append([]string{}, target.VisibleTelemetryTypes...),
		Posture:               evasionAppInsightsPosture(target),
	}
}

func evasionAppInsightsPosture(target models.AppInsightsAppTarget) string {
	parts := []string{}
	if len(target.FilteringClues) > 0 {
		parts = append(parts, "filtering clue(s) visible")
	}
	if len(target.SamplingClues) > 0 {
		parts = append(parts, "sampling clue(s) visible")
	}
	if len(target.LoggingLevelClues) > 0 {
		parts = append(parts, "logging-level clue(s) visible")
	}
	if len(parts) == 0 {
		return "instrumentation clue(s) only"
	}
	return strings.Join(parts, "; ")
}

func evasionAppInsightsDisruptionRank(target models.AppInsightsAppTarget, controlOK bool) (int, string) {
	rank := 1
	reasons := []string{}
	if len(target.FilteringClues) > 0 && len(target.SamplingClues) > 0 {
		rank = 5
		reasons = append(reasons, "filtering and sampling posture clues are both visible")
	} else if len(target.FilteringClues) > 0 {
		rank = 4
		reasons = append(reasons, "filtering posture clues can exclude selected telemetry")
	} else if len(target.SamplingClues) > 0 {
		rank = 3
		reasons = append(reasons, "sampling posture clues can reduce retained event examples")
	} else if len(target.LoggingLevelClues) > 0 {
		rank = 2
		reasons = append(reasons, "logging-level posture clues can narrow trace detail")
	}
	if controlOK {
		reasons = append(reasons, "current identity has visible app configuration write control")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "visible posture does not support a stronger dynamic disruption ranking")
	}
	return rank, strings.Join(reasons, "; ")
}

func evasionAppInsightsNotCollectedByDefault() []models.EvasionBoundaryNote {
	return []models.EvasionBoundaryNote{
		{Name: "setting values", Classification: "recon safety", Reason: "Default output uses setting names as posture clues and does not print instrumentation keys or connection strings."},
		{Name: "code-level processors", Classification: "proof boundary", Reason: "Telemetry processor bodies usually live in source code or binaries, outside management-plane posture."},
		{Name: "host.json body", Classification: "collector issue", Reason: "Function sampling can live in host.json; this helper only uses visible app setting names by default."},
		{Name: "true unsampled count", Classification: "proof boundary", Reason: "Current posture cannot prove how many events were dropped or retained."},
		{Name: "detector failure", Classification: "proof boundary", Reason: "The command does not inspect detections, so it cannot claim a rule missed activity."},
	}
}

func evasionAppInsightsSummary(target models.AppInsightsAppTarget, rank int, controlOK bool) string {
	parts := []string{fmt.Sprintf("target %q ranks %d/5 for Application Insights truth-disruption posture", target.Name, rank)}
	if len(target.FilteringClues) > 0 {
		parts = append(parts, fmt.Sprintf("%d filtering clue(s)", len(target.FilteringClues)))
	}
	if len(target.SamplingClues) > 0 {
		parts = append(parts, fmt.Sprintf("%d sampling clue(s)", len(target.SamplingClues)))
	}
	if controlOK {
		parts = append(parts, "current identity can modify app configuration from visible RBAC")
	} else {
		parts = append(parts, "current identity write control is not proven")
	}
	return strings.Join(parts, "; ") + "."
}
