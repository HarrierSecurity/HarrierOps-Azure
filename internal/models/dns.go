package models

type DnsZoneAsset struct {
	ID                              string   `json:"id"`
	Name                            string   `json:"name"`
	ResourceGroup                   string   `json:"resource_group"`
	Location                        *string  `json:"location"`
	ZoneKind                        string   `json:"zone_kind"`
	RecordSetCount                  *int     `json:"record_set_count"`
	MaxRecordSetCount               *int     `json:"max_record_set_count"`
	NameServers                     []string `json:"name_servers"`
	LinkedVirtualNetworkCount       *int     `json:"linked_virtual_network_count"`
	RegistrationVirtualNetworkCount *int     `json:"registration_virtual_network_count"`
	PrivateEndpointReferenceCount   *int     `json:"private_endpoint_reference_count"`
	Summary                         string   `json:"summary"`
	RelatedIDs                      []string `json:"related_ids"`
}

type DnsMetadata = RuntimeCommandMetadata

type DnsOutput struct {
	DNSZones []DnsZoneAsset `json:"dns_zones"`
	Findings []Finding      `json:"findings"`
	Issues   []Issue        `json:"issues"`
	Metadata DnsMetadata    `json:"metadata"`
}
