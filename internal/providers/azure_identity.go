package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"

	"harrierops-azure/internal/models"
)

const (
	armManagementScope           = "https://management.azure.com/.default"
	azureResourceManagerEndpoint = "https://management.azure.com"
)

var highImpactRoleNames = map[string]struct{}{
	"owner":                     {},
	"contributor":               {},
	"user access administrator": {},
}

var builtInHighImpactRolesByID = map[string]string{
	"8e3af657-a8ff-443c-a75c-2fe8c4bcb635": "Owner",
	"b24988ac-6180-42a0-ab88-20f7382dd24c": "Contributor",
	"18d7d88d-d35e-4fb5-a5c3-7773c20a72d9": "User Access Administrator",
}

type livePrincipalRecord struct {
	id                  string
	principalType       string
	displayName         string
	sources             []string
	scopeIDs            []string
	roleNames           []string
	roleAssignmentCount int
	identityNames       []string
	attachedTo          []string
	isCurrentIdentity   bool
}

type userAssignedIdentityDetails struct {
	principalID *string
	clientID    *string
}

func (provider AzureProvider) RBAC(ctx context.Context, tenant string, subscription string) (RBACFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return RBACFacts{}, err
	}

	return provider.collectRBACFacts(ctx, session), nil
}

func (provider AzureProvider) Permissions(ctx context.Context, tenant string, subscription string) (PermissionsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return PermissionsFacts{}, err
	}

	rbacFacts := provider.collectRBACFacts(ctx, session)
	managedIdentityFacts := provider.collectManagedIdentityFacts(ctx, tenant, subscription, session, rbacFacts)
	whoamiFacts, err := provider.WhoAmI(ctx, tenant, subscription)
	if err != nil {
		return PermissionsFacts{}, err
	}

	principalRecords := map[string]livePrincipalRecord{}
	ensureRecord := func(principalID string) livePrincipalRecord {
		record, ok := principalRecords[principalID]
		if ok {
			return record
		}
		return livePrincipalRecord{
			id:            principalID,
			principalType: "unknown",
		}
	}

	for _, principal := range rbacFacts.Principals {
		if principal.ID == "" {
			continue
		}
		record := ensureRecord(principal.ID)
		record.principalType = normalizePrincipalType(record.principalType, principal.PrincipalType)
		if record.displayName == "" {
			record.displayName = principal.DisplayName
		}
		record.sources = appendUniqueString(record.sources, "rbac")
		principalRecords[principal.ID] = record
	}

	for _, assignment := range rbacFacts.RoleAssignments {
		if assignment.PrincipalID == "" {
			continue
		}
		record := ensureRecord(assignment.PrincipalID)
		record.principalType = normalizePrincipalType(record.principalType, assignment.PrincipalType)
		record.roleNames = appendUniqueString(record.roleNames, assignment.RoleName)
		record.scopeIDs = appendUniqueString(record.scopeIDs, assignment.ScopeID)
		record.roleAssignmentCount++
		record.sources = appendUniqueString(record.sources, "rbac")
		principalRecords[assignment.PrincipalID] = record
	}

	if whoamiFacts.Principal.ID != "" {
		record := ensureRecord(whoamiFacts.Principal.ID)
		record.principalType = normalizePrincipalType(record.principalType, whoamiFacts.Principal.PrincipalType)
		if record.displayName == "" {
			record.displayName = whoamiFacts.Principal.DisplayName
		}
		record.isCurrentIdentity = true
		record.sources = appendUniqueString(record.sources, "whoami")
		for _, scope := range whoamiFacts.EffectiveScopes {
			record.scopeIDs = appendUniqueString(record.scopeIDs, scope.ID)
		}
		principalRecords[whoamiFacts.Principal.ID] = record
	}

	for _, identity := range managedIdentityFacts.Identities {
		if identity.PrincipalID == nil || *identity.PrincipalID == "" {
			continue
		}
		record := ensureRecord(*identity.PrincipalID)
		record.principalType = normalizePrincipalType(record.principalType, "ManagedIdentity")
		if record.displayName == "" {
			record.displayName = identity.Name
		}
		record.identityNames = appendUniqueString(record.identityNames, identity.Name)
		for _, scopeID := range identity.ScopeIDs {
			record.scopeIDs = appendUniqueString(record.scopeIDs, scopeID)
		}
		for _, attachedID := range identity.AttachedTo {
			record.attachedTo = appendUniqueString(record.attachedTo, attachedID)
		}
		record.sources = appendUniqueString(record.sources, "managed-identities")
		principalRecords[*identity.PrincipalID] = record
	}

	permissions := make([]PermissionFact, 0, len(principalRecords))
	principals := make([]PermissionPrincipalFact, 0, len(principalRecords))
	for _, record := range principalRecords {
		roleNames := sortedUniqueStrings(record.roleNames)
		scopeIDs := sortedUniqueStrings(record.scopeIDs)
		highImpactRoles := highImpactRoleNamesFromRoles(roleNames)
		permissions = append(permissions, PermissionFact{
			PrincipalID:         record.id,
			DisplayName:         firstNonEmpty(record.displayName, record.id),
			PrincipalType:       firstNonEmpty(record.principalType, "unknown"),
			HighImpactRoles:     highImpactRoles,
			AllRoleNames:        roleNames,
			RoleAssignmentCount: record.roleAssignmentCount,
			ScopeCount:          len(scopeIDs),
			ScopeIDs:            scopeIDs,
			Privileged:          len(highImpactRoles) > 0,
			IsCurrentIdentity:   record.isCurrentIdentity,
		})
		principals = append(principals, PermissionPrincipalFact{
			ID:            record.id,
			Sources:       sortedUniqueStrings(record.sources),
			IdentityNames: sortedUniqueStrings(record.identityNames),
			AttachedTo:    sortedUniqueStrings(record.attachedTo),
		})
	}

	sort.SliceStable(permissions, func(i int, j int) bool {
		left := permissions[i]
		right := permissions[j]
		switch {
		case left.Privileged != right.Privileged:
			return left.Privileged
		case left.IsCurrentIdentity != right.IsCurrentIdentity:
			return left.IsCurrentIdentity
		case left.RoleAssignmentCount != right.RoleAssignmentCount:
			return left.RoleAssignmentCount > right.RoleAssignmentCount
		case left.DisplayName != right.DisplayName:
			return left.DisplayName < right.DisplayName
		default:
			return left.PrincipalID < right.PrincipalID
		}
	})
	sort.SliceStable(principals, func(i int, j int) bool {
		if principals[i].ID != principals[j].ID {
			return principals[i].ID < principals[j].ID
		}
		return len(principals[i].AttachedTo) > len(principals[j].AttachedTo)
	})

	issues := append([]models.Issue{}, rbacFacts.Issues...)
	issues = append(issues, managedIdentityFacts.Issues...)
	issues = append(issues, whoamiFacts.Issues...)

	return PermissionsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		Permissions:    permissions,
		Principals:     principals,
		Issues:         issues,
	}, nil
}

