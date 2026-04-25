package commands

import (
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

const toolVersion = "dev"

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

func withArtifactContext(metadata models.Metadata, request Request, principal models.Principal, authMode string, tokenSource string) models.Metadata {
	metadata.ArtifactContext = artifactContext(metadata.Command, request, principal)
	metadata.AuthMode = models.StringPtr(authMode)
	metadata.TokenSource = models.StringPtr(tokenSource)
	return metadata
}

func withScopedArtifactContext(metadata models.ScopedCommandMetadata, request Request, principal models.Principal, authMode string, tokenSource string) models.ScopedCommandMetadata {
	metadata.ArtifactContext = artifactContext(metadata.Command, request, principal)
	metadata.AuthMode = models.StringPtr(authMode)
	metadata.TokenSource = models.StringPtr(tokenSource)
	return metadata
}

func withPrincipalsArtifactContext(metadata models.PrincipalsMetadata, request Request, principal models.Principal, authMode string, tokenSource string) models.PrincipalsMetadata {
	metadata.ArtifactContext = artifactContext(metadata.Command, request, principal)
	metadata.AuthMode = models.StringPtr(authMode)
	metadata.TokenSource = models.StringPtr(tokenSource)
	return metadata
}

func withRuntimeArtifactContext(metadata models.RuntimeCommandMetadata, request Request, principal models.Principal, authMode string, tokenSource string) models.RuntimeCommandMetadata {
	metadata.ArtifactContext = artifactContext(metadata.Command, request, principal)
	metadata.AuthMode = models.StringPtr(authMode)
	metadata.TokenSource = models.StringPtr(tokenSource)
	return metadata
}

func withAutomationArtifactContext(metadata models.AutomationMetadata, request Request, principal models.Principal, authMode string, tokenSource string) models.AutomationMetadata {
	metadata.ArtifactContext = artifactContext(metadata.Command, request, principal)
	metadata.AuthMode = models.StringPtr(authMode)
	metadata.TokenSource = models.StringPtr(tokenSource)
	return metadata
}

func withSessionArtifacts(metadata models.ScopedCommandMetadata, artifacts []models.SessionArtifact) models.ScopedCommandMetadata {
	if len(artifacts) == 0 {
		return metadata
	}
	metadata.SessionArtifacts = append([]models.SessionArtifact{}, artifacts...)
	return metadata
}

func withRuntimeSessionArtifacts(metadata models.RuntimeCommandMetadata, artifacts []models.SessionArtifact) models.RuntimeCommandMetadata {
	if len(artifacts) == 0 {
		return metadata
	}
	metadata.SessionArtifacts = append([]models.SessionArtifact{}, artifacts...)
	return metadata
}

func withMetadataSessionArtifacts(metadata models.Metadata, artifacts []models.SessionArtifact) models.Metadata {
	if len(artifacts) == 0 {
		return metadata
	}
	metadata.SessionArtifacts = append([]models.SessionArtifact{}, artifacts...)
	return metadata
}

func artifactContext(command string, request Request, principal models.Principal) *models.ArtifactContext {
	return &models.ArtifactContext{
		ToolVersion: toolVersion,
		CurrentPrincipal: models.ArtifactPrincipal{
			ID:            principal.ID,
			PrincipalType: principal.PrincipalType,
			TenantID:      principal.TenantID,
		},
		CommandOptions: artifactCommandOptions(command, request),
	}
}

func artifactIdentityFactsFromContext(context *models.ArtifactContext, authMode *string, tokenSource *string) providers.ArtifactIdentityFacts {
	principal := models.Principal{}
	if context != nil {
		principal = models.Principal{
			ID:            context.CurrentPrincipal.ID,
			PrincipalType: context.CurrentPrincipal.PrincipalType,
			TenantID:      context.CurrentPrincipal.TenantID,
		}
	}
	return providers.ArtifactIdentityFacts{
		CurrentPrincipal: principal,
		AuthMode:         stringPtrValue(authMode),
		TokenSource:      stringPtrValue(tokenSource),
	}
}

func artifactCommandOptions(command string, request Request) map[string]string {
	options := map[string]string{}
	if command == "devops" && request.DevOpsOrganization != "" {
		options["devops_organization"] = request.DevOpsOrganization
	}
	if command == "role-trusts" && request.RoleTrustsMode != "" {
		options["role_trusts_mode"] = string(request.RoleTrustsMode.Semantic())
	}
	if command == "chains" && request.ChainFamily != "" {
		options["chain_family"] = request.ChainFamily
	}
	if command == "persistence" && request.PersistenceSurface != "" {
		options["persistence_surface"] = request.PersistenceSurface
	}
	if command == "evasion" && request.EvasionSurface != "" {
		options["evasion_surface"] = request.EvasionSurface
	}
	if command == "resourcehijacking" && request.ResourceHijackingSurface != "" {
		options["resourcehijacking_surface"] = request.ResourceHijackingSurface
	}
	if command == "pathmasking" && request.PathMaskingSurface != "" {
		options["pathmasking_surface"] = request.PathMaskingSurface
	}
	return options
}

func artifactWorkspace(outDir string) string {
	if outDir == "" {
		return "."
	}
	return outDir
}
