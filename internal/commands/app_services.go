package commands

import (
	"context"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func appServicesHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.AppServices(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		appServices := sortedByLess(facts.AppServices, appServiceLess)

		return models.AppServicesOutput{
			AppServices: appServices,
			Findings:    []models.Finding{},
			Issues:      facts.Issues,
			Metadata:    withRuntimeArtifactContext(runtimeCommandMetadata("app-services", now, facts.TenantID, facts.SubscriptionID), request, facts.CurrentPrincipal, facts.AuthMode, facts.TokenSource),
		}, nil
	}
}

func appServiceLess(left models.AppServiceAsset, right models.AppServiceAsset) bool {
	leftExposed := appServiceExposurePriority(left)
	rightExposed := appServiceExposurePriority(right)
	if leftExposed != rightExposed {
		return leftExposed
	}

	leftIdentity := left.WorkloadIdentityType != nil && *left.WorkloadIdentityType != ""
	rightIdentity := right.WorkloadIdentityType != nil && *right.WorkloadIdentityType != ""
	if leftIdentity != rightIdentity {
		return leftIdentity
	}

	leftRank := appServiceHardeningRank(left)
	rightRank := appServiceHardeningRank(right)
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

func appServiceExposurePriority(item models.AppServiceAsset) bool {
	return (item.DefaultHostname != nil && *item.DefaultHostname != "") ||
		strings.EqualFold(optionalString(item.PublicNetworkAccess), "enabled")
}

func appServiceHardeningRank(item models.AppServiceAsset) [4]int {
	httpsRank := 0
	if item.HTTPSOnly {
		httpsRank = 1
	}

	tlsRank := map[string]int{
		"1.0": 0,
		"1.1": 1,
		"1.2": 2,
		"1.3": 3,
	}[strings.TrimSpace(optionalString(item.MinTLSVersion))]
	switch strings.TrimSpace(optionalString(item.MinTLSVersion)) {
	case "1.0", "1.1", "1.2", "1.3":
	case "":
		tlsRank = 5
	default:
		tlsRank = 4
	}

	ftpsRank := map[string]int{
		"allallowed": 0,
		"ftpsonly":   1,
		"disabled":   2,
	}[strings.ToLower(optionalString(item.FTPSState))]
	switch strings.ToLower(optionalString(item.FTPSState)) {
	case "allallowed", "ftpsonly", "disabled":
	case "":
		ftpsRank = 4
	default:
		ftpsRank = 3
	}

	clientCertRank := 0
	if item.ClientCertEnabled {
		clientCertRank = 1
	}

	return [4]int{httpsRank, tlsRank, ftpsRank, clientCertRank}
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
