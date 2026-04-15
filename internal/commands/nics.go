package commands

import (
	"context"
	"sort"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func nicsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.NICs(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		nicAssets := append([]models.NicAsset{}, facts.NICAssets...)
		sort.SliceStable(nicAssets, func(i int, j int) bool {
			left := nicAssets[i]
			right := nicAssets[j]

			if len(left.PublicIPIDs) != len(right.PublicIPIDs) {
				return len(left.PublicIPIDs) > len(right.PublicIPIDs)
			}

			leftUnusualAttachment := left.AttachedAssetID == nil || left.AttachedAssetName == nil
			rightUnusualAttachment := right.AttachedAssetID == nil || right.AttachedAssetName == nil
			if leftUnusualAttachment != rightUnusualAttachment {
				return !leftUnusualAttachment
			}

			leftBoundarySignalCount := len(left.SubnetIDs) + len(left.VnetIDs)
			if left.NetworkSecurityGroupID != nil {
				leftBoundarySignalCount++
			}
			rightBoundarySignalCount := len(right.SubnetIDs) + len(right.VnetIDs)
			if right.NetworkSecurityGroupID != nil {
				rightBoundarySignalCount++
			}
			if leftBoundarySignalCount != rightBoundarySignalCount {
				return leftBoundarySignalCount > rightBoundarySignalCount
			}

			if len(left.PrivateIPs) != len(right.PrivateIPs) {
				return len(left.PrivateIPs) > len(right.PrivateIPs)
			}

			leftAttachedName := ""
			if left.AttachedAssetName != nil {
				leftAttachedName = *left.AttachedAssetName
			}
			rightAttachedName := ""
			if right.AttachedAssetName != nil {
				rightAttachedName = *right.AttachedAssetName
			}
			if leftAttachedName != rightAttachedName {
				return leftAttachedName < rightAttachedName
			}

			return left.Name < right.Name
		})

		return models.NicsOutput{
			Findings:  []models.Finding{},
			Issues:    facts.Issues,
			Metadata:  networkMetadata(now, facts.TenantID, facts.SubscriptionID, "nics"),
			NicAssets: nicAssets,
		}, nil
	}
}
