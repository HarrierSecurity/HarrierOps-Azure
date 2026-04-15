package providers

import (
	"context"
	"strconv"
	"strings"

	"harrierops-azure/internal/models"
)

func (provider AzureProvider) AuthPolicies(ctx context.Context, tenant string, subscription string) (AuthPoliciesFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return AuthPoliciesFacts{}, err
	}

	graphToken, err := accessToken(ctx, session.credential, graphScope)
	if err != nil {
		return AuthPoliciesFacts{}, err
	}

	issues := []models.Issue{}
	authPolicies := []models.AuthPolicySummary{}

	defaults, err := graphGetObject(ctx, graphToken, graphObjectURL("/policies/identitySecurityDefaultsEnforcementPolicy", nil))
	if err != nil {
		issues = append(issues, issueFromError("auth_policies.security_defaults", err))
	} else {
		authPolicies = append(authPolicies, securityDefaultsSummary(defaults))
	}

	authorizationPolicy, err := graphGetObject(ctx, graphToken, graphObjectURL("/policies/authorizationPolicy", nil))
	if err != nil {
		issues = append(issues, issueFromError("auth_policies.authorization_policy", err))
	} else {
		authPolicies = append(authPolicies, authorizationPolicySummary(authorizationPolicy))
	}

	conditionalAccessPolicies, err := graphListObjects(ctx, graphToken, graphCollectionURL("/identity/conditionalAccess/policies", nil))
	if err != nil {
		issues = append(issues, issueFromError("auth_policies.conditional_access", err))
	} else {
		for _, policy := range conditionalAccessPolicies {
			authPolicies = append(authPolicies, conditionalAccessPolicySummary(policy))
		}
	}

	return AuthPoliciesFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		AuthPolicies:   authPolicies,
		Issues:         issues,
	}, nil
}

func securityDefaultsSummary(policy map[string]any) models.AuthPolicySummary {
	enabled := mapBoolValue(policy, "isEnabled")
	controls := []string{}
	summary := "Security defaults are disabled for the tenant."
	state := "disabled"
	if enabled {
		controls = []string{"baseline-mfa", "legacy-auth-protection"}
		summary = "Security defaults are enabled for the tenant."
		state = "enabled"
	}

	return models.AuthPolicySummary{
		Controls:   controls,
		Name:       firstNonEmpty(mapStringValue(policy, "displayName"), "Security Defaults"),
		PolicyType: "security-defaults",
		RelatedIDs: filterNonEmpty(mapStringValue(policy, "id")),
		Scope:      models.StringPtr("tenant"),
		State:      state,
		Summary:    summary,
	}
}

func authorizationPolicySummary(policy map[string]any) models.AuthPolicySummary {
	controls := []string{}

	inviteSetting := mapStringValue(policy, "allowInvitesFrom")
	if inviteSetting != "" {
		controls = append(controls, "guest-invites:"+inviteSetting)
	}
	if mapBoolValue(policy, "allowUserConsentForRiskyApps") {
		controls = append(controls, "risky-app-consent:enabled")
	}
	if mapBoolValue(policy, "allowedToUseSSPR") {
		controls = append(controls, "sspr:enabled")
	}
	if mapBoolValue(policy, "blockMsolPowerShell") {
		controls = append(controls, "legacy-msol-powershell:blocked")
	}

	defaultPermissions := mapValue(policy, "defaultUserRolePermissions")
	if mapBoolValue(defaultPermissions, "allowedToCreateApps") {
		controls = append(controls, "users-can-register-apps")
	}
	if mapBoolValue(defaultPermissions, "allowedToCreateSecurityGroups") {
		controls = append(controls, "users-can-create-security-groups")
	}
	if len(listValue(defaultPermissions, "permissionGrantPoliciesAssigned")) > 0 {
		controls = append(controls, "user-consent:self-service")
	}

	summaryParts := []string{}
	if inviteSetting != "" {
		summaryParts = append(summaryParts, "guest invites: "+inviteSetting)
	}
	if mapBoolValue(defaultPermissions, "allowedToCreateApps") {
		summaryParts = append(summaryParts, "users can register apps")
	} else {
		summaryParts = append(summaryParts, "users cannot register apps")
	}
	if len(listValue(defaultPermissions, "permissionGrantPoliciesAssigned")) > 0 {
		summaryParts = append(summaryParts, "self-service permission grant policies assigned")
	}
	if mapBoolValue(policy, "allowUserConsentForRiskyApps") {
		summaryParts = append(summaryParts, "risky app consent enabled")
	}
	if mapBoolValue(policy, "blockMsolPowerShell") {
		summaryParts = append(summaryParts, "legacy MSOL PowerShell blocked")
	}

	summary := "Authorization policy retrieved."
	if len(summaryParts) > 0 {
		summary = strings.Join(summaryParts, "; ")
	}

	return models.AuthPolicySummary{
		Controls:   controls,
		Name:       firstNonEmpty(mapStringValue(policy, "displayName"), "Authorization Policy"),
		PolicyType: "authorization-policy",
		RelatedIDs: filterNonEmpty(mapStringValue(policy, "id")),
		Scope:      models.StringPtr("tenant"),
		State:      "configured",
		Summary:    summary,
	}
}

