package providers

import (
	"context"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

const (
	privescCurrentFootholdDirectControl = "current-foothold-direct-control"
	privescVisiblePrivilegedLead        = "visible-privileged-lead"
	privescIngressBackedWorkloadID      = "ingress-backed-workload-identity"
)

func (p StaticProvider) Privesc(ctx context.Context, tenant string, subscription string) (PrivescFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	return PrivescFacts{
		TenantID:       session.TenantID,
		SubscriptionID: session.Subscription.ID,
		Paths: []models.PrivescPathSummary{
			{
				Asset:            nil,
				CurrentIdentity:  true,
				ImpactRoles:      []string{"Owner"},
				StartingFoothold: "azurefox-lab-sp (current foothold)",
				MissingProof:     "HO-Azure does not prove which exact abuse action is the best next step from this row alone.",
				NextReview:       "Check rbac for the exact assignment evidence and scope behind this current-identity escalation lead.",
				OperatorSignal:   "Current foothold already has direct control.",
				PathType:         privescCurrentFootholdDirectControl,
				Priority:         "high",
				Principal:        "azurefox-lab-sp",
				PrincipalID:      "33333333-3333-3333-3333-333333333333",
				PrincipalType:    "ServicePrincipal",
				ProvenPath:       "Current foothold 'azurefox-lab-sp' already holds high-impact RBAC (Owner) on visible scope.",
				RelatedIDs: []string{
					"33333333-3333-3333-3333-333333333333",
					"/subscriptions/22222222-2222-2222-2222-222222222222",
				},
				Summary: "Current foothold 'azurefox-lab-sp' already holds high-impact RBAC (Owner) on visible scope. HO-Azure does not prove which exact abuse action is the best next step from this row alone. Check rbac for the exact assignment evidence and scope behind this current-identity escalation lead.",
			},
			{
				Asset:            models.StringPtr("vm-web-01"),
				CurrentIdentity:  false,
				ImpactRoles:      []string{"Owner"},
				StartingFoothold: "azurefox-lab-sp (current foothold)",
				MissingProof:     "HO-Azure does not prove control of the workload or successful token use from it.",
				NextReview:       "Check managed-identities for the workload-to-identity anchor behind this ingress-backed lead.",
				OperatorSignal:   "Visible ingress-backed lead; not yet rooted in current foothold.",
				PathType:         privescIngressBackedWorkloadID,
				Priority:         "medium",
				Principal:        "ua-app",
				PrincipalID:      "33333333-3333-3333-3333-333333333333",
				PrincipalType:    "ManagedIdentity",
				ProvenPath:       "Public workload 'vm-web-01' carries identity 'ua-app' with high-impact RBAC (Owner).",
				RelatedIDs: []string{
					"/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-workload/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
					"33333333-3333-3333-3333-333333333333",
					"/subscriptions/22222222-2222-2222-2222-222222222222/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
					"/subscriptions/22222222-2222-2222-2222-222222222222",
				},
				Summary: "Public workload 'vm-web-01' carries identity 'ua-app' with high-impact RBAC (Owner). HO-Azure does not prove control of the workload or successful token use from it. Check managed-identities for the workload-to-identity anchor behind this ingress-backed lead.",
			},
		},
		Issues: []models.Issue{},
	}, nil
}

func (p AzureProvider) Privesc(ctx context.Context, tenant string, subscription string) (PrivescFacts, error) {
	permissionsFacts, err := p.Permissions(ctx, tenant, subscription)
	if err != nil {
		return PrivescFacts{}, err
	}
	principalsFacts, err := p.Principals(ctx, tenant, subscription)
	if err != nil {
		return PrivescFacts{}, err
	}
	managedIdentityFacts, err := p.ManagedIdentities(ctx, tenant, subscription)
	if err != nil {
		return PrivescFacts{}, err
	}
	vmFacts, err := p.VMs(ctx, tenant, subscription)
	if err != nil {
		return PrivescFacts{}, err
	}
	return privescFactsFromSources(permissionsFacts, principalsFacts, managedIdentityFacts, vmFacts), nil
}

func privescFactsFromSources(
	permissionsFacts PermissionsFacts,
	principalsFacts PrincipalsFacts,
	managedIdentityFacts ManagedIdentitiesFacts,
	vmFacts VMsFacts,
) PrivescFacts {
	principalByID := make(map[string]models.PrincipalSummary, len(principalsFacts.Principals))
	currentFootholdLabel := ""
	for _, principal := range principalsFacts.Principals {
		principalByID[principal.ID] = principal
		if principal.IsCurrentIdentity && currentFootholdLabel == "" {
			currentFootholdLabel = firstNonEmpty(stringValue(principal.DisplayName), principal.ID)
		}
	}

	identitiesByPrincipal := map[string][]models.ManagedIdentity{}
	for _, identity := range managedIdentityFacts.Identities {
		if identity.PrincipalID == nil || strings.TrimSpace(*identity.PrincipalID) == "" {
			continue
		}
		principalID := strings.TrimSpace(*identity.PrincipalID)
		identitiesByPrincipal[principalID] = append(identitiesByPrincipal[principalID], identity)
	}

	vmByID := make(map[string]models.VmAsset, len(vmFacts.VMAssets))
	for _, vm := range vmFacts.VMAssets {
		vmByID[vm.ID] = vm
	}

	paths := make([]models.PrivescPathSummary, 0, len(permissionsFacts.Permissions))
	for _, permission := range permissionsFacts.Permissions {
		if !permission.Privileged {
			continue
		}

		principalName := firstNonEmpty(permission.DisplayName, permission.PrincipalID, "unknown")
		impactRoles := append([]string{}, permission.HighImpactRoles...)
		principalID := firstNonEmpty(permission.PrincipalID, "unknown")
		currentIdentity := permission.IsCurrentIdentity
		pathType := privescPathType("direct-role-abuse", currentIdentity)
		startingFoothold := privescStartingFoothold(currentIdentity, principalName, currentFootholdLabel)
		operatorSignal := privescOperatorSignal("direct-role-abuse", currentIdentity)
		provenPath := privescProvenPath(principalName, "direct-role-abuse", "", impactRoles, currentIdentity)
		missingProof := privescMissingProof("direct-role-abuse", currentIdentity)
		nextReview := privescNextReviewHint("direct-role-abuse", currentIdentity)

		paths = append(paths, models.PrivescPathSummary{
			Asset:            nil,
			CurrentIdentity:  currentIdentity,
			ImpactRoles:      impactRoles,
			StartingFoothold: startingFoothold,
			MissingProof:     missingProof,
			NextReview:       nextReview,
			OperatorSignal:   operatorSignal,
			PathType:         pathType,
			Priority:         privescPriority(currentIdentity),
			Principal:        principalName,
			PrincipalID:      principalID,
			PrincipalType:    firstNonEmpty(permission.PrincipalType, "unknown"),
			ProvenPath:       provenPath,
			RelatedIDs:       append([]string{principalID}, permission.ScopeIDs...),
			Summary:          privescSummary(provenPath, missingProof, nextReview),
		})

		for _, identity := range identitiesByPrincipal[principalID] {
			for _, attachedID := range identity.AttachedTo {
				vmAsset, ok := vmByID[attachedID]
				if !ok || len(vmAsset.PublicIPs) == 0 {
					continue
				}

				identityName := firstNonEmpty(identity.Name, principalName)
				assetName := firstNonEmpty(vmAsset.Name, attachedID)
				pathType = privescPathType("public-identity-pivot", false)
				startingFoothold = privescStartingFoothold(false, identityName, currentFootholdLabel)
				operatorSignal = privescOperatorSignal("public-identity-pivot", false)
				provenPath = privescProvenPath(identityName, "public-identity-pivot", assetName, impactRoles, false)
				missingProof = privescMissingProof("public-identity-pivot", false)
				nextReview = privescNextReviewHint("public-identity-pivot", false)

				relatedIDs := []string{identity.ID, principalID, attachedID}
				if principal, ok := principalByID[principalID]; ok {
					relatedIDs = append(relatedIDs, principal.ScopeIDs...)
				}

				paths = append(paths, models.PrivescPathSummary{
					Asset:            models.StringPtr(assetName),
					CurrentIdentity:  false,
					ImpactRoles:      impactRoles,
					StartingFoothold: startingFoothold,
					MissingProof:     missingProof,
					NextReview:       nextReview,
					OperatorSignal:   operatorSignal,
					PathType:         pathType,
					Priority:         "medium",
					Principal:        identityName,
					PrincipalID:      principalID,
					PrincipalType:    "ManagedIdentity",
					ProvenPath:       provenPath,
					RelatedIDs:       relatedIDs,
					Summary:          privescSummary(provenPath, missingProof, nextReview),
				})
			}
		}
	}

	sort.SliceStable(paths, func(i int, j int) bool {
		left := paths[i]
		right := paths[j]
		leftKey := []int{
			privescPriorityRank(left.Priority),
			privescCurrentIdentityRank(left.CurrentIdentity),
			privescPathSortRank(left.PathType),
		}
		rightKey := []int{
			privescPriorityRank(right.Priority),
			privescCurrentIdentityRank(right.CurrentIdentity),
			privescPathSortRank(right.PathType),
		}
		for index := range leftKey {
			if leftKey[index] != rightKey[index] {
				return leftKey[index] < rightKey[index]
			}
		}
		if left.Principal != right.Principal {
			return left.Principal < right.Principal
		}
		return valueOrEmptyString(left.Asset) < valueOrEmptyString(right.Asset)
	})

	issues := append([]models.Issue{}, permissionsFacts.Issues...)
	issues = append(issues, managedIdentityFacts.Issues...)
	issues = append(issues, vmFacts.Issues...)

	return PrivescFacts{
		TenantID:       firstNonEmpty(permissionsFacts.TenantID, principalsFacts.TenantID, managedIdentityFacts.TenantID, vmFacts.TenantID),
		SubscriptionID: firstNonEmpty(permissionsFacts.SubscriptionID, principalsFacts.SubscriptionID, managedIdentityFacts.SubscriptionID, vmFacts.SubscriptionID),
		Paths:          paths,
		Issues:         issues,
	}
}

func privescPathType(rawPathType string, currentIdentity bool) string {
	if rawPathType == "public-identity-pivot" {
		return privescIngressBackedWorkloadID
	}
	if currentIdentity {
		return privescCurrentFootholdDirectControl
	}
	return privescVisiblePrivilegedLead
}

func privescOperatorSignal(rawPathType string, currentIdentity bool) string {
	if rawPathType == "public-identity-pivot" {
		if currentIdentity {
			return "Current foothold already reaches an ingress-backed workload identity path."
		}
		return "Visible ingress-backed lead; not yet rooted in current foothold."
	}
	if currentIdentity {
		return "Current foothold already has direct control."
	}
	return "Visible privileged lead; not yet rooted in current foothold."
}

func privescProvenPath(principalName string, rawPathType string, assetName string, impactRoles []string, currentIdentity bool) string {
	roleText := strings.Join(impactRoles, ", ")
	if roleText == "" {
		roleText = "high-impact roles"
	}

	if rawPathType == "public-identity-pivot" {
		asset := firstNonEmpty(assetName, "visible workload")
		return "Public workload '" + asset + "' carries identity '" + principalName + "' with high-impact RBAC (" + roleText + ")."
	}
	if currentIdentity {
		return "Current foothold '" + principalName + "' already holds high-impact RBAC (" + roleText + ") on visible scope."
	}
	return "Visible principal '" + principalName + "' already holds high-impact RBAC (" + roleText + ") on visible scope."
}

func privescMissingProof(rawPathType string, currentIdentity bool) string {
	if rawPathType == "public-identity-pivot" {
		return "HO-Azure does not prove control of the workload or successful token use from it."
	}
	if currentIdentity {
		return "HO-Azure does not prove which exact abuse action is the best next step from this row alone."
	}
	return "HO-Azure does not prove the current identity can act as or control this principal."
}

func privescNextReviewHint(rawPathType string, currentIdentity bool) string {
	if rawPathType == "public-identity-pivot" {
		return "Check managed-identities for the workload-to-identity anchor behind this ingress-backed lead."
	}
	if currentIdentity {
		return "Check rbac for the exact assignment evidence and scope behind this current-identity escalation lead."
	}
	return "Check role-trusts for paths that could let the current identity influence this privileged principal."
}

func privescSummary(provenPath string, missingProof string, nextReview string) string {
	return strings.TrimSpace(provenPath + " " + missingProof + " " + nextReview)
}

func privescStartingFoothold(currentIdentity bool, principalName string, currentFootholdLabel string) string {
	if currentIdentity {
		return principalName + " (current foothold)"
	}
	if currentFootholdLabel != "" {
		return currentFootholdLabel + " (current foothold)"
	}
	return "unknown current foothold"
}

func privescPriority(currentIdentity bool) string {
	if currentIdentity {
		return "high"
	}
	return "medium"
}

func privescPriorityRank(priority string) int {
	switch strings.ToLower(strings.TrimSpace(priority)) {
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

func privescPathSortRank(pathType string) int {
	switch strings.TrimSpace(pathType) {
	case privescIngressBackedWorkloadID:
		return 0
	case privescVisiblePrivilegedLead:
		return 1
	case privescCurrentFootholdDirectControl:
		return 2
	default:
		return 9
	}
}

func valueOrEmptyString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func privescCurrentIdentityRank(currentIdentity bool) int {
	if currentIdentity {
		return 0
	}
	return 1
}
