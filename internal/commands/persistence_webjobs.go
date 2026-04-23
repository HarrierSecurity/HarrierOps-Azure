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

type persistenceWebJobStepDefinition struct {
	Action     string
	APISurface string
}

var persistenceWebJobSteps = []persistenceWebJobStepDefinition{
	{Action: "create or reuse parent app service", APISurface: "Microsoft.Web/sites"},
	{Action: "add or replace webjob package", APISurface: "site/wwwroot/app_data/jobs"},
	{Action: "set or reuse webjob mode", APISurface: "continuouswebjobs / triggeredwebjobs / scheduler"},
	{Action: "reuse inherited app execution context", APISurface: "parent app identity / app settings / connection strings"},
	{Action: "leave or repurpose rerun path", APISurface: "Kudu and App Service runtime discovery / schedule / manual trigger"},
}

func buildPersistenceWebJobsOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.PersistenceSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	webJobsFuture := runGroupedCommandOutput[models.WebJobsOutput](group, ctx, request, webJobsHandler(provider, now), "webjobs")
	appServicesFuture := runGroupedCommandOutput[models.AppServicesOutput](group, ctx, request, appServicesHandler(provider, now), "app-services")
	managedIdentitiesFuture := runGroupedCommandOutput[models.ManagedIdentitiesOutput](group, ctx, request, managedIdentitiesHandler(provider, now), "managed-identities")
	permissionsFuture := runGroupedCommandOutput[models.PermissionsOutput](group, ctx, request, permissionsHandler(provider, now), "permissions")
	rbacFuture := runGroupedCommandOutput[models.RbacOutput](group, ctx, request, rbacHandler(provider, now), "rbac")

	webJobs, err := webJobsFuture.wait()
	if err != nil {
		return nil, err
	}
	appServices, err := appServicesFuture.wait()
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
		stringPtrValue(webJobs.Metadata.SubscriptionID),
		stringPtrValue(appServices.Metadata.SubscriptionID),
		stringPtrValue(permissions.Metadata.SubscriptionID),
	)
	tenantID := firstNonEmpty(
		request.Tenant,
		stringPtrValue(webJobs.Metadata.TenantID),
		stringPtrValue(appServices.Metadata.TenantID),
		stringPtrValue(permissions.Metadata.TenantID),
	)

	appsByID := make(map[string]models.AppServiceAsset, len(appServices.AppServices))
	for _, app := range appServices.AppServices {
		if strings.TrimSpace(app.ID) == "" {
			continue
		}
		appsByID[app.ID] = app
	}

	evidence := buildPersistencePrincipalEvidence(permissions.Permissions, rbac.RoleAssignments)

	managedIdentitiesByAttachment := persistenceAppServiceManagedIdentitiesByAttachment(managedIdentities.Identities)
	items := sortedByLess(webJobs.WebJobs, webJobLess)
	rows := make([]models.PersistenceWebJob, 0, len(items))
	for _, job := range items {
		control, controlOK := persistenceAutomationControl(job.ParentAppID, evidence.currentIdentityAssignments)
		currentContext := persistenceCurrentIdentityContext(evidence.currentIdentity, control, controlOK)
		capabilitySteps := persistenceWebJobCapabilitySteps(controlOK)
		parentApp, parentAppVisible := appsByID[job.ParentAppID]
		attachedManagedIdentities := persistenceWebJobAttachedManagedIdentities(job, parentApp, parentAppVisible, managedIdentitiesByAttachment)
		executionOptions := persistenceWebJobExecutionContextOptions(job, parentApp, parentAppVisible, attachedManagedIdentities)
		strongestContext, strongestContextHasAzureControl := persistenceWebJobExecutionContext(
			job,
			parentApp,
			parentAppVisible,
			attachedManagedIdentities,
			evidence.permissionsByPrincipal,
			evidence.assignmentsByPrincipal,
		)
		nearbyNames := persistenceWebJobNearbyNames(items, job.Name)

		rows = append(rows, models.PersistenceWebJob{
			ID:                      job.ID,
			Name:                    job.Name,
			ResourceGroup:           job.ResourceGroup,
			Location:                job.Location,
			CapabilitySteps:         capabilitySteps,
			CurrentIdentityContext:  currentContext,
			ExecutionContextOptions: executionOptions,
			CurrentState: models.PersistenceWebJobState{
				Mode:                             job.Mode,
				JobType:                          job.JobType,
				Status:                           job.Status,
				DetailedStatus:                   job.DetailedStatus,
				LatestRunStatus:                  job.LatestRunStatus,
				LatestRunTrigger:                 job.LatestRunTrigger,
				RunCommand:                       job.RunCommand,
				ScheduleExpression:               job.ScheduleExpression,
				SchedulerLogsURL:                 job.SchedulerLogsURL,
				ParentAppName:                    job.ParentAppName,
				ParentHostname:                   firstNonEmptyPtr(job.ParentHostname, parentApp.DefaultHostname),
				ParentRuntime:                    parentApp.RuntimeStack,
				ParentPublicNetworkAccess:        parentApp.PublicNetworkAccess,
				ParentIdentityType:               firstNonEmptyPtr(job.ParentIdentityType, parentApp.WorkloadIdentityType),
				ParentAppSettingsCount:           parentApp.AppSettingsCount,
				ParentKeyVaultReferenceCount:     parentApp.KeyVaultReferenceCount,
				ParentConnectionStringCount:      parentApp.ConnectionStringCount,
				StrongestVisibleExecutionContext: strongestContext,
				NearbyThematicNames:              nearbyNames,
			},
			StillUnmapped: persistenceWebJobStillUnmapped(),
			Summary:       persistenceWebJobSummary(job, controlOK, strongestContext, strongestContextHasAzureControl),
			RelatedIDs:    mergeRelatedIDs(job.RelatedIDs, parentApp.RelatedIDs),
		})
	}

	issues := append([]models.Issue{}, webJobs.Issues...)
	issues = append(issues, appServices.Issues...)
	issues = append(issues, managedIdentities.Issues...)
	issues = append(issues, permissions.Issues...)
	issues = append(issues, rbac.Issues...)

	return models.PersistenceWebJobsOutput{
		Metadata:           scopedMetadata(now, request, tenantID, subscriptionID, "persistence"),
		GroupedCommandName: "persistence",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		WebJobs:            rows,
		Issues:             issues,
	}, nil
}

