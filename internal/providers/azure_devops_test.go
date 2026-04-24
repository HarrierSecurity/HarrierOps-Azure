package providers

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestDevopsListValuesUsesBasicAuthAndRedirectSuppressHeaders(t *testing.T) {
	t.Helper()

	var sawAuthorization string
	var sawAccept string
	var sawRedirectSuppress string
	var sawPassThrough string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuthorization = r.Header.Get("Authorization")
		sawAccept = r.Header.Get("Accept")
		sawRedirectSuppress = r.Header.Get("X-TFS-FedAuthRedirect")
		sawPassThrough = r.Header.Get("X-VSS-ForceMsaPassThrough")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[{"name":"Contoso Proof Lab"}]}`))
	}))
	defer server.Close()

	values, err := devopsListValuesWithClient(context.Background(), "abc123", server.URL, server.Client())
	if err != nil {
		t.Fatalf("devopsListValuesWithClient() error = %v, want nil", err)
	}
	if len(values) != 1 || stringMapValue(values[0], "name") != "Contoso Proof Lab" {
		t.Fatalf("devopsListValuesWithClient() values = %#v, want one parsed project row", values)
	}

	wantAuthorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(":abc123"))
	if sawAuthorization != wantAuthorization {
		t.Fatalf("Authorization header = %q, want %q", sawAuthorization, wantAuthorization)
	}
	if sawAccept != "application/json" {
		t.Fatalf("Accept header = %q, want application/json", sawAccept)
	}
	if sawRedirectSuppress != "Suppress" {
		t.Fatalf("X-TFS-FedAuthRedirect = %q, want Suppress", sawRedirectSuppress)
	}
	if sawPassThrough != "true" {
		t.Fatalf("X-VSS-ForceMsaPassThrough = %q, want true", sawPassThrough)
	}
}

func TestDevopsListValuesSurfacesNonJSONResponseTruthfully(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body>sign in required</body></html>"))
	}))
	defer server.Close()

	_, err := devopsListValuesWithClient(context.Background(), "abc123", server.URL, server.Client())
	if err == nil {
		t.Fatal("devopsListValuesWithClient() error = nil, want non-JSON response error")
	}
	if !strings.Contains(err.Error(), `received non-JSON Azure DevOps response`) {
		t.Fatalf("error = %q, want non-JSON Azure DevOps response guidance", err)
	}
	if !strings.Contains(err.Error(), `text/html; charset=utf-8`) {
		t.Fatalf("error = %q, want content-type included", err)
	}
	if !strings.Contains(err.Error(), `sign in required`) {
		t.Fatalf("error = %q, want body snippet included", err)
	}
}

func TestDevopsOrganizationsFromFederatedCredentialsUsesAzureSideClues(t *testing.T) {
	t.Helper()

	organizations := devopsOrganizationsFromFederatedCredentials([]map[string]any{
		{
			"subject":     "sc://contoso-devops-lab/Contoso Proof Lab/af-rg-reader",
			"description": "Azure DevOps service connection for https://dev.azure.com/contoso-devops-lab/Azurefox%20Proof%20Lab",
		},
		{
			"subject":     "repo:harrierops/ho-azure:ref:refs/heads/main",
			"description": "Legacy Azure DevOps org: contoso-devops-lab",
		},
		{
			"issuer": "https://vstoken.dev.azure.com/11111111-1111-1111-1111-111111111111",
		},
	})

	if !slices.Equal(organizations, []string{"contoso-devops-lab"}) {
		t.Fatalf("organizations = %#v, want contoso-devops-lab", organizations)
	}
}

func TestDevopsOrganizationsFromFederatedCredentialsKeepsMultipleOrgDecisionExplicit(t *testing.T) {
	t.Helper()

	organizations := devopsOrganizationsFromFederatedCredentials([]map[string]any{
		{"description": "Azure DevOps organization=contoso"},
		{"description": "https://fabrikam.visualstudio.com/DefaultCollection"},
	})

	if !slices.Equal(organizations, []string{"contoso", "fabrikam"}) {
		t.Fatalf("organizations = %#v, want contoso and fabrikam", organizations)
	}
}

func TestGraphListObjectsLimitStopsInsidePageWithoutFollowingMoreLinks(t *testing.T) {
	t.Helper()

	requests := 0
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/first":
			_, _ = w.Write([]byte(`{"value":[{"id":"1"},{"id":"2"},{"id":"3"}],"@odata.nextLink":"` + server.URL + `/next"}`))
		case "/next":
			t.Fatal("graphListObjectsLimit followed @odata.nextLink after reaching the configured cap")
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	items, truncated, err := graphListObjectsLimit(context.Background(), "token", server.URL+"/first", 2)
	if err != nil {
		t.Fatalf("graphListObjectsLimit() error = %v, want nil", err)
	}
	if len(items) != 2 {
		t.Fatalf("graphListObjectsLimit() returned %d items, want 2", len(items))
	}
	if !truncated {
		t.Fatal("graphListObjectsLimit() truncated = false, want true")
	}
	if requests != 1 {
		t.Fatalf("graphListObjectsLimit() made %d requests, want 1", requests)
	}
}

func TestBuildDevopsPipelineAssetPreservesYAMLTargetsAndPermissionProof(t *testing.T) {
	t.Helper()

	project := map[string]any{
		"id":   "80791807-25db-4094-a70c-1ba32a0c7370",
		"name": "Contoso Proof Lab",
	}
	definition := map[string]any{
		"id":   "3",
		"name": "contoso-proof-targeted",
		"path": "\\",
		"repository": map[string]any{
			"id":            "2f15c90d-94e2-4c2f-8730-954ea594c4a1",
			"name":          "contoso-proof",
			"type":          "TfsGit",
			"url":           "https://dev.azure.com/contoso-devops-lab/Azurefox%20Proof%20Lab/_git/contoso-proof",
			"defaultBranch": "main",
		},
		"triggers": []any{
			map[string]any{"triggerType": "continuousIntegration"},
		},
		"process": map[string]any{
			"type":         2,
			"yamlFilename": "pipelines/named-target.yml",
		},
	}
	repositoriesByID := map[string]map[string]any{
		"2f15c90d-94e2-4c2f-8730-954ea594c4a1": {
			"defaultBranch": "refs/heads/main",
		},
	}
	serviceEndpointsByName := map[string]map[string]any{
		"af-rg-reader": {
			"id":   "55555555-5555-5555-5555-555555555555",
			"name": "af-rg-reader",
			"type": "azurerm",
			"authorization": map[string]any{
				"scheme": "WorkloadIdentityFederation",
				"parameters": map[string]any{
					"serviceprincipalid": "66666666-6666-6666-6666-666666666666",
					"tenantid":           "11111111-1111-1111-1111-111111111111",
				},
			},
			"data": map[string]any{
				"spnObjectId":    "77777777-7777-7777-7777-777777777777",
				"subscriptionId": "22222222-2222-2222-2222-222222222222",
			},
		},
	}
	repositoryPermission := &devopsPermissionSnapshot{resolved: map[string]string{
		"GenericRead":       "Allow",
		"GenericContribute": "Allow",
	}}
	buildPermission := &devopsPermissionSnapshot{resolved: map[string]string{
		"ViewBuildDefinition": "Allow",
		"EditBuildDefinition": "Allow",
		"QueueBuilds":         "Allow",
	}}
	yamlContent := `trigger: none
