package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"harrierops-azure/internal/artifacts"
	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

type groupedSurfaceBuilder func(context.Context, providers.Provider, func() time.Time, Request, contracts.SurfaceContract) (any, error)

type familyStepDefinition struct {
	Action           string
	APISurface       string
	NeedsWrite       bool
	DownstreamEffect string
	Boundary         string
}

type groupedFamilyConfig struct {
	CommandName            string
	CurrentBehavior        string
	CommandState           string
	InputModes             []string
	PreferredArtifactOrder []string
	Selector               func(Request) string
	Overview               func(func() time.Time, Request, *string) any
	SurfaceNames           func() []string
	SurfaceContract        func(string) (contracts.SurfaceContract, bool)
	SurfaceBuilders        map[string]groupedSurfaceBuilder
}

type familyEvidenceFutures struct {
	permissions asyncCommandOutput[models.PermissionsOutput]
	rbac        asyncCommandOutput[models.RbacOutput]
}

type familyEvidence struct {
	permissions      models.PermissionsOutput
	rbac             models.RbacOutput
	principal        persistencePrincipalEvidence
	sessionArtifacts []models.SessionArtifact
}

var helperArtifactAnchorCommands = []string{
	"whoami",
	"permissions",
	"rbac",
	"principals",
	"automation",
	"logic-apps",
	"api-mgmt",
	"relay",
	"dcr",
	"diagnostic-settings",
	"appinsights",
	"monitoring-sinks",
	"app-services",
	"vm-extensions",
	"event-grid",
	"storage",
	"keyvault",
	"managed-identities",
	"vms",
}

func runFamilyEvidence(group commandOutputGroup, ctx context.Context, request Request, provider providers.Provider, now func() time.Time) familyEvidenceFutures {
	expected := helperArtifactExpectedSessions(ctx, request, provider, now, "permissions", "rbac")
	return runFamilyEvidenceWithExpected(group, ctx, request, provider, now, expected)
}

func runFamilyEvidenceWithExpected(group commandOutputGroup, ctx context.Context, request Request, provider providers.Provider, now func() time.Time, expected map[string]artifacts.ExpectedSession) familyEvidenceFutures {
	return familyEvidenceFutures{
		permissions: runPermissionsOutput(group, ctx, request, provider, now, expected),
		rbac:        runRBACOutput(group, ctx, request, provider, now, expected),
	}
}

func (futures familyEvidenceFutures) wait() (familyEvidence, error) {
	permissions, rbac, principal, sessionArtifacts, err := waitPermissionsRBACBundle(futures.permissions, futures.rbac)
	if err != nil {
		return familyEvidence{}, err
	}
	return familyEvidence{
		permissions:      permissions,
		rbac:             rbac,
		principal:        principal,
		sessionArtifacts: sessionArtifacts,
	}, nil
}

func waitPermissionsRBACBundle(
	permissionsFuture asyncCommandOutput[models.PermissionsOutput],
	rbacFuture asyncCommandOutput[models.RbacOutput],
) (models.PermissionsOutput, models.RbacOutput, persistencePrincipalEvidence, []models.SessionArtifact, error) {
	permissions, permissionsSource, err := permissionsFuture.waitWithSource()
	if err != nil {
		return models.PermissionsOutput{}, models.RbacOutput{}, persistencePrincipalEvidence{}, nil, err
	}
	rbac, rbacSource, err := rbacFuture.waitWithSource()
	if err != nil {
		return models.PermissionsOutput{}, models.RbacOutput{}, persistencePrincipalEvidence{}, nil, err
	}
	sessionArtifacts := []models.SessionArtifact{}
	if permissionsSource != nil {
		sessionArtifacts = append(sessionArtifacts, *permissionsSource)
	}
	if rbacSource != nil {
		sessionArtifacts = append(sessionArtifacts, *rbacSource)
	}
	return permissions, rbac, buildPersistencePrincipalEvidence(permissions.Permissions, rbac.RoleAssignments), sessionArtifacts, nil
}

func helperArtifactExpectedSessions(ctx context.Context, request Request, provider providers.Provider, now func() time.Time, commands ...string) map[string]artifacts.ExpectedSession {
	workspace := artifactWorkspace(request.OutDir)
	hasCandidate := false
	for _, command := range commands {
		if artifacts.HasSessionArtifact(workspace, command) {
			hasCandidate = true
			break
		}
	}
	if !hasCandidate {
		return nil
	}

	anchor, ok := loadHelperArtifactAnchor(ctx, request, provider, now)
	if !ok {
		return nil
	}
	expected := make(map[string]artifacts.ExpectedSession, len(commands))
	for _, command := range commands {
		expected[command] = artifacts.ExpectedSession{
			Command:        command,
			SchemaVersion:  contracts.AzureFoxSchemaVersion,
			ToolVersion:    toolVersion,
			TenantID:       anchor.TenantID,
			SubscriptionID: anchor.SubscriptionID,
			CurrentPrincipal: models.ArtifactPrincipal{
				ID:            anchor.CurrentPrincipal.ID,
				PrincipalType: anchor.CurrentPrincipal.PrincipalType,
				TenantID:      anchor.CurrentPrincipal.TenantID,
			},
			AuthMode:       anchor.AuthMode,
			TokenSource:    anchor.TokenSource,
			CommandOptions: artifactCommandOptions(command, request),
			MaxAge:         60 * time.Minute,
			Now:            now().UTC(),
		}
	}
	return expected
}

