package commands

import (
	"context"
	"sort"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

var roleTrustTypeRanks = map[string]int{
	"federated-credential":     0,
	"service-principal-owner":  1,
	"app-owner":                2,
	"app-to-service-principal": 3,
}

var roleTrustEvidenceRanks = map[string]int{
	"graph-federated-credential": 0,
	"graph-owner":                1,
	"graph-app-role-assignment":  2,
}

func roleTrustsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.RoleTrusts(ctx, request.Tenant, request.Subscription, request.RoleTrustsMode)
		if err != nil {
			return nil, err
		}

		trusts := append([]models.RoleTrustSummary{}, facts.Trusts...)
		sort.SliceStable(trusts, func(i int, j int) bool {
			left := trusts[i]
			right := trusts[j]

			leftConfirmed := strings.ToLower(left.Confidence) == "confirmed"
			rightConfirmed := strings.ToLower(right.Confidence) == "confirmed"
			if leftConfirmed != rightConfirmed {
				return leftConfirmed
			}

			leftTrustRank, leftEvidenceRank := roleTrustPriority(left)
			rightTrustRank, rightEvidenceRank := roleTrustPriority(right)
			if leftTrustRank != rightTrustRank {
				return leftTrustRank < rightTrustRank
			}
			if leftEvidenceRank != rightEvidenceRank {
				return leftEvidenceRank < rightEvidenceRank
			}

			leftFollowOn := roleTrustFollowOnRank(left.FollowOnKind)
			rightFollowOn := roleTrustFollowOnRank(right.FollowOnKind)
			if leftFollowOn != rightFollowOn {
				return leftFollowOn < rightFollowOn
			}

			leftSource := firstNonEmpty(stringPtrValue(left.SourceName), left.SourceObjectID)
			rightSource := firstNonEmpty(stringPtrValue(right.SourceName), right.SourceObjectID)
			if leftSource != rightSource {
				return leftSource < rightSource
			}

			leftTarget := firstNonEmpty(stringPtrValue(left.TargetName), left.TargetObjectID)
			rightTarget := firstNonEmpty(stringPtrValue(right.TargetName), right.TargetObjectID)
			if leftTarget != rightTarget {
				return leftTarget < rightTarget
			}

			return left.SourceObjectID < right.SourceObjectID
		})

		return models.RoleTrustsOutput{
			Metadata: scopedMetadata(now, request, facts.TenantID, facts.SubscriptionID, "role-trusts"),
			Mode:     facts.Mode,
			Trusts:   trusts,
			Issues:   facts.Issues,
		}, nil
	}
}

func roleTrustPriority(trust models.RoleTrustSummary) (int, int) {
	trustRank, ok := roleTrustTypeRanks[strings.ToLower(strings.TrimSpace(trust.TrustType))]
	if !ok {
		trustRank = 9
	}

	evidenceRank, ok := roleTrustEvidenceRanks[strings.ToLower(strings.TrimSpace(trust.EvidenceType))]
	if !ok {
		evidenceRank = 9
	}

	return trustRank, evidenceRank
}

func roleTrustFollowOnRank(kind models.RoleTrustFollowOnKind) int {
	switch kind {
	case models.RoleTrustFollowOnPrivilegeConfirmation:
		return 0
	case models.RoleTrustFollowOnOwnershipReview:
		return 1
	case models.RoleTrustFollowOnOutsideTenant:
		return 2
	default:
		return 9
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
