package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func applicationGatewayHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.ApplicationGateway(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		gateways := sortedByLess(facts.ApplicationGateways, applicationGatewayLess)

		return models.ApplicationGatewayOutput{
			ApplicationGateways: gateways,
			Findings:            []models.Finding{},
			Issues:              facts.Issues,
			Metadata:            runtimeCommandMetadata("application-gateway", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func applicationGatewayLess(left models.ApplicationGatewayAsset, right models.ApplicationGatewayAsset) bool {
	if (left.PublicFrontendCount == 0) != (right.PublicFrontendCount == 0) {
		return left.PublicFrontendCount > 0
	}
	if left.PublicFrontendCount != right.PublicFrontendCount {
		return left.PublicFrontendCount > right.PublicFrontendCount
	}
	if applicationGatewaySharedBreadth(left) != applicationGatewaySharedBreadth(right) {
		return applicationGatewaySharedBreadth(left) > applicationGatewaySharedBreadth(right)
	}
	if left.ListenerCount != right.ListenerCount {
		return left.ListenerCount > right.ListenerCount
	}
	if left.RequestRoutingRuleCount != right.RequestRoutingRuleCount {
		return left.RequestRoutingRuleCount > right.RequestRoutingRuleCount
	}
	if left.BackendTargetCount != right.BackendTargetCount {
		return left.BackendTargetCount > right.BackendTargetCount
	}
	if left.BackendPoolCount != right.BackendPoolCount {
		return left.BackendPoolCount > right.BackendPoolCount
	}
	if applicationGatewayWAFRank(left) != applicationGatewayWAFRank(right) {
		return applicationGatewayWAFRank(left) < applicationGatewayWAFRank(right)
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func applicationGatewaySharedBreadth(item models.ApplicationGatewayAsset) int {
	return item.ListenerCount + item.RequestRoutingRuleCount + item.BackendTargetCount + item.BackendPoolCount
}

func applicationGatewayWAFRank(item models.ApplicationGatewayAsset) int {
	if item.FirewallPolicyID != nil && *item.FirewallPolicyID != "" {
		switch normalizedLower(applicationGatewayModeValue(item.WAFMode)) {
		case "prevention":
			return 3
		case "detection":
			return 1
		default:
			return 2
		}
	}
	if item.WAFEnabled != nil && !*item.WAFEnabled {
		return 0
	}
	switch normalizedLower(applicationGatewayModeValue(item.WAFMode)) {
	case "prevention":
		return 3
	case "detection":
		return 1
	default:
		if item.WAFEnabled != nil && *item.WAFEnabled {
			return 2
		}
		return 4
	}
}

func applicationGatewayModeValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
