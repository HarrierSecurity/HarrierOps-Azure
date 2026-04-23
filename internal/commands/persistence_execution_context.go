package commands

import "harrierops-azure/internal/models"

type persistencePrincipalEvidence struct {
	permissionsByPrincipal     map[string]models.PermissionRow
	assignmentsByPrincipal     map[string][]models.RoleAssignment
	currentIdentity            models.PermissionRow
	currentIdentityVisible     bool
	currentIdentityAssignments []models.RoleAssignment
}

func buildPersistencePrincipalEvidence(
	permissions []models.PermissionRow,
	assignments []models.RoleAssignment,
) persistencePrincipalEvidence {
	evidence := persistencePrincipalEvidence{
		permissionsByPrincipal:     make(map[string]models.PermissionRow, len(permissions)),
		assignmentsByPrincipal:     make(map[string][]models.RoleAssignment),
		currentIdentityAssignments: make([]models.RoleAssignment, 0),
	}

	for _, permission := range permissions {
		if permission.PrincipalID == "" {
			continue
		}
		evidence.permissionsByPrincipal[permission.PrincipalID] = permission
		if permission.IsCurrentIdentity && !evidence.currentIdentityVisible {
			evidence.currentIdentity = permission
			evidence.currentIdentityVisible = true
		}
	}

	for _, assignment := range assignments {
		if assignment.PrincipalID != "" {
			evidence.assignmentsByPrincipal[assignment.PrincipalID] = append(evidence.assignmentsByPrincipal[assignment.PrincipalID], assignment)
		}
		if evidence.currentIdentityVisible && assignment.PrincipalID == evidence.currentIdentity.PrincipalID {
			evidence.currentIdentityAssignments = append(evidence.currentIdentityAssignments, assignment)
		}
	}

	return evidence
}

type persistencePrincipalRoleContextOptions struct {
	fallbackName                string
	kind                        string
	principalID                 *string
	identityType                *string
	permissionsByPrincipal      map[string]models.PermissionRow
	assignmentsByPrincipal      map[string][]models.RoleAssignment
	resolvedSummary             func(name string, roleSummary string) string
	lowerImpactSummary          func(name string) string
	unresolvedPrivilegedSummary func(name string, roleSummary string) string
	noAssignmentsSummary        func(name string) string
	rbacOnlyCarriesAzureControl bool
}

func persistencePrincipalRoleContext(
	opts persistencePrincipalRoleContextOptions,
) (*models.PersistenceRoleContext, bool, bool) {
	if opts.principalID == nil || stringPtrValue(opts.principalID) == "" {
		return nil, false, false
	}

	name := opts.fallbackName
	if permission, ok := opts.permissionsByPrincipal[*opts.principalID]; ok {
		name = firstNonEmpty(permission.DisplayName, name)
		roleNames := append([]string{}, permission.HighImpactRoles...)
		if len(roleNames) == 0 {
			roleNames = append(roleNames, permission.AllRoleNames...)
		}
		summary := opts.resolvedSummary(name, persistenceRoleSummary(roleNames, permission.ScopeIDs))
		if !permission.Privileged {
			summary = opts.lowerImpactSummary(name)
		}
		return &models.PersistenceRoleContext{
			Name:         name,
			Kind:         opts.kind,
			PrincipalID:  opts.principalID,
			IdentityType: opts.identityType,
			RoleNames:    dedupeStrings(roleNames),
			ScopeIDs:     dedupeStrings(permission.ScopeIDs),
			Summary:      summary,
		}, permission.Privileged, true
	}

	assignments := opts.assignmentsByPrincipal[*opts.principalID]
	roleNames, scopeIDs, privileged := persistenceFunctionAssignmentsRoleContext(assignments)
	summary := opts.noAssignmentsSummary(name)
	switch {
	case len(assignments) == 0:
		summary = opts.noAssignmentsSummary(name)
	case !privileged:
		summary = opts.lowerImpactSummary(name)
	default:
		summary = opts.unresolvedPrivilegedSummary(name, persistenceRoleSummary(roleNames, scopeIDs))
	}

	return &models.PersistenceRoleContext{
		Name:         name,
		Kind:         opts.kind,
		PrincipalID:  opts.principalID,
		IdentityType: opts.identityType,
		RoleNames:    dedupeStrings(roleNames),
		ScopeIDs:     dedupeStrings(scopeIDs),
		Summary:      summary,
	}, opts.rbacOnlyCarriesAzureControl && privileged, true
}
