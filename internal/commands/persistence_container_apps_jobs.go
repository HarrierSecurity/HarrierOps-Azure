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

type persistenceContainerAppsJobStepDefinition struct {
	Action     string
	APISurface string
}

var persistenceContainerAppsJobSteps = []persistenceContainerAppsJobStepDefinition{
	{Action: "create or reuse job in environment", APISurface: "Microsoft.App/jobs + properties.environmentId"},
	{Action: "point job at image or command", APISurface: "properties.template.containers"},
	{Action: "choose trigger mode", APISurface: "configuration.triggerType / scheduleTriggerConfig / eventTriggerConfig"},
	{Action: "set execution shape and access posture", APISurface: "replica settings / secrets / registries / identity"},
	{Action: "deploy or update stored job definition", APISurface: "Microsoft.App/jobs"},
	{Action: "start or rely on later executions", APISurface: "manual start / schedule / event scale rule"},
	{Action: "preserve or reuse execution path", APISurface: "stored job definition + trigger + image + identity"},
}

func buildPersistenceContainerAppsJobsOutput(
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	contract contracts.PersistenceSurfaceContract,
) (any, error) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	jobsFuture := runGroupedCommandOutput[models.ContainerAppsJobsOutput](group, ctx, request, containerAppsJobsHandler(provider, now), "container-apps-jobs")
	backingFutures := startPersistenceBackingFutures(group, ctx, provider, now, request)

	jobs, err := jobsFuture.wait()
	if err != nil {
		return nil, err
	}
	backing, err := backingFutures.wait(request, jobs.Metadata.TenantID, jobs.Metadata.SubscriptionID, jobs.Issues)

	if err != nil {
		return nil, err
	}
	managedIdentitiesByID := persistenceContainerAppsJobsManagedIdentitiesByID(backing.managedIdentities.Identities)
	items := sortedByLess(jobs.ContainerAppsJobs, containerAppsJobLess)
	rows := make([]models.PersistenceContainerAppsJob, 0, len(items))
	for _, job := range items {
		control, controlOK := persistenceContainerAppsJobControl(job.ID, backing.evidence.currentIdentityAssignments)
		currentContext := persistenceCurrentIdentityContext(backing.evidence.currentIdentity, control, controlOK)
		capabilitySteps := persistenceContainerAppsJobCapabilitySteps(controlOK)
		executionContextOptions := persistenceContainerAppsJobExecutionContextOptions(job)
		strongestContext, strongestContextHasAzureControl := persistenceContainerAppsJobExecutionContext(
			job,
			managedIdentitiesByID,
			backing.evidence.permissionsByPrincipal,
			backing.evidence.assignmentsByPrincipal,
		)
		nearbyNames := persistenceContainerAppsJobNearbyNames(items, job.Name)

		rows = append(rows, models.PersistenceContainerAppsJob{
			ID:                      job.ID,
			Name:                    job.Name,
			ResourceGroup:           job.ResourceGroup,
			Location:                job.Location,
			CapabilitySteps:         capabilitySteps,
			CurrentIdentityContext:  currentContext,
			ExecutionContextOptions: executionContextOptions,
			CurrentState:            persistenceContainerAppsJobStateFromAsset(job, strongestContext, nearbyNames),
			StillUnmapped:           persistenceContainerAppsJobStillUnmapped(),
			Summary:                 persistenceContainerAppsJobSummary(job, controlOK, strongestContext, strongestContextHasAzureControl),
			RelatedIDs:              mergeRelatedIDs(job.RelatedIDs),
		})
	}

	return models.PersistenceContainerAppsJobsOutput{
		Metadata:           withSessionArtifacts(scopedMetadata(now, request, backing.tenantID, backing.subscriptionID, "persistence"), backing.sessionArtifacts),
		GroupedCommandName: "persistence",
		Surface:            contract.Name,
		InputMode:          "live",
		CommandState:       contract.Status,
		Summary:            contract.Summary,
		BackingCommands:    append([]string{}, contract.BackingCommands...),
		ContainerAppsJobs:  rows,
		Issues:             backing.issues,
	}, nil
}

