package artifacts

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
)

type testArtifactPayload struct {
	Metadata models.Metadata `json:"metadata"`
	Value    string          `json:"value"`
}

func TestLoadSessionArtifactMatchesStrictContext(t *testing.T) {
	now := time.Date(2026, time.April, 25, 18, 0, 0, 0, time.UTC)
	workspace := t.TempDir()
	writeTestArtifact(t, workspace, "rbac", now.Add(-38*time.Minute), "1111", "2222", "3333")

	result, ok, err := LoadSessionArtifact[testArtifactPayload](workspace, ExpectedSession{
		Command:        "rbac",
		SchemaVersion:  contracts.AzureFoxSchemaVersion,
		ToolVersion:    "dev",
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
	})
	if err != nil {
		t.Fatalf("load session artifact: %v", err)
	}
	if !ok {
		t.Fatalf("expected artifact reuse")
	}
	if result.Payload.Value != "from-artifact" {
		t.Fatalf("payload value = %q", result.Payload.Value)
	}
	if result.Source.Command != "rbac" || result.Source.AgeSeconds != int((38*time.Minute).Seconds()) {
		t.Fatalf("unexpected source: %#v", result.Source)
	}
}

func TestLoadSessionArtifactRejectsMismatchedContext(t *testing.T) {
	now := time.Date(2026, time.April, 25, 18, 0, 0, 0, time.UTC)
	workspace := t.TempDir()
	writeTestArtifact(t, workspace, "rbac", now.Add(-10*time.Minute), "1111", "2222", "3333")

	cases := []struct {
		name      string
		tenant    string
		sub       string
		principal string
		maxAge    time.Duration
		options   map[string]string
	}{
		{name: "tenant", tenant: "other", sub: "2222", principal: "3333", maxAge: time.Hour, options: map[string]string{}},
		{name: "subscription", tenant: "1111", sub: "other", principal: "3333", maxAge: time.Hour, options: map[string]string{}},
		{name: "principal", tenant: "1111", sub: "2222", principal: "other", maxAge: time.Hour, options: map[string]string{}},
		{name: "options", tenant: "1111", sub: "2222", principal: "3333", maxAge: time.Hour, options: map[string]string{"mode": "full"}},
		{name: "freshness", tenant: "1111", sub: "2222", principal: "3333", maxAge: time.Minute, options: map[string]string{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok, err := LoadSessionArtifact[testArtifactPayload](workspace, ExpectedSession{
				Command:        "rbac",
				SchemaVersion:  contracts.AzureFoxSchemaVersion,
				ToolVersion:    "dev",
				TenantID:       tc.tenant,
				SubscriptionID: tc.sub,
				CurrentPrincipal: models.ArtifactPrincipal{
					ID:       tc.principal,
					TenantID: tc.tenant,
				},
				AuthMode:       "fixture",
				TokenSource:    "fixture",
				CommandOptions: tc.options,
				MaxAge:         tc.maxAge,
				Now:            now,
			})
			if err != nil {
				t.Fatalf("load session artifact: %v", err)
			}
			if ok {
				t.Fatalf("expected mismatched artifact to be rejected")
			}
		})
	}
}

func TestLoadSessionArtifactFallsBackFromMalformedJSONToValidLoot(t *testing.T) {
	now := time.Date(2026, time.April, 25, 18, 0, 0, 0, time.UTC)
	workspace := t.TempDir()
	writeRawTestArtifact(t, workspace, "json", "rbac", []byte(`{`))
	writeTestArtifactLane(t, workspace, "loot", "rbac", now.Add(-10*time.Minute), "1111", "2222", "3333")

	result, ok, err := LoadSessionArtifact[testArtifactPayload](workspace, ExpectedSession{
		Command:        "rbac",
		SchemaVersion:  contracts.AzureFoxSchemaVersion,
		ToolVersion:    "dev",
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
	})
	if err != nil {
		t.Fatalf("load session artifact: %v", err)
	}
	if !ok {
		t.Fatalf("expected valid loot artifact to be reused")
	}
	if result.Source.Path != filepath.Join(workspace, "loot", "rbac.json") {
		t.Fatalf("expected loot source path, got %q", result.Source.Path)
	}
}

func TestLoadSessionArtifactIgnoresMalformedArtifactWhenNoValidCandidateExists(t *testing.T) {
	now := time.Date(2026, time.April, 25, 18, 0, 0, 0, time.UTC)
	workspace := t.TempDir()
	writeRawTestArtifact(t, workspace, "json", "rbac", []byte(`{`))

	_, ok, err := LoadSessionArtifact[testArtifactPayload](workspace, ExpectedSession{
		Command:        "rbac",
		SchemaVersion:  contracts.AzureFoxSchemaVersion,
		ToolVersion:    "dev",
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
	})
	if err != nil {
		t.Fatalf("malformed candidate should fall through to live refresh, got error: %v", err)
	}
	if ok {
		t.Fatalf("expected malformed artifact not to be reused")
	}
}

