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

var pathMaskingAPIMSteps = []familyStepDefinition{
	{Action: "select public gateway", APISurface: "Microsoft.ApiManagement/service", DownstreamEffect: "Clients see the APIM hostname instead of the backend service path.", Boundary: "Gateway posture does not prove traffic volume."},
	{Action: "identify backend indirection", APISurface: "Microsoft.ApiManagement/service/backends", DownstreamEffect: "Shows where the published API surface can forward requests behind the gateway.", Boundary: "Backend hostname posture does not prove ownership or reachability."},
	{Action: "apply route or transform policy", APISurface: "APIM policies and backend settings", NeedsWrite: true, DownstreamEffect: "Can rewrite paths, switch backends, normalize auth, or keep the true route opaque to callers.", Boundary: "Policy XML bodies are not collected by default."},
	{Action: "preserve consumer-facing contract", APISurface: "APIs, operations, products, subscriptions", NeedsWrite: true, DownstreamEffect: "Existing products and subscriptions can keep caller behavior normal while the backend path stays abstracted.", Boundary: "Consumer use and request contents require runtime logs."},
	{Action: "blend as API gateway operations", APISurface: "APIM service configuration", DownstreamEffect: "Normal cover stories include API versioning, throttling, partner exposure, backend migration, and failover.", Boundary: "Cover story is not an intent claim."},
}

func buildPathMaskingAPIMOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.PathMaskingSurfaceContract,
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

	targets := make([]models.PathMaskingAPIMTarget, 0, len(apiMgmt.ApiManagementServices))
	for _, service := range apiMgmt.ApiManagementServices {
		control, controlOK := resourceHijackingAPIMControl(service.ID, evidence.principal.currentIdentityAssignments)
		rank, reason := pathMaskingAPIMRank(service, controlOK)
		targets = append(targets, models.PathMaskingAPIMTarget{
			ID:                     service.ID,
			Name:                   service.Name,
			ResourceGroup:          service.ResourceGroup,
			Location:               service.Location,
			MaskingRank:            rank,
			MaskingReason:          reason,
			CapabilitySteps:        pathMaskingCapabilitySteps(pathMaskingAPIMSteps, controlOK),
			CurrentIdentityContext: pathMaskingRoleContext(evidence.principal.currentIdentity, control, controlOK, "APIM route or backend policy control", "APIM route/backend write"),
			CurrentState:           pathMaskingAPIMState(service),
			NotCollectedByDefault:  pathMaskingAPIMNotCollectedByDefault(),
			Summary:                pathMaskingAPIMSummary(service, rank, controlOK),
			RelatedIDs:             mergeRelatedIDs(service.RelatedIDs),
		})
	}
	sort.SliceStable(targets, func(i, j int) bool {
		if targets[i].MaskingRank != targets[j].MaskingRank {
			return targets[i].MaskingRank > targets[j].MaskingRank
		}
		return targets[i].Name < targets[j].Name
	})

	issues := familyIssues(apiMgmt.Issues, evidence)

	return models.PathMaskingAPIMOutput{
		Metadata:           scopedMetadata(now, request, firstNonEmpty(request.Tenant, stringPtrValue(apiMgmt.Metadata.TenantID), stringPtrValue(evidence.permissions.Metadata.TenantID)), firstNonEmpty(request.Subscription, stringPtrValue(apiMgmt.Metadata.SubscriptionID), stringPtrValue(evidence.permissions.Metadata.SubscriptionID)), "pathmasking"),
		GroupedCommandName: "pathmasking",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		Targets:            targets,
		Issues:             issues,
	}, nil
}

func pathMaskingCapabilitySteps(steps []familyStepDefinition, controlOK bool) []models.PathMaskingCapabilityStep {
	return familyCapabilitySteps(steps, controlOK)
}

func pathMaskingRoleContext(currentIdentity models.PermissionRow, control persistenceCurrentIdentityControl, controlOK bool, controlName string, controlLabel string) *models.PathMaskingRoleContext {
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
	return &models.PathMaskingRoleContext{
		Name:         name,
		Kind:         "current-foothold",
		PrincipalID:  stringPtrIf(currentIdentity.PrincipalID),
		RoleNames:    dedupeStrings(roleNames),
		ScopeIDs:     dedupeStrings(scopeIDs),
		ControlLabel: label,
		Summary:      summary,
	}
}

