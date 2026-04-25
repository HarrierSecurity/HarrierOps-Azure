package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

const armLogicAppsAPIVersion = "2019-05-01"

func (provider AzureProvider) LogicApps(ctx context.Context, tenant string, subscription string) (LogicAppsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return LogicAppsFacts{}, err
	}

	workflows, err := armListObjects(
		ctx,
		session.credential,
		"/subscriptions/"+session.subscription.ID+"/providers/Microsoft.Logic/workflows",
		armLogicAppsAPIVersion,
	)
	if err != nil {
		return LogicAppsFacts{
			ArtifactIdentityFacts: azureArtifactIdentityFacts(session),
			TenantID:              session.tenantID,
			SubscriptionID:        session.subscription.ID,
			Workflows:             []models.LogicAppWorkflowAsset{},
			Issues:                []models.Issue{issueFromError("logic-apps.workflows", err)},
		}, nil
	}

	rows := make([]models.LogicAppWorkflowAsset, 0, len(workflows))
	issues := []models.Issue{}
	for _, workflow := range workflows {
		workflowID := mapStringValue(workflow, "id")
		hydrated := workflow
		if workflowID != "" {
			detailed, getErr := armGetObject(ctx, session.credential, workflowID, armLogicAppsAPIVersion)
			if getErr != nil {
				issues = append(issues, issueFromError("logic-apps.workflow["+workflowID+"]", getErr))
			} else {
				hydrated = detailed
			}
		}
		rows = append(rows, logicAppWorkflowAsset(hydrated))
	}

	return LogicAppsFacts{
		ArtifactIdentityFacts: azureArtifactIdentityFacts(session),
		TenantID:              session.tenantID,
		SubscriptionID:        session.subscription.ID,
		Workflows:             rows,
		Issues:                issues,
	}, nil
}

func logicAppWorkflowAsset(workflow map[string]any) models.LogicAppWorkflowAsset {
	workflowID := mapStringValue(workflow, "id")
	resourceGroup, workflowName := resourceGroupAndNameFromID(workflowID)
	properties := mapValue(workflow, "properties")
	definition := mapValue(properties, "definition")
	identity := mapValue(workflow, "identity")

	name := firstNonEmpty(mapStringValue(workflow, "name"), workflowName, "unknown")
	identityType := stringPtr(mapStringValue(identity, "type"))
	triggerTypes := logicAppTriggerTypes(definition)
	recurrenceSummary := logicAppRecurrenceSummary(definition)
	downstreamActionKinds := logicAppDownstreamActionKinds(definition)
	connectorReferences := logicAppConnectorReferences(definition)
	parameterNames := logicAppParameterNames(definition)
	resourceReferences := logicAppDownstreamResourceReferences(definition)
	externallyCallableRequestTrigger := logicAppHasRequestTrigger(definition)
	classification := logicAppClassification(externallyCallableRequestTrigger, recurrenceSummary != nil, identityType != nil, len(downstreamActionKinds) > 0)
	identityIDs := logicAppIdentityIDs(workflowID, identity)

	return models.LogicAppWorkflowAsset{
		ID:                               firstNonEmpty(workflowID, "/unknown/"+name),
		Name:                             name,
		Classification:                   classification,
		ResourceGroup:                    resourceGroup,
		Location:                         stringPtr(mapStringValue(workflow, "location")),
		Platform:                         models.StringPtr("Consumption"),
		WorkflowKind:                     stringPtr(mapStringValue(workflow, "kind")),
		State:                            stringPtr(firstNonEmpty(mapStringValue(properties, "state"), mapStringValue(workflow, "state"))),
		IdentityType:                     identityType,
		PrincipalID:                      stringPtr(mapStringValue(identity, "principalId", "principal_id")),
		ClientID:                         stringPtr(mapStringValue(identity, "clientId", "client_id")),
		IdentityIDs:                      identityIDs,
		TriggerCount:                     len(mapValue(definition, "triggers")),
		ActionCount:                      logicAppActionCount(definition),
		BranchCount:                      logicAppBranchCount(definition),
		TriggerTypes:                     triggerTypes,
		ExternallyCallableRequestTrigger: externallyCallableRequestTrigger,
		RecurrenceSummary:                recurrenceSummary,
		DownstreamActionKinds:            downstreamActionKinds,
		ConnectorReferences:              connectorReferences,
		ParameterNames:                   parameterNames,
		DownstreamResourceReferences:     resourceReferences,
		Summary: logicAppOperatorSummary(
			externallyCallableRequestTrigger,
			recurrenceSummary,
			identityType,
			downstreamActionKinds,
			connectorReferences,
			resourceReferences,
			classification,
		),
		RelatedIDs: dedupeStrings(append(append([]string{workflowID}, identityIDs...), resourceReferences...)),
	}
}

func logicAppIdentityIDs(workflowID string, identity map[string]any) []string {
	ids := sortedKeys(mapValue(identity, "userAssignedIdentities", "user_assigned_identities"))
	if identityIncludesType(stringPtr(mapStringValue(identity, "type")), "SystemAssigned") && workflowID != "" {
		ids = append(ids, workflowID+"/identities/system")
	}
	sort.Strings(ids)
	return dedupeStrings(ids)
}

