package commands

import (
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestBuildDevopsDeploymentRecordKeepsPipelineTargetClueWhenTargetInventoryIsEmpty(t *testing.T) {
	pipeline := models.DevopsPipelineAsset{
		ID:                                   "https://dev.azure.com/contoso/project/_build?definitionId=3",
		Name:                                 "contoso-proof-targeted",
		ProjectName:                          "project",
		AzureServiceConnectionPrincipalIDs:   []string{"sp-object-id"},
		TargetClues:                          []string{"App Service", "App Service: app-public-api-6402b6"},
		ExecutionModes:                       []string{"auto-trigger"},
		TrustedInputRefs:                     []string{"repository:azure-repos:contoso-proof@refs/heads/main"},
		PrimaryTrustedInputRef:               "repository:azure-repos:contoso-proof@refs/heads/main",
		PrimaryInjectionSurface:              "repo-content",
		CurrentOperatorCanContributeSource:   boolPtr(true),
		CurrentOperatorCanEdit:               boolPtr(true),
		CurrentOperatorInjectionSurfaceTypes: []string{"repo-content", "definition-edit"},
		ConsequenceTypes:                     []string{"redeploy-workload", "reintroduce-config"},
	}

	record, ok := buildDevopsDeploymentRecord(
		pipeline,
		"app-services",
		map[string][]deploymentTarget{"app-services": {}},
		nil,
		nil,
		map[string]models.PermissionRow{},
		map[string][]models.RoleTrustSummary{},
		map[string][]models.RoleTrustSummary{},
		map[string]models.KeyVaultAsset{},
		nil,
	)

	if !ok {
		t.Fatal("buildDevopsDeploymentRecord() ok = false, want row preserved")
	}
	if record.TargetResolution != "visibility blocked" {
		t.Fatalf("target_resolution = %q, want visibility blocked", record.TargetResolution)
	}
	if record.LikelyAzureImpact == nil || *record.LikelyAzureImpact != "Azure footprint not yet mapped; visible app service clues only" {
		t.Fatalf("likely_azure_impact = %v, want clue-only impact", record.LikelyAzureImpact)
	}
	if record.TargetVisibility == nil || !strings.Contains(*record.TargetVisibility, "no visible targets") {
		t.Fatalf("target_visibility = %#v, want empty-inventory explanation", record.TargetVisibility)
	}
	if record.Actionability == nil || *record.Actionability != "currently actionable" {
		t.Fatalf("actionability = %v, want currently actionable", record.Actionability)
	}
}
