package providers

import (
	"context"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"

	"harrierops-azure/internal/models"
)

func (provider AzureProvider) VMExtensions(ctx context.Context, tenant string, subscription string) (VMExtensionsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return VMExtensionsFacts{}, err
	}

	state, err := provider.computeNetworkState(session)
	if err != nil {
		return VMExtensionsFacts{}, err
	}

	vms := state.vmSnapshot(ctx)
	vmss := state.vmssSnapshot(ctx)
	extensions := []models.VMExtensionAsset{}
	issues := append([]models.Issue{}, vms.issues...)
	issues = append(issues, vmss.issues...)

	for _, vm := range vms.assets {
		rows, extensionIssues := state.collector.collectVMExtensionAssets(ctx, vm)
		extensions = append(extensions, rows...)
		issues = append(issues, extensionIssues...)
	}
	for _, scaleSet := range vmss.assets {
		rows, extensionIssues := state.collector.collectVMSSVMExtensionAssets(ctx, scaleSet)
		extensions = append(extensions, rows...)
		issues = append(issues, extensionIssues...)
	}

	return VMExtensionsFacts{
		ArtifactIdentityFacts: azureArtifactIdentityFacts(session),
		TenantID:              session.tenantID,
		SubscriptionID:        session.subscription.ID,
		VMExtensions:          extensions,
		Issues:                issues,
	}, nil
}

func (collector computeNetworkCollector) collectVMExtensionAssets(ctx context.Context, vm models.VmAsset) ([]models.VMExtensionAsset, []models.Issue) {
	response, err := collector.clients.vmExtensions.List(ctx, vm.ResourceGroup, vm.Name, nil)
	if err != nil {
		return nil, []models.Issue{issueFromError("vm-extensions.vm["+vm.ResourceGroup+"/"+vm.Name+"]", err)}
	}

	extensions := []models.VMExtensionAsset{}
	for _, extension := range response.Value {
		if extension == nil {
			continue
		}
		extensions = append(extensions, vmExtensionAssetFromVM(extension, vm))
	}
	return extensions, nil
}

func (collector computeNetworkCollector) collectVMSSVMExtensionAssets(ctx context.Context, vmss models.VmssAsset) ([]models.VMExtensionAsset, []models.Issue) {
	extensions := []models.VMExtensionAsset{}
	issues := []models.Issue{}
	pager := collector.clients.vmssExtensions.NewListPager(vmss.ResourceGroup, vmss.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("vm-extensions.vmss["+vmss.ResourceGroup+"/"+vmss.Name+"]", err))
			break
		}
		for _, extension := range page.Value {
			if extension == nil {
				continue
			}
			extensions = append(extensions, vmExtensionAssetFromVMSS(extension, vmss))
		}
	}
	return extensions, issues
}

func vmExtensionAssetFromVM(extension *armcompute.VirtualMachineExtension, target models.VmAsset) models.VMExtensionAsset {
	raw := map[string]any{}
	decodeJSONInto(extension, &raw)
	properties := mapValue(raw, "properties")
	extensionID := firstNonEmpty(mapStringValue(raw, "id"), stringValue(extension.ID), target.ID+"/extensions/"+firstNonEmpty(mapStringValue(raw, "name"), stringValue(extension.Name), "unknown"))
	name := firstNonEmpty(mapStringValue(raw, "name"), stringValue(extension.Name), resourceNameFromID(extensionID), "unknown")

	asset := vmExtensionAssetFromProperties(raw, properties, vmExtensionTargetContext{
		extensionID:       extensionID,
		extensionName:     name,
		location:          firstNonEmpty(mapStringValue(raw, "location"), stringValue(extension.Location), target.Location),
		resourceGroup:     firstNonEmpty(resourceGroupFromID(extensionID), target.ResourceGroup),
		targetID:          target.ID,
		targetIdentityIDs: target.IdentityIDs,
		targetKind:        "vm",
		targetLabel:       "VM",
		targetName:        target.Name,
	})
	asset.InstanceViewStatuses = vmExtensionInstanceViewStatuses(properties)
	asset.Summary = vmExtensionSummary(asset)
	return asset
}

func vmExtensionAssetFromVMSS(extension *armcompute.VirtualMachineScaleSetExtension, target models.VmssAsset) models.VMExtensionAsset {
	raw := map[string]any{}
	decodeJSONInto(extension, &raw)
	properties := mapValue(raw, "properties")
	extensionID := firstNonEmpty(mapStringValue(raw, "id"), stringValue(extension.ID), target.ID+"/extensions/"+firstNonEmpty(mapStringValue(raw, "name"), stringValue(extension.Name), "unknown"))
	name := firstNonEmpty(mapStringValue(raw, "name"), stringValue(extension.Name), resourceNameFromID(extensionID), "unknown")

	asset := vmExtensionAssetFromProperties(raw, properties, vmExtensionTargetContext{
		extensionID:           extensionID,
		extensionName:         name,
		location:              firstNonEmpty(mapStringValue(raw, "location"), target.Location),
		resourceGroup:         firstNonEmpty(resourceGroupFromID(extensionID), target.ResourceGroup),
		targetID:              target.ID,
		targetIdentityIDs:     target.IdentityIDs,
		targetKind:            "vmss",
		targetLabel:           "VMSS",
		targetName:            target.Name,
		vmssOrchestrationMode: target.OrchestrationMode,
		vmssUpgradeMode:       target.UpgradeMode,
	})
	asset.Summary = vmExtensionSummary(asset)
	return asset
}

