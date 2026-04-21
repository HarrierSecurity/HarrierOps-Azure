package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"

	"harrierops-azure/internal/models"
)

const (
	devopsScope             = "499b84ac-1321-427f-aa17-267ca6975798/.default"
	devopsGitNamespaceID    = "2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87"
	devopsBuildNamespaceID  = "33344d9c-fc72-4d6f-aba5-fa317101a7e9"
	devopsGraphUsersVersion = "7.1-preview.1"
)

type devopsPermissionSnapshot struct {
	resolved map[string]string
}

func (snapshot devopsPermissionSnapshot) allows(permissionName string) bool {
	value := strings.TrimSpace(snapshot.resolved[permissionName])
	return strings.HasPrefix(value, "Allow")
}

func (provider AzureProvider) Devops(ctx context.Context, tenant string, subscription string, organization string) (DevopsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return DevopsFacts{}, err
	}

	issues := []models.Issue{}
	organization = strings.TrimSpace(organization)
	if organization == "" {
		return DevopsFacts{
			TenantID:       session.tenantID,
			SubscriptionID: session.subscription.ID,
			TokenSource:    session.tokenSource,
			AuthMode:       session.authMode,
			Issues: []models.Issue{{
				Kind:    "unknown",
				Message: "devops: Azure DevOps organization not configured; rerun with --devops-organization or set AZUREFOX_DEVOPS_ORG.",
				Context: map[string]string{"collector": "devops"},
				Scope:   "devops",
			}},
		}, nil
	}

	token, err := session.credential.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{devopsScope}})
	if err != nil {
		return DevopsFacts{}, fmt.Errorf("authenticate Azure DevOps scope: %w", err)
	}

	currentDescriptor := ""
	currentDescriptor, err = devopsCurrentUserDescriptor(ctx, token.Token, organization, session.claims)
	if err != nil {
		issues = append(issues, issueFromError("devops.current_identity_descriptor", err))
	}

	projects, err := devopsListValues(ctx, token.Token, "https://dev.azure.com/"+url.PathEscape(organization)+"/_apis/projects?api-version=7.1&$top=200")
	if err != nil {
		return DevopsFacts{
			TenantID:           session.tenantID,
			SubscriptionID:     session.subscription.ID,
			DevOpsOrganization: organization,
			TokenSource:        session.tokenSource,
			AuthMode:           session.authMode,
			Issues:             []models.Issue{issueFromError("devops.projects", err)},
		}, nil
	}

	allRepositories := []map[string]any{}
	type projectContext struct {
		project          map[string]any
		serviceEndpoints []map[string]any
		variableGroups   []map[string]any
		repositories     []map[string]any
		definitions      []map[string]any
	}
	contexts := make([]projectContext, 0, len(projects))

	for _, project := range projects {
		projectName := stringMapValue(project, "name")
		if projectName == "" {
			continue
		}
		projectPath := url.PathEscape(projectName)

		serviceEndpoints, serviceEndpointErr := devopsListValues(ctx, token.Token, "https://dev.azure.com/"+organization+"/"+projectPath+"/_apis/serviceendpoint/endpoints?api-version=7.1")
		if serviceEndpointErr != nil {
			issues = append(issues, issueFromError("devops["+projectName+"].service_endpoints", serviceEndpointErr))
		}
		variableGroups, variableGroupErr := devopsListValues(ctx, token.Token, "https://dev.azure.com/"+organization+"/"+projectPath+"/_apis/distributedtask/variablegroups?api-version=7.1")
		if variableGroupErr != nil {
			issues = append(issues, issueFromError("devops["+projectName+"].variable_groups", variableGroupErr))
		}
		repositories, repositoryErr := devopsListValues(ctx, token.Token, "https://dev.azure.com/"+organization+"/"+projectPath+"/_apis/git/repositories?includeAllUrls=true&api-version=7.1")
		if repositoryErr != nil {
			issues = append(issues, issueFromError("devops["+projectName+"].repositories", repositoryErr))
		}
		definitions, definitionErr := devopsListValues(ctx, token.Token, "https://dev.azure.com/"+organization+"/"+projectPath+"/_apis/build/definitions?includeAllProperties=true&api-version=7.1&$top=200")
		if definitionErr != nil {
			issues = append(issues, issueFromError("devops["+projectName+"].build_definitions", definitionErr))
		}

		contexts = append(contexts, projectContext{
			project:          project,
			serviceEndpoints: serviceEndpoints,
			variableGroups:   variableGroups,
			repositories:     repositories,
			definitions:      definitions,
		})
		allRepositories = append(allRepositories, repositories...)
	}

	repositoriesByID := map[string]map[string]any{}
	for _, repository := range allRepositories {
		if id := stringMapValue(repository, "id"); id != "" {
			repositoriesByID[strings.ToLower(id)] = repository
		}
	}

	pipelines := []models.DevopsPipelineAsset{}
	repoPermissionsByToken := map[string]devopsPermissionSnapshot{}
	buildPermissionsByToken := map[string]devopsPermissionSnapshot{}
	for _, projectContext := range contexts {
		project := projectContext.project
		projectID := stringMapValue(project, "id")
		projectName := stringMapValue(project, "name")
		projectPath := url.PathEscape(projectName)
		serviceEndpointsByID := map[string]map[string]any{}
		serviceEndpointsByName := map[string]map[string]any{}
		for _, endpoint := range projectContext.serviceEndpoints {
			if id := stringMapValue(endpoint, "id"); id != "" {
				serviceEndpointsByID[strings.ToLower(id)] = endpoint
			}
			if name := stringMapValue(endpoint, "name"); name != "" {
				serviceEndpointsByName[strings.ToLower(name)] = endpoint
			}
		}

		variableGroupsByID := map[string]map[string]any{}
		variableGroupsByName := map[string]map[string]any{}
		for _, group := range projectContext.variableGroups {
			if id := stringMapValue(group, "id"); id != "" {
				variableGroupsByID[strings.ToLower(id)] = group
			}
			if name := stringMapValue(group, "name"); name != "" {
				variableGroupsByName[strings.ToLower(name)] = group
			}
		}

		for _, definition := range projectContext.definitions {
			definitionID := firstNonEmpty(stringMapValue(definition, "id"), mapStringValue(definition, "id"))
			detailedDefinition := definition
			if definitionID != "" {
				definitionURL := "https://dev.azure.com/" + url.PathEscape(organization) + "/" + projectPath + "/_apis/build/definitions/" + url.PathEscape(definitionID) + "?api-version=7.1"
				detailDefinition, detailErr := devopsGetObject(ctx, token.Token, definitionURL)
				if detailErr != nil {
					issues = append(issues, issueFromError("devops["+projectName+"].build_definition["+definitionID+"]", detailErr))
				} else if len(detailDefinition) > 0 {
					detailedDefinition = detailDefinition
				}
			}

			repositoryID := stringPtrValueOrNil(mapValue(detailedDefinition, "repository")["id"])
			var repositoryPermission *devopsPermissionSnapshot
			if currentDescriptor != "" && projectID != "" && repositoryID != nil && *repositoryID != "" {
				repositoryToken := "repoV2/" + projectID + "/" + *repositoryID
				snapshot, ok := repoPermissionsByToken[repositoryToken]
				if !ok {
					snapshot, err = devopsPermissionSnapshotForToken(ctx, token.Token, organization, devopsGitNamespaceID, currentDescriptor, repositoryToken)
					if err != nil {
						issues = append(issues, issueFromError("devops["+projectName+"].repo_permissions["+*repositoryID+"]", err))
					} else {
						repoPermissionsByToken[repositoryToken] = snapshot
					}
				}
				if len(snapshot.resolved) > 0 {
					repositoryPermission = &snapshot
				}
			}

			var buildPermission *devopsPermissionSnapshot
			if currentDescriptor != "" && projectID != "" && definitionID != "" {
				buildToken := projectID + "/" + definitionID
				snapshot, ok := buildPermissionsByToken[buildToken]
				if !ok {
					snapshot, err = devopsPermissionSnapshotForToken(ctx, token.Token, organization, devopsBuildNamespaceID, currentDescriptor, buildToken)
					if err != nil {
						issues = append(issues, issueFromError("devops["+projectName+"].build_permissions["+definitionID+"]", err))
					} else {
						buildPermissionsByToken[buildToken] = snapshot
					}
				}
				if len(snapshot.resolved) > 0 {
					buildPermission = &snapshot
				}
			}

			yamlContent := ""
			yamlPath := mapStringValue(mapValue(detailedDefinition, "process"), "yamlFilename")
			if projectName != "" && repositoryID != nil && *repositoryID != "" && yamlPath != "" {
				repositoryRecord := repositoriesByID[strings.ToLower(*repositoryID)]
				defaultBranch := firstNonEmpty(
					mapStringValue(mapValue(detailedDefinition, "repository"), "defaultBranch"),
					mapStringValue(repositoryRecord, "defaultBranch"),
				)
				content, contentErr := devopsRepositoryFileContent(ctx, token.Token, organization, projectName, *repositoryID, yamlPath, defaultBranch)
				if contentErr != nil {
					issues = append(issues, issueFromError("devops["+projectName+"].repository_file["+yamlPath+"]", contentErr))
				} else {
					yamlContent = content
				}
			}

			pipeline, definitionIssues := buildDevopsPipelineAsset(
				organization,
				project,
				detailedDefinition,
				yamlContent,
				repositoriesByID,
				serviceEndpointsByID,
				serviceEndpointsByName,
				variableGroupsByID,
				variableGroupsByName,
				repositoryPermission,
				buildPermission,
			)
			if pipeline.ID == "" {
				continue
			}
			if len(definitionIssues) > 0 {
				pipeline.PartialRead = true
				issues = append(issues, definitionIssues...)
			}
			pipelines = append(pipelines, pipeline)
		}
	}

	return DevopsFacts{
		TenantID:           session.tenantID,
		SubscriptionID:     session.subscription.ID,
		DevOpsOrganization: organization,
		TokenSource:        session.tokenSource,
		AuthMode:           session.authMode,
		Pipelines:          pipelines,
		Issues:             issues,
	}, nil
}