func persistenceContainerAppsJobStateFromAsset(
	job models.ContainerAppsJobAsset,
	strongestContext *models.PersistenceRoleContext,
	nearbyNames []string,
) models.PersistenceContainerAppsJobState {
	return models.PersistenceContainerAppsJobState{
		EnvironmentID:                    job.EnvironmentID,
		TriggerType:                      job.TriggerType,
		ScheduleExpression:               job.ScheduleExpression,
		EventRules:                       append([]models.ContainerAppsJobEventRule{}, job.EventRules...),
		ContainerImages:                  append([]string{}, job.ContainerImages...),
		Command:                          append([]string{}, job.Command...),
		Parallelism:                      job.Parallelism,
		ReplicaCompletionCount:           job.ReplicaCompletionCount,
		ReplicaRetryLimit:                job.ReplicaRetryLimit,
		ReplicaTimeout:                   job.ReplicaTimeout,
		IdentityType:                     job.WorkloadIdentityType,
		WorkloadPrincipalID:              job.WorkloadPrincipalID,
		WorkloadClientID:                 job.WorkloadClientID,
		WorkloadIdentityIDs:              append([]string{}, job.WorkloadIdentityIDs...),
		SecretCount:                      job.SecretCount,
		KeyVaultSecretCount:              job.KeyVaultSecretCount,
		RegistryServers:                  append([]string{}, job.RegistryServers...),
		RegistryIdentityCount:            job.RegistryIdentityCount,
		RegistryPasswordRefCount:         job.RegistryPasswordRefCount,
		StrongestVisibleExecutionContext: strongestContext,
		NearbyThematicNames:              append([]string{}, nearbyNames...),
	}
}

