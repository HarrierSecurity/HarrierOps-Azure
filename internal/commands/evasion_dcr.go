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

type evasionDCRStepDefinition struct {
	Action           string
	APISurface       string
	NeedsRuleWrite   bool
	NeedsAssociation bool
	DownstreamEffect string
	Boundary         string
}

var evasionDCRSteps = []evasionDCRStepDefinition{
	{
		Action:           "choose or create DCR",
		APISurface:       "Microsoft.Insights/dataCollectionRules/write",
		NeedsRuleWrite:   true,
		DownstreamEffect: "Sets the Azure Monitor object that can define collection, data flows, destinations, and transform posture.",
		Boundary:         "Does not prove a monitored agent has applied the rule.",
	},
	{
		Action:           "associate monitored scope",
		APISurface:       "Microsoft.Insights/dataCollectionRuleAssociations/write",
		NeedsAssociation: true,
		DownstreamEffect: "Selects which visible resource scope receives the DCR collection and routing posture.",
		Boundary:         "Does not prove runtime agent state or log arrival.",
	},
	{
		Action:           "select data sources and streams",
		APISurface:       "dataSources / dataFlows.streams",
		NeedsRuleWrite:   true,
		DownstreamEffect: "Controls which telemetry classes are collected, including host and security-adjacent streams when present.",
		Boundary:         "Ranks by visible stream value only; missing expected streams require a defended baseline.",
	},
	{
		Action:           "configure data flows and transformations",
		APISurface:       "dataFlows.transformKql",
		NeedsRuleWrite:   true,
		DownstreamEffect: "Can filter or reshape selected records before storage while the pipeline still appears configured.",
		Boundary:         "Prints only transform presence, fingerprint, and length; it does not print transformKql or infer intent.",
	},
	{
		Action:           "select destinations",
		APISurface:       "destinations / dataFlows.destinations",
		NeedsRuleWrite:   true,
		DownstreamEffect: "Chooses where collected data goes, which can preserve logging while moving it away from a SOC workspace.",
		Boundary:         "Does not claim destination drift without an expected destination model.",
	},
	{
		Action:           "save or re-associate rule",
		APISurface:       "DCR write plus association write",
		NeedsRuleWrite:   true,
		NeedsAssociation: true,
		DownstreamEffect: "Makes the Azure-side collection posture durable as management-plane configuration.",
		Boundary:         "Persistence here means Azure configuration remains until changed; runtime application is not proven.",
	},
	{
		Action:           "shape defender truth",
		APISurface:       "streams, dataFlows, destinations, transformKql",
		NeedsRuleWrite:   true,
		DownstreamEffect: "Visible levers can narrow collection, reroute telemetry, or transform records without a full logging disablement.",
		Boundary:         "Does not claim malicious filtering or downstream detector failure from posture alone.",
	},
	{
		Action:           "preserve Azure-side config",
		APISurface:       "stored DCR and association resources",
		NeedsRuleWrite:   true,
		DownstreamEffect: "The changed rule or association can remain in place like normal monitoring migration, cost, or schema configuration.",
		Boundary:         "History, author, and timing require activity-log evidence not collected here by default.",
	},
	{
		Action:           "blend as monitoring change",
		APISurface:       "DCR metadata and ARM posture",
		DownstreamEffect: "Common cover stories include AMA migration, workspace consolidation, cost control, schema normalization, and noise reduction.",
		Boundary:         "Cover story is an administrative explanation, not a claim of benign or malicious intent.",
	},
}

func buildEvasionDCROutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.EvasionSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	expected := helperArtifactExpectedSessions(ctx, request, provider, now, "dcr", "permissions", "rbac")
	dcrFuture := runHelperOutput[models.DCROutput](group, ctx, request, dcrHandler(provider, now), "dcr", expected)
	evidenceFutures := runFamilyEvidenceWithExpected(group, ctx, request, provider, now, expected)

	dcrOutput, dcrSource, err := dcrFuture.waitWithSource()
	if err != nil {
		return nil, err
	}
	evidence, err := evidenceFutures.wait()
	if err != nil {
		return nil, err
	}

	sinks := providers.MonitoringSinksFromDCRReferences(dcrOutput.DCRs)
	dcrs := make([]models.EvasionDCR, 0, len(dcrOutput.DCRs))
	for _, dcr := range dcrOutput.DCRs {
		ruleControl, ruleControlOK := evasionDCRRuleControl(dcr.ID, evidence.principal.currentIdentityAssignments)
		associationControl, associationControlOK := evasionDCRAssociationControl(dcr, evidence.principal.currentIdentityAssignments)
		currentContext := evasionCurrentIdentityContext(evidence.principal.currentIdentity, ruleControl, ruleControlOK, associationControl, associationControlOK)
		state := evasionDCRState(dcr)
		rank, reason := evasionDCRDisruptionRank(dcr, ruleControlOK, associationControlOK)
		dcrs = append(dcrs, models.EvasionDCR{
			ID:                     dcr.ID,
			Name:                   dcr.Name,
			ResourceGroup:          dcr.ResourceGroup,
			Location:               dcr.Location,
			DisruptionRank:         rank,
			DisruptionReason:       reason,
			CapabilitySteps:        evasionDCRCapabilitySteps(ruleControlOK, associationControlOK),
			CurrentIdentityContext: currentContext,
			CurrentState:           state,
			NotCollectedByDefault:  evasionDCRNotCollectedByDefault(),
			Summary:                evasionDCRSummary(dcr, rank, ruleControlOK, associationControlOK),
			RelatedIDs:             mergeRelatedIDs(dcr.RelatedIDs),
		})
	}
	sort.SliceStable(dcrs, func(i, j int) bool {
		if dcrs[i].DisruptionRank != dcrs[j].DisruptionRank {
			return dcrs[i].DisruptionRank > dcrs[j].DisruptionRank
		}
		if dcrs[i].CurrentState.TransformationCount != dcrs[j].CurrentState.TransformationCount {
			return dcrs[i].CurrentState.TransformationCount > dcrs[j].CurrentState.TransformationCount
		}
		return dcrs[i].Name < dcrs[j].Name
	})

	issues := familyIssues(dcrOutput.Issues, evidence)

	return models.EvasionDCROutput{
		Metadata: withSessionArtifacts(
			scopedMetadata(
				now,
				request,
				firstNonEmpty(request.Tenant, stringPtrValue(dcrOutput.Metadata.TenantID), stringPtrValue(evidence.permissions.Metadata.TenantID)),
				firstNonEmpty(request.Subscription, stringPtrValue(dcrOutput.Metadata.SubscriptionID), stringPtrValue(evidence.permissions.Metadata.SubscriptionID)),
				"evasion",
			),
			appendSessionArtifact(evidence.sessionArtifacts, dcrSource),
		),
		GroupedCommandName: "evasion",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		MonitoringSinks:    sinks,
		DCRs:               dcrs,
		Issues:             issues,
	}, nil
}

func evasionDCRRuleControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	return evasionDCRBestControl(resourceID, assignments, "Microsoft.Insights/dataCollectionRules/write")
}

func evasionDCRAssociationControl(dcr models.DCRAsset, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	bestRank := 99
	best := persistenceCurrentIdentityControl{}
	targets := []string{dcr.ID}
	for _, association := range dcr.Associations {
		targets = append(targets, association.TargetID, association.ID)
	}
	for _, targetID := range targets {
		control, ok := evasionDCRBestControl(targetID, assignments, "Microsoft.Insights/dataCollectionRuleAssociations/write")
		if !ok {
			continue
		}
		rank, rankOK := persistenceScopeRank(control.ScopeID, targetID)
		if !rankOK || rank >= bestRank {
			continue
		}
		bestRank = rank
		best = control
	}
	return best, bestRank != 99
}

func evasionDCRBestControl(resourceID string, assignments []models.RoleAssignment, action string) (persistenceCurrentIdentityControl, bool) {
	bestRank := 99
	best := persistenceCurrentIdentityControl{}
	for _, assignment := range assignments {
		if !persistenceRoleAssignmentAllowsNamedOrActionControl(assignment, action, "Owner", "Contributor", "Monitoring Contributor") {
			continue
		}
		rank, ok := persistenceScopeRank(assignment.ScopeID, resourceID)
		if !ok || rank >= bestRank {
			continue
		}
		bestRank = rank
		best = persistenceCurrentIdentityControl{
			RoleName: fmt.Sprintf("%s at %s", assignment.RoleName, persistenceScopeLabel(assignment.ScopeID)),
			ScopeID:  assignment.ScopeID,
		}
	}
	return best, bestRank != 99
}

