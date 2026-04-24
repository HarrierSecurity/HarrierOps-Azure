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

var pathMaskingRelaySteps = []familyStepDefinition{
	{Action: "select Relay namespace", APISurface: "Microsoft.Relay/namespaces", DownstreamEffect: "Azure becomes the visible rendezvous point while the listener and backend can stay off the direct public path.", Boundary: "Namespace posture does not prove a listener is currently connected."},
	{Action: "identify Hybrid Connections", APISurface: "Microsoft.Relay/namespaces/hybridConnections", DownstreamEffect: "Hybrid Connections show named private-path channels that can bridge callers toward internal services.", Boundary: "Hybrid Connection posture does not identify the backend process."},
	{Action: "review authorization rules", APISurface: "Microsoft.Relay/namespaces/authorizationRules", DownstreamEffect: "Authorization rules show where management-plane control could sustain or reshape send/listen access paths.", Boundary: "The command does not retrieve keys or prove data-plane use."},
	{Action: "change namespace or connection posture", APISurface: "Relay namespace and Hybrid Connection configuration", NeedsWrite: true, DownstreamEffect: "Can add, remove, or reconfigure the cloud rendezvous path while preserving an Azure-native integration story.", Boundary: "Write capability is inferred only from visible management-plane RBAC."},
	{Action: "blend as private connectivity", APISurface: "Relay service configuration", DownstreamEffect: "Normal cover stories include hybrid integration, firewall avoidance for approved apps, partner connectivity, and private service migration.", Boundary: "Cover story is not an intent claim."},
}

func buildPathMaskingRelayOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.PathMaskingSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	relayFuture := runGroupedCommandOutput[models.RelayOutput](group, ctx, request, relayHandler(provider, now), "relay")
	evidenceFutures := runFamilyEvidence(group, ctx, request, provider, now)

	relay, err := relayFuture.wait()
	if err != nil {
		return nil, err
	}
	evidence, err := evidenceFutures.wait()
	if err != nil {
		return nil, err
	}

	targets := make([]models.PathMaskingRelayTarget, 0, len(relay.Namespaces))
	for _, namespace := range relay.Namespaces {
		control, controlOK := pathMaskingRelayControl(namespace.ID, evidence.principal.currentIdentityAssignments)
		rank, reason := pathMaskingRelayRank(namespace, controlOK)
		targets = append(targets, models.PathMaskingRelayTarget{
			ID:                     namespace.ID,
			Name:                   namespace.Name,
			ResourceGroup:          namespace.ResourceGroup,
			Location:               namespace.Location,
			MaskingRank:            rank,
			MaskingReason:          reason,
			CapabilitySteps:        pathMaskingCapabilitySteps(pathMaskingRelaySteps, controlOK),
			CurrentIdentityContext: pathMaskingRoleContext(evidence.principal.currentIdentity, control, controlOK, "Relay namespace or Hybrid Connection write control", "Relay write"),
			CurrentState:           pathMaskingRelayState(namespace),
			NotCollectedByDefault:  pathMaskingRelayNotCollectedByDefault(),
			Summary:                pathMaskingRelaySummary(namespace, rank, controlOK),
			RelatedIDs:             mergeRelatedIDs(namespace.RelatedIDs),
		})
	}
	sort.SliceStable(targets, func(i, j int) bool {
		if targets[i].MaskingRank != targets[j].MaskingRank {
			return targets[i].MaskingRank > targets[j].MaskingRank
		}
		return targets[i].Name < targets[j].Name
	})

	issues := familyIssues(relay.Issues, evidence)

	return models.PathMaskingRelayOutput{
		Metadata: scopedMetadata(
			now,
			request,
			firstNonEmpty(request.Tenant, stringPtrValue(relay.Metadata.TenantID), stringPtrValue(evidence.permissions.Metadata.TenantID)),
			firstNonEmpty(request.Subscription, stringPtrValue(relay.Metadata.SubscriptionID), stringPtrValue(evidence.permissions.Metadata.SubscriptionID)),
			"pathmasking",
		),
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

func pathMaskingRelayControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	return resourceHijackingBestControl(
		resourceID,
		assignments,
		[]string{"Microsoft.Relay/namespaces/write", "Microsoft.Relay/namespaces/hybridConnections/write"},
		"Owner",
		"Contributor",
		"Azure Relay Owner",
	)
}

func pathMaskingRelayState(namespace models.RelayNamespaceAsset) models.PathMaskingRelayState {
	names := make([]string, 0, len(namespace.HybridConnections))
	listenerParts := []string{}
	attachments := []string{}
	for _, connection := range namespace.HybridConnections {
		names = append(names, connection.Name)
		if relayIntValue(connection.ListenerCount) > 0 {
			listenerParts = append(listenerParts, fmt.Sprintf("%s=%d", connection.Name, relayIntValue(connection.ListenerCount)))
		}
		for _, app := range connection.AppServiceAttachments {
			attachments = append(attachments, connection.Name+"->"+app)
		}
	}
	listenerSummary := "no listener counts visible"
	if len(listenerParts) > 0 {
		listenerSummary = strings.Join(listenerParts, "; ")
	}
	sort.Strings(names)
	sort.Strings(attachments)
	return models.PathMaskingRelayState{
		ServiceBusEndpoint:     namespace.ServiceBusEndpoint,
		HybridConnectionCount:  namespace.HybridConnectionCount,
		AuthorizationRuleCount: namespace.AuthorizationRuleCount,
		HybridConnectionNames:  names,
		ListenerSummary:        listenerSummary,
		AppServiceAttachments:  attachments,
		Posture:                pathMaskingRelayPosture(namespace, listenerSummary),
	}
}

