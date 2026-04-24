package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func containerAppsJobsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.ContainerAppsJobs(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		jobs := sortedByLess(facts.ContainerAppsJobs, containerAppsJobLess)

		return models.ContainerAppsJobsOutput{
			ContainerAppsJobs: jobs,
			Findings:          []models.Finding{},
			Issues:            facts.Issues,
			Metadata:          runtimeCommandMetadata("container-apps-jobs", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func containerAppsJobLess(left models.ContainerAppsJobAsset, right models.ContainerAppsJobAsset) bool {
	leftScheduled := stringPtrValue(left.ScheduleExpression) != ""
	rightScheduled := stringPtrValue(right.ScheduleExpression) != ""
	if leftScheduled != rightScheduled {
		return leftScheduled
	}

	leftEvent := len(left.EventRules) > 0
	rightEvent := len(right.EventRules) > 0
	if leftEvent != rightEvent {
		return leftEvent
	}

	leftIdentity := stringPtrValue(left.WorkloadIdentityType) != ""
	rightIdentity := stringPtrValue(right.WorkloadIdentityType) != ""
	if leftIdentity != rightIdentity {
		return leftIdentity
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}
