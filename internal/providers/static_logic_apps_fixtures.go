package providers

import (
	"strings"

	"harrierops-azure/internal/models"
)

const (
	staticLogicAppsWorkflowResourceGroup = "rg-workflow"
	staticLogicAppsIdentityResourceGroup = "rg-identity"
)

type staticLogicAppWorkflowFixture struct {
	name                             string
	classification                   string
	location                         string
	platform                         string
	state                            string
	triggerTypes                     []string
	recurrenceSummary                string
	externallyCallableRequestTrigger bool
	downstreamActionKinds            []string
	summary                          string
	identity                         *staticLogicAppIdentityFixture
}

type staticLogicAppIdentityFixture struct {
	workflowName          string
	displayName           string
	principalID           string
	clientID              string
	workflowIdentityType  string
	principalIdentityType string
	userAssignedName      string
	userAssigned          bool
	roleAssignmentID      string
	roleName              string
}

func staticLogicAppWorkflowFixtures() []staticLogicAppWorkflowFixture {
	return []staticLogicAppWorkflowFixture{
		{
			name:                             "la-inbound-redeploy",
			classification:                   "persistence-capable",
			location:                         "centralus",
			platform:                         "Consumption",
			state:                            "Enabled",
			triggerTypes:                     []string{"request"},
			externallyCallableRequestTrigger: true,
			downstreamActionKinds:            []string{"automation", "external-http"},
			summary:                          "Request trigger is visible from workflow definition, so this Logic App already looks like a callable re-entry path. Workflow uses managed identity (SystemAssigned), and visible actions touch Azure Automation and external HTTP destinations.",
			identity: &staticLogicAppIdentityFixture{
				workflowName:          "la-inbound-redeploy",
				displayName:           "la-inbound-redeploy-identity",
				principalID:           "56565656-5656-5656-5656-565656565656",
				clientID:              "78787878-7878-7878-7878-787878787878",
				workflowIdentityType:  "SystemAssigned",
				principalIdentityType: "systemAssigned",
				roleAssignmentID:      "ra-3",
				roleName:              "Contributor",
			},
		},
		{
			name:                  "la-nightly-sync",
			classification:        "persistence-capable",
			location:              "centralus",
			platform:              "Consumption",
			state:                 "Enabled",
			triggerTypes:          []string{"recurrence"},
			recurrenceSummary:     "Day/1",
			downstreamActionKinds: []string{"storage", "connector"},
			summary:               "Recurrence is visible from workflow definition (Day/1), so Azure already has a durable schedule for this workflow. Visible downstream actions touch storage and connector-backed service paths, but no workflow identity is exposed from the current read path.",
		},
		{
			name:                  "la-event-router",
			classification:        "execution-capable-only",
			location:              "eastus",
			platform:              "Consumption",
			state:                 "Enabled",
			triggerTypes:          []string{"api-connection"},
			downstreamActionKinds: []string{"function", "messaging"},
			summary:               "Visible trigger and action posture suggest workflow-driven execution, but the current definition does not yet show a durable request or recurrence trigger. Workflow uses a user-assigned managed identity and visibly reaches Azure Functions and messaging paths.",
			identity: &staticLogicAppIdentityFixture{
				workflowName:          "la-event-router",
				displayName:           "ua-workflow-router",
				principalID:           "90909090-9090-9090-9090-909090909090",
				clientID:              "abababab-abab-abab-abab-abababababab",
				workflowIdentityType:  "UserAssigned",
				principalIdentityType: "userAssigned",
				userAssignedName:      "ua-workflow-router",
				userAssigned:          true,
				roleAssignmentID:      "ra-4",
				roleName:              "Contributor",
			},
		},
	}
}

func staticLogicAppIdentityFixtures() []staticLogicAppIdentityFixture {
	fixtures := []staticLogicAppIdentityFixture{}
	for _, workflow := range staticLogicAppWorkflowFixtures() {
		if workflow.identity != nil {
			fixtures = append(fixtures, *workflow.identity)
		}
	}
	return fixtures
}

