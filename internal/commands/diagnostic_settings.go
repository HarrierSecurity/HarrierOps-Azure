package commands

import (
	"context"
	"sort"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func diagnosticSettingsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.DiagnosticSettings(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}
		sources := append([]models.DiagnosticSettingsSource{}, facts.Sources...)
		sort.SliceStable(sources, func(i, j int) bool {
			if diagnosticSettingsCommandRank(sources[i]) != diagnosticSettingsCommandRank(sources[j]) {
				return diagnosticSettingsCommandRank(sources[i]) > diagnosticSettingsCommandRank(sources[j])
			}
			if sources[i].Type != sources[j].Type {
				return sources[i].Type < sources[j].Type
			}
			return sources[i].Name < sources[j].Name
		})
		return models.DiagnosticSettingsOutput{
			Sources:  sources,
			Findings: []models.Finding{},
			Issues:   facts.Issues,
			Metadata: withRuntimeArtifactContext(runtimeCommandMetadata("diagnostic-settings", now, facts.TenantID, facts.SubscriptionID), request, facts.CurrentPrincipal, facts.AuthMode, facts.TokenSource),
		}, nil
	}
}

func diagnosticSettingsCommandRank(source models.DiagnosticSettingsSource) int {
	rank := 0
	if source.HasNonWorkspaceDestination {
		rank += 2
	}
	if len(source.DisabledCategories) > 0 {
		rank += 2
	}
	if source.HasHighSignalLogPosture {
		rank += 2
	}
	if !source.HasDiagnosticSettings {
		rank++
	}
	return rank
}
