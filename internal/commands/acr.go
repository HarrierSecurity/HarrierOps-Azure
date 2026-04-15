package commands

import (
	"context"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func acrHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Acr(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		registries := sortedByLess(facts.Registries, acrRegistryLess)

		return models.AcrOutput{
			Registries: registries,
			Findings:   []models.Finding{},
			Issues:     facts.Issues,
			Metadata:   runtimeCommandMetadata("acr", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func acrRegistryLess(left models.AcrRegistryAsset, right models.AcrRegistryAsset) bool {
	if acrRegistryPostureRank(left) != acrRegistryPostureRank(right) {
		return acrRegistryPostureRank(left) < acrRegistryPostureRank(right)
	}
	if acrIntValue(left.EnabledWebhookCount) != acrIntValue(right.EnabledWebhookCount) {
		return acrIntValue(left.EnabledWebhookCount) > acrIntValue(right.EnabledWebhookCount)
	}
	if acrIntValue(left.ReplicationCount) != acrIntValue(right.ReplicationCount) {
		return acrIntValue(left.ReplicationCount) > acrIntValue(right.ReplicationCount)
	}
	if acrGovernanceWeaknessScore(left) != acrGovernanceWeaknessScore(right) {
		return acrGovernanceWeaknessScore(left) > acrGovernanceWeaknessScore(right)
	}
	if acrIntValue(left.BroadWebhookScopeCount) != acrIntValue(right.BroadWebhookScopeCount) {
		return acrIntValue(left.BroadWebhookScopeCount) > acrIntValue(right.BroadWebhookScopeCount)
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func acrRegistryPostureRank(registry models.AcrRegistryAsset) int {
	publicEnabled := strings.EqualFold(valueOrString(registry.PublicNetworkAccess), "enabled")
	adminEnabled := registry.AdminUserEnabled != nil && *registry.AdminUserEnabled
	anonymousEnabled := registry.AnonymousPullEnabled != nil && *registry.AnonymousPullEnabled

	switch {
	case publicEnabled && (adminEnabled || anonymousEnabled):
		return 0
	case publicEnabled:
		return 1
	case adminEnabled || anonymousEnabled:
		return 2
	default:
		return 3
	}
}

func acrGovernanceWeaknessScore(registry models.AcrRegistryAsset) int {
	score := 0
	for _, value := range []*string{
		registry.QuarantinePolicyStatus,
		registry.RetentionPolicyStatus,
		registry.TrustPolicyStatus,
	} {
		if strings.EqualFold(valueOrString(value), "disabled") {
			score++
		}
	}
	return score
}

func acrIntValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
