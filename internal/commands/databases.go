package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func databasesHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Databases(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		servers := sortedByLess(facts.DatabaseServers, databasesLess)

		return models.DatabasesOutput{
			DatabaseServers: servers,
			Findings:        []models.Finding{},
			Issues:          facts.Issues,
			Metadata:        runtimeCommandMetadata("databases", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func databasesLess(left models.DatabaseServerAsset, right models.DatabaseServerAsset) bool {
	if databasesExposurePriority(left) != databasesExposurePriority(right) {
		return databasesExposurePriority(left)
	}
	if databasesTLSRank(left) != databasesTLSRank(right) {
		return databasesTLSRank(left) < databasesTLSRank(right)
	}
	if databasesCountValue(left.DatabaseCount) != databasesCountValue(right.DatabaseCount) {
		return databasesCountValue(left.DatabaseCount) > databasesCountValue(right.DatabaseCount)
	}
	if databasesHasIdentity(left) != databasesHasIdentity(right) {
		return databasesHasIdentity(left)
	}
	if left.Engine != right.Engine {
		return left.Engine < right.Engine
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func databasesExposurePriority(item models.DatabaseServerAsset) bool {
	return normalizedLower(databasesStringValue(item.PublicNetworkAccess)) == "enabled"
}

func databasesTLSRank(item models.DatabaseServerAsset) int {
	switch databasesStringValue(item.MinimalTLSVersion) {
	case "1.0":
		return 0
	case "1.1":
		return 1
	case "1.2":
		return 2
	case "1.3":
		return 3
	default:
		return 4
	}
}

func databasesCountValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func databasesHasIdentity(item models.DatabaseServerAsset) bool {
	return databasesStringValue(item.WorkloadIdentityType) != ""
}

func databasesStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
