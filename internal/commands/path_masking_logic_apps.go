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

var pathMaskingLogicAppSteps = []familyStepDefinition{
	{Action: "select trusted workflow", APISurface: "Microsoft.Logic/workflows", DownstreamEffect: "Keeps the visible activity inside an existing Azure integration resource rather than a direct caller-to-target path.", Boundary: "Workflow posture does not prove the workflow ran."},
	{Action: "identify trigger entry point", APISurface: "request, recurrence, api-connection, or event trigger", DownstreamEffect: "Request, schedule, and connector triggers can make the workflow the front door instead of the operator or caller.", Boundary: "Trigger posture does not prove invocation or caller identity."},
	{Action: "map downstream relay actions", APISurface: "HTTP, api-connection, or service actions", DownstreamEffect: "Downstream actions show where the workflow can forward, transform, or broker activity through trusted connectors.", Boundary: "Default output does not print full workflow bodies or payloads."},
	{Action: "change workflow route", APISurface: "workflow definition", NeedsWrite: true, DownstreamEffect: "Can repoint HTTP actions, reshape branches, or preserve the trigger while changing the downstream path.", Boundary: "Write capability is inferred only from visible management-plane RBAC."},
	{Action: "blend as integration maintenance", APISurface: "workflow configuration", DownstreamEffect: "Normal cover stories include connector repair, retry tuning, endpoint migration, and integration modernization.", Boundary: "Cover story is not an intent claim."},
}

func buildPathMaskingLogicAppsOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.PathMaskingSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	expected := helperArtifactExpectedSessions(ctx, request, provider, now, "logic-apps", "permissions", "rbac")
	logicAppsFuture := runHelperOutput[models.LogicAppsOutput](group, ctx, request, logicAppsHandler(provider, now), "logic-apps", expected)
	evidenceFutures := runFamilyEvidenceWithExpected(group, ctx, request, provider, now, expected)

	logicApps, logicAppsSource, err := logicAppsFuture.waitWithSource()
	if err != nil {
		return nil, err
	}
	evidence, err := evidenceFutures.wait()
	if err != nil {
		return nil, err
	}

	targets := make([]models.PathMaskingLogicAppTarget, 0, len(logicApps.Workflows))
	for _, workflow := range logicApps.Workflows {
		control, controlOK := resourceHijackingLogicAppControl(workflow.ID, evidence.principal.currentIdentityAssignments)
		rank, reason := pathMaskingLogicAppRank(workflow, controlOK)
		targets = append(targets, models.PathMaskingLogicAppTarget{
			ID:                     workflow.ID,
			Name:                   workflow.Name,
			ResourceGroup:          workflow.ResourceGroup,
			Location:               workflow.Location,
			MaskingRank:            rank,
			MaskingReason:          reason,
			CapabilitySteps:        pathMaskingCapabilitySteps(pathMaskingLogicAppSteps, controlOK),
			CurrentIdentityContext: pathMaskingRoleContext(evidence.principal.currentIdentity, control, controlOK, "Logic App route or relay workflow write control", "workflow write"),
			CurrentState:           pathMaskingLogicAppState(workflow),
			NotCollectedByDefault:  pathMaskingLogicAppNotCollectedByDefault(),
			Summary:                pathMaskingLogicAppSummary(workflow, rank, controlOK),
			RelatedIDs:             mergeRelatedIDs(workflow.RelatedIDs),
		})
	}
	sort.SliceStable(targets, func(i, j int) bool {
		if targets[i].MaskingRank != targets[j].MaskingRank {
			return targets[i].MaskingRank > targets[j].MaskingRank
		}
		return targets[i].Name < targets[j].Name
	})

	issues := familyIssues(logicApps.Issues, evidence)

	return models.PathMaskingLogicAppsOutput{
		Metadata: withSessionArtifacts(
			scopedMetadata(
				now,
				request,
				firstNonEmpty(request.Tenant, stringPtrValue(logicApps.Metadata.TenantID), stringPtrValue(evidence.permissions.Metadata.TenantID)),
				firstNonEmpty(request.Subscription, stringPtrValue(logicApps.Metadata.SubscriptionID), stringPtrValue(evidence.permissions.Metadata.SubscriptionID)),
				"pathmasking",
			),
			appendSessionArtifact(evidence.sessionArtifacts, logicAppsSource),
		),
		GroupedCommandName: "pathmasking",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		Targets:            targets,
		Issues:             issues,
	}, nil
}