func devopsListValues(ctx context.Context, bearerToken string, requestURL string) ([]map[string]any, error) {
	return devopsListValuesWithClient(ctx, bearerToken, requestURL, http.DefaultClient)
}

func devopsListValuesWithClient(ctx context.Context, bearerToken string, requestURL string, client *http.Client) ([]map[string]any, error) {
	body, _, err := devopsReadBodyWithClient(ctx, bearerToken, requestURL, client)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Value []map[string]any `json:"value"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode Azure DevOps JSON response: %w", err)
	}
	return payload.Value, nil
}

func devopsRequestHeaders(bearerToken string) map[string]string {
	return map[string]string{
		"Authorization":             "Basic " + base64.StdEncoding.EncodeToString([]byte(":"+bearerToken)),
		"Accept":                    "application/json",
		"X-TFS-FedAuthRedirect":     "Suppress",
		"X-VSS-ForceMsaPassThrough": "true",
	}
}

func devopsBodySnippet(body string) string {
	if len(body) <= 160 {
		return body
	}
	return body[:160] + "..."
}

func devopsCurrentUserDescriptor(ctx context.Context, bearerToken string, organization string, claims map[string]string) (string, error) {
	connectionDescriptor := ""
	connectionData, err := devopsGetObject(
		ctx,
		bearerToken,
		"https://dev.azure.com/"+url.PathEscape(organization)+"/_apis/connectionData?connectOptions=1&lastChangeId=-1&lastChangeId64=-1&api-version=7.1",
	)
	if err == nil {
		if descriptor := mapStringValue(mapValue(connectionData, "authenticatedUser"), "descriptor"); descriptor != "" {
			connectionDescriptor = descriptor
		}
	}

	users, err := devopsListValues(ctx, bearerToken, "https://vssps.dev.azure.com/"+url.PathEscape(organization)+"/_apis/graph/users?api-version="+devopsGraphUsersVersion)
	if err != nil {
		return "", err
	}

	matchValues := []string{
		strings.ToLower(strings.TrimSpace(claims["email"])),
		strings.ToLower(strings.TrimSpace(claims["preferred_username"])),
		strings.ToLower(strings.TrimSpace(claims["upn"])),
		strings.ToLower(strings.TrimSpace(claims["unique_name"])),
		strings.ToLower(strings.TrimSpace(claims["name"])),
	}

	for _, user := range users {
		candidates := []string{
			strings.ToLower(strings.TrimSpace(stringMapValue(user, "mailAddress"))),
			strings.ToLower(strings.TrimSpace(stringMapValue(user, "principalName"))),
			strings.ToLower(strings.TrimSpace(stringMapValue(user, "displayName"))),
		}
		for _, matchValue := range matchValues {
			if matchValue == "" {
				continue
			}
			if slices.Contains(candidates, matchValue) {
				if descriptor := devopsPermissionSubject(user); descriptor != "" {
					return descriptor, nil
				}
			}
		}
	}

	if looksLikeUserClaims(claims) {
		candidates := []map[string]any{}
		for _, user := range users {
			if stringMapValue(user, "descriptor") == "" {
				continue
			}
			if strings.TrimSpace(stringMapValue(user, "mailAddress")) == "" && strings.TrimSpace(stringMapValue(user, "principalName")) == "" {
				continue
			}
			candidates = append(candidates, user)
		}
		if len(candidates) == 1 {
			if descriptor := devopsPermissionSubject(candidates[0]); descriptor != "" {
				return descriptor, nil
			}
		}
	}

	if connectionDescriptor != "" {
		return connectionDescriptor, nil
	}

	return "", fmt.Errorf("resolve Azure DevOps identity descriptor for current principal")
}

func devopsPermissionSubject(user map[string]any) string {
	subjects := []string{}
	if descriptor := stringMapValue(user, "descriptor"); descriptor != "" {
		subjects = append(subjects, descriptor)
	}
	if strings.EqualFold(stringMapValue(user, "origin"), "msa") {
		if originID := stringMapValue(user, "originId"); originID != "" {
			subjects = append(subjects, "Microsoft.IdentityModel.Claims.ClaimsIdentity;"+originID+"@Live.com")
		}
	}
	return strings.Join(sortedUniqueStrings(subjects), ",")
}

func devopsPermissionSnapshotForToken(ctx context.Context, bearerToken string, organization string, namespaceID string, descriptor string, token string) (devopsPermissionSnapshot, error) {
	requestURL := "https://dev.azure.com/" + url.PathEscape(organization) + "/_apis/accesscontrollists/" + namespaceID
	query := url.Values{}
	query.Set("token", token)
	query.Set("descriptors", descriptor)
	query.Set("includeExtendedInfo", "true")
	query.Set("recurse", "false")
	query.Set("api-version", "7.1")

	acls, err := devopsACLValues(ctx, bearerToken, requestURL+"?"+query.Encode())
	if err != nil {
		return devopsPermissionSnapshot{}, err
	}

	resolved := map[string]string{}
	namespacePermissions := devopsPermissionBits(namespaceID)
	for _, acl := range acls {
		for _, ace := range mapValue(acl, "acesDictionary") {
			effectiveAllow := mapIntValue(mapValue(ace, "extendedInfo"), "effectiveAllow")
			if effectiveAllow == 0 {
				effectiveAllow = mapIntValue(ace, "allow")
			}
			for bit, name := range namespacePermissions {
				if effectiveAllow&bit != 0 {
					resolved[name] = "Allow"
				}
			}
		}
	}
	return devopsPermissionSnapshot{resolved: resolved}, nil
}

func devopsPermissionBits(namespaceID string) map[int]string {
	switch namespaceID {
	case devopsGitNamespaceID:
		return map[int]string{
			2: "GenericRead",
			4: "GenericContribute",
		}
	case devopsBuildNamespaceID:
		return map[int]string{
			128:  "QueueBuilds",
			1024: "ViewBuildDefinition",
			2048: "EditBuildDefinition",
		}
	default:
		return map[int]string{}
	}
}

func devopsACLValues(ctx context.Context, bearerToken string, requestURL string) ([]map[string]any, error) {
	body, contentType, err := devopsReadBody(ctx, bearerToken, requestURL)
	if err != nil {
		return nil, err
	}

	var wrapped struct {
		Value []map[string]any `json:"value"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && wrapped.Value != nil {
		return wrapped.Value, nil
	}

	var direct []map[string]any
	if err := json.Unmarshal(body, &direct); err == nil {
		return direct, nil
	}

	return nil, fmt.Errorf("decode Azure DevOps ACL response (%s)", firstNonEmpty(contentType, "unknown content-type"))
}

