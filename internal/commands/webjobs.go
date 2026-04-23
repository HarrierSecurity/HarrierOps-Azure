package commands

import (
	"context"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func webJobsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.WebJobs(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		webJobs := sortedByLess(facts.WebJobs, webJobLess)

		return models.WebJobsOutput{
			Findings: []models.Finding{},
			Issues:   facts.Issues,
			Metadata: runtimeCommandMetadata("webjobs", now, facts.TenantID, facts.SubscriptionID),
			WebJobs:  webJobs,
		}, nil
	}
}

func webJobLess(left models.WebJobAsset, right models.WebJobAsset) bool {
	leftModeRank := webJobModeRank(left.Mode)
	rightModeRank := webJobModeRank(right.Mode)
	if leftModeRank != rightModeRank {
		return leftModeRank < rightModeRank
	}

	leftIdentity := left.ParentIdentityType != nil && *left.ParentIdentityType != ""
	rightIdentity := right.ParentIdentityType != nil && *right.ParentIdentityType != ""
	if leftIdentity != rightIdentity {
		return leftIdentity
	}

	leftHostname := left.ParentHostname != nil && *left.ParentHostname != ""
	rightHostname := right.ParentHostname != nil && *right.ParentHostname != ""
	if leftHostname != rightHostname {
		return leftHostname
	}

	if left.ParentAppName != right.ParentAppName {
		return left.ParentAppName < right.ParentAppName
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func webJobModeRank(mode string) int {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "continuous":
		return 0
	case "scheduled":
		return 1
	case "triggered/manual":
		return 2
	default:
		return 3
	}
}
