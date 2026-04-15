package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func aksHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.AKS(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		clusters := sortedByLess(facts.AksClusters, aksClusterLess)

		return models.AksOutput{
			AksClusters: clusters,
			Findings:    []models.Finding{},
			Issues:      facts.Issues,
			Metadata:    runtimeCommandMetadata("aks", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func aksClusterLess(left models.AksClusterAsset, right models.AksClusterAsset) bool {
	leftControlRank := aksControlPlaneRank(left)
	rightControlRank := aksControlPlaneRank(right)
	if leftControlRank != rightControlRank {
		return leftControlRank < rightControlRank
	}

	leftHasIdentity := left.ClusterIdentityType != nil
	rightHasIdentity := right.ClusterIdentityType != nil
	if leftHasIdentity != rightHasIdentity {
		return leftHasIdentity
	}

	leftFederation := aksFederationCueCount(left)
	rightFederation := aksFederationCueCount(right)
	if leftFederation != rightFederation {
		return leftFederation > rightFederation
	}

	if len(left.AddonNames) != len(right.AddonNames) {
		return len(left.AddonNames) > len(right.AddonNames)
	}

	leftAuthRank := aksAuthCueRank(left)
	rightAuthRank := aksAuthCueRank(right)
	if leftAuthRank != rightAuthRank {
		return leftAuthRank < rightAuthRank
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func aksControlPlaneRank(cluster models.AksClusterAsset) int {
	if cluster.FQDN != nil && cluster.PrivateClusterEnabled != nil && !*cluster.PrivateClusterEnabled {
		return 0
	}
	if cluster.FQDN != nil && cluster.PublicFQDNEnabled != nil && *cluster.PublicFQDNEnabled {
		return 1
	}
	return 2
}

func aksFederationCueCount(cluster models.AksClusterAsset) int {
	count := 0
	if cluster.OIDCIssuerEnabled != nil && *cluster.OIDCIssuerEnabled {
		count++
	}
	if cluster.WorkloadIdentityEnabled != nil && *cluster.WorkloadIdentityEnabled {
		count++
	}
	return count
}

func aksAuthCueRank(cluster models.AksClusterAsset) string {
	return aksBoolRank(cluster.LocalAccountsDisabled, false) +
		aksBoolRank(cluster.AADManaged, false) +
		aksBoolRank(cluster.AzureRBACEnabled, false)
}

func aksBoolRank(value *bool, falseWins bool) string {
	if value == nil {
		return "2"
	}
	if *value == falseWins {
		return "0"
	}
	return "1"
}