func devopsGetObject(ctx context.Context, bearerToken string, requestURL string) (map[string]any, error) {
	body, _, err := devopsReadBody(ctx, bearerToken, requestURL)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode Azure DevOps object response: %w", err)
	}
	return payload, nil
}

func devopsRepositoryFileContent(ctx context.Context, bearerToken string, organization string, projectName string, repositoryID string, path string, defaultBranch string) (string, error) {
	requestURL := "https://dev.azure.com/" + url.PathEscape(organization) + "/" + url.PathEscape(projectName) + "/_apis/git/repositories/" + url.PathEscape(repositoryID) + "/items"
	query := url.Values{}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	query.Set("path", path)
	query.Set("includeContent", "true")
	query.Set("api-version", "7.1")
	branch := devopsBranchName(defaultBranch)
	if branch != "" {
		query.Set("versionDescriptor.version", branch)
		query.Set("versionDescriptor.versionType", "branch")
	}

	item, err := devopsGetObject(ctx, bearerToken, requestURL+"?"+query.Encode())
	if err != nil {
		return "", err
	}
	return mapStringValue(item, "content"), nil
}

func devopsBranchName(defaultBranch string) string {
	branch := strings.TrimSpace(defaultBranch)
	branch = strings.TrimPrefix(branch, "refs/heads/")
	if branch == "" {
		return "main"
	}
	return branch
}

func devopsReadBody(ctx context.Context, bearerToken string, requestURL string) ([]byte, string, error) {
	return devopsReadBodyWithClient(ctx, bearerToken, requestURL, http.DefaultClient)
}

func devopsReadBodyWithClient(ctx context.Context, bearerToken string, requestURL string, client *http.Client) ([]byte, string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, "", err
	}
	for key, value := range devopsRequestHeaders(bearerToken) {
		request.Header.Set(key, value)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}
	contentType := strings.ToLower(strings.TrimSpace(response.Header.Get("Content-Type")))
	bodyText := strings.TrimSpace(string(body))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, contentType, fmt.Errorf("%s (%s): %s", response.Status, firstNonEmpty(contentType, "unknown content-type"), bodyText)
	}
	if !strings.Contains(contentType, "json") && !strings.Contains(contentType, "text/plain") {
		return nil, contentType, fmt.Errorf(
			"received non-JSON Azure DevOps response (content-type %q): %s; confirm org access and interactive sign-in for this session",
			firstNonEmpty(contentType, "unknown"),
			devopsBodySnippet(bodyText),
		)
	}
	return body, contentType, nil
}

