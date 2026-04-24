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

var resourceHijackingLogicAppSteps = []familyStepDefinition{
	{Action: "select trusted workflow", APISurface: "Microsoft.Logic/workflows", DownstreamEffect: "Keeps the existing automation resource, trigger path, and operational context in place.", Boundary: "Workflow posture does not prove the workflow ran."},
	{Action: "edit workflow definition", APISurface: "workflow definition", NeedsWrite: true, DownstreamEffect: "Can add, remove, or repurpose actions while the trusted Logic App remains the same resource.", Boundary: "Default output does not print full workflow definition bodies."},
	{Action: "repurpose trigger", APISurface: "request, recurrence, api-connection, or event trigger", NeedsWrite: true, DownstreamEffect: "Can keep an existing inbound, scheduled, or connector trigger while changing what happens after it fires.", Boundary: "Trigger posture does not prove trigger invocation."},
	{Action: "reuse connector or identity context", APISurface: "workflow identity and connection references", NeedsWrite: true, DownstreamEffect: "Can run altered workflow logic through already-integrated connectors or managed identity context.", Boundary: "Connector credential values and secret material are not collected."},
	{Action: "blend as integration maintenance", APISurface: "workflow configuration", DownstreamEffect: "Normal cover stories include connector refresh, integration repair, retry handling, or workflow modernization.", Boundary: "Cover story is not an intent claim."},
}

func buildResourceHijackingLogicAppsOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.ResourceHijackingSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	logicAppsFuture := runGroupedCommandOutput[models.LogicAppsOutput](group, ctx, request, logicAppsHandler(provider, now), "logic-apps")
	evidenceFutures := runFamilyEvidence(group, ctx, request, provider, now)

	logicApps, err := logicAppsFuture.wait()
	if err != nil {
		return nil, err
	}
	evidence, err := evidenceFutures.wait()
	if err != nil {
		return nil, err
	}

	targets := make([]models.ResourceHijackingLogicAppTarget, 0, len(logicApps.Workflows))
	for _, workflow := range logicApps.Workflows {
		control, controlOK := resourceHijackingLogicAppControl(workflow.ID, evidence.principal.currentIdentityAssignments)
		rank, reason := resourceHijackingLogicAppTakeoverRank(workflow, controlOK)
		targets = append(targets, models.ResourceHijackingLogicAppTarget{
			ID:                     workflow.ID,
			Name:                   workflow.Name,
			ResourceGroup:          workflow.ResourceGroup,
			Location:               workflow.Location,
			TakeoverRank:           rank,
			TakeoverReason:         reason,
			CapabilitySteps:        resourceHijackingLogicAppCapabilitySteps(controlOK),
			CurrentIdentityContext: resourceHijackingRoleContext(evidence.principal.currentIdentity, control, controlOK, "Logic App workflow write control", "workflow write"),
			CurrentState:           resourceHijackingLogicAppState(workflow),
			NotCollectedByDefault:  resourceHijackingLogicAppNotCollectedByDefault(),
			Summary:                resourceHijackingLogicAppSummary(workflow, rank, controlOK),
			RelatedIDs:             mergeRelatedIDs(workflow.RelatedIDs),
		})
	}
	sort.SliceStable(targets, func(i, j int) bool {
		if targets[i].TakeoverRank != targets[j].TakeoverRank {
			return targets[i].TakeoverRank > targets[j].TakeoverRank
		}
		return targets[i].Name < targets[j].Name
	})

	issues := familyIssues(logicApps.Issues, evidence)

	return models.ResourceHijackingLogicAppsOutput{
		Metadata: scopedMetadata(
			now,
			request,
			firstNonEmpty(request.Tenant, stringPtrValue(logicApps.Metadata.TenantID), stringPtrValue(evidence.permissions.Metadata.TenantID)),
			firstNonEmpty(request.Subscription, stringPtrValue(logicApps.Metadata.SubscriptionID), stringPtrValue(evidence.permissions.Metadata.SubscriptionID)),
			"resourcehijacking",
		),
		GroupedCommandName: "resourcehijacking",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		Targets:            targets,
		Issues:             issues,
	}, nil
}

func resourceHijackingLogicAppControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	return resourceHijackingBestControl(
		resourceID,
		assignments,
		[]string{"Microsoft.Logic/workflows/write"},
		"Owner",
		"Contributor",
		"Logic App Contributor",
	)
}

func resourceHijackingLogicAppCapabilitySteps(controlOK bool) []models.ResourceHijackingCapabilityStep {
	return familyCapabilitySteps(resourceHijackingLogicAppSteps, controlOK)
}

func resourceHijackingLogicAppState(workflow models.LogicAppWorkflowAsset) models.ResourceHijackingLogicAppState {
	return familyLogicAppState(workflow, resourceHijackingLogicAppPosture(workflow))
}

func resourceHijackingLogicAppPosture(workflow models.LogicAppWorkflowAsset) string {
	return familyLogicAppPosture(workflow, "Logic App workflow visible without stronger trigger, action, or identity posture")
}

func resourceHijackingLogicAppTakeoverRank(workflow models.LogicAppWorkflowAsset, controlOK bool) (int, string) {
	rank := 1
	reasons := []string{}
	hasTrigger := workflow.ExternallyCallableRequestTrigger || len(workflow.TriggerTypes) > 0 || workflow.RecurrenceSummary != nil
	hasDownstream := len(workflow.DownstreamActionKinds) > 0 || len(workflow.DownstreamResourceReferences) > 0 || len(workflow.ConnectorReferences) > 0
	hasIdentity := strings.TrimSpace(stringPtrValue(workflow.IdentityType)) != ""
	switch {
	case workflow.ExternallyCallableRequestTrigger && hasDownstream && hasIdentity:
		rank = 5
		reasons = append(reasons, "external trigger, downstream action posture, and workflow identity are visible")
	case hasTrigger && hasDownstream && hasIdentity:
		rank = 4
		reasons = append(reasons, "trigger, downstream action posture, and workflow identity are visible")
	case hasTrigger && hasDownstream:
		rank = 3
		reasons = append(reasons, "trigger and downstream action posture are visible")
	case hasTrigger || hasDownstream:
		rank = 2
		reasons = append(reasons, "partial trigger or downstream action posture is visible")
	}
	if controlOK {
		reasons = append(reasons, "current identity has visible Logic App workflow write control")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "visible posture does not support a stronger dynamic takeover ranking")
	}
	return rank, strings.Join(reasons, "; ")
}

func resourceHijackingLogicAppNotCollectedByDefault() []models.ResourceHijackingBoundaryNote {
	return []models.ResourceHijackingBoundaryNote{
		{Name: "full workflow definition body", Classification: "collector issue", Reason: "The helper reports trigger and downstream action posture but does not print complete workflow JSON by default."},
		{Name: "connector credential values", Classification: "recon safety", Reason: "Connector secrets and connection credential material are not safe default output."},
		{Name: "run history", Classification: "proof boundary", Reason: "Management-plane posture cannot prove the modified workflow ran or completed downstream actions."},
		{Name: "data handled by actions", Classification: "proof boundary", Reason: "The command does not inspect connector or action payload content."},
		{Name: "activity history", Classification: "API/noise", Reason: "Broad workflow change history is not needed for default posture and should be a narrow follow-up for timing or actor proof."},
	}
}

func resourceHijackingLogicAppSummary(workflow models.LogicAppWorkflowAsset, rank int, controlOK bool) string {
	parts := []string{fmt.Sprintf("workflow %q ranks %d/5 for Logic App resource-hijack posture", workflow.Name, rank)}
	if len(workflow.TriggerTypes) > 0 {
		parts = append(parts, fmt.Sprintf("%d trigger type(s)", len(workflow.TriggerTypes)))
	}
	if len(workflow.DownstreamActionKinds) > 0 {
		parts = append(parts, fmt.Sprintf("%d downstream action kind(s)", len(workflow.DownstreamActionKinds)))
	}
	if controlOK {
		parts = append(parts, "current identity can modify workflow posture from visible RBAC")
	} else {
		parts = append(parts, "current identity workflow write control is not proven")
	}
	return strings.Join(parts, "; ") + "."
}
