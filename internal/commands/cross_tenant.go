package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func crossTenantHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.CrossTenant(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		return models.CrossTenantOutput{
			CrossTenantPaths: sortedByLess(facts.CrossTenantPaths, crossTenantLess),
			Findings:         []models.Finding{},
			Issues:           facts.Issues,
			Metadata:         commandMetadata("cross-tenant", now, request, facts.TenantID, facts.SubscriptionID, ""),
		}, nil
	}
}

func crossTenantLess(left models.CrossTenantPathSummary, right models.CrossTenantPathSummary) bool {
	leftPriority := crossTenantPriorityRank(left.Priority)
	rightPriority := crossTenantPriorityRank(right.Priority)
	if leftPriority != rightPriority {
		return leftPriority < rightPriority
	}

	leftSignal := crossTenantSignalRank(left.SignalType)
	rightSignal := crossTenantSignalRank(right.SignalType)
	if leftSignal != rightSignal {
		return leftSignal < rightSignal
	}

	leftScope := crossTenantScopeRank(left)
	rightScope := crossTenantScopeRank(right)
	if leftScope != rightScope {
		return leftScope < rightScope
	}

	leftTenant := firstNonEmpty(stringPtrValue(left.TenantName), stringPtrValue(left.TenantID))
	rightTenant := firstNonEmpty(stringPtrValue(right.TenantName), stringPtrValue(right.TenantID))
	if leftTenant != rightTenant {
		return leftTenant < rightTenant
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func crossTenantPriorityRank(priority string) int {
	switch normalizedLower(priority) {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	default:
		return 9
	}
}

func crossTenantSignalRank(signalType string) int {
	switch normalizedLower(signalType) {
	case "lighthouse":
		return 0
	case "external-sp":
		return 1
	case "policy":
		return 2
	default:
		return 9
	}
}

func crossTenantScopeRank(item models.CrossTenantPathSummary) int {
	if normalizedLower(item.SignalType) != "lighthouse" {
		return 9
	}

	scope := normalizedLower(stringPtrValue(item.Scope))
	switch {
	case hasPrefixNormalized(scope, "subscription::"):
		return 0
	case hasPrefixNormalized(scope, "resource-group::"):
		return 1
	default:
		return 2
	}
}

func hasPrefixNormalized(value string, prefix string) bool {
	return len(value) >= len(prefix) && value[:len(prefix)] == prefix
}
