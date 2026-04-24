package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) DCR(_ context.Context, tenant string, subscription string) (DCRFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID
	workspaceID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.OperationalInsights/workspaces/law-soc-prod"
	eventHubID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.EventHub/namespaces/eh-monitor/authorizationRules/send"
	vmTargetID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01"
	vmssTargetID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-compute/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-batch"
	prodDCRID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.Insights/dataCollectionRules/dcr-prod-host"
	migrationDCRID := "/subscriptions/" + subscriptionID + "/resourceGroups/rg-monitor/providers/Microsoft.Insights/dataCollectionRules/dcr-ama-migration"
	prodAssocID := vmTargetID + "/providers/Microsoft.Insights/dataCollectionRuleAssociations/prod-host-association"
	migrationAssocID := vmssTargetID + "/providers/Microsoft.Insights/dataCollectionRuleAssociations/batch-migration-association"
	transformFingerprint := "31c5a1b7dd8e"
	transformLength := 84

	facts := DCRFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		DCRs: []models.DCRAsset{
			{
				ID:            prodDCRID,
				Name:          "dcr-prod-host",
				ResourceGroup: "rg-monitor",
				Location:      "eastus",
				Description:   models.StringPtr("Production host collection rule with cost-control transform"),
				DataSources: []models.DCRDataSource{
					{Name: "windows-security-events", Type: "windowsEventLogs", Streams: []string{"Microsoft-WindowsEvent"}},
					{Name: "linux-syslog", Type: "syslog", Streams: []string{"Microsoft-Syslog"}},
				},
				DataFlows: []models.DCRDataFlow{
					{
						Streams:                 []string{"Microsoft-WindowsEvent"},
						Destinations:            []string{"soc-workspace"},
						TransformKqlPresent:     true,
						TransformKqlFingerprint: &transformFingerprint,
						TransformKqlLength:      &transformLength,
					},
					{
						Streams:      []string{"Microsoft-Syslog"},
						Destinations: []string{"soc-workspace"},
					},
				},
				Destinations: []models.DCRDestination{
					{Name: "soc-workspace", Type: "logAnalytics", ResourceID: &workspaceID, Detail: models.StringPtr("soc-workspace")},
				},
				Associations: []models.DCRAssociation{
					{
						ID:                   prodAssocID,
						Name:                 "prod-host-association",
						TargetID:             vmTargetID,
						DataCollectionRuleID: &prodDCRID,
						Description:          models.StringPtr("Production host association"),
					},
				},
				DataSourceTypes:     []string{"syslog", "windowsEventLogs"},
				Streams:             []string{"Microsoft-Syslog", "Microsoft-WindowsEvent"},
				HighSignalStreams:   []string{"Microsoft-WindowsEvent", "Microsoft-Syslog"},
				DestinationTypes:    []string{"logAnalytics"},
				TransformationCount: 1,
				AssociationCount:    1,
				RelatedIDs:          []string{prodAssocID, prodDCRID, vmTargetID, workspaceID},
			},
			{
				ID:            migrationDCRID,
				Name:          "dcr-ama-migration",
				ResourceGroup: "rg-monitor",
				Location:      "eastus",
				Description:   models.StringPtr("AMA migration routing for batch fleet"),
				DataSources: []models.DCRDataSource{
					{Name: "perf-default", Type: "performanceCounters", Streams: []string{"Microsoft-Perf"}},
					{Name: "custom-text", Type: "logFiles", Streams: []string{"Custom-AppText_CL"}},
				},
				DataFlows: []models.DCRDataFlow{
					{
						Streams:      []string{"Microsoft-Perf", "Custom-AppText_CL"},
						Destinations: []string{"migration-eventhub"},
					},
				},
				Destinations: []models.DCRDestination{
					{Name: "migration-eventhub", Type: "eventHubs", ResourceID: &eventHubID, Detail: models.StringPtr("monitoring-migration")},
				},
				Associations: []models.DCRAssociation{
					{
						ID:                   migrationAssocID,
						Name:                 "batch-migration-association",
						TargetID:             vmssTargetID,
						DataCollectionRuleID: &migrationDCRID,
						Description:          models.StringPtr("Batch fleet AMA migration"),
					},
				},
				DataSourceTypes:  []string{"logFiles", "performanceCounters"},
				Streams:          []string{"Custom-AppText_CL", "Microsoft-Perf"},
				DestinationTypes: []string{"eventHubs"},
				AssociationCount: 1,
				RelatedIDs:       []string{eventHubID, migrationAssocID, migrationDCRID, vmssTargetID},
			},
		},
		Issues: []models.Issue{},
	}
	for index := range facts.DCRs {
		facts.DCRs[index].Summary = dcrSummary(facts.DCRs[index])
	}
	return facts, nil
}