type vmExtensionTargetContext struct {
	extensionID           string
	extensionName         string
	location              string
	resourceGroup         string
	targetID              string
	targetIdentityIDs     []string
	targetKind            string
	targetLabel           string
	targetName            string
	vmssOrchestrationMode *string
	vmssUpgradeMode       *string
}

func vmExtensionAssetFromProperties(raw map[string]any, properties map[string]any, target vmExtensionTargetContext) models.VMExtensionAsset {
	settings := mapValue(properties, "settings")
	publicSettingKeys := sortedKeys(settings)
	fileURIs := vmExtensionFileURIs(settings)
	fileHosts := vmExtensionFileURIHosts(fileURIs)
	fileCount := len(fileURIs)
	commandClue := vmExtensionCommandClue(settings)
	protectedSettingsPresent := vmExtensionOptionalNonEmptyBool(properties, "protectedSettings", "protected_settings")
	keyVaultProtectedSettings := vmExtensionOptionalNonEmptyBool(properties, "protectedSettingsFromKeyVault", "protected_settings_from_key_vault")

	sourceClues := []string{}
	if len(fileHosts) > 0 {
		sourceClues = append(sourceClues, "fileUris hosts "+strings.Join(fileHosts, ", "))
	}
	if commandClue != nil {
		sourceClues = append(sourceClues, "public commandToExecute visible")
	}

	rerunClues := []string{}
	if tag := firstNonEmpty(mapStringValue(properties, "forceUpdateTag", "force_update_tag"), mapStringValue(raw, "forceUpdateTag", "force_update_tag")); tag != "" {
		rerunClues = append(rerunClues, "forceUpdateTag="+tag)
	}
	if timestamp := firstNonEmpty(mapStringValue(settings, "timestamp"), strconv.Itoa(mapIntValue(settings, "timestamp"))); timestamp != "" && timestamp != "0" {
		rerunClues = append(rerunClues, "timestamp="+timestamp)
	}

	return models.VMExtensionAsset{
		AutoUpgradeMinorVersion:   optionalBoolPtr(properties, "autoUpgradeMinorVersion", "auto_upgrade_minor_version"),
		CommandClue:               commandClue,
		EnableAutomaticUpgrade:    optionalBoolPtr(properties, "enableAutomaticUpgrade", "enable_automatic_upgrade"),
		ExtensionType:             stringPtr(firstNonEmpty(mapStringValue(properties, "type"), mapStringValue(raw, "type"))),
		FileURIHosts:              fileHosts,
		FileURICount:              intPtr(fileCount),
		ForceUpdateTag:            stringPtr(firstNonEmpty(mapStringValue(properties, "forceUpdateTag", "force_update_tag"), mapStringValue(raw, "forceUpdateTag", "force_update_tag"))),
		ID:                        firstNonEmpty(target.extensionID, "/unknown/"+target.extensionName),
		InstanceViewStatuses:      []string{},
		KeyVaultProtectedSettings: keyVaultProtectedSettings,
		Location:                  target.location,
		Name:                      target.extensionName,
		ProtectedSettingsPresent:  protectedSettingsPresent,
		ProvisionAfterExtensions:  stringListValue(properties, "provisionAfterExtensions", "provision_after_extensions"),
		ProvisioningState:         stringPtr(mapStringValue(properties, "provisioningState", "provisioning_state")),
		Publisher:                 stringPtr(mapStringValue(properties, "publisher")),
		PublicSettingKeys:         publicSettingKeys,
		RelatedIDs:                sortedUniqueStrings(append([]string{target.extensionID, target.targetID}, target.targetIdentityIDs...)),
		ResourceGroup:             target.resourceGroup,
		RerunClues:                sortedUniqueStrings(rerunClues),
		SourceClues:               sortedUniqueStrings(sourceClues),
		SuppressFailures:          optionalBoolPtr(properties, "suppressFailures", "suppress_failures"),
		TargetID:                  target.targetID,
		TargetIdentityIDs:         sortedUniqueStrings(target.targetIdentityIDs),
		TargetKind:                target.targetKind,
		TargetName:                target.targetName,
		TypeHandlerVersion:        stringPtr(mapStringValue(properties, "typeHandlerVersion", "type_handler_version")),
		VMSSOrchestrationMode:     target.vmssOrchestrationMode,
		VMSSUpgradeMode:           target.vmssUpgradeMode,
	}
}

func vmExtensionFileURIs(settings map[string]any) []string {
	values := stringListValue(settings, "fileUris", "file_uris")
	if len(values) > 0 {
		return values
	}
	if uri := firstNonEmpty(mapStringValue(settings, "fileUri", "file_uri"), mapStringValue(settings, "scriptUri", "script_uri")); uri != "" {
		return []string{uri}
	}
	return []string{}
}