func persistenceContainerAppsJobControl(resourceID string, assignments []models.RoleAssignment) (persistenceCurrentIdentityControl, bool) {
	bestRank := 99
	best := persistenceCurrentIdentityControl{}
	for _, assignment := range assignments {
		if !persistenceRoleAssignmentAllowsNamedOrActionControl(
			assignment,
			"Microsoft.App/jobs/write",
			"owner",
			"contributor",
			"container apps contributor",
			"azure container apps contributor",
		) {
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

func persistenceContainerAppsJobCapabilitySteps(controlOK bool) []models.PersistenceCapabilityStep {
	steps := make([]models.PersistenceCapabilityStep, 0, len(persistenceContainerAppsJobSteps))
	for _, step := range persistenceContainerAppsJobSteps {
		steps = append(steps, models.PersistenceCapabilityStep{
			Action:     step.Action,
			APISurface: step.APISurface,
			Status:     persistenceContainerAppsJobStepStatus(step.Action, controlOK),
		})
	}
	return steps
}

func persistenceContainerAppsJobStepStatus(action string, controlOK bool) string {
	if !controlOK {
		return "not proven"
	}
	switch action {
	case "create or reuse job in environment",
		"point job at image or command",
		"choose trigger mode",
		"set execution shape and access posture",
		"deploy or update stored job definition",
		"preserve or reuse execution path":
		return "yes"
	default:
		return "not proven"
	}
}

func persistenceContainerAppsJobExecutionContextOptions(job models.ContainerAppsJobAsset) []string {
	options := []string{}
	if strings.TrimSpace(stringPtrValue(job.WorkloadIdentityType)) != "" {
		options = append(options, "managed identity")
	}
	if len(job.ContainerImages) > 0 {
		options = append(options, "container image visible")
	}
	if len(job.Command) > 0 {
		options = append(options, "command clue visible")
	}
	switch strings.ToLower(strings.TrimSpace(stringPtrValue(job.TriggerType))) {
	case "schedule", "scheduled":
		options = append(options, "scheduled trigger")
	case "event", "event-driven":
		options = append(options, "event trigger")
	case "manual":
		options = append(options, "manual trigger")
	}
	if intPtrValue(job.SecretCount) > 0 {
		options = append(options, "secret references")
	}
	if intPtrValue(job.KeyVaultSecretCount) > 0 {
		options = append(options, "Key Vault-backed secrets")
	}
	if len(job.RegistryServers) > 0 || intPtrValue(job.RegistryIdentityCount) > 0 || intPtrValue(job.RegistryPasswordRefCount) > 0 {
		options = append(options, "registry posture")
	}
	if strings.TrimSpace(stringPtrValue(job.EnvironmentID)) != "" {
		options = append(options, "environment linkage")
	}
	return dedupeStrings(options)
}

func persistenceContainerAppsJobsManagedIdentitiesByID(identities []models.ManagedIdentity) map[string]models.ManagedIdentity {
	byID := make(map[string]models.ManagedIdentity, len(identities))
	for _, identity := range identities {
		if strings.TrimSpace(identity.ID) == "" {
			continue
		}
		byID[strings.ToLower(strings.TrimSpace(identity.ID))] = identity
	}
	return byID
}

func persistenceContainerAppsJobExecutionContext(
	job models.ContainerAppsJobAsset,
	managedIdentitiesByID map[string]models.ManagedIdentity,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool) {
	candidates := []persistenceContainerAppsJobExecutionCandidate{}
	if strings.TrimSpace(stringPtrValue(job.WorkloadPrincipalID)) != "" {
		candidates = append(candidates, persistenceContainerAppsJobExecutionCandidate{
			name:         persistenceContainerAppsJobSystemIdentityLabel(job.Name),
			principalID:  job.WorkloadPrincipalID,
			identityType: job.WorkloadIdentityType,
		})
	}
	for _, identityID := range job.WorkloadIdentityIDs {
		identity, ok := managedIdentitiesByID[strings.ToLower(strings.TrimSpace(identityID))]
		if !ok || strings.TrimSpace(stringPtrValue(identity.PrincipalID)) == "" {
			continue
		}
		candidates = append(candidates, persistenceContainerAppsJobExecutionCandidate{
			name:         firstNonEmpty(identity.Name, persistenceResourceNameFromID(identity.ID), "container apps job identity"),
			principalID:  identity.PrincipalID,
			identityType: models.StringPtr(firstNonEmpty(identity.IdentityType, "userAssigned")),
		})
	}

	ranked := []persistenceContainerAppsJobRankedExecutionCandidate{}
	for _, candidate := range candidates {
		context, privileged, ok := persistenceContainerAppsJobRoleContext(candidate, permissionsByPrincipal, assignmentsByPrincipal)
		if !ok {
			continue
		}
		ranked = append(ranked, persistenceContainerAppsJobRankedExecutionCandidate{
			candidate:  candidate,
			context:    context,
			privileged: privileged,
		})
	}
	if len(ranked) == 0 {
		return nil, false
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		left := ranked[i]
		right := ranked[j]
		if left.privileged != right.privileged {
			return left.privileged
		}
		leftRoles := len(left.context.RoleNames)
		rightRoles := len(right.context.RoleNames)
		if leftRoles != rightRoles {
			return leftRoles > rightRoles
		}
		return left.candidate.name < right.candidate.name
	})
	return ranked[0].context, ranked[0].privileged
}

type persistenceContainerAppsJobExecutionCandidate struct {
	name         string
	principalID  *string
	identityType *string
}

type persistenceContainerAppsJobRankedExecutionCandidate struct {
	candidate  persistenceContainerAppsJobExecutionCandidate
	context    *models.PersistenceRoleContext
	privileged bool
}

func persistenceContainerAppsJobRoleContext(
	candidate persistenceContainerAppsJobExecutionCandidate,
	permissionsByPrincipal map[string]models.PermissionRow,
	assignmentsByPrincipal map[string][]models.RoleAssignment,
) (*models.PersistenceRoleContext, bool, bool) {
	return persistencePrincipalRoleContext(persistencePrincipalRoleContextOptions{
		fallbackName:           candidate.name,
		kind:                   "container-apps-job-execution-context",
		principalID:            candidate.principalID,
		identityType:           candidate.identityType,
		permissionsByPrincipal: permissionsByPrincipal,
		assignmentsByPrincipal: assignmentsByPrincipal,
		resolvedSummary: func(name string, roleSummary string) string {
			return fmt.Sprintf("The strongest visible execution context here is the Container Apps job identity `%s`, which already holds %s.", name, roleSummary)
		},
		lowerImpactSummary: func(name string) string {
			return fmt.Sprintf("Container Apps job identity `%s` is visible here, but only lower-impact Azure role assignments are visible from current scope.", name)
		},
		unresolvedPrivilegedSummary: func(name string, _ string) string {
			return fmt.Sprintf("Container Apps job identity `%s` is visible here, and raw Azure role-assignment rows for its principal ID suggest stronger Azure control, but that principal is not resolved as a standalone permissions row here.", name)
		},
		noAssignmentsSummary: func(name string) string {
			return fmt.Sprintf("Container Apps job identity `%s` is visible here, but no Azure role-assignment rows are found for its principal ID.", name)
		},
		rbacOnlyCarriesAzureControl: false,
	})
}

func persistenceContainerAppsJobStillUnmapped() []string {
	return []string{
		"the current command does not retrieve raw secret values, registry passwords, or environment variable values, so it reports only safe secret and registry posture",
		"the current command does not inspect container image contents or command behavior, so operator intent is not inferred from image names or command clues alone",
		"the current command does not replay job execution, logs, or execution history, so conclusions stop at stored management-plane job definition and trigger posture",
		"the current command does not inspect downstream queue contents, event payloads, or pipeline data that may drive an event-triggered job run",
	}
}

func persistenceContainerAppsJobSummary(
	job models.ContainerAppsJobAsset,
	controlOK bool,
	strongestContext *models.PersistenceRoleContext,
	strongestContextHasAzureControl bool,
) string {
	if controlOK && persistenceContainerAppsJobShowsReusablePosture(job) && strongestContext != nil && strongestContextHasAzureControl {
		return fmt.Sprintf("Current identity can preserve or reuse Container Apps job '%s' as scheduled or event-driven persistence, and the strongest visible job execution context already carries Azure control.", job.Name)
	}
	if controlOK && persistenceContainerAppsJobShowsReusablePosture(job) {
		return fmt.Sprintf("Current identity can preserve or reuse Container Apps job '%s' as Container Apps Jobs persistence, with visible trigger, image, execution settings, and safe access posture from the current read path.", job.Name)
	}
	if controlOK {
		return fmt.Sprintf("Current identity can deploy or update Container Apps job '%s', but the current read path only confirms part of the later rerun story.", job.Name)
	}
	if persistenceContainerAppsJobShowsReusablePosture(job) {
		return fmt.Sprintf("Container Apps job '%s' already shows a stored trigger, image, and execution context, but the current identity does not yet have a proven path to repurpose it here.", job.Name)
	}
	return fmt.Sprintf("Container Apps job '%s' is visible, but the current identity does not yet have a proven path to turn it into reusable Container Apps Jobs persistence.", job.Name)
}

func persistenceContainerAppsJobShowsReusablePosture(job models.ContainerAppsJobAsset) bool {
	if strings.TrimSpace(stringPtrValue(job.TriggerType)) != "" {
		return true
	}
	if strings.TrimSpace(stringPtrValue(job.ScheduleExpression)) != "" {
		return true
	}
	if len(job.EventRules) > 0 {
		return true
	}
	if len(job.ContainerImages) > 0 || len(job.Command) > 0 {
		return true
	}
	return strings.TrimSpace(stringPtrValue(job.EnvironmentID)) != ""
}

func persistenceContainerAppsJobNearbyNames(
	jobs []models.ContainerAppsJobAsset,
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

func persistenceContainerAppsJobSystemIdentityLabel(jobName string) string {
	name := strings.TrimSpace(jobName)
	if name == "" {
		return "system-assigned identity for Container Apps job"
	}
	return fmt.Sprintf("system-assigned identity for Container Apps job %q", name)
}