func buildDevopsPipelineAsset(
	organization string,
	project map[string]any,
	definition map[string]any,
	yamlContent string,
	repositoriesByID map[string]map[string]any,
	serviceEndpointsByID map[string]map[string]any,
	serviceEndpointsByName map[string]map[string]any,
	variableGroupsByID map[string]map[string]any,
	variableGroupsByName map[string]map[string]any,
	repositoryPermission *devopsPermissionSnapshot,
	buildPermission *devopsPermissionSnapshot,
) (models.DevopsPipelineAsset, []models.Issue) {
	definitionID := firstNonEmpty(stringMapValue(definition, "id"), mapStringValue(definition, "id"))
	name := stringMapValue(definition, "name")
	projectID := stringMapValue(project, "id")
	projectName := stringMapValue(project, "name")
	if definitionID == "" || name == "" || projectName == "" {
		return models.DevopsPipelineAsset{}, nil
	}

	repository := mapValue(definition, "repository")
	repositoryID := stringPtrValueOrNil(repository["id"])
	repositoryRecord := map[string]any{}
	if repositoryID != nil && *repositoryID != "" {
		repositoryRecord = repositoriesByID[strings.ToLower(*repositoryID)]
	}
	repositoryName := firstNonEmpty(stringMapValue(repository, "name"), mapStringValue(definition, "repository", "name"))
	repositoryType := firstNonEmpty(stringMapValue(repository, "type"), mapStringValue(definition, "repository", "type"))
	repositoryURL := firstNonEmpty(stringMapValue(repository, "url"), mapStringValue(definition, "repository", "url"))
	defaultBranch := firstNonEmpty(
		stringMapValue(repositoryRecord, "defaultBranch"),
		stringMapValue(repository, "defaultBranch"),
		mapStringValue(definition, "repository", "defaultBranch"),
	)
	repositoryHostType := devopsRepositoryHostType(repositoryType, repositoryURL)
	sourceVisibilityState := devopsSourceVisibilityState(repositoryHostType, repositoryID, repositoriesByID)

	rawDefinition, _ := json.Marshal(definition)
	combinedText := string(rawDefinition)
	if strings.TrimSpace(yamlContent) != "" {
		combinedText += "\n" + yamlContent
	}
	scanText := strings.ToLower(combinedText)

	matchedServiceEndpoints := devopsMatchedServiceEndpoints(scanText, serviceEndpointsByID, serviceEndpointsByName)
	matchedVariableGroups := devopsMatchedVariableGroups(definition, scanText, variableGroupsByID, variableGroupsByName)
	triggerTypes := devopsTriggerTypes(definition)
	executionModes := devopsExecutionModes(triggerTypes)
	targetClues := devopsTargetClues(combinedText, name)
	azureNames, azureTypes, azureSchemes, azureIDs, principalIDs, clientIDs, tenantIDs, subscriptionIDs := devopsEndpointDetails(matchedServiceEndpoints)
	secretVariableNames := devopsSecretVariableNames(definition, matchedVariableGroups)
	keyVaultGroupNames, keyVaultNames := devopsKeyVaultGroups(matchedVariableGroups)
	variableGroupNames := devopsGroupNames(matchedVariableGroups)
	upstreamSources := devopsUpstreamSources(repositoryHostType, repositoryName, defaultBranch, executionModes)
	sourceJoinIDs := devopsSourceJoinIDs(organization, projectID, repositoryID, repositoryName, repositoryURL, repositoryHostType)
	trustedInputs := devopsTrustedInputs(
		repositoryHostType,
		repositoryName,
		defaultBranch,
		sourceVisibilityState,
		sourceJoinIDs,
		executionModes,
		repositoryPermission,
	)
	trustedInputs = append(trustedInputs, devopsExternalURLTrustedInputs(combinedText)...)
	trustedInputTypes := []string{}
	trustedInputRefs := []string{}
	trustedInputJoinIDs := []string{}
	for _, input := range trustedInputs {
		trustedInputTypes = append(trustedInputTypes, input.InputType)
		trustedInputRefs = append(trustedInputRefs, input.Ref)
		trustedInputJoinIDs = append(trustedInputJoinIDs, input.JoinIDs...)
	}

	primaryInjectionSurface := ""
	if len(trustedInputs) > 0 && len(trustedInputs[0].SurfaceTypes) > 0 {
		primaryInjectionSurface = trustedInputs[0].SurfaceTypes[0]
	}
	primaryTrustedInputRef := ""
	if len(trustedInputs) > 0 {
		primaryTrustedInputRef = trustedInputs[0].Ref
	}

	secretSupportTypes := devopsSecretSupportTypes(variableGroupNames, secretVariableNames, keyVaultGroupNames, azureNames)
	secretDependencyIDs := devopsSecretDependencyIDs(matchedVariableGroups)
	riskCues := devopsRiskCues(executionModes, azureNames, secretVariableNames, keyVaultGroupNames)
	consequenceTypes := devopsConsequenceTypes(targetClues, executionModes, secretSupportTypes)
	triggerJoinIDs := devopsTriggerJoinIDs(organization, projectName, definitionID, executionModes, upstreamSources)
	identityJoinIDs := dedupeStrings(append(append([]string{}, azureIDs...), append(principalIDs, clientIDs...)...))
	currentOperatorCanViewDefinition := boolPtr(true)
	if buildPermission != nil {
		currentOperatorCanViewDefinition = boolPtr(buildPermission.allows("ViewBuildDefinition"))
	}
	currentOperatorCanQueue := (*bool)(nil)
	currentOperatorCanEdit := (*bool)(nil)
	if buildPermission != nil {
		currentOperatorCanQueue = boolPtr(buildPermission.allows("QueueBuilds"))
		currentOperatorCanEdit = boolPtr(buildPermission.allows("EditBuildDefinition"))
	}
	currentOperatorCanViewSource := (*bool)(nil)
	currentOperatorCanContributeSource := (*bool)(nil)
	if repositoryPermission != nil {
		canViewSource := repositoryPermission.allows("GenericRead") || repositoryPermission.allows("GenericContribute")
		currentOperatorCanViewSource = boolPtr(canViewSource)
		currentOperatorCanContributeSource = boolPtr(repositoryPermission.allows("GenericContribute"))
	}
	currentOperatorInjectionSurfaceTypes := devopsCurrentOperatorInjectionSurfaceTypes(currentOperatorCanContributeSource, currentOperatorCanEdit, executionModes)
	summary := devopsSummary(
		name,
		projectName,
		trustedInputs,
		executionModes,
		azureNames,
		variableGroupNames,
		secretVariableNames,
		keyVaultNames,
		targetClues,
	)

	return models.DevopsPipelineAsset{
		ID:                                    "https://dev.azure.com/" + organization + "/" + url.PathEscape(projectName) + "/_build?definitionId=" + definitionID,
		DefinitionID:                          definitionID,
		Name:                                  name,
		ProjectID:                             projectID,
		ProjectName:                           projectName,
		Path:                                  firstNonEmpty(stringMapValue(definition, "path"), "\\"),
		RepositoryID:                          repositoryID,
		RepositoryName:                        repositoryName,
		RepositoryType:                        repositoryType,
		RepositoryURL:                         repositoryURL,
		RepositoryHostType:                    repositoryHostType,
		SourceVisibilityState:                 sourceVisibilityState,
		DefaultBranch:                         defaultBranch,
		TriggerTypes:                          triggerTypes,
		VariableGroupNames:                    variableGroupNames,
		SecretVariableCount:                   len(secretVariableNames),
		SecretVariableNames:                   secretVariableNames,
		KeyVaultGroupNames:                    keyVaultGroupNames,
		KeyVaultNames:                         keyVaultNames,
		AzureServiceConnectionNames:           azureNames,
		AzureServiceConnectionTypes:           azureTypes,
		AzureServiceConnectionAuthSchemes:     azureSchemes,
		AzureServiceConnectionIDs:             azureIDs,
		AzureServiceConnectionPrincipalIDs:    principalIDs,
		AzureServiceConnectionClientIDs:       clientIDs,
		AzureServiceConnectionTenantIDs:       tenantIDs,
		AzureServiceConnectionSubscriptionIDs: subscriptionIDs,
		TargetClues:                           targetClues,
		RiskCues:                              riskCues,
		ExecutionModes:                        executionModes,
		UpstreamSources:                       upstreamSources,
		TrustedInputs:                         trustedInputs,
		TrustedInputTypes:                     trustedInputTypes,
		TrustedInputRefs:                      trustedInputRefs,
		TrustedInputJoinIDs:                   dedupeStrings(trustedInputJoinIDs),
		PrimaryInjectionSurface:               primaryInjectionSurface,
		PrimaryTrustedInputRef:                primaryTrustedInputRef,
		SourceJoinIDs:                         sourceJoinIDs,
		TriggerJoinIDs:                        triggerJoinIDs,
		IdentityJoinIDs:                       identityJoinIDs,
		SecretSupportTypes:                    secretSupportTypes,
		SecretDependencyIDs:                   secretDependencyIDs,
		InjectionSurfaceTypes:                 devopsInjectionSurfaceTypes(trustedInputs),
		CurrentOperatorInjectionSurfaceTypes:  currentOperatorInjectionSurfaceTypes,
		EditPathState:                         devopsEditPathState(repositoryName),
		QueuePathState:                        "unknown",
		RerunPathState:                        "unknown",
		ApprovalPathState:                     "unknown",
		CurrentOperatorCanViewDefinition:      currentOperatorCanViewDefinition,
		CurrentOperatorCanQueue:               currentOperatorCanQueue,
		CurrentOperatorCanEdit:                currentOperatorCanEdit,
		CurrentOperatorCanViewSource:          currentOperatorCanViewSource,
		CurrentOperatorCanContributeSource:    currentOperatorCanContributeSource,
		ConsequenceTypes:                      consequenceTypes,
		MissingExecutionPath:                  len(executionModes) == 0,
		MissingInjectionPoint:                 len(currentOperatorInjectionSurfaceTypes) == 0,
		MissingTargetMapping:                  len(targetClues) == 0,
		PartialRead:                           false,
		Summary:                               summary,
		RelatedIDs:                            dedupeStrings(append([]string{"https://dev.azure.com/" + organization + "/" + url.PathEscape(projectName) + "/_build?definitionId=" + definitionID}, append(identityJoinIDs, secretDependencyIDs...)...)),
	}, devopsDefinitionIssues(repositoryName, repositoryHostType, sourceVisibilityState, trustedInputs)
}

