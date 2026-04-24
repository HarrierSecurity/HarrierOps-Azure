package models

type VMExtensionAsset struct {
	AutoUpgradeMinorVersion   *bool    `json:"auto_upgrade_minor_version"`
	CommandClue               *string  `json:"command_clue,omitempty"`
	EnableAutomaticUpgrade    *bool    `json:"enable_automatic_upgrade"`
	ExtensionType             *string  `json:"extension_type"`
	FileURIHosts              []string `json:"file_uri_hosts"`
	FileURICount              *int     `json:"file_uri_count"`
	ForceUpdateTag            *string  `json:"force_update_tag,omitempty"`
	ID                        string   `json:"id"`
	InstanceViewStatuses      []string `json:"instance_view_statuses"`
	KeyVaultProtectedSettings *bool    `json:"key_vault_protected_settings"`
	Location                  string   `json:"location"`
	Name                      string   `json:"name"`
	ProtectedSettingsPresent  *bool    `json:"protected_settings_present"`
	ProvisionAfterExtensions  []string `json:"provision_after_extensions"`
	ProvisioningState         *string  `json:"provisioning_state"`
	Publisher                 *string  `json:"publisher"`
	PublicSettingKeys         []string `json:"public_setting_keys"`
	RelatedIDs                []string `json:"related_ids"`
	ResourceGroup             string   `json:"resource_group"`
	RerunClues                []string `json:"rerun_clues"`
	SourceClues               []string `json:"source_clues"`
	Summary                   string   `json:"summary"`
	SuppressFailures          *bool    `json:"suppress_failures"`
	TargetID                  string   `json:"target_id"`
	TargetIdentityIDs         []string `json:"target_identity_ids"`
	TargetKind                string   `json:"target_kind"`
	TargetName                string   `json:"target_name"`
	TypeHandlerVersion        *string  `json:"type_handler_version"`
	VMSSOrchestrationMode     *string  `json:"vmss_orchestration_mode,omitempty"`
	VMSSUpgradeMode           *string  `json:"vmss_upgrade_mode,omitempty"`
}

type VMExtensionsMetadata = RuntimeCommandMetadata

type VMExtensionsOutput struct {
	Findings     []Finding            `json:"findings"`
	Issues       []Issue              `json:"issues"`
	Metadata     VMExtensionsMetadata `json:"metadata"`
	VMExtensions []VMExtensionAsset   `json:"vm_extensions"`
}