func (provider AzureProvider) ManagedIdentities(ctx context.Context, tenant string, subscription string) (ManagedIdentitiesFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return ManagedIdentitiesFacts{}, err
	}

	rbacFacts := provider.collectRBACFacts(ctx, session)
	return provider.collectManagedIdentityFacts(ctx, tenant, subscription, session, rbacFacts), nil
}

func (provider AzureProvider) collectRBACFacts(ctx context.Context, session azureSession) RBACFacts {
	subscriptionScope := "/subscriptions/" + session.subscription.ID
	clientFactory, err := armauthorization.NewClientFactory(session.subscription.ID, session.credential, nil)
	if err != nil {
		return RBACFacts{
			TenantID: session.tenantID,
			Scopes: []models.ScopeRef{
				{DisplayName: session.subscription.DisplayName, ID: subscriptionScope, ScopeType: "subscription"},
			},
			Issues: []models.Issue{issueFromError("rbac.authorization_client_factory", err)},
		}
	}

	assignments := []map[string]any{}
	assignmentsPager := clientFactory.NewRoleAssignmentsClient().NewListForSubscriptionPager(nil)
	for assignmentsPager.More() {
		page, pagerErr := assignmentsPager.NextPage(ctx)
		if pagerErr != nil {
			return RBACFacts{
				TenantID: session.tenantID,
				Scopes: []models.ScopeRef{
					{DisplayName: session.subscription.DisplayName, ID: subscriptionScope, ScopeType: "subscription"},
				},
				Issues: []models.Issue{issueFromError("rbac.role_assignments", pagerErr)},
			}
		}
		for _, assignment := range page.Value {
			if assignment == nil {
				continue
			}
			assignmentMap := map[string]any{}
			decodeJSONInto(assignment, &assignmentMap)
			assignments = append(assignments, assignmentMap)
		}
	}

	roleDefinitionsClient := clientFactory.NewRoleDefinitionsClient()

	currentPrincipalID, currentDisplayName := currentPrincipalFromClaims(session.claims)
	currentPrincipalType := principalTypeFromClaims(session.claims)

	scopes := map[string]models.ScopeRef{
		subscriptionScope: {
			DisplayName: session.subscription.DisplayName,
			ID:          subscriptionScope,
			ScopeType:   "subscription",
		},
	}
	principals := map[string]models.Principal{}
	roleAssignments := []models.RoleAssignment{}
	roleNameCache := map[string]string{}
	issues := []models.Issue{}

	for _, assignment := range assignments {
		properties := mapValue(assignment, "properties")
		principalID := mapStringValue(properties, "principalId")
		if principalID == "" {
			continue
		}

		assignmentScope := firstNonEmpty(mapStringValue(properties, "scope"), subscriptionScope)
		if _, ok := scopes[assignmentScope]; !ok {
			scopes[assignmentScope] = models.ScopeRef{
				ID:        assignmentScope,
				ScopeType: scopeTypeFromID(assignmentScope),
			}
		}

		roleDefinitionID := mapStringValue(properties, "roleDefinitionId")
		roleName, resolveErr := resolveRoleDefinitionName(ctx, roleDefinitionsClient, roleDefinitionID, roleNameCache)
		if resolveErr != nil {
			issues = append(issues, issueFromError("rbac.role_definition["+roleDefinitionID+"]", resolveErr))
		}

		principalType := mapStringValue(properties, "principalType")
		displayName := ""
		if principalID == currentPrincipalID {
			principalType = normalizePrincipalType(principalType, currentPrincipalType)
			displayName = currentDisplayName
		}
		principals[principalID] = models.Principal{
			DisplayName:   displayName,
			ID:            principalID,
			PrincipalType: firstNonEmpty(principalType, "unknown"),
			TenantID:      session.tenantID,
		}
		roleAssignments = append(roleAssignments, models.RoleAssignment{
			ID:               firstNonEmpty(mapStringValue(assignment, "id"), mapStringValue(assignment, "name")),
			PrincipalID:      principalID,
			PrincipalType:    firstNonEmpty(principalType, "unknown"),
			RoleDefinitionID: roleDefinitionID,
			RoleName:         roleName,
			ScopeID:          assignmentScope,
		})
	}

	principalRows := make([]models.Principal, 0, len(principals))
	for _, principal := range principals {
		principalRows = append(principalRows, principal)
	}
	sort.SliceStable(principalRows, func(i int, j int) bool {
		if principalRows[i].DisplayName != principalRows[j].DisplayName {
			return principalRows[i].DisplayName < principalRows[j].DisplayName
		}
		return principalRows[i].ID < principalRows[j].ID
	})

	scopeRows := make([]models.ScopeRef, 0, len(scopes))
	for _, scope := range scopes {
		scopeRows = append(scopeRows, scope)
	}
	sort.SliceStable(scopeRows, func(i int, j int) bool {
		if scopeRows[i].ScopeType != scopeRows[j].ScopeType {
			return scopeRank(scopeRows[i].ScopeType) < scopeRank(scopeRows[j].ScopeType)
		}
		return scopeRows[i].ID < scopeRows[j].ID
	})
	sort.SliceStable(roleAssignments, func(i int, j int) bool {
		left := roleAssignments[i]
		right := roleAssignments[j]
		switch {
		case left.ScopeID != right.ScopeID:
			return left.ScopeID < right.ScopeID
		case left.PrincipalID != right.PrincipalID:
			return left.PrincipalID < right.PrincipalID
		case left.RoleName != right.RoleName:
			return left.RoleName < right.RoleName
		default:
			return left.ID < right.ID
		}
	})

	return RBACFacts{
		TenantID:        session.tenantID,
		Principals:      principalRows,
		Scopes:          scopeRows,
		RoleAssignments: roleAssignments,
		Issues:          issues,
	}
}

