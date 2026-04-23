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

const (
	persistenceCurrentBehavior = "Grouped persistence walkthroughs. Use `ho-azure persistence` or `ho-azure persistence help` to list surfaces, then `ho-azure persistence <surface>` to run an implemented surface."
	persistenceCommandState    = contracts.StatusImplemented
)

var (
	persistenceInputModes            = []string{"live"}
	persistencePreferredArtifactMode = []string{"loot", "json"}
)

type persistenceSurfaceBuilder func(context.Context, providers.Provider, func() time.Time, Request, contracts.PersistenceSurfaceContract) (any, error)

var persistenceSurfaceBuilders = map[string]persistenceSurfaceBuilder{
	"automation":  buildPersistenceAutomationOutput,
	"app-service": buildPersistenceAppServiceOutput,
	"azure-ml":    buildPersistenceAzureMLOutput,
	"functions":   buildPersistenceFunctionsOutput,
	"logic-apps":  buildPersistenceLogicAppsOutput,
	"webjobs":     buildPersistenceWebJobsOutput,
}

func persistenceHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		surface := strings.TrimSpace(request.PersistenceSurface)
		if surface == "" {
			return buildPersistenceOverview(now, request, nil), nil
		}

		contract, ok := contracts.PersistenceSurface(surface)
		if !ok {
			return nil, fmt.Errorf("unknown persistence surface %q", surface)
		}
		builder, ok := persistenceSurfaceBuilders[surface]
		if !ok {
			return nil, fmt.Errorf("persistence surface %q is not implemented yet; scaffold contract is in place for migration", surface)
		}
		return builder(ctx, provider, now, request, contract)
	}
}

func buildPersistenceOverview(now func() time.Time, request Request, selectedSurface *string) models.PersistenceOverviewOutput {
	surfaces := make([]models.PersistenceSurfaceDescriptor, 0, len(contracts.PersistenceSurfaceNames()))
	for _, name := range contracts.PersistenceSurfaceNames() {
		surface, _ := contracts.PersistenceSurface(name)
		surfaces = append(surfaces, models.PersistenceSurfaceDescriptor{
			Surface:          surface.Name,
			State:            surface.Status,
			Summary:          surface.Summary,
			OperatorQuestion: surface.OperatorQuestion,
			BackingCommands:  append([]string{}, surface.BackingCommands...),
		})
	}

	return models.PersistenceOverviewOutput{
		Metadata:               scopedMetadata(now, request, request.Tenant, request.Subscription, "persistence"),
		GroupedCommandName:     "persistence",
		CommandState:           persistenceCommandState,
		CurrentBehavior:        persistenceCurrentBehavior,
		PlannedInputModes:      append([]string{}, persistenceInputModes...),
		PreferredArtifactOrder: append([]string{}, persistencePreferredArtifactMode...),
		SelectedSurface:        selectedSurface,
		Surfaces:               surfaces,
		Issues:                 []models.Issue{},
	}
}

type persistenceAutomationStepDefinition struct {
	Action                  string
	APISurface              string
	NeedsBroadResourceWrite bool
}

var persistenceAutomationSteps = []persistenceAutomationStepDefinition{
	{Action: "create or modify account", APISurface: "automationAccounts", NeedsBroadResourceWrite: true},
	{Action: "add or edit runbook", APISurface: "automationAccounts/runbooks"},
	{Action: "upload or replace code", APISurface: "runbook content update"},
	{Action: "publish runbook", APISurface: "runbook publish action"},
	{Action: "attach or reuse exec ctx", APISurface: "account identity / automation assets", NeedsBroadResourceWrite: true},
	{Action: "create schedule", APISurface: "automationAccounts/schedules"},
	{Action: "link schedule to runbook", APISurface: "automationAccounts/jobSchedules"},
	{Action: "create webhook", APISurface: "automation webhook resource"},
}

func buildPersistenceAutomationOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.PersistenceSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	automationFuture := runGroupedCommandOutput[models.AutomationOutput](group, ctx, request, automationHandler(provider, now), "automation")
	permissionsFuture := runGroupedCommandOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions")
	rbacFuture := runGroupedCommandOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac")

	automation, err := automationFuture.wait()
	if err != nil {
		return nil, err
	}
	permissions, err := permissionsFuture.wait()
	if err != nil {
		return nil, err
	}
	rbac, err := rbacFuture.wait()
	if err != nil {
		return nil, err
	}

	subscriptionID := firstNonEmpty(
		request.Subscription,
		stringPtrValue(automation.Metadata.SubscriptionID),
		stringPtrValue(permissions.Metadata.SubscriptionID),
	)
	tenantID := firstNonEmpty(
		request.Tenant,
		stringPtrValue(automation.Metadata.TenantID),
		stringPtrValue(permissions.Metadata.TenantID),
	)

	evidence := buildPersistencePrincipalEvidence(permissions.Permissions, rbac.RoleAssignments)

	accounts := make([]models.PersistenceAutomationAccount, 0, len(automation.AutomationAccounts))
	for _, account := range automation.AutomationAccounts {
		control, controlOK := persistenceAutomationControl(account.ID, evidence.currentIdentityAssignments)
		currentContext := persistenceCurrentIdentityContext(evidence.currentIdentity, control, controlOK)
		capabilitySteps := persistenceAutomationCapabilitySteps(control, controlOK)
		executionOptions := persistenceAutomationExecutionContextOptions(account)
		strongestContext, strongestContextHasAzureControl := persistenceAutomationExecutionContext(account, evidence.permissionsByPrincipal, evidence.assignmentsByPrincipal)
		nearbyNames := persistenceAutomationNearbyNames(automation.AutomationAccounts, account.Name)
		stillUnmapped := persistenceAutomationStillUnmapped(account)

		accounts = append(accounts, models.PersistenceAutomationAccount{
			ID:                      account.ID,
			Name:                    account.Name,
			ResourceGroup:           account.ResourceGroup,
			Location:                account.Location,
			CapabilitySteps:         capabilitySteps,
			CurrentIdentityContext:  currentContext,
			ExecutionContextOptions: executionOptions,
			CurrentState: models.PersistenceAutomationAccountState{
				RunbookCount:                     account.RunbookCount,
				PublishedRunbookCount:            account.PublishedRunbookCount,
				PublishedRunbookNames:            append([]string{}, account.PublishedRunbookNames...),
				ScheduleCount:                    account.ScheduleCount,
				JobScheduleCount:                 account.JobScheduleCount,
				WebhookCount:                     account.WebhookCount,
				HybridWorkerGroupCount:           account.HybridWorkerGroupCount,
				CredentialCount:                  account.CredentialCount,
				CertificateCount:                 account.CertificateCount,
				ConnectionCount:                  account.ConnectionCount,
				VariableCount:                    account.VariableCount,
				EncryptedVariableCount:           account.EncryptedVariableCount,
				PrimaryStartMode:                 account.PrimaryStartMode,
				PrimaryRunbookName:               account.PrimaryRunbookName,
				IdentityType:                     account.IdentityType,
				StrongestVisibleExecutionContext: strongestContext,
				NearbyThematicNames:              nearbyNames,
				MissingTargetMapping:             account.MissingTargetMapping,
			},
			StillUnmapped: stillUnmapped,
			Summary:       persistenceAutomationSummary(account, controlOK, strongestContext, strongestContextHasAzureControl),
			RelatedIDs:    mergeRelatedIDs(account.RelatedIDs),
		})
	}

	issues := append([]models.Issue{}, automation.Issues...)
	issues = append(issues, permissions.Issues...)
	issues = append(issues, rbac.Issues...)

	return models.PersistenceAutomationOutput{
		Metadata:           scopedMetadata(now, request, tenantID, subscriptionID, "persistence"),
		GroupedCommandName: "persistence",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		AutomationAccounts: accounts,
		Issues:             issues,
	}, nil
}

