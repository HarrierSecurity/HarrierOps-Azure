package providers

import (
	"context"
	"strings"

	"harrierops-azure/internal/models"
)

func (provider StaticProvider) CrossTenant(_ context.Context, tenant string, subscription string) (CrossTenantFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID
	subscriptionScope := "/subscriptions/" + subscriptionID

	return CrossTenantFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		CrossTenantPaths: []models.CrossTenantPathSummary{
			{
				AttackPath: "control",
				ID:         subscriptionScope + "/providers/Microsoft.ManagedServices/registrationAssignments/lh-sub-contoso",
				Name:       "Contoso baseline ops",
				Posture:    models.StringPtr("strongest=Owner; eligible=1; delegated-role-assign=yes"),
				Priority:   "high",
				RelatedIDs: []string{
					subscriptionScope + "/providers/Microsoft.ManagedServices/registrationAssignments/lh-sub-contoso",
					subscriptionScope,
					subscriptionScope + "/providers/Microsoft.ManagedServices/registrationDefinitions/lh-def-contoso-sub",
				},
				Scope:      models.StringPtr("subscription::" + subscriptionID),
				SignalType: "lighthouse",
				Summary:    "managed by Contoso Corp.; strongest role Owner; 1 eligible authorization(s)",
				TenantID:   models.StringPtr("33333333-3333-3333-3333-333333333333"),
				TenantName: models.StringPtr("Contoso Corp."),
			},
			{
				AttackPath: "pivot",
				ID:         "sp-external-ci",
				Name:       "external-ci-bridge",
				Posture:    models.StringPtr("roles=Owner; assignments=2; scopes=1"),
				Priority:   "high",
				RelatedIDs: []string{"sp-external-ci"},
				Scope:      models.StringPtr("tenant"),
				SignalType: "external-sp",
				Summary:    "Service principal 'external-ci-bridge' appears to be owned by another tenant and holds high-impact Azure role assignments in the current environment.",
				TenantID:   models.StringPtr("66666666-6666-6666-6666-666666666666"),
				TenantName: nil,
			},
			{
				AttackPath: "entry",
				ID:         "authorizationPolicy",
				Name:       "Authorization Policy",
				Posture:    models.StringPtr("guest-invites=everyone; app-registration=yes; user-consent=self-service"),
				Priority:   "high",
				RelatedIDs: []string{"authorizationPolicy"},
				Scope:      models.StringPtr("tenant"),
				SignalType: "policy",
				Summary:    "guest invites: everyone; users can register apps; self-service permission grant policies assigned; legacy MSOL PowerShell blocked",
				TenantID:   models.StringPtr(session.TenantID),
				TenantName: nil,
			},
			{
				AttackPath: "control",
				ID:         subscriptionScope + "/resourceGroups/rg-platform/providers/Microsoft.ManagedServices/registrationAssignments/lh-rg-platform-contrib",
				Name:       "Fabrikam platform support",
				Posture:    models.StringPtr("strongest=Contributor; eligible=0"),
				Priority:   "low",
				RelatedIDs: []string{
					subscriptionScope + "/resourceGroups/rg-platform/providers/Microsoft.ManagedServices/registrationAssignments/lh-rg-platform-contrib",
					subscriptionScope + "/resourceGroups/rg-platform",
					subscriptionScope + "/providers/Microsoft.ManagedServices/registrationDefinitions/lh-def-fabrikam-rg",
				},
				Scope:      models.StringPtr("resource-group::rg-platform"),
				SignalType: "lighthouse",
				Summary:    "managed by Fabrikam Ops; strongest role Contributor",
				TenantID:   models.StringPtr("44444444-4444-4444-4444-444444444444"),
				TenantName: models.StringPtr("Fabrikam Ops"),
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (provider AzureProvider) CrossTenant(ctx context.Context, tenant string, subscription string) (CrossTenantFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return CrossTenantFacts{}, err
	}

	lighthouseFacts, err := provider.Lighthouse(ctx, tenant, subscription)
	if err != nil {
		return CrossTenantFacts{}, err
	}
	authPolicyFacts, err := provider.AuthPolicies(ctx, tenant, subscription)
	if err != nil {
		return CrossTenantFacts{}, err
	}
	principalFacts, err := provider.Principals(ctx, tenant, subscription)
	if err != nil {
		return CrossTenantFacts{}, err
	}

	issues := append([]models.Issue{}, lighthouseFacts.Issues...)
	issues = append(issues, authPolicyFacts.Issues...)
	issues = append(issues, principalFacts.Issues...)

	principalByID := map[string]models.PrincipalSummary{}
	for _, principal := range principalFacts.Principals {
		if principal.ID == "" {
			continue
		}
		principalByID[principal.ID] = principal
	}

	paths := make([]models.CrossTenantPathSummary, 0, len(lighthouseFacts.LighthouseDelegations))
	for _, delegation := range lighthouseFacts.LighthouseDelegations {
		paths = append(paths, crossTenantLighthouseRow(delegation))
	}

	if session.tenantID != "" {
		graphToken, err := accessToken(ctx, session.credential, graphScope)
		if err != nil {
			issues = append(issues, issueFromError("cross_tenant.service_principals", err))
		} else {
			servicePrincipals, err := graphListObjects(ctx, graphToken, graphCollectionURL("/servicePrincipals", map[string]string{
				"$select": "id,displayName,appId,appOwnerOrganizationId",
			}))
			if err != nil {
				issues = append(issues, issueFromError("cross_tenant.service_principals", err))
			} else {
				for _, servicePrincipal := range servicePrincipals {
					principalID := mapStringValue(servicePrincipal, "id")
					ownerTenantID := mapStringValue(servicePrincipal, "appOwnerOrganizationId")
					if principalID == "" || ownerTenantID == "" || strings.EqualFold(ownerTenantID, session.tenantID) {
						continue
					}
					paths = append(paths, crossTenantExternalServicePrincipalRow(servicePrincipal, principalByID[principalID]))
				}
			}
		}
	}

	paths = append(paths, crossTenantPolicyRows(authPolicyFacts.AuthPolicies, session.tenantID)...)

	return CrossTenantFacts{
		TenantID:         session.tenantID,
		SubscriptionID:   session.subscription.ID,
		CrossTenantPaths: paths,
		Issues:           issues,
	}, nil
}

func crossTenantLighthouseRow(item models.LighthouseDelegationAsset) models.CrossTenantPathSummary {
	scopeLabel := firstNonEmpty(stringPtrValue(item.ScopeDisplayName), item.ScopeID)
	scope := "subscription::" + scopeLabel
	if item.ScopeType == "resource_group" {
		scope = "resource-group::" + scopeLabel
	}

	postureParts := []string{"strongest=" + firstNonEmpty(stringPtrValue(item.StrongestRoleName), "unknown")}
	postureParts = append(postureParts, "eligible="+intText(item.EligibleAuthorizationCount))
	if item.HasDelegatedRoleAssignments {
		postureParts = append(postureParts, "delegated-role-assign=yes")
	}

	priority := "medium"
	switch {
	case item.ScopeType == "subscription" && (item.HasOwnerRole || item.HasUserAccessAdministrator):
		priority = "high"
	case item.ScopeType == "subscription" || item.HasOwnerRole:
		priority = "medium"
	default:
		priority = "low"
	}

	return models.CrossTenantPathSummary{
		AttackPath: "control",
		ID:         item.ID,
		Name:       firstNonEmpty(stringPtrValue(item.RegistrationDefinitionName), item.Name, "lighthouse"),
		Posture:    models.StringPtr(strings.Join(postureParts, "; ")),
		Priority:   priority,
		RelatedIDs: append([]string{}, item.RelatedIDs...),
		Scope:      models.StringPtr(scope),
		SignalType: "lighthouse",
		Summary:    firstNonEmpty(item.Summary, "Outside tenant has delegated management visibility at this scope."),
		TenantID:   item.ManagedByTenantID,
		TenantName: item.ManagedByTenantName,
	}
}

func crossTenantExternalServicePrincipalRow(servicePrincipal map[string]any, principal models.PrincipalSummary) models.CrossTenantPathSummary {
	assignmentCount := principal.RoleAssignmentCount
	scopeCount := len(principal.ScopeIDs)
	roleNames := append([]string{}, principal.RoleNames...)
	highImpact := false
	for _, roleName := range roleNames {
		if crossTenantHighImpactRole(roleName) {
			highImpact = true
			break
		}
	}

	priority := "low"
	if highImpact {
		priority = "high"
	} else if assignmentCount > 0 {
		priority = "medium"
	}

	postureParts := []string{}
	if len(roleNames) > 0 {
		limit := len(roleNames)
		if limit > 3 {
			limit = 3
		}
		postureParts = append(postureParts, "roles="+strings.Join(roleNames[:limit], ","))
	} else {
		postureParts = append(postureParts, "roles=none-visible")
	}
	postureParts = append(postureParts, "assignments="+intText(assignmentCount))
	postureParts = append(postureParts, "scopes="+intText(scopeCount))

	displayName := firstNonEmpty(
		mapStringValue(servicePrincipal, "displayName"),
		stringPtrValue(principal.DisplayName),
		mapStringValue(servicePrincipal, "appId"),
		mapStringValue(servicePrincipal, "id"),
		"external service principal",
	)

	summary := "Service principal '" + displayName + "' appears to be owned by another tenant and is readable in the current tenant, but no Azure role assignments are visible through the current read path."
	if highImpact {
		summary = "Service principal '" + displayName + "' appears to be owned by another tenant and holds high-impact Azure role assignments in the current environment."
	} else if assignmentCount > 0 {
		summary = "Service principal '" + displayName + "' appears to be owned by another tenant and also holds visible Azure role assignments in the current environment."
	}

	return models.CrossTenantPathSummary{
		AttackPath: "pivot",
		ID:         mapStringValue(servicePrincipal, "id"),
		Name:       displayName,
		Posture:    models.StringPtr(strings.Join(postureParts, "; ")),
		Priority:   priority,
		RelatedIDs: filterNonEmpty(mapStringValue(servicePrincipal, "id")),
		Scope:      models.StringPtr("tenant"),
		SignalType: "external-sp",
		Summary:    summary,
		TenantID:   models.StringPtr(mapStringValue(servicePrincipal, "appOwnerOrganizationId")),
		TenantName: nil,
	}
}

func crossTenantPolicyRows(authPolicies []models.AuthPolicySummary, tenantID string) []models.CrossTenantPathSummary {
	rows := []models.CrossTenantPathSummary{}
	for _, policy := range authPolicies {
		if policy.PolicyType != "authorization-policy" {
			continue
		}

		controls := map[string]struct{}{}
		for _, control := range policy.Controls {
			controls[control] = struct{}{}
		}

		guestInvites := ""
		for control := range controls {
			if strings.HasPrefix(control, "guest-invites:") {
				guestInvites = strings.TrimPrefix(control, "guest-invites:")
				break
			}
		}
		usersCanRegisterApps := false
		_, usersCanRegisterApps = controls["users-can-register-apps"]
		userConsent := false
		_, userConsent = controls["user-consent:self-service"]

		if guestInvites == "" && !usersCanRegisterApps && !userConsent {
			continue
		}

		postureParts := []string{}
		if guestInvites != "" {
			postureParts = append(postureParts, "guest-invites="+guestInvites)
		}
		if usersCanRegisterApps {
			postureParts = append(postureParts, "app-registration=yes")
		}
		if userConsent {
			postureParts = append(postureParts, "user-consent=self-service")
		}

		priority := "low"
		switch {
		case guestInvites == "everyone":
			priority = "high"
		case usersCanRegisterApps || userConsent:
			priority = "medium"
		}

		rowID := firstNonEmpty(firstString(policy.RelatedIDs), policy.Name)
		rows = append(rows, models.CrossTenantPathSummary{
			AttackPath: "entry",
			ID:         rowID,
			Name:       firstNonEmpty(policy.Name, "Authorization Policy"),
			Posture:    models.StringPtr(strings.Join(postureParts, "; ")),
			Priority:   priority,
			RelatedIDs: append([]string{}, policy.RelatedIDs...),
			Scope:      models.StringPtr("tenant"),
			SignalType: "policy",
			Summary:    firstNonEmpty(policy.Summary, "Tenant policy may make outside-tenant entry or consent easier to extend."),
			TenantID:   models.StringPtr(tenantID),
			TenantName: nil,
		})
	}
	return rows
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func crossTenantHighImpactRole(roleName string) bool {
	switch strings.ToLower(strings.TrimSpace(roleName)) {
	case "owner", "contributor", "user access administrator":
		return true
	default:
		return false
	}
}
