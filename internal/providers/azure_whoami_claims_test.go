package providers

import "testing"

func TestPrincipalTypeFromClaims(t *testing.T) {
	testCases := []struct {
		name   string
		claims map[string]string
		want   string
	}{
		{
			name: "delegated user token stays user even when appid is present",
			claims: map[string]string{
				"appid":       "04b07795-8ddb-461a-bbee-02f9e1bf7b46",
				"idtyp":       "user",
				"oid":         "1058bd62-c9bd-4332-b6c4-bcf3f90f1c4e",
				"scp":         "user_impersonation",
				"unique_name": "live.com#farleycolby@gmail.com",
			},
			want: "User",
		},
		{
			name: "managed identity signal wins when xms_mirid is present",
			claims: map[string]string{
				"appid":     "11111111-2222-3333-4444-555555555555",
				"idtyp":     "app",
				"oid":       "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
				"xms_mirid": "/subscriptions/sub-1/resourcegroups/rg-1/providers/Microsoft.ManagedIdentity/userAssignedIdentities/mi-runner",
			},
			want: "ManagedIdentity",
		},
		{
			name: "app-only token without managed identity stays service principal",
			claims: map[string]string{
				"appid": "11111111-2222-3333-4444-555555555555",
				"idtyp": "app",
				"oid":   "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			},
			want: "ServicePrincipal",
		},
		{
			name: "oid without appid falls back to user",
			claims: map[string]string{
				"oid": "1058bd62-c9bd-4332-b6c4-bcf3f90f1c4e",
			},
			want: "User",
		},
		{
			name:   "empty claims stay unknown",
			claims: map[string]string{},
			want:   "Unknown",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := principalTypeFromClaims(testCase.claims)
			if got != testCase.want {
				t.Fatalf("principalTypeFromClaims() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestCurrentPrincipalFromClaimsUsesObjectIDNotAppID(t *testing.T) {
	principalID, displayName := currentPrincipalFromClaims(map[string]string{
		"appid": "11111111-2222-3333-4444-555555555555",
		"oid":   "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		"sub":   "subject-value",
		"name":  "runner-sp",
	})

	if principalID != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Fatalf("expected oid-backed principal ID, got %q", principalID)
	}
	if displayName != "runner-sp" {
		t.Fatalf("expected explicit display name, got %q", displayName)
	}
}

func TestCurrentPrincipalFromClaimsFallsBackToSubjectBeforeAppID(t *testing.T) {
	principalID, displayName := currentPrincipalFromClaims(map[string]string{
		"appid": "11111111-2222-3333-4444-555555555555",
		"sub":   "subject-value",
	})

	if principalID != "subject-value" {
		t.Fatalf("expected subject-backed fallback principal ID, got %q", principalID)
	}
	if displayName != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("expected appid to remain only a name fallback, got %q", displayName)
	}
}
