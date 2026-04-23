package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

const (
	staticAzureMLWorkspacePrincipalID    = "5a5a5656-5656-5656-5656-565656565656"
	staticAzureMLUserAssignedPrincipalID = "6a6a5656-5656-5656-5656-565656565656"
	staticAzureMLUserAssignedClientID    = "7a7a5656-5656-5656-5656-565656565656"
)

func (StaticProvider) AzureML(_ context.Context, tenant string, subscription string) (AzureMLFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return AzureMLFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Workspaces: []models.AzureMLWorkspaceAsset{
			{
				ID:                   "/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub",
				Name:                 "ml-ops-hub",
				Classification:       "execution-capable",
				ResourceGroup:        "rg-ml",
				Location:             models.StringPtr("eastus"),
				WorkspaceKind:        models.StringPtr("Default"),
				State:                models.StringPtr("Succeeded"),
				PublicNetworkAccess:  models.StringPtr("Enabled"),
				IdentityType:         models.StringPtr("SystemAssigned,UserAssigned"),
				PrincipalID:          models.StringPtr(staticAzureMLWorkspacePrincipalID),
				IdentityIDs:          []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub/identities/system", "/subscriptions/" + subscriptionID + "/resourceGroups/rg-id/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-ml-ops"},
				ComputeCount:         2,
				ComputeTypes:         []string{"ComputeCluster", "ComputeInstance"},
				JobCount:             2,
				JobTypes:             []string{"Command", "Pipeline"},
				ScheduleCount:        1,
				ScheduleTriggerTypes: []string{"Cron"},
				EndpointCount:        1,
				EndpointAuthModes:    []string{"AADToken"},
				EndpointPublicAccess: []string{"Enabled"},
				DatastoreCount:       2,
				DatastoreTypes:       []string{"AzureBlob", "AzureDataLakeGen2"},
				StorageAccountID:     models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stamlops"),
				KeyVaultID:           models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-sec/providers/Microsoft.KeyVault/vaults/kv-amlops"),
				ContainerRegistryID:  models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-app/providers/Microsoft.ContainerRegistry/registries/cramlops"),
				Summary:              "Visible Azure ML workspace already shows execution-capable runtime surfaces through compute, jobs, and an online endpoint. Cron-backed scheduling is also visible, and the workspace carries managed identity plus linked datastore and storage context for follow-up.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub/identities/system",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-id/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-ml-ops",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stamlops",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-sec/providers/Microsoft.KeyVault/vaults/kv-amlops",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-app/providers/Microsoft.ContainerRegistry/registries/cramlops",
				},
			},
			{
				ID:                   "/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-nightly-train",
				Name:                 "ml-nightly-train",
				Classification:       "supporting-persistence-context",
				ResourceGroup:        "rg-ml",
				Location:             models.StringPtr("centralus"),
				WorkspaceKind:        models.StringPtr("Default"),
				State:                models.StringPtr("Succeeded"),
				PublicNetworkAccess:  models.StringPtr("Disabled"),
				ScheduleCount:        1,
				ScheduleTriggerTypes: []string{"Recurrence"},
				DatastoreCount:       1,
				DatastoreTypes:       []string{"AzureBlob"},
				StorageAccountID:     models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stmltrain"),
				Summary:              "Visible recurrence-backed scheduling makes this workspace relevant as persistence-adjacent ML context. The current control-plane read path does not yet prove a stronger compute, job, or serving surface behind that schedule.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-nightly-train",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stmltrain",
				},
			},
			{
				ID:                    "/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-catalog",
				Name:                  "ml-catalog",
				Classification:        "supporting-context",
				ResourceGroup:         "rg-ml",
				Location:              models.StringPtr("westus2"),
				WorkspaceKind:         models.StringPtr("Default"),
				State:                 models.StringPtr("Succeeded"),
				PublicNetworkAccess:   models.StringPtr("Enabled"),
				DatastoreCount:        2,
				DatastoreTypes:        []string{"AzureBlob", "AzureDataLakeGen2"},
				StorageAccountID:      models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stmlcatalog"),
				ApplicationInsightsID: models.StringPtr("/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.Insights/components/ai-mlcatalog"),
				Summary:               "Visible Azure ML workspace currently reads more like supporting context than an active execution surface. Storage-linked and datastore relationships are visible, but no compute, job, schedule, or online endpoint is confirmed from the current read path.",
				RelatedIDs: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-catalog",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stmlcatalog",
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.Insights/components/ai-mlcatalog",
				},
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func staticAzureMLRBACPrincipals(tenantID string) []models.Principal {
	return []models.Principal{
		{
			DisplayName:   "ml-ops-hub-workspace-identity",
			ID:            staticAzureMLWorkspacePrincipalID,
			PrincipalType: "ServicePrincipal",
			TenantID:      tenantID,
		},
		{
			DisplayName:   "ua-ml-ops",
			ID:            staticAzureMLUserAssignedPrincipalID,
			PrincipalType: "ServicePrincipal",
			TenantID:      tenantID,
		},
	}
}

