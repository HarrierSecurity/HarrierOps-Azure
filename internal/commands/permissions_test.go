package commands

import "testing"

func TestPermissionsScopeSummaryUsesExactSingleScopeLabel(t *testing.T) {
	got := permissionsSummary(
		"app-empty-mi-system",
		"ServicePrincipal",
		[]string{"Contributor"},
		[]string{"/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-apps"},
		1,
		true,
		false,
		true,
		false,
		false,
		"Check managed-identities for the workload pivot behind this direct control row.",
	)

	want := "ServicePrincipal 'app-empty-mi-system' already has direct control visible through Contributor at resource group rg-apps, and current scope also shows a workload pivot. Check managed-identities for the workload pivot behind this direct control row."
	if got != want {
		t.Fatalf("permissionsSummary() = %q, want %q", got, want)
	}
}

func TestPermissionsScopeSummaryUsesSubscriptionScopeForRootAssignment(t *testing.T) {
	got := permissionsSummary(
		"azurefox-lab-sp",
		"ServicePrincipal",
		[]string{"Owner"},
		[]string{"/subscriptions/22222222-2222-2222-2222-222222222222"},
		1,
		true,
		true,
		false,
		false,
		false,
		"Check privesc for the direct abuse or escalation path behind this current identity.",
	)

	want := "Current identity 'azurefox-lab-sp' already has direct control visible through Owner at subscription scope. Check privesc for the direct abuse or escalation path behind this current identity."
	if got != want {
		t.Fatalf("permissionsSummary() = %q, want %q", got, want)
	}
}
