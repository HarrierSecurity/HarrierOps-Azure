package commands

import (
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
)

func commandMetadata(
	command string,
	now func() time.Time,
	request Request,
	tenantID string,
	subscriptionID string,
	tokenSource string,
) models.Metadata {
	return models.Metadata{
		Command:            command,
		DevOpsOrganization: models.StringPtr(request.DevOpsOrganization),
		GeneratedAt:        now().UTC().Format(time.RFC3339),
		SchemaVersion:      contracts.AzureFoxSchemaVersion,
		SubscriptionID:     models.StringPtr(subscriptionID),
		TenantID:           models.StringPtr(tenantID),
		TokenSource:        models.StringPtr(tokenSource),
	}
}

func runtimeCommandMetadata(
	command string,
	now func() time.Time,
	tenantID string,
	subscriptionID string,
) models.RuntimeCommandMetadata {
	return models.RuntimeCommandMetadata{
		Command:        command,
		GeneratedAt:    now().UTC().Format(time.RFC3339),
		SchemaVersion:  contracts.AzureFoxSchemaVersion,
		SubscriptionID: models.StringPtr(subscriptionID),
		TenantID:       models.StringPtr(tenantID),
		TokenSource:    nil,
	}
}

func whoAmIMetadata(
	now func() time.Time,
	request Request,
	tenantID string,
	subscriptionID string,
	tokenSource string,
	authMode string,
) models.WhoAmIMetadata {
	return models.WhoAmIMetadata{
		AuthMode: models.StringPtr(authMode),
		Metadata: commandMetadata("whoami", now, request, tenantID, subscriptionID, tokenSource),
	}
}

func scopedMetadata(
	now func() time.Time,
	request Request,
	tenantID string,
	subscriptionID string,
	command string,
) models.PermissionsMetadata {
	return models.ScopedCommandMetadata{
		SchemaVersion:      contracts.AzureFoxSchemaVersion,
		Command:            command,
		GeneratedAt:        now().UTC().Format(time.RFC3339),
		TenantID:           models.StringPtr(tenantID),
		SubscriptionID:     models.StringPtr(subscriptionID),
		DevOpsOrganization: models.StringPtr(request.DevOpsOrganization),
		TokenSource:        nil,
		AuthMode:           nil,
	}
}

func networkMetadata(
	now func() time.Time,
	tenantID string,
	subscriptionID string,
	command string,
) models.NetworkCommandMetadata {
	return models.NetworkCommandMetadata{
		Command:        command,
		GeneratedAt:    now().UTC().Format(time.RFC3339),
		SchemaVersion:  contracts.AzureFoxSchemaVersion,
		SubscriptionID: models.StringPtr(subscriptionID),
		TenantID:       models.StringPtr(tenantID),
		TokenSource:    nil,
	}
}
