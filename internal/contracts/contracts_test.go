package contracts

import (
	"slices"
	"strings"
	"testing"
)

func TestImplementedContractsAreSelfConsistent(t *testing.T) {
	for _, contract := range ImplementedCommands() {
		t.Run(contract.Name, func(t *testing.T) {
			if contract.Name == "" {
				t.Fatalf("implemented contract must have a name")
			}
			if contract.Status != StatusImplemented {
				t.Fatalf("expected %s to be implemented, got %q", contract.Name, contract.Status)
			}
			if !strings.HasSuffix(contract.Model, "Output") {
				t.Fatalf("expected %s model to end with Output, got %q", contract.Name, contract.Model)
			}
			if strings.TrimSpace(contract.OperatorQuestion) == "" {
				t.Fatalf("expected %s operator question to be non-empty", contract.Name)
			}
			if len(contract.TopLevelFields) == 0 {
				t.Fatalf("expected %s top-level fields to be populated", contract.Name)
			}
			if !slices.Contains(contract.TopLevelFields, "metadata") {
				t.Fatalf("expected %s top-level fields to include metadata", contract.Name)
			}
			if !slices.Contains(contract.TopLevelFields, "issues") {
				t.Fatalf("expected %s top-level fields to include issues", contract.Name)
			}
		})
	}
}

func TestCommandSpecificFlagsStayAttachedToTheirContracts(t *testing.T) {
	roleTrusts, ok := Command("role-trusts")
	if !ok {
		t.Fatalf("expected role-trusts contract to exist")
	}
	if len(roleTrusts.Flags) != 1 || roleTrusts.Flags[0].Name != "mode" {
		t.Fatalf("expected role-trusts to declare its mode flag, got %#v", roleTrusts.Flags)
	}

	devops, ok := Command("devops")
	if !ok {
		t.Fatalf("expected devops contract to exist")
	}
	if len(devops.Flags) != 1 || devops.Flags[0].Name != "devops-organization" {
		t.Fatalf("expected devops to declare its organization flag, got %#v", devops.Flags)
	}
}

func TestChainFamilyContractsHaveMigrationHome(t *testing.T) {
	for _, familyName := range []string{"credential-path", "deployment-path", "escalation-path", "compute-control"} {
		family, ok := Family(familyName)
		if !ok {
			t.Fatalf("expected family contract %q to exist", familyName)
		}
		if family.GroupedCommand != "chains" {
			t.Fatalf("expected family %q to live under chains, got %q", familyName, family.GroupedCommand)
		}
		if len(family.BackingCommands) == 0 {
			t.Fatalf("expected family %q to name backing commands", familyName)
		}
	}
}
