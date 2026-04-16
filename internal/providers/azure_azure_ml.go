package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"harrierops-azure/internal/models"
)

const armAzureMLAPIVersion = "2024-04-01"

func (provider AzureProvider) AzureML(ctx context.Context, tenant string, subscription string) (AzureMLFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return AzureMLFacts{}, err
	}

	workspaces, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.MachineLearningServices/workspaces",
		armAzureMLAPIVersion,
	)
	if err != nil {
		return AzureMLFacts{
			TenantID:       session.tenantID,
			SubscriptionID: session.subscription.ID,
			Workspaces:     []models.AzureMLWorkspaceAsset{},
			Issues:         []models.Issue{issueFromError("azure-ml.workspaces", err)},
		}, nil
	}

	rows := make([]models.AzureMLWorkspaceAsset, 0, len(workspaces))
	issues := []models.Issue{}
	for _, workspace := range workspaces {
		workspaceID := mapStringValue(workspace, "id")
		hydrated := workspace
		if workspaceID != "" {
			detailed, getErr := armGetObject(ctx, session.credential, workspaceID, armAzureMLAPIVersion)
			if getErr != nil {
				issues = append(issues, issueFromError("azure-ml.workspace["+workspaceID+"]", getErr))
			} else {
				hydrated = detailed
			}
		}

		computes, computeIssues := azureMLListChildObjects(ctx, session.credential, workspaceID, "computes")
		issues = append(issues, computeIssues...)
		jobs, jobIssues := azureMLListChildObjects(ctx, session.credential, workspaceID, "jobs")
		issues = append(issues, jobIssues...)
		schedules, scheduleIssues := azureMLListChildObjects(ctx, session.credential, workspaceID, "schedules")
		issues = append(issues, scheduleIssues...)
		endpoints, endpointIssues := azureMLListChildObjects(ctx, session.credential, workspaceID, "onlineEndpoints")
		issues = append(issues, endpointIssues...)
		datastores, datastoreIssues := azureMLListChildObjects(ctx, session.credential, workspaceID, "datastores")
		issues = append(issues, datastoreIssues...)

		rows = append(rows, azureMLWorkspaceAsset(hydrated, computes, jobs, schedules, endpoints, datastores))
	}

	return AzureMLFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		Workspaces:     rows,
		Issues:         issues,
	}, nil
}

func azureMLListChildObjects(ctx context.Context, credential azcore.TokenCredential, parentID string, childType string) ([]map[string]any, []models.Issue) {
	if parentID == "" {
		return []map[string]any{}, nil
	}

	items, err := armListObjects(ctx, credential, parentID+"/"+childType, armAzureMLAPIVersion)
	if err != nil {
		return []map[string]any{}, []models.Issue{issueFromError("azure-ml."+childType+"["+parentID+"]", err)}
	}
	return items, nil
}