func pathMaskingRelayPosture(namespace models.RelayNamespaceAsset, listenerSummary string) string {
	parts := []string{}
	if relayIntValue(namespace.HybridConnectionCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d Hybrid Connection(s)", relayIntValue(namespace.HybridConnectionCount)))
	}
	if relayIntValue(namespace.AuthorizationRuleCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d authorization rule(s)", relayIntValue(namespace.AuthorizationRuleCount)))
	}
	if listenerSummary != "no listener counts visible" {
		parts = append(parts, "listener counts "+listenerSummary)
	}
	if pathMaskingRelayAttachmentCount(namespace) > 0 {
		parts = append(parts, fmt.Sprintf("%d App Service Hybrid Connection attachment(s)", pathMaskingRelayAttachmentCount(namespace)))
	}
	if len(parts) == 0 {
		return "Relay namespace visible without stronger Hybrid Connection or authorization posture"
	}
	return strings.Join(parts, "; ")
}

func pathMaskingRelayRank(namespace models.RelayNamespaceAsset, controlOK bool) (int, string) {
	rank := 1
	reasons := []string{}
	hasHybrid := relayIntValue(namespace.HybridConnectionCount) > 0 || len(namespace.HybridConnections) > 0
	hasAuthRules := relayIntValue(namespace.AuthorizationRuleCount) > 0
	hasListener := false
	hasAttachment := pathMaskingRelayAttachmentCount(namespace) > 0
	for _, connection := range namespace.HybridConnections {
		if relayIntValue(connection.ListenerCount) > 0 {
			hasListener = true
			break
		}
	}
	switch {
	case hasHybrid && hasAuthRules && hasListener && hasAttachment:
		rank = 5
		reasons = append(reasons, "Hybrid Connection, authorization rule, listener-count, and App Service attachment posture are visible")
	case hasHybrid && hasAuthRules:
		rank = 4
		reasons = append(reasons, "Hybrid Connection and authorization rule posture are visible")
	case hasHybrid:
		rank = 3
		reasons = append(reasons, "Hybrid Connection posture is visible")
	case hasAuthRules:
		rank = 2
		reasons = append(reasons, "authorization rule posture is visible without a visible Hybrid Connection")
	}
	if controlOK {
		reasons = append(reasons, "current identity has visible Relay namespace or Hybrid Connection write control")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "visible posture does not support a stronger dynamic pathmasking ranking")
	}
	return rank, strings.Join(reasons, "; ")
}

func pathMaskingRelayNotCollectedByDefault() []models.PathMaskingBoundaryNote {
	return []models.PathMaskingBoundaryNote{
		{Name: "listener runtime state", Classification: "proof boundary", Reason: "Management-plane posture and listener counts do not prove a current listener process, host, or session."},
		{Name: "backend process and host", Classification: "proof boundary", Reason: "Relay names and endpoints do not identify the private service or process behind the listener."},
		{Name: "traffic contents", Classification: "proof boundary", Reason: "The command does not inspect Relay traffic or payloads."},
		{Name: "authorization keys", Classification: "recon safety", Reason: "Authorization rules are counted, but key material is not retrieved or printed."},
		{Name: "App Service backend internals", Classification: "proof boundary", Reason: "App Service Hybrid Connection attachments can show reachability posture, but they do not identify the private listener host, process, or traffic contents."},
	}
}

func pathMaskingRelaySummary(namespace models.RelayNamespaceAsset, rank int, controlOK bool) string {
	parts := []string{fmt.Sprintf("namespace %q ranks %d/5 for Relay pathmasking posture", namespace.Name, rank)}
	if relayIntValue(namespace.HybridConnectionCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d Hybrid Connection(s)", relayIntValue(namespace.HybridConnectionCount)))
	}
	if relayIntValue(namespace.AuthorizationRuleCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d authorization rule(s)", relayIntValue(namespace.AuthorizationRuleCount)))
	}
	if attachmentCount := pathMaskingRelayAttachmentCount(namespace); attachmentCount > 0 {
		parts = append(parts, fmt.Sprintf("%d App Service attachment(s)", attachmentCount))
	}
	if controlOK {
		parts = append(parts, "current identity can change Relay path posture from visible RBAC")
	} else {
		parts = append(parts, "current identity Relay write control is not proven")
	}
	return strings.Join(parts, "; ") + "."
}

func pathMaskingRelayAttachmentCount(namespace models.RelayNamespaceAsset) int {
	total := 0
	for _, connection := range namespace.HybridConnections {
		total += len(connection.AppServiceAttachments)
	}
	return total
}