func (provider AzureProvider) collectManagedIdentityFacts(ctx context.Context, tenant string, subscription string, session azureSession, rbacFacts RBACFacts) ManagedIdentitiesFacts {
	subscriptionScope := "/subscriptions/" + session.subscription.ID

	vmFacts, vmErr := provider.VMs(ctx, tenant, subscription)
	vmssFacts, vmssErr := provider.VMSS(ctx, tenant, subscription)
	appServiceFacts, appErr := provider.AppServices(ctx, tenant, subscription)
	functionFacts, functionErr := provider.Functions(ctx, tenant, subscription)
	logicAppsFacts, logicAppsErr := provider.LogicApps(ctx, tenant, subscription)
	userAssignedClient, userAssignedErr := armmsi.NewUserAssignedIdentitiesClient(session.subscription.ID, session.credential, nil)

	issues := append([]models.Issue{}, rbacFacts.Issues...)
	if vmErr != nil {
		issues = append(issues, issueFromError("managed-identities.vms", vmErr))
	}
	if vmssErr != nil {
		issues = append(issues, issueFromError("managed-identities.vmss", vmssErr))
	}
	if appErr != nil {
		issues = append(issues, issueFromError("managed-identities.app-services", appErr))
	}
	if functionErr != nil {
		issues = append(issues, issueFromError("managed-identities.functions", functionErr))
	}
	if logicAppsErr != nil {
		issues = append(issues, issueFromError("managed-identities.logic-apps", logicAppsErr))
	}
	if userAssignedErr != nil {
		issues = append(issues, issueFromError("managed-identities.user_assigned_client", userAssignedErr))
	}

	assignmentsByPrincipal := map[string][]models.RoleAssignment{}
	for _, assignment := range rbacFacts.RoleAssignments {
		assignmentsByPrincipal[assignment.PrincipalID] = append(assignmentsByPrincipal[assignment.PrincipalID], assignment)
	}

	userAssignedCache := map[string]userAssignedIdentityDetails{}
	identityMap := map[string]models.ManagedIdentity{}

	for _, vm := range vmFacts.VMAssets {
		exposure := models.WorkloadExposureNone
		if len(vm.PublicIPs) > 0 {
			exposure = models.WorkloadExposurePublic
		}
		for _, identityID := range vm.IdentityIDs {
			if strings.HasSuffix(identityID, "/identities/system") {
				identityMap[identityID] = managedIdentityFromAttachment(
					identityID,
					vm.Name+"-system",
					"systemAssigned",
					nil,
					nil,
					vm.ID,
					"VM",
					vm.Name,
					exposure,
					subscriptionScope,
					nil,
				)
				continue
			}
			details, detailIssues := loadUserAssignedIdentityDetails(ctx, userAssignedClient, identityID, userAssignedCache)
			issues = append(issues, detailIssues...)
			identityMap[identityID] = managedIdentityFromAttachment(
				identityID,
				resourceNameFromID(identityID),
				"userAssigned",
				details.principalID,
				details.clientID,
				vm.ID,
				"VM",
				vm.Name,
				exposure,
				subscriptionScope,
				assignmentsByPrincipal[stringPtrValue(details.principalID)],
			)
		}
	}

	for _, vmss := range vmssFacts.VMSSAssets {
		exposure := models.WorkloadExposureNone
		if vmss.PublicIPConfigurationCount > 0 || vmss.InboundNATPoolCount > 0 || vmss.LoadBalancerBackendPoolCount > 0 || vmss.ApplicationGatewayBackendPoolCount > 0 {
			exposure = models.WorkloadExposureExposed
		}
		if identityIncludesType(vmss.IdentityType, "SystemAssigned") {
			systemID := vmss.ID + "/identities/system"
			identityMap[systemID] = managedIdentityFromAttachment(
				systemID,
				vmss.Name+"-system",
				"systemAssigned",
				vmss.PrincipalID,
				vmss.ClientID,
				vmss.ID,
				"VMSS",
				vmss.Name,
				exposure,
				subscriptionScope,
				assignmentsByPrincipal[stringPtrValue(vmss.PrincipalID)],
			)
		}
		for _, identityID := range vmss.IdentityIDs {
			if strings.HasSuffix(identityID, "/identities/system") {
				continue
			}
			details, detailIssues := loadUserAssignedIdentityDetails(ctx, userAssignedClient, identityID, userAssignedCache)
			issues = append(issues, detailIssues...)
			identityMap[identityID] = managedIdentityFromAttachment(
				identityID,
				resourceNameFromID(identityID),
				"userAssigned",
				details.principalID,
				details.clientID,
				vmss.ID,
				"VMSS",
				vmss.Name,
				exposure,
				subscriptionScope,
				assignmentsByPrincipal[stringPtrValue(details.principalID)],
			)
		}
	}

	for _, app := range appServiceFacts.AppServices {
		exposure := models.WorkloadExposureNone
		if app.DefaultHostname != nil && *app.DefaultHostname != "" {
			exposure = models.WorkloadExposurePublic
		}
		if identityIncludesType(app.WorkloadIdentityType, "SystemAssigned") {
			systemID := app.ID + "/identities/system"
			identityMap[systemID] = managedIdentityFromAttachment(
				systemID,
				app.Name+"-system",
				"systemAssigned",
				app.WorkloadPrincipalID,
				app.WorkloadClientID,
				app.ID,
				"AppService",
				app.Name,
				exposure,
				subscriptionScope,
				assignmentsByPrincipal[stringPtrValue(app.WorkloadPrincipalID)],
			)
		}
		for _, identityID := range app.WorkloadIdentityIDs {
			details, detailIssues := loadUserAssignedIdentityDetails(ctx, userAssignedClient, identityID, userAssignedCache)
			issues = append(issues, detailIssues...)
			identityMap[identityID] = managedIdentityFromAttachment(
				identityID,
				resourceNameFromID(identityID),
				"userAssigned",
				details.principalID,
				details.clientID,
				app.ID,
				"AppService",
				app.Name,
				exposure,
				subscriptionScope,
				assignmentsByPrincipal[stringPtrValue(details.principalID)],
			)
		}
	}

	for _, app := range functionFacts.FunctionApps {
		exposure := models.WorkloadExposureNone
		if app.DefaultHostname != nil && *app.DefaultHostname != "" {
			exposure = models.WorkloadExposurePublic
		}
		if identityIncludesType(app.WorkloadIdentityType, "SystemAssigned") {
			systemID := app.ID + "/identities/system"
			identityMap[systemID] = managedIdentityFromAttachment(
				systemID,
				app.Name+"-system",
				"systemAssigned",
				app.WorkloadPrincipalID,
				app.WorkloadClientID,
				app.ID,
				"FunctionApp",
				app.Name,
				exposure,
				subscriptionScope,
				assignmentsByPrincipal[stringPtrValue(app.WorkloadPrincipalID)],
			)
		}
		for _, identityID := range app.WorkloadIdentityIDs {
			details, detailIssues := loadUserAssignedIdentityDetails(ctx, userAssignedClient, identityID, userAssignedCache)
			issues = append(issues, detailIssues...)
			identityMap[identityID] = managedIdentityFromAttachment(
				identityID,
				resourceNameFromID(identityID),
				"userAssigned",
				details.principalID,
				details.clientID,
				app.ID,
				"FunctionApp",
				app.Name,
				exposure,
				subscriptionScope,
				assignmentsByPrincipal[stringPtrValue(details.principalID)],
			)
		}
	}

	for _, workflow := range logicAppsFacts.Workflows {
		exposure := models.WorkloadExposureNone
		if identityIncludesType(workflow.IdentityType, "SystemAssigned") {
			systemID := workflow.ID + "/identities/system"
			identityMap[systemID] = managedIdentityFromAttachment(
				systemID,
				workflow.Name+"-identity",
				"systemAssigned",
				workflow.PrincipalID,
				workflow.ClientID,
				workflow.ID,
				"LogicApp",
				workflow.Name,
				exposure,
				subscriptionScope,
				assignmentsByPrincipal[stringPtrValue(workflow.PrincipalID)],
			)
		}
		for _, identityID := range workflow.IdentityIDs {
			if strings.HasSuffix(identityID, "/identities/system") {
				continue
			}
			details, detailIssues := loadUserAssignedIdentityDetails(ctx, userAssignedClient, identityID, userAssignedCache)
			issues = append(issues, detailIssues...)
			identityMap[identityID] = managedIdentityFromAttachment(
				identityID,
				resourceNameFromID(identityID),
				"userAssigned",
				details.principalID,
				details.clientID,
				workflow.ID,
				"LogicApp",
				workflow.Name,
				exposure,
				subscriptionScope,
				assignmentsByPrincipal[stringPtrValue(details.principalID)],
			)
		}
	}

	identities := make([]models.ManagedIdentity, 0, len(identityMap))
	principalIDs := map[string]struct{}{}
	for _, identity := range identityMap {
		identities = append(identities, identity)
		if identity.PrincipalID != nil && *identity.PrincipalID != "" {
			principalIDs[*identity.PrincipalID] = struct{}{}
		}
	}
	sort.SliceStable(identities, func(i int, j int) bool {
		left := identities[i]
		right := identities[j]
		switch {
		case left.DirectControlVisible != right.DirectControlVisible:
			return left.DirectControlVisible
		case managedIdentityExposureRank(left.WorkloadExposure) != managedIdentityExposureRank(right.WorkloadExposure):
			return managedIdentityExposureRank(left.WorkloadExposure) < managedIdentityExposureRank(right.WorkloadExposure)
		case left.Name != right.Name:
			return left.Name < right.Name
		default:
			return left.ID < right.ID
		}
	})

	roleAssignments := []models.ManagedIdentityRoleAssignment{}
	findings := []models.ManagedIdentityFinding{}
	for _, assignment := range rbacFacts.RoleAssignments {
		if _, ok := principalIDs[assignment.PrincipalID]; !ok {
			continue
		}
		roleAssignments = append(roleAssignments, models.ManagedIdentityRoleAssignment{
			ID:               assignment.ID,
			ScopeID:          assignment.ScopeID,
			PrincipalID:      assignment.PrincipalID,
			PrincipalType:    assignment.PrincipalType,
			RoleDefinitionID: assignment.RoleDefinitionID,
			RoleName:         assignment.RoleName,
		})
	}
	sort.SliceStable(roleAssignments, func(i int, j int) bool {
		switch {
		case roleAssignments[i].PrincipalID != roleAssignments[j].PrincipalID:
			return roleAssignments[i].PrincipalID < roleAssignments[j].PrincipalID
		case roleAssignments[i].RoleName != roleAssignments[j].RoleName:
			return roleAssignments[i].RoleName < roleAssignments[j].RoleName
		default:
			return roleAssignments[i].ScopeID < roleAssignments[j].ScopeID
		}
	})

	for _, identity := range identities {
		if identity.PrincipalID == nil || *identity.PrincipalID == "" {
			continue
		}
		assignments := assignmentsByPrincipal[*identity.PrincipalID]
		highImpactRoles := roleNamesForAssignments(assignments)
		if len(highImpactRoles) == 0 {
			continue
		}
		relatedIDs := []string{identity.ID}
		for _, assignment := range assignments {
			if isHighImpactRole(assignment.RoleName, assignment.RoleDefinitionID) {
				relatedIDs = append(relatedIDs, assignment.ID)
			}
		}
		findings = append(findings, models.ManagedIdentityFinding{
			ID:          "identity-privileged-" + identity.ID,
			Severity:    "high",
			Title:       "Managed identity has elevated role assignment",
			Description: "Identity '" + identity.Name + "' is assigned one or more high-impact roles (" + strings.Join(highImpactRoles, ", ") + ").",
			RelatedIDs:  dedupeStrings(relatedIDs),
		})
	}
	sort.SliceStable(findings, func(i int, j int) bool {
		return findings[i].ID < findings[j].ID
	})

	return ManagedIdentitiesFacts{
		TenantID:        session.tenantID,
		SubscriptionID:  session.subscription.ID,
		Identities:      identities,
		RoleAssignments: roleAssignments,
		Findings:        findings,
		Issues:          issues,
	}
}

