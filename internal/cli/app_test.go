package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"harrierops-azure/internal/commands"
	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/providers"
)

func newTestApp() *App {
	fixedNow := func() time.Time {
		return time.Date(2026, time.April, 13, 12, 0, 0, 0, time.UTC)
	}
	return New(commands.NewRegistry(providers.NewStaticProvider(), fixedNow))
}

type outputCase struct {
	name   string
	args   []string
	golden string
}

type artifactCase struct {
	name          string
	args          []string
	artifactBase  string
	jsonGolden    string
	lootGolden    string
	csvGolden     string
	tableGolden   string
	tableContains []string
}

func implementedArtifactCases() []artifactCase {
	overrides := map[string]artifactCase{
		"whoami": {
			tableContains: []string{"azurefox whoami", "principal_type"},
			lootGolden:    "whoami.loot.golden.json",
			tableGolden:   "",
		},
		"rbac": {
			tableContains: []string{"azurefox rbac", "role_definition_id"},
			tableGolden:   "",
		},
		"inventory": {
			tableContains: []string{"azurefox inventory", "top_type"},
			tableGolden:   "",
		},
		"permissions": {
			tableContains: []string{"azurefox permissions", "operator signal"},
			tableGolden:   "",
		},
	}

	cases := []artifactCase{}
	for _, contract := range contracts.ImplementedCommands() {
		artifact := artifactCase{
			name:         contract.Name,
			args:         []string{contract.Name, "--output", "json"},
			artifactBase: contract.Name,
			jsonGolden:   contract.Name + ".golden.json",
			lootGolden:   contract.Name + ".golden.json",
			csvGolden:    contract.Name + ".golden.csv",
			tableGolden:  contract.Name + ".golden.table.txt",
		}
		if override, ok := overrides[contract.Name]; ok {
			if override.lootGolden != "" {
				artifact.lootGolden = override.lootGolden
			}
			artifact.tableGolden = override.tableGolden
			artifact.tableContains = append([]string{}, override.tableContains...)
		}
		cases = append(cases, artifact)
	}

	cases = append(cases, artifactCase{
		name:         "role-trusts-full",
		args:         []string{"role-trusts", "--mode", "full", "--output", "json"},
		artifactBase: "role-trusts",
		jsonGolden:   "role-trusts-full.golden.json",
		lootGolden:   "role-trusts-full.golden.json",
		csvGolden:    "role-trusts-full.golden.csv",
		tableGolden:  "role-trusts-full.golden.table.txt",
	})

	return cases
}

func TestJSONOutputIsDeterministic(t *testing.T) {
	for _, artifact := range implementedArtifactCases() {
		tc := outputCase{name: artifact.name, args: append([]string{}, artifact.args...), golden: artifact.jsonGolden}
		t.Run(tc.name, func(t *testing.T) {
			stdout, _ := runSuccess(t, tc.args...)
			assertMatchesGolden(t, stdout, tc.golden)
		})
	}
}

func TestCSVColumnsStayStable(t *testing.T) {
	for _, artifact := range implementedArtifactCases() {
		args := append([]string{}, artifact.args...)
		args[len(args)-1] = "csv"
		tc := outputCase{name: artifact.name, args: args, golden: artifact.csvGolden}
		t.Run(tc.name, func(t *testing.T) {
			stdout, _ := runSuccess(t, tc.args...)
			assertMatchesGolden(t, stdout, tc.golden)
		})
	}
}

func TestTableOutputStaysStable(t *testing.T) {
	for _, artifact := range implementedArtifactCases() {
		if artifact.tableGolden == "" {
			continue
		}
		args := append([]string{}, artifact.args...)
		args[len(args)-1] = "table"
		tc := outputCase{name: artifact.name, args: args, golden: artifact.tableGolden}
		t.Run(tc.name, func(t *testing.T) {
			stdout, _ := runSuccess(t, tc.args...)
			assertMatchesGolden(t, stdout, tc.golden)
		})
	}
}

