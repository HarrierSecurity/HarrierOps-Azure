package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func containerAppsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.ContainerApps(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		containerApps := sortedByLess(facts.ContainerApps, containerAppLess)

		return models.ContainerAppsOutput{
			ContainerApps: containerApps,
			Findings:      []models.Finding{},
			Issues:        facts.Issues,
			Metadata:      runtimeCommandMetadata("container-apps", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func containerAppLess(left models.ContainerAppAsset, right models.ContainerAppAsset) bool {
	leftExternal := left.ExternalIngressEnabled != nil && *left.ExternalIngressEnabled
	rightExternal := right.ExternalIngressEnabled != nil && *right.ExternalIngressEnabled
	if leftExternal != rightExternal {
		return leftExternal
	}

	leftIdentity := left.WorkloadIdentityType != nil && *left.WorkloadIdentityType != ""
	rightIdentity := right.WorkloadIdentityType != nil && *right.WorkloadIdentityType != ""
	if leftIdentity != rightIdentity {
		return leftIdentity
	}

	leftHostname := left.DefaultHostname != nil && *left.DefaultHostname != ""
	rightHostname := right.DefaultHostname != nil && *right.DefaultHostname != ""
	if leftHostname != rightHostname {
		return leftHostname
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}