func managedIdentityFromAttachment(identityID string, identityName string, identityType string, principalID *string, clientID *string, attachedTo string, attachedKind string, attachedName string, exposure models.WorkloadExposure, subscriptionScope string, assignments []models.RoleAssignment) models.ManagedIdentity {
	highImpactRoles := roleNamesForAssignments(assignments)
	directControlVisible := len(highImpactRoles) > 0
	visibilityBlocked := principalID == nil || *principalID == ""
	operatorSignal, nextReview, summary := managedIdentityNarrative(attachedKind, attachedName, identityName, exposure, directControlVisible, visibilityBlocked, highImpactRoles)
	scopeIDs := []string{subscriptionScope}
	if principalID != nil && *principalID != "" {
		scopeIDs = scopeIDs[:0]
		for _, assignment := range assignments {
			scopeIDs = append(scopeIDs, assignment.ScopeID)
		}
		if len(scopeIDs) == 0 {
			scopeIDs = []string{subscriptionScope}
		}
	}

	return models.ManagedIdentity{
		ID:                   identityID,
		Name:                 firstNonEmpty(identityName, resourceNameFromID(identityID)),
		IdentityType:         identityType,
		PrincipalID:          principalID,
		ClientID:             clientID,
		AttachedTo:           []string{attachedTo},
		ScopeIDs:             sortedUniqueStrings(scopeIDs),
		OperatorSignal:       stringPtr(operatorSignal),
		NextReview:           stringPtr(nextReview),
		Summary:              stringPtr(summary),
		WorkloadExposure:     exposure,
		DirectControlVisible: directControlVisible,
	}
}