func staticLogicAppWorkflowID(subscriptionID string, workflowName string) string {
	return "/subscriptions/" + subscriptionID + "/resourceGroups/" + staticLogicAppsWorkflowResourceGroup +
		"/providers/Microsoft.Logic/workflows/" + workflowName
}

func staticLogicAppIdentityID(subscriptionID string, fixture staticLogicAppIdentityFixture) string {
	if fixture.userAssigned {
		return "/subscriptions/" + subscriptionID + "/resourceGroups/" + staticLogicAppsIdentityResourceGroup +
			"/providers/Microsoft.ManagedIdentity/userAssignedIdentities/" + fixture.userAssignedName
	}
	return staticLogicAppWorkflowID(subscriptionID, fixture.workflowName) + "/identities/system"
}

func staticLogicAppScopeID(subscriptionID string) string {
	return "/subscriptions/" + subscriptionID + "/resourceGroups/" + staticLogicAppsWorkflowResourceGroup
}

func staticLogicAppRelatedIDs(subscriptionID string, workflow staticLogicAppWorkflowFixture) []string {
	relatedIDs := []string{staticLogicAppWorkflowID(subscriptionID, workflow.name)}
	if workflow.identity != nil {
		relatedIDs = append(relatedIDs, staticLogicAppIdentityID(subscriptionID, *workflow.identity))
	}
	return relatedIDs
}

func staticLogicAppOperatorSignal() string {
	return "Internal Logic App workload pivot; direct control visible."
}

func staticLogicAppNextReview() string {
	return "Check permissions for direct control on this identity, then logic-apps for the workflow context behind this workload pivot."
}

func staticLogicAppManagedIdentitySummary(workflowName string, identityName string, roleName string) string {
	return "Logic App '" + workflowName + "' exposes managed identity '" + identityName +
		"'. Current scope already shows direct control through high-impact roles (" + roleName + "). " +
		staticLogicAppNextReview()
}

func staticLogicAppFindingDescription(identityName string, roleName string) string {
	return "Identity '" + identityName + "' is assigned one or more high-impact roles (" + roleName + ")."
}

func staticLogicAppRecurrenceSummaryPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return models.StringPtr(value)
}

func staticLogicAppRBACPrincipals(tenantID string) []models.Principal {
	principals := []models.Principal{}
	for _, fixture := range staticLogicAppIdentityFixtures() {
		principals = append(principals, models.Principal{
			DisplayName:   fixture.displayName,
			ID:            fixture.principalID,
			PrincipalType: "ServicePrincipal",
			TenantID:      tenantID,
		})
	}
	return principals
}

func staticLogicAppRBACRoleAssignments(subscriptionID string) []models.RoleAssignment {
	assignments := []models.RoleAssignment{}
	for _, fixture := range staticLogicAppIdentityFixtures() {
		assignments = append(assignments, models.RoleAssignment{
			ID:               fixture.roleAssignmentID,
			PrincipalID:      fixture.principalID,
			PrincipalType:    "ServicePrincipal",
			RoleDefinitionID: "rd-contributor",
			RoleName:         fixture.roleName,
			ScopeID:          staticLogicAppScopeID(subscriptionID),
		})
	}
	return assignments
}

func staticLogicAppPermissionFacts(subscriptionID string) []PermissionFact {
	facts := []PermissionFact{}
	for _, fixture := range staticLogicAppIdentityFixtures() {
		facts = append(facts, PermissionFact{
			PrincipalID:         fixture.principalID,
			DisplayName:         fixture.displayName,
			PrincipalType:       "ServicePrincipal",
			HighImpactRoles:     []string{fixture.roleName},
			AllRoleNames:        []string{fixture.roleName},
			RoleAssignmentCount: 1,
			ScopeCount:          1,
			ScopeIDs:            []string{staticLogicAppScopeID(subscriptionID)},
			Privileged:          true,
			IsCurrentIdentity:   false,
		})
	}
	return facts
}

