package commands

import (
	"context"
	"strings"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func authPoliciesHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.AuthPolicies(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		findings := authPolicyFindings(facts.AuthPolicies, facts.Issues)
		policies := sortedByLess(facts.AuthPolicies, func(left, right models.AuthPolicySummary) bool {
			return authPolicyLess(left, right, findings, facts.Issues)
		})

		return models.AuthPoliciesOutput{
			AuthPolicies: policies,
			Findings:     findings,
			Issues:       facts.Issues,
			Metadata: models.AuthPoliciesMetadata{
				Command:        "auth-policies",
				GeneratedAt:    now().UTC().Format(time.RFC3339),
				SchemaVersion:  contracts.AzureFoxSchemaVersion,
				SubscriptionID: models.StringPtr(facts.SubscriptionID),
				TenantID:       models.StringPtr(facts.TenantID),
				TokenSource:    nil,
			},
		}, nil
	}
}

func authPolicyFindings(policies []models.AuthPolicySummary, issues []models.Issue) []models.AuthPolicyFinding {
	findings := []models.AuthPolicyFinding{}

	var securityDefaults *models.AuthPolicySummary
	var authorizationPolicy *models.AuthPolicySummary
	conditionalAccess := []models.AuthPolicySummary{}
	for i := range policies {
		policy := &policies[i]
		switch policy.PolicyType {
		case "security-defaults":
			securityDefaults = policy
		case "authorization-policy":
			authorizationPolicy = policy
		case "conditional-access":
			conditionalAccess = append(conditionalAccess, *policy)
		}
	}

	conditionalAccessUnreadable := false
	for _, issue := range issues {
		if issue.Context["collector"] == "auth_policies.conditional_access" {
			conditionalAccessUnreadable = true
			break
		}
	}

	if securityDefaults != nil && securityDefaults.State == "disabled" {
		findings = append(findings, models.AuthPolicyFinding{
			ID:          "auth-policy-security-defaults-disabled",
			Severity:    "medium",
			Title:       "Security defaults are disabled",
			Description: "Tenant-wide security defaults are disabled. Review whether Conditional Access policies provide equivalent baseline MFA and legacy-auth controls.",
			RelatedIDs:  append([]string{}, securityDefaults.RelatedIDs...),
		})
	}

	if authorizationPolicy != nil {
		controls := map[string]struct{}{}
		for _, control := range authorizationPolicy.Controls {
			controls[control] = struct{}{}
		}

		if _, ok := controls["users-can-register-apps"]; ok {
			findings = append(findings, models.AuthPolicyFinding{
				ID:          "auth-policy-users-can-register-apps",
				Severity:    "medium",
				Title:       "Users can register applications",
				Description: "Default user permissions allow application registration. Review whether that app-creation surface is expected for this tenant.",
				RelatedIDs:  append([]string{}, authorizationPolicy.RelatedIDs...),
			})
		}
		if _, ok := controls["guest-invites:everyone"]; ok {
			findings = append(findings, models.AuthPolicyFinding{
				ID:          "auth-policy-guest-invites-everyone",
				Severity:    "medium",
				Title:       "Guest invitations are broadly allowed",
				Description: "The authorization policy allows guest invitations from everyone in the tenant. Validate whether that guest-invite surface is intentional.",
				RelatedIDs:  append([]string{}, authorizationPolicy.RelatedIDs...),
			})
		}
		if _, ok := controls["risky-app-consent:enabled"]; ok {
			findings = append(findings, models.AuthPolicyFinding{
				ID:          "auth-policy-risky-app-consent-enabled",
				Severity:    "high",
				Title:       "Risky app consent is enabled",
				Description: "Authorization policy allows user consent for risky apps. Review whether that consent posture is expected for this tenant.",
				RelatedIDs:  append([]string{}, authorizationPolicy.RelatedIDs...),
			})
		} else if _, ok := controls["user-consent:self-service"]; ok {
			findings = append(findings, models.AuthPolicyFinding{
				ID:          "auth-policy-user-consent-enabled",
				Severity:    "medium",
				Title:       "User consent is available to default users",
				Description: "Default user permissions include self-service permission-grant policy assignment. Review which delegated or application access paths that enables.",
				RelatedIDs:  append([]string{}, authorizationPolicy.RelatedIDs...),
			})
		}
	}

	enabledCA := 0
	for _, policy := range conditionalAccess {
		if strings.EqualFold(policy.State, "enabled") {
			enabledCA++
		}
	}
	if securityDefaults != nil && securityDefaults.State == "disabled" && enabledCA == 0 && !conditionalAccessUnreadable {
		findings = append(findings, models.AuthPolicyFinding{
			ID:          "auth-policy-no-active-enforcement-visible",
			Severity:    "medium",
			Title:       "No active auth enforcement visible",
			Description: "Security defaults are disabled and no enabled Conditional Access policies are visible from the current read path. Validate whether stronger auth controls exist outside the currently readable policy surface.",
			RelatedIDs:  append([]string{}, securityDefaults.RelatedIDs...),
		})
	}

	return findings
}

