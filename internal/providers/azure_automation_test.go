package providers

import "testing"

func TestAutomationIdentityHelpersPreserveOnlyRealIdentityIDs(t *testing.T) {
	accountID := "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Automation/automationAccounts/aa-ops"
	principalID := "25735885-4d19-475f-a313-c643041061e0"

	identity := map[string]any{
		"type":        "SystemAssigned",
		"principalId": principalID,
		"tenantId":    "tenant-value",
	}

	identityIDs := automationIdentityIDs(accountID, identity)
	if len(identityIDs) != 1 || identityIDs[0] != accountID+"/identities/system" {
		t.Fatalf("automationIdentityIDs() = %#v, want only system identity path", identityIDs)
	}

	identityJoinIDs := automationIdentityJoinIDs(identityIDs, &principalID, nil)
	if len(identityJoinIDs) != 2 || identityJoinIDs[0] != accountID+"/identities/system" || identityJoinIDs[1] != principalID {
		t.Fatalf("automationIdentityJoinIDs() = %#v, want system identity path plus principal ID", identityJoinIDs)
	}

	relatedIDs := automationRelatedIDs(accountID, identityIDs)
	if len(relatedIDs) != 2 || relatedIDs[0] != accountID || relatedIDs[1] != accountID+"/identities/system" {
		t.Fatalf("automationRelatedIDs() = %#v, want account ID plus system identity path", relatedIDs)
	}
}

func TestAutomationIdentityHelpersPreserveUserAssignedIdentityIDs(t *testing.T) {
	accountID := "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Automation/automationAccounts/aa-ops"
	userAssignedID := "/subscriptions/sub/resourceGroups/rg-id/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-ops"

	identity := map[string]any{
		"type": "UserAssigned",
		"userAssignedIdentities": map[string]any{
			userAssignedID: map[string]any{},
		},
	}

	identityIDs := automationIdentityIDs(accountID, identity)
	if len(identityIDs) != 1 || identityIDs[0] != userAssignedID {
		t.Fatalf("automationIdentityIDs() = %#v, want only user-assigned identity ID", identityIDs)
	}
}

func TestAutomationSKUNameFallsBackToProperties(t *testing.T) {
	account := map[string]any{}
	properties := map[string]any{
		"sku": map[string]any{
			"name": "Basic",
		},
	}

	got := automationSKUName(account, properties)
	if got == nil || *got != "Basic" {
		t.Fatalf("automationSKUName() = %v, want Basic", got)
	}
}
