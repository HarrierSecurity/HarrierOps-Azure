package providers

import "testing"

func TestContainerAppsJobSummaryExtractsScheduleJobFields(t *testing.T) {
	summary := containerAppsJobSummary(map[string]any{
		"id":       "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/jobs/nightly-reconcile",
		"name":     "nightly-reconcile",
		"location": "eastus",
		"identity": map[string]any{
			"type":        "SystemAssigned",
			"principalId": "principal-id",
			"clientId":    "client-id",
		},
		"properties": map[string]any{
			"environmentId": "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/managedEnvironments/env-prod",
			"configuration": map[string]any{
				"triggerType":       "Schedule",
				"replicaRetryLimit": 3,
				"replicaTimeout":    1800,
				"scheduleTriggerConfig": map[string]any{
					"cronExpression":         "0 3 * * *",
					"parallelism":            1,
					"replicaCompletionCount": 1,
				},
				"secrets": []any{
					map[string]any{"name": "plain-secret", "value": "redacted"},
					map[string]any{"name": "kv-secret", "keyVaultUrl": "https://kv.vault.azure.net/secrets/job"},
				},
				"registries": []any{
					map[string]any{"server": "ghcr.io", "identity": "system"},
				},
			},
			"template": map[string]any{
				"containers": []any{
					map[string]any{
						"name":    "reconcile",
						"image":   "ghcr.io/harrierops/jobs/reconcile:1.4.2",
						"command": []any{"/app/reconcile"},
						"args":    []any{"--tenant", "prod"},
					},
				},
			},
		},
	})

	if summary.ScheduleExpression == nil || *summary.ScheduleExpression != "0 3 * * *" {
		t.Fatalf("ScheduleExpression = %v, want cron expression", summary.ScheduleExpression)
	}
	if summary.Parallelism == nil || *summary.Parallelism != 1 {
		t.Fatalf("Parallelism = %v, want 1", summary.Parallelism)
	}
	if summary.ReplicaTimeout == nil || *summary.ReplicaTimeout != 1800 {
		t.Fatalf("ReplicaTimeout = %v, want 1800", summary.ReplicaTimeout)
	}
	if len(summary.ContainerImages) != 1 || summary.ContainerImages[0] != "ghcr.io/harrierops/jobs/reconcile:1.4.2" {
		t.Fatalf("ContainerImages = %#v, want image clue", summary.ContainerImages)
	}
	if len(summary.Command) != 1 || summary.Command[0] != "reconcile: /app/reconcile --tenant prod" {
		t.Fatalf("Command = %#v, want command clue", summary.Command)
	}
	if summary.SecretCount == nil || *summary.SecretCount != 2 {
		t.Fatalf("SecretCount = %v, want 2", summary.SecretCount)
	}
	if summary.KeyVaultSecretCount == nil || *summary.KeyVaultSecretCount != 1 {
		t.Fatalf("KeyVaultSecretCount = %v, want 1", summary.KeyVaultSecretCount)
	}
	if summary.WorkloadPrincipalID == nil || *summary.WorkloadPrincipalID != "principal-id" {
		t.Fatalf("WorkloadPrincipalID = %v, want principal-id", summary.WorkloadPrincipalID)
	}
}

func TestContainerAppsJobSummaryExtractsEventRulesSafely(t *testing.T) {
	userAssignedID := "/subscriptions/sub/resourceGroups/rg-id/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-job"
	summary := containerAppsJobSummary(map[string]any{
		"id":       "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/jobs/queue-drain",
		"name":     "queue-drain",
		"location": "eastus",
		"identity": map[string]any{
			"type": "SystemAssigned, UserAssigned",
			"userAssignedIdentities": map[string]any{
				userAssignedID: map[string]any{},
			},
		},
		"properties": map[string]any{
			"environmentId": "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/managedEnvironments/env-internal",
			"configuration": map[string]any{
				"triggerType":       "Event",
				"replicaRetryLimit": 3,
				"replicaTimeout":    3600,
				"eventTriggerConfig": map[string]any{
					"parallelism":            2,
					"replicaCompletionCount": 2,
					"scale": map[string]any{
						"rules": []any{
							map[string]any{
								"name":     "orders-queue",
								"type":     "azure-queue",
								"identity": userAssignedID,
								"auth": []any{
									map[string]any{"secretRef": "queue-connection", "triggerParameter": "connection"},
								},
							},
						},
					},
				},
				"registries": []any{
					map[string]any{
						"server":            "contoso.azurecr.io",
						"passwordSecretRef": "acr-password",
					},
				},
			},
			"template": map[string]any{
				"containers": []any{
					map[string]any{"image": "contoso.azurecr.io/jobs/queue-drain:2026.04"},
				},
			},
		},
	})

	if len(summary.EventRules) != 1 {
		t.Fatalf("EventRules = %#v, want one rule", summary.EventRules)
	}
	rule := summary.EventRules[0]
	if rule.Name != "orders-queue" || rule.Type != "azure-queue" {
		t.Fatalf("EventRules[0] = %#v, want named azure-queue rule", rule)
	}
	if len(rule.AuthSecretRefs) != 1 || rule.AuthSecretRefs[0] != "queue-connection" {
		t.Fatalf("AuthSecretRefs = %#v, want safe auth reference name", rule.AuthSecretRefs)
	}
	if rule.Identity == nil || *rule.Identity != userAssignedID {
		t.Fatalf("Event rule identity = %v, want user-assigned identity id", rule.Identity)
	}
	if summary.ScheduleExpression != nil {
		t.Fatalf("ScheduleExpression = %v, want nil for event job", summary.ScheduleExpression)
	}
	if summary.RegistryPasswordRefCount == nil || *summary.RegistryPasswordRefCount != 1 {
		t.Fatalf("RegistryPasswordRefCount = %v, want 1", summary.RegistryPasswordRefCount)
	}
	if len(summary.RegistryServers) != 1 || summary.RegistryServers[0] != "contoso.azurecr.io" {
		t.Fatalf("RegistryServers = %#v, want safe registry server", summary.RegistryServers)
	}
	if len(summary.WorkloadIdentityIDs) != 1 || summary.WorkloadIdentityIDs[0] != userAssignedID {
		t.Fatalf("WorkloadIdentityIDs = %#v, want attached user-assigned identity", summary.WorkloadIdentityIDs)
	}
}
