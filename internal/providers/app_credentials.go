package providers

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"harrierops-azure/internal/models"
)

func (provider AzureProvider) AppCredentials(ctx context.Context, tenant string, subscription string) (AppCredentialsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return AppCredentialsFacts{}, err
	}

	graphToken, err := accessToken(ctx, session.credential, graphScope)
	if err != nil {
		return AppCredentialsFacts{}, err
	}

	rbacFacts := provider.collectRBACFacts(ctx, session)
	roleContexts := appCredentialRoleContextByPrincipal(rbacFacts)
	currentPrincipalID, currentPrincipalName := currentPrincipalFromSession(session)
	issues := append([]models.Issue{}, rbacFacts.Issues...)

	servicePrincipals, err := graphListObjects(ctx, graphToken, graphCollectionURL("/servicePrincipals", map[string]string{
		"$select": "id,appId,displayName,servicePrincipalType,passwordCredentials,keyCredentials",
	}))
	if err != nil {
		issues = append(issues, issueFromError("app_credentials.service_principals", err))
		servicePrincipals = []map[string]any{}
	}

	applications, err := graphListObjects(ctx, graphToken, graphCollectionURL("/applications", map[string]string{
		"$select": "id,appId,displayName,passwordCredentials,keyCredentials",
	}))
	if err != nil {
		issues = append(issues, issueFromError("app_credentials.applications", err))
		applications = []map[string]any{}
	}

	servicePrincipalByAppID := map[string]map[string]any{}
	for _, item := range servicePrincipals {
		if appID := mapStringValue(item, "appId"); appID != "" {
			servicePrincipalByAppID[appID] = item
		}
	}

	applicationOwnerRequests := make([]graphBatchRequest, 0, len(applications))
	applicationFederatedRequests := make([]graphBatchRequest, 0, len(applications))
	for _, application := range applications {
		appObjectID := mapStringValue(application, "id")
		if appObjectID == "" {
			continue
		}
		applicationOwnerRequests = append(applicationOwnerRequests, graphBatchRequest{
			Key: appObjectID,
			URL: graphCollectionURL("/applications/"+url.PathEscape(appObjectID)+"/owners", map[string]string{
				"$select": "id,displayName,userPrincipalName,appId,servicePrincipalType",
			}),
		})
		applicationFederatedRequests = append(applicationFederatedRequests, graphBatchRequest{
			Key: appObjectID,
			URL: graphCollectionURL("/applications/"+url.PathEscape(appObjectID)+"/federatedIdentityCredentials", map[string]string{
				"$select": "id,issuer,subject,name",
			}),
		})
	}

	servicePrincipalOwnerRequests := make([]graphBatchRequest, 0, len(servicePrincipals))
	for _, servicePrincipal := range servicePrincipals {
		spID := mapStringValue(servicePrincipal, "id")
		if spID == "" {
			continue
		}
		servicePrincipalOwnerRequests = append(servicePrincipalOwnerRequests, graphBatchRequest{
			Key: spID,
			URL: graphCollectionURL("/servicePrincipals/"+url.PathEscape(spID)+"/owners", map[string]string{
				"$select": "id,displayName,userPrincipalName,appId,servicePrincipalType",
			}),
		})
	}

	applicationOwnersByID, applicationOwnerErrs := graphBatchListObjectsByKey(ctx, graphToken, applicationOwnerRequests)
	federatedByApplicationID, federatedErrs := graphBatchListObjectsByKey(ctx, graphToken, applicationFederatedRequests)
	servicePrincipalOwnersByID, servicePrincipalOwnerErrs := graphBatchListObjectsByKey(ctx, graphToken, servicePrincipalOwnerRequests)

	rows := []models.AppCredentialSummary{}

	for _, application := range applications {
		appObjectID := mapStringValue(application, "id")
		if appObjectID == "" {
			continue
		}
		appDisplayName := firstNonEmpty(mapStringValue(application, "displayName"), mapStringValue(application, "appId"), appObjectID)
		backingSP := servicePrincipalByAppID[mapStringValue(application, "appId")]
		roleContext := appCredentialRoleContext(roleContexts, mapStringValue(backingSP, "id"), "Application")

		if ownerErr, ok := applicationOwnerErrs[appObjectID]; ok {
			issues = append(issues, issueFromError("app_credentials.applications["+appObjectID+"].owners", ownerErr))
		}
		if credentialErr, ok := federatedErrs[appObjectID]; ok {
			issues = append(issues, issueFromError("app_credentials.applications["+appObjectID+"].federated_credentials", credentialErr))
		}

		for _, owner := range applicationOwnersByID[appObjectID] {
			if mapStringValue(owner, "id") == currentPrincipalID {
				rows = append(rows,
					directlyAddableApplicationCredentialRow(owner, application, backingSP, roleContext, currentPrincipalName),
					directlyAddableFederatedTrustRow(owner, application, backingSP, roleContext, currentPrincipalName),
				)
			}
		}

		rows = append(rows, existingGraphCredentialRows(
			"Application",
			appObjectID,
			appDisplayName,
			backingSP,
			application,
			roleContext,
		)...)
		rows = append(rows, existingFederatedTrustRows(
			application,
			backingSP,
			federatedByApplicationID[appObjectID],
			roleContext,
		)...)
	}

	for _, servicePrincipal := range servicePrincipals {
		spID := mapStringValue(servicePrincipal, "id")
		if spID == "" {
			continue
		}
		spDisplayName := firstNonEmpty(mapStringValue(servicePrincipal, "displayName"), spID)
		roleContext := appCredentialRoleContext(roleContexts, spID, "ServicePrincipal")

		if ownerErr, ok := servicePrincipalOwnerErrs[spID]; ok {
			issues = append(issues, issueFromError("app_credentials.service_principals["+spID+"].owners", ownerErr))
		}

		for _, owner := range servicePrincipalOwnersByID[spID] {
			if mapStringValue(owner, "id") != currentPrincipalID {
				continue
			}
			rows = append(rows, directlyAddableServicePrincipalRow(owner, servicePrincipal, roleContext, currentPrincipalName))
		}

		rows = append(rows, existingGraphCredentialRows(
			"ServicePrincipal",
			spID,
			spDisplayName,
			servicePrincipal,
			servicePrincipal,
			roleContext,
		)...)
	}

	return AppCredentialsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		AppCredentials: rows,
		Issues:         issues,
	}, nil
}

