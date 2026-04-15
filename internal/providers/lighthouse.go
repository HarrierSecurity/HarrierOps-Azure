package providers

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"harrierops-azure/internal/models"
)

const armManagedServicesAPIVersion = "2022-10-01"

func (provider StaticProvider) Lighthouse(_ context.Context, tenant string, subscription string) (LighthouseFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	subscriptionScope := "/subscriptions/" + subscriptionID
	subscriptionAssignmentID := subscriptionScope + "/providers/Microsoft.ManagedServices/registrationAssignments/lh-sub-contoso"
	subscriptionDefinitionID := subscriptionScope + "/providers/Microsoft.ManagedServices/registrationDefinitions/lh-def-contoso-sub"
	platformScope := subscriptionScope + "/resourceGroups/rg-platform"
	platformAssignmentID := platformScope + "/providers/Microsoft.ManagedServices/registrationAssignments/lh-rg-platform-contrib"
	platformDefinitionID := subscriptionScope + "/providers/Microsoft.ManagedServices/registrationDefinitions/lh-def-fabrikam-rg"
	loggingScope := subscriptionScope + "/resourceGroups/rg-logging"
	loggingAssignmentID := loggingScope + "/providers/Microsoft.ManagedServices/registrationAssignments/lh-rg-logging-reader"
	loggingDefinitionID := subscriptionScope + "/providers/Microsoft.ManagedServices/registrationDefinitions/lh-def-logging-reader"

	return LighthouseFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		LighthouseDelegations: []models.LighthouseDelegationAsset{
			{
				AuthorizationCount:          2,
				DefinitionProvisioningState: models.StringPtr("Succeeded"),
				Description:                 models.StringPtr("Contoso subscription delegation"),
				EligibleAuthorizationCount:  1,
				EligiblePrincipalCount:      1,
				EligibleRoleNames:           []string{"Contributor"},
				HasDelegatedRoleAssignments: true,
				HasOwnerRole:                true,
				HasUserAccessAdministrator:  true,
				ID:                          subscriptionAssignmentID,
				ManagedByTenantID:           models.StringPtr("33333333-3333-3333-3333-333333333333"),
				ManagedByTenantName:         models.StringPtr("Contoso Corp."),
				ManageeTenantID:             models.StringPtr(session.TenantID),
				ManageeTenantName:           models.StringPtr("AzureFox Lab Tenant"),
				Name:                        "lh-sub-contoso",
				PlanName:                    models.StringPtr("contoso-plan"),
				PlanProduct:                 models.StringPtr("ops"),
				PlanPublisher:               models.StringPtr("contoso"),
				PrincipalCount:              2,
				ProvisioningState:           models.StringPtr("Succeeded"),
				RegistrationDefinitionID:    models.StringPtr(subscriptionDefinitionID),
				RegistrationDefinitionName:  models.StringPtr("Contoso baseline ops"),
				RelatedIDs:                  []string{subscriptionAssignmentID, subscriptionScope, subscriptionDefinitionID},
				ResourceGroup:               nil,
				RoleNames:                   []string{"Owner", "User Access Administrator"},
				ScopeDisplayName:            models.StringPtr(subscriptionID),
				ScopeID:                     subscriptionScope,
				ScopeType:                   "subscription",
				StrongestRoleName:           models.StringPtr("Owner"),
				Summary:                     "managed by Contoso Corp.; strongest role Owner; 1 eligible authorization(s)",
			},
			{
				AuthorizationCount:          1,
				DefinitionProvisioningState: models.StringPtr("Succeeded"),
				Description:                 models.StringPtr("Fabrikam platform support delegation"),
				EligibleAuthorizationCount:  0,
				EligiblePrincipalCount:      0,
				EligibleRoleNames:           []string{},
				HasDelegatedRoleAssignments: false,
				HasOwnerRole:                false,
				HasUserAccessAdministrator:  false,
				ID:                          platformAssignmentID,
				ManagedByTenantID:           models.StringPtr("44444444-4444-4444-4444-444444444444"),
				ManagedByTenantName:         models.StringPtr("Fabrikam Ops"),
				ManageeTenantID:             models.StringPtr(session.TenantID),
				ManageeTenantName:           models.StringPtr("AzureFox Lab Tenant"),
				Name:                        "lh-rg-platform-contrib",
				PlanName:                    nil,
				PlanProduct:                 nil,
				PlanPublisher:               nil,
				PrincipalCount:              1,
				ProvisioningState:           models.StringPtr("Succeeded"),
				RegistrationDefinitionID:    models.StringPtr(platformDefinitionID),
				RegistrationDefinitionName:  models.StringPtr("Fabrikam platform support"),
				RelatedIDs:                  []string{platformAssignmentID, platformScope, platformDefinitionID},
				ResourceGroup:               models.StringPtr("rg-platform"),
				RoleNames:                   []string{"Contributor"},
				ScopeDisplayName:            models.StringPtr("rg-platform"),
				ScopeID:                     platformScope,
				ScopeType:                   "resource_group",
				StrongestRoleName:           models.StringPtr("Contributor"),
				Summary:                     "managed by Fabrikam Ops; strongest role Contributor",
			},
			{
				AuthorizationCount:          1,
				DefinitionProvisioningState: models.StringPtr("Succeeded"),
				Description:                 models.StringPtr("Northwind logging review"),
				EligibleAuthorizationCount:  0,
				EligiblePrincipalCount:      0,
				EligibleRoleNames:           []string{},
				HasDelegatedRoleAssignments: false,
				HasOwnerRole:                false,
				HasUserAccessAdministrator:  false,
				ID:                          loggingAssignmentID,
				ManagedByTenantID:           models.StringPtr("55555555-5555-5555-5555-555555555555"),
				ManagedByTenantName:         models.StringPtr("Northwind MSP"),
				ManageeTenantID:             models.StringPtr(session.TenantID),
				ManageeTenantName:           models.StringPtr("AzureFox Lab Tenant"),
				Name:                        "lh-rg-logging-reader",
				PlanName:                    nil,
				PlanProduct:                 nil,
				PlanPublisher:               nil,
				PrincipalCount:              1,
				ProvisioningState:           models.StringPtr("Succeeded"),
				RegistrationDefinitionID:    models.StringPtr(loggingDefinitionID),
				RegistrationDefinitionName:  models.StringPtr("Northwind logging review"),
				RelatedIDs:                  []string{loggingAssignmentID, loggingScope, loggingDefinitionID},
				ResourceGroup:               models.StringPtr("rg-logging"),
				RoleNames:                   []string{"Reader"},
				ScopeDisplayName:            models.StringPtr("rg-logging"),
				ScopeID:                     loggingScope,
				ScopeType:                   "resource_group",
				StrongestRoleName:           models.StringPtr("Reader"),
				Summary:                     "managed by Northwind MSP; strongest role Reader",
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (provider AzureProvider) Lighthouse(ctx context.Context, tenant string, subscription string) (LighthouseFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return LighthouseFacts{}, err
	}

	subscriptionScope := "/subscriptions/" + session.subscription.ID
	roleNameCache := map[string]string{}
	seenAssignmentIDs := map[string]struct{}{}
	delegations := []models.LighthouseDelegationAsset{}
	issues := []models.Issue{}

	assignments, err := listManagedServicesAssignments(ctx, session.credential, subscriptionScope)
	if err != nil {
		issues = append(issues, issueFromError("lighthouse.subscription", err))
	} else {
		for _, assignment := range assignments {
			asset := lighthouseDelegationSummary(ctx, session.credential, assignment, subscriptionScope, roleNameCache)
			if asset.ID == "" || lighthouseSeen(seenAssignmentIDs, asset.ID) {
				continue
			}
			delegations = append(delegations, asset)
		}
	}

	resourceGroupsClient := session.clientFactory.NewResourceGroupsClient()
	resourceGroupPager := resourceGroupsClient.NewListPager(nil)
	for resourceGroupPager.More() {
		page, err := resourceGroupPager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("lighthouse.resource_groups", err))
			break
		}
		for _, resourceGroup := range page.Value {
			resourceGroupName := stringValue(resourceGroup.Name)
			if resourceGroupName == "" {
				continue
			}

			scope := subscriptionScope + "/resourceGroups/" + resourceGroupName
			assignments, err := listManagedServicesAssignments(ctx, session.credential, scope)
			if err != nil {
				issues = append(issues, issueFromError("lighthouse.resource_group["+resourceGroupName+"]", err))
				continue
			}
			for _, assignment := range assignments {
				asset := lighthouseDelegationSummary(ctx, session.credential, assignment, scope, roleNameCache)
				if asset.ID == "" || lighthouseSeen(seenAssignmentIDs, asset.ID) {
					continue
				}
				delegations = append(delegations, asset)
			}
		}
	}

	return LighthouseFacts{
		TenantID:              session.tenantID,
		SubscriptionID:        session.subscription.ID,
		LighthouseDelegations: delegations,
		Issues:                issues,
	}, nil
}

