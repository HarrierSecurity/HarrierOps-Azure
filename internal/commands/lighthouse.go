package commands

import (
	"context"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func lighthouseHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Lighthouse(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		delegations := sortedByLess(facts.LighthouseDelegations, lighthouseLess)

		return models.LighthouseOutput{
			Findings:              []models.Finding{},
			Issues:                facts.Issues,
			LighthouseDelegations: delegations,
			Metadata:              commandMetadata("lighthouse", now, request, facts.TenantID, facts.SubscriptionID, ""),
		}, nil
	}
}

func lighthouseLess(left models.LighthouseDelegationAsset, right models.LighthouseDelegationAsset) bool {
	if left.ScopeType != right.ScopeType {
		return left.ScopeType == "subscription"
	}

	leftRoleRank, leftDelegatedRank := lighthouseRoleRank(left)
	rightRoleRank, rightDelegatedRank := lighthouseRoleRank(right)
	if leftRoleRank != rightRoleRank {
		return leftRoleRank < rightRoleRank
	}
	if leftDelegatedRank != rightDelegatedRank {
		return leftDelegatedRank < rightDelegatedRank
	}

	leftAuthZero := left.AuthorizationCount == 0
	rightAuthZero := right.AuthorizationCount == 0
	if leftAuthZero != rightAuthZero {
		return !leftAuthZero
	}
	if left.AuthorizationCount != right.AuthorizationCount {
		return left.AuthorizationCount > right.AuthorizationCount
	}
	if left.EligibleAuthorizationCount != right.EligibleAuthorizationCount {
		return left.EligibleAuthorizationCount > right.EligibleAuthorizationCount
	}

	leftAssignmentStateRank, leftDefinitionStateRank := lighthouseStateRank(left)
	rightAssignmentStateRank, rightDefinitionStateRank := lighthouseStateRank(right)
	if leftAssignmentStateRank != rightAssignmentStateRank {
		return leftAssignmentStateRank < rightAssignmentStateRank
	}
	if leftDefinitionStateRank != rightDefinitionStateRank {
		return leftDefinitionStateRank < rightDefinitionStateRank
	}

	leftManagedBy := firstNonEmpty(trimmedStringPtrValue(left.ManagedByTenantName), trimmedStringPtrValue(left.ManagedByTenantID))
	rightManagedBy := firstNonEmpty(trimmedStringPtrValue(right.ManagedByTenantName), trimmedStringPtrValue(right.ManagedByTenantID))
	if leftManagedBy != rightManagedBy {
		return leftManagedBy < rightManagedBy
	}

	leftScopeName := trimmedStringPtrValue(left.ScopeDisplayName)
	rightScopeName := trimmedStringPtrValue(right.ScopeDisplayName)
	if leftScopeName != rightScopeName {
		return leftScopeName < rightScopeName
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func lighthouseRoleRank(item models.LighthouseDelegationAsset) (int, int) {
	roleName := normalizedLower(trimmedStringPtrValue(item.StrongestRoleName))

	roleRank := 5
	switch {
	case item.HasOwnerRole:
		roleRank = 0
	case item.HasUserAccessAdministrator:
		roleRank = 1
	case roleName == "contributor":
		roleRank = 2
	case roleName != "" && roleName != "reader":
		roleRank = 3
	case roleName == "reader":
		roleRank = 4
	}

	delegatedRank := 1
	if item.HasDelegatedRoleAssignments {
		delegatedRank = 0
	}
	return roleRank, delegatedRank
}

func lighthouseStateRank(item models.LighthouseDelegationAsset) (int, int) {
	assignmentRank := 1
	if state := normalizedLower(trimmedStringPtrValue(item.ProvisioningState)); state != "" && state != "succeeded" {
		assignmentRank = 0
	}

	definitionRank := 1
	if state := normalizedLower(trimmedStringPtrValue(item.DefinitionProvisioningState)); state != "" && state != "succeeded" {
		definitionRank = 0
	}

	return assignmentRank, definitionRank
}

func trimmedStringPtrValue(value *string) string {
	return strings.TrimSpace(stringPtrValue(value))
}
