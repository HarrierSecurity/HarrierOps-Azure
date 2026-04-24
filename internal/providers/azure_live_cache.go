package providers

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"

	"harrierops-azure/internal/models"
)

type liveAzureCache struct {
	mu                   sync.Mutex
	sessions             map[string]*onceValue[azureSession]
	webAppsStates        map[string]*liveWebAppsState
	computeStates        map[string]*liveComputeNetworkState
	devopsOrgDiscoveries map[string]*onceValue[devopsOrganizationDiscovery]
}

func newLiveAzureCache() *liveAzureCache {
	return &liveAzureCache{
		sessions:             map[string]*onceValue[azureSession]{},
		webAppsStates:        map[string]*liveWebAppsState{},
		computeStates:        map[string]*liveComputeNetworkState{},
		devopsOrgDiscoveries: map[string]*onceValue[devopsOrganizationDiscovery]{},
	}
}

type onceValue[T any] struct {
	mu    sync.Mutex
	ready chan struct{}
	value T
	valid bool
}

func (value *onceValue[T]) get(load func() (T, error)) (T, error) {
	for {
		value.mu.Lock()
		if value.valid {
			cached := value.value
			value.mu.Unlock()
			return cached, nil
		}
		if value.ready != nil {
			ready := value.ready
			value.mu.Unlock()
			<-ready
			continue
		}
		value.ready = make(chan struct{})
		value.mu.Unlock()

		result, err := load()

		value.mu.Lock()
		if err == nil {
			value.value = result
			value.valid = true
		}
		ready := value.ready
		value.ready = nil
		value.mu.Unlock()

		close(ready)
		if err != nil {
			return result, err
		}
		return result, nil
	}
}

func sessionRequestKey(tenant string, subscription string) string {
	return tenant + "::" + subscription
}

func sessionStateKey(session azureSession) string {
	return session.tenantID + "::" + session.subscription.ID
}

func (provider AzureProvider) cachedSession(ctx context.Context, tenant string, subscription string) (azureSession, error) {
	if provider.cache == nil {
		return provider.buildSession(ctx, tenant, subscription)
	}

	cacheKey := sessionRequestKey(tenant, subscription)

	provider.cache.mu.Lock()
	entry := provider.cache.sessions[cacheKey]
	if entry == nil {
		entry = &onceValue[azureSession]{}
		provider.cache.sessions[cacheKey] = entry
	}
	provider.cache.mu.Unlock()

	return entry.get(func() (azureSession, error) {
		return provider.buildSession(ctx, tenant, subscription)
	})
}

func (provider AzureProvider) buildSession(ctx context.Context, tenant string, subscription string) (azureSession, error) {
	credential, tokenSource, authMode, claims, tenantID, err := newAzureCredential(ctx, tenant)
	if err != nil {
		return azureSession{}, err
	}

	subscriptionsClient, err := armsubscriptions.NewClient(credential, nil)
	if err != nil {
		return azureSession{}, fmt.Errorf("build subscriptions client: %w", err)
	}

	subscriptionRef, err := resolveSubscription(ctx, subscriptionsClient, subscription)
	if err != nil {
		return azureSession{}, err
	}

	clientFactory, err := armresources.NewClientFactory(subscriptionRef.ID, credential, nil)
	if err != nil {
		return azureSession{}, fmt.Errorf("build resource client factory: %w", err)
	}

	return azureSession{
		claims:        claims,
		credential:    credential,
		tokenSource:   tokenSource,
		authMode:      authMode,
		tenantID:      tenantID,
		subscription:  subscriptionRef,
		clientFactory: clientFactory,
	}, nil
}

type liveWebAppResource struct {
	appMap            map[string]any
	assetKind         string
	resourceGroup     string
	name              string
	config            *onceValue[map[string]any]
	connectionStrings *onceValue[map[string]any]
	functions         *onceValue[[]models.FunctionChildAsset]
	settings          *onceValue[map[string]any]
	sourceControl     *onceValue[map[string]any]
}

const armWebAppsAPIVersion = "2021-03-01"

type liveWebAppsState struct {
	client         *armappservice.WebAppsClient
	credential     azcore.TokenCredential
	subscriptionID string
	apps           *onceValue[[]*liveWebAppResource]
}