func persistenceCurrentIdentityContext(
	currentIdentity models.PermissionRow,
	control persistenceCurrentIdentityControl,
	controlOK bool,
) *models.PersistenceRoleContext {
	if strings.TrimSpace(currentIdentity.DisplayName) == "" && !controlOK {
		return nil
	}

	name := firstNonEmpty(currentIdentity.DisplayName, "current identity")
	roleNames := append([]string{}, currentIdentity.HighImpactRoles...)
	if len(roleNames) == 0 {
		roleNames = append(roleNames, currentIdentity.AllRoleNames...)
	}
	scopeIDs := append([]string{}, currentIdentity.ScopeIDs...)
	summary := "Current foothold identity is visible, but no direct resource-control role is confirmed here yet."
	if controlOK {
		summary = fmt.Sprintf("Current foothold `%s` already holds %s.", name, control.RoleName)
		role := strings.TrimSpace(strings.SplitN(control.RoleName, " at ", 2)[0])
		if role != "" {
			roleNames = []string{role}
		}
		if control.ScopeID != "" {
			scopeIDs = []string{control.ScopeID}
		}
	}

	return &models.PersistenceRoleContext{
		Name:        name,
		Kind:        "current-foothold",
		PrincipalID: stringPtrIf(currentIdentity.PrincipalID),
		RoleNames:   dedupeStrings(roleNames),
		ScopeIDs:    dedupeStrings(scopeIDs),
		Summary:     summary,
	}
}

func persistenceAutomationCapabilitySteps(
	control persistenceCurrentIdentityControl,
	controlOK bool,
) []models.PersistenceCapabilityStep {
	controlRole := strings.ToLower(strings.TrimSpace(strings.SplitN(control.RoleName, " at ", 2)[0]))
	steps := make([]models.PersistenceCapabilityStep, 0, len(persistenceAutomationSteps))
	for _, step := range persistenceAutomationSteps {
		status := "not proven"
		if controlOK {
			status = "yes"
			if step.NeedsBroadResourceWrite && controlRole == "automation contributor" {
				status = "not proven"
			}
		}
		steps = append(steps, models.PersistenceCapabilityStep{
			Action:     step.Action,
			APISurface: step.APISurface,
			Status:     status,
		})
	}
	return steps
}

func persistenceAutomationExecutionContextOptions(account models.AutomationAccountAsset) []string {
	options := []string{}
	if strings.TrimSpace(stringPtrValue(account.IdentityType)) != "" {
		options = append(options, "managed identity")
	}
	if intPtrValue(account.CredentialCount) > 0 {
		options = append(options, "stored credentials")
	}
	if intPtrValue(account.ConnectionCount) > 0 {
		options = append(options, "connections")
	}
	if intPtrValue(account.CertificateCount) > 0 {
		options = append(options, "certificates")
	}
	if intPtrValue(account.VariableCount) > 0 {
		options = append(options, "variables")
	}
	return dedupeStrings(options)
}

func persistenceAutomationExecutionContext(
	account models.AutomationAccountAsset,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool) {
	context, carriesAzureControl, _ := persistencePrincipalRoleContext(persistencePrincipalRoleContextOptions{
		fallbackName:           firstNonEmpty(account.Name+"-identity", "automation identity"),
		kind:                   "automation-execution-context",
		principalID:            account.PrincipalID,
		identityType:           account.IdentityType,
		permissionsByPrincipal: permissionsByPrincipal,
		assignmentsByPrincipal: assignmentsByPrincipal,
		resolvedSummary: func(name string, roleSummary string) string {
			return fmt.Sprintf("The strongest visible execution context here is the Automation Account identity `%s`, which already holds %s.", name, roleSummary)
		},
		lowerImpactSummary: func(name string) string {
			return fmt.Sprintf("Automation identity `%s` is visible here, but only lower-impact Azure role assignments are visible from current scope.", name)
		},
		unresolvedPrivilegedSummary: func(name string, roleSummary string) string {
			return fmt.Sprintf("The strongest visible execution context here is the Automation Account identity `%s`, which already holds %s.", name, roleSummary)
		},
		noAssignmentsSummary: func(name string) string {
			return fmt.Sprintf("Automation identity `%s` is visible here, but no Azure role-assignment rows are found for its principal ID.", name)
		},
		rbacOnlyCarriesAzureControl: true,
	})
	return context, carriesAzureControl
}

