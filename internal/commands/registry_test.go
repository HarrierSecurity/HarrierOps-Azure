package commands

import (
	"testing"
	"time"

	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/providers"
)

func TestRegistryCoversImplementedContracts(t *testing.T) {
	registry := NewRegistry(providers.NewStaticProvider(), func() time.Time { return time.Unix(0, 0) })

	for _, name := range registry.CommandNames() {
		contract, ok := contracts.Command(name)
		if !ok || contract.Status != contracts.StatusImplemented {
			continue
		}

		definition := registry.definitions[name]
		if definition.Handler == nil {
			t.Fatalf("expected implemented command %q to have a handler", name)
		}
	}
}