func managedIdentityNarrative(attachedKind string, attachedName string, identityName string, exposure models.WorkloadExposure, directControlVisible bool, visibilityBlocked bool, highImpactRoles []string) (string, string, string) {
	exposureLabel := managedIdentityExposureLabel(attachedKind, exposure)
	switch attachedKind {
	case "VM":
		switch {
		case directControlVisible:
			nextReview := "Check permissions for direct control on this identity, then vms for the host context behind the workload pivot."
			return exposureLabel + "; direct control visible.", nextReview, "VM '" + attachedName + "' gives a " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "'. Current scope already shows direct control through high-impact roles (" + strings.Join(highImpactRoles, ", ") + "). " + nextReview
		case visibilityBlocked:
			nextReview := "Check vms for the host context behind this workload pivot; current scope does not yet show direct control on this identity."
			return exposureLabel + "; visibility blocked.", nextReview, "VM '" + attachedName + "' gives a " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "', but current scope does not show the backing principal cleanly. " + nextReview
		default:
			nextReview := "Check permissions for direct control on this identity, then vms for the host context behind the workload pivot."
			return exposureLabel + "; direct control not confirmed.", nextReview, "VM '" + attachedName + "' gives a " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "'. Current scope does not confirm direct control. " + nextReview
		}
	case "VMSS":
		switch {
		case directControlVisible:
			nextReview := "Check permissions for direct control on this identity, then vmss for the fleet context behind the workload pivot."
			return exposureLabel + "; direct control visible.", nextReview, "VMSS '" + attachedName + "' gives an " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "'. Current scope already shows direct control through high-impact roles (" + strings.Join(highImpactRoles, ", ") + "). " + nextReview
		case visibilityBlocked:
			nextReview := "Check vmss for the fleet context behind this workload pivot; current scope does not yet show direct control on this identity."
			return exposureLabel + "; visibility blocked.", nextReview, "VMSS '" + attachedName + "' gives an " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "', but current scope does not show the backing principal cleanly. " + nextReview
		default:
			nextReview := "Check vmss for the fleet context behind this workload pivot, then permissions to confirm direct control."
			return exposureLabel + "; direct control not confirmed.", nextReview, "VMSS '" + attachedName + "' gives an " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "'. Current scope does not confirm direct control. " + nextReview
		}
	case "LogicApp":
		switch {
		case directControlVisible:
			nextReview := "Check permissions for direct control on this identity, then logic-apps for the workflow context behind this workload pivot."
			return exposureLabel + "; direct control visible.", nextReview, "Logic App '" + attachedName + "' gives a " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "'. Current scope already shows direct control through high-impact roles (" + strings.Join(highImpactRoles, ", ") + "). " + nextReview
		case visibilityBlocked:
			nextReview := "Check logic-apps for the backing workflow context; current scope does not yet show direct control on this identity."
			return exposureLabel + "; visibility blocked.", nextReview, "Logic App '" + attachedName + "' gives a " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "', but current scope does not show the backing principal cleanly. " + nextReview
		default:
			nextReview := "Check logic-apps for the workflow context behind this workload pivot, then permissions to confirm direct control."
			return exposureLabel + "; direct control not confirmed.", nextReview, "Logic App '" + attachedName + "' gives a " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "'. Current scope does not confirm direct control. " + nextReview
		}
	default:
		workloadLabel := "App Service"
		if attachedKind == "FunctionApp" {
			workloadLabel = "Function App"
		}
		switch {
		case directControlVisible:
			nextReview := "Check permissions for direct control on this identity, then env-vars for secret-bearing config on this workload."
			return exposureLabel + "; direct control visible.", nextReview, workloadLabel + " '" + attachedName + "' gives a " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "'. Current scope already shows direct control through high-impact roles (" + strings.Join(highImpactRoles, ", ") + "). " + nextReview
		case visibilityBlocked:
			nextReview := "Check env-vars for the backing workload context; current scope does not yet show direct control on this identity."
			return exposureLabel + "; visibility blocked.", nextReview, workloadLabel + " '" + attachedName + "' gives a " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "', but current scope does not show the backing principal cleanly. " + nextReview
		default:
			nextReview := "Check env-vars for secret-bearing config on this workload, then permissions to confirm direct control."
			return exposureLabel + "; direct control not confirmed.", nextReview, workloadLabel + " '" + attachedName + "' gives a " + strings.ToLower(exposureLabel) + " into managed identity '" + identityName + "'. Current scope does not confirm direct control. " + nextReview
		}
	}
}