func evasionCurrentIdentityContext(
	currentIdentity models.PermissionRow,
	ruleControl persistenceCurrentIdentityControl,
	ruleControlOK bool,
	associationControl persistenceCurrentIdentityControl,
	associationControlOK bool,
) *models.EvasionRoleContext {
	if strings.TrimSpace(currentIdentity.DisplayName) == "" && !ruleControlOK && !associationControlOK {
		return nil
	}

	name := firstNonEmpty(currentIdentity.DisplayName, "current identity")
	roleNames := append([]string{}, currentIdentity.HighImpactRoles...)
	if len(roleNames) == 0 {
		roleNames = append(roleNames, currentIdentity.AllRoleNames...)
	}
	scopeIDs := append([]string{}, currentIdentity.ScopeIDs...)
	summary := "Current foothold identity is visible, but DCR or association write control is not proven here."
	controlLabel := "not proven"
	if ruleControlOK && associationControlOK {
		summary = fmt.Sprintf("Current foothold `%s` has visible DCR write control and association write control.", name)
		roleNames = []string{evasionControlRoleName(ruleControl), evasionControlRoleName(associationControl)}
		scopeIDs = []string{ruleControl.ScopeID, associationControl.ScopeID}
		controlLabel = "DCR + association write"
	} else if ruleControlOK {
		summary = fmt.Sprintf("Current foothold `%s` has visible DCR write control; association write control is not proven from the current evidence.", name)
		roleNames = []string{evasionControlRoleName(ruleControl)}
		scopeIDs = []string{ruleControl.ScopeID}
		controlLabel = "DCR write"
	} else if associationControlOK {
		summary = fmt.Sprintf("Current foothold `%s` has visible association write control; DCR content write control is not proven from the current evidence.", name)
		roleNames = []string{evasionControlRoleName(associationControl)}
		scopeIDs = []string{associationControl.ScopeID}
		controlLabel = "association write"
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

func evasionControlRoleName(control persistenceCurrentIdentityControl) string {
	return strings.TrimSpace(strings.SplitN(control.RoleName, " at ", 2)[0])
}

func evasionDCRCapabilitySteps(ruleControlOK bool, associationControlOK bool) []models.EvasionCapabilityStep {
	steps := make([]models.EvasionCapabilityStep, 0, len(evasionDCRSteps))
	for _, step := range evasionDCRSteps {
		status := "visible posture only"
		canAct := false
		if step.NeedsRuleWrite && step.NeedsAssociation {
			if ruleControlOK && associationControlOK {
				status = "yes"
				canAct = true
			} else if ruleControlOK || associationControlOK {
				status = "partial"
				canAct = true
			} else {
				status = "not proven"
			}
		} else if step.NeedsRuleWrite {
			if ruleControlOK {
				status = "yes"
				canAct = true
			} else {
				status = "not proven"
			}
		} else if step.NeedsAssociation {
			if associationControlOK {
				status = "yes"
				canAct = true
			} else {
				status = "not proven"
			}
		}
		steps = append(steps, models.EvasionCapabilityStep{
			Action:           step.Action,
			APISurface:       step.APISurface,
			Status:           status,
			CanAct:           canAct,
			DownstreamEffect: step.DownstreamEffect,
			Boundary:         step.Boundary,
		})
	}
	return steps
}

func evasionDCRState(dcr models.DCRAsset) models.EvasionDCRState {
	return models.EvasionDCRState{
		DataSourceTypes:       append([]string{}, dcr.DataSourceTypes...),
		Streams:               append([]string{}, dcr.Streams...),
		HighSignalStreams:     append([]string{}, dcr.HighSignalStreams...),
		DestinationTypes:      append([]string{}, dcr.DestinationTypes...),
		AssociationTargets:    evasionDCRAssociationTargets(dcr),
		TransformationCount:   dcr.TransformationCount,
		AssociationCount:      dcr.AssociationCount,
		TransformationPosture: evasionDCRTransformationPosture(dcr),
		DestinationPosture:    evasionDCRDestinationPosture(dcr),
	}
}

func evasionDCRAssociationTargets(dcr models.DCRAsset) []string {
	targets := make([]string, 0, len(dcr.Associations))
	for _, association := range dcr.Associations {
		if association.TargetID != "" {
			targets = append(targets, association.TargetID)
		}
	}
	sort.Strings(targets)
	return targets
}

func evasionDCRTransformationPosture(dcr models.DCRAsset) string {
	if dcr.TransformationCount == 0 {
		return "no transformation posture visible"
	}
	if len(dcr.HighSignalStreams) > 0 {
		return "transformation posture is visible on a DCR with high-signal streams"
	}
	return "transformation posture is visible, but stream impact needs monitoring context"
}

func evasionDCRDestinationPosture(dcr models.DCRAsset) string {
	if len(dcr.Destinations) == 0 {
		return "no destination visible"
	}
	return "operator-selected destinations visible: " + strings.Join(dcr.DestinationTypes, ", ")
}

func evasionDCRDisruptionRank(dcr models.DCRAsset, ruleControlOK bool, associationControlOK bool) (int, string) {
	rank := 1
	reasons := []string{}
	if dcr.TransformationCount > 0 && len(dcr.HighSignalStreams) > 0 {
		rank = 5
		reasons = append(reasons, "transformations can alter selected high-signal data while collection remains configured")
	} else if dcr.TransformationCount > 0 {
		rank = 4
		reasons = append(reasons, "transformations can filter or reshape collected records before storage")
	} else if len(dcr.HighSignalStreams) > 0 {
		rank = 3
		reasons = append(reasons, "high-signal streams are visible collection levers")
	} else if len(dcr.Destinations) > 0 {
		rank = 2
		reasons = append(reasons, "destinations can move collected data without disabling collection")
	}
	if ruleControlOK && associationControlOK {
		reasons = append(reasons, "current identity has visible rule and association write control")
	} else if ruleControlOK {
		reasons = append(reasons, "current identity has visible rule write control")
	} else if associationControlOK {
		reasons = append(reasons, "current identity has visible association write control")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "visible posture does not support a stronger dynamic disruption ranking")
	}
	return rank, strings.Join(reasons, "; ")
}

func evasionDCRNotCollectedByDefault() []models.EvasionBoundaryNote {
	return []models.EvasionBoundaryNote{
		{
			Name:           "transformKql body",
			Classification: "operational anomaly",
			Reason:         "Presence, length, and fingerprint are enough for posture; printing full transform logic can expose sensitive filtering logic and encourages overclaiming intent.",
		},
		{
			Name:           "log arrival or missing-record proof",
			Classification: "proof boundary",
			Reason:         "Management-plane DCR posture cannot prove which records arrived, were dropped, or were later queried from Log Analytics.",
		},
		{
			Name:           "agent applied-state",
			Classification: "product-model gap",
			Reason:         "DCR association shows intended scope, not that the Azure Monitor Agent applied the rule on the target at runtime.",
		},
		{
			Name:           "activity-log history and actor timing",
			Classification: "API/noise",
			Reason:         "Broad activity-log pulls can be noisy and are not required for the default posture view; use history only when sequencing is explicitly needed.",
		},
		{
			Name:           "expected SOC destination baseline",
			Classification: "scope/sequencing",
			Reason:         "Destination drift needs a defended expected-workspace model before the tool can say the current destination is wrong.",
		},
	}
}

func evasionDCRSummary(dcr models.DCRAsset, rank int, ruleControlOK bool, associationControlOK bool) string {
	parts := []string{fmt.Sprintf("DCR %q ranks %d/5 for quiet truth-disruption posture", dcr.Name, rank)}
	if dcr.TransformationCount > 0 {
		parts = append(parts, fmt.Sprintf("%d transformation clue(s)", dcr.TransformationCount))
	}
	if len(dcr.HighSignalStreams) > 0 {
		parts = append(parts, "high-signal streams: "+strings.Join(dcr.HighSignalStreams, ", "))
	}
	if len(dcr.DestinationTypes) > 0 {
		parts = append(parts, "destinations: "+strings.Join(dcr.DestinationTypes, ", "))
	}
	if ruleControlOK && associationControlOK {
		parts = append(parts, "current identity can affect both rule content and association scope from visible RBAC")
	} else if ruleControlOK {
		parts = append(parts, "current identity can affect rule content from visible RBAC")
	} else if associationControlOK {
		parts = append(parts, "current identity can affect association scope from visible RBAC")
	} else {
		parts = append(parts, "current identity write control is not proven")
	}
	return strings.Join(parts, "; ") + "."
}