func devopsRepositoryHostType(repositoryType string, repositoryURL string) string {
	switch strings.ToLower(repositoryType) {
	case "tfsgit", "azureReposGit":
		return "azure-repos"
	case "github", "githubenterprise":
		return "github"
	}
	if strings.Contains(strings.ToLower(repositoryURL), "github.com") {
		return "github"
	}
	if strings.Contains(strings.ToLower(repositoryURL), "dev.azure.com") {
		return "azure-repos"
	}
	return strings.ToLower(repositoryType)
}

func devopsSourceVisibilityState(repositoryHostType string, repositoryID *string, repositoriesByID map[string]map[string]any) string {
	if repositoryHostType == "github" {
		return "external-reference"
	}
	if repositoryID != nil && *repositoryID != "" {
		if _, ok := repositoriesByID[strings.ToLower(*repositoryID)]; ok {
			return "visible"
		}
		return "inferred-only"
	}
	return "external-reference"
}

func devopsMatchedServiceEndpoints(scanText string, byID map[string]map[string]any, byName map[string]map[string]any) []map[string]any {
	matches := []map[string]any{}
	for key, endpoint := range byID {
		if key != "" && strings.Contains(scanText, key) {
			matches = append(matches, endpoint)
		}
	}
	for key, endpoint := range byName {
		if key != "" && strings.Contains(scanText, key) && !devopsContainsObject(matches, endpoint) {
			matches = append(matches, endpoint)
		}
	}
	return matches
}

func devopsMatchedVariableGroups(definition map[string]any, scanText string, byID map[string]map[string]any, byName map[string]map[string]any) []map[string]any {
	matches := []map[string]any{}
	for _, rawGroup := range sliceValue(definition["variableGroups"]) {
		if groupID := mapStringValue(rawGroup, "id"); groupID != "" {
			if group, ok := byID[strings.ToLower(groupID)]; ok {
				matches = append(matches, group)
			}
		}
		if groupID, ok := rawGroup.(string); ok {
			if group, found := byID[strings.ToLower(groupID)]; found {
				matches = append(matches, group)
			}
		}
	}
	for key, group := range byID {
		if key != "" && strings.Contains(scanText, key) && !devopsContainsObject(matches, group) {
			matches = append(matches, group)
		}
	}
	for key, group := range byName {
		if key != "" && strings.Contains(scanText, key) && !devopsContainsObject(matches, group) {
			matches = append(matches, group)
		}
	}
	return matches
}