func managedIdentityExposureLabel(attachedKind string, exposure models.WorkloadExposure) string {
	switch exposure {
	case models.WorkloadExposurePublic:
		return "Public " + managedIdentityWorkloadLabel(attachedKind) + " workload pivot"
	case models.WorkloadExposureExposed:
		return "Exposed " + managedIdentityWorkloadLabel(attachedKind) + " workload pivot"
	default:
		return "Internal " + managedIdentityWorkloadLabel(attachedKind) + " workload pivot"
	}
}

func managedIdentityWorkloadLabel(attachedKind string) string {
	switch attachedKind {
	case "FunctionApp":
		return "Function App"
	case "AppService":
		return "App Service"
	case "LogicApp":
		return "Logic App"
	default:
		return attachedKind
	}
}

func managedIdentityExposureRank(value models.WorkloadExposure) int {
	switch value {
	case models.WorkloadExposurePublic:
		return 0
	case models.WorkloadExposureExposed:
		return 1
	default:
		return 2
	}
}

func roleNamesForAssignments(assignments []models.RoleAssignment) []string {
	names := []string{}
	for _, assignment := range assignments {
		if isHighImpactRole(assignment.RoleName, assignment.RoleDefinitionID) {
			name := assignment.RoleName
			if name == "" {
				name = builtInHighImpactRolesByID[roleDefinitionGUID(assignment.RoleDefinitionID)]
			}
			names = append(names, name)
		}
	}
	return sortedUniqueStrings(names)
}