func azureMLWorkspaceAsset(
	workspace map[string]any,
	computes []map[string]any,
	jobs []map[string]any,
	schedules []map[string]any,
	endpoints []map[string]any,
	datastores []map[string]any,
) models.AzureMLWorkspaceAsset {
	workspaceID := mapStringValue(workspace, "id")
	resourceGroup, workspaceName := resourceGroupAndNameFromID(workspaceID)
	properties := mapValue(workspace, "properties")
	identity := mapValue(workspace, "identity")

	name := firstNonEmpty(mapStringValue(workspace, "name"), workspaceName, "unknown")
	identityType := stringPtr(mapStringValue(identity, "type"))
	identityIDs := azureMLIdentityIDs(workspaceID, identity)
	computeTypes := azureMLComputeTypes(computes)
	jobTypes := azureMLJobTypes(jobs)
	scheduleTriggerTypes := azureMLScheduleTriggerTypes(schedules)
	endpointAuthModes, endpointPublicAccess := azureMLEndpointPosture(endpoints)
	datastoreTypes := azureMLDatastoreTypes(datastores)
	classification := azureMLClassification(len(computes) > 0, len(jobs) > 0, len(endpoints) > 0, len(schedules) > 0)

	return models.AzureMLWorkspaceAsset{
		ID:                    firstNonEmpty(workspaceID, "/unknown/"+name),
		Name:                  name,
		Classification:        classification,
		ResourceGroup:         resourceGroup,
		Location:              stringPtr(mapStringValue(workspace, "location")),
		WorkspaceKind:         stringPtr(mapStringValue(workspace, "kind")),
		State:                 stringPtr(firstNonEmpty(mapStringValue(properties, "provisioningState", "provisioning_state"), mapStringValue(workspace, "state"))),
		PublicNetworkAccess:   stringPtr(mapStringValue(properties, "publicNetworkAccess", "public_network_access")),
		IdentityType:          identityType,
		PrincipalID:           stringPtr(mapStringValue(identity, "principalId", "principal_id")),
		IdentityIDs:           identityIDs,
		ComputeCount:          len(computes),
		ComputeTypes:          computeTypes,
		JobCount:              len(jobs),
		JobTypes:              jobTypes,
		ScheduleCount:         len(schedules),
		ScheduleTriggerTypes:  scheduleTriggerTypes,
		EndpointCount:         len(endpoints),
		EndpointAuthModes:     endpointAuthModes,
		EndpointPublicAccess:  endpointPublicAccess,
		DatastoreCount:        len(datastores),
		DatastoreTypes:        datastoreTypes,
		StorageAccountID:      stringPtr(mapStringValue(properties, "storageAccount", "storage_account")),
		KeyVaultID:            stringPtr(mapStringValue(properties, "keyVault", "key_vault")),
		ContainerRegistryID:   stringPtr(mapStringValue(properties, "containerRegistry", "container_registry")),
		ApplicationInsightsID: stringPtr(mapStringValue(properties, "applicationInsights", "application_insights")),
		Summary: azureMLOperatorSummary(
			len(computes),
			len(jobs),
			len(schedules),
			len(endpoints),
			identityType,
			len(datastores),
			classification,
		),
		RelatedIDs: azureMLRelatedIDs(
			workspaceID,
			identityIDs,
			mapStringValue(properties, "storageAccount", "storage_account"),
			mapStringValue(properties, "keyVault", "key_vault"),
			mapStringValue(properties, "containerRegistry", "container_registry"),
			mapStringValue(properties, "applicationInsights", "application_insights"),
			azureMLResourceIDs(datastores),
		),
	}
}

func azureMLIdentityIDs(workspaceID string, identity map[string]any) []string {
	ids := sortedKeys(identity, "userAssignedIdentities", "user_assigned_identities")
	if identityIncludesType(stringPtr(mapStringValue(identity, "type")), "SystemAssigned") && workspaceID != "" {
		ids = append(ids, workspaceID+"/identities/system")
	}
	sort.Strings(ids)
	return dedupeStrings(ids)
}

func azureMLComputeTypes(computes []map[string]any) []string {
	types := []string{}
	for _, compute := range computes {
		properties := mapValue(compute, "properties")
		computeType := strings.TrimSpace(mapStringValue(properties, "computeType", "compute_type"))
		if computeType == "" {
			continue
		}
		types = append(types, computeType)
	}
	sort.Strings(types)
	return dedupeStrings(types)
}

func azureMLJobTypes(jobs []map[string]any) []string {
	types := []string{}
	for _, job := range jobs {
		properties := mapValue(job, "properties")
		jobType := strings.TrimSpace(firstNonEmpty(
			mapStringValue(properties, "jobType", "job_type"),
			mapStringValue(job, "kind"),
		))
		if jobType == "" {
			continue
		}
		types = append(types, jobType)
	}
	sort.Strings(types)
	return dedupeStrings(types)
}

func azureMLScheduleTriggerTypes(schedules []map[string]any) []string {
	types := []string{}
	for _, schedule := range schedules {
		properties := mapValue(schedule, "properties")
		trigger := mapValue(properties, "trigger")
		triggerType := strings.TrimSpace(mapStringValue(trigger, "triggerType", "trigger_type"))
		if triggerType == "" {
			continue
		}
		types = append(types, triggerType)
	}
	sort.Strings(types)
	return dedupeStrings(types)
}

