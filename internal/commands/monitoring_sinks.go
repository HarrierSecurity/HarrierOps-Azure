package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

type monitoringSinksSourceProvider interface {
	MonitoringSinksFromSources(context.Context, string, string, *providers.DCRFacts, *providers.DiagnosticSettingsFacts) (providers.MonitoringSinksFacts, error)
}

func monitoringSinksHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, sessionArtifacts, err := monitoringSinksFacts(ctx, request, provider, now)
		if err != nil {
			return nil, err
		}
		return models.MonitoringSinksOutput{
			Sinks:    sortedByLess(facts.Sinks, monitoringSinkLess),
			Findings: []models.Finding{},
			Issues:   facts.Issues,
			Metadata: withRuntimeSessionArtifacts(
				withRuntimeArtifactContext(runtimeCommandMetadata("monitoring-sinks", now, facts.TenantID, facts.SubscriptionID), request, facts.CurrentPrincipal, facts.AuthMode, facts.TokenSource),
				sessionArtifacts,
			),
		}, nil
	}
}

func monitoringSinksFacts(ctx context.Context, request Request, provider providers.Provider, now func() time.Time) (providers.MonitoringSinksFacts, []models.SessionArtifact, error) {
	sourceProvider, ok := provider.(monitoringSinksSourceProvider)
	if !ok {
		facts, err := provider.MonitoringSinks(ctx, request.Tenant, request.Subscription)
		return facts, nil, err
	}

	group := newCommandOutputGroup(2)
	expected := helperArtifactExpectedSessions(ctx, request, provider, now, "dcr", "diagnostic-settings")
	dcrFuture := runHelperOutput[models.DCROutput](group, ctx, request, dcrHandler(provider, now), "dcr", expected)
	diagnosticFuture := runHelperOutput[models.DiagnosticSettingsOutput](group, ctx, request, diagnosticSettingsHandler(provider, now), "diagnostic-settings", expected)

	dcr, dcrSource, err := dcrFuture.waitWithSource()
	if err != nil {
		return providers.MonitoringSinksFacts{}, nil, err
	}
	diagnosticSettings, diagnosticSource, err := diagnosticFuture.waitWithSource()
	if err != nil {
		return providers.MonitoringSinksFacts{}, nil, err
	}

	dcrFacts := dcrFactsFromOutput(dcr)
	diagnosticFacts := diagnosticSettingsFactsFromOutput(diagnosticSettings)
	facts, err := sourceProvider.MonitoringSinksFromSources(ctx, request.Tenant, request.Subscription, &dcrFacts, &diagnosticFacts)
	if err != nil {
		return providers.MonitoringSinksFacts{}, nil, err
	}

	sessionArtifacts := []models.SessionArtifact{}
	if dcrSource != nil {
		sessionArtifacts = append(sessionArtifacts, *dcrSource)
	}
	if diagnosticSource != nil {
		sessionArtifacts = append(sessionArtifacts, *diagnosticSource)
	}
	return facts, sessionArtifacts, nil
}

func dcrFactsFromOutput(output models.DCROutput) providers.DCRFacts {
	return providers.DCRFacts{
		ArtifactIdentityFacts: artifactIdentityFactsFromRuntimeMetadata(output.Metadata),
		TenantID:              stringPtrValue(output.Metadata.TenantID),
		SubscriptionID:        stringPtrValue(output.Metadata.SubscriptionID),
		DCRs:                  append([]models.DCRAsset{}, output.DCRs...),
		Issues:                append([]models.Issue{}, output.Issues...),
	}
}

func diagnosticSettingsFactsFromOutput(output models.DiagnosticSettingsOutput) providers.DiagnosticSettingsFacts {
	return providers.DiagnosticSettingsFacts{
		ArtifactIdentityFacts: artifactIdentityFactsFromRuntimeMetadata(output.Metadata),
		TenantID:              stringPtrValue(output.Metadata.TenantID),
		SubscriptionID:        stringPtrValue(output.Metadata.SubscriptionID),
		Sources:               append([]models.DiagnosticSettingsSource{}, output.Sources...),
		Issues:                append([]models.Issue{}, output.Issues...),
	}
}

func artifactIdentityFactsFromRuntimeMetadata(metadata models.RuntimeCommandMetadata) providers.ArtifactIdentityFacts {
	return artifactIdentityFactsFromContext(metadata.ArtifactContext, metadata.AuthMode, metadata.TokenSource)
}

func monitoringSinkLess(left models.MonitoringSinkAsset, right models.MonitoringSinkAsset) bool {
	if left.ReferenceCount != right.ReferenceCount {
		return left.ReferenceCount > right.ReferenceCount
	}
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	return left.Name < right.Name
}
