package commands

import (
	"testing"

	"harrierops-azure/internal/models"
)

func TestComputeControlResolveIdentityBindingNormalizesManagedIdentityIDs(t *testing.T) {
	workload := models.WorkloadSummary{
		AssetID:      "/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.Web/sites/app-uami-ctrl-a43cfa",
		AssetKind:    "AppService",
		AssetName:    "app-uami-ctrl-a43cfa",
		IdentityIDs:  []string{"/subscriptions/sub-1/resourcegroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app"},
		IdentityType: models.StringPtr("UserAssigned"),
	}
	surface := models.TokenCredentialSurfaceSummary{
		AssetID:        workload.AssetID,
		AssetKind:      workload.AssetKind,
		AssetName:      workload.AssetName,
		SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,
		RelatedIDs:     []string{"/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app"},
		AccessPath:     "token",
		Priority:       "high",
		Summary:        "App Service can request tokens through attached managed identity.",
		OperatorSignal: "managed identity token surface",
	}
	managedByID := map[string]models.ManagedIdentity{
		computeControlARMIDKey("/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app"): {
			ID:          "/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
			Name:        "ua-app",
			PrincipalID: models.StringPtr("principal-1"),
		},
	}

	binding := computeControlResolveIdentityBinding(surface, workload, managedByID, nil)
	if binding == nil {
		t.Fatal("expected identity binding, got nil")
	}
	if binding.PrincipalID != "principal-1" {
		t.Fatalf("expected principal-1, got %q", binding.PrincipalID)
	}
	if binding.IdentityName != "ua-app" {
		t.Fatalf("expected ua-app identity name, got %q", binding.IdentityName)
	}
}

func TestComputeControlAttachedIdentityBindingNormalizesAttachedAssetIDs(t *testing.T) {
	principalID := "principal-1"
	assetID := "/subscriptions/sub-1/resourcegroups/rg-workload/providers/Microsoft.Web/sites/app-uami-ctrl-a43cfa"
	managedByPrincipal := map[string][]models.ManagedIdentity{
		principalID: {
			{
				ID:          "/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
				Name:        "ua-app",
				PrincipalID: models.StringPtr(principalID),
				AttachedTo:  []string{"/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.Web/sites/app-uami-ctrl-a43cfa"},
			},
		},
	}

	binding := computeControlAttachedIdentityBinding(principalID, assetID, managedByPrincipal)
	if binding == nil {
		t.Fatal("expected attachment binding, got nil")
	}
	if binding.IdentityName != "ua-app" {
		t.Fatalf("expected ua-app identity name, got %q", binding.IdentityName)
	}
}

func TestComputeControlResolveMixedIdentityBindingNormalizesUserAssignedCorroborationIDs(t *testing.T) {
	workload := models.WorkloadSummary{
		AssetID:             "/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.Web/sites/func-orders-a43cfa",
		AssetKind:           "FunctionApp",
		AssetName:           "func-orders-a43cfa",
		IdentityPrincipalID: models.StringPtr("system-principal"),
		IdentityIDs:         []string{"/subscriptions/sub-1/resourcegroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app"},
	}
	envRows := []models.EnvVarSummary{
		{
			AssetID:                   workload.AssetID,
			AssetKind:                 workload.AssetKind,
			AssetName:                 workload.AssetName,
			SettingName:               "PAYMENT_API_KEY",
			KeyVaultReferenceIdentity: models.StringPtr("/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app"),
		},
	}
	managedByID := map[string]models.ManagedIdentity{
		computeControlARMIDKey("/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app"): {
			ID:          "/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
			Name:        "ua-app",
			PrincipalID: models.StringPtr("principal-1"),
		},
	}
	surface := models.TokenCredentialSurfaceSummary{
		AssetID:        workload.AssetID,
		AssetKind:      workload.AssetKind,
		AssetName:      workload.AssetName,
		SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,
		AccessPath:     "token",
		Priority:       "high",
		Summary:        "Function App can request tokens through attached managed identity.",
		OperatorSignal: "managed identity token surface",
	}

	binding := computeControlResolveMixedIdentityBinding(surface, workload, envRows, managedByID, nil)
	if binding == nil {
		t.Fatal("expected mixed identity binding, got nil")
	}
	if binding.IdentityName != "ua-app" {
		t.Fatalf("expected ua-app identity name, got %q", binding.IdentityName)
	}
	if binding.IdentityChoiceBasis != "env-vars:PAYMENT_API_KEY" {
		t.Fatalf("expected env-vars corroboration basis, got %q", binding.IdentityChoiceBasis)
	}
}