func listManagedServicesAssignments(ctx context.Context, credential azcore.TokenCredential, scope string) ([]map[string]any, error) {
	path := scope + "/providers/Microsoft.ManagedServices/registrationAssignments?$expandRegistrationDefinition=true"
	return armListObjects(ctx, credential, path, armManagedServicesAPIVersion)
}

func lighthouseDelegationSummary(
	ctx context.Context,
	credential azcore.TokenCredential,
	assignment map[string]any,
	scope string,
	roleNameCache map[string]string,
) models.LighthouseDelegationAsset {
	assignmentID := mapStringValue(assignment, "id")
	properties := mapValue(assignment, "properties")
	definition := mapValue(properties, "registrationDefinition")
	definitionProperties := mapValue(definition, "properties")
	authorizations := listValue(definitionProperties, "authorizations")
	eligibleAuthorizations := listValue(definitionProperties, "eligibleAuthorizations", "eligible_authorizations")

	roleNames := dedupeStrings(lighthouseRoleNames(ctx, credential, scope, authorizations, roleNameCache))
	eligibleRoleNames := dedupeStrings(lighthouseRoleNames(ctx, credential, scope, eligibleAuthorizations, roleNameCache))
	strongestRoleName := lighthouseStrongestRoleName(append(append([]string{}, roleNames...), eligibleRoleNames...))

	principalCount := len(lighthousePrincipalIDs(authorizations))
	eligiblePrincipalCount := len(lighthousePrincipalIDs(eligibleAuthorizations))

	scopeID := strings.TrimSuffix(assignmentID, "/providers/Microsoft.ManagedServices/registrationAssignments/"+mapStringValue(assignment, "name"))
	if scopeID == "" {
		scopeID = scope
	}
	scopeType := "subscription"
	if strings.Contains(strings.ToLower(scopeID), "/resourcegroups/") {
		scopeType = "resource_group"
	}
	resourceGroup := models.StringPtr(resourceGroupFromID(scopeID))
	if resourceGroup != nil && *resourceGroup == "" {
		resourceGroup = nil
	}
	scopeDisplayNameText := firstNonEmpty(stringPtrValue(resourceGroup), resourceNameFromID(scopeID))
	scopeDisplayName := models.StringPtr(scopeDisplayNameText)

	hasOwnerRole := lighthouseContainsRole(roleNames, eligibleRoleNames, "Owner")
	hasUserAccessAdministrator := lighthouseContainsRole(roleNames, eligibleRoleNames, "User Access Administrator")
	hasDelegatedRoleAssignments := lighthouseHasDelegatedRoleAssignments(authorizations)

	managedByName := firstNonEmpty(
		mapStringValue(definitionProperties, "managedByTenantName", "managed_by_tenant_name"),
		mapStringValue(definitionProperties, "managedByTenantId", "managed_by_tenant_id"),
	)
	summaryParts := []string{}
	if managedByName != "" {
		summaryParts = append(summaryParts, "managed by "+managedByName)
	}
	if strongestRoleName != nil && *strongestRoleName != "" {
		summaryParts = append(summaryParts, "strongest role "+*strongestRoleName)
	}
	if len(eligibleAuthorizations) > 0 {
		summaryParts = append(summaryParts, intText(len(eligibleAuthorizations))+" eligible authorization(s)")
	}
	assignmentState := mapStringValue(properties, "provisioningState", "provisioning_state")
	if assignmentState != "" && !strings.EqualFold(assignmentState, "Succeeded") {
		summaryParts = append(summaryParts, "assignment state "+assignmentState)
	}
	summary := strings.Join(summaryParts, "; ")
	if summary == "" {
		summary = "Azure Lighthouse delegation visible at this scope."
	}

	plan := mapValue(definition, "plan")
	registrationDefinitionID := models.StringPtr(mapStringValue(properties, "registrationDefinitionId", "registration_definition_id"))

	return models.LighthouseDelegationAsset{
		AuthorizationCount:          len(authorizations),
		DefinitionProvisioningState: models.StringPtr(mapStringValue(definitionProperties, "provisioningState", "provisioning_state")),
		Description:                 models.StringPtr(mapStringValue(definitionProperties, "description")),
		EligibleAuthorizationCount:  len(eligibleAuthorizations),
		EligiblePrincipalCount:      eligiblePrincipalCount,
		EligibleRoleNames:           eligibleRoleNames,
		HasDelegatedRoleAssignments: hasDelegatedRoleAssignments,
		HasOwnerRole:                hasOwnerRole,
		HasUserAccessAdministrator:  hasUserAccessAdministrator,
		ID:                          assignmentID,
		ManagedByTenantID:           models.StringPtr(mapStringValue(definitionProperties, "managedByTenantId", "managed_by_tenant_id")),
		ManagedByTenantName:         models.StringPtr(mapStringValue(definitionProperties, "managedByTenantName", "managed_by_tenant_name")),
		ManageeTenantID:             models.StringPtr(mapStringValue(definitionProperties, "manageeTenantId", "managee_tenant_id")),
		ManageeTenantName:           models.StringPtr(mapStringValue(definitionProperties, "manageeTenantName", "managee_tenant_name")),
		Name:                        firstNonEmpty(mapStringValue(assignment, "name"), resourceNameFromID(assignmentID), "unknown"),
		PlanName:                    models.StringPtr(mapStringValue(plan, "name")),
		PlanProduct:                 models.StringPtr(mapStringValue(plan, "product")),
		PlanPublisher:               models.StringPtr(mapStringValue(plan, "publisher")),
		PrincipalCount:              principalCount,
		ProvisioningState:           models.StringPtr(assignmentState),
		RegistrationDefinitionID:    registrationDefinitionID,
		RegistrationDefinitionName: models.StringPtr(firstNonEmpty(
			mapStringValue(definitionProperties, "registrationDefinitionName", "registration_definition_name"),
			mapStringValue(definition, "name"),
		)),
		RelatedIDs: dedupeStrings([]string{
			assignmentID,
			scopeID,
			stringPtrValue(registrationDefinitionID),
		}),
		ResourceGroup:     resourceGroup,
		RoleNames:         roleNames,
		ScopeDisplayName:  scopeDisplayName,
		ScopeID:           scopeID,
		ScopeType:         scopeType,
		StrongestRoleName: strongestRoleName,
		Summary:           summary,
	}
}

