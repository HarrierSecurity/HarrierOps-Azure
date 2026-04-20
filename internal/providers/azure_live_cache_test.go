package providers

import (
	"errors"
	"testing"
)

func TestOnceValueRetriesAfterError(t *testing.T) {
	cache := &onceValue[int]{}
	loadCalls := 0
	wantErr := errors.New("transient failure")

	if _, err := cache.get(func() (int, error) {
		loadCalls++
		return 0, wantErr
	}); !errors.Is(err, wantErr) {
		t.Fatalf("onceValue.get() error = %v, want %v", err, wantErr)
	}

	value, err := cache.get(func() (int, error) {
		loadCalls++
		return 42, nil
	})
	if err != nil {
		t.Fatalf("onceValue.get() second call error = %v, want nil", err)
	}
	if value != 42 {
		t.Fatalf("onceValue.get() second call value = %d, want 42", value)
	}
	if loadCalls != 2 {
		t.Fatalf("onceValue.get() loadCalls = %d, want 2", loadCalls)
	}
}

func TestOnceValueCachesSuccessfulResult(t *testing.T) {
	cache := &onceValue[int]{}
	loadCalls := 0

	first, err := cache.get(func() (int, error) {
		loadCalls++
		return 7, nil
	})
	if err != nil {
		t.Fatalf("onceValue.get() first call error = %v, want nil", err)
	}
	second, err := cache.get(func() (int, error) {
		loadCalls++
		return 9, nil
	})
	if err != nil {
		t.Fatalf("onceValue.get() second call error = %v, want nil", err)
	}
	if first != 7 || second != 7 {
		t.Fatalf("onceValue.get() cached values = (%d, %d), want (7, 7)", first, second)
	}
	if loadCalls != 1 {
		t.Fatalf("onceValue.get() loadCalls = %d, want 1", loadCalls)
	}
}

func TestNewLiveWebAppResourcePreservesRawListMap(t *testing.T) {
	appMap := map[string]any{
		"id":                  "/subscriptions/sub/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-public-api",
		"kind":                "app,linux",
		"name":                "app-public-api",
		"publicNetworkAccess": "Enabled",
	}

	resource := newLiveWebAppResource(appMap)
	if resource.assetKind != "AppService" {
		t.Fatalf("newLiveWebAppResource().assetKind = %q, want AppService", resource.assetKind)
	}
	if resource.resourceGroup != "rg-apps" {
		t.Fatalf("newLiveWebAppResource().resourceGroup = %q, want rg-apps", resource.resourceGroup)
	}
	if resource.name != "app-public-api" {
		t.Fatalf("newLiveWebAppResource().name = %q, want app-public-api", resource.name)
	}
	if got := mapStringValue(resource.appMap, "publicNetworkAccess"); got != "Enabled" {
		t.Fatalf("newLiveWebAppResource().appMap publicNetworkAccess = %q, want Enabled", got)
	}
}