func staticAzureMLRBACRoleAssignments(subscriptionID string) []models.RoleAssignment {
	return []models.RoleAssignment{
		{
			ID:               "ra-aml-1",
			PrincipalID:      staticAzureMLWorkspacePrincipalID,
			PrincipalType:    "ServicePrincipal",
			RoleDefinitionID: "rd-contributor",
			RoleName:         "Contributor",
			ScopeID:          "/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml",
		},
		{
			ID:               "ra-aml-2",
			PrincipalID:      staticAzureMLUserAssignedPrincipalID,
			PrincipalType:    "ServicePrincipal",
			RoleDefinitionID: "rd-owner",
			RoleName:         "Owner",
			ScopeID:          "/subscriptions/" + subscriptionID,
		},
	}
}

func staticAzureMLPermissionFacts(subscriptionID string) []PermissionFact {
	return []PermissionFact{
		{
			PrincipalID:         staticAzureMLWorkspacePrincipalID,
			DisplayName:         "ml-ops-hub-workspace-identity",
			PrincipalType:       "ServicePrincipal",
			HighImpactRoles:     []string{"Contributor"},
			AllRoleNames:        []string{"Contributor"},
			RoleAssignmentCount: 1,
			ScopeCount:          1,
			ScopeIDs:            []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml"},
			Privileged:          true,
			IsCurrentIdentity:   false,
		},
		{
			PrincipalID:         staticAzureMLUserAssignedPrincipalID,
			DisplayName:         "ua-ml-ops",
			PrincipalType:       "ServicePrincipal",
			HighImpactRoles:     []string{"Owner"},
			AllRoleNames:        []string{"Owner"},
			RoleAssignmentCount: 1,
			ScopeCount:          1,
			ScopeIDs:            []string{"/subscriptions/" + subscriptionID},
			Privileged:          true,
			IsCurrentIdentity:   false,
		},
	}
}

func staticAzureMLPermissionPrincipals(subscriptionID string) []PermissionPrincipalFact {
	return []PermissionPrincipalFact{
		{
			ID:            staticAzureMLWorkspacePrincipalID,
			Sources:       []string{"rbac"},
			IdentityNames: []string{"ml-ops-hub-workspace-identity"},
			AttachedTo: []string{
				"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub",
			},
		},
		{
			ID:            staticAzureMLUserAssignedPrincipalID,
			Sources:       []string{"rbac", "managed-identities"},
			IdentityNames: []string{"ua-ml-ops"},
			AttachedTo: []string{
				"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub",
			},
		},
	}
}

func staticAzureMLManagedIdentities(subscriptionID string) []models.ManagedIdentity {
	workspaceID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub"
	return []models.ManagedIdentity{
		{
			ID:           workspaceID + "/identities/system",
			Name:         "ml-ops-hub-workspace-identity",
			IdentityType: "systemAssigned",
			PrincipalID:  models.StringPtr(staticAzureMLWorkspacePrincipalID),
			AttachedTo:   []string{workspaceID},
			ScopeIDs:     []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml"},
			OperatorSignal: models.StringPtr(
				"Public Azure ML workload pivot; direct control visible.",
			),
			NextReview: models.StringPtr(
				"Check azure-ml for the workspace, compute, and execution context behind this workload pivot.",
			),
			Summary: models.StringPtr(
				"Azure ML workspace 'ml-ops-hub' gives a public Azure ML workload pivot into managed identity 'ml-ops-hub-workspace-identity'. Current scope already shows direct control through high-impact roles (Contributor). Check azure-ml for the workspace, compute, and execution context behind this workload pivot.",
			),
			WorkloadExposure:     models.WorkloadExposurePublic,
			DirectControlVisible: true,
		},
		{
			ID:           "/subscriptions/" + subscriptionID + "/resourceGroups/rg-id/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-ml-ops",
			Name:         "ua-ml-ops",
			IdentityType: "userAssigned",
			PrincipalID:  models.StringPtr(staticAzureMLUserAssignedPrincipalID),
			ClientID:     models.StringPtr(staticAzureMLUserAssignedClientID),
			AttachedTo:   []string{workspaceID},
			ScopeIDs:     []string{"/subscriptions/" + subscriptionID},
			OperatorSignal: models.StringPtr(
				"Public Azure ML workload pivot; direct control visible.",
			),
			NextReview: models.StringPtr(
				"Check azure-ml for the workspace, compute, and execution context behind this workload pivot.",
			),
			Summary: models.StringPtr(
				"Azure ML workspace 'ml-ops-hub' gives a public Azure ML workload pivot into managed identity 'ua-ml-ops'. Current scope already shows direct control through high-impact roles (Owner). Check azure-ml for the workspace, compute, and execution context behind this workload pivot.",
			),
			WorkloadExposure:     models.WorkloadExposurePublic,
			DirectControlVisible: true,
		},
	}
}

