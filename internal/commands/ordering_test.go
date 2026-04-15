package commands

import (
	"testing"

	"harrierops-azure/internal/models"
)

func TestEndpointLessPrefersIPCaseInsensitively(t *testing.T) {
	left := models.EndpointSummary{
		Endpoint:        "52.160.10.20",
		EndpointType:    "IP",
		SourceAssetName: "vm-web-01",
	}
	right := models.EndpointSummary{
		Endpoint:        "app-public-api.azurewebsites.net",
		EndpointType:    "hostname",
		SourceAssetName: "app-public-api",
	}

	if !endpointLess(left, right) {
		t.Fatal("expected IP endpoint to sort ahead of hostname endpoint")
	}
}

func TestExposurePriorityRankNormalizesCase(t *testing.T) {
	if exposurePriorityRank("HIGH") != 0 {
		t.Fatalf("expected HIGH to rank as 0, got %d", exposurePriorityRank("HIGH"))
	}
	if exposurePriorityRank(" Medium ") != 1 {
		t.Fatalf("expected Medium to rank as 1, got %d", exposurePriorityRank(" Medium "))
	}
	if exposurePriorityRank("LoW") != 2 {
		t.Fatalf("expected LoW to rank as 2, got %d", exposurePriorityRank("LoW"))
	}
}

func TestBuildTokensCredentialsFindingsHandlesPriorityCaseAndUnknownSurface(t *testing.T) {
	surfaces := []models.TokenCredentialSurfaceSummary{
		{
			AssetID:           "vm-1",
			AssetName:         "vm-web-01",
			OperatorSignal:    "managed identity token",
			Priority:          "HIGH",
			PubliclyReachable: false,
			RelatedIDs:        []string{"vm-1", "identity-1"},
			Summary:           "Managed identity token path is visible.",
			SurfaceType:       models.TokenCredentialSurfaceManagedIdentityToken,
			NextReviewKind:    models.TokenCredentialReviewManagedIdentityAndPermissions,
		},
		{
			AccessPath:     "future-path",
			AssetID:        "mystery-1",
			AssetName:      "mystery",
			OperatorSignal: "future surface",
			RelatedIDs:     []string{"mystery-1"},
			Summary:        "Unknown credential-bearing surface is visible.",
			SurfaceType:    models.TokenCredentialSurfaceType("future-surface"),
		},
	}

	findings := buildTokensCredentialsFindings(surfaces)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	if findings[0].Severity != "high" {
		t.Fatalf("expected managed identity finding severity to normalize to high, got %q", findings[0].Severity)
	}
	if findings[0].Title != "Publicly reachable workload can mint tokens with managed identity" {
		t.Fatalf("unexpected managed identity title: %q", findings[0].Title)
	}
	if findings[1].ID != "tokens-credentials-unclassified-mystery-1-future-path-future-surface" {
		t.Fatalf("unexpected unknown-surface finding id: %q", findings[1].ID)
	}
	if findings[1].Severity != "low" {
		t.Fatalf("expected unknown surface finding severity low, got %q", findings[1].Severity)
	}
}

func TestArmDeploymentLessPrefersFailedDeployment(t *testing.T) {
	left := models.ArmDeploymentSummary{Name: "failed", ProvisioningState: "Failed"}
	right := models.ArmDeploymentSummary{Name: "succeeded", ProvisioningState: "Succeeded"}

	if !armDeploymentLess(left, right) {
		t.Fatal("expected failed deployment to sort ahead of succeeded deployment")
	}
}

func TestFunctionAppExposurePriorityNormalizesPublicNetworkCase(t *testing.T) {
	enabledLower := "enabled"
	item := models.FunctionAppAsset{PublicNetworkAccess: &enabledLower}

	if !functionAppExposurePriority(item) {
		t.Fatal("expected lowercase enabled public network access to count as exposed")
	}
}

func TestAppServiceHardeningRankDoesNotTreatUnknownValuesAsWeakest(t *testing.T) {
	unknownTLS := "future"
	unknownFTPS := "mystery"
	weakTLS := "1.0"
	weakFTPS := "AllAllowed"

	unknown := appServiceHardeningRank(models.AppServiceAsset{
		MinTLSVersion: &unknownTLS,
		FTPSState:     &unknownFTPS,
	})
	weak := appServiceHardeningRank(models.AppServiceAsset{
		MinTLSVersion: &weakTLS,
		FTPSState:     &weakFTPS,
	})

	if unknown[1] <= weak[1] {
		t.Fatalf("expected unknown TLS version to rank after weakest known TLS, got unknown=%d weak=%d", unknown[1], weak[1])
	}
	if unknown[2] <= weak[2] {
		t.Fatalf("expected unknown FTPS state to rank after weakest known FTPS posture, got unknown=%d weak=%d", unknown[2], weak[2])
	}
}

func TestContainerAppLessPrefersExternalIngressThenIdentity(t *testing.T) {
	trueValue := true
	systemAssigned := "SystemAssigned"
	external := models.ContainerAppAsset{
		Name:                   "external",
		ExternalIngressEnabled: &trueValue,
	}
	identityOnly := models.ContainerAppAsset{
		Name:                 "identity",
		WorkloadIdentityType: &systemAssigned,
	}

	if !containerAppLess(external, identityOnly) {
		t.Fatal("expected external ingress app to sort ahead of internal identity-bearing app")
	}
}

func TestContainerInstanceLessPrefersPublicEndpointThenIdentityThenFQDN(t *testing.T) {
	systemAssigned := "SystemAssigned"
	withPublicIP := models.ContainerInstanceAsset{
		Name:            "public-ip",
		PublicIPAddress: models.StringPtr("52.160.10.30"),
	}
	withIdentity := models.ContainerInstanceAsset{
		Name:                 "identity",
		WorkloadIdentityType: &systemAssigned,
	}
	withFQDN := models.ContainerInstanceAsset{
		Name: "fqdn",
		FQDN: models.StringPtr("aci-public-api.eastus.azurecontainer.io"),
	}

	if !containerInstanceLess(withPublicIP, withIdentity) {
		t.Fatal("expected public endpoint container group to sort ahead of identity-only group")
	}
	if !containerInstanceLess(withIdentity, models.ContainerInstanceAsset{Name: "plain"}) {
		t.Fatal("expected identity-bearing container group to sort ahead of plain group")
	}
	if !containerInstanceLess(withFQDN, models.ContainerInstanceAsset{Name: "plain-2"}) {
		t.Fatal("expected FQDN-bearing container group to sort ahead of plain group")
	}
}