func authPolicyLess(left, right models.AuthPolicySummary, findings []models.AuthPolicyFinding, issues []models.Issue) bool {
	leftStateA, leftStateB := authPolicyStateRank(left)
	rightStateA, rightStateB := authPolicyStateRank(right)

	leftHasFinding := authPolicyHasFinding(left, findings)
	rightHasFinding := authPolicyHasFinding(right, findings)
	if leftHasFinding != rightHasFinding {
		return leftHasFinding
	}

	leftHasIssue := authPolicyHasIssue(left, issues)
	rightHasIssue := authPolicyHasIssue(right, issues)
	if leftHasIssue != rightHasIssue {
		return leftHasIssue
	}

	if leftStateA != rightStateA {
		return leftStateA < rightStateA
	}
	if leftStateB != rightStateB {
		return leftStateB < rightStateB
	}

	leftSecurityDefaults := left.PolicyType == "security-defaults"
	rightSecurityDefaults := right.PolicyType == "security-defaults"
	if leftSecurityDefaults != rightSecurityDefaults {
		return leftSecurityDefaults
	}

	leftAuthorization := left.PolicyType == "authorization-policy"
	rightAuthorization := right.PolicyType == "authorization-policy"
	if leftAuthorization != rightAuthorization {
		return leftAuthorization
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.PolicyType < right.PolicyType
}

func authPolicyHasFinding(policy models.AuthPolicySummary, findings []models.AuthPolicyFinding) bool {
	if len(policy.RelatedIDs) == 0 {
		return false
	}
	related := map[string]struct{}{}
	for _, id := range policy.RelatedIDs {
		related[id] = struct{}{}
	}
	for _, finding := range findings {
		for _, id := range finding.RelatedIDs {
			if _, ok := related[id]; ok {
				return true
			}
		}
	}
	return false
}

func authPolicyHasIssue(policy models.AuthPolicySummary, issues []models.Issue) bool {
	collectorName := map[string]string{
		"security-defaults":    "auth_policies.security_defaults",
		"authorization-policy": "auth_policies.authorization_policy",
		"conditional-access":   "auth_policies.conditional_access",
	}[policy.PolicyType]
	if collectorName == "" {
		return false
	}
	for _, issue := range issues {
		if issue.Context["collector"] == collectorName {
			return true
		}
	}
	return false
}

func authPolicyStateRank(policy models.AuthPolicySummary) (int, int) {
	state := strings.ToLower(policy.State)
	controls := map[string]struct{}{}
	for _, control := range policy.Controls {
		controls[strings.ToLower(control)] = struct{}{}
	}

	switch policy.PolicyType {
	case "security-defaults":
		if state == "disabled" {
			return 0, 0
		}
		return 2, 0
	case "authorization-policy":
		riskyControls := []string{
			"risky-app-consent:enabled",
			"guest-invites:everyone",
			"users-can-register-apps",
			"user-consent:self-service",
		}
		for _, control := range riskyControls {
			if _, ok := controls[control]; ok {
				return 0, 0
			}
		}
		return 1, 0
	case "conditional-access":
		switch state {
		case "disabled":
			return 0, 0
		case "enabledforreportingbutnotenforced":
			return 1, 0
		case "enabled":
			return 2, 0
		default:
			return 3, 0
		}
	default:
		return 9, 0
	}
}
