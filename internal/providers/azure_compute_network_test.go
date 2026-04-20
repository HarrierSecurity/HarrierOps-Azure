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
