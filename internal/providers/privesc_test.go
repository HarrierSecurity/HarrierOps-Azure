package providers

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestPrivescMarksPreferredPathAndExplainsWhyItWon(t *testing.T) {
	servicePrincipalID := "sp-current"
	managedPrincipalID := "mi-nearby"
	vmID := "/subscriptions/sub-1/resourceGroups/rg-1/providers/Microsoft.Compute/virtualMachines/vm-public"

	permissions := PermissionsFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		Permissions: []PermissionFact{
			{
				DisplayName:       "svc-current",
				PrincipalID:       servicePrincipalID,
				PrincipalType:     "ServicePrincipal",
				HighImpactRoles:   []string{"Owner"},
				Privileged:        true,
				IsCurrentIdentity: true,
			},
			{
				DisplayName:       "mi-nearby",
				PrincipalID:       managedPrincipalID,
				PrincipalType:     "ManagedIdentity",
				HighImpactRoles:   []string{"Owner"},
				Privileged:        true,
				IsCurrentIdentity: false,
			},
		},
	}

	principals := PrincipalsFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		Principals: []models.PrincipalSummary{
			{
				ID:                servicePrincipalID,
				DisplayName:       models.StringPtr("svc-current"),
				PrincipalType:     "ServicePrincipal",
				IsCurrentIdentity: true,
			},
			{
				ID:                managedPrincipalID,
				DisplayName:       models.StringPtr("mi-nearby"),
				PrincipalType:     "ManagedIdentity",
				IsCurrentIdentity: false,
			},
		},
	}

	managedIdentities := ManagedIdentitiesFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		Identities: []models.ManagedIdentity{
			{
				ID:          "mi-1",
				Name:        "mi-nearby",
				PrincipalID: models.StringPtr(managedPrincipalID),
				AttachedTo:  []string{vmID},
			},
		},
	}

	vms := VMsFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		VMAssets: []models.VmAsset{
			{
				ID:        vmID,
				Name:      "vm-public",
				PublicIPs: []string{"52.160.10.20"},
			},
		},
	}

	facts := PrivescFactsFromSources(permissions, principals, managedIdentities, vms)
	if len(facts.Paths) < 2 {
		t.Fatalf("expected at least two privesc paths, got %d", len(facts.Paths))
	}

	if !facts.Paths[0].Preferred {
		t.Fatalf("expected first privesc path to be marked preferred")
	}
	if got := facts.Paths[0].PreferredReason; got == "" {
		t.Fatalf("expected preferred privesc path to explain why it won")
	}
	if facts.Paths[1].Preferred {
		t.Fatalf("expected only the winning privesc path to be marked preferred")
	}
}

func TestPrivescPrefersVisiblePrivilegedLeadBeforeIngressPivotWhenPrivilegeSimilar(t *testing.T) {
	servicePrincipalID := "sp-direct"
	managedPrincipalID := "sp-managed"
	vmID := "/subscriptions/sub-1/resourceGroups/rg-1/providers/Microsoft.Compute/virtualMachines/vm-public"

	permissions := PermissionsFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		Permissions: []PermissionFact{
			{
				DisplayName:       "svc-direct",
				PrincipalID:       servicePrincipalID,
				PrincipalType:     "ServicePrincipal",
				HighImpactRoles:   []string{"Owner"},
				Privileged:        true,
				IsCurrentIdentity: false,
			},
			{
				DisplayName:       "sp-managed",
				PrincipalID:       managedPrincipalID,
				PrincipalType:     "ServicePrincipal",
				HighImpactRoles:   []string{"Owner"},
				Privileged:        true,
				IsCurrentIdentity: false,
			},
		},
	}

	principals := PrincipalsFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		Principals: []models.PrincipalSummary{
			{
				ID:                servicePrincipalID,
				DisplayName:       models.StringPtr("svc-direct"),
				PrincipalType:     "ServicePrincipal",
				IsCurrentIdentity: false,
			},
			{
				ID:                managedPrincipalID,
				DisplayName:       models.StringPtr("sp-managed"),
				PrincipalType:     "ServicePrincipal",
				IsCurrentIdentity: false,
			},
		},
	}

	managedIdentities := ManagedIdentitiesFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		Identities: []models.ManagedIdentity{
			{
				ID:          "mi-1",
				Name:        "ua-pivot",
				PrincipalID: models.StringPtr(managedPrincipalID),
				AttachedTo:  []string{vmID},
			},
		},
	}

	vms := VMsFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		VMAssets: []models.VmAsset{
			{
				ID:        vmID,
				Name:      "vm-public",
				PublicIPs: []string{"52.160.10.20"},
			},
		},
	}

	facts := PrivescFactsFromSources(permissions, principals, managedIdentities, vms)
	if len(facts.Paths) < 3 {
		t.Fatalf("expected at least three privesc paths, got %d", len(facts.Paths))
	}

	if got := facts.Paths[0].PathType; got != privescVisiblePrivilegedLead {
		t.Fatalf("expected a visible privileged lead to sort ahead of ingress-backed workload identity when privilege is similar, got first path type %q", got)
	}
}

