package commands

import (
	"context"
	"strconv"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func azureMLHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.AzureML(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		workspaces := sortedByLess(facts.Workspaces, azureMLLess)
		for idx := range workspaces {
			workspaces[idx] = decorateAzureMLArtifact(workspaces[idx])
		}

		return models.AzureMLOutput{
			Findings:   []models.Finding{},
			Issues:     facts.Issues,
			Metadata:   runtimeCommandMetadata("azure-ml", now, facts.TenantID, facts.SubscriptionID),
			Workspaces: workspaces,
		}, nil
	}
}

func azureMLLess(left models.AzureMLWorkspaceAsset, right models.AzureMLWorkspaceAsset) bool {
	leftRank := azureMLClassificationRank(left.Classification)
	rightRank := azureMLClassificationRank(right.Classification)
	if leftRank != rightRank {
		return leftRank < rightRank
	}

	if left.EndpointCount != right.EndpointCount {
		return left.EndpointCount > right.EndpointCount
	}
	if left.ScheduleCount != right.ScheduleCount {
		return left.ScheduleCount > right.ScheduleCount
	}
	if left.ComputeCount != right.ComputeCount {
		return left.ComputeCount > right.ComputeCount
	}
	if left.JobCount != right.JobCount {
		return left.JobCount > right.JobCount
	}

	leftIdentity := left.IdentityType != nil && *left.IdentityType != ""
	rightIdentity := right.IdentityType != nil && *right.IdentityType != ""
	if leftIdentity != rightIdentity {
		return leftIdentity
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func azureMLClassificationRank(classification string) int {
	switch classification {
	case "execution-capable":
		return 0
	case "supporting-persistence-context":
		return 1
	default:
		return 2
	}
}

func decorateAzureMLArtifact(workspace models.AzureMLWorkspaceAsset) models.AzureMLWorkspaceAsset {
	workspace.Runtime = compactArtifactValue(azureMLArtifactRuntime(workspace))
	workspace.Serving = compactArtifactValue(azureMLArtifactServing(workspace))
	workspace.Identity = compactArtifactValue(azureMLArtifactIdentity(workspace))
	workspace.Storage = compactArtifactValue(azureMLArtifactStorage(workspace))
	return workspace
}

func azureMLArtifactRuntime(workspace models.AzureMLWorkspaceAsset) string {
	parts := []string{}
	if workspace.ComputeCount > 0 {
		parts = append(parts, "compute="+strconv.Itoa(workspace.ComputeCount))
	}
	if workspace.JobCount > 0 {
		parts = append(parts, "jobs="+strconv.Itoa(workspace.JobCount))
	}
	if workspace.ScheduleCount > 0 {
		parts = append(parts, "schedules="+strconv.Itoa(workspace.ScheduleCount))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func azureMLArtifactServing(workspace models.AzureMLWorkspaceAsset) string {
	if workspace.EndpointCount == 0 {
		return "-"
	}
	parts := []string{"endpoints=" + strconv.Itoa(workspace.EndpointCount)}
	if len(workspace.EndpointAuthModes) > 0 {
		parts = append(parts, "auth="+strings.Join(workspace.EndpointAuthModes, ","))
	}
	if len(workspace.EndpointPublicAccess) > 0 {
		parts = append(parts, "public="+strings.Join(workspace.EndpointPublicAccess, ","))
	}
	return strings.Join(parts, "; ")
}

func azureMLArtifactIdentity(workspace models.AzureMLWorkspaceAsset) string {
	parts := []string{}
	if workspace.IdentityType != nil && *workspace.IdentityType != "" {
		parts = append(parts, *workspace.IdentityType)
	}
	userAssigned := len(workspace.IdentityIDs)
	if strings.Contains(strings.ToLower(stringPtrValue(workspace.IdentityType)), "systemassigned") && userAssigned > 0 {
		userAssigned--
	}
	if userAssigned > 0 {
		parts = append(parts, "user-assigned="+strconv.Itoa(userAssigned))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func azureMLArtifactStorage(workspace models.AzureMLWorkspaceAsset) string {
	parts := []string{}
	if workspace.DatastoreCount > 0 {
		parts = append(parts, "datastores="+strconv.Itoa(workspace.DatastoreCount))
	}
	if workspace.StorageAccountID != nil && *workspace.StorageAccountID != "" {
		parts = append(parts, "storage-account")
	}
	if workspace.KeyVaultID != nil && *workspace.KeyVaultID != "" {
		parts = append(parts, "key-vault")
	}
	if workspace.ContainerRegistryID != nil && *workspace.ContainerRegistryID != "" {
		parts = append(parts, "container-registry")
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}
