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

func TestChainsFamilyBuildersCoverImplementedFamilies(t *testing.T) {
	for _, familyName := range contracts.FamilyNames() {
		family, ok := contracts.Family(familyName)
		if !ok || family.Status != contracts.StatusImplemented {
			continue
		}
		if chainsFamilyBuilders[familyName] == nil {
			t.Fatalf("expected implemented family %q to have a chains builder", familyName)
		}
	}
}

func TestPersistenceSurfaceBuildersCoverImplementedSurfaces(t *testing.T) {
	for _, surfaceName := range contracts.PersistenceSurfaceNames() {
		surface, ok := contracts.PersistenceSurface(surfaceName)
		if !ok || surface.Status != contracts.StatusImplemented {
			continue
		}
		if persistenceSurfaceBuilders[surfaceName] == nil {
			t.Fatalf("expected implemented persistence surface %q to have a builder", surfaceName)
		}
	}
}
