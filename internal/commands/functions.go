package commands

import (
	"context"
	"strconv"
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
		for idx := range functionApps {
			functionApps[idx] = decorateFunctionAppArtifact(functionApps[idx])
		}

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

func decorateFunctionAppArtifact(app models.FunctionAppAsset) models.FunctionAppAsset {
	app.Runtime = compactArtifactValue(functionArtifactRuntime(app))
	app.Identity = compactArtifactValue(functionArtifactIdentity(app))
	app.Deployment = compactArtifactValue(functionArtifactDeployment(app))
	return app
}

func compactArtifactValue(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "-" {
		return nil
	}
	return models.StringPtr(trimmed)
}

func functionArtifactRuntime(app models.FunctionAppAsset) string {
	parts := make([]string, 0, 2)
	if app.RuntimeStack != nil && *app.RuntimeStack != "" {
		parts = append(parts, *app.RuntimeStack)
	}
	if app.FunctionsExtensionVersion != nil && *app.FunctionsExtensionVersion != "" {
		parts = append(parts, "functions="+*app.FunctionsExtensionVersion)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func functionArtifactIdentity(app models.FunctionAppAsset) string {
	parts := make([]string, 0, 2)
	if app.WorkloadIdentityType != nil && *app.WorkloadIdentityType != "" {
		parts = append(parts, *app.WorkloadIdentityType)
	}
	if len(app.WorkloadIdentityIDs) > 0 {
		parts = append(parts, "user-assigned="+strconv.Itoa(len(app.WorkloadIdentityIDs)))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func functionArtifactDeployment(app models.FunctionAppAsset) string {
	parts := make([]string, 0, 3)
	if app.AzureWebJobsStorageValueType != nil && *app.AzureWebJobsStorageValueType != "" {
		parts = append(parts, "storage="+*app.AzureWebJobsStorageValueType)
	}
	if app.RunFromPackage != nil {
		if *app.RunFromPackage {
			parts = append(parts, "package=yes")
		} else {
			parts = append(parts, "package=disabled")
		}
	}
	if app.KeyVaultReferenceCount != nil {
		parts = append(parts, "kv-refs="+strconv.Itoa(*app.KeyVaultReferenceCount))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}
