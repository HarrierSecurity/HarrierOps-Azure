package providers

import (
	"context"
	"fmt"

	"harrierops-azure/internal/models"
)

func (StaticProvider) AppCredentials(_ context.Context, tenant string, subscription string) (AppCredentialsFacts, error) {
	session := staticFixtureSession(tenant, subscription)
	buildSPRoleContext := "Backed identity already holds high-impact Azure roles (Owner) across 2 visible scopes."

	rows := []models.AppCredentialSummary{
		{
			RowClass:                    "directly_addable_federated_trust",
			TargetObjectType:            "Application",
			TargetObjectID:              "55555555-5555-5555-5555-555555555555",
			TargetObjectName:            "build-app",
			BackingServicePrincipalID:   models.StringPtr("66666666-6666-6666-6666-666666666666"),
			BackingServicePrincipalName: models.StringPtr("build-sp"),
			CredentialType:              models.StringPtr("federated"),
			ControlPath:                 "application-owner",
			RoleContext:                 buildSPRoleContext,
			TenantContext:               "current-tenant",
			CurrentEvidence:             "Current identity 'azurefox-lab-sp' visibly owns application 'build-app', and federated trust lives on that application object.",
			MissingProof:                "This row shows direct control of the federated-trust surface, not which external subject would be trusted after a change.",
			OperatorActionability:       "Treat this as a visible path to add, replace, or widen federated trust on the application.",
			RecommendedFixFocus:         "Remove the ownership path that lets the current identity control federated trust on this application.",
			Summary:                     "Current identity 'azurefox-lab-sp' visibly owns application 'build-app', and federated trust lives on that application object. Backed identity already holds high-impact Azure roles (Owner) across 2 visible scopes. This row shows direct control of the federated-trust surface, not which external subject would be trusted after a change.",
			RelatedIDs:                  []string{"33333333-3333-3333-3333-333333333333", "55555555-5555-5555-5555-555555555555", "66666666-6666-6666-6666-666666666666"},
		},
		{
			RowClass:                    "directly_addable",
			TargetObjectType:            "Application",
			TargetObjectID:              "55555555-5555-5555-5555-555555555555",
			TargetObjectName:            "build-app",
			BackingServicePrincipalID:   models.StringPtr("66666666-6666-6666-6666-666666666666"),
			BackingServicePrincipalName: models.StringPtr("build-sp"),
			CredentialType:              models.StringPtr("password-or-key"),
			ControlPath:                 "application-owner",
			RoleContext:                 buildSPRoleContext,
			TenantContext:               "current-tenant",
			CurrentEvidence:             "Current identity 'azurefox-lab-sp' visibly owns application 'build-app', and application ownership can change authentication material accepted here.",
			MissingProof:                "",
			OperatorActionability:       "Treat this as a visible path to add or replace authentication material on the application object.",
			RecommendedFixFocus:         "Remove the ownership path that lets the current identity control this application.",
			Summary:                     "Current identity 'azurefox-lab-sp' visibly owns application 'build-app', and application ownership can change authentication material accepted here. Backed identity already holds high-impact Azure roles (Owner) across 2 visible scopes.",
			RelatedIDs:                  []string{"33333333-3333-3333-3333-333333333333", "55555555-5555-5555-5555-555555555555", "66666666-6666-6666-6666-666666666666"},
		},
		{
			RowClass:                    "directly_addable",
			TargetObjectType:            "ServicePrincipal",
			TargetObjectID:              "66666666-6666-6666-6666-666666666666",
			TargetObjectName:            "build-sp",
			BackingServicePrincipalID:   models.StringPtr("66666666-6666-6666-6666-666666666666"),
			BackingServicePrincipalName: models.StringPtr("build-sp"),
			CredentialType:              models.StringPtr("password-or-key"),
			ControlPath:                 "service-principal-owner",
			RoleContext:                 buildSPRoleContext,
			TenantContext:               "current-tenant",
			CurrentEvidence:             "Current identity 'azurefox-lab-sp' visibly owns service principal 'build-sp', and service-principal ownership can change authentication material Azure accepts here.",
			MissingProof:                "",
			OperatorActionability:       "Treat this as a visible path to add or replace authentication material on the service principal.",
			RecommendedFixFocus:         "Remove the owner-level control path that lets the current identity control this service principal.",
			Summary:                     "Current identity 'azurefox-lab-sp' visibly owns service principal 'build-sp', and service-principal ownership can change authentication material Azure accepts here. Backed identity already holds high-impact Azure roles (Owner) across 2 visible scopes.",
			RelatedIDs:                  []string{"33333333-3333-3333-3333-333333333333", "66666666-6666-6666-6666-666666666666"},
		},
		{
			RowClass:                    "federated_trust_present",
			TargetObjectType:            "Application",
			TargetObjectID:              "55555555-5555-5555-5555-555555555555",
			TargetObjectName:            "build-app",
			BackingServicePrincipalID:   models.StringPtr("66666666-6666-6666-6666-666666666666"),
			BackingServicePrincipalName: models.StringPtr("build-sp"),
			CredentialType:              models.StringPtr("federated"),
			ControlPath:                 "existing-federated-trust",
			RoleContext:                 buildSPRoleContext,
			TenantContext:               "current-tenant",
			CurrentEvidence:             "Application 'build-app' already trusts federated subject 'repo:TacoRocket/AzureFox:ref:refs/heads/main' from issuer 'https://token.actions.githubusercontent.com'.",
			MissingProof:                "This row shows existing federated trust, not that the current identity can change it.",
			OperatorActionability:       "Review whether this external trust path is still required and who can modify the application that carries it.",
			RecommendedFixFocus:         "Remove or tighten federated trust that no longer needs to yield Azure-facing access.",
			Summary:                     "Application 'build-app' already trusts federated subject 'repo:TacoRocket/AzureFox:ref:refs/heads/main' from issuer 'https://token.actions.githubusercontent.com'. Backed identity already holds high-impact Azure roles (Owner) across 2 visible scopes. This row shows existing federated trust, not that the current identity can change it.",
			RelatedIDs:                  []string{"55555555-5555-5555-5555-555555555555", "fic-build-main", "66666666-6666-6666-6666-666666666666"},
		},
		{
			RowClass:                    "existing_credential",
			TargetObjectType:            "ServicePrincipal",
			TargetObjectID:              "66666666-6666-6666-6666-666666666666",
			TargetObjectName:            "build-sp",
			BackingServicePrincipalID:   models.StringPtr("66666666-6666-6666-6666-666666666666"),
			BackingServicePrincipalName: models.StringPtr("build-sp"),
			CredentialType:              models.StringPtr("password"),
			ControlPath:                 "existing-auth-material",
			RoleContext:                 buildSPRoleContext,
			TenantContext:               "current-tenant",
			CurrentEvidence:             "ServicePrincipal 'build-sp' already has 1 visible password credential metadata entry.",
			MissingProof:                "This row shows existing authentication material, not a current-identity path to change it.",
			OperatorActionability:       "Review who can modify this object and whether the existing credential material is still needed.",
			RecommendedFixFocus:         "Remove stale authentication material and tighten ownership over the identity that accepts it.",
			Summary:                     "ServicePrincipal 'build-sp' already has 1 visible password credential metadata entry. Backed identity already holds high-impact Azure roles (Owner) across 2 visible scopes. This row shows existing authentication material, not a current-identity path to change it.",
			RelatedIDs:                  []string{"66666666-6666-6666-6666-666666666666", "pwd-build-sp-main"},
		},
	}

	models.SortAppCredentialRows(rows)

	return AppCredentialsFacts{
		TenantID:       session.TenantID,
		SubscriptionID: session.Subscription.ID,
		AppCredentials: rows,
		Issues: []models.Issue{
			{
				Kind:    "note",
				Message: fmt.Sprintf("Static fixture data includes %d representative app-credential rows for deterministic testing.", len(rows)),
				Scope:   "app_credentials.fixture",
			},
		},
	}, nil
}
