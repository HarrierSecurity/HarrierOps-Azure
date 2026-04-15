package commands

import (
	"context"
	"sort"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func vmssHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.VMSS(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		vmssAssets := append([]models.VmssAsset{}, facts.VMSSAssets...)
		sort.SliceStable(vmssAssets, func(i int, j int) bool {
			left := vmssAssets[i]
			right := vmssAssets[j]

			if vmssHasFrontendPriority(left) != vmssHasFrontendPriority(right) {
				return vmssHasFrontendPriority(left)
			}
			if (left.IdentityType != nil) != (right.IdentityType != nil) {
				return left.IdentityType != nil
			}
			if left.PublicIPConfigurationCount != right.PublicIPConfigurationCount {
				return left.PublicIPConfigurationCount > right.PublicIPConfigurationCount
			}

			leftInstances := 0
			if left.InstanceCount != nil {
				leftInstances = *left.InstanceCount
			}
			rightInstances := 0
			if right.InstanceCount != nil {
				rightInstances = *right.InstanceCount
			}
			if leftInstances != rightInstances {
				return leftInstances > rightInstances
			}

			if vmssOrchestrationRank(left.OrchestrationMode) != vmssOrchestrationRank(right.OrchestrationMode) {
				return vmssOrchestrationRank(left.OrchestrationMode) < vmssOrchestrationRank(right.OrchestrationMode)
			}
			if vmssUpgradeRank(left.UpgradeMode) != vmssUpgradeRank(right.UpgradeMode) {
				return vmssUpgradeRank(left.UpgradeMode) < vmssUpgradeRank(right.UpgradeMode)
			}
			return left.Name < right.Name
		})

		return models.VmssOutput{
			Findings:   []models.Finding{},
			Issues:     facts.Issues,
			Metadata:   commandMetadata("vmss", now, request, facts.TenantID, facts.SubscriptionID, ""),
			VmssAssets: vmssAssets,
		}, nil
	}
}

func vmssHasFrontendPriority(asset models.VmssAsset) bool {
	return asset.PublicIPConfigurationCount > 0 ||
		asset.InboundNATPoolCount > 0 ||
		asset.LoadBalancerBackendPoolCount > 0 ||
		asset.ApplicationGatewayBackendPoolCount > 0
}

func vmssOrchestrationRank(value *string) int {
	if value == nil {
		return 9
	}
	switch normalizedLower(*value) {
	case "uniform":
		return 0
	case "flexible":
		return 1
	default:
		return 9
	}
}

func vmssUpgradeRank(value *string) int {
	if value == nil {
		return 9
	}
	switch normalizedLower(*value) {
	case "rolling":
		return 0
	case "automatic":
		return 1
	case "manual":
		return 2
	default:
		return 9
	}
}