func lighthouseRoleNames(
	ctx context.Context,
	credential azcore.TokenCredential,
	scope string,
	authorizations []any,
	roleNameCache map[string]string,
) []string {
	roleNames := []string{}
	for _, authorization := range authorizations {
		item, ok := authorization.(map[string]any)
		if !ok {
			continue
		}
		roleDefinitionID := mapStringValue(item, "roleDefinitionId", "role_definition_id")
		if roleDefinitionID == "" {
			continue
		}
		roleName, err := resolveRoleDefinitionName(ctx, credential, scope, roleDefinitionID, roleNameCache)
		if err != nil || strings.TrimSpace(roleName) == "" {
			roleName = firstNonEmpty(builtInHighImpactRolesByID[roleDefinitionGUID(roleDefinitionID)], "Unknown")
		}
		roleNames = append(roleNames, roleName)
	}
	return roleNames
}

func lighthousePrincipalIDs(authorizations []any) []string {
	principalIDs := []string{}
	for _, authorization := range authorizations {
		item, ok := authorization.(map[string]any)
		if !ok {
			continue
		}
		principalIDs = append(principalIDs, mapStringValue(item, "principalId", "principal_id"))
	}
	return dedupeStrings(principalIDs)
}

func lighthouseHasDelegatedRoleAssignments(authorizations []any) bool {
	for _, authorization := range authorizations {
		item, ok := authorization.(map[string]any)
		if !ok {
			continue
		}
		if len(listValue(item, "delegatedRoleDefinitionIds", "delegated_role_definition_ids")) > 0 {
			return true
		}
	}
	return false
}

func lighthouseStrongestRoleName(roleNames []string) *string {
	available := dedupeStrings(roleNames)
	if len(available) == 0 {
		return nil
	}
	best := available[0]
	for _, roleName := range available[1:] {
		if lighthouseRolePriority(roleName) < lighthouseRolePriority(best) {
			best = roleName
		}
	}
	return models.StringPtr(best)
}

func lighthouseRolePriority(roleName string) string {
	switch strings.ToLower(strings.TrimSpace(roleName)) {
	case "owner":
		return "0:owner"
	case "user access administrator":
		return "1:user access administrator"
	case "contributor":
		return "2:contributor"
	case "reader":
		return "4:reader"
	default:
		return "3:" + strings.ToLower(strings.TrimSpace(roleName))
	}
}

func lighthouseContainsRole(roleNames []string, eligibleRoleNames []string, want string) bool {
	for _, roleName := range append(append([]string{}, roleNames...), eligibleRoleNames...) {
		if strings.EqualFold(strings.TrimSpace(roleName), want) {
			return true
		}
	}
	return false
}

func lighthouseSeen(seen map[string]struct{}, assignmentID string) bool {
	if _, ok := seen[assignmentID]; ok {
		return true
	}
	seen[assignmentID] = struct{}{}
	return false
}
