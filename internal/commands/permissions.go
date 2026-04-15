package commands

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func permissionsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Permissions(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		subscriptionID := request.Subscription
		if subscriptionID == "" {
			subscriptionID = facts.SubscriptionID
		}

		return models.PermissionsOutput{
			Metadata:    scopedMetadata(now, request, facts.TenantID, subscriptionID, "permissions"),
			Permissions: enrichPermissionRows(facts.Permissions, facts.Principals),
			Issues:      facts.Issues,
		}, nil
	}
}

type enrichedPermissionRow struct {
	row               models.PermissionRow
	workloadPivotRank int
}

func enrichPermissionRows(permissions []providers.PermissionFact, principals []providers.PermissionPrincipalFact) []models.PermissionRow {
	principalsByID := make(map[string]providers.PermissionPrincipalFact, len(principals))
	for _, principal := range principals {
		principalsByID[principal.ID] = principal
	}

	enriched := make([]enrichedPermissionRow, 0, len(permissions))
	for _, permission := range permissions {
		principal := principalsByID[permission.PrincipalID]
		hasWorkloadPivot := len(principal.AttachedTo) > 0
		workloadVisibilityBlocked := !hasWorkloadPivot && (len(principal.IdentityNames) > 0 || contains(principal.Sources, "managed-identities"))
		workloadPivotRank := 9
		if len(principal.IdentityNames) > 0 && hasWorkloadPivot {
			workloadPivotRank = 0
		} else if hasWorkloadPivot {
			workloadPivotRank = 1
		}
		trustExpansionFollowOn := permission.Privileged &&
			!permission.IsCurrentIdentity &&
			!hasWorkloadPivot &&
			!workloadVisibilityBlocked &&
			(strings.EqualFold(permission.PrincipalType, "ServicePrincipal") || contains(principal.Sources, "managed-identities"))

		nextReview := permissionsNextReviewHint(
			permission.Privileged,
			permission.IsCurrentIdentity,
			hasWorkloadPivot,
			workloadVisibilityBlocked,
			trustExpansionFollowOn,
		)

		enriched = append(enriched, enrichedPermissionRow{
			row: models.PermissionRow{
				PrincipalID:         permission.PrincipalID,
				DisplayName:         permission.DisplayName,
				PrincipalType:       permission.PrincipalType,
				Priority:            permissionsPriority(permission.Privileged, permission.IsCurrentIdentity, hasWorkloadPivot),
				HighImpactRoles:     permission.HighImpactRoles,
				AllRoleNames:        permission.AllRoleNames,
				RoleAssignmentCount: permission.RoleAssignmentCount,
				ScopeCount:          permission.ScopeCount,
				ScopeIDs:            permission.ScopeIDs,
				Privileged:          permission.Privileged,
				IsCurrentIdentity:   permission.IsCurrentIdentity,
				OperatorSignal: permissionsOperatorSignal(
					permission.Privileged,
					permission.IsCurrentIdentity,
					hasWorkloadPivot,
					workloadVisibilityBlocked,
					trustExpansionFollowOn,
				),
				NextReview: nextReview,
				Summary: permissionsSummary(
					permission.DisplayName,
					permission.PrincipalType,
					permission.HighImpactRoles,
					permission.ScopeCount,
					permission.Privileged,
					permission.IsCurrentIdentity,
					hasWorkloadPivot,
					workloadVisibilityBlocked,
					trustExpansionFollowOn,
					nextReview,
				),
			},
			workloadPivotRank: workloadPivotRank,
		})
	}

	sort.SliceStable(enriched, func(i int, j int) bool {
		left := enriched[i]
		right := enriched[j]

		leftKey := []int{
			permissionPriorityRank(left.row.Priority),
			boolRank(!left.row.Privileged),
			permissionFollowOnRank(left.row.NextReview),
			left.workloadPivotRank,
			permissionRoleRank(left.row.HighImpactRoles),
			-left.row.ScopeCount,
			-left.row.RoleAssignmentCount,
		}
		rightKey := []int{
			permissionPriorityRank(right.row.Priority),
			boolRank(!right.row.Privileged),
			permissionFollowOnRank(right.row.NextReview),
			right.workloadPivotRank,
			permissionRoleRank(right.row.HighImpactRoles),
			-right.row.ScopeCount,
			-right.row.RoleAssignmentCount,
		}

		for index := range leftKey {
			if leftKey[index] != rightKey[index] {
				return leftKey[index] < rightKey[index]
			}
		}
		if left.row.DisplayName != right.row.DisplayName {
			return left.row.DisplayName < right.row.DisplayName
		}
		return left.row.PrincipalID < right.row.PrincipalID
	})

	rows := make([]models.PermissionRow, 0, len(enriched))
	for _, item := range enriched {
		rows = append(rows, item.row)
	}
	return rows
}

