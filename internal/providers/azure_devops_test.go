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
		_, _ = w.Write([]byte(`{"value":[{"name":"Azurefox Proof Lab"}]}`))
	}))
	defer server.Close()

	values, err := devopsListValuesWithClient(context.Background(), "abc123", server.URL, server.Client())
	if err != nil {
		t.Fatalf("devopsListValuesWithClient() error = %v, want nil", err)
	}
	if len(values) != 1 || stringMapValue(values[0], "name") != "Azurefox Proof Lab" {
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

func TestBuildDevopsPipelineAssetPreservesYAMLTargetsAndPermissionProof(t *testing.T) {
	t.Helper()

	project := map[string]any{
		"id":   "80791807-25db-4094-a70c-1ba32a0c7370",
		"name": "Azurefox Proof Lab",
	}
	definition := map[string]any{
		"id":   "3",
		"name": "lab-proof-targeted",
		"path": "\\",
		"repository": map[string]any{
			"id":            "2f15c90d-94e2-4c2f-8730-954ea594c4a1",
			"name":          "lab-proof",
			"type":          "TfsGit",
			"url":           "https://dev.azure.com/azurefox-proof-lab-foxlab/Azurefox%20Proof%20Lab/_git/lab-proof",
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
			"id":   "0b746b31-635e-4b09-890c-af76e7c4638a",
			"name": "af-rg-reader",
			"type": "azurerm",
			"authorization": map[string]any{
				"scheme": "WorkloadIdentityFederation",
				"parameters": map[string]any{
					"serviceprincipalid": "2f4a317c-1ee4-49d6-ac44-56bb6fc56b08",
					"tenantid":           "0bdb6f5d-f8c3-4525-8417-b1fa701482cd",
				},
			},
			"data": map[string]any{
				"spnObjectId":    "808a05cd-7de5-42c6-b06f-f98ebf154b3d",
				"subscriptionId": "b436881a-b87e-44b6-a5d6-7bf7f4bc9c88",
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
		"azurefox-proof-lab-foxlab",
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
