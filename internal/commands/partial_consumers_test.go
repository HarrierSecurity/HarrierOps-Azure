package commands

import (
	"context"
	"strings"
	"testing"
	"time"

	"harrierops-azure/internal/artifacts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func TestPartialConsumerMonitoringSinksUsesSourceArtifacts(t *testing.T) {
	workspace := t.TempDir()
	now := fixedArtifactTestTime()
	provider := providers.NewStaticProvider()
	writeCommandArtifact(t, workspace, "dcr", dcrHandler(provider, now))
	writeCommandArtifact(t, workspace, "diagnostic-settings", diagnosticSettingsHandler(provider, now))

	payload, err := monitoringSinksHandler(blockingMonitoringSinksProvider{StaticProvider: provider}, now)(context.Background(), Request{OutDir: workspace})
	if err != nil {
		t.Fatalf("monitoring-sinks handler: %v", err)
	}
	output := payload.(models.MonitoringSinksOutput)
	assertSessionArtifactCommands(t, output.Metadata.SessionArtifacts, "dcr", "diagnostic-settings")
}

func TestPartialConsumerResourceTrustsUsesSourceArtifacts(t *testing.T) {
	workspace := t.TempDir()
	now := fixedArtifactTestTime()
	provider := providers.NewStaticProvider()
	writeCommandArtifact(t, workspace, "storage", storageHandler(provider, now))
	writeCommandArtifact(t, workspace, "keyvault", keyVaultHandler(provider, now))

	payload, err := resourceTrustsHandler(blockingResourceTrustsProvider{StaticProvider: provider}, now)(context.Background(), Request{OutDir: workspace})
	if err != nil {
		t.Fatalf("resource-trusts handler: %v", err)
	}
	output := payload.(models.ResourceTrustsOutput)
	assertSessionArtifactCommands(t, output.Metadata.SessionArtifacts, "storage", "keyvault")
}

func TestPartialConsumerIdentityCommandsUseSourceArtifacts(t *testing.T) {
	workspace := t.TempDir()
	now := fixedArtifactTestTime()
	provider := providers.NewStaticProvider()
	writeCommandArtifact(t, workspace, "rbac", rbacHandler(provider, now))
	writeCommandArtifact(t, workspace, "whoami", whoAmIHandler(provider, now))
	writeCommandArtifact(t, workspace, "managed-identities", managedIdentitiesHandler(provider, now))

	request := Request{OutDir: workspace}
	principalsPayload, err := principalsHandler(blockingIdentityProvider{StaticProvider: provider}, now)(context.Background(), request)
	if err != nil {
		t.Fatalf("principals handler: %v", err)
	}
	principalsOutput := principalsPayload.(models.PrincipalsOutput)
	assertSessionArtifactCommands(t, principalsOutput.Metadata.SessionArtifacts, "rbac", "whoami", "managed-identities")

	permissionsPayload, err := permissionsHandler(blockingIdentityProvider{StaticProvider: provider}, now)(context.Background(), request)
	if err != nil {
		t.Fatalf("permissions handler: %v", err)
	}
	permissionsOutput := permissionsPayload.(models.PermissionsOutput)
	assertSessionArtifactCommands(t, permissionsOutput.Metadata.SessionArtifacts, "rbac", "whoami", "managed-identities")
}

func TestPartialConsumerPrivescUsesSourceArtifacts(t *testing.T) {
	workspace := t.TempDir()
	now := fixedArtifactTestTime()
	provider := providers.NewStaticProvider()
	writeCommandArtifact(t, workspace, "permissions", permissionsHandler(provider, now))
	writeCommandArtifact(t, workspace, "principals", principalsHandler(provider, now))
	writeCommandArtifact(t, workspace, "managed-identities", managedIdentitiesHandler(provider, now))
	writeCommandArtifact(t, workspace, "vms", vmsHandler(provider, now))

	payload, err := privescHandler(blockingPrivescProvider{StaticProvider: provider}, now)(context.Background(), Request{OutDir: workspace})
	if err != nil {
		t.Fatalf("privesc handler: %v", err)
	}
	output := payload.(models.PrivescOutput)
	assertSessionArtifactCommands(t, output.Metadata.SessionArtifacts, "permissions", "principals", "managed-identities", "vms")
}

func TestPartialConsumerRejectsMismatchedSourceArtifact(t *testing.T) {
	workspace := t.TempDir()
	now := fixedArtifactTestTime()
	provider := providers.NewStaticProvider()
	writeCommandArtifact(t, workspace, "dcr", dcrHandler(provider, now))
	writeCommandArtifactWithRequest(t, workspace, "diagnostic-settings", diagnosticSettingsHandler(provider, now), Request{
		OutDir:       workspace,
		Subscription: "different-subscription",
	})

	payload, err := monitoringSinksHandler(blockingMonitoringSinksProvider{StaticProvider: provider}, now)(context.Background(), Request{OutDir: workspace})
	if err != nil {
		t.Fatalf("monitoring-sinks handler: %v", err)
	}
	output := payload.(models.MonitoringSinksOutput)
	assertSessionArtifactCommands(t, output.Metadata.SessionArtifacts, "dcr")
}