pool:
  vmImage: ubuntu-latest
steps:
- task: AzureWebApp@1
  inputs:
    azureSubscription: af-rg-reader
    appName: app-public-api-6402b6
`

	pipeline, issues := buildDevopsPipelineAsset(
		"contoso-devops-lab",
		project,
		definition,
		yamlContent,
		repositoriesByID,
		map[string]map[string]any{},
		serviceEndpointsByName,
		map[string]map[string]any{},
		map[string]map[string]any{},
		repositoryPermission,
		buildPermission,
	)

	if len(issues) != 0 {
		t.Fatalf("buildDevopsPipelineAsset() issues = %#v, want none", issues)
	}
	if !slices.Contains(pipeline.AzureServiceConnectionNames, "af-rg-reader") {
		t.Fatalf("azure_service_connection_names = %#v, want af-rg-reader", pipeline.AzureServiceConnectionNames)
	}
	if !slices.Equal(pipeline.AzureServiceConnectionAuthSchemes, []string{"WorkloadIdentityFederation"}) {
		t.Fatalf("azure_service_connection_auth_schemes = %#v, want WorkloadIdentityFederation", pipeline.AzureServiceConnectionAuthSchemes)
	}
	if !slices.Equal(pipeline.AzureServiceConnectionPrincipalIDs, []string{"77777777-7777-7777-7777-777777777777"}) {
		t.Fatalf("azure_service_connection_principal_ids = %#v, want spnObjectId", pipeline.AzureServiceConnectionPrincipalIDs)
	}
	if !slices.Equal(pipeline.AzureServiceConnectionClientIDs, []string{"66666666-6666-6666-6666-666666666666"}) {
		t.Fatalf("azure_service_connection_client_ids = %#v, want serviceprincipalid", pipeline.AzureServiceConnectionClientIDs)
	}
	if !slices.Equal(pipeline.AzureServiceConnectionTenantIDs, []string{"11111111-1111-1111-1111-111111111111"}) {
		t.Fatalf("azure_service_connection_tenant_ids = %#v, want tenantid", pipeline.AzureServiceConnectionTenantIDs)
	}
	if !slices.Equal(pipeline.AzureServiceConnectionSubscriptionIDs, []string{"22222222-2222-2222-2222-222222222222"}) {
		t.Fatalf("azure_service_connection_subscription_ids = %#v, want subscriptionId", pipeline.AzureServiceConnectionSubscriptionIDs)
	}
	if !slices.Contains(pipeline.TargetClues, "App Service") || !slices.Contains(pipeline.TargetClues, "App Service: app-public-api-6402b6") {
		t.Fatalf("target_clues = %#v, want generic and exact App Service clues", pipeline.TargetClues)
	}
	if pipeline.CurrentOperatorCanContributeSource == nil || !*pipeline.CurrentOperatorCanContributeSource {
		t.Fatalf("current_operator_can_contribute_source = %#v, want true", pipeline.CurrentOperatorCanContributeSource)
	}
	if pipeline.CurrentOperatorCanEdit == nil || !*pipeline.CurrentOperatorCanEdit {
		t.Fatalf("current_operator_can_edit = %#v, want true", pipeline.CurrentOperatorCanEdit)
	}
	if pipeline.CurrentOperatorCanQueue == nil || !*pipeline.CurrentOperatorCanQueue {
		t.Fatalf("current_operator_can_queue = %#v, want true", pipeline.CurrentOperatorCanQueue)
	}
	if pipeline.MissingTargetMapping {
		t.Fatal("missing_target_mapping = true, want false")
	}
	if pipeline.MissingInjectionPoint {
		t.Fatal("missing_injection_point = true, want false")
	}
	if pipeline.TrustedInputs[0].CurrentOperatorAccessState != "write" {
		t.Fatalf("trusted_inputs[0].current_operator_access_state = %q, want write", pipeline.TrustedInputs[0].CurrentOperatorAccessState)
	}
	if !pipeline.TrustedInputs[0].CurrentOperatorCanPoison {
		t.Fatal("trusted_inputs[0].current_operator_can_poison = false, want true")
	}
	if !slices.Contains(pipeline.CurrentOperatorInjectionSurfaceTypes, "repo-content") || !slices.Contains(pipeline.CurrentOperatorInjectionSurfaceTypes, "definition-edit") {
		t.Fatalf("current_operator_injection_surface_types = %#v, want repo-content and definition-edit", pipeline.CurrentOperatorInjectionSurfaceTypes)
	}
}
