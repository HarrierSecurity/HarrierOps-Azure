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
