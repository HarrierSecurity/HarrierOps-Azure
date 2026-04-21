package providers

import (
	"context"
	"testing"

	"harrierops-azure/internal/models"
)

func TestStorageSummaryPreservesNetworkDefaultActionFromNetworkAcls(t *testing.T) {
	account := map[string]any{
		"id":       "/subscriptions/sub/providers/Microsoft.Storage/storageAccounts/stpublic",
		"name":     "stpublic",
		"location": "centralus",
		"properties": map[string]any{
			"allowBlobPublicAccess": true,
			"networkAcls": map[string]any{
				"defaultAction": "Allow",
			},
			"privateEndpointConnections": []any{},
			"publicNetworkAccess":        "Enabled",
			"supportsHttpsTrafficOnly":   true,
		},
	}

	summary := storageSummary(context.Background(), azureSession{}, account, &[]models.Issue{})

	if summary.NetworkDefaultAction == nil || *summary.NetworkDefaultAction != "Allow" {
		t.Fatalf("storageSummary().NetworkDefaultAction = %v, want Allow", summary.NetworkDefaultAction)
	}
	if !containsStringValue(summary.AnonymousAccessIndicators, "network_default_action=Allow") {
		t.Fatalf("storageSummary().AnonymousAccessIndicators = %v, want network_default_action=Allow", summary.AnonymousAccessIndicators)
	}
}