func permissionsOperatorSignal(privileged bool, isCurrentIdentity bool, hasWorkloadPivot bool, workloadVisibilityBlocked bool, trustExpansionFollowOn bool) string {
	if !privileged {
		return "Direct control not confirmed."
	}
	if isCurrentIdentity {
		return "Direct control visible; current foothold."
	}
	if hasWorkloadPivot {
		return "Direct control visible; workload pivot visible."
	}
	if workloadVisibilityBlocked {
		return "Direct control visible; visibility blocked."
	}
	if trustExpansionFollowOn {
		return "Direct control visible; trust expansion follow-on."
	}
	return "Direct control visible; exact assignment review next."
}

func permissionsNextReviewHint(privileged bool, isCurrentIdentity bool, hasWorkloadPivot bool, workloadVisibilityBlocked bool, trustExpansionFollowOn bool) string {
	if !privileged {
		return "Check rbac for the exact assignment evidence behind this lower-signal row."
	}
	if isCurrentIdentity {
		return "Check privesc for the direct abuse or escalation path behind this current identity."
	}
	if hasWorkloadPivot {
		return "Check managed-identities for the workload pivot behind this direct control row."
	}
	if workloadVisibilityBlocked {
		return "Check managed-identities; current scope does not yet show the workload pivot behind this direct-control row."
	}
	if trustExpansionFollowOn {
		return "Check role-trusts for trust expansion around who can influence this principal."
	}
	return "Check rbac for the exact assignment scope behind this direct-control row."
}

func permissionsPriority(privileged bool, isCurrentIdentity bool, hasWorkloadPivot bool) string {
	if !privileged {
		return "low"
	}
	if isCurrentIdentity || hasWorkloadPivot {
		return "high"
	}
	return "medium"
}

func permissionsSummary(
	principalName string,
	principalType string,
	highImpactRoles []string,
	scopeCount int,
	privileged bool,
	isCurrentIdentity bool,
	hasWorkloadPivot bool,
	workloadVisibilityBlocked bool,
	trustExpansionFollowOn bool,
	nextReview string,
) string {
	if !privileged {
		return "Principal '" + principalName + "' does not yet show direct control from visible RBAC. " + nextReview
	}

	roleText := strings.Join(highImpactRoles, ", ")
	if roleText == "" {
		roleText = "high-impact roles"
	}

	scopeText := "subscription-wide"
	if scopeCount > 1 {
		scopeText = strconv.Itoa(scopeCount) + " visible scopes"
	}

	if isCurrentIdentity {
		return "Current identity '" + principalName + "' already has direct control visible through " + roleText + " across " + scopeText + ". " + nextReview
	}
	if hasWorkloadPivot {
		return principalType + " '" + principalName + "' already has direct control visible through " + roleText + " across " + scopeText + ", and current scope also shows a workload pivot. " + nextReview
	}
	if workloadVisibilityBlocked {
		return principalType + " '" + principalName + "' already has direct control visible through " + roleText + " across " + scopeText + ", but the backing workload pivot stays visibility blocked from current scope. " + nextReview
	}
	if trustExpansionFollowOn {
		return principalType + " '" + principalName + "' already has direct control visible through " + roleText + " across " + scopeText + ". The next useful question is trust expansion, not more privilege ranking. " + nextReview
	}
	return principalType + " '" + principalName + "' already has direct control visible through " + roleText + " across " + scopeText + ". " + nextReview
}

func permissionPriorityRank(priority string) int {
	switch strings.ToLower(priority) {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	default:
		return 9
	}
}

func permissionFollowOnRank(nextReview string) int {
	text := strings.ToLower(nextReview)
	switch {
	case strings.Contains(text, "check privesc"):
		return 0
	case strings.Contains(text, "check managed-identities"):
		return 1
	case strings.Contains(text, "check role-trusts"):
		return 2
	default:
		return 3
	}
}

func permissionRoleRank(highImpactRoles []string) int {
	roles := make([]string, 0, len(highImpactRoles))
	for _, role := range highImpactRoles {
		roles = append(roles, strings.ToLower(role))
	}
	switch {
	case contains(roles, "owner"):
		return 0
	case contains(roles, "user access administrator"):
		return 1
	case contains(roles, "contributor"):
		return 2
	default:
		return 9
	}
}

func boolRank(value bool) int {
	if value {
		return 1
	}
	return 0
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if strings.EqualFold(value, needle) {
			return true
		}
	}
	return false
}
