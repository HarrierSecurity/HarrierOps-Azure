package providers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"

	"harrierops-azure/internal/models"
)

func (provider AzureProvider) WebJobs(ctx context.Context, tenant string, subscription string) (WebJobsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return WebJobsFacts{}, err
	}

	webAppsState, err := provider.webAppsState(session)
	if err != nil {
		return WebJobsFacts{}, fmt.Errorf("build web apps client: %w", err)
	}

	rows := []models.WebJobAsset{}
	issues := []models.Issue{}
	apps, listErr := webAppsState.list(ctx)
	if listErr != nil {
		issues = append(issues, issueFromError("webjobs.web_apps", listErr))
	}
	for _, app := range apps {
		if app.assetKind != "AppService" {
			continue
		}

		webJobs, err := webAppsState.webJobAssets(ctx, app)
		if err != nil {
			issues = append(issues, issueFromError("webjobs["+app.resourceGroup+"/"+app.name+"]", err))
			continue
		}
		rows = append(rows, webJobs...)
	}

	return WebJobsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		WebJobs:        rows,
		Issues:         issues,
	}, nil
}

func (state *liveWebAppsState) webJobAssets(ctx context.Context, app *liveWebAppResource) ([]models.WebJobAsset, error) {
	if app.resourceGroup == "" || app.name == "" {
		return []models.WebJobAsset{}, nil
	}

	rows := []models.WebJobAsset{}

	continuousPager := state.client.NewListContinuousWebJobsPager(app.resourceGroup, app.name, nil)
	for continuousPager.More() {
		page, err := continuousPager.NextPage(ctx)
		if err != nil {
			if webJobsListNotFound(err) {
				break
			}
			return rows, err
		}
		for _, webJob := range page.Value {
			if webJob == nil {
				continue
			}
			rows = append(rows, continuousWebJobAsset(app, webJob))
		}
	}

	triggeredPager := state.client.NewListTriggeredWebJobsPager(app.resourceGroup, app.name, nil)
	for triggeredPager.More() {
		page, err := triggeredPager.NextPage(ctx)
		if err != nil {
			if webJobsListNotFound(err) {
				break
			}
			return rows, err
		}
		for _, webJob := range page.Value {
			if webJob == nil {
				continue
			}
			rows = append(rows, triggeredWebJobAsset(app, webJob))
		}
	}

	return rows, nil
}

func webJobsListNotFound(err error) bool {
	var responseErr *azcore.ResponseError
	return errors.As(err, &responseErr) && responseErr.StatusCode == 404
}

func continuousWebJobAsset(app *liveWebAppResource, webJob *armappservice.ContinuousWebJob) models.WebJobAsset {
	properties := &armappservice.ContinuousWebJobProperties{}
	if webJob.Properties != nil {
		properties = webJob.Properties
	}
	status := continuousWebJobStatusString(properties.Status)
	jobType := webJobTypeString(properties.WebJobType)
	runCommand := properties.RunCommand
	detailedStatus := properties.DetailedStatus
	parentAppID, parentAppName, parentHostname, parentIdentityType, parentIdentityIDs, relatedIDs := webJobParentContext(app)

	return models.WebJobAsset{
		DetailedStatus:     detailedStatus,
		ID:                 firstNonEmpty(stringPtrValue(webJob.ID), parentAppID+"/continuouswebjobs/"+stringPtrValue(webJob.Name)),
		JobType:            jobType,
		Location:           mapStringValue(app.appMap, "location"),
		Mode:               "continuous",
		Name:               firstNonEmpty(stringPtrValue(webJob.Name), "unknown"),
		ParentAppID:        parentAppID,
		ParentAppName:      parentAppName,
		ParentHostname:     parentHostname,
		ParentIdentityIDs:  parentIdentityIDs,
		ParentIdentityType: parentIdentityType,
		RelatedIDs:         dedupeStrings(append(relatedIDs, stringPtrValue(webJob.ID))),
		ResourceGroup:      app.resourceGroup,
		RunCommand:         runCommand,
		Status:             status,
		Summary: webJobSummary(
			firstNonEmpty(stringPtrValue(webJob.Name), "unknown"),
			"continuous",
			parentAppName,
			parentHostname,
			parentIdentityType,
			status,
			runCommand,
			nil,
			nil,
		),
	}
}

