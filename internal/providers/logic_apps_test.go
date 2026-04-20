package providers

import "testing"

func TestLogicAppWorkflowAssetClassifiesRequestTriggerAsPersistenceCapable(t *testing.T) {
	asset := logicAppWorkflowAsset(map[string]any{
		"id":       "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Logic/workflows/la-request",
		"name":     "la-request",
		"location": "centralus",
		"identity": map[string]any{"type": "SystemAssigned"},
		"properties": map[string]any{
			"state": "Enabled",
			"definition": map[string]any{
				"triggers": map[string]any{
					"manual": map[string]any{"type": "Request"},
				},
				"actions": map[string]any{
					"notify": map[string]any{"type": "Http"},
				},
			},
		},
	})

	if asset.Classification != "persistence-capable" {
		t.Fatalf("expected persistence-capable classification, got %q", asset.Classification)
	}
	if !asset.ExternallyCallableRequestTrigger {
		t.Fatalf("expected request trigger to be marked externally callable")
	}
	if len(asset.DownstreamActionKinds) != 1 || asset.DownstreamActionKinds[0] != "external-http" {
		t.Fatalf("expected external-http downstream, got %#v", asset.DownstreamActionKinds)
	}
}

func TestLogicAppWorkflowAssetCollectsNestedActionCategories(t *testing.T) {
	asset := logicAppWorkflowAsset(map[string]any{
		"id":   "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Logic/workflows/la-nested",
		"name": "la-nested",
		"properties": map[string]any{
			"definition": map[string]any{
				"triggers": map[string]any{
					"when_event_happens": map[string]any{"type": "ApiConnection"},
				},
				"actions": map[string]any{
					"scope": map[string]any{
						"type": "Scope",
						"actions": map[string]any{
							"call_function": map[string]any{"type": "Function"},
							"send_message":  map[string]any{"type": "ApiConnection", "inputs": map[string]any{"path": "/servicebus/queues/foo/messages"}},
						},
					},
				},
			},
		},
	})

	if asset.Classification != "execution-capable-only" {
		t.Fatalf("expected execution-capable-only classification, got %q", asset.Classification)
	}
	if got := asset.DownstreamActionKinds; len(got) != 2 || got[0] != "function" || got[1] != "messaging" {
		t.Fatalf("unexpected downstream categories: %#v", got)
	}
}

func TestLogicAppWorkflowAssetPreservesOnlySystemIdentityPath(t *testing.T) {
	workflowID := "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Logic/workflows/la-inbound"
	asset := logicAppWorkflowAsset(map[string]any{
		"id":   workflowID,
		"name": "la-inbound",
		"identity": map[string]any{
			"type":        "SystemAssigned",
			"principalId": "principal-value",
			"tenantId":    "tenant-value",
		},
		"properties": map[string]any{
			"definition": map[string]any{},
		},
	})

	wantIdentityID := workflowID + "/identities/system"
	if len(asset.IdentityIDs) != 1 || asset.IdentityIDs[0] != wantIdentityID {
		t.Fatalf("logicAppWorkflowAsset().IdentityIDs = %#v, want only system identity path", asset.IdentityIDs)
	}
	if len(asset.RelatedIDs) != 2 || asset.RelatedIDs[0] != workflowID || asset.RelatedIDs[1] != wantIdentityID {
		t.Fatalf("logicAppWorkflowAsset().RelatedIDs = %#v, want workflow ID plus system identity path", asset.RelatedIDs)
	}
}

func TestLogicAppWorkflowAssetPreservesUserAssignedIdentityIDs(t *testing.T) {
	userAssignedID := "/subscriptions/sub/resourceGroups/rg-id/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-workflow-router"
	asset := logicAppWorkflowAsset(map[string]any{
		"id":   "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Logic/workflows/la-router",
		"name": "la-router",
		"identity": map[string]any{
			"type": "UserAssigned",
			"userAssignedIdentities": map[string]any{
				userAssignedID: map[string]any{},
			},
		},
		"properties": map[string]any{
			"definition": map[string]any{},
		},
	})

	if len(asset.IdentityIDs) != 1 || asset.IdentityIDs[0] != userAssignedID {
		t.Fatalf("logicAppWorkflowAsset().IdentityIDs = %#v, want user-assigned identity ID", asset.IdentityIDs)
	}
}