type blockingMonitoringSinksProvider struct {
	providers.StaticProvider
}

func (provider blockingMonitoringSinksProvider) MonitoringSinks(_ context.Context, _ string, _ string) (providers.MonitoringSinksFacts, error) {
	panic("monitoring-sinks should compose from source artifacts")
}

func (provider blockingMonitoringSinksProvider) MonitoringSinksFromSources(ctx context.Context, tenant string, subscription string, dcrFacts *providers.DCRFacts, diagnosticFacts *providers.DiagnosticSettingsFacts) (providers.MonitoringSinksFacts, error) {
	if dcrFacts == nil || diagnosticFacts == nil {
		panic("monitoring-sinks missing source facts")
	}
	return provider.StaticProvider.MonitoringSinksFromSources(ctx, tenant, subscription, dcrFacts, diagnosticFacts)
}

type blockingResourceTrustsProvider struct {
	providers.StaticProvider
}

func (provider blockingResourceTrustsProvider) ResourceTrusts(_ context.Context, _ string, _ string) (providers.ResourceTrustsFacts, error) {
	panic("resource-trusts should compose from source artifacts")
}

type blockingIdentityProvider struct {
	providers.StaticProvider
}

func (provider blockingIdentityProvider) Principals(_ context.Context, _ string, _ string) (providers.PrincipalsFacts, error) {
	panic("principals should compose from source artifacts")
}

func (provider blockingIdentityProvider) PrincipalsFromSources(ctx context.Context, tenant string, subscription string, rbacFacts providers.RBACFacts, whoamiFacts providers.WhoAmIFacts, managedIdentityFacts providers.ManagedIdentitiesFacts) (providers.PrincipalsFacts, error) {
	return providers.PrincipalsFactsFromSources(tenant, subscription, rbacFacts, whoamiFacts, managedIdentityFacts), nil
}

func (provider blockingIdentityProvider) Permissions(_ context.Context, _ string, _ string) (providers.PermissionsFacts, error) {
	panic("permissions should compose from source artifacts")
}

func (provider blockingIdentityProvider) PermissionsFromSources(ctx context.Context, tenant string, subscription string, rbacFacts providers.RBACFacts, whoamiFacts providers.WhoAmIFacts, managedIdentityFacts providers.ManagedIdentitiesFacts) (providers.PermissionsFacts, error) {
	return providers.PermissionsFactsFromSources(tenant, subscription, rbacFacts, whoamiFacts, managedIdentityFacts), nil
}

type blockingPrivescProvider struct {
	providers.StaticProvider
}

func (provider blockingPrivescProvider) Privesc(_ context.Context, _ string, _ string) (providers.PrivescFacts, error) {
	panic("privesc should compose from source artifacts")
}

func (provider blockingPrivescProvider) PrivescFromSources(ctx context.Context, permissionsFacts providers.PermissionsFacts, principalsFacts providers.PrincipalsFacts, managedIdentityFacts providers.ManagedIdentitiesFacts, vmFacts providers.VMsFacts) (providers.PrivescFacts, error) {
	return providers.PrivescFactsFromSources(permissionsFacts, principalsFacts, managedIdentityFacts, vmFacts), nil
}

func fixedArtifactTestTime() func() time.Time {
	return func() time.Time {
		return time.Date(2026, time.April, 25, 18, 0, 0, 0, time.UTC)
	}
}

func writeCommandArtifact(t *testing.T, workspace string, command string, handler Handler) {
	t.Helper()
	writeCommandArtifactWithRequest(t, workspace, command, handler, Request{OutDir: workspace})
}

func writeCommandArtifactWithRequest(t *testing.T, workspace string, command string, handler Handler, request Request) {
	t.Helper()
	payload, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("%s handler: %v", command, err)
	}
	if _, err := artifacts.Write(command, payload, workspace, models.RenderContext{
		Tenant:       request.Tenant,
		Subscription: request.Subscription,
	}); err != nil {
		t.Fatalf("write %s artifact: %v", command, err)
	}
}

func assertSessionArtifactCommands(t *testing.T, artifacts []models.SessionArtifact, expected ...string) {
	t.Helper()
	if len(artifacts) != len(expected) {
		t.Fatalf("expected %d source artifact(s), got %d: %#v", len(expected), len(artifacts), artifacts)
	}
	seen := map[string]bool{}
	for _, artifact := range artifacts {
		seen[artifact.Command] = true
		if !strings.Contains(artifact.Context, "same tenant") {
			t.Fatalf("unexpected artifact context for %s: %q", artifact.Command, artifact.Context)
		}
	}
	for _, command := range expected {
		if !seen[command] {
			t.Fatalf("missing source artifact %q in %#v", command, artifacts)
		}
	}
}