func pathMaskingLogicAppState(workflow models.LogicAppWorkflowAsset) models.PathMaskingLogicAppState {
	return familyLogicAppState(workflow, pathMaskingLogicAppPosture(workflow))
}

func pathMaskingLogicAppPosture(workflow models.LogicAppWorkflowAsset) string {
	return familyLogicAppPosture(workflow, "Logic App workflow visible without stronger trigger, downstream, or identity posture")
}

func pathMaskingLogicAppRank(workflow models.LogicAppWorkflowAsset, controlOK bool) (int, string) {
	rank := 1
	reasons := []string{}
	hasTrigger := workflow.ExternallyCallableRequestTrigger || len(workflow.TriggerTypes) > 0 || workflow.RecurrenceSummary != nil
	hasDownstream := len(workflow.DownstreamActionKinds) > 0 || len(workflow.DownstreamResourceReferences) > 0 || len(workflow.ConnectorReferences) > 0
	hasHTTPOrConnector := workflow.ExternallyCallableRequestTrigger || len(workflow.ConnectorReferences) > 0 || hasActionKind(workflow.DownstreamActionKinds, "http") || hasActionKind(workflow.DownstreamActionKinds, "api-connection")
	hasIdentity := strings.TrimSpace(stringPtrValue(workflow.IdentityType)) != ""
	switch {
	case hasHTTPOrConnector && hasDownstream && hasIdentity:
		rank = 5
		reasons = append(reasons, "request or connector path, downstream action posture, and workflow identity are visible")
	case hasTrigger && hasDownstream && hasIdentity:
		rank = 4
		reasons = append(reasons, "trigger, downstream action posture, and workflow identity are visible")
	case hasTrigger && hasDownstream:
		rank = 3
		reasons = append(reasons, "trigger and downstream action posture are visible")
	case hasTrigger || hasDownstream:
		rank = 2
		reasons = append(reasons, "partial trigger or downstream path posture is visible")
	}
	if controlOK {
		reasons = append(reasons, "current identity has visible Logic App route or relay workflow write control")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "visible posture does not support a stronger dynamic pathmasking ranking")
	}
	return rank, strings.Join(reasons, "; ")
}

func hasActionKind(values []string, needle string) bool {
	needle = strings.ToLower(strings.TrimSpace(needle))
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), needle) {
			return true
		}
	}
	return false
}

func pathMaskingLogicAppNotCollectedByDefault() []models.PathMaskingBoundaryNote {
	return []models.PathMaskingBoundaryNote{
		{Name: "full workflow definition body", Classification: "collector issue", Reason: "The helper reports trigger and downstream action posture but does not print complete workflow JSON by default."},
		{Name: "run history", Classification: "proof boundary", Reason: "Management-plane posture cannot prove the workflow ran, succeeded, or carried traffic."},
		{Name: "connector credential values", Classification: "recon safety", Reason: "Connector secrets and connection credential material are not safe default output."},
		{Name: "payload and response contents", Classification: "proof boundary", Reason: "The command does not inspect trigger payloads, action payloads, or response data."},
		{Name: "workflow change history", Classification: "API/noise", Reason: "Broad workflow history is not needed for default posture and should stay a narrow follow-up for timing or actor proof."},
	}
}

func pathMaskingLogicAppSummary(workflow models.LogicAppWorkflowAsset, rank int, controlOK bool) string {
	parts := []string{fmt.Sprintf("workflow %q ranks %d/5 for Logic Apps pathmasking posture", workflow.Name, rank)}
	if len(workflow.TriggerTypes) > 0 {
		parts = append(parts, fmt.Sprintf("%d trigger type(s)", len(workflow.TriggerTypes)))
	}
	if len(workflow.DownstreamActionKinds) > 0 {
		parts = append(parts, fmt.Sprintf("%d downstream action kind(s)", len(workflow.DownstreamActionKinds)))
	}
	if controlOK {
		parts = append(parts, "current identity can change workflow path posture from visible RBAC")
	} else {
		parts = append(parts, "current identity workflow write control is not proven")
	}
	return strings.Join(parts, "; ") + "."
}
