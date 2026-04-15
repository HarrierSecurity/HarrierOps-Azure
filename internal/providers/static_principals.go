package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (provider StaticProvider) Principals(ctx context.Context, tenant string, subscription string) (PrincipalsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	subscriptionID := session.Subscription.ID

	return PrincipalsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: subscriptionID,
		Principals: []models.PrincipalSummary{
			{
				AttachedTo: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachines/vm-web-01",
				},
				DisplayName:         models.StringPtr("azurefox-lab-sp"),
				ID:                  "33333333-3333-3333-3333-333333333333",
				IdentityNames:       []string{"ua-app"},
				IdentityTypes:       []string{"userAssigned"},
				IsCurrentIdentity:   true,
				PrincipalType:       "ServicePrincipal",
				RoleAssignmentCount: 1,
				RoleNames:           []string{"Owner"},
				ScopeIDs:            []string{"/subscriptions/" + subscriptionID},
				Sources:             []string{"rbac", "whoami", "managed-identities"},
				TenantID:            models.StringPtr(session.TenantID),
			},
			{
				AttachedTo:          []string{},
				DisplayName:         models.StringPtr("operator@lab.local"),
				ID:                  "44444444-4444-4444-4444-444444444444",
				IdentityNames:       []string{},
				IdentityTypes:       []string{},
				IsCurrentIdentity:   false,
				PrincipalType:       "User",
				RoleAssignmentCount: 1,
				RoleNames:           []string{"Reader"},
				ScopeIDs:            []string{"/subscriptions/" + subscriptionID},
				Sources:             []string{"rbac"},
				TenantID:            models.StringPtr(session.TenantID),
			},
			{
				AttachedTo: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/func-orders",
				},
				DisplayName:         models.StringPtr("func-orders-system"),
				ID:                  "cccc2222-2222-2222-2222-222222222222",
				IdentityNames:       []string{"func-orders-system"},
				IdentityTypes:       []string{"systemAssigned"},
				IsCurrentIdentity:   false,
				PrincipalType:       "ServicePrincipal",
				RoleAssignmentCount: 1,
				RoleNames:           []string{"Contributor"},
				ScopeIDs:            []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps"},
				Sources:             []string{"managed-identities"},
				TenantID:            models.StringPtr(session.TenantID),
			},
			{
				AttachedTo: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps/providers/Microsoft.Web/sites/app-empty-mi",
				},
				DisplayName:         models.StringPtr("app-empty-mi-system"),
				ID:                  "eeee3333-3333-3333-3333-333333333333",
				IdentityNames:       []string{"app-empty-mi-system"},
				IdentityTypes:       []string{"systemAssigned"},
				IsCurrentIdentity:   false,
				PrincipalType:       "ServicePrincipal",
				RoleAssignmentCount: 1,
				RoleNames:           []string{"Contributor"},
				ScopeIDs:            []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-apps"},
				Sources:             []string{"managed-identities"},
				TenantID:            models.StringPtr(session.TenantID),
			},
			{
				AttachedTo: []string{
					"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload/providers/Microsoft.Compute/virtualMachineScaleSets/vmss-edge-01",
				},
				DisplayName:         models.StringPtr("vmss-edge-01-system"),
				ID:                  "77770000-0000-0000-0000-000000000001",
				IdentityNames:       []string{"vmss-edge-01-system"},
				IdentityTypes:       []string{"systemAssigned"},
				IsCurrentIdentity:   false,
				PrincipalType:       "ServicePrincipal",
				RoleAssignmentCount: 1,
				RoleNames:           []string{"Contributor"},
				ScopeIDs:            []string{"/subscriptions/" + subscriptionID + "/resourceGroups/rg-workload"},
				Sources:             []string{"managed-identities"},
				TenantID:            models.StringPtr(session.TenantID),
			},
		},
		Issues: []models.Issue{},
	}, nil
}
