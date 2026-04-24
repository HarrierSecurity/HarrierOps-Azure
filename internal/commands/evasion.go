package commands

import (
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

const (
	evasionCurrentBehavior = "Grouped evasion walkthroughs. Use `ho-azure evasion` or `ho-azure evasion help` to list surfaces, then `ho-azure evasion <surface>` to run an implemented surface."
	evasionCommandState    = contracts.StatusImplemented
)

var (
	evasionInputModes            = []string{"live"}
	evasionPreferredArtifactMode = []string{"loot", "json"}
)

var evasionSurfaceBuilders = map[string]groupedSurfaceBuilder{
	"appinsights":         buildEvasionAppInsightsOutput,
	"dcr":                 buildEvasionDCROutput,
	"diagnostic-settings": buildEvasionDiagnosticSettingsOutput,
}

func evasionHandler(provider providers.Provider, now func() time.Time) Handler {
	return groupedFamilyHandler(provider, now, evasionFamilyConfig())
}

func buildEvasionOverview(now func() time.Time, request Request, selectedSurface *string) any {
	config := evasionFamilyConfig()
	return models.EvasionOverviewOutput{
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

func evasionFamilyConfig() groupedFamilyConfig {
	return groupedFamilyConfig{
		CommandName:            "evasion",
		CurrentBehavior:        evasionCurrentBehavior,
		CommandState:           evasionCommandState,
		InputModes:             evasionInputModes,
		PreferredArtifactOrder: evasionPreferredArtifactMode,
		Selector:               func(request Request) string { return request.EvasionSurface },
		Overview:               buildEvasionOverview,
		SurfaceNames:           contracts.EvasionSurfaceNames,
		SurfaceContract:        contracts.EvasionSurface,
		SurfaceBuilders:        evasionSurfaceBuilders,
	}
}
