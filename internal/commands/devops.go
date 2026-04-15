package commands

import (
	"context"
	"strings"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func devopsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Devops(ctx, request.Tenant, request.Subscription, request.DevOpsOrganization)
		if err != nil {
			return nil, err
		}

		return models.DevopsOutput{
			Metadata: models.DevopsMetadata{
				SchemaVersion:      contracts.AzureFoxSchemaVersion,
				Command:            "devops",
				GeneratedAt:        now().UTC().Format(time.RFC3339),
				TenantID:           models.StringPtr(facts.TenantID),
				SubscriptionID:     models.StringPtr(facts.SubscriptionID),
				DevOpsOrganization: models.StringPtr(facts.DevOpsOrganization),
				TokenSource:        models.StringPtr(facts.TokenSource),
				AuthMode:           models.StringPtr(facts.AuthMode),
			},
			Pipelines: sortedByLess(facts.Pipelines, devopsLess),
			Findings:  []models.Finding{},
			Issues:    facts.Issues,
		}, nil
	}
}

func devopsLess(left models.DevopsPipelineAsset, right models.DevopsPipelineAsset) bool {
	leftHasCIOrSchedule := hasAnyFold(left.TriggerTypes, "continuousintegration", "schedule")
	rightHasCIOrSchedule := hasAnyFold(right.TriggerTypes, "continuousintegration", "schedule")

	switch {
	case len(left.AzureServiceConnectionNames) == 0 && len(right.AzureServiceConnectionNames) > 0:
		return false
	case len(left.AzureServiceConnectionNames) > 0 && len(right.AzureServiceConnectionNames) == 0:
		return true
	case left.SecretVariableCount == 0 && right.SecretVariableCount > 0:
		return false
	case left.SecretVariableCount > 0 && right.SecretVariableCount == 0:
		return true
	case len(left.KeyVaultGroupNames) == 0 && len(right.KeyVaultGroupNames) > 0:
		return false
	case len(left.KeyVaultGroupNames) > 0 && len(right.KeyVaultGroupNames) == 0:
		return true
	case !leftHasCIOrSchedule && rightHasCIOrSchedule:
		return false
	case leftHasCIOrSchedule && !rightHasCIOrSchedule:
		return true
	case len(left.TargetClues) == 0 && len(right.TargetClues) > 0:
		return false
	case len(left.TargetClues) > 0 && len(right.TargetClues) == 0:
		return true
	case left.ProjectName != right.ProjectName:
		return left.ProjectName < right.ProjectName
	case left.Name != right.Name:
		return left.Name < right.Name
	default:
		return left.ID < right.ID
	}
}

func hasAnyFold(values []string, wanted ...string) bool {
	for _, value := range values {
		for _, candidate := range wanted {
			if equalFold(value, candidate) {
				return true
			}
		}
	}
	return false
}

func equalFold(left string, right string) bool {
	return strings.EqualFold(left, right)
}