func triggeredWebJobAsset(app *liveWebAppResource, webJob *armappservice.TriggeredWebJob) models.WebJobAsset {
	properties := &armappservice.TriggeredWebJobProperties{}
	if webJob.Properties != nil {
		properties = webJob.Properties
	}
	latestRunStatus := (*string)(nil)
	latestRunTrigger := (*string)(nil)
	if properties.LatestRun != nil {
		latestRunStatus = triggeredWebJobStatusString(properties.LatestRun.Status)
		latestRunTrigger = properties.LatestRun.Trigger
	}
	mode := webJobTriggeredMode(properties.SchedulerLogsURL, latestRunTrigger, properties.Settings)
	scheduleExpression := webJobScheduleExpression(properties.Settings)
	jobType := webJobTypeString(properties.WebJobType)
	runCommand := properties.RunCommand
	parentAppID, parentAppName, parentHostname, parentIdentityType, parentIdentityIDs, relatedIDs := webJobParentContext(app)

	return models.WebJobAsset{
		ID:                 firstNonEmpty(stringPtrValue(webJob.ID), parentAppID+"/triggeredwebjobs/"+stringPtrValue(webJob.Name)),
		JobType:            jobType,
		LatestRunStatus:    latestRunStatus,
		LatestRunTrigger:   latestRunTrigger,
		Location:           mapStringValue(app.appMap, "location"),
		Mode:               mode,
		Name:               firstNonEmpty(stringPtrValue(webJob.Name), "unknown"),
		ParentAppID:        parentAppID,
		ParentAppName:      parentAppName,
		ParentHostname:     parentHostname,
		ParentIdentityIDs:  parentIdentityIDs,
		ParentIdentityType: parentIdentityType,
		RelatedIDs:         dedupeStrings(append(relatedIDs, stringPtrValue(webJob.ID))),
		ResourceGroup:      app.resourceGroup,
		RunCommand:         runCommand,
		ScheduleExpression: scheduleExpression,
		SchedulerLogsURL:   properties.SchedulerLogsURL,
		Status:             latestRunStatus,
		Summary: webJobSummary(
			firstNonEmpty(stringPtrValue(webJob.Name), "unknown"),
			mode,
			parentAppName,
			parentHostname,
			parentIdentityType,
			latestRunStatus,
			runCommand,
			scheduleExpression,
			latestRunTrigger,
		),
	}
}

func webJobParentContext(app *liveWebAppResource) (string, string, *string, *string, []string, []string) {
	parentAppID := mapStringValue(app.appMap, "id")
	parentAppName := firstNonEmpty(mapStringValue(app.appMap, "name"), "unknown")
	parentHostname := webAppDefaultHostname(app.appMap)
	identity := mapValue(app.appMap, "identity")
	parentIdentityType := stringPtr(mapStringValue(identity, "type"))
	parentIdentityIDs := sortedKeys(mapValue(identity, "userAssignedIdentities"), "user_assigned_identities")
	parentPrincipalID := stringPtr(mapStringValue(identity, "principalId", "principal_id"))

	relatedIDs := append([]string{parentAppID, stringPtrValue(parentPrincipalID)}, parentIdentityIDs...)
	return parentAppID, parentAppName, parentHostname, parentIdentityType, parentIdentityIDs, dedupeStrings(relatedIDs)
}

func webJobTriggeredMode(schedulerLogsURL *string, latestRunTrigger *string, settings map[string]interface{}) string {
	if stringPtrValue(schedulerLogsURL) != "" {
		return "scheduled"
	}
	if webJobSettingsHintScheduled(settings) {
		return "scheduled"
	}
	if webJobTriggerLooksScheduled(latestRunTrigger) {
		return "scheduled"
	}
	return "triggered/manual"
}

func continuousWebJobStatusString(status *armappservice.ContinuousWebJobStatus) *string {
	if status == nil {
		return nil
	}
	return stringPtr(string(*status))
}

