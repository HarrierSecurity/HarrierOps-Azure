package commands

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func armDeploymentsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.ArmDeployments(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		deployments := append([]models.ArmDeploymentSummary{}, facts.Deployments...)
		sort.SliceStable(deployments, func(i int, j int) bool {
			return armDeploymentLess(deployments[i], deployments[j])
		})

		return models.ArmDeploymentsOutput{
			Deployments: deployments,
			Findings:    buildArmDeploymentFindings(deployments),
			Issues:      facts.Issues,
			Metadata:    commandMetadata("arm-deployments", now, request, facts.TenantID, facts.SubscriptionID, ""),
		}, nil
	}
}

func buildArmDeploymentFindings(deployments []models.ArmDeploymentSummary) []models.ArmDeploymentFinding {
	findings := make([]models.ArmDeploymentFinding, 0, len(deployments))
	for _, deployment := range deployments {
		state := strings.ToLower(deployment.ProvisioningState)

		if state == "failed" || state == "canceled" {
			findings = append(findings, models.ArmDeploymentFinding{
				ID:          "arm-deployment-failed-" + deployment.ID,
				Severity:    "medium",
				Title:       "Deployment did not complete successfully",
				Description: "Deployment '" + deployment.Name + "' ended in state '" + valueOrUnknown(deployment.ProvisioningState) + "'. Review the deployment history for leaked config context, partial resource creation, or operator troubleshooting artifacts.",
				RelatedIDs:  []string{deployment.ID},
			})
		}

		if deployment.OutputsCount > 0 {
			findings = append(findings, models.ArmDeploymentFinding{
				ID:          "arm-deployment-outputs-" + deployment.ID,
				Severity:    "medium",
				Title:       "Deployment exposes output values",
				Description: "Deployment '" + deployment.Name + "' includes " + strconv.Itoa(deployment.OutputsCount) + " recorded output values. Validate whether any outputs reveal useful endpoints, identifiers, or sensitive configuration.",
				RelatedIDs:  []string{deployment.ID},
			})
		}

		if deployment.TemplateLink != nil || deployment.ParametersLink != nil {
			findings = append(findings, models.ArmDeploymentFinding{
				ID:          "arm-deployment-remote-link-" + deployment.ID,
				Severity:    "low",
				Title:       "Deployment references linked template content",
				Description: "Deployment '" + deployment.Name + "' uses linked template or parameter content. Review those linked artifacts for exposed configuration, trust assumptions, or reusable infrastructure patterns.",
				RelatedIDs:  []string{deployment.ID},
			})
		}
	}
	return findings
}

func armDeploymentLess(left models.ArmDeploymentSummary, right models.ArmDeploymentSummary) bool {
	leftLinkCount := 0
	if left.TemplateLink != nil {
		leftLinkCount++
	}
	if left.ParametersLink != nil {
		leftLinkCount++
	}
	rightLinkCount := 0
	if right.TemplateLink != nil {
		rightLinkCount++
	}
	if right.ParametersLink != nil {
		rightLinkCount++
	}

	if armDeploymentStateRank(left.ProvisioningState) != armDeploymentStateRank(right.ProvisioningState) {
		return armDeploymentStateRank(left.ProvisioningState) < armDeploymentStateRank(right.ProvisioningState)
	}
	if leftLinkCount != rightLinkCount {
		return leftLinkCount > rightLinkCount
	}
	if left.OutputsCount != right.OutputsCount {
		return left.OutputsCount > right.OutputsCount
	}

	leftResourceSignal := max(left.OutputResourceCount, len(left.Providers))
	rightResourceSignal := max(right.OutputResourceCount, len(right.Providers))
	if leftResourceSignal != rightResourceSignal {
		return leftResourceSignal > rightResourceSignal
	}
	if (left.ScopeType == "subscription") != (right.ScopeType == "subscription") {
		return left.ScopeType != "subscription"
	}
	leftResourceGroup := ""
	if left.ResourceGroup != nil {
		leftResourceGroup = *left.ResourceGroup
	}
	rightResourceGroup := ""
	if right.ResourceGroup != nil {
		rightResourceGroup = *right.ResourceGroup
	}
	if leftResourceGroup != rightResourceGroup {
		return leftResourceGroup < rightResourceGroup
	}
	return left.Name < right.Name
}

func armDeploymentStateRank(value string) int {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch {
	case normalized == "failed":
		return 0
	case normalized != "" && normalized != "succeeded":
		return 1
	default:
		return 2
	}
}

func valueOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}