func loadHelperArtifactAnchor(ctx context.Context, request Request, provider providers.Provider, now func() time.Time) (artifacts.SessionAnchor, bool) {
	workspace := artifactWorkspace(request.OutDir)
	anchor, ok, err := artifacts.LoadSessionAnchorFromCommands(workspace, helperArtifactAnchorCommands, contracts.AzureFoxSchemaVersion, toolVersion, 30*time.Minute, now().UTC())
	if err == nil && ok {
		return anchor, true
	}

	whoami, err := runCommandOutput[models.WhoAmIOutput](ctx, request, whoAmIHandler(provider, now), "whoami")
	if err != nil {
		return artifacts.SessionAnchor{}, false
	}
	if _, err := artifacts.Write("whoami", whoami, workspace, models.RenderContext{
		Tenant:       whoami.TenantID,
		Subscription: whoami.Subscription.ID,
	}); err != nil {
		return artifacts.SessionAnchor{}, false
	}
	return artifacts.SessionAnchor{
		TenantID:       whoami.TenantID,
		SubscriptionID: whoami.Subscription.ID,
		CurrentPrincipal: models.ArtifactPrincipal{
			ID:            whoami.Principal.ID,
			PrincipalType: whoami.Principal.PrincipalType,
			TenantID:      whoami.Principal.TenantID,
		},
		AuthMode:    stringPtrValue(whoami.Metadata.AuthMode),
		TokenSource: stringPtrValue(whoami.Metadata.TokenSource),
	}, true
}

func runPermissionsOutput(
	group commandOutputGroup,
	ctx context.Context,
	request Request,
	provider providers.Provider,
	now func() time.Time,
	expected map[string]artifacts.ExpectedSession,
) asyncCommandOutput[models.PermissionsOutput] {
	return runHelperOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions", expected)
}

func runRBACOutput(
	group commandOutputGroup,
	ctx context.Context,
	request Request,
	provider providers.Provider,
	now func() time.Time,
	expected map[string]artifacts.ExpectedSession,
) asyncCommandOutput[models.RbacOutput] {
	return runHelperOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac", expected)
}

func runHelperOutput[T any](
	group commandOutputGroup,
	ctx context.Context,
	request Request,
	handler Handler,
	command string,
	expected map[string]artifacts.ExpectedSession,
) asyncCommandOutput[T] {
	if session, ok := expected[command]; ok {
		return runGroupedCommandOutputWithArtifact[T](group, ctx, request, handler, session)
	}
	return runGroupedCommandOutputWritingArtifact[T](group, ctx, request, handler, command)
}

func familyIssues(base []models.Issue, evidence familyEvidence) []models.Issue {
	issues := append([]models.Issue{}, base...)
	issues = append(issues, evidence.permissions.Issues...)
	issues = append(issues, evidence.rbac.Issues...)
	return issues
}

func appendSessionArtifact(base []models.SessionArtifact, source *models.SessionArtifact) []models.SessionArtifact {
	if source == nil {
		return append([]models.SessionArtifact{}, base...)
	}
	items := append([]models.SessionArtifact{}, base...)
	items = append(items, *source)
	return items
}

func groupedFamilyHandler(provider providers.Provider, now func() time.Time, config groupedFamilyConfig) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		surface := strings.TrimSpace(config.Selector(request))
		if surface == "" {
			return config.Overview(now, request, nil), nil
		}

		contract, ok := config.SurfaceContract(surface)
		if !ok {
			return nil, fmt.Errorf("unknown %s surface %q", config.CommandName, surface)
		}
		builder, ok := config.SurfaceBuilders[surface]
		if !ok {
			return nil, fmt.Errorf("%s surface %q is not implemented yet", config.CommandName, surface)
		}
		return builder(ctx, provider, now, request, contract)
	}
}

func groupedFamilySurfaceDescriptors(config groupedFamilyConfig) []models.FamilySurfaceDescriptor {
	surfaces := make([]models.FamilySurfaceDescriptor, 0, len(config.SurfaceNames()))
	for _, name := range config.SurfaceNames() {
		surface, _ := config.SurfaceContract(name)
		surfaces = append(surfaces, models.FamilySurfaceDescriptor{
			Surface:          surface.Name,
			State:            surface.Status,
			Summary:          surface.Summary,
			OperatorQuestion: surface.OperatorQuestion,
			BackingCommands:  append([]string{}, surface.BackingCommands...),
		})
	}
	return surfaces
}

