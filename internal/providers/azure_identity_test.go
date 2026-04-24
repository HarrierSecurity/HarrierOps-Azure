package providers

import (
	"context"
	"strings"
	"testing"

	"harrierops-azure/internal/models"
)

func TestManagedIdentityNarrativeUsesLogicAppSpecificFollowup(t *testing.T) {
	operatorSignal, nextReview, summary := managedIdentityNarrative(
		"LogicApp",
		"la-inbound-redeploy",
		"la-inbound-redeploy-identity",
		models.WorkloadExposureNone,
		true,
		false,
		[]string{"Contributor"},
	)

	if !strings.Contains(operatorSignal, "Logic App") {
		t.Fatalf("expected Logic App operator signal, got %q", operatorSignal)
	}
	if strings.Contains(nextReview, "env-vars") {
		t.Fatalf("expected Logic App-specific next review, got %q", nextReview)
	}
	if !strings.Contains(nextReview, "logic-apps") {
		t.Fatalf("expected logic-apps next review guidance, got %q", nextReview)
	}
	if !strings.Contains(summary, "Logic App 'la-inbound-redeploy'") {
		t.Fatalf("expected Logic App summary, got %q", summary)
	}
}

func TestRoleDefinitionDetailFromMapCollectsManagementActionLists(t *testing.T) {
	detail := roleDefinitionDetailFromMap(map[string]any{
		"name": "custom-role-id",
		"properties": map[string]any{
			"roleName": "Container Apps Job Operator",
			"permissions": []any{
				map[string]any{
					"actions":        []any{"Microsoft.App/jobs/write", "Microsoft.App/jobs/read", "Microsoft.App/jobs/write"},
					"notActions":     []any{"Microsoft.App/jobs/delete"},
					"dataActions":    []any{"Microsoft.App/jobs/executions/logs/read"},
					"notDataActions": []any{"Microsoft.App/jobs/secrets/read"},
				},
			},
		},
	})

	if detail.roleName != "Container Apps Job Operator" {
		t.Fatalf("expected role name to be collected, got %q", detail.roleName)
	}
	if got, want := strings.Join(detail.actions, ","), "Microsoft.App/jobs/read,Microsoft.App/jobs/write"; got != want {
		t.Fatalf("expected deduped sorted actions %q, got %q", want, got)
	}
	if got, want := strings.Join(detail.notActions, ","), "Microsoft.App/jobs/delete"; got != want {
		t.Fatalf("expected notActions %q, got %q", want, got)
	}
	if got, want := strings.Join(detail.dataActions, ","), "Microsoft.App/jobs/executions/logs/read"; got != want {
		t.Fatalf("expected dataActions %q, got %q", want, got)
	}
	if got, want := strings.Join(detail.notDataActions, ","), "Microsoft.App/jobs/secrets/read"; got != want {
		t.Fatalf("expected notDataActions %q, got %q", want, got)
	}
}

func TestResolveRoleDefinitionDetailKeepsBuiltInFastPathNameOnly(t *testing.T) {
	detail, err := resolveRoleDefinitionDetail(
		context.Background(),
		nil,
		"/subscriptions/sub/providers/Microsoft.Authorization/roleDefinitions/8e3af657-a8ff-443c-a75c-2fe8c4bcb635",
		map[string]roleDefinitionDetail{},
	)
	if err != nil {
		t.Fatalf("expected built-in fast path to avoid role-definition client, got %v", err)
	}
	if detail.roleName != "Owner" {
		t.Fatalf("expected Owner role name, got %q", detail.roleName)
	}
	if len(detail.actions) != 0 {
		t.Fatalf("expected built-in fast path to avoid synthetic action collection, got %v", detail.actions)
	}
}

func TestMergeManagedIdentityAttachmentNormalizesIDAndAccumulatesAttachments(t *testing.T) {
	identityMap := map[string]models.ManagedIdentity{}

	mergeManagedIdentityAttachment(identityMap, managedIdentityFromAttachment(
		"/subscriptions/sub-1/resourceGroups/RG-Apps/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
		"ua-app",
		"userAssigned",
		models.StringPtr("principal-1"),
		models.StringPtr("client-1"),
		"/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.Web/sites/app-uami-ctrl-a43cfa",
		"AppService",
		"app-uami-ctrl-a43cfa",
		models.WorkloadExposurePublic,
		"/subscriptions/sub-1",
		[]models.RoleAssignment{{RoleName: "Owner"}},
	))
	mergeManagedIdentityAttachment(identityMap, managedIdentityFromAttachment(
		"/subscriptions/sub-1/resourcegroups/rg-apps/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app",
		"ua-app",
		"userAssigned",
		models.StringPtr("principal-1"),
		models.StringPtr("client-1"),
		"/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.Web/sites/func-orders-a43cfa",
		"FunctionApp",
		"func-orders-a43cfa",
		models.WorkloadExposurePublic,
		"/subscriptions/sub-1",
		[]models.RoleAssignment{{RoleName: "Owner"}},
	))

	if len(identityMap) != 1 {
		t.Fatalf("expected one normalized identity row, got %d", len(identityMap))
	}

	identity := identityMap[armIDJoinKey("/subscriptions/sub-1/resourceGroups/rg-apps/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app")]
	if identity.ID != "/subscriptions/sub-1/resourceGroups/RG-Apps/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua-app" {
		t.Fatalf("expected first identity ID to be preserved, got %q", identity.ID)
	}
	if len(identity.AttachedTo) != 2 {
		t.Fatalf("expected two attachments, got %v", identity.AttachedTo)
	}
	if !containsString(identity.AttachedTo, "/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.Web/sites/app-uami-ctrl-a43cfa") {
		t.Fatalf("expected app service attachment to be preserved, got %v", identity.AttachedTo)
	}
	if !containsString(identity.AttachedTo, "/subscriptions/sub-1/resourceGroups/rg-workload/providers/Microsoft.Web/sites/func-orders-a43cfa") {
		t.Fatalf("expected function app attachment to be preserved, got %v", identity.AttachedTo)
	}
	if identity.OperatorSignal == nil || !strings.Contains(*identity.OperatorSignal, "Multiple workload attachments; representative signal:") {
		t.Fatalf("expected representative operator signal for multi-attachment row, got %v", identity.OperatorSignal)
	}
	if identity.NextReview == nil || !strings.Contains(*identity.NextReview, "Review the attached_to list for the full workload set.") {
		t.Fatalf("expected representative next review for multi-attachment row, got %v", identity.NextReview)
	}
	if identity.Summary == nil || !strings.Contains(*identity.Summary, "Managed identity 'ua-app' is attached to multiple visible workloads (2 attachments).") {
		t.Fatalf("expected representative summary for multi-attachment row, got %v", identity.Summary)
	}
	if identity.Summary == nil || !strings.Contains(*identity.Summary, "Representative workload signal:") {
		t.Fatalf("expected representative summary marker for multi-attachment row, got %v", identity.Summary)
	}
}
