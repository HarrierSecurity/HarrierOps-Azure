package providers

import (
	"context"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

const armAutomationAPIVersion = "2024-10-23"

func (provider AzureProvider) Automation(ctx context.Context, tenant string, subscription string) (AutomationFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return AutomationFacts{}, err
	}

	accounts, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.Automation/automationAccounts",
		armAutomationAPIVersion,
	)
	if err != nil {
		return AutomationFacts{
			TenantID:           session.tenantID,
			SubscriptionID:     session.subscription.ID,
			AutomationAccounts: []models.AutomationAccountAsset{},
			Issues:             []models.Issue{issueFromError("automation.accounts", err)},
		}, nil
	}

	accountsOut := make([]models.AutomationAccountAsset, 0, len(accounts))
	issues := []models.Issue{}
	for _, account := range accounts {
		accountsOut = append(accountsOut, automationAccountSummary(ctx, session, account, &issues))
	}

	return AutomationFacts{
		TenantID:           session.tenantID,
		SubscriptionID:     session.subscription.ID,
		AutomationAccounts: accountsOut,
		Issues:             issues,
	}, nil
}

func automationAccountSummary(
	ctx context.Context,
	session azureSession,
	account map[string]any,
	issues *[]models.Issue,
) models.AutomationAccountAsset {
	accountID := mapStringValue(account, "id")
	resourceGroup, accountName := resourceGroupAndNameFromID(accountID)
	hydrated := account
	if accountID != "" && resourceGroup != "" && accountName != "" {
		detailed, err := armGetObject(ctx, session.credential, accountID, armAutomationAPIVersion)
		if err != nil {
			*issues = append(*issues, issueFromError("automation["+resourceGroup+"/"+accountName+"].account", err))
		} else {
			hydrated = detailed
		}
	}

	runbooks := automationListByAccount(ctx, session, accountID, resourceGroup, accountName, "runbooks", "runbook", issues)
	schedules := automationListByAccount(ctx, session, accountID, resourceGroup, accountName, "schedules", "schedule", issues)
	jobSchedules := automationListByAccount(ctx, session, accountID, resourceGroup, accountName, "jobSchedules", "job_schedule", issues)
	webhooks := automationListByAccount(ctx, session, accountID, resourceGroup, accountName, "webhooks", "webhook", issues)
	credentials := automationListByAccount(ctx, session, accountID, resourceGroup, accountName, "credentials", "credential", issues)
	certificates := automationListByAccount(ctx, session, accountID, resourceGroup, accountName, "certificates", "certificate", issues)
	connections := automationListByAccount(ctx, session, accountID, resourceGroup, accountName, "connections", "connection", issues)
	variables := automationListByAccount(ctx, session, accountID, resourceGroup, accountName, "variables", "variable", issues)
	hybridWorkerGroups := automationListByAccount(ctx, session, accountID, resourceGroup, accountName, "hybridRunbookWorkerGroups", "hybrid_runbook_worker_group", issues)

	identity := mapValue(hydrated, "identity")
	properties := mapValue(hydrated, "properties")
	identityType := stringPtr(mapStringValue(identity, "type"))
	principalID := stringPtr(mapStringValue(identity, "principalId", "principal_id"))
	clientID := stringPtr(mapStringValue(identity, "clientId", "client_id"))
	identityIDs := automationIdentityIDs(accountID, identity)
	publishedRunbookCount := automationPublishedRunbookCount(runbooks)
	encryptedVariableCount := automationEncryptedVariableCount(variables)
	startModes := automationStartModes(
		publishedRunbookCount,
		automationCount(schedules),
		automationCount(jobSchedules),
		automationCount(webhooks),
		automationCount(hybridWorkerGroups),
	)
	publishedRunbookNames := automationPublishedRunbookNames(runbooks)
	scheduleRunbookNames := automationRunbookNamesFromTriggers(jobSchedules)
	webhookRunbookNames := automationRunbookNamesFromTriggers(webhooks)
	primaryStartMode, primaryRunbookName := automationPrimaryRunPath(
		startModes,
		publishedRunbookNames,
		scheduleRunbookNames,
		webhookRunbookNames,
		automationCount(hybridWorkerGroups),
	)
	hybridWorkerGroupIDs := automationObjectRawIDs(hybridWorkerGroups)
	triggerJoinIDs := dedupeStrings(append(
		append(
			automationObjectJoinIDs(jobSchedules, "automation-job-schedule"),
			automationObjectJoinIDs(webhooks, "automation-webhook")...,
		),
		automationObjectJoinIDs(hybridWorkerGroups, "automation-hybrid-worker")...,
	))
	identityJoinIDs := automationIdentityJoinIDs(identityIDs, principalID, clientID)
	secretSupportTypes := automationSecretSupportTypes(
		automationCount(credentials),
		automationCount(certificates),
		automationCount(connections),
		encryptedVariableCount,
	)
	secretDependencyIDs := dedupeStrings(append(
		append(
			append(
				automationObjectJoinIDs(credentials, "automation-credential"),
				automationObjectJoinIDs(certificates, "automation-certificate")...,
			),
			automationObjectJoinIDs(connections, "automation-connection")...,
		),
		automationSecretVariableJoinIDs(variables)...,
	))

	name := firstNonEmpty(mapStringValue(hydrated, "name"), resourceNameFromID(accountID), "unknown")
	return models.AutomationAccountAsset{
		ID:                     firstNonEmpty(accountID, "/unknown/"+name),
		Name:                   name,
		ResourceGroup:          resourceGroup,
		Location:               stringPtr(mapStringValue(hydrated, "location")),
		State:                  stringPtr(firstNonEmpty(mapStringValue(hydrated, "state"), mapStringValue(properties, "state"))),
		SKUName:                automationSKUName(hydrated, properties),
		IdentityType:           identityType,
		PrincipalID:            principalID,
		ClientID:               clientID,
		IdentityIDs:            identityIDs,
		RunbookCount:           automationCount(runbooks),
		PublishedRunbookCount:  publishedRunbookCount,
		PublishedRunbookNames:  publishedRunbookNames,
		ScheduleCount:          automationCount(schedules),
		ScheduleDefinitions:    automationScheduleDefinitions(schedules),
		JobScheduleCount:       automationCount(jobSchedules),
		WebhookCount:           automationCount(webhooks),
		HybridWorkerGroupCount: automationCount(hybridWorkerGroups),
		CredentialCount:        automationCount(credentials),
		CertificateCount:       automationCount(certificates),
		ConnectionCount:        automationCount(connections),
		VariableCount:          automationCount(variables),
		EncryptedVariableCount: encryptedVariableCount,
		StartModes:             startModes,
		PrimaryStartMode:       primaryStartMode,
		PrimaryRunbookName:     primaryRunbookName,
		ScheduleRunbookNames:   scheduleRunbookNames,
		WebhookRunbookNames:    webhookRunbookNames,
		HybridWorkerGroupIDs:   hybridWorkerGroupIDs,
		TriggerJoinIDs:         triggerJoinIDs,
		IdentityJoinIDs:        identityJoinIDs,
		SecretSupportTypes:     secretSupportTypes,
		SecretDependencyIDs:    secretDependencyIDs,
		ConsequenceTypes:       automationConsequenceTypes(startModes, secretSupportTypes),
		MissingExecutionPath:   automationMissingExecutionPath(startModes, scheduleRunbookNames, webhookRunbookNames),
		MissingTargetMapping:   true,
		Summary: automationOperatorSummary(
			name,
			identityType,
			automationCount(runbooks),
			publishedRunbookCount,
			automationCount(schedules),
			automationCount(jobSchedules),
			automationCount(webhooks),
			automationCount(hybridWorkerGroups),
			automationCount(credentials),
			automationCount(certificates),
			automationCount(connections),
			automationCount(variables),
			encryptedVariableCount,
		),
		RelatedIDs: automationRelatedIDs(accountID, identityIDs),
	}
}

