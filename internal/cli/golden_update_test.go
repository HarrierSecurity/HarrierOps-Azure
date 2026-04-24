package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateGoldens(t *testing.T) {
	if os.Getenv("HO_AZURE_UPDATE_GOLDENS") != "1" {
		t.Skip("set HO_AZURE_UPDATE_GOLDENS=1 to refresh deterministic CLI goldens")
	}

	filter := strings.TrimSpace(os.Getenv("HO_AZURE_GOLDEN_FILTER"))
	for _, artifact := range implementedArtifactCases() {
		if filter != "" && !strings.Contains(artifact.name, filter) {
			continue
		}

		tempDir := t.TempDir()
		args := append(append([]string{}, artifact.args...), "--outdir", tempDir)
		jsonOut, _ := runSuccess(t, args...)
		writeGolden(t, artifact.jsonGolden, jsonOut)
		writeGolden(t, artifact.lootGolden, readFile(t, filepath.Join(tempDir, "loot", artifact.artifactBase+".json")))
		writeGolden(t, artifact.csvGolden, readFile(t, filepath.Join(tempDir, "csv", artifact.artifactBase+".csv")))

		if artifact.tableGolden == "" {
			continue
		}
		writeGolden(t, artifact.tableGolden, readFile(t, filepath.Join(tempDir, "table", artifact.artifactBase+".txt")))
	}
}

func writeGolden(t *testing.T, name string, content string) {
	t.Helper()
	if name == "" {
		return
	}
	path := filepath.Join("..", "..", "testdata", name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
