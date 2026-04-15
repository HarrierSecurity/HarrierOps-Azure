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