func highImpactRoleNamesFromRoles(roleNames []string) []string {
	highImpactRoles := []string{}
	for _, roleName := range roleNames {
		if _, ok := highImpactRoleNames[strings.ToLower(strings.TrimSpace(roleName))]; ok {
			highImpactRoles = append(highImpactRoles, roleName)
		}
	}
	return sortedUniqueStrings(highImpactRoles)
}

func isHighImpactRole(roleName string, roleDefinitionID string) bool {
	if _, ok := highImpactRoleNames[strings.ToLower(strings.TrimSpace(roleName))]; ok {
		return true
	}
	_, ok := builtInHighImpactRolesByID[roleDefinitionGUID(roleDefinitionID)]
	return ok
}

func roleDefinitionGUID(roleDefinitionID string) string {
	parts := strings.Split(strings.Trim(roleDefinitionID, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return strings.ToLower(parts[len(parts)-1])
}

func resolveRoleDefinitionName(ctx context.Context, client *armauthorization.RoleDefinitionsClient, roleDefinitionID string, cache map[string]string) (string, error) {
	if roleDefinitionID == "" {
		return "", nil
	}
	cacheKey := strings.ToLower(roleDefinitionID)
	if roleName, ok := cache[cacheKey]; ok {
		return roleName, nil
	}
	if roleName, ok := builtInHighImpactRolesByID[roleDefinitionGUID(roleDefinitionID)]; ok {
		cache[cacheKey] = roleName
		return roleName, nil
	}

	if client == nil {
		return "", fmt.Errorf("role definitions client unavailable")
	}
	roleDefinitionResponse, err := client.GetByID(ctx, roleDefinitionID, nil)
	if err != nil {
		return "", err
	}
	roleDefinition := map[string]any{}
	decodeJSONInto(roleDefinitionResponse.RoleDefinition, &roleDefinition)
	roleName := firstNonEmpty(mapStringValue(mapValue(roleDefinition, "properties"), "roleName"), mapStringValue(roleDefinition, "name"))
	cache[cacheKey] = roleName
	return roleName, nil
}

func loadUserAssignedIdentityDetails(ctx context.Context, client *armmsi.UserAssignedIdentitiesClient, identityID string, cache map[string]userAssignedIdentityDetails) (userAssignedIdentityDetails, []models.Issue) {
	if details, ok := cache[identityID]; ok {
		return details, nil
	}

	if client == nil {
		return userAssignedIdentityDetails{}, nil
	}
	resourceGroup, identityName := resourceGroupAndNameFromID(identityID)
	if resourceGroup == "" || identityName == "" {
		return userAssignedIdentityDetails{}, nil
	}
	resourceResponse, err := client.Get(ctx, resourceGroup, identityName, nil)
	if err != nil {
		return userAssignedIdentityDetails{}, []models.Issue{issueFromError("managed-identities.user-assigned["+identityID+"]", err)}
	}
	resource := map[string]any{}
	decodeJSONInto(resourceResponse.Identity, &resource)
	properties := mapValue(resource, "properties")
	details := userAssignedIdentityDetails{
		principalID: stringPtr(mapStringValue(properties, "principalId")),
		clientID:    stringPtr(mapStringValue(properties, "clientId")),
	}
	cache[identityID] = details
	return details, nil
}

func armListObjects(ctx context.Context, credential azcore.TokenCredential, path string, apiVersion string) ([]map[string]any, error) {
	nextURL := armURL(path, apiVersion)
	items := []map[string]any{}
	for nextURL != "" {
		payload, err := authorizedJSONGet(ctx, credential, armManagementScope, nextURL)
		if err != nil {
			return nil, err
		}
		for _, item := range listValue(payload, "value") {
			if mapped, ok := item.(map[string]any); ok {
				items = append(items, mapped)
			}
		}
		nextURL = mapStringValue(payload, "nextLink")
	}
	return items, nil
}

func armGetObject(ctx context.Context, credential azcore.TokenCredential, path string, apiVersion string) (map[string]any, error) {
	return authorizedJSONGet(ctx, credential, armManagementScope, armURL(path, apiVersion))
}

func armURL(path string, apiVersion string) string {
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		if strings.Contains(path, "api-version=") {
			return path
		}
		separator := "?"
		if strings.Contains(path, "?") {
			separator = "&"
		}
		return path + separator + "api-version=" + url.QueryEscape(apiVersion)
	}
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	return azureResourceManagerEndpoint + path + separator + "api-version=" + url.QueryEscape(apiVersion)
}

func authorizedJSONGet(ctx context.Context, credential azcore.TokenCredential, scope string, rawURL string) (map[string]any, error) {
	token, err := accessToken(ctx, credential, scope)
	if err != nil {
		return nil, err
	}
	return authorizedJSONGetWithToken(ctx, token, rawURL)
}

func accessToken(ctx context.Context, credential azcore.TokenCredential, scope string) (string, error) {
	token, err := credential.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{scope}})
	if err != nil {
		return "", fmt.Errorf("get token: %w", err)
	}
	return token.Token, nil
}

