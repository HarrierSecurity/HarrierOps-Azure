package commands

import (
	"context"
	"strings"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func snapshotsDisksHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.SnapshotsDisks(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		return models.SnapshotsDisksOutput{
			Metadata: models.SnapshotsDisksMetadata{
				SchemaVersion:  contracts.AzureFoxSchemaVersion,
				Command:        "snapshots-disks",
				GeneratedAt:    now().UTC().Format(time.RFC3339),
				TenantID:       models.StringPtr(facts.TenantID),
				SubscriptionID: models.StringPtr(facts.SubscriptionID),
				TokenSource:    nil,
			},
			SnapshotDiskAssets: sortedByLess(facts.SnapshotDiskAssets, snapshotDiskLess),
			Findings:           []models.Finding{},
			Issues:             facts.Issues,
		}, nil
	}
}

func snapshotDiskLess(left models.SnapshotDiskAsset, right models.SnapshotDiskAsset) bool {
	leftDetachedRank := left.AttachmentState != "detached"
	rightDetachedRank := right.AttachmentState != "detached"
	if leftDetachedRank != rightDetachedRank {
		return !leftDetachedRank
	}

	leftSnapshotRank := left.AssetKind != "snapshot"
	rightSnapshotRank := right.AssetKind != "snapshot"
	if leftSnapshotRank != rightSnapshotRank {
		return !leftSnapshotRank
	}

	leftPublicRank := strings.ToLower(stringPtrValue(left.PublicNetworkAccess)) != "enabled"
	rightPublicRank := strings.ToLower(stringPtrValue(right.PublicNetworkAccess)) != "enabled"
	if leftPublicRank != rightPublicRank {
		return !leftPublicRank
	}

	leftPriority := snapshotDiskPrioritySortValue(left)
	rightPriority := snapshotDiskPrioritySortValue(right)
	if leftPriority != rightPriority {
		return leftPriority < rightPriority
	}

	leftAttachedName := stringPtrValue(left.AttachedToName)
	rightAttachedName := stringPtrValue(right.AttachedToName)
	if leftAttachedName != rightAttachedName {
		return leftAttachedName < rightAttachedName
	}

	leftSourceName := stringPtrValue(left.SourceResourceName)
	rightSourceName := stringPtrValue(right.SourceResourceName)
	if leftSourceName != rightSourceName {
		return leftSourceName < rightSourceName
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func snapshotDiskPrioritySortValue(item models.SnapshotDiskAsset) int {
	score := 0
	if item.DiskAccessID != nil {
		score -= 2
	}
	if item.MaxShares != nil && *item.MaxShares != 1 {
		score -= 2
	}
	if strings.EqualFold(stringPtrValue(item.NetworkAccessPolicy), "allowall") {
		score -= 2
	}
	if strings.EqualFold(stringPtrValue(item.PublicNetworkAccess), "enabled") {
		score -= 1
	}
	if item.DiskEncryptionSetID == nil {
		score -= 1
	}
	if item.AttachmentState == "detached" {
		score -= 2
	}
	if item.AssetKind == "snapshot" {
		score -= 1
	}
	return score
}
