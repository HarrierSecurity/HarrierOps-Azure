package providers

import (
	"context"
	"fmt"
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

func TestEventGridEnumerationScopesBuildsTopicTypePaths(t *testing.T) {
	scopes := eventGridEnumerationScopes("sub", []map[string]any{
		{
			"name": "Microsoft.Resources.Subscriptions",
			"properties": map[string]any{
				"resourceRegionType": "GlobalResource",
			},
		},
		{
			"name": "Microsoft.Storage.StorageAccounts",
			"properties": map[string]any{
				"resourceRegionType": "RegionalResource",
				"supportedLocations": []any{"centralus", "westus"},
			},
		},
		{
			"name": "Microsoft.KeyVault.Vaults",
			"properties": map[string]any{
				"supportedScopesForSource": []any{"Resource"},
				"supportedLocations":       []any{"eastus", "centralus"},
			},
		},
	}, []string{"centralus", "eastus"})

	got := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		got = append(got, scope.path)
	}

	want := []string{
		"/subscriptions/sub/providers/Microsoft.EventGrid/topicTypes/Microsoft.Resources.Subscriptions/eventSubscriptions",
		"/subscriptions/sub/providers/Microsoft.EventGrid/locations/centralus/topicTypes/Microsoft.Storage.StorageAccounts/eventSubscriptions",
		"/subscriptions/sub/providers/Microsoft.EventGrid/locations/centralus/topicTypes/Microsoft.KeyVault.Vaults/eventSubscriptions",
		"/subscriptions/sub/providers/Microsoft.EventGrid/locations/eastus/topicTypes/Microsoft.KeyVault.Vaults/eventSubscriptions",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("eventGridEnumerationScopes() = %#v, want %#v", got, want)
	}
}

func TestEventGridTopicTypeLocationsFallsBackWhenSupportedLocationsMissing(t *testing.T) {
	got := eventGridTopicTypeLocations(map[string]any{}, []string{"centralus", "eastus", "centralus"})
	want := []string{"centralus", "eastus"}
	if !slices.Equal(got, want) {
		t.Fatalf("eventGridTopicTypeLocations() = %#v, want %#v", got, want)
	}
}

func TestEventGridItemsFromScopesDedupesAndIgnoresMissingScopes(t *testing.T) {
	rows, issues := eventGridItemsFromScopes(context.Background(), []eventGridEnumerationScope{
		{path: "subscription", issueScope: "event-grid.topic-type[global]"},
		{path: "regional", issueScope: "event-grid.topic-type[storage@centralus]"},
		{path: "missing", issueScope: "event-grid.topic-type[storage@eastus]"},
		{path: "broken", issueScope: "event-grid.topic-type[keyvault@centralus]"},
	}, func(_ context.Context, path string) ([]map[string]any, error) {
		switch path {
		case "subscription":
			return []map[string]any{
				{"id": "/subscriptions/sub/providers/Microsoft.EventGrid/eventSubscriptions/subscription-route"},
			}, nil
		case "regional":
			return []map[string]any{
				{"id": "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/st/providers/Microsoft.EventGrid/eventSubscriptions/storage-route"},
				{"id": "/subscriptions/sub/providers/Microsoft.EventGrid/eventSubscriptions/subscription-route"},
			}, nil
		case "missing":
			return nil, fmt.Errorf("GET https://management.azure.com/example: 404 Not Found")
		case "broken":
			return nil, fmt.Errorf("GET https://management.azure.com/example: 500 Internal Server Error")
		default:
			return nil, fmt.Errorf("unexpected path %q", path)
		}
	})

	gotIDs := []string{}
	for _, row := range rows {
		gotIDs = append(gotIDs, mapStringValue(row, "id"))
	}
	wantIDs := []string{
		"/subscriptions/sub/providers/Microsoft.EventGrid/eventSubscriptions/subscription-route",
		"/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/st/providers/Microsoft.EventGrid/eventSubscriptions/storage-route",
	}
	if !slices.Equal(gotIDs, wantIDs) {
		t.Fatalf("eventGridItemsFromScopes() ids = %#v, want %#v", gotIDs, wantIDs)
	}

	if len(issues) != 1 {
		t.Fatalf("expected one surfaced issue, got %#v", issues)
	}
	if issues[0].Scope != "event-grid.topic-type[keyvault@centralus]" {
		t.Fatalf("unexpected issue scope: %#v", issues[0])
	}
}
