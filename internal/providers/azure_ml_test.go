package providers

import (
	"slices"
	"strings"
	"testing"
)

func TestAzureMLWorkspaceAssetExecutionCapable(t *testing.T) {
	asset := azureMLWorkspaceAsset(
		map[string]any{
			"id":       "/subscriptions/sub/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops",
			"name":     "ml-ops",
			"kind":     "Default",
			"location": "eastus",
			"identity": map[string]any{
				"type":        "SystemAssigned,UserAssigned",
				"principalId": "11111111-1111-1111-1111-111111111111",
				"userAssignedIdentities": map[string]any{
					"/subscriptions/sub/resourceGroups/rg-id/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-ml": map[string]any{},
				},
			},
			"properties": map[string]any{
				"publicNetworkAccess": "Enabled",
				"storageAccount":      "/subscriptions/sub/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stml",
				"keyVault":            "/subscriptions/sub/resourceGroups/rg-sec/providers/Microsoft.KeyVault/vaults/kvml",
			},
		},
		[]map[string]any{
			{
				"id": "/subscriptions/sub/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops/computes/cpu-train",
				"properties": map[string]any{
					"computeType": "ComputeInstance",
				},
			},
		},
		[]map[string]any{
			{
				"id": "/subscriptions/sub/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops/jobs/nightly-train",
				"properties": map[string]any{
					"jobType":   "Command",
					"computeId": "/subscriptions/sub/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops/computes/cpu-train",
				},
			},
		},
		nil,
		[]map[string]any{
			{
				"id": "/subscriptions/sub/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops/onlineEndpoints/fraud-score",
				"properties": map[string]any{
					"authMode":            "AADToken",
					"publicNetworkAccess": "Enabled",
				},
			},
		},
		[]map[string]any{
			{
				"id": "/subscriptions/sub/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-ops/datastores/workspaceblobstore",
				"properties": map[string]any{
					"datastoreType": "AzureBlob",
				},
			},
		},
	)

	if asset.Classification != "execution-capable" {
		t.Fatalf("expected execution-capable classification, got %q", asset.Classification)
	}
	if asset.ComputeCount != 1 || asset.JobCount != 1 || asset.EndpointCount != 1 {
		t.Fatalf("expected compute/job/endpoint counts to be present, got %#v", asset)
	}
	if !slices.Equal(asset.ComputeTypes, []string{"ComputeInstance"}) {
		t.Fatalf("unexpected compute types: %#v", asset.ComputeTypes)
	}
	if !slices.Equal(asset.JobTypes, []string{"Command"}) {
		t.Fatalf("unexpected job types: %#v", asset.JobTypes)
	}
	if !slices.Equal(asset.EndpointAuthModes, []string{"AADToken"}) {
		t.Fatalf("unexpected endpoint auth modes: %#v", asset.EndpointAuthModes)
	}
	if asset.DatastoreCount != 1 || !slices.Equal(asset.DatastoreTypes, []string{"AzureBlob"}) {
		t.Fatalf("unexpected datastore detail: count=%d types=%#v", asset.DatastoreCount, asset.DatastoreTypes)
	}
	if !strings.Contains(asset.Summary, "execution-capable") {
		t.Fatalf("expected execution-capable summary, got %q", asset.Summary)
	}
}

func TestAzureMLWorkspaceAssetScheduleOnlyBecomesSupportingPersistenceContext(t *testing.T) {
	asset := azureMLWorkspaceAsset(
		map[string]any{
			"id":   "/subscriptions/sub/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-schedule",
			"name": "ml-schedule",
		},
		nil,
		nil,
		[]map[string]any{
			{
				"id": "/subscriptions/sub/resourceGroups/rg-ml/providers/Microsoft.MachineLearningServices/workspaces/ml-schedule/schedules/weekly-run",
				"properties": map[string]any{
					"trigger": map[string]any{
						"triggerType": "Cron",
					},
				},
			},
		},
		nil,
		nil,
	)

	if asset.Classification != "supporting-persistence-context" {
		t.Fatalf("expected supporting-persistence-context classification, got %q", asset.Classification)
	}
	if asset.ScheduleCount != 1 || !slices.Equal(asset.ScheduleTriggerTypes, []string{"Cron"}) {
		t.Fatalf("unexpected schedule detail: count=%d types=%#v", asset.ScheduleCount, asset.ScheduleTriggerTypes)
	}
	if strings.Contains(asset.Summary, "what notebooks") {
		t.Fatalf("summary drifted into notebook overclaiming: %q", asset.Summary)
	}
}