func persistenceRoleSummary(roleNames []string, scopeIDs []string) string {
	roleText := naturalJoin(dedupeStrings(roleNames))
	if roleText == "" {
		roleText = "visible Azure roles"
	}
	if len(scopeIDs) == 0 {
		return roleText
	}
	if len(scopeIDs) == 1 {
		return fmt.Sprintf("%s at %s", roleText, persistenceScopeLabel(scopeIDs[0]))
	}
	return fmt.Sprintf("%s across %d visible scopes", roleText, len(scopeIDs))
}

func persistenceAutomationStillUnmapped(account models.AutomationAccountAsset) []string {
	items := []string{
		"the exact runbook code, content source, or operator intent behind this Automation surface",
	}
	if account.MissingTargetMapping {
		items = append(items, "the exact downstream resources, workflows, or credentials this automation path would modify")
	}
	items = append(items, "the full schedule cadence, webhook URI, or trigger usefulness beyond the currently modeled metadata")
	return dedupeStrings(items)
}

func persistenceAutomationSummary(
	account models.AutomationAccountAsset,
	controlOK bool,
	strongestContext *models.PersistenceRoleContext,
	strongestContextHasAzureControl bool,
) string {
	if controlOK && strongestContext != nil && strongestContextHasAzureControl {
		return fmt.Sprintf("Current identity can manage Automation Account '%s' end to end, and the strongest visible execution context already carries Azure control.", account.Name)
	}
	if controlOK {
		return fmt.Sprintf("Current identity can manage Automation Account '%s' end to end from current RBAC evidence.", account.Name)
	}
	return fmt.Sprintf("Automation Account '%s' is visible, but the current identity does not yet have a proven end-to-end management path here.", account.Name)
}

func persistenceAutomationNearbyNames(
	accounts []models.AutomationAccountAsset,
	currentAccountName string,
) []string {
	seen := map[string]struct{}{}
	candidates := []string{}
	for _, account := range accounts {
		pools := [][]string{
			{account.Name},
			account.PublishedRunbookNames,
			account.ScheduleRunbookNames,
			account.WebhookRunbookNames,
		}
		for _, pool := range pools {
			for _, name := range pool {
				name = strings.TrimSpace(name)
				if name == "" || strings.EqualFold(name, currentAccountName) {
					continue
				}
				if persistenceAutomationNameScore(name) == 0 {
					continue
				}
				if _, ok := seen[strings.ToLower(name)]; ok {
					continue
				}
				seen[strings.ToLower(name)] = struct{}{}
				candidates = append(candidates, name)
			}
		}
	}
	sort.SliceStable(candidates, func(i int, j int) bool {
		leftScore := persistenceAutomationNameScore(candidates[i])
		rightScore := persistenceAutomationNameScore(candidates[j])
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return false
	})
	if len(candidates) > 4 {
		return append([]string{}, candidates[:4]...)
	}
	return candidates
}

func persistenceAutomationNameScore(name string) int {
	lower := strings.ToLower(name)
	strongKeywords := []string{"baseline", "nightly", "maintenance", "reconcile", "reapply", "backup", "schedule", "patch"}
	for _, keyword := range strongKeywords {
		if strings.Contains(lower, keyword) {
			return 2
		}
	}
	weakKeywords := []string{"rotate", "sync", "agent", "config", "job"}
	for _, keyword := range weakKeywords {
		if strings.Contains(lower, keyword) {
			return 1
		}
	}
	return 0
}