func currentPrincipalFromSession(session azureSession) (string, string) {
	return currentPrincipalFromClaims(session.claims)
}

func appCredentialRoleContextByPrincipal(rbacFacts RBACFacts) map[string]string {
	type principalRoleState struct {
		allRoles  []string
		highRoles []string
		scopeIDs  []string
	}
	states := map[string]*principalRoleState{}
	for _, assignment := range rbacFacts.RoleAssignments {
		if assignment.PrincipalID == "" {
			continue
		}
		state := states[assignment.PrincipalID]
		if state == nil {
			state = &principalRoleState{}
			states[assignment.PrincipalID] = state
		}
		state.allRoles = appendUniqueString(state.allRoles, assignment.RoleName)
		state.scopeIDs = appendUniqueString(state.scopeIDs, assignment.ScopeID)
		if _, ok := highImpactRoleNames[strings.ToLower(strings.TrimSpace(assignment.RoleName))]; ok {
			state.highRoles = appendUniqueString(state.highRoles, assignment.RoleName)
		}
	}

	contexts := map[string]string{}
	for principalID, state := range states {
		scopeLabel := "1 visible scope"
		if len(state.scopeIDs) != 1 {
			scopeLabel = fmt.Sprintf("%d visible scopes", len(state.scopeIDs))
		}
		if len(state.highRoles) > 0 {
			contexts[principalID] = fmt.Sprintf(
				"Backed identity already holds high-impact Azure roles (%s) across %s.",
				strings.Join(state.highRoles, ", "),
				scopeLabel,
			)
			continue
		}
		if len(state.allRoles) > 0 {
			contexts[principalID] = fmt.Sprintf(
				"Backed identity already holds visible Azure roles (%s) across %s.",
				strings.Join(state.allRoles, ", "),
				scopeLabel,
			)
		}
	}
	return contexts
}

func appCredentialRoleContext(roleContexts map[string]string, principalID string, targetType string) string {
	if principalID == "" {
		if targetType == "Application" {
			return "No visible Azure-facing service principal is linked to this application in the current environment."
		}
		return "No visible Azure RBAC is attached to this service principal in the current subscription."
	}
	if context := strings.TrimSpace(roleContexts[principalID]); context != "" {
		return context
	}
	return "No visible Azure RBAC is attached to this identity in the current subscription."
}

func graphCredentialCount(item map[string]any, field string) int {
	raw, ok := item[field]
	if !ok || raw == nil {
		return 0
	}
	switch typed := raw.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	default:
		return 0
	}
}

func existingGraphCredentialRows(
	targetType string,
	targetID string,
	targetName string,
	backingSP map[string]any,
	item map[string]any,
	roleContext string,
) []models.AppCredentialSummary {
	rows := []models.AppCredentialSummary{}
	if passwordCount := graphCredentialCount(item, "passwordCredentials"); passwordCount > 0 {
		rows = append(rows, existingCredentialRow(
			targetType,
			targetID,
			targetName,
			backingSP,
			"password",
			passwordCount,
			roleContext,
		))
	}
	if keyCount := graphCredentialCount(item, "keyCredentials"); keyCount > 0 {
		rows = append(rows, existingCredentialRow(
			targetType,
			targetID,
			targetName,
			backingSP,
			"key",
			keyCount,
			roleContext,
		))
	}
	return rows
}