func TestLoadSessionAnchorUsesFreshWhoamiArtifact(t *testing.T) {
	now := time.Date(2026, time.April, 25, 18, 0, 0, 0, time.UTC)
	workspace := t.TempDir()
	writeTestArtifact(t, workspace, "whoami", now.Add(-29*time.Minute), "1111", "2222", "3333")

	anchor, ok, err := LoadSessionAnchor(workspace, contracts.AzureFoxSchemaVersion, "dev", 30*time.Minute, now)
	if err != nil {
		t.Fatalf("load session anchor: %v", err)
	}
	if !ok {
		t.Fatalf("expected fresh whoami anchor")
	}
	if anchor.TenantID != "1111" || anchor.SubscriptionID != "2222" || anchor.CurrentPrincipal.ID != "3333" {
		t.Fatalf("unexpected anchor: %#v", anchor)
	}
}

func TestLoadSessionAnchorRejectsStaleWhoamiArtifact(t *testing.T) {
	now := time.Date(2026, time.April, 25, 18, 0, 0, 0, time.UTC)
	workspace := t.TempDir()
	writeTestArtifact(t, workspace, "whoami", now.Add(-31*time.Minute), "1111", "2222", "3333")

	_, ok, err := LoadSessionAnchor(workspace, contracts.AzureFoxSchemaVersion, "dev", 30*time.Minute, now)
	if err != nil {
		t.Fatalf("load session anchor: %v", err)
	}
	if ok {
		t.Fatalf("expected stale whoami anchor to be rejected")
	}
}

func TestLoadSessionAnchorCanUseFreshHelperArtifact(t *testing.T) {
	now := time.Date(2026, time.April, 25, 18, 0, 0, 0, time.UTC)
	workspace := t.TempDir()
	writeTestArtifact(t, workspace, "permissions", now.Add(-12*time.Minute), "1111", "2222", "3333")

	anchor, ok, err := LoadSessionAnchorFromCommands(workspace, []string{"whoami", "permissions"}, contracts.AzureFoxSchemaVersion, "dev", 30*time.Minute, now)
	if err != nil {
		t.Fatalf("load session anchor: %v", err)
	}
	if !ok {
		t.Fatalf("expected fresh helper artifact anchor")
	}
	if anchor.TenantID != "1111" || anchor.SubscriptionID != "2222" || anchor.CurrentPrincipal.ID != "3333" {
		t.Fatalf("unexpected anchor: %#v", anchor)
	}
}

func writeTestArtifact(t *testing.T, workspace string, command string, generatedAt time.Time, tenant string, subscription string, principal string) {
	t.Helper()
	writeTestArtifactLane(t, workspace, "json", command, generatedAt, tenant, subscription, principal)
}

func writeTestArtifactLane(t *testing.T, workspace string, lane string, command string, generatedAt time.Time, tenant string, subscription string, principal string) {
	t.Helper()
	payload := testArtifactPayload{
		Metadata: models.Metadata{
			AuthMode:       models.StringPtr("fixture"),
			Command:        command,
			GeneratedAt:    generatedAt.Format(time.RFC3339),
			SchemaVersion:  contracts.AzureFoxSchemaVersion,
			SubscriptionID: models.StringPtr(subscription),
			TenantID:       models.StringPtr(tenant),
			TokenSource:    models.StringPtr("fixture"),
			ArtifactContext: &models.ArtifactContext{
				ToolVersion: "dev",
				CurrentPrincipal: models.ArtifactPrincipal{
					ID:       principal,
					TenantID: tenant,
				},
				CommandOptions: map[string]string{},
			},
		},
		Value: "from-artifact",
	}
	content := `{
  "metadata": {
    "auth_mode": "` + *payload.Metadata.AuthMode + `",
    "command": "` + payload.Metadata.Command + `",
    "generated_at": "` + payload.Metadata.GeneratedAt + `",
    "schema_version": "` + payload.Metadata.SchemaVersion + `",
    "subscription_id": "` + *payload.Metadata.SubscriptionID + `",
    "tenant_id": "` + *payload.Metadata.TenantID + `",
    "token_source": "` + *payload.Metadata.TokenSource + `",
    "artifact_context": {
      "tool_version": "dev",
      "current_principal": {
        "id": "` + payload.Metadata.ArtifactContext.CurrentPrincipal.ID + `",
        "tenant_id": "` + payload.Metadata.ArtifactContext.CurrentPrincipal.TenantID + `"
      },
      "command_options": {}
    }
  },
  "value": "from-artifact"
}
`
	path := filepath.Join(workspace, lane, command+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir artifact dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
}

func writeRawTestArtifact(t *testing.T, workspace string, lane string, command string, content []byte) {
	t.Helper()
	path := filepath.Join(workspace, lane, command+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir artifact dir: %v", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write raw artifact: %v", err)
	}
}
