package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func automationHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Automation(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		return models.AutomationOutput{
			Metadata: models.AutomationMetadata{
				SchemaVersion:  contracts.AzureFoxSchemaVersion,
				Command:        "automation",
				GeneratedAt:    now().UTC().Format(time.RFC3339),
				TenantID:       models.StringPtr(facts.TenantID),
				SubscriptionID: models.StringPtr(facts.SubscriptionID),
				TokenSource:    nil,
			},
			AutomationAccounts: sortedByLess(facts.AutomationAccounts, automationLess),
			Findings:           []models.Finding{},
			Issues:             facts.Issues,
		}, nil
	}
}

func automationLess(left models.AutomationAccountAsset, right models.AutomationAccountAsset) bool {
	leftSecureAssets := automationSecureAssetTotal(left)
	rightSecureAssets := automationSecureAssetTotal(right)

	switch {
	case intPtrValue(left.HybridWorkerGroupCount) == 0 && intPtrValue(right.HybridWorkerGroupCount) > 0:
		return false
	case intPtrValue(left.HybridWorkerGroupCount) > 0 && intPtrValue(right.HybridWorkerGroupCount) == 0:
		return true
	case left.IdentityType == nil && right.IdentityType != nil:
		return false
	case left.IdentityType != nil && right.IdentityType == nil:
		return true
	case intPtrValue(left.WebhookCount) == 0 && intPtrValue(right.WebhookCount) > 0:
		return false
	case intPtrValue(left.WebhookCount) > 0 && intPtrValue(right.WebhookCount) == 0:
		return true
	case intPtrValue(left.PublishedRunbookCount) != intPtrValue(right.PublishedRunbookCount):
		return intPtrValue(left.PublishedRunbookCount) > intPtrValue(right.PublishedRunbookCount)
	case intPtrValue(left.JobScheduleCount) != intPtrValue(right.JobScheduleCount):
		return intPtrValue(left.JobScheduleCount) > intPtrValue(right.JobScheduleCount)
	case leftSecureAssets != rightSecureAssets:
		return leftSecureAssets > rightSecureAssets
	case intPtrValue(left.RunbookCount) != intPtrValue(right.RunbookCount):
		return intPtrValue(left.RunbookCount) > intPtrValue(right.RunbookCount)
	case left.Name != right.Name:
		return left.Name < right.Name
	default:
		return left.ID < right.ID
	}
}

func automationSecureAssetTotal(item models.AutomationAccountAsset) int {
	return intPtrValue(item.CredentialCount) +
		intPtrValue(item.CertificateCount) +
		intPtrValue(item.ConnectionCount) +
		intPtrValue(item.EncryptedVariableCount)
}