func existingFederatedTrustRows(
	application map[string]any,
	backingSP map[string]any,
	credentials []map[string]any,
	roleContext string,
) []models.AppCredentialSummary {
	rows := make([]models.AppCredentialSummary, 0, len(credentials))
	for _, credential := range credentials {
		rows = append(rows, federatedTrustRow(application, backingSP, credential, roleContext))
	}
	return rows
}

func existingCredentialRow(
	targetType string,
	targetID string,
	targetName string,
	backingSP map[string]any,
	credentialType string,
	count int,
	roleContext string,
) models.AppCredentialSummary {
	countLabel := fmt.Sprintf("%d visible %s credential metadata entr", count, credentialType)
	if count == 1 {
		countLabel += "y"
	} else {
		countLabel += "ies"
	}
	currentEvidence := fmt.Sprintf(
		"%s '%s' already has %s.",
		targetType,
		targetName,
		countLabel,
	)
	missingProof := "This row shows existing authentication material, not a current-identity path to change it."
	operatorAction := "Review who can modify this object and whether the existing credential material is still needed."
	recommendedFix := "Remove stale authentication material and tighten ownership over the identity that accepts it."
	summary := currentEvidence + " " + roleContext + " " + missingProof

	return models.AppCredentialSummary{
		RowClass:                  "existing_credential",
		TargetObjectType:          targetType,
		TargetObjectID:            targetID,
		TargetObjectName:          targetName,
		BackingServicePrincipalID: stringPtr(mapStringValue(backingSP, "id")),
		BackingServicePrincipalName: stringPtr(firstNonEmpty(
			mapStringValue(backingSP, "displayName"),
			mapStringValue(backingSP, "id"),
		)),
		CredentialType:        stringPtr(credentialType),
		ControlPath:           "existing-auth-material",
		RoleContext:           roleContext,
		TenantContext:         "current-tenant",
		CurrentEvidence:       currentEvidence,
		MissingProof:          missingProof,
		OperatorActionability: operatorAction,
		RecommendedFixFocus:   recommendedFix,
		Summary:               summary,
		RelatedIDs:            dedupeStrings([]string{targetID, mapStringValue(backingSP, "id")}),
	}
}

func federatedTrustRow(application map[string]any, backingSP map[string]any, credential map[string]any, roleContext string) models.AppCredentialSummary {
	appID := mapStringValue(application, "id")
	appDisplayName := firstNonEmpty(mapStringValue(application, "displayName"), mapStringValue(application, "appId"), appID)
	subject := firstNonEmpty(mapStringValue(credential, "subject"), "unknown subject")
	issuer := firstNonEmpty(mapStringValue(credential, "issuer"), "unknown issuer")
	currentEvidence := fmt.Sprintf(
		"Application '%s' already trusts federated subject '%s' from issuer '%s'.",
		appDisplayName,
		subject,
		issuer,
	)
	missingProof := "This row shows existing federated trust, not that the current identity can change it."
	operatorAction := "Review whether this external trust path is still required and who can modify the application that carries it."
	recommendedFix := "Remove or tighten federated trust that no longer needs to yield Azure-facing access."
	summary := currentEvidence + " " + roleContext + " " + missingProof

	return models.AppCredentialSummary{
		RowClass:                  "federated_trust_present",
		TargetObjectType:          "Application",
		TargetObjectID:            appID,
		TargetObjectName:          appDisplayName,
		BackingServicePrincipalID: stringPtr(mapStringValue(backingSP, "id")),
		BackingServicePrincipalName: stringPtr(firstNonEmpty(
			mapStringValue(backingSP, "displayName"),
			mapStringValue(backingSP, "id"),
		)),
		CredentialType:        stringPtr("federated"),
		ControlPath:           "existing-federated-trust",
		RoleContext:           roleContext,
		TenantContext:         "current-tenant",
		CurrentEvidence:       currentEvidence,
		MissingProof:          missingProof,
		OperatorActionability: operatorAction,
		RecommendedFixFocus:   recommendedFix,
		Summary:               summary,
		RelatedIDs: dedupeStrings([]string{
			appID,
			mapStringValue(credential, "id"),
			mapStringValue(backingSP, "id"),
		}),
	}
}

