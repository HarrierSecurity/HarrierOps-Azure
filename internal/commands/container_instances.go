package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func containerInstancesHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.ContainerInstances(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		containerInstances := sortedByLess(facts.ContainerInstances, containerInstanceLess)

		return models.ContainerInstancesOutput{
			ContainerInstances: containerInstances,
			Findings:           []models.Finding{},
			Issues:             facts.Issues,
			Metadata:           runtimeCommandMetadata("container-instances", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func containerInstanceLess(left models.ContainerInstanceAsset, right models.ContainerInstanceAsset) bool {
	leftPublic := stringPtrValue(left.PublicIPAddress) != "" || stringPtrValue(left.FQDN) != ""
	rightPublic := stringPtrValue(right.PublicIPAddress) != "" || stringPtrValue(right.FQDN) != ""
	if leftPublic != rightPublic {
		return leftPublic
	}

	leftIdentity := stringPtrValue(left.WorkloadIdentityType) != ""
	rightIdentity := stringPtrValue(right.WorkloadIdentityType) != ""
	if leftIdentity != rightIdentity {
		return leftIdentity
	}

	leftFQDN := stringPtrValue(left.FQDN) != ""
	rightFQDN := stringPtrValue(right.FQDN) != ""
	if leftFQDN != rightFQDN {
		return leftFQDN
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}