func familyCapabilitySteps(steps []familyStepDefinition, controlOK bool) []models.FamilyCapabilityStep {
	rows := make([]models.FamilyCapabilityStep, 0, len(steps))
	for _, step := range steps {
		status := "visible posture only"
		canAct := false
		if step.NeedsWrite {
			if controlOK {
				status = "yes"
				canAct = true
			} else {
				status = "not proven"
			}
		}
		rows = append(rows, models.FamilyCapabilityStep{
			Action:           step.Action,
			APISurface:       step.APISurface,
			Status:           status,
			CanAct:           canAct,
			DownstreamEffect: step.DownstreamEffect,
			Boundary:         step.Boundary,
		})
	}
	return rows
}

type familyAPIMPostureOptions struct {
	IncludeAPICount                bool
	IncludeSubscriptionCount       bool
	IncludeActiveSubscriptionCount bool
	IncludeNamedValueSecretPosture bool
}

func familyAPIMPostureParts(service models.ApiMgmtServiceAsset, options familyAPIMPostureOptions) []string {
	parts := []string{}
	if len(service.GatewayHostnames) > 0 {
		parts = append(parts, fmt.Sprintf("%d gateway hostname(s)", len(service.GatewayHostnames)))
	}
	if len(service.BackendHostnames) > 0 {
		parts = append(parts, fmt.Sprintf("%d backend hostname(s)", len(service.BackendHostnames)))
	}
	if options.IncludeAPICount && apiMgmtIntValue(service.APICount) > 0 {
		parts = append(parts, fmt.Sprintf("%d API(s)", apiMgmtIntValue(service.APICount)))
	}
	if options.IncludeSubscriptionCount && apiMgmtIntValue(service.SubscriptionCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d subscription(s)", apiMgmtIntValue(service.SubscriptionCount)))
	}
	if len(service.PolicyControlTypes) > 0 {
		parts = append(parts, "policy controls: "+strings.Join(service.PolicyControlTypes, ", "))
	}
	if options.IncludeActiveSubscriptionCount && apiMgmtIntValue(service.ActiveSubscriptionCount) > 0 {
		parts = append(parts, fmt.Sprintf("%d active subscription(s)", apiMgmtIntValue(service.ActiveSubscriptionCount)))
	}
	if options.IncludeNamedValueSecretPosture && (apiMgmtIntValue(service.NamedValueSecretCount) > 0 || apiMgmtIntValue(service.NamedValueKeyVaultCount) > 0) {
		parts = append(parts, "secret or Key Vault named-value posture")
	}
	return parts
}

func familyLogicAppState(workflow models.LogicAppWorkflowAsset, posture string) models.FamilyLogicAppState {
	return models.FamilyLogicAppState{
		Platform:                         workflow.Platform,
		State:                            workflow.State,
		TriggerTypes:                     append([]string{}, workflow.TriggerTypes...),
		ExternallyCallableRequestTrigger: workflow.ExternallyCallableRequestTrigger,
		RecurrenceSummary:                workflow.RecurrenceSummary,
		DownstreamActionKinds:            append([]string{}, workflow.DownstreamActionKinds...),
		ConnectorReferences:              append([]string{}, workflow.ConnectorReferences...),
		ParameterNames:                   append([]string{}, workflow.ParameterNames...),
		DownstreamResourceReferences:     append([]string{}, workflow.DownstreamResourceReferences...),
		IdentityType:                     workflow.IdentityType,
		IdentityIDs:                      append([]string{}, workflow.IdentityIDs...),
		Posture:                          posture,
	}
}

func familyLogicAppPosture(workflow models.LogicAppWorkflowAsset, emptyPosture string) string {
	parts := []string{}
	if workflow.ExternallyCallableRequestTrigger {
		parts = append(parts, "externally callable request trigger")
	}
	if workflow.RecurrenceSummary != nil && strings.TrimSpace(*workflow.RecurrenceSummary) != "" {
		parts = append(parts, "recurrence trigger "+*workflow.RecurrenceSummary)
	}
	if len(workflow.TriggerTypes) > 0 {
		parts = append(parts, "trigger types "+strings.Join(workflow.TriggerTypes, ", "))
	}
	if len(workflow.DownstreamActionKinds) > 0 {
		parts = append(parts, "downstream actions "+strings.Join(workflow.DownstreamActionKinds, ", "))
	}
	if len(workflow.ConnectorReferences) > 0 {
		parts = append(parts, "connector references "+strings.Join(workflow.ConnectorReferences, ", "))
	}
	if len(workflow.DownstreamResourceReferences) > 0 {
		parts = append(parts, fmt.Sprintf("%d downstream resource reference(s)", len(workflow.DownstreamResourceReferences)))
	}
	if strings.TrimSpace(stringPtrValue(workflow.IdentityType)) != "" {
		parts = append(parts, "managed identity posture")
	}
	if len(parts) == 0 {
		return emptyPosture
	}
	return strings.Join(parts, "; ")
}
