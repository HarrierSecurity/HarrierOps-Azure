package providers

import (
	"context"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

const armSQLServersAPIVersion = "2023-08-01"
const armPostgreSQLFlexibleServersAPIVersion = "2024-08-01"
const armMySQLFlexibleServersAPIVersion = "2023-12-30"

func (provider AzureProvider) Databases(ctx context.Context, tenant string, subscription string) (DatabasesFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return DatabasesFacts{}, err
	}

	rows := []models.DatabaseServerAsset{}
	issues := []models.Issue{}

	sqlServers, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.Sql/servers",
		armSQLServersAPIVersion,
	)
	if err != nil {
		issues = append(issues, issueFromError("databases.sql_servers", err))
	} else {
		for _, server := range sqlServers {
			rows = append(rows, databaseServerSummary(server, databaseListByServer(ctx, session, server, armSQLServersAPIVersion, &issues), "AzureSql"))
		}
	}

	postgresServers, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.DBforPostgreSQL/flexibleServers",
		armPostgreSQLFlexibleServersAPIVersion,
	)
	if err != nil {
		issues = append(issues, issueFromError("databases.postgresql_flexible_servers", err))
	} else {
		for _, server := range postgresServers {
			rows = append(rows, databaseServerSummary(server, databaseListByServer(ctx, session, server, armPostgreSQLFlexibleServersAPIVersion, &issues), "PostgreSqlFlexible"))
		}
	}

	mysqlServers, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.DBforMySQL/flexibleServers",
		armMySQLFlexibleServersAPIVersion,
	)
	if err != nil {
		issues = append(issues, issueFromError("databases.mysql_flexible_servers", err))
	} else {
		for _, server := range mysqlServers {
			rows = append(rows, databaseServerSummary(server, databaseListByServer(ctx, session, server, armMySQLFlexibleServersAPIVersion, &issues), "MySqlFlexible"))
		}
	}

	return DatabasesFacts{
		TenantID:        session.tenantID,
		SubscriptionID:  session.subscription.ID,
		DatabaseServers: rows,
		Issues:          issues,
	}, nil
}

func databaseListByServer(
	ctx context.Context,
	session azureSession,
	server map[string]any,
	apiVersion string,
	issues *[]models.Issue,
) []map[string]any {
	serverID := mapStringValue(server, "id")
	resourceGroup, serverName := resourceGroupAndNameFromID(serverID)
	if serverID == "" || resourceGroup == "" || serverName == "" {
		return nil
	}

	databases, err := armListObjects(ctx, session.credential, serverID+"/databases", apiVersion)
	if err != nil {
		*issues = append(*issues, issueFromError("databases["+resourceGroup+"/"+serverName+"].databases", err))
		return nil
	}
	return databases
}

func databaseServerSummary(server map[string]any, databases []map[string]any, engine string) models.DatabaseServerAsset {
	serverID := mapStringValue(server, "id")
	properties := mapValue(server, "properties")
	identity := mapValue(server, "identity")
	network := mapValue(properties, "network")

	userDatabaseNames := visibleUserDatabaseNames(databases, engine)
	workloadIdentityIDs := sortedKeys(mapValue(identity, "userAssignedIdentities", "user_assigned_identities"))
	workloadIdentityType := stringPtr(mapStringValue(identity, "type"))
	workloadPrincipalID := stringPtr(mapStringValue(identity, "principalId", "principal_id"))
	workloadClientID := stringPtr(mapStringValue(identity, "clientId", "client_id"))
	fullyQualifiedDomainName := stringPtr(mapStringValue(properties, "fullyQualifiedDomainName", "fully_qualified_domain_name"))
	publicNetworkAccess := stringPtr(firstNonEmpty(
		mapStringValue(properties, "publicNetworkAccess", "public_network_access"),
		mapStringValue(network, "publicNetworkAccess", "public_network_access"),
	))
	minimalTLSVersion := stringPtr(mapStringValue(properties, "minimalTlsVersion", "minimal_tls_version"))
	serverVersion := stringPtr(mapStringValue(properties, "version"))
	highAvailabilityMode := stringPtr(databaseNormalizedEnum(mapStringValue(mapValue(properties, "highAvailability", "high_availability"), "mode")))
	delegatedSubnetResourceID := stringPtr(firstNonEmpty(
		mapStringValue(network, "delegatedSubnetResourceId", "delegated_subnet_resource_id"),
		mapStringValue(properties, "delegatedSubnetResourceId", "delegated_subnet_resource_id"),
	))
	privateDNSZoneResourceID := stringPtr(firstNonEmpty(
		mapStringValue(network, "privateDnsZoneArmResourceId", "private_dns_zone_arm_resource_id"),
		mapStringValue(network, "privateDnsZoneResourceId", "private_dns_zone_resource_id"),
		mapStringValue(properties, "privateDnsZoneArmResourceId", "private_dns_zone_arm_resource_id"),
		mapStringValue(properties, "privateDnsZoneResourceId", "private_dns_zone_resource_id"),
	))

	var databaseCount *int
	if databases != nil {
		count := len(userDatabaseNames)
		databaseCount = &count
	}

	serverName := firstNonEmpty(mapStringValue(server, "name"), resourceNameFromID(serverID), "unknown")
	return models.DatabaseServerAsset{
		DatabaseCount:             databaseCount,
		DelegatedSubnetResourceID: delegatedSubnetResourceID,
		Engine:                    engine,
		FullyQualifiedDomainName:  fullyQualifiedDomainName,
		HighAvailabilityMode:      highAvailabilityMode,
		ID:                        firstNonEmpty(serverID, "/unknown/"+serverName),
		Location:                  stringPtr(mapStringValue(server, "location")),
		MinimalTLSVersion:         minimalTLSVersion,
		Name:                      serverName,
		PrivateDNSZoneResourceID:  privateDNSZoneResourceID,
		PublicNetworkAccess:       publicNetworkAccess,
		RelatedIDs:                dedupeStrings(append([]string{serverID, stringPtrValue(workloadPrincipalID)}, workloadIdentityIDs...)),
		ResourceGroup:             resourceGroupFromID(serverID),
		ServerVersion:             serverVersion,
		State:                     stringPtr(mapStringValue(properties, "state")),
		Summary: databaseServerOperatorSummary(
			engine,
			serverName,
			fullyQualifiedDomainName,
			workloadIdentityType,
			publicNetworkAccess,
			minimalTLSVersion,
			serverVersion,
			highAvailabilityMode,
			delegatedSubnetResourceID,
			privateDNSZoneResourceID,
			databaseCount,
			userDatabaseNames,
		),
		UserDatabaseNames:    userDatabaseNames,
		WorkloadClientID:     workloadClientID,
		WorkloadIdentityIDs:  workloadIdentityIDs,
		WorkloadIdentityType: workloadIdentityType,
		WorkloadPrincipalID:  workloadPrincipalID,
	}
}

