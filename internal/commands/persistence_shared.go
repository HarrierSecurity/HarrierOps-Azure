package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

type persistenceCurrentIdentityControl struct {
	RoleName string
	ScopeID  string
}

type persistenceBackingFutures struct {
	managedIdentities asyncCommandOutput[models.ManagedIdentitiesOutput]
	permissions       asyncCommandOutput[models.PermissionsOutput]
	rbac              asyncCommandOutput[models.RbacOutput]
}

type persistenceBackingData struct {
	managedIdentities models.ManagedIdentitiesOutput
	permissions       models.PermissionsOutput
	rbac              models.RbacOutput
	evidence          persistencePrincipalEvidence
	tenantID          string
	subscriptionID    string
	issues            []models.Issue
}

func startPersistenceBackingFutures(
	group commandOutputGroup,
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
) persistenceBackingFutures {
	return persistenceBackingFutures{
		managedIdentities: runGroupedCommandOutput[models.ManagedIdentitiesOutput](group, ctx, request, managedIdentitiesHandler(provider, now), "managed-identities"),
		permissions:       runGroupedCommandOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions"),
		rbac:              runGroupedCommandOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac"),
	}
}

func (futures persistenceBackingFutures) wait(
	request Request,
	primaryTenantID *string,
	primarySubscriptionID *string,
	primaryIssues []models.Issue,
) (persistenceBackingData, error) {
	managedIdentities, err := futures.managedIdentities.wait()
	if err != nil {
		return persistenceBackingData{}, err
	}
	permissions, err := futures.permissions.wait()
	if err != nil {
		return persistenceBackingData{}, err
	}
	rbac, err := futures.rbac.wait()
	if err != nil {
		return persistenceBackingData{}, err
	}

	issues := append([]models.Issue{}, primaryIssues...)
	issues = append(issues, managedIdentities.Issues...)
	issues = append(issues, permissions.Issues...)
	issues = append(issues, rbac.Issues...)

	return persistenceBackingData{
		managedIdentities: managedIdentities,
		permissions:       permissions,
		rbac:              rbac,
		evidence:          buildPersistencePrincipalEvidence(permissions.Permissions, rbac.RoleAssignments),
		tenantID: firstNonEmpty(
			request.Tenant,
			stringPtrValue(primaryTenantID),
			stringPtrValue(permissions.Metadata.TenantID),
		),
		subscriptionID: firstNonEmpty(
			request.Subscription,
			stringPtrValue(primarySubscriptionID),
			stringPtrValue(permissions.Metadata.SubscriptionID),
		),
		issues: issues,
	}, nil
}

func persistenceAutomationControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	bestRank := 99
	best := persistenceCurrentIdentityControl{}
	for _, assignment := range assignments {
		role := strings.ToLower(strings.TrimSpace(assignment.RoleName))
		if role != "owner" && role != "contributor" && role != "automation contributor" {
			continue
		}
		rank, ok := persistenceScopeRank(assignment.ScopeID, resourceID)
		if !ok || rank >= bestRank {
			continue
		}
		bestRank = rank
		best = persistenceCurrentIdentityControl{
			RoleName: fmt.Sprintf("%s at %s", assignment.RoleName, persistenceScopeLabel(assignment.ScopeID)),
			ScopeID:  assignment.ScopeID,
		}
	}
	return best, bestRank != 99
}

func persistenceRoleAssignmentAllowsManagementAction(assignment models.RoleAssignment, targetAction string) bool {
	targetAction = strings.ToLower(strings.TrimSpace(targetAction))
	if targetAction == "" || len(assignment.Actions) == 0 {
		return false
	}
	for _, notAction := range assignment.NotActions {
		if persistenceAzureActionPatternMatches(notAction, targetAction) {
			return false
		}
	}
	for _, action := range assignment.Actions {
		if persistenceAzureActionPatternMatches(action, targetAction) {
			return true
		}
	}
	return false
}

func persistenceRoleAssignmentAllowsNamedOrActionControl(assignment models.RoleAssignment, targetAction string, roleNames ...string) bool {
	targetAction = strings.ToLower(strings.TrimSpace(targetAction))
	for _, notAction := range assignment.NotActions {
		if persistenceAzureActionPatternMatches(notAction, targetAction) {
			return false
		}
	}
	if persistenceRoleAssignmentAllowsManagementAction(assignment, targetAction) {
		return true
	}
	role := strings.ToLower(strings.TrimSpace(assignment.RoleName))
	for _, roleName := range roleNames {
		if role == strings.ToLower(strings.TrimSpace(roleName)) {
			return true
		}
	}
	return false
}

func persistenceAzureActionPatternMatches(pattern string, targetAction string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	targetAction = strings.ToLower(strings.TrimSpace(targetAction))
	if pattern == "" || targetAction == "" {
		return false
	}
	if pattern == "*" {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return pattern == targetAction
	}
	parts := strings.Split(pattern, "*")
	position := 0
	if !strings.HasPrefix(pattern, "*") {
		first := parts[0]
		if !strings.HasPrefix(targetAction, first) {
			return false
		}
		position = len(first)
		parts = parts[1:]
	}
	for index, part := range parts {
		if part == "" {
			continue
		}
		foundAt := strings.Index(targetAction[position:], part)
		if foundAt < 0 {
			return false
		}
		position += foundAt + len(part)
		if index == len(parts)-1 && !strings.HasSuffix(pattern, "*") && position != len(targetAction) {
			return false
		}
	}
	return strings.HasSuffix(pattern, "*") || position == len(targetAction)
}

func persistenceScopeRank(scopeID string, resourceID string) (int, bool) {
	scopeID = strings.TrimSpace(scopeID)
	resourceID = strings.TrimSpace(resourceID)
	if scopeID == "" || resourceID == "" {
		return 0, false
	}
	scopeLower := strings.ToLower(strings.TrimRight(scopeID, "/"))
	resourceLower := strings.ToLower(strings.TrimRight(resourceID, "/"))
	if scopeLower == resourceLower {
		return 0, true
	}
	if !strings.HasPrefix(resourceLower, scopeLower+"/") {
		return 0, false
	}
	if strings.Contains(scopeLower, "/resourcegroups/") {
		return 1, true
	}
	if strings.Contains(scopeLower, "/subscriptions/") {
		return 2, true
	}
	return 3, true
}

func persistenceScopeLabel(scopeID string) string {
	if strings.Contains(scopeID, "/subscriptions/") && !strings.Contains(scopeID, "/resourceGroups/") {
		return "subscription scope"
	}
	if strings.Contains(scopeID, "/resourceGroups/") {
		return "resource group " + armScopeName(scopeID)
	}
	return "a parent scope of this resource"
}