func automationListByAccount(
	ctx context.Context,
	session azureSession,
	accountID string,
	resourceGroup string,
	accountName string,
	path string,
	scopeSuffix string,
	issues *[]models.Issue,
) []map[string]any {
	if accountID == "" || resourceGroup == "" || accountName == "" {
		return nil
	}
	values, err := armListObjects(ctx, session.credential, accountID+"/"+path, armAutomationAPIVersion)
	if err != nil {
		*issues = append(*issues, issueFromError("automation["+resourceGroup+"/"+accountName+"]."+scopeSuffix, err))
		return nil
	}
	return values
}

func automationIdentityIDs(accountID string, identity map[string]any) []string {
	ids := sortedKeys(mapValue(identity, "userAssignedIdentities", "user_assigned_identities"))
	if identityIncludesType(stringPtr(mapStringValue(identity, "type")), "SystemAssigned") && accountID != "" {
		ids = append(ids, accountID+"/identities/system")
	}
	sort.Strings(ids)
	return dedupeStrings(ids)
}

func automationIdentityJoinIDs(identityIDs []string, principalID *string, clientID *string) []string {
	return dedupeStrings(append(append([]string{}, identityIDs...), stringPtrValue(principalID), stringPtrValue(clientID)))
}

func automationRelatedIDs(accountID string, identityIDs []string) []string {
	return dedupeStrings(append([]string{accountID}, identityIDs...))
}

