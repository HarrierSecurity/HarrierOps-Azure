package models

type VmAsset struct {
	ID            string   `json:"id"`
	IdentityIDs   []string `json:"identity_ids"`
	Location      string   `json:"location"`
	Name          string   `json:"name"`
	NICIDs        []string `json:"nic_ids"`
	PowerState    string   `json:"power_state"`
	PrivateIPs    []string `json:"private_ips"`
	PublicIPs     []string `json:"public_ips"`
	ResourceGroup string   `json:"resource_group"`
	VMType        string   `json:"vm_type"`
}

type VmsFinding struct {
	Description string   `json:"description"`
	ID          string   `json:"id"`
	RelatedIDs  []string `json:"related_ids"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
}

type VmsOutput struct {
	Findings []VmsFinding `json:"findings"`
	Issues   []Issue      `json:"issues"`
	Metadata Metadata     `json:"metadata"`
	VMAssets []VmAsset    `json:"vm_assets"`
}
