package providers

import (
	"slices"
	"strings"
	"testing"
)

func TestVMExtensionCommandSummaryRedactsArguments(t *testing.T) {
	t.Helper()

	settings := map[string]any{
		"commandToExecute": `powershell.exe -ExecutionPolicy Bypass -Command "Invoke-WebRequest https://example.blob.core.windows.net/bootstrap.ps1?sig=secret -Headers @{Authorization='Bearer abc'} -OutFile C:\temp\bootstrap.ps1"`,
	}

	clue := vmExtensionCommandClue(settings)
	if clue == nil {
		t.Fatal("vmExtensionCommandClue() = nil, want redacted clue")
	}
	if *clue != "public commandToExecute visible; executable=powershell.exe; arguments redacted" {
		t.Fatalf("command clue = %q, want redacted executable summary", *clue)
	}
	for _, forbidden := range []string{"Bearer", "secret", "sig=", "bootstrap.ps1", "Authorization"} {
		if strings.Contains(*clue, forbidden) {
			t.Fatalf("command clue = %q leaked forbidden token %q", *clue, forbidden)
		}
	}
}

func TestVMExtensionAssetFromPropertiesClassifiesVMSSAndProtectedSettings(t *testing.T) {
	t.Helper()

	asset := vmExtensionAssetFromProperties(
		map[string]any{"id": "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/vmss/extensions/script", "name": "script"},
		map[string]any{
			"publisher":                "Microsoft.Azure.Extensions",
			"type":                     "CustomScript",
			"typeHandlerVersion":       "2.1",
			"forceUpdateTag":           "rerun-42",
			"protectedSettings":        map[string]any{"commandToExecute": "hidden"},
			"suppressFailures":         true,
			"provisionAfterExtensions": []any{"dependency"},
			"settings": map[string]any{
				"fileUris":  []any{"https://scripts.example.invalid/bootstrap.sh?sig=secret"},
				"timestamp": "42",
			},
		},
		vmExtensionTargetContext{
			extensionID:           "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/vmss/extensions/script",
			extensionName:         "script",
			location:              "eastus",
			resourceGroup:         "rg",
			targetID:              "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/vmss",
			targetIdentityIDs:     []string{"/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/job-id"},
			targetKind:            "vmss",
			targetLabel:           "VMSS",
			targetName:            "vmss",
			vmssOrchestrationMode: stringPtr("Flexible"),
			vmssUpgradeMode:       stringPtr("Manual"),
		},
	)

	if asset.TargetKind != "vmss" || asset.TargetName != "vmss" {
		t.Fatalf("target = %s/%s, want vmss/vmss", asset.TargetKind, asset.TargetName)
	}
	if !boolPtrIsTrue(asset.ProtectedSettingsPresent) {
		t.Fatalf("ProtectedSettingsPresent = %#v, want true", asset.ProtectedSettingsPresent)
	}
	if !boolPtrIsTrue(asset.SuppressFailures) {
		t.Fatalf("SuppressFailures = %#v, want true", asset.SuppressFailures)
	}
	if !slices.Contains(asset.FileURIHosts, "scripts.example.invalid") {
		t.Fatalf("FileURIHosts = %#v, want scripts.example.invalid", asset.FileURIHosts)
	}
	if !slices.Contains(asset.RerunClues, "forceUpdateTag=rerun-42") || !slices.Contains(asset.RerunClues, "timestamp=42") {
		t.Fatalf("RerunClues = %#v, want forceUpdateTag and timestamp", asset.RerunClues)
	}
	if len(asset.SourceClues) != 1 || asset.SourceClues[0] != "fileUris hosts scripts.example.invalid" {
		t.Fatalf("SourceClues = %#v, want fileUri host only", asset.SourceClues)
	}
}