func automationSKUName(account map[string]any, properties map[string]any) *string {
	return stringPtr(firstNonEmpty(
		mapStringValue(mapValue(account, "sku"), "name"),
		mapStringValue(mapValue(properties, "sku"), "name"),
	))
}

func automationPublishedRunbookCount(runbooks []map[string]any) *int {
	if runbooks == nil {
		return nil
	}
	count := 0
	for _, runbook := range runbooks {
		state := strings.ToLower(firstNonEmpty(
			mapStringValue(mapValue(runbook, "properties"), "state"),
			mapStringValue(runbook, "state"),
		))
		if state == "published" {
			count++
		}
	}
	return intPtr(count)
}

func automationEncryptedVariableCount(variables []map[string]any) *int {
	if variables == nil {
		return nil
	}
	count := 0
	for _, variable := range variables {
		properties := mapValue(variable, "properties")
		if mapBoolValue(properties, "isEncrypted", "is_encrypted") || mapBoolValue(variable, "isEncrypted", "is_encrypted") {
			count++
		}
	}
	return intPtr(count)
}

func automationCount(items []map[string]any) *int {
	if items == nil {
		return nil
	}
	count := len(items)
	return intPtr(count)
}

func automationStartModes(
	publishedRunbookCount *int,
	scheduleCount *int,
	jobScheduleCount *int,
	webhookCount *int,
	hybridWorkerGroupCount *int,
) []string {
	modes := []string{}
	if automationIntValue(scheduleCount) > 0 {
		modes = append(modes, "schedule")
	}
	if automationIntValue(jobScheduleCount) > 0 {
		modes = append(modes, "job-schedule")
	}
	if automationIntValue(webhookCount) > 0 {
		modes = append(modes, "webhook")
	}
	if automationIntValue(hybridWorkerGroupCount) > 0 {
		modes = append(modes, "hybrid-worker")
	}
	if automationIntValue(publishedRunbookCount) > 0 && len(modes) == 0 {
		modes = append(modes, "manual-only")
	}
	return dedupeStrings(modes)
}

