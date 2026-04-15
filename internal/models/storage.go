package models

type StorageAsset struct {
	AllowSharedKeyAccess      *bool    `json:"allow_shared_key_access"`
	AnonymousAccessIndicators []string `json:"anonymous_access_indicators"`
	ContainerCount            *int     `json:"container_count"`
	DNSEndpointType           *string  `json:"dns_endpoint_type"`
	FileShareCount            *int     `json:"file_share_count"`
	HTTPSTrafficOnlyEnabled   *bool    `json:"https_traffic_only_enabled"`
	ID                        string   `json:"id"`
	IsHNSEnabled              *bool    `json:"is_hns_enabled"`
	IsSFTPEnabled             *bool    `json:"is_sftp_enabled"`
	Location                  *string  `json:"location"`
	MinimumTLSVersion         *string  `json:"minimum_tls_version"`
	Name                      string   `json:"name"`
	NetworkDefaultAction      *string  `json:"network_default_action"`
	NFSV3Enabled              *bool    `json:"nfs_v3_enabled"`
	PrivateEndpointEnabled    bool     `json:"private_endpoint_enabled"`
	PublicAccess              bool     `json:"public_access"`
	PublicNetworkAccess       *string  `json:"public_network_access"`
	QueueCount                *int     `json:"queue_count"`
	ResourceGroup             string   `json:"resource_group"`
	TableCount                *int     `json:"table_count"`
}

type StorageFinding struct {
	Description string   `json:"description"`
	ID          string   `json:"id"`
	RelatedIDs  []string `json:"related_ids,omitempty"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
}

type StorageMetadata = Metadata

type StorageOutput struct {
	Findings      []StorageFinding `json:"findings"`
	Issues        []Issue          `json:"issues"`
	Metadata      StorageMetadata  `json:"metadata"`
	StorageAssets []StorageAsset   `json:"storage_assets"`
}
