package render

import (
	"testing"

	"harrierops-azure/internal/contracts"
)

func TestRendererRegistryCoversImplementedContracts(t *testing.T) {
	for _, name := range contracts.CommandNames() {
		contract, ok := contracts.Command(name)
		if !ok || contract.Status != contracts.StatusImplemented {
			continue
		}

		entry, err := renderRegistryEntry(name)
		if err != nil {
			t.Fatalf("expected render registry entry for %q: %v", name, err)
		}
		if entry.table == nil {
			t.Fatalf("expected table renderer for %q", name)
		}
		if entry.csv == nil {
			t.Fatalf("expected csv renderer for %q", name)
		}
	}
}

func TestChainsFamilyRenderersCoverImplementedFamilies(t *testing.T) {
	for _, familyName := range contracts.FamilyNames() {
		family, ok := contracts.Family(familyName)
		if !ok || family.Status != contracts.StatusImplemented {
			continue
		}
		if chainsFamilyTableRenderers[familyName] == nil {
			t.Fatalf("expected implemented family %q to have a table renderer", familyName)
		}
		if chainsFamilyCSVRenderers[familyName] == nil {
			t.Fatalf("expected implemented family %q to have a csv renderer", familyName)
		}
	}
}

func TestPersistenceCommandRendererExists(t *testing.T) {
	entry, err := renderRegistryEntry("persistence")
	if err != nil {
		t.Fatalf("expected persistence renderer entry: %v", err)
	}
	if entry.table == nil {
		t.Fatalf("expected table renderer for persistence")
	}
	if entry.csv == nil {
		t.Fatalf("expected csv renderer for persistence")
	}
}