func devopsTriggerTypes(definition map[string]any) []string {
	triggers := []string{}
	for _, rawTrigger := range sliceValue(definition["triggers"]) {
		triggerType := firstNonEmpty(mapStringValue(rawTrigger, "triggerType"), mapStringValue(rawTrigger, "settingsSourceType"))
		if triggerType != "" {
			triggers = appendUniqueString(triggers, triggerType)
		}
	}
	return triggers
}

func devopsContainsObject(objects []map[string]any, candidate map[string]any) bool {
	candidateID := stringMapValue(candidate, "id")
	candidateName := stringMapValue(candidate, "name")
	for _, object := range objects {
		if candidateID != "" && candidateID == stringMapValue(object, "id") {
			return true
		}
		if candidateID == "" && candidateName != "" && candidateName == stringMapValue(object, "name") {
			return true
		}
	}
	return false
}

func devopsExecutionModes(triggerTypes []string) []string {
	modes := []string{}
	for _, triggerType := range triggerTypes {
		switch strings.ToLower(triggerType) {
		case "continuousintegration":
			modes = appendUniqueString(modes, "auto-trigger")
		case "pullrequest":
			modes = appendUniqueString(modes, "pull-request")
		case "schedule":
			modes = appendUniqueString(modes, "schedule")
		case "buildcompletion", "artifact":
			modes = appendUniqueString(modes, "artifact")
		default:
			modes = appendUniqueString(modes, strings.ToLower(triggerType))
		}
	}
	return modes
}

func devopsTargetClues(scanText string, name string) []string {
	clues := []string{}
	combined := strings.ToLower(name) + " " + strings.ToLower(scanText)
	if strings.Contains(combined, "aks") || strings.Contains(combined, "kubernetes") || strings.Contains(combined, "helm") || strings.Contains(combined, "kubectl") {
		clues = appendUniqueString(clues, "AKS/Kubernetes")
	}
	if strings.Contains(combined, "appservice") || strings.Contains(combined, "app service") || strings.Contains(combined, "webapp") || strings.Contains(combined, "azurewebapp@") {
		clues = appendUniqueString(clues, "App Service")
	}
	if strings.Contains(combined, "functionapp") || strings.Contains(combined, "function app") || strings.Contains(combined, "functions") {
		clues = appendUniqueString(clues, "Functions")
	}
	if strings.Contains(combined, "terraform") || strings.Contains(combined, "bicep") || strings.Contains(combined, "arm") || strings.Contains(combined, "az deployment") {
		clues = appendUniqueString(clues, "ARM/Bicep/Terraform")
	}
	if strings.Contains(combined, "acr") || strings.Contains(combined, "container") || strings.Contains(combined, "docker") {
		clues = appendUniqueString(clues, "ACR/Containers")
	}
	for _, match := range regexp.MustCompile(`(?im)appname\s*:\s*["']?([a-z0-9._-]+)["']?`).FindAllStringSubmatch(scanText, -1) {
		if len(match) == 2 {
			clues = appendUniqueString(clues, "App Service")
			clues = appendUniqueString(clues, "App Service: "+strings.TrimSpace(match[1]))
		}
	}
	return clues
}

func devopsEndpointDetails(endpoints []map[string]any) ([]string, []string, []string, []string, []string, []string, []string, []string) {
	names := []string{}
	types := []string{}
	schemes := []string{}
	ids := []string{}
	principalIDs := []string{}
	clientIDs := []string{}
	tenantIDs := []string{}
	subscriptionIDs := []string{}
	for _, endpoint := range endpoints {
		names = appendUniqueString(names, stringMapValue(endpoint, "name"))
		types = appendUniqueString(types, stringMapValue(endpoint, "type"))
		ids = appendUniqueString(ids, stringMapValue(endpoint, "id"))
		schemes = appendUniqueString(schemes, mapStringValue(endpoint, "authorization", "scheme"))
		principalID := firstNonEmpty(
			mapStringValue(endpoint, "data", "servicePrincipalObjectId"),
			mapStringValue(endpoint, "data", "spnObjectId"),
		)
		clientID := firstNonEmpty(
			mapStringValue(endpoint, "authorization", "parameters", "serviceprincipalid"),
			mapStringValue(endpoint, "data", "servicePrincipalId"),
		)
		principalIDs = appendUniqueString(principalIDs, principalID)
		clientIDs = appendUniqueString(clientIDs, clientID)
		tenantIDs = appendUniqueString(tenantIDs, firstNonEmpty(mapStringValue(endpoint, "authorization", "parameters", "tenantid"), mapStringValue(endpoint, "data", "subscriptionTenantId")))
		subscriptionIDs = appendUniqueString(subscriptionIDs, mapStringValue(endpoint, "data", "subscriptionId"))
	}
	return sortedUniqueStrings(names), sortedUniqueStrings(types), sortedUniqueStrings(schemes), sortedUniqueStrings(ids), sortedUniqueStrings(principalIDs), sortedUniqueStrings(clientIDs), sortedUniqueStrings(tenantIDs), sortedUniqueStrings(subscriptionIDs)
}

func devopsSecretVariableNames(definition map[string]any, groups []map[string]any) []string {
	names := []string{}
	for key, value := range mapValue(definition, "variables") {
		if mapBoolValue(value, "isSecret") {
			names = appendUniqueString(names, key)
		}
	}
	for _, group := range groups {
		for key, value := range mapValue(group, "variables") {
			if mapBoolValue(value, "isSecret") {
				names = appendUniqueString(names, key)
			}
		}
	}
	return sortedUniqueStrings(names)
}

func devopsKeyVaultGroups(groups []map[string]any) ([]string, []string) {
	groupNames := []string{}
	keyVaultNames := []string{}
	for _, group := range groups {
		groupType := strings.ToLower(stringMapValue(group, "type"))
		if groupType == "azurekeyvault" || groupType == "vstsazurermkeyvault" {
			groupNames = appendUniqueString(groupNames, stringMapValue(group, "name"))
			keyVaultNames = appendUniqueString(keyVaultNames, firstNonEmpty(
				mapStringValue(group, "providerData", "vault"),
				mapStringValue(group, "providerData", "vaultName"),
			))
		}
	}
	return sortedUniqueStrings(groupNames), sortedUniqueStrings(keyVaultNames)
}

