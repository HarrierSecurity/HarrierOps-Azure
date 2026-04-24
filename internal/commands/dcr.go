package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func dcrHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.DCR(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		dcrs := sortedByLess(facts.DCRs, dcrLess)

		return models.DCROutput{
			DCRs:     dcrs,
			Findings: []models.Finding{},
			Issues:   facts.Issues,
			Metadata: runtimeCommandMetadata("dcr", now, facts.TenantID, facts.SubscriptionID),
		}, nil
	}
}

func dcrLess(left models.DCRAsset, right models.DCRAsset) bool {
	leftTransform := left.TransformationCount > 0
	rightTransform := right.TransformationCount > 0
	if leftTransform != rightTransform {
		return leftTransform
	}

	leftHighSignal := len(left.HighSignalStreams)
	rightHighSignal := len(right.HighSignalStreams)
	if leftHighSignal != rightHighSignal {
		return leftHighSignal > rightHighSignal
	}

	leftAssociations := left.AssociationCount
	rightAssociations := right.AssociationCount
	if leftAssociations != rightAssociations {
		return leftAssociations > rightAssociations
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}
