package models

type DatabaseServerAsset struct {
	DatabaseCount             *int     `json:"database_count"`
	DelegatedSubnetResourceID *string  `json:"delegated_subnet_resource_id"`
	Engine                    string   `json:"engine"`
	FullyQualifiedDomainName  *string  `json:"fully_qualified_domain_name"`
	HighAvailabilityMode      *string  `json:"high_availability_mode"`
	ID                        string   `json:"id"`
	Location                  *string  `json:"location"`
	MinimalTLSVersion         *string  `json:"minimal_tls_version"`
	Name                      string   `json:"name"`
	PrivateDNSZoneResourceID  *string  `json:"private_dns_zone_resource_id"`
	PublicNetworkAccess       *string  `json:"public_network_access"`
	RelatedIDs                []string `json:"related_ids"`
	ResourceGroup             string   `json:"resource_group"`
	ServerVersion             *string  `json:"server_version"`
	State                     *string  `json:"state"`
	Summary                   string   `json:"summary"`
	UserDatabaseNames         []string `json:"user_database_names"`
	WorkloadClientID          *string  `json:"workload_client_id"`
	WorkloadIdentityIDs       []string `json:"workload_identity_ids"`
	WorkloadIdentityType      *string  `json:"workload_identity_type"`
	WorkloadPrincipalID       *string  `json:"workload_principal_id"`
}

type DatabasesMetadata = RuntimeCommandMetadata

type DatabasesOutput struct {
	DatabaseServers []DatabaseServerAsset `json:"database_servers"`
	Findings        []Finding             `json:"findings"`
	Issues          []Issue               `json:"issues"`
	Metadata        DatabasesMetadata     `json:"metadata"`
}
