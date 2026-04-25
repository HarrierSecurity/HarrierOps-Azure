package providers

import (
	"testing"

	"harrierops-azure/internal/models"
)

func TestPrincipalsFactsPromotesManagedIdentityOverServicePrincipal(t *testing.T) {
	principalID := "mi-principal"

	facts := PrincipalsFactsFromSources(
		"tenant-1",
		"sub-1",
		RBACFacts{
			Principals: []models.Principal{
				{
					ID:            principalID,
					DisplayName:   "build-runner",
					PrincipalType: "ServicePrincipal",
				},
			},
		},
		WhoAmIFacts{},
		ManagedIdentitiesFacts{
			Identities: []models.ManagedIdentity{
				{
					ID:          "mi-1",
					Name:        "build-runner-mi",
					PrincipalID: models.StringPtr(principalID),
				},
			},
		},
	)

	if len(facts.Principals) != 1 {
		t.Fatalf("expected 1 principal, got %d", len(facts.Principals))
	}
	if got := facts.Principals[0].PrincipalType; got != "ManagedIdentity" {
		t.Fatalf("expected managed identity evidence to promote principal type, got %q", got)
	}
}
