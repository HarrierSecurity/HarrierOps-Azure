package models

type NicAsset struct {
	AttachedAssetID        *string  `json:"attached_asset_id"`
	AttachedAssetName      *string  `json:"attached_asset_name"`
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	NetworkSecurityGroupID *string  `json:"network_security_group_id"`
	PrivateIPs             []string `json:"private_ips"`
	PublicIPIDs            []string `json:"public_ip_ids"`
	SubnetIDs              []string `json:"subnet_ids"`
	VnetIDs                []string `json:"vnet_ids"`
}

type NicsOutput struct {
	Findings  []Finding              `json:"findings"`
	Issues    []Issue                `json:"issues"`
	Metadata  NetworkCommandMetadata `json:"metadata"`
	NicAssets []NicAsset             `json:"nic_assets"`
}
