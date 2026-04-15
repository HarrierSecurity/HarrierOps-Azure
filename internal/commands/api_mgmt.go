package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func apiMgmtHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.ApiMgmt(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		services := sortedByLess(facts.ApiManagementServices, apiMgmtLess)

		return models.ApiMgmtOutput{
			ApiManagementServices: services,
			Findings:              []models.Finding{},
			Issues:                facts.Issues,
			Metadata:              runtimeCommandMetadata("api-mgmt", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func apiMgmtLess(left models.ApiMgmtServiceAsset, right models.ApiMgmtServiceAsset) bool {
	if apiMgmtPriority(left) != apiMgmtPriority(right) {
		return apiMgmtPriority(left)
	}
	if apiMgmtIntValue(left.NamedValueSecretCount) != apiMgmtIntValue(right.NamedValueSecretCount) {
		return apiMgmtIntValue(left.NamedValueSecretCount) > apiMgmtIntValue(right.NamedValueSecretCount)
	}
	if apiMgmtIntValue(left.NamedValueKeyVaultCount) != apiMgmtIntValue(right.NamedValueKeyVaultCount) {
		return apiMgmtIntValue(left.NamedValueKeyVaultCount) > apiMgmtIntValue(right.NamedValueKeyVaultCount)
	}
	if apiMgmtIntValue(left.SubscriptionCount) != apiMgmtIntValue(right.SubscriptionCount) {
		return apiMgmtIntValue(left.SubscriptionCount) > apiMgmtIntValue(right.SubscriptionCount)
	}
	if len(left.BackendHostnames) != len(right.BackendHostnames) {
		return len(left.BackendHostnames) > len(right.BackendHostnames)
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func apiMgmtPriority(service models.ApiMgmtServiceAsset) bool {
	return len(service.GatewayHostnames) > 0 || normalizedLower(valueOrString(service.PublicNetworkAccess)) == "enabled"
}

func apiMgmtIntValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func valueOrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
