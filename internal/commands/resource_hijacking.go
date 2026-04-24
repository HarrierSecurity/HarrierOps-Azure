package commands

import (
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

const (
	resourceHijackingCurrentBehavior = "Grouped resourcehijacking walkthroughs. Use `ho-azure resourcehijacking` or `ho-azure resourcehijacking help` to list surfaces, then `ho-azure resourcehijacking <surface>` to run an implemented surface."
	resourceHijackingCommandState    = contracts.StatusImplemented
)

var (
	resourceHijackingInputModes            = []string{"live"}
	resourceHijackingPreferredArtifactMode = []string{"loot", "json"}
)

var resourceHijackingSurfaceBuilders = map[string]groupedSurfaceBuilder{
	"api-mgmt":   buildResourceHijackingAPIMOutput,
	"automation": buildResourceHijackingAutomationOutput,
	"logic-apps": buildResourceHijackingLogicAppsOutput,
}

func resourceHijackingHandler(provider providers.Provider, now func() time.Time) Handler {
	return groupedFamilyHandler(provider, now, resourceHijackingFamilyConfig())
}

func buildResourceHijackingOverview(now func() time.Time, request Request, selectedSurface *string) any {
	config := resourceHijackingFamilyConfig()
	return models.ResourceHijackingOverviewOutput{
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

func resourceHijackingFamilyConfig() groupedFamilyConfig {
	return groupedFamilyConfig{
		CommandName:            "resourcehijacking",
		CurrentBehavior:        resourceHijackingCurrentBehavior,
		CommandState:           resourceHijackingCommandState,
		InputModes:             resourceHijackingInputModes,
		PreferredArtifactOrder: resourceHijackingPreferredArtifactMode,
		Selector:               func(request Request) string { return request.ResourceHijackingSurface },
		Overview:               buildResourceHijackingOverview,
		SurfaceNames:           contracts.ResourceHijackingSurfaceNames,
		SurfaceContract:        contracts.ResourceHijackingSurface,
		SurfaceBuilders:        resourceHijackingSurfaceBuilders,
	}
}