func triggeredWebJobStatusString(status *armappservice.TriggeredWebJobStatus) *string {
	if status == nil {
		return nil
	}
	return stringPtr(string(*status))
}

func webJobTypeString(jobType *armappservice.WebJobType) *string {
	if jobType == nil {
		return nil
	}
	return stringPtr(string(*jobType))
}

func webJobSettingsHintScheduled(settings map[string]interface{}) bool {
	for key, value := range settings {
		loweredKey := strings.ToLower(strings.TrimSpace(key))
		if strings.Contains(loweredKey, "schedule") || strings.Contains(loweredKey, "cron") {
			return true
		}
		loweredValue := strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
		if strings.Contains(loweredValue, "schedule") || strings.Contains(loweredValue, "cron") || strings.Contains(loweredValue, "ncrontab") {
			return true
		}
	}
	return false
}

func webJobScheduleExpression(settings map[string]interface{}) *string {
	if len(settings) == 0 {
		return nil
	}

	type candidate struct {
		rank  int
		value string
	}

	best := candidate{rank: -1}
	for key, rawValue := range settings {
		loweredKey := strings.ToLower(strings.TrimSpace(key))
		value := strings.TrimSpace(fmt.Sprint(rawValue))
		if value == "" {
			continue
		}

		rank := -1
		switch {
		case loweredKey == "schedule":
			rank = 4
		case loweredKey == "cron":
			rank = 4
		case loweredKey == "ncrontab":
			rank = 4
		case strings.Contains(loweredKey, "ncrontab"):
			rank = 3
		case strings.Contains(loweredKey, "cron"):
			rank = 3
		case strings.Contains(loweredKey, "schedule"):
			rank = 2
		default:
			continue
		}

		if rank > best.rank {
			best = candidate{rank: rank, value: value}
		}
	}

	if best.rank < 0 {
		return nil
	}
	return stringPtr(best.value)
}

func webJobTriggerLooksScheduled(trigger *string) bool {
	lowered := strings.ToLower(strings.TrimSpace(stringPtrValue(trigger)))
	return strings.Contains(lowered, "schedule") || strings.Contains(lowered, "timer") || strings.Contains(lowered, "cron")
}

func webJobSummary(name string, mode string, parentAppName string, parentHostname *string, parentIdentityType *string, status *string, runCommand *string, scheduleExpression *string, latestRunTrigger *string) string {
	modePhrase := "is a webjob"
	switch mode {
	case "continuous":
		modePhrase = "is a continuous WebJob"
	case "scheduled":
		modePhrase = "is a scheduled WebJob"
	case "triggered/manual":
		modePhrase = "is a triggered/manual WebJob"
	}

	parentPhrase := "parent App Service '" + parentAppName + "' has no default hostname visible from the current read path"
	if parentHostname != nil && *parentHostname != "" {
		parentPhrase = "parent App Service '" + parentAppName + "' publishes hostname '" + *parentHostname + "'"
	}

	identityPhrase := "no managed identity is visible on the parent app"
	if parentIdentityType != nil && *parentIdentityType != "" {
		identityPhrase = "the parent app uses managed identity (" + *parentIdentityType + ")"
	}

	executionParts := []string{}
	if status != nil && *status != "" {
		executionParts = append(executionParts, "status "+*status)
	}
	if latestRunTrigger != nil && *latestRunTrigger != "" {
		executionParts = append(executionParts, "latest visible trigger "+*latestRunTrigger)
	}
	if scheduleExpression != nil && *scheduleExpression != "" {
		executionParts = append(executionParts, "schedule '"+*scheduleExpression+"'")
	}
	if runCommand != nil && *runCommand != "" {
		executionParts = append(executionParts, "run command '"+*runCommand+"'")
	}
	executionPhrase := "execution detail is limited from the current read path"
	if len(executionParts) > 0 {
		executionPhrase = strings.Join(executionParts, ", ")
	}

	return "WebJob '" + name + "' " + modePhrase + " under App Service '" + parentAppName + "'; " + executionPhrase + "; " + parentPhrase + "; " + identityPhrase + "."
}
