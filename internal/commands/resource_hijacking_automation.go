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

var resourceHijackingAutomationSteps = []familyStepDefinition{
	{Action: "select trusted automation account", APISurface: "Microsoft.Automation/automationAccounts", DownstreamEffect: "Keeps the existing operations automation account, modules, assets, and expected maintenance context in place.", Boundary: "Account posture does not prove any job ran."},
	{Action: "edit published runbook", APISurface: "Microsoft.Automation/automationAccounts/runbooks", NeedsWrite: true, DownstreamEffect: "Can change script logic that operators already expect Azure Automation to run.", Boundary: "Default output does not print runbook script content."},
	{Action: "reuse schedule or webhook trigger", APISurface: "job schedules and webhooks", NeedsWrite: true, DownstreamEffect: "Can preserve the existing invocation path while changing what the runbook does.", Boundary: "Trigger posture does not prove invocation."},
	{Action: "reuse automation identity or worker context", APISurface: "automation account identity and hybrid workers", NeedsWrite: true, DownstreamEffect: "Can run altered automation through already-integrated identity, worker, or secure-asset context.", Boundary: "Host state, job output, and secure asset values are not collected."},
	{Action: "blend as maintenance automation", APISurface: "automation account configuration", DownstreamEffect: "Normal cover stories include patching, remediation, cleanup, and scheduled maintenance script updates.", Boundary: "Cover story is not an intent claim."},
}

func buildResourceHijackingAutomationOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.ResourceHijackingSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	automationFuture := runGroupedCommandOutput[models.AutomationOutput](group, ctx, request, automationHandler(provider, now), "automation")
	evidenceFutures := runFamilyEvidence(group, ctx, request, provider, now)

	automation, err := automationFuture.wait()
	if err != nil {
		return nil, err
	}
	evidence, err := evidenceFutures.wait()
	if err != nil {
		return nil, err
	}

	targets := make([]models.ResourceHijackingAutomationTarget, 0, len(automation.AutomationAccounts))
	for _, account := range automation.AutomationAccounts {
		control, controlOK := persistenceAutomationControl(account.ID, evidence.principal.currentIdentityAssignments)
		rank, reason := resourceHijackingAutomationTakeoverRank(account, controlOK)
		targets = append(targets, models.ResourceHijackingAutomationTarget{
			ID:                     account.ID,
			Name:                   account.Name,
			ResourceGroup:          account.ResourceGroup,
			Location:               account.Location,
			TakeoverRank:           rank,
			TakeoverReason:         reason,
			CapabilitySteps:        resourceHijackingAutomationCapabilitySteps(controlOK),
			CurrentIdentityContext: resourceHijackingRoleContext(evidence.principal.currentIdentity, control, controlOK, "Automation account or runbook write control", "automation write"),
			CurrentState:           resourceHijackingAutomationState(account),
			NotCollectedByDefault:  resourceHijackingAutomationNotCollectedByDefault(),
			Summary:                resourceHijackingAutomationSummary(account, rank, controlOK),
			RelatedIDs:             mergeRelatedIDs(account.RelatedIDs),
		})
	}
	sort.SliceStable(targets, func(i, j int) bool {
		if targets[i].TakeoverRank != targets[j].TakeoverRank {
			return targets[i].TakeoverRank > targets[j].TakeoverRank
		}
		return targets[i].Name < targets[j].Name
	})

	issues := familyIssues(automation.Issues, evidence)

	return models.ResourceHijackingAutomationOutput{
		Metadata: scopedMetadata(
			now,
			request,
			firstNonEmpty(request.Tenant, stringPtrValue(automation.Metadata.TenantID), stringPtrValue(evidence.permissions.Metadata.TenantID)),
			firstNonEmpty(request.Subscription, stringPtrValue(automation.Metadata.SubscriptionID), stringPtrValue(evidence.permissions.Metadata.SubscriptionID)),
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

func resourceHijackingAutomationCapabilitySteps(controlOK bool) []models.ResourceHijackingCapabilityStep {
	return familyCapabilitySteps(resourceHijackingAutomationSteps, controlOK)
}

func resourceHijackingAutomationState(account models.AutomationAccountAsset) models.ResourceHijackingAutomationState {
	return models.ResourceHijackingAutomationState{
		State:                  account.State,
		IdentityType:           account.IdentityType,
		PublishedRunbookCount:  account.PublishedRunbookCount,
		PublishedRunbookNames:  append([]string{}, account.PublishedRunbookNames...),
		RunbookTypes:           append([]string{}, account.RunbookTypes...),
		RunbookCommandClues:    append([]string{}, account.RunbookCommandClues...),
		RunbookResourceClues:   append([]string{}, account.RunbookResourceClues...),
		ScheduleCount:          account.ScheduleCount,
		JobScheduleCount:       account.JobScheduleCount,
		WebhookCount:           account.WebhookCount,
		HybridWorkerGroupCount: account.HybridWorkerGroupCount,
		PrimaryStartMode:       account.PrimaryStartMode,
		PrimaryRunbookName:     account.PrimaryRunbookName,
		ScheduleRunbookNames:   append([]string{}, account.ScheduleRunbookNames...),
		WebhookRunbookNames:    append([]string{}, account.WebhookRunbookNames...),
		ConsequenceTypes:       append([]string{}, account.ConsequenceTypes...),
		Posture:                resourceHijackingAutomationPosture(account),
	}
}

func resourceHijackingAutomationPosture(account models.AutomationAccountAsset) string {
	parts := []string{}
	if apiMgmtIntValue(account.PublishedRunbookCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d published runbook(s)", apiMgmtIntValue(account.PublishedRunbookCount)))
	}
	if len(account.RunbookTypes) > 0 {
		parts = append(parts, "runbook types "+strings.Join(account.RunbookTypes, ", "))
	}
	if len(account.RunbookCommandClues) > 0 {
		parts = append(parts, "command clues "+strings.Join(account.RunbookCommandClues, ", "))
	}
	if len(account.RunbookResourceClues) > 0 {
		parts = append(parts, "resource clues "+strings.Join(account.RunbookResourceClues, ", "))
	}
	if apiMgmtIntValue(account.JobScheduleCount) > 0 || apiMgmtIntValue(account.WebhookCount) > 0 {
		parts = append(parts, "schedule or webhook trigger posture")
	}
	if strings.TrimSpace(stringPtrValue(account.IdentityType)) != "" {
		parts = append(parts, "managed identity posture")
	}
	if apiMgmtIntValue(account.HybridWorkerGroupCount) > 0 {
		parts = append(parts, "hybrid worker posture")
	}
	if len(account.ConsequenceTypes) > 0 {
		parts = append(parts, "consequence types "+strings.Join(account.ConsequenceTypes, ", "))
	}
	if len(parts) == 0 {
		return "Automation account visible without stronger runbook, trigger, or identity posture"
	}
	return strings.Join(parts, "; ")
}

func resourceHijackingAutomationTakeoverRank(account models.AutomationAccountAsset, controlOK bool) (int, string) {
	rank := 1
	reasons := []string{}
	hasRunbook := apiMgmtIntValue(account.PublishedRunbookCount) > 0 || len(account.PublishedRunbookNames) > 0
	hasRunbookPosture := hasRunbook && (len(account.RunbookTypes) > 0 || len(account.RunbookCommandClues) > 0 || len(account.RunbookResourceClues) > 0)
	hasTrigger := apiMgmtIntValue(account.JobScheduleCount) > 0 || apiMgmtIntValue(account.WebhookCount) > 0
	hasIdentity := strings.TrimSpace(stringPtrValue(account.IdentityType)) != ""
	hasWorker := apiMgmtIntValue(account.HybridWorkerGroupCount) > 0
	switch {
	case hasRunbookPosture && hasTrigger && hasIdentity:
		rank = 5
		reasons = append(reasons, "published runbook posture, trigger posture, and automation identity are visible")
	case hasRunbookPosture && hasTrigger && hasWorker:
		rank = 4
		reasons = append(reasons, "published runbook posture, trigger posture, and hybrid worker context are visible")
	case hasRunbook && hasTrigger:
		rank = 3
		reasons = append(reasons, "published runbook and trigger posture are visible")
	case hasRunbook || hasTrigger:
		rank = 2
		reasons = append(reasons, "partial runbook or trigger posture is visible")
	}
	if controlOK {
		reasons = append(reasons, "current identity has visible Automation account or runbook write control")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "visible posture does not support a stronger dynamic takeover ranking")
	}
	return rank, strings.Join(reasons, "; ")
}

func resourceHijackingAutomationNotCollectedByDefault() []models.ResourceHijackingBoundaryNote {
	return []models.ResourceHijackingBoundaryNote{
		{Name: "runbook script content", Classification: "recon safety", Reason: "The live helper reports safe runbook type and trigger posture by default; content-derived command/resource clues require a narrower review path and raw script bodies are not printed."},
		{Name: "secure asset values", Classification: "recon safety", Reason: "Automation credentials, certificates, connections, and encrypted variables are not safe default output."},
		{Name: "job output and status history", Classification: "proof boundary", Reason: "Management-plane posture cannot prove a changed runbook executed or what it produced."},
		{Name: "hybrid worker host state", Classification: "proof boundary", Reason: "Automation account posture cannot prove guest-side host impact without worker or host evidence."},
		{Name: "activity history", Classification: "API/noise", Reason: "Broad Automation history pulls are not needed for default posture and should be a narrow follow-up for timing or actor proof."},
	}
}

func resourceHijackingAutomationSummary(account models.AutomationAccountAsset, rank int, controlOK bool) string {
	parts := []string{fmt.Sprintf("account %q ranks %d/5 for Automation resource-hijack posture", account.Name, rank)}
	if apiMgmtIntValue(account.PublishedRunbookCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d published runbook(s)", apiMgmtIntValue(account.PublishedRunbookCount)))
	}
	if apiMgmtIntValue(account.JobScheduleCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d job schedule(s)", apiMgmtIntValue(account.JobScheduleCount)))
	}
	if apiMgmtIntValue(account.WebhookCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d webhook(s)", apiMgmtIntValue(account.WebhookCount)))
	}
	if controlOK {
		parts = append(parts, "current identity can modify Automation posture from visible RBAC")
	} else {
		parts = append(parts, "current identity Automation write control is not proven")
	}
	return strings.Join(parts, "; ") + "."
}
