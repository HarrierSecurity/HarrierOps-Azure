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

type persistenceAppServiceStepDefinition struct {
	Action     string
	APISurface string
}

var persistenceAppServiceSteps = []persistenceAppServiceStepDefinition{
	{Action: "create or reuse app service", APISurface: "Microsoft.Web/sites"},
	{Action: "set or reuse deployment path", APISurface: "source control / package deployment / site config"},
	{Action: "change app settings or identity attachment", APISurface: "app settings / connection strings / managed identity"},
	{Action: "deploy or replace application code", APISurface: "zip deploy / package deploy / publish"},
	{Action: "expose or reuse HTTP/HTTPS entry path", APISurface: "hostname / TLS / public network access"},
}

func buildPersistenceAppServiceOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.PersistenceSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	appServicesFuture := runGroupedCommandOutput[models.AppServicesOutput](group, ctx, request, appServicesHandler(provider, now), "app-services")
	envVarsFuture := runGroupedCommandOutput[models.EnvVarsOutput](group, ctx, request, envVarsHandler(provider, now), "env-vars")
	managedIdentitiesFuture := runGroupedCommandOutput[models.ManagedIdentitiesOutput](group, ctx, request, managedIdentitiesHandler(provider, now), "managed-identities")
	permissionsFuture := runGroupedCommandOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions")
	rbacFuture := runGroupedCommandOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac")

	appServices, err := appServicesFuture.wait()
	if err != nil {
		return nil, err
	}
	envVars, err := envVarsFuture.wait()
	if err != nil {
		return nil, err
	}
	managedIdentities, err := managedIdentitiesFuture.wait()
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
		stringPtrValue(appServices.Metadata.SubscriptionID),
		stringPtrValue(permissions.Metadata.SubscriptionID),
	)
	tenantID := firstNonEmpty(
		request.Tenant,
		stringPtrValue(appServices.Metadata.TenantID),
		stringPtrValue(permissions.Metadata.TenantID),
	)

	envVarsByAsset := make(map[string][]models.EnvVarSummary)
	for _, item := range envVars.EnvVars {
		if item.AssetID == "" || !strings.EqualFold(item.AssetKind, "AppService") {
			continue
		}
		envVarsByAsset[item.AssetID] = append(envVarsByAsset[item.AssetID], item)
	}

	evidence := buildPersistencePrincipalEvidence(permissions.Permissions, rbac.RoleAssignments)

	managedIdentitiesByAttachment := persistenceAppServiceManagedIdentitiesByAttachment(managedIdentities.Identities)
	apps := sortedByLess(appServices.AppServices, appServiceLess)
	rows := make([]models.PersistenceAppService, 0, len(apps))
	for _, app := range apps {
		control, controlOK := persistenceAppServiceControl(app.ID, evidence.currentIdentityAssignments)
		currentContext := persistenceCurrentIdentityContext(evidence.currentIdentity, control, controlOK)
		attachedManagedIdentities := persistenceAppServiceAttachedManagedIdentities(app, managedIdentitiesByAttachment)
		capabilitySteps := persistenceAppServiceCapabilitySteps(controlOK)
		appEnvVars := envVarsByAsset[app.ID]
		executionContextOptions := persistenceAppServiceExecutionContextOptions(app, appEnvVars)
		nearbyNames := persistenceAppServiceNearbyNames(apps, app.Name)
		strongestContext, strongestContextHasAzureControl := persistenceAppServiceExecutionContext(
			app,
			attachedManagedIdentities,
			evidence.permissionsByPrincipal,
			evidence.assignmentsByPrincipal,
		)

		rows = append(rows, models.PersistenceAppService{
			ID:                      app.ID,
			Name:                    app.Name,
			ResourceGroup:           app.ResourceGroup,
			Location:                app.Location,
			CapabilitySteps:         capabilitySteps,
			CurrentIdentityContext:  currentContext,
			ExecutionContextOptions: executionContextOptions,
			CurrentState: models.PersistenceAppServiceState{
				State:                            app.State,
				Hostname:                         app.DefaultHostname,
				PublicNetworkAccess:              app.PublicNetworkAccess,
				Runtime:                          app.RuntimeStack,
				Deployment:                       app.Deployment,
				DeploymentRepoURL:                app.DeploymentRepoURL,
				DeploymentBranch:                 app.DeploymentBranch,
				DeploymentIsGitHubAction:         app.DeploymentIsGitHubAction,
				DeploymentManualIntegration:      app.DeploymentManualIntegration,
				IdentityType:                     app.WorkloadIdentityType,
				AppSettingsCount:                 app.AppSettingsCount,
				KeyVaultReferenceCount:           app.KeyVaultReferenceCount,
				SensitiveSettingCount:            app.SensitiveSettingCount,
				ConnectionStringCount:            app.ConnectionStringCount,
				KeyVaultConnectionStringCount:    app.KeyVaultConnectionStringCount,
				ConnectionStringTypes:            append([]string{}, app.ConnectionStringTypes...),
				RunFromPackage:                   app.RunFromPackage,
				HTTPSOnly:                        boolPtr(app.HTTPSOnly),
				MinTLSVersion:                    app.MinTLSVersion,
				FTPSState:                        app.FTPSState,
				VisibleSensitiveSettingNames:     persistenceAppServiceSensitiveSettingNames(appEnvVars),
				StrongestVisibleExecutionContext: strongestContext,
				NearbyThematicNames:              nearbyNames,
			},
			StillUnmapped: persistenceAppServiceStillUnmapped(),
			Summary:       persistenceAppServiceSummary(app, controlOK, strongestContext, strongestContextHasAzureControl),
			RelatedIDs:    mergeRelatedIDs(app.RelatedIDs),
		})
	}

	issues := append([]models.Issue{}, appServices.Issues...)
	issues = append(issues, envVars.Issues...)
	issues = append(issues, managedIdentities.Issues...)
	issues = append(issues, permissions.Issues...)
	issues = append(issues, rbac.Issues...)

	return models.PersistenceAppServiceOutput{
		Metadata:           scopedMetadata(now, request, tenantID, subscriptionID, "persistence"),
		GroupedCommandName: "persistence",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		AppServices:        rows,
		Issues:             issues,
	}, nil
}

func persistenceAppServiceControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	bestRank := 99
	best := persistenceCurrentIdentityControl{}
	for _, assignment := range assignments {
		role := strings.ToLower(strings.TrimSpace(assignment.RoleName))
		if role != "owner" && role != "contributor" {
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

func persistenceAppServiceCapabilitySteps(controlOK bool) []models.PersistenceCapabilityStep {
	steps := make([]models.PersistenceCapabilityStep, 0, len(persistenceAppServiceSteps))
	for _, step := range persistenceAppServiceSteps {
		status := "not proven"
		if controlOK {
			status = "yes"
		}
		steps = append(steps, models.PersistenceCapabilityStep{
			Action:     step.Action,
			APISurface: step.APISurface,
			Status:     status,
		})
	}
	return steps
}

func persistenceAppServiceExecutionContextOptions(app models.AppServiceAsset, envVars []models.EnvVarSummary) []string {
	options := []string{}
	if strings.TrimSpace(stringPtrValue(app.WorkloadIdentityType)) != "" {
		options = append(options, "managed identity")
	}
	if intPtrValue(app.KeyVaultReferenceCount) > 0 {
		options = append(options, "Key Vault-backed settings")
	}
	if intPtrValue(app.SensitiveSettingCount) > 0 {
		options = append(options, "sensitive-looking app settings")
	}
	if intPtrValue(app.ConnectionStringCount) > 0 {
		options = append(options, "connection strings")
	}
	if intPtrValue(app.KeyVaultConnectionStringCount) > 0 {
		options = append(options, "Key Vault-backed connection strings")
	}
	for _, item := range envVars {
		if item.LooksSensitive && item.ValueType == "plain-text" {
			options = append(options, "plain-text sensitive-looking app settings")
			break
		}
	}
	return dedupeStrings(options)
}

func persistenceAppServiceExecutionContext(
	app models.AppServiceAsset,
	attachedManagedIdentities []models.ManagedIdentity,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool) {
	candidates := []persistenceAppServiceExecutionCandidate{}
	seenPrincipalIDs := map[string]struct{}{}
	for _, identity := range attachedManagedIdentities {
		candidate, ok := persistenceAppServiceExecutionCandidateFromManagedIdentity(identity, permissionsByPrincipal, assignmentsByPrincipal)
		if !ok {
			continue
		}
		if candidate.principalID != nil && strings.TrimSpace(*candidate.principalID) != "" {
			seenPrincipalIDs[strings.ToLower(strings.TrimSpace(*candidate.principalID))] = struct{}{}
		}
		candidates = append(candidates, candidate)
	}
	if app.WorkloadPrincipalID != nil && strings.TrimSpace(*app.WorkloadPrincipalID) != "" {
		key := strings.ToLower(strings.TrimSpace(*app.WorkloadPrincipalID))
		if _, ok := seenPrincipalIDs[key]; !ok {
			candidate, ok := persistenceAppServiceExecutionCandidateFromFallback(app, permissionsByPrincipal, assignmentsByPrincipal)
			if ok {
				candidates = append(candidates, candidate)
			}
		}
	}
	if len(candidates) == 0 {
		return nil, false
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return persistenceAppServiceExecutionCandidateLess(candidates[i], candidates[j])
	})
	best := candidates[0]
	return best.context, best.privileged
}

type persistenceAppServiceExecutionCandidate struct {
	name         string
	principalID  *string
	identityType *string
	context      *models.PersistenceRoleContext
	privileged   bool
}

func persistenceAppServiceExecutionCandidateFromManagedIdentity(
	identity models.ManagedIdentity,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (persistenceAppServiceExecutionCandidate, bool) {
	context, privileged, ok := persistenceAppServiceRoleContext(
		identity.Name,
		stringPtrIf(identity.IdentityType),
		identity.PrincipalID,
		permissionsByPrincipal,
		assignmentsByPrincipal,
	)
	if !ok {
		return persistenceAppServiceExecutionCandidate{}, false
	}
	return persistenceAppServiceExecutionCandidate{
		name:         identity.Name,
		principalID:  identity.PrincipalID,
		identityType: stringPtrIf(identity.IdentityType),
		context:      context,
		privileged:   privileged,
	}, true
}

func persistenceAppServiceExecutionCandidateFromFallback(
	app models.AppServiceAsset,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (persistenceAppServiceExecutionCandidate, bool) {
	name := persistenceAppServiceIdentityName(app)
	context, privileged, ok := persistenceAppServiceRoleContext(
		name,
		app.WorkloadIdentityType,
		app.WorkloadPrincipalID,
		permissionsByPrincipal,
		assignmentsByPrincipal,
	)
	if !ok {
		return persistenceAppServiceExecutionCandidate{}, false
	}
	return persistenceAppServiceExecutionCandidate{
		name:         name,
		principalID:  app.WorkloadPrincipalID,
		identityType: app.WorkloadIdentityType,
		context:      context,
		privileged:   privileged,
	}, true
}

func persistenceAppServiceRoleContext(
	name string,
	identityType *string,
	principalID *string,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool, bool) {
	return persistencePrincipalRoleContext(persistencePrincipalRoleContextOptions{
		fallbackName:           name,
		kind:                   "app-service-execution-context",
		principalID:            principalID,
		identityType:           identityType,
		permissionsByPrincipal: permissionsByPrincipal,
		assignmentsByPrincipal: assignmentsByPrincipal,
		resolvedSummary: func(name string, roleSummary string) string {
			return fmt.Sprintf("The strongest visible execution context here is the App Service identity `%s`, which already holds %s.", name, roleSummary)
		},
		lowerImpactSummary: func(name string) string {
			return fmt.Sprintf("App Service identity `%s` is visible here, but only lower-impact Azure role assignments are visible from current scope.", name)
		},
		unresolvedPrivilegedSummary: func(name string, _ string) string {
			return fmt.Sprintf("App Service identity `%s` is visible here, and raw Azure role-assignment rows for its principal ID suggest stronger Azure control, but that principal is not resolved as a standalone permissions row here.", name)
		},
		noAssignmentsSummary: func(name string) string {
			return fmt.Sprintf("App Service identity `%s` is visible here, but no Azure role-assignment rows are found for its principal ID.", name)
		},
		rbacOnlyCarriesAzureControl: false,
	})
}

func persistenceAppServiceExecutionCandidateLess(
	left, right persistenceAppServiceExecutionCandidate,
) bool {
	leftContext := left.context
	rightContext := right.context
	if leftContext == nil {
		return false
	}
	if rightContext == nil {
		return true
	}
	if left.privileged != right.privileged {
		return left.privileged
	}
	leftRoleRank := permissionRoleRank(leftContext.RoleNames)
	rightRoleRank := permissionRoleRank(rightContext.RoleNames)
	if leftRoleRank != rightRoleRank {
		return leftRoleRank < rightRoleRank
	}
	leftScopeRank := persistenceFunctionScopeBreadthRank(leftContext.ScopeIDs)
	rightScopeRank := persistenceFunctionScopeBreadthRank(rightContext.ScopeIDs)
	if leftScopeRank != rightScopeRank {
		return leftScopeRank < rightScopeRank
	}
	return leftContext.Name < rightContext.Name
}

func persistenceAppServiceStillUnmapped() []string {
	return []string{
		"the current command does not retrieve deployed application packages, repository contents, or source bundles, so operator intent is not inferred from code here",
		"the current command does not replay HTTP traffic or inspect runtime-side request handling, so conclusions stop at visible management-plane exposure posture",
		"the current command does not infer downstream API, data, or automation impact from runtime stack or setting names alone",
		"this App Service view stops at the main web host; use `persistence webjobs` when you need App Service WebJobs background-execution depth",
	}
}

func persistenceAppServiceSummary(
	app models.AppServiceAsset,
	controlOK bool,
	strongestContext *models.PersistenceRoleContext,
	strongestContextHasAzureControl bool,
) string {
	if controlOK && persistenceAppServiceShowsReusablePosture(app) && strongestContext != nil && strongestContextHasAzureControl {
		return fmt.Sprintf("Current identity can repurpose App Service '%s' as reusable App Service persistence, and the strongest visible execution context already carries Azure control.", app.Name)
	}
	if controlOK && persistenceAppServiceShowsReusablePosture(app) {
		return fmt.Sprintf("Current identity can repurpose App Service '%s' as reusable App Service persistence, with visible deployment, configuration, and reachable-host posture from the current read path.", app.Name)
	}
	if controlOK {
		return fmt.Sprintf("Current identity can build or repurpose App Service '%s', but the current read path only confirms part of the later reachable-host story.", app.Name)
	}
	if persistenceAppServiceShowsReusablePosture(app) {
		return fmt.Sprintf("App Service '%s' already shows durable deployment, configuration, and reachable-host posture, but the current identity does not yet have a proven path to repurpose it here.", app.Name)
	}
	return fmt.Sprintf("App Service '%s' is visible, but the current identity does not yet have a proven path to turn it into reusable App Service persistence.", app.Name)
}

func persistenceAppServiceShowsReusablePosture(app models.AppServiceAsset) bool {
	if strings.TrimSpace(stringPtrValue(app.DefaultHostname)) != "" {
		return true
	}
	if strings.EqualFold(stringPtrValue(app.PublicNetworkAccess), "Enabled") {
		return true
	}
	if strings.TrimSpace(stringPtrValue(app.Deployment)) != "" {
		return true
	}
	if intPtrValue(app.AppSettingsCount) > 0 || intPtrValue(app.ConnectionStringCount) > 0 {
		return true
	}
	return strings.TrimSpace(stringPtrValue(app.RuntimeStack)) != ""
}

func persistenceAppServiceManagedIdentitiesByAttachment(identities []models.ManagedIdentity) map[string][]models.ManagedIdentity {
	byAttachment := make(map[string][]models.ManagedIdentity)
	for _, identity := range identities {
		for _, attachment := range identity.AttachedTo {
			if attachment == "" {
				continue
			}
			byAttachment[attachment] = append(byAttachment[attachment], identity)
		}
	}
	return byAttachment
}

func persistenceAppServiceAttachedManagedIdentities(
	app models.AppServiceAsset,
	byAttachment map[string][]models.ManagedIdentity,
) []models.ManagedIdentity {
	items := append([]models.ManagedIdentity{}, byAttachment[app.ID]...)
	sort.SliceStable(items, func(i, j int) bool {
		leftSystem := strings.EqualFold(items[i].IdentityType, "systemAssigned")
		rightSystem := strings.EqualFold(items[j].IdentityType, "systemAssigned")
		if leftSystem != rightSystem {
			return leftSystem
		}
		return items[i].Name < items[j].Name
	})
	return items
}

func persistenceAppServiceSensitiveSettingNames(envVars []models.EnvVarSummary) []string {
	names := []string{}
	seen := map[string]struct{}{}
	for _, item := range envVars {
		if !item.LooksSensitive {
			continue
		}
		name := strings.TrimSpace(item.SettingName)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func persistenceAppServiceIdentityName(app models.AppServiceAsset) string {
	if strings.Contains(strings.ToLower(stringPtrValue(app.WorkloadIdentityType)), "userassigned") && len(app.WorkloadIdentityIDs) > 0 {
		return persistenceResourceNameFromID(app.WorkloadIdentityIDs[0])
	}
	return firstNonEmpty(app.Name+"-system", "app-service identity")
}

func persistenceAppServiceNearbyNames(
	apps []models.AppServiceAsset,
	currentAppName string,
) []string {
	seen := map[string]struct{}{}
	candidates := []string{}
	for _, app := range apps {
		name := strings.TrimSpace(app.Name)
		if name == "" || strings.EqualFold(name, currentAppName) {
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
	sort.SliceStable(candidates, func(i, j int) bool {
		leftScore := persistenceAutomationNameScore(candidates[i])
		rightScore := persistenceAutomationNameScore(candidates[j])
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return candidates[i] < candidates[j]
	})
	if len(candidates) > 4 {
		return append([]string{}, candidates[:4]...)
	}
	return candidates
}

func boolPtr(value bool) *bool {
	copied := value
	return &copied
}
