package providers

import "testing"

func TestExistingGraphCredentialRowsDoNotRequireVisibleControlContext(t *testing.T) {
	rows := existingGraphCredentialRows(
		"Application",
		"app-1",
		"build-app",
		map[string]any{},
		map[string]any{
			"passwordCredentials": []any{map[string]any{"id": "pwd-1"}},
			"keyCredentials":      []any{map[string]any{"id": "key-1"}},
		},
		"No visible Azure RBAC is attached to this identity in the current subscription.",
	)

	if len(rows) != 2 {
		t.Fatalf("existingGraphCredentialRows() len = %d, want 2", len(rows))
	}
	if rows[0].RowClass != "existing_credential" || rows[1].RowClass != "existing_credential" {
		t.Fatalf("existingGraphCredentialRows() row classes = %q, %q, want existing_credential rows", rows[0].RowClass, rows[1].RowClass)
	}
	if got := valueOrEmptyStringPtr(rows[0].CredentialType); got != "password" {
		t.Fatalf("existingGraphCredentialRows() first credential_type = %q, want password", got)
	}
	if got := valueOrEmptyStringPtr(rows[1].CredentialType); got != "key" {
		t.Fatalf("existingGraphCredentialRows() second credential_type = %q, want key", got)
	}
}

func TestExistingFederatedTrustRowsDoNotRequireVisibleControlContext(t *testing.T) {
	rows := existingFederatedTrustRows(
		map[string]any{
			"id":          "app-1",
			"displayName": "build-app",
		},
		map[string]any{},
		[]map[string]any{{
			"id":      "fic-1",
			"issuer":  "https://token.actions.githubusercontent.com",
			"subject": "repo:harrierops/ho-azure:ref:refs/heads/main",
		}},
		"No visible Azure-facing service principal is linked to this application in the current environment.",
	)

	if len(rows) != 1 {
		t.Fatalf("existingFederatedTrustRows() len = %d, want 1", len(rows))
	}
	if rows[0].RowClass != "federated_trust_present" {
		t.Fatalf("existingFederatedTrustRows() row_class = %q, want federated_trust_present", rows[0].RowClass)
	}
	if got := valueOrEmptyStringPtr(rows[0].CredentialType); got != "federated" {
		t.Fatalf("existingFederatedTrustRows() credential_type = %q, want federated", got)
	}
}

func valueOrEmptyStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