func logicAppTriggerTypes(definition map[string]any) []string {
	triggers := mapValue(definition, "triggers")
	types := []string{}
	for _, rawTrigger := range triggers {
		trigger, ok := rawTrigger.(map[string]any)
		if !ok {
			continue
		}
		triggerType := normalizeLogicAppType(mapStringValue(trigger, "type"))
		if triggerType == "" {
			continue
		}
		types = append(types, triggerType)
	}
	sort.Strings(types)
	return dedupeStrings(types)
}

func logicAppHasRequestTrigger(definition map[string]any) bool {
	triggers := mapValue(definition, "triggers")
	for _, rawTrigger := range triggers {
		trigger, ok := rawTrigger.(map[string]any)
		if !ok {
			continue
		}
		if strings.EqualFold(mapStringValue(trigger, "type"), "Request") {
			return true
		}
	}
	return false
}

func logicAppRecurrenceSummary(definition map[string]any) *string {
	triggers := mapValue(definition, "triggers")
	for _, rawTrigger := range triggers {
		trigger, ok := rawTrigger.(map[string]any)
		if !ok {
			continue
		}
		recurrence := mapValue(trigger, "recurrence")
		if len(recurrence) == 0 && !strings.EqualFold(mapStringValue(trigger, "type"), "Recurrence") {
			continue
		}
		frequency := firstNonEmpty(mapStringValue(recurrence, "frequency"), mapStringValue(trigger, "frequency"))
		interval := mapIntValue(recurrence, "interval")
		if interval == 0 {
			interval = mapIntValue(trigger, "interval")
		}
		if frequency == "" {
			frequency = "Recurring"
		}
		summary := frequency
		if interval > 0 {
			summary = fmt.Sprintf("%s/%d", frequency, interval)
		}

		schedule := mapValue(recurrence, "schedule")
		weekdays := logicAppStringList(schedule, "weekDays", "week_days")
		if len(weekdays) > 0 {
			summary += " weekdays=" + strings.Join(weekdays, ",")
		}
		return models.StringPtr(summary)
	}
	return nil
}

func logicAppDownstreamActionKinds(definition map[string]any) []string {
	categories := []string{}
	logicAppWalkActions(mapValue(definition, "actions"), func(action map[string]any) {
		category := logicAppActionCategory(action)
		if category != "" {
			categories = append(categories, category)
		}
	})
	sort.Strings(categories)
	return dedupeStrings(categories)
}

func logicAppActionCount(definition map[string]any) int {
	count := 0
	logicAppWalkActions(mapValue(definition, "actions"), func(action map[string]any) {
		count++
	})
	return count
}

func logicAppBranchCount(definition map[string]any) int {
	count := 0
	logicAppWalkActions(mapValue(definition, "actions"), func(action map[string]any) {
		if len(mapValue(action, "actions")) > 0 || len(mapValue(mapValue(action, "else"), "actions")) > 0 || len(mapValue(action, "cases")) > 0 {
			count++
		}
	})
	return count
}

func logicAppConnectorReferences(definition map[string]any) []string {
	values := []string{}
	logicAppWalkActions(mapValue(definition, "actions"), func(action map[string]any) {
		values = append(values, logicAppStringMatches(action, func(value string) bool {
			normalized := strings.ToLower(value)
			return strings.Contains(normalized, "/managedapis/") ||
				strings.Contains(normalized, "apiconnections") ||
				strings.Contains(normalized, "serviceproviderconnections")
		})...)
	})
	return dedupeStrings(logicAppShortRefs(values))
}

func logicAppParameterNames(definition map[string]any) []string {
	return sortedKeys(mapValue(definition, "parameters"))
}

func logicAppDownstreamResourceReferences(definition map[string]any) []string {
	values := []string{}
	logicAppWalkActions(mapValue(definition, "actions"), func(action map[string]any) {
		values = append(values, logicAppStringMatches(action, func(value string) bool {
			normalized := strings.ToLower(value)
			return strings.Contains(normalized, "/subscriptions/") && strings.Contains(normalized, "/providers/")
		})...)
	})
	return dedupeStrings(values)
}

func logicAppStringMatches(value any, match func(string) bool) []string {
	values := []string{}
	switch typed := value.(type) {
	case map[string]any:
		for _, child := range typed {
			values = append(values, logicAppStringMatches(child, match)...)
		}
	case []any:
		for _, child := range typed {
			values = append(values, logicAppStringMatches(child, match)...)
		}
	case string:
		if match(typed) && !logicAppLooksSecretString(typed) {
			values = append(values, typed)
		}
	}
	return values
}

