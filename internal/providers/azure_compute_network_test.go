package providers

import (
	"testing"

	"harrierops-azure/internal/models"
)

func TestEndpointsFromContainerInstancesPreserveWorkloadIdentityIDs(t *testing.T) {
	identityID := "/subscriptions/sub/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app"
	endpoints := endpointsFromContainerInstances([]models.ContainerInstanceAsset{
		{
			ID:                  "/subscriptions/sub/resourceGroups/rg-workload/providers/Microsoft.ContainerInstance/containerGroups/aci-web",
			Name:                "aci-web",
			PublicIPAddress:     stringPtr("52.165.253.164"),
			WorkloadPrincipalID: stringPtr("principal-id"),
			WorkloadIdentityIDs: []string{identityID},
		},
	})

	if len(endpoints) != 1 {
		t.Fatalf("endpointsFromContainerInstances() len = %d, want 1", len(endpoints))
	}
	if !containsStringValue(endpoints[0].RelatedIDs, identityID) {
		t.Fatalf("endpointsFromContainerInstances().RelatedIDs = %v, want user-assigned identity id present", endpoints[0].RelatedIDs)
	}
}

func TestComposeNetworkEffectivePreservesContainerInstanceIdentityIDs(t *testing.T) {
	identityID := "/subscriptions/sub/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app"
	endpoints := endpointsFromContainerInstances([]models.ContainerInstanceAsset{
		{
			ID:                  "/subscriptions/sub/resourceGroups/rg-workload/providers/Microsoft.ContainerInstance/containerGroups/aci-web",
			Name:                "aci-web",
			PublicIPAddress:     stringPtr("52.165.253.164"),
			WorkloadIdentityIDs: []string{identityID},
		},
	})

	rows := composeNetworkEffective(endpoints, nil)
	if len(rows) != 1 {
		t.Fatalf("composeNetworkEffective() len = %d, want 1", len(rows))
	}
	if !containsStringValue(rows[0].RelatedIDs, identityID) {
		t.Fatalf("composeNetworkEffective().RelatedIDs = %v, want user-assigned identity id present", rows[0].RelatedIDs)
	}
}

func TestTokenCredentialSurfacesFromEnvVarsIncludesFunctionAppAzureWebJobsStorage(t *testing.T) {
	identityID := "/subscriptions/sub/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app"
	surfaces := tokenCredentialSurfacesFromEnvVars([]models.EnvVarSummary{
		{
			AssetID:             "/subscriptions/sub/resourceGroups/rg-workload/providers/Microsoft.Web/sites/app-public-api",
			AssetKind:           "AppService",
			AssetName:           "app-public-api",
			Location:            "centralus",
			LooksSensitive:      false,
			ResourceGroup:       "rg-workload",
			SettingName:         "API_BASE_URL",
			ValueType:           "plain-text",
			WorkloadPrincipalID: stringPtr("principal-app"),
		},
		{
			AssetID:             "/subscriptions/sub/resourceGroups/rg-workload/providers/Microsoft.Web/sites/func-orders",
			AssetKind:           "FunctionApp",
			AssetName:           "func-orders",
			Location:            "centralus",
			LooksSensitive:      false,
			ResourceGroup:       "rg-workload",
			SettingName:         "AzureWebJobsStorage",
			ValueType:           "plain-text",
			WorkloadPrincipalID: stringPtr("principal-func"),
			WorkloadIdentityIDs: []string{identityID},
		},
	})

	if len(surfaces) != 1 {
		t.Fatalf("tokenCredentialSurfacesFromEnvVars() len = %d, want 1", len(surfaces))
	}
	if surfaces[0].SurfaceType != models.TokenCredentialSurfacePlainTextSecret {
		t.Fatalf("tokenCredentialSurfacesFromEnvVars().SurfaceType = %q, want %q", surfaces[0].SurfaceType, models.TokenCredentialSurfacePlainTextSecret)
	}
	if surfaces[0].AssetKind != "FunctionApp" || surfaces[0].AssetName != "func-orders" {
		t.Fatalf("tokenCredentialSurfacesFromEnvVars() asset = %s %s, want FunctionApp func-orders", surfaces[0].AssetKind, surfaces[0].AssetName)
	}
	if surfaces[0].OperatorSignal != "setting=AzureWebJobsStorage" {
		t.Fatalf("tokenCredentialSurfacesFromEnvVars().OperatorSignal = %q, want setting=AzureWebJobsStorage", surfaces[0].OperatorSignal)
	}
	if !containsStringValue(surfaces[0].RelatedIDs, identityID) {
		t.Fatalf("tokenCredentialSurfacesFromEnvVars().RelatedIDs = %v, want user-assigned identity id present", surfaces[0].RelatedIDs)
	}
	if !containsStringValue(surfaces[0].RelatedIDs, "principal-func") {
		t.Fatalf("tokenCredentialSurfacesFromEnvVars().RelatedIDs = %v, want principal id present", surfaces[0].RelatedIDs)
	}
}
