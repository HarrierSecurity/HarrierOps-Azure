package providers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"harrierops-azure/internal/models"
)

//go:embed fixtures/devops.json
var staticDevopsFixtureJSON []byte

func (StaticProvider) Devops(_ context.Context, tenant string, subscription string, organization string) (DevopsFacts, error) {
	session := staticFixtureSession(tenant, subscription)

	var fixture struct {
		Pipelines []models.DevopsPipelineAsset `json:"pipelines"`
	}
	if err := json.Unmarshal(staticDevopsFixtureJSON, &fixture); err != nil {
		return DevopsFacts{}, fmt.Errorf("decode static devops fixture: %w", err)
	}

	return DevopsFacts{
		TenantID:           session.TenantID,
		SubscriptionID:     session.Subscription.ID,
		DevOpsOrganization: organization,
		Pipelines:          fixture.Pipelines,
		Issues:             []models.Issue{},
	}, nil
}
