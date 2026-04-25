package providers

import (
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

type principalRecord struct {
	id                  string
	principalType       string
	displayName         string
	tenantID            string
	sources             []string
	scopeIDs            []string
	roleNames           []string
	roleAssignmentCount int
	identityNames       []string
	identityTypes       []string
	attachedTo          []string
	isCurrentIdentity   bool
}

func PrincipalsFactsFromSources(
	tenantID string,
	subscriptionID string,
	rbacFacts RBACFacts,
	whoamiFacts WhoAmIFacts,
	managedIdentityFacts ManagedIdentitiesFacts,
) PrincipalsFacts {
	records := map[string]principalRecord{}
	issues := append([]models.Issue{}, rbacFacts.Issues...)
	issues = append(issues, whoamiFacts.Issues...)
	issues = append(issues, managedIdentityFacts.Issues...)

	ensureRecord := func(principalID string) principalRecord {
		if record, ok := records[principalID]; ok {
			return record
		}
		return principalRecord{
			id:            principalID,
			principalType: "unknown",
			tenantID:      tenantID,
		}
	}

	for _, principal := range rbacFacts.Principals {
		if principal.ID == "" {
			continue
		}
		record := ensureRecord(principal.ID)
		record.principalType = normalizePrincipalType(record.principalType, principal.PrincipalType)
		if record.displayName == "" {
			record.displayName = principal.DisplayName
		}
		if record.tenantID == "" {
			record.tenantID = principal.TenantID
		}
		record.sources = appendUniqueString(record.sources, "rbac")
		records[principal.ID] = record
	}

	for _, assignment := range rbacFacts.RoleAssignments {
		if assignment.PrincipalID == "" {
			continue
		}
		record := ensureRecord(assignment.PrincipalID)
		record.principalType = normalizePrincipalType(record.principalType, assignment.PrincipalType)
		record.roleNames = appendUniqueString(record.roleNames, assignment.RoleName)
		record.scopeIDs = appendUniqueString(record.scopeIDs, assignment.ScopeID)
		record.roleAssignmentCount++
		record.sources = appendUniqueString(record.sources, "rbac")
		records[assignment.PrincipalID] = record
	}

	if whoamiFacts.Principal.ID != "" {
		record := ensureRecord(whoamiFacts.Principal.ID)
		record.principalType = normalizePrincipalType(record.principalType, whoamiFacts.Principal.PrincipalType)
		if record.displayName == "" {
			record.displayName = whoamiFacts.Principal.DisplayName
		}
		if record.tenantID == "" {
			record.tenantID = whoamiFacts.Principal.TenantID
		}
		record.isCurrentIdentity = true
		record.sources = appendUniqueString(record.sources, "whoami")
		for _, scope := range whoamiFacts.EffectiveScopes {
			record.scopeIDs = appendUniqueString(record.scopeIDs, scope.ID)
		}
		records[whoamiFacts.Principal.ID] = record
	}

	for _, identity := range managedIdentityFacts.Identities {
		if identity.PrincipalID == nil || *identity.PrincipalID == "" {
			continue
		}
		record := ensureRecord(*identity.PrincipalID)
		record.principalType = normalizePrincipalType(record.principalType, "ManagedIdentity")
		if record.displayName == "" {
			record.displayName = identity.Name
		}
		record.identityNames = appendUniqueString(record.identityNames, identity.Name)
		record.identityTypes = appendUniqueString(record.identityTypes, identity.IdentityType)
		for _, scopeID := range identity.ScopeIDs {
			record.scopeIDs = appendUniqueString(record.scopeIDs, scopeID)
		}
		for _, attachedID := range identity.AttachedTo {
			record.attachedTo = appendUniqueString(record.attachedTo, attachedID)
		}
		record.sources = appendUniqueString(record.sources, "managed-identities")
		records[*identity.PrincipalID] = record
	}

	principals := make([]models.PrincipalSummary, 0, len(records))
	for _, record := range records {
		principals = append(principals, models.PrincipalSummary{
			AttachedTo:          sortedUniqueStrings(record.attachedTo),
			DisplayName:         models.StringPtr(record.displayName),
			ID:                  record.id,
			IdentityNames:       sortedUniqueStrings(record.identityNames),
			IdentityTypes:       sortedUniqueStrings(record.identityTypes),
			IsCurrentIdentity:   record.isCurrentIdentity,
			PrincipalType:       firstNonEmpty(record.principalType, "unknown"),
			RoleAssignmentCount: record.roleAssignmentCount,
			RoleNames:           sortedUniqueStrings(record.roleNames),
			ScopeIDs:            sortedUniqueStrings(record.scopeIDs),
			Sources:             sortedUniqueStrings(record.sources),
			TenantID:            models.StringPtr(record.tenantID),
		})
	}

	sort.SliceStable(principals, func(i int, j int) bool {
		left := principals[i]
		right := principals[j]
		switch {
		case principalHasHighImpactRolesByName(left.RoleNames) != principalHasHighImpactRolesByName(right.RoleNames):
			return principalHasHighImpactRolesByName(left.RoleNames)
		case len(left.AttachedTo) > 0 && len(right.AttachedTo) == 0:
			return true
		case len(left.AttachedTo) == 0 && len(right.AttachedTo) > 0:
			return false
		case len(left.ScopeIDs) != len(right.ScopeIDs):
			return len(left.ScopeIDs) > len(right.ScopeIDs)
		case left.RoleAssignmentCount != right.RoleAssignmentCount:
			return left.RoleAssignmentCount > right.RoleAssignmentCount
		case stringPtrValue(left.DisplayName) != stringPtrValue(right.DisplayName):
			return stringPtrValue(left.DisplayName) < stringPtrValue(right.DisplayName)
		default:
			return left.ID < right.ID
		}
	})

	return PrincipalsFacts{
		TenantID:         tenantID,
		SubscriptionID:   subscriptionID,
		CurrentPrincipal: whoamiFacts.Principal,
		TokenSource:      whoamiFacts.TokenSource,
		AuthMode:         whoamiFacts.AuthMode,
		Principals:       principals,
		Issues:           issues,
	}
}

func principalHasHighImpactRolesByName(roleNames []string) bool {
	for _, roleName := range roleNames {
		if _, ok := highImpactRoleNames[normalizeRoleName(roleName)]; ok {
			return true
		}
	}
	return false
}

func normalizeRoleName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