func staticLogicAppPermissionPrincipals(subscriptionID string) []PermissionPrincipalFact {
	principals := []PermissionPrincipalFact{}
	for _, fixture := range staticLogicAppIdentityFixtures() {
		principals = append(principals, PermissionPrincipalFact{
			ID:            fixture.principalID,
			Sources:       []string{"managed-identities"},
			IdentityNames: []string{fixture.displayName},
			AttachedTo: []string{
				staticLogicAppWorkflowID(subscriptionID, fixture.workflowName),
			},
		})
	}
	return principals
}

func staticLogicAppPrincipalSummaries(tenantID string, subscriptionID string) []models.PrincipalSummary {
	principals := []models.PrincipalSummary{}
	for _, fixture := range staticLogicAppIdentityFixtures() {
		principals = append(principals, models.PrincipalSummary{
			AttachedTo: []string{
				staticLogicAppWorkflowID(subscriptionID, fixture.workflowName),
			},
			DisplayName:         models.StringPtr(fixture.displayName),
			ID:                  fixture.principalID,
			IdentityNames:       []string{fixture.displayName},
			IdentityTypes:       []string{fixture.principalIdentityType},
			IsCurrentIdentity:   false,
			PrincipalType:       "ServicePrincipal",
			RoleAssignmentCount: 1,
			RoleNames:           []string{fixture.roleName},
			ScopeIDs:            []string{staticLogicAppScopeID(subscriptionID)},
			Sources:             []string{"managed-identities"},
			TenantID:            models.StringPtr(tenantID),
		})
	}
	return principals
}

func staticLogicAppManagedIdentities(subscriptionID string) []models.ManagedIdentity {
	identities := []models.ManagedIdentity{}
	for _, fixture := range staticLogicAppIdentityFixtures() {
		identities = append(identities, models.ManagedIdentity{
			ID:           staticLogicAppIdentityID(subscriptionID, fixture),
			Name:         fixture.displayName,
			IdentityType: fixture.principalIdentityType,
			PrincipalID:  models.StringPtr(fixture.principalID),
			ClientID:     models.StringPtr(fixture.clientID),
			AttachedTo: []string{
				staticLogicAppWorkflowID(subscriptionID, fixture.workflowName),
			},
			ScopeIDs:             []string{staticLogicAppScopeID(subscriptionID)},
			OperatorSignal:       models.StringPtr(staticLogicAppOperatorSignal()),
			NextReview:           models.StringPtr(staticLogicAppNextReview()),
			Summary:              models.StringPtr(staticLogicAppManagedIdentitySummary(fixture.workflowName, fixture.displayName, fixture.roleName)),
			WorkloadExposure:     models.WorkloadExposureNone,
			DirectControlVisible: true,
		})
	}
	return identities
}

func staticLogicAppManagedIdentityRoleAssignments(subscriptionID string) []models.ManagedIdentityRoleAssignment {
	assignments := []models.ManagedIdentityRoleAssignment{}
	for _, fixture := range staticLogicAppIdentityFixtures() {
		assignments = append(assignments, models.ManagedIdentityRoleAssignment{
			ID:               fixture.roleAssignmentID,
			ScopeID:          staticLogicAppScopeID(subscriptionID),
			PrincipalID:      fixture.principalID,
			PrincipalType:    "ServicePrincipal",
			RoleDefinitionID: "rd-contributor",
			RoleName:         fixture.roleName,
		})
	}
	return assignments
}

func staticLogicAppManagedIdentityFindings(subscriptionID string) []models.ManagedIdentityFinding {
	findings := []models.ManagedIdentityFinding{}
	for _, fixture := range staticLogicAppIdentityFixtures() {
		findings = append(findings, models.ManagedIdentityFinding{
			ID:          "identity-privileged-" + staticLogicAppIdentityID(subscriptionID, fixture),
			Severity:    "high",
			Title:       "Managed identity has elevated role assignment",
			Description: staticLogicAppFindingDescription(fixture.displayName, fixture.roleName),
			RelatedIDs: []string{
				staticLogicAppIdentityID(subscriptionID, fixture),
				fixture.roleAssignmentID,
			},
		})
	}
	return findings
}