func vmExtensionFileURIHosts(fileURIs []string) []string {
	hosts := []string{}
	for _, fileURI := range fileURIs {
		if host := hostnameFromURL(fileURI); strings.TrimSpace(host) != "" {
			hosts = append(hosts, host)
		}
	}
	return sortedUniqueStrings(hosts)
}

func vmExtensionCommandClue(settings map[string]any) *string {
	command := firstNonEmpty(mapStringValue(settings, "commandToExecute", "command_to_execute"), mapStringValue(settings, "command", "script"))
	if strings.TrimSpace(command) == "" {
		return nil
	}
	return stringPtr(vmExtensionCommandSummary(command))
}

func vmExtensionCommandSummary(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "public commandToExecute visible; command text redacted"
	}
	executable := vmExtensionCommandExecutable(fields[0])
	if executable == "" {
		return "public commandToExecute visible; command text redacted"
	}
	return "public commandToExecute visible; executable=" + executable + "; arguments redacted"
}

func vmExtensionCommandExecutable(value string) string {
	value = strings.Trim(strings.TrimSpace(value), `"'`)
	if value == "" {
		return ""
	}
	if index := strings.LastIndexAny(value, `/\`); index >= 0 && index < len(value)-1 {
		value = value[index+1:]
	}
	value = strings.Trim(strings.TrimSpace(value), `"'`)
	if value == "" || strings.Contains(value, "://") {
		return ""
	}
	return value
}

func vmExtensionHasNonEmptyValue(input map[string]any, keys ...string) bool {
	for _, key := range keys {
		value, exists := input[key]
		if !exists {
			continue
		}
		switch typed := value.(type) {
		case nil:
			return false
		case map[string]any:
			return len(typed) > 0
		case []any:
			return len(typed) > 0
		case string:
			return strings.TrimSpace(typed) != ""
		default:
			return true
		}
	}
	return false
}

func vmExtensionOptionalNonEmptyBool(input map[string]any, keys ...string) *bool {
	for _, key := range keys {
		if _, exists := input[key]; exists {
			return boolPtr(vmExtensionHasNonEmptyValue(input, key))
		}
	}
	return nil
}

func vmExtensionInstanceViewStatuses(properties map[string]any) []string {
	instanceView := mapValue(properties, "instanceView", "instance_view")
	values := []string{}
	for _, raw := range append(listValue(instanceView, "statuses"), listValue(instanceView, "substatuses", "sub_statuses")...) {
		status := mapValue(raw)
		value := firstNonEmpty(mapStringValue(status, "code"), mapStringValue(status, "displayStatus", "display_status"))
		if value != "" {
			values = append(values, value)
		}
	}
	return sortedUniqueStrings(values)
}

func vmExtensionSummary(asset models.VMExtensionAsset) string {
	targetLabel := "target"
	switch asset.TargetKind {
	case "vm":
		targetLabel = "VM"
	case "vmss":
		targetLabel = "VMSS"
	}

	handler := firstNonEmpty(stringPtrValue(asset.Publisher), "unknown publisher") + "/" + firstNonEmpty(stringPtrValue(asset.ExtensionType), "unknown extension")
	if version := stringPtrValue(asset.TypeHandlerVersion); version != "" {
		handler += " " + version
	}

	sourceText := "does not expose public script or command source clues"
	if len(asset.FileURIHosts) > 0 {
		sourceText = "shows public file URI " + pluralize("host", len(asset.FileURIHosts)) + " " + humanJoin(asset.FileURIHosts)
	}
	if asset.CommandClue != nil {
		if len(asset.FileURIHosts) > 0 {
			sourceText += ", exposes a public command clue"
		} else {
			sourceText = "exposes a public command clue"
		}
	}

	protectedText := "does not expose protected-settings metadata in this response"
	if boolPtrIsTrue(asset.KeyVaultProtectedSettings) {
		protectedText = "uses Key Vault-referenced protected settings"
	} else if boolPtrIsTrue(asset.ProtectedSettingsPresent) {
		protectedText = "has protected settings present"
	}

	statusText := "Visible status: " + firstNonEmpty(stringPtrValue(asset.ProvisioningState), "unknown") + "."
	if len(asset.InstanceViewStatuses) > 0 {
		statusText = "Visible status: " + humanJoin(asset.InstanceViewStatuses) + "."
	}

	return targetLabel + " extension '" + firstNonEmpty(asset.Name, "unknown") + "' targets " + targetLabel + " '" + firstNonEmpty(asset.TargetName, "unknown") + "' with " + handler + ", " + sourceText + ", and " + protectedText + ". " + statusText
}

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

func humanJoin(values []string) string {
	cleaned := []string{}
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			cleaned = append(cleaned, value)
		}
	}
	switch len(cleaned) {
	case 0:
		return ""
	case 1:
		return cleaned[0]
	case 2:
		return cleaned[0] + " and " + cleaned[1]
	default:
		return strings.Join(cleaned[:len(cleaned)-1], ", ") + ", and " + cleaned[len(cleaned)-1]
	}
}
