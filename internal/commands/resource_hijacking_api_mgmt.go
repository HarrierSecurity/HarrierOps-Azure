package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

var resourceHijackingAPIMSteps = []familyStepDefinition{
	{Action: "select trusted gateway", APISurface: "Microsoft.ApiManagement/service", DownstreamEffect: "Keeps clients on the existing APIM hostname and API surface.", Boundary: "Gateway posture does not prove traffic volume."},
	{Action: "identify backend control point", APISurface: "Microsoft.ApiManagement/service/backends", DownstreamEffect: "Shows where APIM can forward requests behind the stable front door.", Boundary: "Backend hostnames do not prove ownership or runtime reachability."},
	{Action: "change backend or routing policy", APISurface: "APIM backend or policy write", NeedsWrite: true, DownstreamEffect: "Can redirect selected API traffic while the published APIM surface remains healthy.", Boundary: "This command does not collect policy XML bodies by default."},
	{Action: "preserve subscriptions and named values", APISurface: "APIM subscriptions and named values", NeedsWrite: true, DownstreamEffect: "Existing consumers, subscription gates, and stored config can keep the route looking operational.", Boundary: "Named-value values and secrets are not printed."},
	{Action: "blend as API operations change", APISurface: "APIM service configuration", DownstreamEffect: "Normal cover stories include backend migration, failover, version routing, and blue/green release work.", Boundary: "Cover story is not an intent claim."},
}

func buildResourceHijackingAPIMOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.ResourceHijackingSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	apiMgmtFuture := runGroupedCommandOutput[models.ApiMgmtOutput](group, ctx, request, apiMgmtHandler(provider, now), "api-mgmt")
	evidenceFutures := runFamilyEvidence(group, ctx, request, provider, now)

	apiMgmt, err := apiMgmtFuture.wait()
	if err != nil {
		return nil, err
	}
	evidence, err := evidenceFutures.wait()
	if err != nil {
		return nil, err
	}

	targets := make([]models.ResourceHijackingAPIMTarget, 0, len(apiMgmt.ApiManagementServices))
	for _, service := range apiMgmt.ApiManagementServices {
		control, controlOK := resourceHijackingAPIMControl(service.ID, evidence.principal.currentIdentityAssignments)
		rank, reason := resourceHijackingAPIMTakeoverRank(service, controlOK)
		targets = append(targets, models.ResourceHijackingAPIMTarget{
			ID:                     service.ID,
			Name:                   service.Name,
			ResourceGroup:          service.ResourceGroup,
			Location:               service.Location,
			TakeoverRank:           rank,
			TakeoverReason:         reason,
			CapabilitySteps:        resourceHijackingAPIMCapabilitySteps(controlOK),
			CurrentIdentityContext: resourceHijackingRoleContext(evidence.principal.currentIdentity, control, controlOK, "APIM backend or policy write control", "APIM backend/policy write"),
			CurrentState:           resourceHijackingAPIMState(service),
			NotCollectedByDefault:  resourceHijackingAPIMNotCollectedByDefault(),
			Summary:                resourceHijackingAPIMSummary(service, rank, controlOK),
			RelatedIDs:             mergeRelatedIDs(service.RelatedIDs),
		})
	}
	sort.SliceStable(targets, func(i, j int) bool {
		if targets[i].TakeoverRank != targets[j].TakeoverRank {
			return targets[i].TakeoverRank > targets[j].TakeoverRank
		}
		return targets[i].Name < targets[j].Name
	})

	issues := familyIssues(apiMgmt.Issues, evidence)

	return models.ResourceHijackingAPIMOutput{
		Metadata: scopedMetadata(
			now,
			request,
			firstNonEmpty(request.Tenant, stringPtrValue(apiMgmt.Metadata.TenantID), stringPtrValue(evidence.permissions.Metadata.TenantID)),
			firstNonEmpty(request.Subscription, stringPtrValue(apiMgmt.Metadata.SubscriptionID), stringPtrValue(evidence.permissions.Metadata.SubscriptionID)),
			"resourcehijacking",
		),
		GroupedCommandName: "resourcehijacking",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		Targets:            targets,
		Issues:             issues,
	}, nil
}

func resourceHijackingAPIMControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	actions := []string{
		"Microsoft.ApiManagement/service/write",
		"Microsoft.ApiManagement/service/backends/write",
		"Microsoft.ApiManagement/service/apis/policies/write",
		"Microsoft.ApiManagement/service/policies/write",
	}
	return resourceHijackingBestControl(resourceID, assignments, actions, "Owner", "Contributor", "API Management Service Contributor")
}

func resourceHijackingBestControl(resourceID string, assignments []models.RoleAssignment, actions []string, roleNames ...string) (persistenceCurrentIdentityControl, bool) {
	bestRank := 99
	best := persistenceCurrentIdentityControl{}
	for _, assignment := range assignments {
		allowed := false
		for _, action := range actions {
			if persistenceRoleAssignmentAllowsNamedOrActionControl(assignment, action, roleNames...) {
				allowed = true
				break
			}
		}
		if !allowed {
			continue
		}
		rank, ok := persistenceScopeRank(assignment.ScopeID, resourceID)
		if !ok || rank >= bestRank {
			continue
		}
		bestRank = rank
		best = persistenceCurrentIdentityControl{
			RoleName: fmt.Sprintf("%s at %s", assignment.RoleName, persistenceScopeLabel(assignment.ScopeID)),
			ScopeID:  assignment.ScopeID,
		}
	}
	return best, bestRank != 99
}

func resourceHijackingRoleContext(
	currentIdentity models.PermissionRow,
	control persistenceCurrentIdentityControl,
	controlOK bool,
	controlName string,
	controlLabel string,
) *models.ResourceHijackingRoleContext {
	if strings.TrimSpace(currentIdentity.DisplayName) == "" && !controlOK {
		return nil
	}
	name := firstNonEmpty(currentIdentity.DisplayName, "current identity")
	roleNames := append([]string{}, currentIdentity.HighImpactRoles...)
	if len(roleNames) == 0 {
		roleNames = append(roleNames, currentIdentity.AllRoleNames...)
	}
	scopeIDs := append([]string{}, currentIdentity.ScopeIDs...)
	summary := "Current foothold identity is visible, but " + controlName + " is not proven here."
	label := "not proven"
	if controlOK {
		summary = fmt.Sprintf("Current foothold `%s` has visible %s.", name, controlName)
		roleNames = []string{evasionControlRoleName(control)}
		scopeIDs = []string{control.ScopeID}
		label = controlLabel
	}
	return &models.ResourceHijackingRoleContext{
		Name:         name,
		Kind:         "current-foothold",
		PrincipalID:  stringPtrIf(currentIdentity.PrincipalID),
		RoleNames:    dedupeStrings(roleNames),
		ScopeIDs:     dedupeStrings(scopeIDs),
		ControlLabel: label,
		Summary:      summary,
	}
}

func resourceHijackingAPIMCapabilitySteps(controlOK bool) []models.ResourceHijackingCapabilityStep {
	return familyCapabilitySteps(resourceHijackingAPIMSteps, controlOK)
}

func resourceHijackingAPIMState(service models.ApiMgmtServiceAsset) models.ResourceHijackingAPIMState {
	return models.ResourceHijackingAPIMState{
		State:                   service.State,
		PublicNetworkAccess:     service.PublicNetworkAccess,
		VirtualNetworkType:      service.VirtualNetworkType,
		GatewayHostnames:        append([]string{}, service.GatewayHostnames...),
		BackendHostnames:        append([]string{}, service.BackendHostnames...),
		APICount:                service.APICount,
		SubscriptionCount:       service.SubscriptionCount,
		ActiveSubscriptionCount: service.ActiveSubscriptionCount,
		BackendCount:            service.BackendCount,
		PolicyCount:             service.PolicyCount,
		PolicyControlTypes:      append([]string{}, service.PolicyControlTypes...),
		NamedValueSecretCount:   service.NamedValueSecretCount,
		NamedValueKeyVaultCount: service.NamedValueKeyVaultCount,
		WorkloadIdentityType:    service.WorkloadIdentityType,
		Posture:                 resourceHijackingAPIMPosture(service),
	}
}