func automationPublishedRunbookNames(runbooks []map[string]any) []string {
	if runbooks == nil {
		return []string{}
	}
	names := []string{}
	for _, runbook := range runbooks {
		state := strings.ToLower(firstNonEmpty(
			mapStringValue(mapValue(runbook, "properties"), "state"),
			mapStringValue(runbook, "state"),
		))
		if state != "published" {
			continue
		}
		if name := automationRunbookName(runbook); name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return dedupeStrings(names)
}

func automationRunbookNamesFromTriggers(items []map[string]any) []string {
	if items == nil {
		return []string{}
	}
	names := []string{}
	for _, item := range items {
		if name := automationRunbookName(item); name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return dedupeStrings(names)
}

func automationRunbookName(item map[string]any) string {
	properties := mapValue(item, "properties")
	for _, candidate := range []string{
		mapStringValue(item, "runbookName", "runbook_name"),
		mapStringValue(properties, "runbookName", "runbook_name"),
		mapStringValue(mapValue(item, "runbook"), "name"),
		mapStringValue(mapValue(properties, "runbook"), "name"),
		mapStringValue(item, "name"),
		mapStringValue(properties, "name"),
	} {
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

func automationScheduleDefinitions(schedules []map[string]any) []string {
	if schedules == nil {
		return []string{}
	}
	definitions := []string{}
	for _, schedule := range schedules {
		if definition := automationScheduleDefinition(schedule); definition != "" {
			definitions = append(definitions, definition)
		}
	}
	sort.Strings(definitions)
	return dedupeStrings(definitions)
}

func automationScheduleDefinition(schedule map[string]any) string {
	properties := mapValue(schedule, "properties")
	name := firstNonEmpty(
		mapStringValue(schedule, "name"),
		mapStringValue(properties, "name"),
		"schedule",
	)
	parts := []string{}
	if frequency := firstNonEmpty(mapStringValue(properties, "frequency"), mapStringValue(schedule, "frequency")); frequency != "" {
		parts = append(parts, "frequency="+frequency)
	}
	if interval := firstNonZeroInt(mapIntValue(properties, "interval"), mapIntValue(schedule, "interval")); interval > 0 {
		parts = append(parts, "interval="+stringValue(interval))
	}
	if timezone := firstNonEmpty(mapStringValue(properties, "timeZone", "time_zone"), mapStringValue(schedule, "timeZone", "time_zone")); timezone != "" {
		parts = append(parts, "timezone="+timezone)
	}
	if start := firstNonEmpty(mapStringValue(properties, "startTime", "start_time"), mapStringValue(schedule, "startTime", "start_time")); start != "" {
		parts = append(parts, "start="+start)
	}
	if expiry := firstNonEmpty(mapStringValue(properties, "expiryTime", "expiry_time"), mapStringValue(schedule, "expiryTime", "expiry_time")); expiry != "" {
		parts = append(parts, "expiry="+expiry)
	}
	if enabled, ok := automationOptionalBool(properties, schedule, "isEnabled", "is_enabled"); ok {
		parts = append(parts, "enabled="+boolText(enabled))
	}
	if advanced := automationAdvancedScheduleSummary(mapValue(properties, "advancedSchedule", "advanced_schedule")); advanced != "" {
		parts = append(parts, advanced)
	}
	if len(parts) == 0 {
		return name
	}
	return name + ": " + strings.Join(parts, "; ")
}

func automationAdvancedScheduleSummary(advanced map[string]any) string {
	if len(advanced) == 0 {
		return ""
	}
	parts := []string{}
	if weekDays := automationStringValues(listValue(advanced, "weekDays", "week_days")); len(weekDays) > 0 {
		parts = append(parts, "weekdays="+strings.Join(weekDays, ","))
	}
	if values := automationIntValues(listValue(advanced, "monthDays", "month_days")); len(values) > 0 {
		parts = append(parts, "monthdays="+strings.Join(values, ","))
	}
	if monthlyOccurrences := listValue(advanced, "monthlyOccurrences", "monthly_occurrences"); len(monthlyOccurrences) > 0 {
		parts = append(parts, "monthlyOccurrences="+stringValue(len(monthlyOccurrences)))
	}
	return strings.Join(parts, "; ")
}

func automationStringValues(values []any) []string {
	items := []string{}
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			if trimmed := strings.TrimSpace(typed); trimmed != "" {
				items = append(items, trimmed)
			}
		}
	}
	sort.Strings(items)
	return dedupeStrings(items)
}

func automationIntValues(values []any) []string {
	items := []string{}
	for _, value := range values {
		switch typed := value.(type) {
		case int:
			items = append(items, stringValue(typed))
		case float64:
			items = append(items, stringValue(int(typed)))
		}
	}
	sort.Strings(items)
	return dedupeStrings(items)
}

func automationOptionalBool(primary map[string]any, secondary map[string]any, keys ...string) (bool, bool) {
	for _, input := range []map[string]any{primary, secondary} {
		for _, key := range keys {
			raw, ok := input[key]
			if !ok {
				continue
			}
			switch value := raw.(type) {
			case bool:
				return value, true
			case string:
				switch strings.ToLower(strings.TrimSpace(value)) {
				case "true", "enabled", "yes":
					return true, true
				case "false", "disabled", "no":
					return false, true
				}
			}
		}
	}
	return false, false
}

func firstNonZeroInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func automationPrimaryRunPath(
	startModes []string,
	publishedRunbookNames []string,
	scheduleRunbookNames []string,
	webhookRunbookNames []string,
	hybridWorkerGroupCount *int,
) (*string, *string) {
	if len(webhookRunbookNames) > 0 {
		return stringPtr("webhook"), stringPtr(webhookRunbookNames[0])
	}
	if len(scheduleRunbookNames) > 0 {
		return stringPtr("schedule"), stringPtr(scheduleRunbookNames[0])
	}
	if len(publishedRunbookNames) > 0 {
		mode := "published-runbook"
		for _, startMode := range startModes {
			if startMode == "manual-only" {
				mode = "manual-only"
				break
			}
		}
		return stringPtr(mode), stringPtr(publishedRunbookNames[0])
	}
	if automationIntValue(hybridWorkerGroupCount) > 0 {
		return stringPtr("hybrid-worker"), nil
	}
	return nil, nil
}

func automationObjectRawIDs(items []map[string]any) []string {
	if items == nil {
		return []string{}
	}
	ids := []string{}
	for _, item := range items {
		if id := mapStringValue(item, "id"); id != "" {
			ids = append(ids, id)
		}
	}
	return dedupeStrings(ids)
}

func automationObjectJoinIDs(items []map[string]any, prefix string) []string {
	if items == nil {
		return []string{}
	}
	ids := []string{}
	for _, item := range items {
		if id := mapStringValue(item, "id"); id != "" {
			ids = append(ids, id)
		}
		if name := firstNonEmpty(mapStringValue(item, "name"), mapStringValue(mapValue(item, "properties"), "name")); name != "" {
			ids = append(ids, prefix+":"+name)
		}
	}
	return dedupeStrings(ids)
}

func automationSecretVariableJoinIDs(variables []map[string]any) []string {
	if variables == nil {
		return []string{}
	}
	ids := []string{}
	for _, variable := range variables {
		properties := mapValue(variable, "properties")
		if !(mapBoolValue(properties, "isEncrypted", "is_encrypted") || mapBoolValue(variable, "isEncrypted", "is_encrypted")) {
			continue
		}
		if id := mapStringValue(variable, "id"); id != "" {
			ids = append(ids, id)
		}
		if name := firstNonEmpty(mapStringValue(variable, "name"), mapStringValue(properties, "name")); name != "" {
			ids = append(ids, "automation-variable:"+name)
		}
	}
	return dedupeStrings(ids)
}

func automationSecretSupportTypes(
	credentialCount *int,
	certificateCount *int,
	connectionCount *int,
	encryptedVariableCount *int,
) []string {
	types := []string{}
	if automationIntValue(credentialCount) > 0 {
		types = append(types, "credentials")
	}
	if automationIntValue(certificateCount) > 0 {
		types = append(types, "certificates")
	}
	if automationIntValue(connectionCount) > 0 {
		types = append(types, "connections")
	}
	if automationIntValue(encryptedVariableCount) > 0 {
		types = append(types, "encrypted-variables")
	}
	return types
}

func automationConsequenceTypes(startModes []string, secretSupportTypes []string) []string {
	consequences := []string{}
	if automationContains(startModes, "schedule") {
		consequences = append(consequences, "run-recurring-execution", "reintroduce-config")
	}
	if len(secretSupportTypes) > 0 {
		consequences = append(consequences, "consume-secret-backed-deployment-material")
	}
	if len(consequences) == 0 {
		consequences = append(consequences, "run-recurring-execution")
	}
	return dedupeStrings(consequences)
}

func automationMissingExecutionPath(startModes []string, scheduleRunbookNames []string, webhookRunbookNames []string) bool {
	for _, mode := range startModes {
		if mode != "manual-only" {
			return false
		}
	}
	return len(scheduleRunbookNames) == 0 && len(webhookRunbookNames) == 0
}

func automationOperatorSummary(
	accountName string,
	identityType *string,
	runbookCount *int,
	publishedRunbookCount *int,
	scheduleCount *int,
	jobScheduleCount *int,
	webhookCount *int,
	hybridWorkerGroupCount *int,
	credentialCount *int,
	certificateCount *int,
	connectionCount *int,
	variableCount *int,
	encryptedVariableCount *int,
) string {
	identityClause := "has no managed identity visible from the current read path"
	if stringPtrValue(identityType) != "" {
		identityClause = "uses managed identity (" + stringPtrValue(identityType) + ")"
	}
	return "Automation account '" + accountName + "' " + identityClause + ". " +
		"Visible execution shape: " + automationRunbookClause(runbookCount, publishedRunbookCount) + "; " +
		automationTriggerClause(scheduleCount, jobScheduleCount, webhookCount) + "; " +
		automationWorkerClause(hybridWorkerGroupCount) + ". " +
		"Secure asset posture: " + automationAssetClause(
		credentialCount,
		certificateCount,
		connectionCount,
		variableCount,
		encryptedVariableCount,
	) + "."
}

func automationRunbookClause(runbookCount *int, publishedRunbookCount *int) string {
	if runbookCount == nil {
		return "runbook visibility unreadable"
	}
	if publishedRunbookCount == nil {
		return stringValue(*runbookCount) + " runbook(s)"
	}
	return stringValue(*publishedRunbookCount) + "/" + stringValue(*runbookCount) + " published runbook(s)"
}

func automationTriggerClause(scheduleCount *int, jobScheduleCount *int, webhookCount *int) string {
	parts := []string{
		automationCountOrUnreadable("schedule", scheduleCount),
		automationCountOrUnreadable("job schedule", jobScheduleCount),
		automationCountOrUnreadable("webhook", webhookCount),
	}
	return strings.Join(parts, ", ")
}

func automationWorkerClause(hybridWorkerGroupCount *int) string {
	if hybridWorkerGroupCount == nil {
		return "Hybrid Runbook Worker visibility unreadable"
	}
	if *hybridWorkerGroupCount == 0 {
		return "no Hybrid Runbook Worker groups visible"
	}
	return stringValue(*hybridWorkerGroupCount) + " Hybrid Runbook Worker group(s)"
}

func automationAssetClause(
	credentialCount *int,
	certificateCount *int,
	connectionCount *int,
	variableCount *int,
	encryptedVariableCount *int,
) string {
	parts := []string{
		automationCountOrUnreadable("credentials", credentialCount),
		automationCountOrUnreadable("certificates", certificateCount),
		automationCountOrUnreadable("connections", connectionCount),
	}
	switch {
	case variableCount == nil:
		parts = append(parts, "variables unreadable")
	case encryptedVariableCount == nil:
		parts = append(parts, "variables "+stringValue(*variableCount))
	default:
		parts = append(parts, "variables "+stringValue(*variableCount)+" ("+stringValue(*encryptedVariableCount)+" encrypted)")
	}
	return strings.Join(parts, ", ")
}

func automationCountOrUnreadable(label string, count *int) string {
	if count == nil {
		return label + " unreadable"
	}
	return label + " " + stringValue(*count)
}

func automationIntValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func automationContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
