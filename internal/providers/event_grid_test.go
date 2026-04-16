package providers

import (
	"slices"
	"testing"
)

func TestEventGridRouteAssetExecutionCapableFunction(t *testing.T) {
	route := eventGridRouteAsset(map[string]any{
		"id":   "/subscriptions/sub/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/stlanding/providers/Microsoft.EventGrid/eventSubscriptions/to-function",
		"name": "to-function",
		"properties": map[string]any{
			"provisioningState": "Succeeded",
			"destination": map[string]any{
				"endpointType": "AzureFunction",
				"properties": map[string]any{
					"resourceId": "/subscriptions/sub/resourceGroups/rg-app/providers/Microsoft.Web/sites/fa-ingest/functions/BlobCreated",
				},
			},
			"filter": map[string]any{
				"includedEventTypes": []any{"Microsoft.Storage.BlobCreated"},
			},
		},
	})

	if route.Classification != "execution-capable" {
		t.Fatalf("expected execution-capable classification, got %q", route.Classification)
	}
	if route.SourceType != "Microsoft.Storage/storageAccounts" {
		t.Fatalf("expected storage-account source type, got %q", route.SourceType)
	}
	if route.DestinationTargetID == nil || *route.DestinationTargetID == "" {
		t.Fatalf("expected destination target id to be present")
	}
}

func TestEventGridRouteAssetWebhookStaysExternalCallback(t *testing.T) {
	route := eventGridRouteAsset(map[string]any{
		"id":   "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.EventGrid/topics/custom/providers/Microsoft.EventGrid/eventSubscriptions/to-webhook",
		"name": "to-webhook",
		"properties": map[string]any{
			"destination": map[string]any{
				"endpointType": "WebHook",
				"properties": map[string]any{
					"endpointUrl": "https://example.invalid/hook",
				},
			},
			"filter": map[string]any{},
		},
	})

	if route.Classification != "external-callback" {
		t.Fatalf("expected external-callback classification, got %q", route.Classification)
	}
	if !route.ExternalDelivery {
		t.Fatalf("expected external delivery to be true")
	}
	if route.DestinationTargetID != nil {
		t.Fatalf("expected no destination target id for webhook route, got %q", *route.DestinationTargetID)
	}
	if !slices.Equal(route.IncludedEventTypes, []string{"All"}) {
		t.Fatalf("expected default event types to be All, got %#v", route.IncludedEventTypes)
	}
}