func azureMLEndpointPosture(endpoints []map[string]any) ([]string, []string) {
	authModes := []string{}
	publicAccess := []string{}
	for _, endpoint := range endpoints {
		properties := mapValue(endpoint, "properties")
		if authMode := strings.TrimSpace(mapStringValue(properties, "authMode", "auth_mode")); authMode != "" {
			authModes = append(authModes, authMode)
		}
		if exposure := strings.TrimSpace(mapStringValue(properties, "publicNetworkAccess", "public_network_access")); exposure != "" {
			publicAccess = append(publicAccess, exposure)
		}
	}
	sort.Strings(authModes)
	sort.Strings(publicAccess)
	return dedupeStrings(authModes), dedupeStrings(publicAccess)
}

func azureMLDatastoreTypes(datastores []map[string]any) []string {
	types := []string{}
	for _, datastore := range datastores {
		properties := mapValue(datastore, "properties")
		datastoreType := strings.TrimSpace(mapStringValue(properties, "datastoreType", "datastore_type"))
		if datastoreType == "" {
			continue
		}
		types = append(types, datastoreType)
	}
	sort.Strings(types)
	return dedupeStrings(types)
}

func azureMLClassification(hasCompute bool, hasJobs bool, hasEndpoints bool, hasSchedules bool) string {
	switch {
	case hasCompute || hasJobs || hasEndpoints:
		return "execution-capable"
	case hasSchedules:
		return "supporting-persistence-context"
	default:
		return "supporting-context"
	}
}

func azureMLOperatorSummary(
	computeCount int,
	jobCount int,
	scheduleCount int,
	endpointCount int,
	identityType *string,
	datastoreCount int,
	classification string,
) string {
	parts := []string{}
	switch classification {
	case "execution-capable":
		parts = append(parts, "Visible Azure ML workspace already shows execution-capable runtime surfaces from the current control-plane read path.")
	case "supporting-persistence-context":
		parts = append(parts, "Visible Azure ML workspace shows repeatable schedule context, but the current read path does not yet prove a stronger runtime or serving surface.")
	default:
		parts = append(parts, "Visible Azure ML workspace currently reads more like supporting context than an active execution surface.")
	}
	if computeCount > 0 {
		parts = append(parts, fmt.Sprintf("%d compute target(s) visible.", computeCount))
	}
	if jobCount > 0 {
		parts = append(parts, fmt.Sprintf("%d job(s) visible.", jobCount))
	}
	if scheduleCount > 0 {
		parts = append(parts, fmt.Sprintf("%d schedule(s) visible.", scheduleCount))
	}
	if endpointCount > 0 {
		parts = append(parts, fmt.Sprintf("%d online endpoint(s) visible.", endpointCount))
	}
	if identityType != nil && *identityType != "" {
		parts = append(parts, "Workspace carries managed identity context ("+*identityType+").")
	}
	if datastoreCount > 0 {
		parts = append(parts, fmt.Sprintf("%d datastore relationship(s) visible.", datastoreCount))
	}
	switch classification {
	case "execution-capable":
		parts = append(parts, "The current control-plane read path still does not prove what notebook, model, or job payload actually runs here.")
	case "supporting-persistence-context":
		parts = append(parts, "The current control-plane read path does not yet prove what the scheduled work does or whether it keeps a durable re-entry path alive.")
	default:
		parts = append(parts, "The current control-plane read path does not yet confirm compute, job, schedule, or endpoint posture strong enough for a heavier execution claim.")
	}
	return strings.Join(parts, " ")
}

func azureMLRelatedIDs(
	workspaceID string,
	identityIDs []string,
	storageAccountID string,
	keyVaultID string,
	containerRegistryID string,
	applicationInsightsID string,
	datastoreIDs []string,
) []string {
	values := []string{workspaceID, storageAccountID, keyVaultID, containerRegistryID, applicationInsightsID}
	values = append(values, identityIDs...)
	values = append(values, datastoreIDs...)
	return dedupeStrings(values)
}

func azureMLResourceIDs(resources []map[string]any) []string {
	ids := []string{}
	for _, resource := range resources {
		if resourceID := strings.TrimSpace(mapStringValue(resource, "id")); resourceID != "" {
			ids = append(ids, resourceID)
		}
	}
	return dedupeStrings(ids)
}