func resourceHijackingAPIMPosture(service models.ApiMgmtServiceAsset) string {
	parts := familyAPIMPostureParts(service, familyAPIMPostureOptions{
		IncludeActiveSubscriptionCount: true,
		IncludeNamedValueSecretPosture: true,
	})
	if len(parts) == 0 {
		return "APIM service visible without stronger routing posture"
	}
	return strings.Join(parts, "; ")
}

func resourceHijackingAPIMTakeoverRank(service models.ApiMgmtServiceAsset, controlOK bool) (int, string) {
	rank := 1
	reasons := []string{}
	hasGateway := len(service.GatewayHostnames) > 0
	hasBackend := len(service.BackendHostnames) > 0 || apiMgmtIntValue(service.BackendCount) > 0 || len(service.PolicyControlTypes) > 0
	hasConsumers := apiMgmtIntValue(service.ActiveSubscriptionCount) > 0 || apiMgmtIntValue(service.APISubscriptionRequiredCount) > 0
	switch {
	case hasGateway && hasBackend && hasConsumers:
		rank = 5
		reasons = append(reasons, "trusted gateway, backend target, and consumer/subscription posture are all visible")
	case hasGateway && hasBackend:
		rank = 4
		reasons = append(reasons, "trusted gateway and backend target posture are visible")
	case hasGateway || hasBackend:
		rank = 3
		reasons = append(reasons, "partial gateway or backend routing posture is visible")
	case apiMgmtIntValue(service.APICount) > 0:
		rank = 2
		reasons = append(reasons, "APIM APIs are visible but stronger backend routing posture is not")
	}
	if controlOK {
		reasons = append(reasons, "current identity has visible APIM backend or policy write control")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "visible posture does not support a stronger dynamic takeover ranking")
	}
	return rank, strings.Join(reasons, "; ")
}

func resourceHijackingAPIMNotCollectedByDefault() []models.ResourceHijackingBoundaryNote {
	return []models.ResourceHijackingBoundaryNote{
		{Name: "policy XML bodies", Classification: "recon safety", Reason: "The flat helper parses safe policy control types, but does not print raw policy XML or named-value expansions."},
		{Name: "named-value values", Classification: "recon safety", Reason: "Default output reports secret and Key Vault named-value counts without printing stored values."},
		{Name: "live request flow", Classification: "proof boundary", Reason: "Management-plane posture cannot prove traffic was routed, captured, or modified."},
		{Name: "backend ownership", Classification: "proof boundary", Reason: "A backend hostname does not prove who controls that endpoint or whether it is reachable."},
		{Name: "activity history", Classification: "API/noise", Reason: "Broad APIM history pulls are not needed for default posture and should be a narrow follow-up only when timing or actor proof matters."},
	}
}

func resourceHijackingAPIMSummary(service models.ApiMgmtServiceAsset, rank int, controlOK bool) string {
	parts := []string{fmt.Sprintf("service %q ranks %d/5 for APIM resource-hijack posture", service.Name, rank)}
	if len(service.GatewayHostnames) > 0 {
		parts = append(parts, fmt.Sprintf("%d gateway hostname(s)", len(service.GatewayHostnames)))
	}
	if len(service.BackendHostnames) > 0 {
		parts = append(parts, fmt.Sprintf("%d backend hostname(s)", len(service.BackendHostnames)))
	}
	if controlOK {
		parts = append(parts, "current identity can modify APIM backend or policy posture from visible RBAC")
	} else {
		parts = append(parts, "current identity backend or policy write control is not proven")
	}
	return strings.Join(parts, "; ") + "."
}
