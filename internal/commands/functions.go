package commands

import (
	"context"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func functionsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Functions(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		functionApps := sortedByLess(facts.FunctionApps, functionAppLess)

		return models.FunctionsOutput{
			Findings:     []models.Finding{},
			FunctionApps: functionApps,
			Issues:       facts.Issues,
			Metadata:     runtimeCommandMetadata("functions", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func functionAppLess(left models.FunctionAppAsset, right models.FunctionAppAsset) bool {
	leftExposed := functionAppExposurePriority(left)
	rightExposed := functionAppExposurePriority(right)
	if leftExposed != rightExposed {
		return leftExposed
	}

	leftIdentity := left.WorkloadIdentityType != nil && *left.WorkloadIdentityType != ""
	rightIdentity := right.WorkloadIdentityType != nil && *right.WorkloadIdentityType != ""
	if leftIdentity != rightIdentity {
		return leftIdentity
	}

	leftPlainText := optionalString(left.AzureWebJobsStorageValueType) == "plain-text"
	rightPlainText := optionalString(right.AzureWebJobsStorageValueType) == "plain-text"
	if leftPlainText != rightPlainText {
		return leftPlainText
	}

	leftRank := functionDeploymentSignalRank(left)
	rightRank := functionDeploymentSignalRank(right)
	if leftRank != rightRank {
		for idx := range leftRank {
			if leftRank[idx] != rightRank[idx] {
				return leftRank[idx] < rightRank[idx]
			}
		}
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func functionAppExposurePriority(item models.FunctionAppAsset) bool {
	return (item.DefaultHostname != nil && *item.DefaultHostname != "") ||
		strings.EqualFold(optionalString(item.PublicNetworkAccess), "enabled")
}

func functionDeploymentSignalRank(item models.FunctionAppAsset) [3]int {
	runFromPackageRank := 1
	if item.RunFromPackage != nil && *item.RunFromPackage {
		runFromPackageRank = 0
	}

	keyVaultRefs := intPtrValue(item.KeyVaultReferenceCount)
	keyVaultRank := -keyVaultRefs

	signalCount := 0
	if item.RunFromPackage != nil && *item.RunFromPackage {
		signalCount++
	}
	if keyVaultRefs > 0 {
		signalCount++
	}

	return [3]int{-signalCount, runFromPackageRank, keyVaultRank}
}

func intPtrValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