func conditionalAccessPolicySummary(policy map[string]any) models.AuthPolicySummary {
	grantControls := stringList(mapValue(mapValue(policy, "grantControls"), "builtInControls"))
	sessionControls := []string{}
	rawSessionControls := mapValue(policy, "sessionControls")
	for key, value := range rawSessionControls {
		switch typed := value.(type) {
		case nil:
		case bool:
			if typed {
				sessionControls = append(sessionControls, key)
			}
		case []any:
			if len(typed) > 0 {
				sessionControls = append(sessionControls, key)
			}
		case map[string]any:
			if len(typed) > 0 {
				sessionControls = append(sessionControls, key)
			}
		default:
			sessionControls = append(sessionControls, key)
		}
	}

	authStrength := mapStringValue(mapValue(mapValue(policy, "grantControls"), "authenticationStrength"), "displayName")
	if authStrength != "" {
		grantControls = append(grantControls, "authentication-strength:"+authStrength)
	}

	users := mapValue(mapValue(policy, "conditions"), "users")
	applications := mapValue(mapValue(policy, "conditions"), "applications")

	scopeParts := []string{}
	if containsString(stringList(mapValue(users, "includeUsers")), "All") {
		scopeParts = append(scopeParts, "users:all")
	}
	if roles := stringList(mapValue(users, "includeRoles")); len(roles) > 0 {
		scopeParts = append(scopeParts, "roles:"+strconv.Itoa(len(roles)))
	}
	if containsString(stringList(mapValue(applications, "includeApplications")), "All") {
		scopeParts = append(scopeParts, "apps:all")
	} else if apps := stringList(mapValue(applications, "includeApplications")); len(apps) > 0 {
		scopeParts = append(scopeParts, "apps:"+strconv.Itoa(len(apps)))
	}

	state := firstNonEmpty(mapStringValue(policy, "state"), "unknown")
	summaryParts := []string{"state: " + state}
	if len(grantControls) > 0 {
		summaryParts = append(summaryParts, "grants: "+strings.Join(grantControls, ", "))
	}
	if len(sessionControls) > 0 {
		summaryParts = append(summaryParts, "session: "+strings.Join(sessionControls, ", "))
	}
	if len(scopeParts) > 0 {
		summaryParts = append(summaryParts, "scope: "+strings.Join(scopeParts, ", "))
	}

	scope := "scoped"
	if len(scopeParts) > 0 {
		scope = strings.Join(scopeParts, ", ")
	}

	return models.AuthPolicySummary{
		Controls:   append(grantControls, sessionControls...),
		Name:       firstNonEmpty(mapStringValue(policy, "displayName"), mapStringValue(policy, "id"), "Conditional Access"),
		PolicyType: "conditional-access",
		RelatedIDs: filterNonEmpty(mapStringValue(policy, "id")),
		Scope:      models.StringPtr(scope),
		State:      state,
		Summary:    strings.Join(summaryParts, "; "),
	}
}

func stringList(input any) []string {
	values := []string{}
	for _, item := range listValue(input) {
		value := strings.TrimSpace(stringValue(item))
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func filterNonEmpty(values ...string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filtered = append(filtered, value)
		}
	}
	return filtered
}