func staticAzureMLManagedIdentityRoleAssignments(subscriptionID string) []models.ManagedIdentityRoleAssignment {
	return []models.ManagedIdentityRoleAssignment{
		{
			ID:               "ra-aml-1",
			ScopeID:          "/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml",
			PrincipalID:      staticAzureMLWorkspacePrincipalID,
			PrincipalType:    "ServicePrincipal",
			RoleDefinitionID: "rd-contributor",
			RoleName:         "Contributor",
		},
		{
			ID:               "ra-aml-2",
			ScopeID:          "/subscriptions/" + subscriptionID,
			PrincipalID:      staticAzureMLUserAssignedPrincipalID,
			PrincipalType:    "ServicePrincipal",
			RoleDefinitionID: "rd-owner",
			RoleName:         "Owner",
		},
	}
}

func staticAzureMLManagedIdentityFindings(subscriptionID string) []models.ManagedIdentityFinding {
	return []models.ManagedIdentityFinding{
		{
			ID:          "identity-privileged-/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub/identities/system",
			Severity:    "high",
			Title:       "Managed identity has elevated role assignment",
			Description: "Identity 'ml-ops-hub-workspace-identity' is assigned one or more high-impact roles (Contributor).",
			RelatedIDs: []string{
				"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub/identities/system",
				"ra-aml-1",
			},
		},
		{
			ID:          "identity-privileged-/subscriptions/" + subscriptionID + "/resourceGroups/rg-id/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-ml-ops",
			Severity:    "high",
			Title:       "Managed identity has elevated role assignment",
			Description: "Identity 'ua-ml-ops' is assigned one or more high-impact roles (Owner).",
			RelatedIDs: []string{
				"/subscriptions/" + subscriptionID + "/resourceGroups/rg-id/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-ml-ops",
				"ra-aml-2",
			},
		},
	}
}

func staticAzureMLPrincipalSummaries(tenantID string, subscriptionID string) []models.PrincipalSummary {
	workspaceID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops-hub"
	return []models.PrincipalSummary{
		{
			AttachedTo: []string{workspaceID},
			DisplayName: models.StringPtr(
				"ml-ops-hub-workspace-identity",
			),
			ID:                  staticAzureMLWorkspacePrincipalID,
			IdentityNames:       []string{"ml-ops-hub-workspace-identity"},
			IdentityTypes:       []string{"systemAssigned"},
			IsCurrentIdentity:   false,
			PrincipalType:       "ServicePrincipal",
			RoleAssignmentCount: 1,
			RoleNames:           []string{"Contributor"},
			ScopeIDs:            []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-ml"},
			Sources:             []string{"managed-identities"},
			TenantID:            models.StringPtr(tenantID),
		},
		{
			AttachedTo: []string{workspaceID},
			DisplayName: models.StringPtr(
				"ua-ml-ops",
			),
			ID:                  staticAzureMLUserAssignedPrincipalID,
			IdentityNames:       []string{"ua-ml-ops"},
			IdentityTypes:       []string{"userAssigned"},
			IsCurrentIdentity:   false,
			PrincipalType:       "ServicePrincipal",
			RoleAssignmentCount: 1,
			RoleNames:           []string{"Owner"},
			ScopeIDs:            []string{"/subscriptions/" + subscriptionID},
			Sources:             []string{"managed-identities"},
			TenantID:            models.StringPtr(tenantID),
		},
	}
}