func visibleUserDatabaseNames(databases []map[string]any, engine string) []string {
	if databases == nil {
		return []string{}
	}

	names := []string{}
	systemNames := databaseSystemNames(engine)
	for _, database := range databases {
		name := mapStringValue(database, "name")
		if name == "" {
			continue
		}
		if _, blocked := systemNames[strings.ToLower(name)]; blocked {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return dedupeStrings(names)
}

func databaseSystemNames(engine string) map[string]struct{} {
	switch engine {
	case "AzureSql":
		return map[string]struct{}{"master": {}}
	case "PostgreSqlFlexible":
		return map[string]struct{}{"postgres": {}, "azure_maintenance": {}}
	case "MySqlFlexible":
		return map[string]struct{}{
			"mysql":              {},
			"information_schema": {},
			"performance_schema": {},
			"sys":                {},
		}
	default:
		return map[string]struct{}{}
	}
}

func databaseServerOperatorSummary(
	engine string,
	serverName string,
	fullyQualifiedDomainName *string,
	workloadIdentityType *string,
	publicNetworkAccess *string,
	minimalTLSVersion *string,
	serverVersion *string,
	highAvailabilityMode *string,
	delegatedSubnetResourceID *string,
	privateDNSZoneResourceID *string,
	databaseCount *int,
	userDatabaseNames []string,
) string {
	endpointPhrase := "does not expose a readable database endpoint from the current read path"
	if stringPtrValue(fullyQualifiedDomainName) != "" {
		endpointPhrase = "publishes endpoint '" + stringPtrValue(fullyQualifiedDomainName) + "'"
	}
	identityPhrase := "has no managed identity visible from the current read path"
	if stringPtrValue(workloadIdentityType) != "" {
		identityPhrase = "uses managed identity (" + stringPtrValue(workloadIdentityType) + ")"
	}

	inventoryParts := []string{}
	if databaseCount != nil {
		inventoryParts = append(inventoryParts, stringValue(*databaseCount)+" user database(s)")
	}
	if len(userDatabaseNames) > 0 {
		inventoryParts = append(inventoryParts, "names: "+strings.Join(userDatabaseNames, ", "))
	}

	postureParts := []string{"public network access " + firstNonEmpty(stringPtrValue(publicNetworkAccess), "unknown")}
	if stringPtrValue(minimalTLSVersion) != "" {
		postureParts = append(postureParts, "minimal TLS "+stringPtrValue(minimalTLSVersion))
	}
	if stringPtrValue(serverVersion) != "" {
		postureParts = append(postureParts, "server version "+stringPtrValue(serverVersion))
	}
	if stringPtrValue(highAvailabilityMode) != "" {
		postureParts = append(postureParts, "HA "+stringPtrValue(highAvailabilityMode))
	}
	if stringPtrValue(delegatedSubnetResourceID) != "" {
		postureParts = append(postureParts, "delegated subnet configured")
	}
	if stringPtrValue(privateDNSZoneResourceID) != "" {
		postureParts = append(postureParts, "private DNS configured")
	}

	inventoryPhrase := "Database inventory is not fully readable from the current read path."
	if len(inventoryParts) > 0 {
		inventoryPhrase = "Visible inventory: " + strings.Join(inventoryParts, ", ") + "."
	}

	return databaseEngineLabel(engine) + " server '" + serverName + "' " + endpointPhrase + " and " + identityPhrase + ". " + inventoryPhrase + " Visible posture: " + strings.Join(postureParts, ", ") + "."
}

func databaseEngineLabel(engine string) string {
	switch engine {
	case "AzureSql":
		return "Azure SQL"
	case "PostgreSqlFlexible":
		return "PostgreSQL Flexible"
	case "MySqlFlexible":
		return "MySQL Flexible"
	default:
		return engine
	}
}

func databaseNormalizedEnum(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), "_", "-"))
}
