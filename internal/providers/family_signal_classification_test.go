package providers

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestAppInsightsSettingClassBoundaries(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "APPLICATIONINSIGHTS_CONNECTION_STRING", want: "instrumentation"},
		{name: "APPINSIGHTS_SAMPLING_PERCENTAGE", want: "sampling"},
		{name: "ApplicationInsights:TelemetryProcessor:DropHealthChecks", want: "filtering"},
		{name: "Logging:LogLevel:ApplicationInsights", want: "logging-level"},
		{name: "FEATURE_FILTER_ENABLED", want: ""},
		{name: "LOGLEVEL_DEFAULT", want: ""},
		{name: "unrelated", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := appInsightsSettingClass(tc.name); got != tc.want {
				t.Fatalf("appInsightsSettingClass(%q)=%q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestDCRStreamSignalRankBoundaries(t *testing.T) {
	tests := []struct {
		stream string
		want   int
	}{
		{stream: "Microsoft-SecurityEvent", want: 5},
		{stream: "Microsoft-WindowsEvent", want: 4},
		{stream: "Microsoft-Syslog", want: 3},
		{stream: "Microsoft-Process", want: 2},
		{stream: "Microsoft-Perf", want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.stream, func(t *testing.T) {
			if got := dcrStreamSignalRank(tc.stream); got != tc.want {
				t.Fatalf("dcrStreamSignalRank(%q)=%d, want %d", tc.stream, got, tc.want)
			}
		})
	}
}

func TestDiagnosticSettingsSignalClassifiers(t *testing.T) {
	if !diagnosticSettingsSourceLooksHighSignal(models.DiagnosticSettingsSource{Type: "Microsoft.KeyVault/vaults"}) {
		t.Fatal("expected Key Vault source to be high signal")
	}
	if diagnosticSettingsSourceLooksHighSignal(models.DiagnosticSettingsSource{Type: "Microsoft.Network/networkWatchers"}) {
		t.Fatal("did not expect Network Watcher source to be high signal")
	}
	if !diagnosticSettingsCategoryLooksHighSignal("AuditEvent") {
		t.Fatal("expected AuditEvent category to be high signal")
	}
	if diagnosticSettingsCategoryLooksHighSignal("AllMetrics") {
		t.Fatal("did not expect AllMetrics category to be high signal")
	}
}

func TestDiagnosticSettingsAllLogsCoversSupportedLogCategories(t *testing.T) {
	supported := []models.DiagnosticSettingsCategory{
		{Name: "AuditEvent", Type: "log", Enabled: true},
		{Name: "SecretNearExpiryEvent", Type: "log", Enabled: true},
		{Name: "AllMetrics", Type: "metric", Enabled: true},
	}

	got := diagnosticSettingsNotExportedSupported(supported, []string{"allLogs"})
	want := []string{"AllMetrics"}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("diagnosticSettingsNotExportedSupported(allLogs)=%v, want %v", got, want)
	}
}

func TestAPIMPolicyControlTypeClassifiers(t *testing.T) {
	policies := []map[string]any{
		{"properties": map[string]any{"value": `<policies><inbound><choose><when condition="true"><set-backend-service base-url="https://api.example" /><set-header name="x-route" exists-action="override" /></when></choose></inbound></policies>`}},
		{"properties": map[string]any{"policyContent": `<policies><inbound><rewrite-uri template="/v2" /><send-request mode="new" /></inbound></policies>`}},
		{"properties": map[string]any{"value": `<policies><inbound><base /></inbound></policies>`}},
	}

	got := apiMgmtPolicyControlTypes(policies)
	want := []string{"backend-routing", "header-auth", "conditional-routing", "request-rewrite", "side-request"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("apiMgmtPolicyControlTypes=%v, want %v", got, want)
	}
}

func TestAutomationRunbookClueClassifiers(t *testing.T) {
	content := `
Connect-AzAccount -Identity
Get-AzResource -ResourceType Microsoft.Web/sites
Invoke-RestMethod -Uri https://management.azure.com/subscriptions/demo/providers/Microsoft.KeyVault/vaults
kubectl get pods
`

	commandClues := automationRunbookCommandClues(content)
	wantCommands := []string{"connect-azaccount", "get-az", "invoke-restmethod", "kubectl"}
	if strings.Join(commandClues, ",") != strings.Join(wantCommands, ",") {
		t.Fatalf("automationRunbookCommandClues=%v, want %v", commandClues, wantCommands)
	}

	resourceClues := automationRunbookResourceClues(content)
	wantResources := []string{"app-service", "key-vault", "azure-management-api"}
	if strings.Join(resourceClues, ",") != strings.Join(wantResources, ",") {
		t.Fatalf("automationRunbookResourceClues=%v, want %v", resourceClues, wantResources)
	}
}