func pathMaskingAPIMState(service models.ApiMgmtServiceAsset) models.PathMaskingAPIMState {
	return models.PathMaskingAPIMState{
		GatewayHostnames:        append([]string{}, service.GatewayHostnames...),
		BackendHostnames:        append([]string{}, service.BackendHostnames...),
		APICount:                service.APICount,
		SubscriptionCount:       service.SubscriptionCount,
		PolicyCount:             service.PolicyCount,
		PolicyControlTypes:      append([]string{}, service.PolicyControlTypes...),
		NamedValueSecretCount:   service.NamedValueSecretCount,
		NamedValueKeyVaultCount: service.NamedValueKeyVaultCount,
		PublicNetworkAccess:     service.PublicNetworkAccess,
		VirtualNetworkType:      service.VirtualNetworkType,
		Posture:                 pathMaskingAPIMPosture(service),
	}
}

func pathMaskingAPIMPosture(service models.ApiMgmtServiceAsset) string {
	parts := familyAPIMPostureParts(service, familyAPIMPostureOptions{
		IncludeAPICount:          true,
		IncludeSubscriptionCount: true,
	})
	if len(parts) == 0 {
		return "APIM service visible without stronger proxy or backend-indirection posture"
	}
	return strings.Join(parts, "; ")
}

func pathMaskingAPIMRank(service models.ApiMgmtServiceAsset, controlOK bool) (int, string) {
	rank := 1
	reasons := []string{}
	hasGateway := len(service.GatewayHostnames) > 0
	hasBackend := len(service.BackendHostnames) > 0 || apiMgmtIntValue(service.BackendCount) > 0 || len(service.PolicyControlTypes) > 0
	hasContract := apiMgmtIntValue(service.APICount) > 0 || apiMgmtIntValue(service.SubscriptionCount) > 0
	gatewayLabel := "gateway"
	if strings.EqualFold(strings.TrimSpace(stringPtrValue(service.PublicNetworkAccess)), "Enabled") {
		gatewayLabel = "public gateway"
	}
	switch {
	case hasGateway && hasBackend && hasContract:
		rank = 5
		reasons = append(reasons, gatewayLabel+", backend indirection, and API/consumer contract posture are visible")
	case hasGateway && hasBackend:
		rank = 4
		reasons = append(reasons, gatewayLabel+" and backend indirection posture are visible")
	case hasGateway || hasBackend:
		rank = 3
		reasons = append(reasons, "partial gateway or backend indirection posture is visible")
	case hasContract:
		rank = 2
		reasons = append(reasons, "API contract posture is visible without stronger backend masking evidence")
	}
	if controlOK {
		reasons = append(reasons, "current identity has visible APIM route or backend policy control")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "visible posture does not support a stronger dynamic pathmasking ranking")
	}
	return rank, strings.Join(reasons, "; ")
}

func pathMaskingAPIMNotCollectedByDefault() []models.PathMaskingBoundaryNote {
	return []models.PathMaskingBoundaryNote{
		{Name: "policy XML bodies", Classification: "recon safety", Reason: "The flat helper parses safe policy control types, but does not print raw policy XML or named-value expansions."},
		{Name: "live request flow", Classification: "proof boundary", Reason: "Management-plane posture cannot prove callers used this path or which backend received traffic."},
		{Name: "request contents", Classification: "proof boundary", Reason: "The command does not inspect APIM gateway logs or request payloads."},
		{Name: "backend ownership", Classification: "proof boundary", Reason: "A backend hostname does not prove who controls the target or what process answers."},
		{Name: "named-value values", Classification: "recon safety", Reason: "Secret and Key Vault-backed named values are counted but values are not printed."},
	}
}

func pathMaskingAPIMSummary(service models.ApiMgmtServiceAsset, rank int, controlOK bool) string {
	parts := []string{fmt.Sprintf("service %q ranks %d/5 for APIM pathmasking posture", service.Name, rank)}
	if len(service.GatewayHostnames) > 0 {
		parts = append(parts, fmt.Sprintf("%d gateway hostname(s)", len(service.GatewayHostnames)))
	}
	if len(service.BackendHostnames) > 0 {
		parts = append(parts, fmt.Sprintf("%d backend hostname(s)", len(service.BackendHostnames)))
	}
	if controlOK {
		parts = append(parts, "current identity can change route or backend posture from visible RBAC")
	} else {
		parts = append(parts, "current identity route or backend control is not proven")
	}
	return strings.Join(parts, "; ") + "."
}
