package render

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestFunctionsTableUsesCompactRowPlusNoteLayout(t *testing.T) {
	runtime := "PYTHON|3.11"
	functionsVersion := "~4"
	hostname := "func-orders.azurewebsites.net"
	identityType := "SystemAssigned, UserAssigned"
	publicAccess := "Enabled"
	minTLS := "1.2"
	ftps := "Disabled"
	storageType := "plain-text"
	runFromPackage := false
	keyVaultRefs := 1
	alwaysOn := true
	httpTrigger := "HTTP"
	timerTrigger := "timer"
	falseValue := false

	output, err := Table("functions", models.FunctionsOutput{
		FunctionApps: []models.FunctionAppAsset{{
			Name:                         "func-orders",
			DefaultHostname:              &hostname,
			RuntimeStack:                 &runtime,
			FunctionsExtensionVersion:    &functionsVersion,
			WorkloadIdentityType:         &identityType,
			WorkloadIdentityIDs:          []string{"identity-1", "identity-2"},
			AzureWebJobsStorageValueType: &storageType,
			RunFromPackage:               &runFromPackage,
			KeyVaultReferenceCount:       &keyVaultRefs,
			HTTPSOnly:                    true,
			PublicNetworkAccess:          &publicAccess,
			MinTLSVersion:                &minTLS,
			FTPSState:                    &ftps,
			AlwaysOn:                     &alwaysOn,
			Summary:                      "Function App 'func-orders' publishes hostname 'func-orders.azurewebsites.net', runs runtime 'PYTHON|3.11', targets Functions runtime '~4', and uses managed identity (SystemAssigned, UserAssigned). Deployment signals: AzureWebJobsStorage as plain-text app setting, 1 Key Vault-backed setting(s). Visible posture: public network access Enabled, HTTPS-only enabled, TLS 1.2, FTPS Disabled, Always On enabled.",
			TriggerTypes:                 []string{httpTrigger, timerTrigger},
			VisibleFunctions: []models.FunctionChildAsset{
				{Name: "OrdersWebhook", TriggerType: &httpTrigger, IsDisabled: &falseValue, InvokeURLTemplate: &hostname, BindingTypes: []string{"httpTrigger", "http"}},
				{Name: "NightlyReconcile", TriggerType: &timerTrigger, IsDisabled: &falseValue, BindingTypes: []string{"timerTrigger"}},
			},
		}},
	}, models.RenderContext{})
	if err != nil {
		t.Fatalf("Table(functions) returned error: %v", err)
	}

	if !strings.Contains(output, "table view is compact by design; the JSON artifact keeps the fuller visible field set") {
		t.Fatalf("expected compact-view hint in output, got:\n%s", output)
	}
	if !strings.Contains(output, "│ note ") {
		t.Fatalf("expected wrapped note section in output, got:\n%s", output)
	}
	if !strings.Contains(output, "triggers") {
		t.Fatalf("expected triggers column in output, got:\n%s", output)
	}
	if !strings.Contains(output, "2 fn; HTTP, timer") {
		t.Fatalf("expected trigger summary in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Visible child functions: OrdersWebhook") {
		t.Fatalf("expected child-function note in output, got:\n%s", output)
	}
	if strings.Contains(output, "│ operator signal") {
		t.Fatalf("did not expect operator signal section in output, got:\n%s", output)
	}
	if strings.Contains(output, "why it matters") {
		t.Fatalf("did not expect legacy why-it-matters column in output, got:\n%s", output)
	}
	if !strings.Contains(output, "│ func-orders ") {
		t.Fatalf("expected padded main-row cell spacing in output, got:\n%s", output)
	}
}