func TestRoleTrustsRejectsInvalidMode(t *testing.T) {
	app := newTestApp()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := app.Run([]string{"role-trusts", "--mode", "slow", "--output", "json"}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d with stderr %q", exitCode, stderr.String())
	}
	if !strings.Contains(stderr.String(), `invalid mode "slow"; valid values: fast, full, fast-old, full-old`) {
		t.Fatalf("expected invalid mode guidance, got stderr %q", stderr.String())
	}
}

func TestRoleTrustsLegacyModesPreserveSemanticOutput(t *testing.T) {
	for _, tc := range []struct {
		name   string
		args   []string
		golden string
	}{
		{name: "fast-old", args: []string{"role-trusts", "--mode", "fast-old", "--output", "json"}, golden: "role-trusts.golden.json"},
		{name: "full-old", args: []string{"role-trusts", "--mode", "full-old", "--output", "json"}, golden: "role-trusts-full.golden.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout, _ := runSuccess(t, tc.args...)
			assertMatchesGolden(t, stdout, tc.golden)
		})
	}
}

func TestJSONCarriesSchemaVersion(t *testing.T) {
	for _, artifact := range implementedArtifactCases() {
		tc := struct {
			name string
			args []string
		}{name: artifact.name, args: append([]string{}, artifact.args...)}
		t.Run(tc.name, func(t *testing.T) {
			stdout, _ := runSuccess(t, tc.args...)
			assertSchemaVersion(t, stdout)
		})
	}
}

func TestArtifactGenerationWritesAllFormats(t *testing.T) {
	for _, tc := range implementedArtifactCases() {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			args := append(append([]string{}, tc.args...), "--outdir", tempDir)
			runSuccess(t, args...)

			assertMatchesGoldenFile(t, filepath.Join(tempDir, "json", tc.artifactBase+".json"), tc.jsonGolden)
			assertMatchesGoldenFile(t, filepath.Join(tempDir, "loot", tc.artifactBase+".json"), tc.lootGolden)
			assertMatchesGoldenFile(t, filepath.Join(tempDir, "csv", tc.artifactBase+".csv"), tc.csvGolden)

			tableArtifact := readFile(t, filepath.Join(tempDir, "table", tc.artifactBase+".txt"))
			if tc.tableGolden != "" {
				if tableArtifact != readTestdata(t, tc.tableGolden) {
					t.Fatalf("table artifact mismatch\nwant:\n%s\ngot:\n%s", readTestdata(t, tc.tableGolden), tableArtifact)
				}
			}
			for _, snippet := range tc.tableContains {
				if !strings.Contains(tableArtifact, snippet) {
					t.Fatalf("table artifact missing %q: %q", snippet, tableArtifact)
				}
			}
		})
	}
}

func runSuccess(t *testing.T, args ...string) (string, string) {
	t.Helper()

	app := newTestApp()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := app.Run(args, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}

	return stdout.String(), stderr.String()
}

func assertSchemaVersion(t *testing.T, content string) {
	t.Helper()

	var decoded struct {
		Metadata struct {
			SchemaVersion string `json:"schema_version"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		t.Fatalf("unmarshal json output: %v", err)
	}
	if decoded.Metadata.SchemaVersion != contracts.AzureFoxSchemaVersion {
		t.Fatalf("expected schema version %q, got %q", contracts.AzureFoxSchemaVersion, decoded.Metadata.SchemaVersion)
	}
}

func assertMatchesGolden(t *testing.T, content string, golden string) {
	t.Helper()
	expected := readTestdata(t, golden)
	if content != expected {
		t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", expected, content)
	}
}

func assertMatchesGoldenFile(t *testing.T, path string, golden string) {
	t.Helper()
	assertMatchesGolden(t, readFile(t, path), golden)
}

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	return readFile(t, filepath.Join("..", "..", "testdata", name))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