func (provider AzureProvider) webAppsState(session azureSession) (*liveWebAppsState, error) {
	if provider.cache == nil {
		client, err := armappservice.NewWebAppsClient(session.subscription.ID, session.credential, nil)
		if err != nil {
			return nil, fmt.Errorf("build web apps client: %w", err)
		}
		return &liveWebAppsState{
			client:         client,
			credential:     session.credential,
			subscriptionID: session.subscription.ID,
			apps:           &onceValue[[]*liveWebAppResource]{},
		}, nil
	}

	cacheKey := sessionStateKey(session)

	provider.cache.mu.Lock()
	state := provider.cache.webAppsStates[cacheKey]
	if state == nil {
		client, err := armappservice.NewWebAppsClient(session.subscription.ID, session.credential, nil)
		if err != nil {
			provider.cache.mu.Unlock()
			return nil, fmt.Errorf("build web apps client: %w", err)
		}
		state = &liveWebAppsState{
			client:         client,
			credential:     session.credential,
			subscriptionID: session.subscription.ID,
			apps:           &onceValue[[]*liveWebAppResource]{},
		}
		provider.cache.webAppsStates[cacheKey] = state
	}
	provider.cache.mu.Unlock()

	return state, nil
}

func (state *liveWebAppsState) list(ctx context.Context) ([]*liveWebAppResource, error) {
	return state.apps.get(func() ([]*liveWebAppResource, error) {
		apps := []*liveWebAppResource{}
		if state.subscriptionID == "" {
			return apps, nil
		}

		appMaps, err := armListObjects(
			ctx,
			state.credential,
			"/subscriptions/"+state.subscriptionID+"/providers/Microsoft.Web/sites",
			armWebAppsAPIVersion,
		)
		if err != nil {
			return apps, err
		}
		for _, appMap := range appMaps {
			apps = append(apps, newLiveWebAppResource(appMap))
		}
		return apps, nil
	})
}

func newLiveWebAppResource(appMap map[string]any) *liveWebAppResource {
	return &liveWebAppResource{
		appMap:            appMap,
		assetKind:         webAssetKind(mapStringValue(appMap, "kind")),
		resourceGroup:     resourceGroupFromID(mapStringValue(appMap, "id")),
		name:              mapStringValue(appMap, "name"),
		config:            &onceValue[map[string]any]{},
		connectionStrings: &onceValue[map[string]any]{},
		functions:         &onceValue[[]models.FunctionChildAsset]{},
		settings:          &onceValue[map[string]any]{},
		sourceControl:     &onceValue[map[string]any]{},
	}
}

func (state *liveWebAppsState) configMap(ctx context.Context, app *liveWebAppResource) (map[string]any, error) {
	if app.resourceGroup == "" || app.name == "" {
		return map[string]any{}, nil
	}
	return app.config.get(func() (map[string]any, error) {
		config, err := state.client.GetConfiguration(ctx, app.resourceGroup, app.name, nil)
		if err != nil {
			return map[string]any{}, err
		}
		configMap := map[string]any{}
		decodeJSONInto(config.SiteConfigResource, &configMap)
		return configMap, nil
	})
}

func (state *liveWebAppsState) settingsMap(ctx context.Context, app *liveWebAppResource) (map[string]any, error) {
	if app.resourceGroup == "" || app.name == "" {
		return map[string]any{}, nil
	}
	return app.settings.get(func() (map[string]any, error) {
		settings, err := state.client.ListApplicationSettings(ctx, app.resourceGroup, app.name, nil)
		if err != nil {
			return map[string]any{}, err
		}
		settingsMap := map[string]any{}
		decodeJSONInto(settings.StringDictionary, &settingsMap)
		return settingsMap, nil
	})
}

func (state *liveWebAppsState) connectionStringsMap(ctx context.Context, app *liveWebAppResource) (map[string]any, error) {
	if app.resourceGroup == "" || app.name == "" {
		return map[string]any{}, nil
	}
	return app.connectionStrings.get(func() (map[string]any, error) {
		connectionStrings, err := state.client.ListConnectionStrings(ctx, app.resourceGroup, app.name, nil)
		if err != nil {
			return map[string]any{}, err
		}
		connectionStringsMap := map[string]any{}
		decodeJSONInto(connectionStrings.ConnectionStringDictionary, &connectionStringsMap)
		return connectionStringsMap, nil
	})
}

func (state *liveWebAppsState) sourceControlMap(ctx context.Context, app *liveWebAppResource) (map[string]any, error) {
	if app.resourceGroup == "" || app.name == "" {
		return map[string]any{}, nil
	}
	return app.sourceControl.get(func() (map[string]any, error) {
		sourceControl, err := state.client.GetSourceControl(ctx, app.resourceGroup, app.name, nil)
		if err != nil {
			var responseErr *azcore.ResponseError
			if errors.As(err, &responseErr) && responseErr.StatusCode == 404 {
				return map[string]any{}, nil
			}
			return map[string]any{}, err
		}
		sourceControlMap := map[string]any{}
		decodeJSONInto(sourceControl.SiteSourceControl, &sourceControlMap)
		return sourceControlMap, nil
	})
}