func authorizedJSONGetWithToken(ctx context.Context, bearerToken string, rawURL string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s: %s", rawURL, resp.Status)
	}

	payload := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func normalizePrincipalType(current string, candidate string) string {
	current = strings.TrimSpace(current)
	candidate = strings.TrimSpace(candidate)

	if candidate == "" {
		if current == "" {
			return "unknown"
		}
		return current
	}
	if strings.EqualFold(current, "unknown") || current == "" {
		return candidate
	}
	if strings.EqualFold(current, candidate) {
		return current
	}
	if strings.EqualFold(current, "ServicePrincipal") && strings.EqualFold(candidate, "ManagedIdentity") {
		return candidate
	}
	if strings.EqualFold(current, "ManagedIdentity") && strings.EqualFold(candidate, "ServicePrincipal") {
		return current
	}
	return current
}

func scopeTypeFromID(scopeID string) string {
	lower := strings.ToLower(scopeID)
	switch {
	case strings.Contains(lower, "/providers/") && strings.Contains(lower, "/resourcegroups/"):
		return "resource"
	case strings.Contains(lower, "/resourcegroups/"):
		return "resource-group"
	case strings.HasPrefix(lower, "/subscriptions/"):
		return "subscription"
	default:
		return "unknown"
	}
}

func scopeRank(scopeType string) int {
	switch scopeType {
	case "subscription":
		return 0
	case "resource-group":
		return 1
	case "resource":
		return 2
	default:
		return 9
	}
}

func appendUniqueString(values []string, value string) []string {
	if strings.TrimSpace(value) == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func identityIncludesType(identityType *string, expected string) bool {
	return strings.Contains(strings.ToLower(stringPtrValue(identityType)), strings.ToLower(expected))
}