func TestPrivescPrefersHigherPrivilegeThenNonHumanIdentity(t *testing.T) {
	managedPrincipalID := "sp-managed"
	userPrincipalID := "user-high"

	permissions := PermissionsFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		Permissions: []PermissionFact{
			{
				DisplayName:       "aaa-user",
				PrincipalID:       userPrincipalID,
				PrincipalType:     "User",
				HighImpactRoles:   []string{"Owner"},
				Privileged:        true,
				IsCurrentIdentity: false,
			},
			{
				DisplayName:       "zzz-automation-sp",
				PrincipalID:       managedPrincipalID,
				PrincipalType:     "ServicePrincipal",
				HighImpactRoles:   []string{"Owner"},
				Privileged:        true,
				IsCurrentIdentity: false,
			},
		},
	}

	principals := PrincipalsFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		Principals: []models.PrincipalSummary{
			{
				ID:                userPrincipalID,
				DisplayName:       models.StringPtr("aaa-user"),
				PrincipalType:     "User",
				IsCurrentIdentity: false,
			},
			{
				ID:                managedPrincipalID,
				DisplayName:       models.StringPtr("zzz-automation-sp"),
				PrincipalType:     "ServicePrincipal",
				IsCurrentIdentity: false,
			},
		},
	}

	facts := PrivescFactsFromSources(permissions, principals, ManagedIdentitiesFacts{}, VMsFacts{})
	if len(facts.Paths) < 2 {
		t.Fatalf("expected at least two privesc paths, got %d", len(facts.Paths))
	}

	if got := facts.Paths[0].Principal; got != "zzz-automation-sp" {
		t.Fatalf("expected managed/non-human identity to sort ahead of user on same privilege, got first principal %q", got)
	}
	if got := facts.Paths[1].Principal; got != "aaa-user" {
		t.Fatalf("expected user to sort after managed/non-human identity on same privilege, got second principal %q", got)
	}
}

func TestPrivescPrefersAutomationThemedIdentityWhenPrivilegeAndTypeAreSimilar(t *testing.T) {
	firstID := "sp-1"
	secondID := "sp-2"

	permissions := PermissionsFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		Permissions: []PermissionFact{
			{
				DisplayName:       "zzz-generic-sp",
				PrincipalID:       firstID,
				PrincipalType:     "ServicePrincipal",
				HighImpactRoles:   []string{"Owner"},
				Privileged:        true,
				IsCurrentIdentity: false,
			},
			{
				DisplayName:       "aaa-pipeline-sp",
				PrincipalID:       secondID,
				PrincipalType:     "ServicePrincipal",
				HighImpactRoles:   []string{"Owner"},
				Privileged:        true,
				IsCurrentIdentity: false,
			},
		},
	}

	principals := PrincipalsFacts{
		TenantID:       "tenant-1",
		SubscriptionID: "sub-1",
		Principals: []models.PrincipalSummary{
			{
				ID:                firstID,
				DisplayName:       models.StringPtr("zzz-generic-sp"),
				PrincipalType:     "ServicePrincipal",
				IsCurrentIdentity: false,
			},
			{
				ID:                secondID,
				DisplayName:       models.StringPtr("aaa-pipeline-sp"),
				PrincipalType:     "ServicePrincipal",
				IsCurrentIdentity: false,
			},
		},
	}

	facts := PrivescFactsFromSources(permissions, principals, ManagedIdentitiesFacts{}, VMsFacts{})
	if len(facts.Paths) < 2 {
		t.Fatalf("expected at least two privesc paths, got %d", len(facts.Paths))
	}

	if got := facts.Paths[0].Principal; got != "aaa-pipeline-sp" {
		t.Fatalf("expected automation/pipeline-themed identity to sort ahead of generic service principal when privilege is otherwise similar, got first principal %q", got)
	}
	if got := facts.Paths[0].PreferredReason; !strings.Contains(got, "pipeline-themed") {
		t.Fatalf("expected preferred reason to surface the pipeline theme tie-break, got %q", got)
	}
}