func devopsGroupNames(groups []map[string]any) []string {
	names := []string{}
	for _, group := range groups {
		names = appendUniqueString(names, stringMapValue(group, "name"))
	}
	return sortedUniqueStrings(names)
}

func devopsUpstreamSources(repositoryHostType string, repositoryName string, defaultBranch string, executionModes []string) []string {
	sources := []string{}
	if repositoryName != "" {
		branch := firstNonEmpty(defaultBranch, "unknown")
		sources = append(sources, "repo:"+repositoryHostType+":"+repositoryName+"@"+branch)
	}
	for _, mode := range executionModes {
		if mode == "schedule" || mode == "artifact" {
			sources = appendUniqueString(sources, mode)
		}
	}
	return sortedUniqueStrings(sources)
}

func devopsSourceJoinIDs(organization string, projectID string, repositoryID *string, repositoryName string, repositoryURL string, repositoryHostType string) []string {
	joinIDs := []string{}
	if repositoryHostType == "azure-repos" && repositoryID != nil && *repositoryID != "" {
		joinIDs = append(joinIDs, "devops-repo://"+organization+"/"+projectID+"/"+*repositoryID)
	}
	if repositoryURL != "" {
		joinIDs = append(joinIDs, "repo-url://"+repositoryURL)
	}
	if repositoryName != "" && repositoryHostType != "" {
		joinIDs = append(joinIDs, "repo-ref://"+repositoryHostType+"/"+repositoryName)
	}
	return dedupeStrings(joinIDs)
}

func devopsTrustedInputs(
	repositoryHostType string,
	repositoryName string,
	defaultBranch string,
	sourceVisibilityState string,
	sourceJoinIDs []string,
	executionModes []string,
	repositoryPermission *devopsPermissionSnapshot,
) []models.DevopsTrustedInput {
	if repositoryName == "" || repositoryHostType == "" {
		return []models.DevopsTrustedInput{}
	}
	accessState := "exists-only"
	canPoison := false
	permissionDetail := "definition reference only"
	permissionSource := "pipeline-definition"
	evidenceBasis := "definition-reference"
	surfaceTypes := []string{"repo-content"}
	if slices.Contains(executionModes, "pull-request") {
		surfaceTypes = appendUniqueString(surfaceTypes, "pull-request")
	}
	if sourceVisibilityState == "external-reference" {
		permissionDetail = "external reference only"
	}
	if repositoryPermission != nil {
		if repositoryPermission.allows("GenericContribute") {
			accessState = "write"
			canPoison = true
			permissionDetail = "GenericContribute allowed"
			permissionSource = "azure-devops-git-permissions"
			evidenceBasis = "repository-permission"
		} else if repositoryPermission.allows("GenericRead") {
			accessState = "read"
			permissionDetail = "GenericRead allowed"
			permissionSource = "azure-devops-git-permissions"
			evidenceBasis = "repository-permission"
		}
	}
	return []models.DevopsTrustedInput{{
		InputType:                    "repository",
		Ref:                          "repository:" + repositoryHostType + ":" + repositoryName + "@" + firstNonEmpty(defaultBranch, "unknown"),
		VisibilityState:              sourceVisibilityState,
		CurrentOperatorAccessState:   accessState,
		CurrentOperatorCanPoison:     canPoison,
		TrustedInputEvidenceBasis:    evidenceBasis,
		TrustedInputPermissionSource: permissionSource,
		TrustedInputPermissionDetail: permissionDetail,
		SurfaceTypes:                 surfaceTypes,
		JoinIDs:                      sourceJoinIDs,
	}}
}

func devopsExternalURLTrustedInputs(scanText string) []models.DevopsTrustedInput {
	matches := regexp.MustCompile(`https?://[^\s"'\\]+`).FindAllString(scanText, -1)
	if len(matches) == 0 {
		return []models.DevopsTrustedInput{}
	}

	urls := []string{}
	for _, match := range matches {
		candidate := strings.TrimRight(strings.TrimSpace(match), ",.)}]")
		if candidate == "" {
			continue
		}
		if !strings.Contains(candidate, "/_apis/") && !strings.Contains(candidate, ".visualstudio.com/") {
			continue
		}
		urls = appendUniqueString(urls, candidate)
	}
	sort.Strings(urls)

	inputs := make([]models.DevopsTrustedInput, 0, len(urls))
	for _, value := range urls {
		inputs = append(inputs, models.DevopsTrustedInput{
			InputType:                    "external-url",
			Ref:                          "external-url:" + value,
			VisibilityState:              "external-reference",
			CurrentOperatorAccessState:   "exists-only",
			CurrentOperatorCanPoison:     false,
			TrustedInputEvidenceBasis:    "definition-reference",
			TrustedInputPermissionSource: "pipeline-definition",
			TrustedInputPermissionDetail: "external reference only",
			SurfaceTypes:                 []string{"external-download"},
			JoinIDs:                      []string{"url://" + value},
		})
	}
	return inputs
}

func devopsCurrentOperatorInjectionSurfaceTypes(currentOperatorCanContributeSource *bool, currentOperatorCanEdit *bool, executionModes []string) []string {
	surfaces := []string{}
	if currentOperatorCanContributeSource != nil && *currentOperatorCanContributeSource {
		surfaces = appendUniqueString(surfaces, "repo-content")
		if slices.Contains(executionModes, "pull-request") {
			surfaces = appendUniqueString(surfaces, "pull-request")
		}
	}
	if currentOperatorCanEdit != nil && *currentOperatorCanEdit {
		surfaces = appendUniqueString(surfaces, "definition-edit")
	}
	return sortedUniqueStrings(surfaces)
}

func devopsSecretSupportTypes(variableGroupNames []string, secretVariableNames []string, keyVaultGroupNames []string, azureNames []string) []string {
	types := []string{}
	if len(variableGroupNames) > 0 {
		types = append(types, "variable-groups")
	}
	if len(secretVariableNames) > 0 {
		types = append(types, "secret-variables")
		for _, name := range secretVariableNames {
			lowerName := strings.ToLower(name)
			if strings.Contains(lowerName, "sign") || strings.Contains(lowerName, "cert") {
				types = appendUniqueString(types, "signing-keys")
			}
			if strings.Contains(lowerName, "acr") || strings.Contains(lowerName, "docker") || strings.Contains(lowerName, "registry") {
				types = appendUniqueString(types, "registry-creds")
			}
			if strings.Contains(lowerName, "password") || strings.Contains(lowerName, "secret") || strings.Contains(lowerName, "token") {
				types = appendUniqueString(types, "deployment-creds")
			}
		}
	}
	if len(keyVaultGroupNames) > 0 {
		types = append(types, "keyvault-backed-inputs")
	}
	if len(azureNames) > 0 {
		types = appendUniqueString(types, "deployment-creds")
	}
	return sortedUniqueStrings(types)
}

