package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) VMExtensions(_ context.Context, tenant string, subscription string) (VMExtensionsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID
	vmID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01"
	vmssID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-compute/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-batch"
	identityID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app"

	return VMExtensionsFacts{
		ArtifactIdentityFacts: staticArtifactIdentityFacts(session),
		TenantID:              session.TenantID,
		SubscriptionID:        subscriptionID,
		VMExtensions: []models.VMExtensionAsset{
			{
				AutoUpgradeMinorVersion:   boolPtr(true),
				CommandClue:               models.StringPtr("public commandToExecute visible; executable=powershell; arguments redacted"),
				EnableAutomaticUpgrade:    boolPtr(false),
				ExtensionType:             models.StringPtr("CustomScriptExtension"),
				FileURIHosts:              []string{"raw.githubusercontent.com", "storageacct.blob.core.windows.net"},
				FileURICount:              intPtr(2),
				ForceUpdateTag:            models.StringPtr("2026-04-23T120000Z"),
				ID:                        vmID + "/extensions/config-bootstrap",
				InstanceViewStatuses:      []string{"ProvisioningState/succeeded"},
				KeyVaultProtectedSettings: nil,
				Location:                  "eastus",
				Name:                      "config-bootstrap",
				ProtectedSettingsPresent:  boolPtr(true),
				ProvisionAfterExtensions:  []string{},
				ProvisioningState:         models.StringPtr("Succeeded"),
				Publisher:                 models.StringPtr("Microsoft.Compute"),
				PublicSettingKeys:         []string{"commandToExecute", "fileUris"},
				RelatedIDs:                []string{vmID, identityID, vmID + "/extensions/config-bootstrap"},
				ResourceGroup:             "rg-workload",
				RerunClues:                []string{"forceUpdateTag=2026-04-23T120000Z"},
				SourceClues:               []string{"fileUris hosts raw.githubusercontent.com, storageacct.blob.core.windows.net", "public commandToExecute visible"},
				Summary:                   "VM extension 'config-bootstrap' targets VM 'vm-web-01' with Microsoft.Compute/CustomScriptExtension 1.10, shows public file URI hosts raw.githubusercontent.com and storageacct.blob.core.windows.net, exposes a public command clue, and has protected settings present. Visible status: Succeeded.",
				SuppressFailures:          boolPtr(false),
				TargetID:                  vmID,
				TargetIdentityIDs:         []string{identityID},
				TargetKind:                "vm",
				TargetName:                "vm-web-01",
				TypeHandlerVersion:        models.StringPtr("1.10"),
			},
			{
				AutoUpgradeMinorVersion:   boolPtr(true),
				EnableAutomaticUpgrade:    boolPtr(true),
				ExtensionType:             models.StringPtr("DependencyAgentLinux"),
				FileURIHosts:              []string{},
				FileURICount:              intPtr(0),
				ID:                        vmID + "/extensions/dependency-agent",
				InstanceViewStatuses:      []string{"ProvisioningState/succeeded"},
				KeyVaultProtectedSettings: nil,
				Location:                  "eastus",
				Name:                      "dependency-agent",
				ProtectedSettingsPresent:  nil,
				ProvisionAfterExtensions:  []string{},
				ProvisioningState:         models.StringPtr("Succeeded"),
				Publisher:                 models.StringPtr("Microsoft.Azure.Monitoring.DependencyAgent"),
				PublicSettingKeys:         []string{},
				RelatedIDs:                []string{vmID, identityID, vmID + "/extensions/dependency-agent"},
				ResourceGroup:             "rg-workload",
				RerunClues:                []string{},
				SourceClues:               []string{},
				Summary:                   "VM extension 'dependency-agent' targets VM 'vm-web-01' with Microsoft.Azure.Monitoring.DependencyAgent/DependencyAgentLinux 9.10, does not expose public script or command source clues, and does not expose protected-settings metadata in this response. Visible status: Succeeded.",
				SuppressFailures:          boolPtr(false),
				TargetID:                  vmID,
				TargetIdentityIDs:         []string{identityID},
				TargetKind:                "vm",
				TargetName:                "vm-web-01",
				TypeHandlerVersion:        models.StringPtr("9.10"),
			},
			{
				AutoUpgradeMinorVersion:   boolPtr(false),
				CommandClue:               models.StringPtr("public commandToExecute visible; executable=maintenance.sh; arguments redacted"),
				EnableAutomaticUpgrade:    boolPtr(false),
				ExtensionType:             models.StringPtr("CustomScript"),
				FileURIHosts:              []string{"scripts.contoso.internal"},
				FileURICount:              intPtr(1),
				ForceUpdateTag:            models.StringPtr("roll-20260423"),
				ID:                        vmssID + "/extensions/maintenance-script",
				InstanceViewStatuses:      []string{},
				KeyVaultProtectedSettings: boolPtr(true),
				Location:                  "eastus",
				Name:                      "maintenance-script",
				ProtectedSettingsPresent:  nil,
				ProvisionAfterExtensions:  []string{},
				ProvisioningState:         models.StringPtr("Succeeded"),
				Publisher:                 models.StringPtr("Microsoft.Azure.Extensions"),
				PublicSettingKeys:         []string{"commandToExecute", "fileUris"},
				RelatedIDs:                []string{vmssID, vmssID + "/extensions/maintenance-script"},
				ResourceGroup:             "rg-compute",
				RerunClues:                []string{"forceUpdateTag=roll-20260423"},
				SourceClues:               []string{"fileUris hosts scripts.contoso.internal", "public commandToExecute visible"},
				Summary:                   "VMSS extension 'maintenance-script' targets scale set 'vmss-batch' with Microsoft.Azure.Extensions/CustomScript 2.1, shows public file URI host scripts.contoso.internal, exposes a public command clue, and uses Key Vault-referenced protected settings. Visible status: Succeeded.",
				SuppressFailures:          boolPtr(false),
				TargetID:                  vmssID,
				TargetIdentityIDs:         []string{},
				TargetKind:                "vmss",
				TargetName:                "vmss-batch",
				TypeHandlerVersion:        models.StringPtr("2.1"),
				VMSSOrchestrationMode:     models.StringPtr("Uniform"),
				VMSSUpgradeMode:           models.StringPtr("Rolling"),
			},
		},
		Issues: []models.Issue{},
	}, nil
}
