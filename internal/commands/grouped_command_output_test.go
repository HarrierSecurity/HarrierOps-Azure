package commands

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"harrierops-azure/internal/artifacts"
	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func TestRunGroupedCommandOutputReusesCommandResult(t *testing.T) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	calls := 0
	handler := func(context.Context, Request) (any, error) {
		calls++
		return "shared-output", nil
	}

	first := runGroupedCommandOutput[string](group, context.Background(), Request{}, handler, "shared")
	second := runGroupedCommandOutput[string](group, context.Background(), Request{}, handler, "shared")

	firstValue, err := first.wait()
	if err != nil {
		t.Fatalf("first wait failed: %v", err)
	}
	secondValue, err := second.wait()
	if err != nil {
		t.Fatalf("second wait failed: %v", err)
	}
	if firstValue != "shared-output" || secondValue != "shared-output" {
		t.Fatalf("unexpected values: first=%q second=%q", firstValue, secondValue)
	}
	if calls != 1 {
		t.Fatalf("expected one backing call, got %d", calls)
	}
}

func TestRunGroupedCommandOutputUsesMatchingSessionArtifact(t *testing.T) {
	now := time.Date(2026, time.April, 25, 18, 0, 0, 0, time.UTC)
	workspace := t.TempDir()
	writeRbacArtifact(t, workspace, now.Add(-20*time.Minute), "1111", "2222", "3333")

	group := newCommandOutputGroup(chainsFanoutLimit)
	calls := 0
	handler := func(context.Context, Request) (any, error) {
		calls++
		return models.RbacOutput{}, nil
	}

	future := runGroupedCommandOutputWithArtifact[models.RbacOutput](
		group,
		context.Background(),
		Request{OutDir: workspace},
		handler,
		artifacts.ExpectedSession{
			Command:        "rbac",
			SchemaVersion:  contracts.AzureFoxSchemaVersion,
			ToolVersion:    toolVersion,
			TenantID:       "1111",
			SubscriptionID: "2222",
			CurrentPrincipal: models.ArtifactPrincipal{
				ID:       "3333",
				TenantID: "1111",
			},
			AuthMode:       "fixture",
			TokenSource:    "fixture",
			CommandOptions: map[string]string{},
			MaxAge:         time.Hour,
			Now:            now,
		},
	)

	output, source, err := future.waitWithSource()
	if err != nil {
		t.Fatalf("wait failed: %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected artifact reuse to suppress live handler call, got %d calls", calls)
	}
	if source == nil || source.Command != "rbac" {
		t.Fatalf("expected source artifact, got %#v", source)
	}
	if output.Metadata.Command != "rbac" {
		t.Fatalf("expected rbac payload, got %#v", output.Metadata)
	}
}

func TestHelperArtifactExpectedSessionsIgnoresFinalGroupedOutputArtifacts(t *testing.T) {
	workspace := t.TempDir()
	path := filepath.Join(workspace, "json", "resourcehijacking.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir grouped artifact dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"metadata":{"command":"resourcehijacking"}}`), 0o644); err != nil {
		t.Fatalf("write grouped artifact: %v", err)
	}

	expected := helperArtifactExpectedSessions(
		context.Background(),
		Request{OutDir: workspace},
		providers.NewStaticProvider(),
		func() time.Time { return time.Date(2026, time.April, 25, 18, 0, 0, 0, time.UTC) },
		"permissions",
		"rbac",
	)
	if expected != nil {
		t.Fatalf("expected final grouped output artifact to be ignored as a helper reuse source, got %#v", expected)
	}
}

func TestHelperArtifactExpectedSessionsUsesCurrentDirectoryWhenOutdirEmpty(t *testing.T) {
	now := time.Date(2026, time.April, 25, 18, 0, 0, 0, time.UTC)
	workspace := t.TempDir()
	writeRbacArtifact(t, workspace, now.Add(-10*time.Minute), "1111", "2222", "3333")

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("chdir workspace: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	expected := helperArtifactExpectedSessions(
		context.Background(),
		Request{},
		providers.NewStaticProvider(),
		func() time.Time { return now },
		"rbac",
	)
	if expected == nil {
		t.Fatalf("expected current-directory artifact to create expected session")
	}
	if expected["rbac"].SubscriptionID != "2222" || expected["rbac"].CurrentPrincipal.ID != "3333" {
		t.Fatalf("unexpected expected session: %#v", expected["rbac"])
	}
}

func TestRunGroupedCommandOutputWritingArtifactReturnsWriteError(t *testing.T) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	handler := func(context.Context, Request) (any, error) {
		return models.RbacOutput{}, nil
	}

	future := runGroupedCommandOutputWritingArtifact[models.RbacOutput](
		group,
		context.Background(),
		Request{OutDir: string([]byte{0})},
		handler,
		"rbac",
	)
	_, err := future.wait()
	if err == nil {
		t.Fatalf("expected artifact write error")
	}
}

func writeRbacArtifact(t *testing.T, workspace string, generatedAt time.Time, tenant string, subscription string, principal string) {
	t.Helper()
	output := models.RbacOutput{
		Metadata: models.Metadata{
			AuthMode:       models.StringPtr("fixture"),
			Command:        "rbac",
			GeneratedAt:    generatedAt.Format(time.RFC3339),
			SchemaVersion:  contracts.AzureFoxSchemaVersion,
			SubscriptionID: models.StringPtr(subscription),
			TenantID:       models.StringPtr(tenant),
			TokenSource:    models.StringPtr("fixture"),
			ArtifactContext: &models.ArtifactContext{
				ToolVersion: toolVersion,
				CurrentPrincipal: models.ArtifactPrincipal{
					ID:       principal,
					TenantID: tenant,
				},
				CommandOptions: map[string]string{},
			},
		},
		Issues:          []models.Issue{},
		Principals:      []models.Principal{},
		RoleAssignments: []models.RoleAssignment{},
		Scopes:          []models.ScopeRef{},
	}
	content, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("marshal artifact: %v", err)
	}
	path := filepath.Join(workspace, "json", "rbac.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir artifact dir: %v", err)
	}
	if err := os.WriteFile(path, append(content, '\n'), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
}
