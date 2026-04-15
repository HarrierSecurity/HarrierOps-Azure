package providers

import (
	"context"

	"harrierops-azure/internal/models"
)

func (StaticProvider) AuthPolicies(_ context.Context, tenant string, subscription string) (AuthPoliciesFacts, error) {
	session := staticFixtureSession(tenant, subscription)

	return AuthPoliciesFacts{
		TenantID:       session.TenantID,
		SubscriptionID: session.Subscription.ID,
		AuthPolicies: []models.AuthPolicySummary{
			{
				Controls:   []string{},
				Name:       "Security Defaults",
				PolicyType: "security-defaults",
				RelatedIDs: []string{"00000000-0000-0000-0000-000000000005"},
				Scope:      models.StringPtr("tenant"),
				State:      "disabled",
				Summary:    "Security defaults are disabled for the tenant.",
			},
			{
				Controls: []string{
					"guest-invites:everyone",
					"sspr:enabled",
					"legacy-msol-powershell:blocked",
					"users-can-register-apps",
					"users-can-create-security-groups",
					"user-consent:self-service",
				},
				Name:       "Authorization Policy",
				PolicyType: "authorization-policy",
				RelatedIDs: []string{"authorizationPolicy"},
				Scope:      models.StringPtr("tenant"),
				State:      "configured",
				Summary:    "guest invites: everyone; users can register apps; self-service permission grant policies assigned; legacy MSOL PowerShell blocked",
			},
			{
				Controls:   []string{"block"},
				Name:       "CA002: Block legacy auth",
				PolicyType: "conditional-access",
				RelatedIDs: []string{"ca-2"},
				Scope:      models.StringPtr("users:all, apps:all"),
				State:      "disabled",
				Summary:    "state: disabled; grants: block; scope: users:all, apps:all",
			},
			{
				Controls:   []string{"mfa"},
				Name:       "CA001: Require multi-factor authentication for admins",
				PolicyType: "conditional-access",
				RelatedIDs: []string{"ca-1"},
				Scope:      models.StringPtr("roles:2, apps:all"),
				State:      "enabledForReportingButNotEnforced",
				Summary:    "state: enabledForReportingButNotEnforced; grants: mfa; scope: roles:2, apps:all",
			},
		},
		Issues: []models.Issue{},
	}, nil
}
