package commands

import (
	"context"
	"strconv"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func logicAppsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.LogicApps(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		workflows := sortedByLess(facts.Workflows, logicAppLess)
		for idx := range workflows {
			workflows[idx] = decorateLogicAppArtifact(workflows[idx])
		}

		return models.LogicAppsOutput{
			Findings:  []models.Finding{},
			Issues:    facts.Issues,
			Metadata:  runtimeCommandMetadata("logic-apps", now, facts.TenantID, facts.SubscriptionID),
			Workflows: workflows,
		}, nil
	}
}

func logicAppLess(left models.LogicAppWorkflowAsset, right models.LogicAppWorkflowAsset) bool {
	leftClass := logicAppClassificationRank(left.Classification)
	rightClass := logicAppClassificationRank(right.Classification)
	if leftClass != rightClass {
		return leftClass < rightClass
	}

	if left.ExternallyCallableRequestTrigger != right.ExternallyCallableRequestTrigger {
		return left.ExternallyCallableRequestTrigger
	}

	leftRecurring := left.RecurrenceSummary != nil && *left.RecurrenceSummary != ""
	rightRecurring := right.RecurrenceSummary != nil && *right.RecurrenceSummary != ""
	if leftRecurring != rightRecurring {
		return leftRecurring
	}

	leftIdentity := left.IdentityType != nil && *left.IdentityType != ""
	rightIdentity := right.IdentityType != nil && *right.IdentityType != ""
	if leftIdentity != rightIdentity {
		return leftIdentity
	}

	if len(left.DownstreamActionKinds) != len(right.DownstreamActionKinds) {
		return len(left.DownstreamActionKinds) > len(right.DownstreamActionKinds)
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func logicAppClassificationRank(classification string) int {
	switch classification {
	case "persistence-capable":
		return 0
	case "execution-capable-only":
		return 1
	default:
		return 2
	}
}

func decorateLogicAppArtifact(workflow models.LogicAppWorkflowAsset) models.LogicAppWorkflowAsset {
	workflow.Trigger = compactArtifactValue(logicAppArtifactTrigger(workflow))
	workflow.Identity = compactArtifactValue(logicAppArtifactIdentity(workflow))
	workflow.Downstream = compactArtifactValue(strings.Join(workflow.DownstreamActionKinds, "; "))
	return workflow
}

func logicAppArtifactTrigger(workflow models.LogicAppWorkflowAsset) string {
	parts := []string{}
	seen := map[string]struct{}{}
	if workflow.ExternallyCallableRequestTrigger {
		parts = append(parts, "request(external)")
		seen["request"] = struct{}{}
	}
	if workflow.RecurrenceSummary != nil && *workflow.RecurrenceSummary != "" {
		parts = append(parts, "recurrence")
		seen["recurrence"] = struct{}{}
	}
	for _, triggerType := range workflow.TriggerTypes {
		if _, ok := seen[triggerType]; ok {
			continue
		}
		parts = append(parts, triggerType)
		seen[triggerType] = struct{}{}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func logicAppArtifactIdentity(workflow models.LogicAppWorkflowAsset) string {
	parts := []string{}
	if workflow.IdentityType != nil && *workflow.IdentityType != "" {
		parts = append(parts, *workflow.IdentityType)
	}
	if len(workflow.IdentityIDs) > 0 {
		userAssignedCount := len(workflow.IdentityIDs)
		if strings.Contains(strings.ToLower(stringPtrValue(workflow.IdentityType)), "systemassigned") && userAssignedCount > 0 {
			userAssignedCount--
		}
		if userAssignedCount > 0 {
			parts = append(parts, "user-assigned="+strconv.Itoa(userAssignedCount))
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}