func logicAppShortRefs(values []string) []string {
	result := []string{}
	for _, value := range values {
		if name := resourceNameFromID(value); name != "" {
			result = append(result, name)
			continue
		}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func logicAppLooksSecretString(value string) bool {
	normalized := strings.ToLower(value)
	return strings.Contains(normalized, "sig=") ||
		strings.Contains(normalized, "code=") ||
		strings.Contains(normalized, "token=") ||
		strings.Contains(normalized, "secret=") ||
		strings.Contains(normalized, "sharedaccesssignature")
}

func logicAppWalkActions(actions map[string]any, visit func(map[string]any)) {
	for _, rawAction := range actions {
		action, ok := rawAction.(map[string]any)
		if !ok {
			continue
		}
		visit(action)
		logicAppWalkActions(mapValue(action, "actions"), visit)
		logicAppWalkActions(mapValue(mapValue(action, "else"), "actions"), visit)
		for _, rawCase := range mapValue(action, "cases") {
			caseMap, ok := rawCase.(map[string]any)
			if !ok {
				continue
			}
			logicAppWalkActions(mapValue(caseMap, "actions"), visit)
		}
	}
}

func logicAppActionCategory(action map[string]any) string {
	actionType := strings.ToLower(mapStringValue(action, "type"))
	payload := strings.ToLower(mustJSON(action))
	switch {
	case actionType == "function" || strings.Contains(payload, "/functions/"):
		return "function"
	case strings.Contains(payload, "microsoft.automation") || strings.Contains(payload, "automationaccounts"):
		return "automation"
	case strings.Contains(payload, "management.azure.com") || strings.Contains(payload, "microsoft.authorization") || strings.Contains(payload, "microsoft.resources"):
		return "azure-management"
	case strings.Contains(payload, "microsoft.storage") || strings.Contains(payload, ".blob.core.windows.net") || strings.Contains(payload, ".queue.core.windows.net") || strings.Contains(payload, ".table.core.windows.net") || strings.Contains(payload, ".file.core.windows.net"):
		return "storage"
	case strings.Contains(payload, "servicebus") || strings.Contains(payload, "eventhub") || strings.Contains(payload, "eventgrid"):
		return "messaging"
	case actionType == "workflow":
		return "workflow"
	case actionType == "http" || actionType == "httpwebhook":
		return "external-http"
	case actionType == "apiconnection" || actionType == "apiconnectionwebhook":
		return "connector"
	case actionType == "serviceprovider":
		return "service-provider"
	default:
		return ""
	}
}

func logicAppClassification(externallyCallableRequestTrigger bool, hasRecurrence bool, hasIdentity bool, hasDownstream bool) string {
	switch {
	case externallyCallableRequestTrigger || hasRecurrence:
		return "persistence-capable"
	case hasIdentity || hasDownstream:
		return "execution-capable-only"
	default:
		return "visibility-limited"
	}
}

func logicAppOperatorSummary(
	externallyCallableRequestTrigger bool,
	recurrenceSummary *string,
	identityType *string,
	downstreamActionKinds []string,
	connectorReferences []string,
	resourceReferences []string,
	classification string,
) string {
	parts := []string{}
	switch classification {
	case "persistence-capable":
		parts = append(parts, "Visible trigger posture suggests durable re-entry for this workflow.")
	case "execution-capable-only":
		parts = append(parts, "Visible trigger or action posture suggests workflow-driven execution, but not a stronger durable re-entry claim yet.")
	default:
		parts = append(parts, "Workflow is visible from the control plane, but the current read path does not yet show a stronger durable or execution story.")
	}
	if externallyCallableRequestTrigger {
		parts = append(parts, "Request trigger is visible from workflow definition.")
	}
	if recurrenceSummary != nil && *recurrenceSummary != "" {
		parts = append(parts, "Recurrence is visible ("+*recurrenceSummary+").")
	}
	if identityType != nil && *identityType != "" {
		parts = append(parts, "Workflow uses managed identity ("+*identityType+").")
	}
	if len(downstreamActionKinds) > 0 {
		parts = append(parts, "Visible actions touch "+strings.Join(downstreamActionKinds, ", ")+".")
	}
	if len(connectorReferences) > 0 {
		parts = append(parts, "Connector references are visible ("+strings.Join(connectorReferences, ", ")+").")
	}
	if len(resourceReferences) > 0 {
		parts = append(parts, fmt.Sprintf("%d downstream resource reference(s) are visible.", len(resourceReferences)))
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func normalizeLogicAppType(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	var words []string
	current := []rune{}
	for index, r := range trimmed {
		if index > 0 && r >= 'A' && r <= 'Z' && len(current) > 0 {
			words = append(words, strings.ToLower(string(current)))
			current = []rune{r}
			continue
		}
		current = append(current, r)
	}
	if len(current) > 0 {
		words = append(words, strings.ToLower(string(current)))
	}
	return strings.Join(words, "-")
}

func logicAppStringList(input map[string]any, keys ...string) []string {
	values := listValue(input, keys...)
	out := make([]string, 0, len(values))
	for _, value := range values {
		text := strings.TrimSpace(stringValue(value))
		if text == "" {
			continue
		}
		out = append(out, text)
	}
	return dedupeStrings(out)
}

func mustJSON(value any) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(bytes)
}