func directlyAddableApplicationCredentialRow(owner map[string]any, application map[string]any, backingSP map[string]any, roleContext string, currentPrincipalName string) models.AppCredentialSummary {
	appID := mapStringValue(application, "id")
	appDisplayName := firstNonEmpty(mapStringValue(application, "displayName"), mapStringValue(application, "appId"), appID)
	currentEvidence := fmt.Sprintf(
		"Current identity '%s' visibly owns application '%s', and application ownership can change authentication material accepted here.",
		currentPrincipalName,
		appDisplayName,
	)
	missingProof := ""
	operatorAction := "Treat this as a visible path to add or replace authentication material on the application object."
	recommendedFix := "Remove the ownership path that lets the current identity control this application."
	summary := currentEvidence + " " + roleContext

	return models.AppCredentialSummary{
		RowClass:                  "directly_addable",
		TargetObjectType:          "Application",
		TargetObjectID:            appID,
		TargetObjectName:          appDisplayName,
		BackingServicePrincipalID: stringPtr(mapStringValue(backingSP, "id")),
		BackingServicePrincipalName: stringPtr(firstNonEmpty(
			mapStringValue(backingSP, "displayName"),
			mapStringValue(backingSP, "id"),
		)),
		CredentialType:        stringPtr("password-or-key"),
		ControlPath:           "application-owner",
		RoleContext:           roleContext,
		TenantContext:         "current-tenant",
		CurrentEvidence:       currentEvidence,
		MissingProof:          missingProof,
		OperatorActionability: operatorAction,
		RecommendedFixFocus:   recommendedFix,
		Summary:               summary,
		RelatedIDs:            dedupeStrings([]string{mapStringValue(owner, "id"), appID, mapStringValue(backingSP, "id")}),
	}
}

func directlyAddableFederatedTrustRow(owner map[string]any, application map[string]any, backingSP map[string]any, roleContext string, currentPrincipalName string) models.AppCredentialSummary {
	appID := mapStringValue(application, "id")
	appDisplayName := firstNonEmpty(mapStringValue(application, "displayName"), mapStringValue(application, "appId"), appID)
	currentEvidence := fmt.Sprintf(
		"Current identity '%s' visibly owns application '%s', and federated trust lives on that application object.",
		currentPrincipalName,
		appDisplayName,
	)
	missingProof := "This row shows direct control of the federated-trust surface, not which external subject would be trusted after a change."
	operatorAction := "Treat this as a visible path to add, replace, or widen federated trust on the application."
	recommendedFix := "Remove the ownership path that lets the current identity control federated trust on this application."
	summary := currentEvidence + " " + roleContext + " " + missingProof

	return models.AppCredentialSummary{
		RowClass:                  "directly_addable_federated_trust",
		TargetObjectType:          "Application",
		TargetObjectID:            appID,
		TargetObjectName:          appDisplayName,
		BackingServicePrincipalID: stringPtr(mapStringValue(backingSP, "id")),
		BackingServicePrincipalName: stringPtr(firstNonEmpty(
			mapStringValue(backingSP, "displayName"),
			mapStringValue(backingSP, "id"),
		)),
		CredentialType:        stringPtr("federated"),
		ControlPath:           "application-owner",
		RoleContext:           roleContext,
		TenantContext:         "current-tenant",
		CurrentEvidence:       currentEvidence,
		MissingProof:          missingProof,
		OperatorActionability: operatorAction,
		RecommendedFixFocus:   recommendedFix,
		Summary:               summary,
		RelatedIDs:            dedupeStrings([]string{mapStringValue(owner, "id"), appID, mapStringValue(backingSP, "id")}),
	}
}

func directlyAddableServicePrincipalRow(owner map[string]any, servicePrincipal map[string]any, roleContext string, currentPrincipalName string) models.AppCredentialSummary {
	spID := mapStringValue(servicePrincipal, "id")
	spDisplayName := firstNonEmpty(mapStringValue(servicePrincipal, "displayName"), spID)
	currentEvidence := fmt.Sprintf(
		"Current identity '%s' visibly owns service principal '%s', and service-principal ownership can change authentication material Azure accepts here.",
		currentPrincipalName,
		spDisplayName,
	)
	operatorAction := "Treat this as a visible path to add or replace authentication material on the service principal."
	recommendedFix := "Remove the owner-level control path that lets the current identity control this service principal."
	summary := currentEvidence + " " + roleContext

	return models.AppCredentialSummary{
		RowClass:                    "directly_addable",
		TargetObjectType:            "ServicePrincipal",
		TargetObjectID:              spID,
		TargetObjectName:            spDisplayName,
		BackingServicePrincipalID:   stringPtr(spID),
		BackingServicePrincipalName: stringPtr(spDisplayName),
		CredentialType:              stringPtr("password-or-key"),
		ControlPath:                 "service-principal-owner",
		RoleContext:                 roleContext,
		TenantContext:               "current-tenant",
		CurrentEvidence:             currentEvidence,
		MissingProof:                "",
		OperatorActionability:       operatorAction,
		RecommendedFixFocus:         recommendedFix,
		Summary:                     summary,
		RelatedIDs:                  dedupeStrings([]string{mapStringValue(owner, "id"), spID}),
	}
}
