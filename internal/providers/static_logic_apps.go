package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) LogicApps(_ context.Context, tenant string, subscription string) (LogicAppsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID
	workflows := []models.LogicAppWorkflowAsset{}

	for _, workflow := range staticLogicAppWorkflowFixtures() {
		asset := models.LogicAppWorkflowAsset{
			ID:                               staticLogicAppWorkflowID(subscriptionID, workflow.name),
			Name:                             workflow.name,
			Classification:                   workflow.classification,
			ResourceGroup:                    staticLogicAppsWorkflowResourceGroup,
			Location:                         models.StringPtr(workflow.location),
			Platform:                         models.StringPtr(workflow.platform),
			State:                            models.StringPtr(workflow.state),
			TriggerTypes:                     workflow.triggerTypes,
			RecurrenceSummary:                staticLogicAppRecurrenceSummaryPtr(workflow.recurrenceSummary),
			ExternallyCallableRequestTrigger: workflow.externallyCallableRequestTrigger,
			DownstreamActionKinds:            workflow.downstreamActionKinds,
			Summary:                          workflow.summary,
			RelatedIDs:                       staticLogicAppRelatedIDs(subscriptionID, workflow),
		}
		if workflow.identity != nil {
			asset.IdentityType = models.StringPtr(workflow.identity.workflowIdentityType)
			asset.PrincipalID = models.StringPtr(workflow.identity.principalID)
			asset.ClientID = models.StringPtr(workflow.identity.clientID)
			asset.IdentityIDs = []string{staticLogicAppIdentityID(subscriptionID, *workflow.identity)}
		}
		workflows = append(workflows, asset)
	}

	return LogicAppsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Workflows:      workflows,
		Issues:         []models.Issue{},
	}, nil
}