func persistenceWebJobCapabilitySteps(controlOK bool) []models.PersistenceCapabilityStep {
	steps := make([]models.PersistenceCapabilityStep, 0, len(persistenceWebJobSteps))
	for _, step := range persistenceWebJobSteps {
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

func persistenceWebJobAttachedManagedIdentities(
	job models.WebJobAsset,
	parentApp models.AppServiceAsset,
	parentAppVisible bool,
	byAttachment map[string][]models.ManagedIdentity,
) []models.ManagedIdentity {
	parentID := job.ParentAppID
	if parentAppVisible && strings.TrimSpace(parentApp.ID) != "" {
		parentID = parentApp.ID
	}
	items := append([]models.ManagedIdentity{}, byAttachment[parentID]...)
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

func persistenceWebJobExecutionContextOptions(
	job models.WebJobAsset,
	parentApp models.AppServiceAsset,
	parentAppVisible bool,
	attachedManagedIdentities []models.ManagedIdentity,
) []string {
	options := []string{}
	if len(attachedManagedIdentities) > 0 || strings.TrimSpace(stringPtrValue(job.ParentIdentityType)) != "" {
		options = append(options, "managed identity")
	}
	if parentAppVisible {
		if intPtrValue(parentApp.AppSettingsCount) > 0 {
			options = append(options, "parent app settings")
		}
		if intPtrValue(parentApp.KeyVaultReferenceCount) > 0 {
			options = append(options, "Key Vault-backed settings")
		}
		if intPtrValue(parentApp.ConnectionStringCount) > 0 {
			options = append(options, "connection strings")
		}
	}
	if strings.TrimSpace(stringPtrValue(job.RunCommand)) != "" {
		options = append(options, "run command visible")
	}
	return dedupeStrings(options)
}

func persistenceWebJobExecutionContext(
	job models.WebJobAsset,
	parentApp models.AppServiceAsset,
	parentAppVisible bool,
	attachedManagedIdentities []models.ManagedIdentity,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool) {
	if len(attachedManagedIdentities) > 0 || parentAppVisible {
		app := parentApp
		if !parentAppVisible {
			app = models.AppServiceAsset{
				ID:                   job.ParentAppID,
				Name:                 job.ParentAppName,
				WorkloadIdentityType: job.ParentIdentityType,
			}
		}
		return persistenceAppServiceExecutionContext(app, attachedManagedIdentities, permissionsByPrincipal, assignmentsByPrincipal)
	}
	return nil, false
}

func persistenceWebJobStillUnmapped() []string {
	return []string{
		"the current command does not retrieve deployed WebJob package contents or script bodies, so operator intent is not inferred from code here",
		"the current command does not replay runtime-side execution, logs, or job history, so conclusions stop at visible management-plane WebJob posture",
		"the current command does not infer downstream API, data, or automation impact from WebJob names, package names, or runtime hints alone",
	}
}

func persistenceWebJobSummary(
	job models.WebJobAsset,
	controlOK bool,
	strongestContext *models.PersistenceRoleContext,
	strongestContextHasAzureControl bool,
) string {
	if controlOK && persistenceWebJobShowsReusablePosture(job) && strongestContext != nil && strongestContextHasAzureControl {
		return fmt.Sprintf("Current identity can repurpose WebJob '%s' under App Service '%s' as reusable WebJobs persistence, and the strongest visible inherited execution context already carries Azure control.", job.Name, job.ParentAppName)
	}
	if controlOK && persistenceWebJobShowsReusablePosture(job) {
		return fmt.Sprintf("Current identity can repurpose WebJob '%s' under App Service '%s' as reusable WebJobs persistence, with visible mode, rerun, and parent-app context from the current read path.", job.Name, job.ParentAppName)
	}
	if controlOK {
		return fmt.Sprintf("Current identity can build or repurpose WebJob '%s' under App Service '%s', but the current read path only confirms part of the later rerun story.", job.Name, job.ParentAppName)
	}
	if persistenceWebJobShowsReusablePosture(job) {
		return fmt.Sprintf("WebJob '%s' under App Service '%s' already shows durable mode and rerun posture, but the current identity does not yet have a proven path to repurpose it here.", job.Name, job.ParentAppName)
	}
	return fmt.Sprintf("WebJob '%s' under App Service '%s' is visible, but the current identity does not yet have a proven path to turn it into reusable WebJobs persistence.", job.Name, job.ParentAppName)
}

func persistenceWebJobShowsReusablePosture(job models.WebJobAsset) bool {
	if strings.TrimSpace(job.Mode) != "" {
		return true
	}
	if strings.TrimSpace(stringPtrValue(job.RunCommand)) != "" {
		return true
	}
	if strings.TrimSpace(stringPtrValue(job.Status)) != "" {
		return true
	}
	if strings.TrimSpace(stringPtrValue(job.LatestRunTrigger)) != "" {
		return true
	}
	return strings.TrimSpace(stringPtrValue(job.ParentHostname)) != ""
}

func persistenceWebJobNearbyNames(
	jobs []models.WebJobAsset,
	currentJobName string,
) []string {
	seen := map[string]struct{}{}
	candidates := []string{}
	for _, job := range jobs {
		name := strings.TrimSpace(job.Name)
		if name == "" || strings.EqualFold(name, currentJobName) {
			continue
		}
		if persistenceAutomationNameScore(name) == 0 {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
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