func (state *liveWebAppsState) functionAssets(ctx context.Context, app *liveWebAppResource) ([]models.FunctionChildAsset, error) {
	if app.resourceGroup == "" || app.name == "" {
		return []models.FunctionChildAsset{}, nil
	}
	return app.functions.get(func() ([]models.FunctionChildAsset, error) {
		rows := []models.FunctionChildAsset{}
		pager := state.client.NewListFunctionsPager(app.resourceGroup, app.name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return rows, err
			}
			for _, function := range page.Value {
				if function == nil {
					continue
				}
				rows = append(rows, functionChildAssetFromEnvelope(function))
			}
		}
		return rows, nil
	})
}

func functionChildAssetFromEnvelope(function *armappservice.FunctionEnvelope) models.FunctionChildAsset {
	data := map[string]any{}
	decodeJSONInto(function, &data)
	return functionChildAssetFromMap(data)
}

type liveNICSnapshot struct {
	assets []models.NicAsset
	byID   map[string]models.NicAsset
	issues []models.Issue
}

type liveVMSnapshot struct {
	assets []models.VmAsset
	issues []models.Issue
}

type liveVMSSSnapshot struct {
	assets []models.VmssAsset
	issues []models.Issue
}

type liveSnapshotDiskSnapshot struct {
	assets []models.SnapshotDiskAsset
	issues []models.Issue
}

type liveComputeNetworkState struct {
	collector computeNetworkCollector
	nics      *onceValue[liveNICSnapshot]
	vms       *onceValue[liveVMSnapshot]
	vmss      *onceValue[liveVMSSSnapshot]
	snapshots *onceValue[liveSnapshotDiskSnapshot]
}

func (provider AzureProvider) computeNetworkState(session azureSession) (*liveComputeNetworkState, error) {
	if provider.cache == nil {
		collector, err := newComputeNetworkCollector(session)
		if err != nil {
			return nil, err
		}
		return &liveComputeNetworkState{
			collector: collector,
			nics:      &onceValue[liveNICSnapshot]{},
			vms:       &onceValue[liveVMSnapshot]{},
			vmss:      &onceValue[liveVMSSSnapshot]{},
			snapshots: &onceValue[liveSnapshotDiskSnapshot]{},
		}, nil
	}

	cacheKey := sessionStateKey(session)

	provider.cache.mu.Lock()
	state := provider.cache.computeStates[cacheKey]
	if state == nil {
		collector, err := newComputeNetworkCollector(session)
		if err != nil {
			provider.cache.mu.Unlock()
			return nil, err
		}
		state = &liveComputeNetworkState{
			collector: collector,
			nics:      &onceValue[liveNICSnapshot]{},
			vms:       &onceValue[liveVMSnapshot]{},
			vmss:      &onceValue[liveVMSSSnapshot]{},
			snapshots: &onceValue[liveSnapshotDiskSnapshot]{},
		}
		provider.cache.computeStates[cacheKey] = state
	}
	provider.cache.mu.Unlock()

	return state, nil
}

func (state *liveComputeNetworkState) nicSnapshot(ctx context.Context) liveNICSnapshot {
	snapshot, _ := state.nics.get(func() (liveNICSnapshot, error) {
		assets, byID, issues := state.collector.collectNICAssets(ctx)
		return liveNICSnapshot{assets: assets, byID: byID, issues: issues}, nil
	})
	return snapshot
}

func (state *liveComputeNetworkState) vmSnapshot(ctx context.Context) liveVMSnapshot {
	snapshot, _ := state.vms.get(func() (liveVMSnapshot, error) {
		nics := state.nicSnapshot(ctx)
		assets, issues := state.collector.collectVMAssets(ctx, nics.byID)
		return liveVMSnapshot{assets: assets, issues: issues}, nil
	})
	return snapshot
}

func (state *liveComputeNetworkState) vmssSnapshot(ctx context.Context) liveVMSSSnapshot {
	snapshot, _ := state.vmss.get(func() (liveVMSSSnapshot, error) {
		assets, issues := state.collector.collectVMSSAssets(ctx)
		return liveVMSSSnapshot{assets: assets, issues: issues}, nil
	})
	return snapshot
}

func (state *liveComputeNetworkState) snapshotDiskSnapshot(ctx context.Context) liveSnapshotDiskSnapshot {
	snapshot, _ := state.snapshots.get(func() (liveSnapshotDiskSnapshot, error) {
		assets, issues := state.collector.collectSnapshotDiskAssets(ctx)
		return liveSnapshotDiskSnapshot{assets: assets, issues: issues}, nil
	})
	return snapshot
}
