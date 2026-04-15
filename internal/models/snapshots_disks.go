package models

type SnapshotDiskAsset struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	AssetKind           string   `json:"asset_kind"`
	ResourceGroup       string   `json:"resource_group"`
	Location            *string  `json:"location"`
	DiskRole            *string  `json:"disk_role"`
	AttachmentState     string   `json:"attachment_state"`
	AttachedToID        *string  `json:"attached_to_id"`
	AttachedToName      *string  `json:"attached_to_name"`
	SourceResourceID    *string  `json:"source_resource_id"`
	SourceResourceName  *string  `json:"source_resource_name"`
	SourceResourceKind  *string  `json:"source_resource_kind"`
	OSType              *string  `json:"os_type"`
	SizeGB              *int     `json:"size_gb"`
	TimeCreated         *string  `json:"time_created"`
	Incremental         *bool    `json:"incremental"`
	NetworkAccessPolicy *string  `json:"network_access_policy"`
	PublicNetworkAccess *string  `json:"public_network_access"`
	DiskAccessID        *string  `json:"disk_access_id"`
	MaxShares           *int     `json:"max_shares"`
	EncryptionType      *string  `json:"encryption_type"`
	DiskEncryptionSetID *string  `json:"disk_encryption_set_id"`
	Summary             string   `json:"summary"`
	RelatedIDs          []string `json:"related_ids"`
}

type SnapshotsDisksMetadata struct {
	SchemaVersion  string  `json:"schema_version"`
	Command        string  `json:"command"`
	GeneratedAt    string  `json:"generated_at"`
	TenantID       *string `json:"tenant_id"`
	SubscriptionID *string `json:"subscription_id"`
	TokenSource    *string `json:"token_source"`
}

type SnapshotsDisksOutput struct {
	Metadata           SnapshotsDisksMetadata `json:"metadata"`
	SnapshotDiskAssets []SnapshotDiskAsset    `json:"snapshot_disk_assets"`
	Findings           []Finding              `json:"findings"`
	Issues             []Issue                `json:"issues"`
}