func devopsSecretDependencyIDs(groups []map[string]any) []string {
	dependencies := []string{}
	for _, group := range groups {
		if id := stringMapValue(group, "id"); id != "" {
			dependencies = appendUniqueString(dependencies, id)
		}
		if keyVault := firstNonEmpty(mapStringValue(group, "providerData", "vault"), mapStringValue(group, "providerData", "vaultName")); keyVault != "" {
			dependencies = appendUniqueString(dependencies, "keyvault:"+keyVault)
		}
	}
	return sortedUniqueStrings(dependencies)
}

func devopsRiskCues(executionModes []string, azureNames []string, secretVariableNames []string, keyVaultGroupNames []string) []string {
	cues := []string{}
	if len(executionModes) > 0 && !slices.Contains(executionModes, "pull-request") {
		cues = append(cues, "auto-triggered")
	}
	if len(azureNames) > 0 {
		cues = append(cues, "azure deployment connection")
	}
	if len(secretVariableNames) > 0 {
		cues = append(cues, "secret-bearing variables")
	}
	if len(keyVaultGroupNames) > 0 {
		cues = append(cues, "key vault-backed variables")
	}
	return cues
}

func devopsConsequenceTypes(targetClues []string, executionModes []string, secretSupportTypes []string) []string {
	consequences := []string{}
	for _, clue := range targetClues {
		switch clue {
		case "AKS/Kubernetes", "App Service", "Functions":
			consequences = appendUniqueString(consequences, "redeploy-workload")
		case "ARM/Bicep/Terraform":
			consequences = appendUniqueString(consequences, "modify-infra")
		case "ACR/Containers":
			consequences = appendUniqueString(consequences, "consume-secret-backed-deployment-material")
		}
	}
	if len(executionModes) > 0 {
		consequences = appendUniqueString(consequences, "run-recurring-execution")
		consequences = appendUniqueString(consequences, "reintroduce-config")
	}
	if len(secretSupportTypes) > 0 {
		consequences = appendUniqueString(consequences, "consume-secret-backed-deployment-material")
	}
	return sortedUniqueStrings(consequences)
}

func devopsTriggerJoinIDs(organization string, projectName string, definitionID string, executionModes []string, upstreamSources []string) []string {
	joinIDs := []string{}
	for _, mode := range executionModes {
		joinIDs = append(joinIDs, "devops-trigger://"+organization+"/"+projectName+"/"+definitionID+"/"+url.PathEscape(mode))
	}
	for _, source := range upstreamSources {
		joinIDs = append(joinIDs, "devops-source://"+url.PathEscape(source))
	}
	return dedupeStrings(joinIDs)
}

func devopsInjectionSurfaceTypes(trustedInputs []models.DevopsTrustedInput) []string {
	surfaces := []string{}
	for _, input := range trustedInputs {
		for _, surface := range input.SurfaceTypes {
			surfaces = appendUniqueString(surfaces, surface)
		}
	}
	return sortedUniqueStrings(surfaces)
}

func devopsEditPathState(repositoryName string) string {
	if repositoryName != "" {
		return "repo-backed"
	}
	return "definition-backed"
}

func devopsSummary(
	name string,
	projectName string,
	trustedInputs []models.DevopsTrustedInput,
	executionModes []string,
	azureNames []string,
	variableGroupNames []string,
	secretVariableNames []string,
	keyVaultNames []string,
	targetClues []string,
) string {
	parts := []string{
		"Build definition '" + name + "' in project '" + projectName + "' exposes an Azure change path.",
	}
	if len(trustedInputs) > 0 {
		parts = append(parts, "trusted inputs include "+devopsTrustedInputLabel(trustedInputs[0])+".")
	}
	if len(executionModes) > 0 {
		parts = append(parts, "execution can start through "+strings.Join(executionModes, ", ")+".")
	}
	if len(trustedInputs) > 0 && trustedInputs[0].CurrentOperatorCanPoison {
		parts = append(parts, "current credentials can poison "+strings.Join(trustedInputs[0].SurfaceTypes, ", ")+" through "+devopsTrustedInputLabel(trustedInputs[0])+".")
	}
	if len(azureNames) > 0 {
		parts = append(parts, "uses Azure-facing service connection(s) "+strings.Join(azureNames, ", ")+".")
	}
	if len(variableGroupNames) > 0 {
		parts = append(parts, "references variable group(s) "+strings.Join(variableGroupNames, ", ")+".")
	}
	if len(secretVariableNames) > 0 {
		parts = append(parts, fmt.Sprintf("surfaces %d secret-marked variable name(s).", len(secretVariableNames)))
	}
	if len(keyVaultNames) > 0 {
		parts = append(parts, "pulls from Key Vault-backed variable support ("+strings.Join(keyVaultNames, ", ")+").")
	}
	if len(targetClues) > 0 {
		parts = append(parts, "source clues ground likely Azure impact in "+strings.Join(targetClues, ", ")+".")
	}
	return strings.Join(parts, " ")
}

func sliceValue(input any) []any {
	switch typed := input.(type) {
	case []any:
		return typed
	case []map[string]any:
		values := make([]any, 0, len(typed))
		for _, item := range typed {
			values = append(values, item)
		}
		return values
	default:
		return nil
	}
}

func devopsTrustedInputLabel(input models.DevopsTrustedInput) string {
	if input.Ref == "" {
		return "a trusted input"
	}
	if strings.HasPrefix(input.Ref, input.InputType+":") {
		return strings.TrimPrefix(input.Ref, input.InputType+":")
	}
	return input.Ref
}

func devopsDefinitionIssues(repositoryName string, repositoryHostType string, sourceVisibilityState string, trustedInputs []models.DevopsTrustedInput) []models.Issue {
	if repositoryName == "" || repositoryHostType == "" || sourceVisibilityState == "visible" || len(trustedInputs) == 0 {
		return []models.Issue{}
	}
	return []models.Issue{partialCollectionIssue(
		"devops.trusted_input_proof",
		"trusted-input proof is still partial for "+repositoryHostType+" source "+repositoryName+"; the definition reference is visible, but Azure DevOps source-control permissions are not yet proven here",
		"",
		repositoryName,
	)}
}
