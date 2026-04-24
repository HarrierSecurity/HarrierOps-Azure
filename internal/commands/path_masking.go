package commands

import (
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

const (
	pathMaskingCurrentBehavior = "Grouped pathmasking walkthroughs. Use `ho-azure pathmasking` or `ho-azure pathmasking help` to list surfaces, then `ho-azure pathmasking <surface>` to run an implemented surface."
	pathMaskingCommandState    = contracts.StatusImplemented
)

var (
	pathMaskingInputModes            = []string{"live"}
	pathMaskingPreferredArtifactMode = []string{"loot", "json"}
)

var pathMaskingSurfaceBuilders = map[string]groupedSurfaceBuilder{
	"api-mgmt":   buildPathMaskingAPIMOutput,
	"logic-apps": buildPathMaskingLogicAppsOutput,
	"relay":      buildPathMaskingRelayOutput,
}

func pathMaskingHandler(provider providers.Provider, now func() time.Time) Handler {
	return groupedFamilyHandler(provider, now, pathMaskingFamilyConfig())
}

func buildPathMaskingOverview(now func() time.Time, request Request, selectedSurface *string) any {
	config := pathMaskingFamilyConfig()
	return models.PathMaskingOverviewOutput{
		Metadata:               scopedMetadata(now, request, request.Tenant, request.Subscription, config.CommandName),
		GroupedCommandName:     config.CommandName,
		CommandState:           config.CommandState,
		CurrentBehavior:        config.CurrentBehavior,
		PlannedInputModes:      append([]string{}, config.InputModes...),
		PreferredArtifactOrder: append([]string{}, config.PreferredArtifactOrder...),
		SelectedSurface:        selectedSurface,
		Surfaces:               groupedFamilySurfaceDescriptors(config),
		Issues:                 []models.Issue{},
	}
}

func pathMaskingFamilyConfig() groupedFamilyConfig {
	return groupedFamilyConfig{
		CommandName:            "pathmasking",
		CurrentBehavior:        pathMaskingCurrentBehavior,
		CommandState:           pathMaskingCommandState,
		InputModes:             pathMaskingInputModes,
		PreferredArtifactOrder: pathMaskingPreferredArtifactMode,
		Selector:               func(request Request) string { return request.PathMaskingSurface },
		Overview:               buildPathMaskingOverview,
		SurfaceNames:           contracts.PathMaskingSurfaceNames,
		SurfaceContract:        contracts.PathMaskingSurface,
		SurfaceBuilders:        pathMaskingSurfaceBuilders,
	}
}
