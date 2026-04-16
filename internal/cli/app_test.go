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
			lootGolden:  "whoami.loot.golden.json",
			tableGolden: "whoami.golden.table.txt",
		},
		"rbac": {
			tableGolden: "rbac.golden.table.txt",
		},
		"inventory": {
			tableGolden: "inventory.golden.table.txt",
		},
		"permissions": {
			tableGolden: "permissions.golden.table.txt",
		},
		"chains": {
			tableGolden: "chains.golden.table.txt",
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
	cases = append(cases, artifactCase{
		name:         "chains-credential-path",
		args:         []string{"chains", "credential-path", "--output", "json"},
		artifactBase: "chains",
		jsonGolden:   "chains-credential-path.golden.json",
		lootGolden:   "chains-credential-path.golden.json",
		csvGolden:    "chains-credential-path.golden.csv",
		tableGolden:  "chains-credential-path.golden.table.txt",
	})
	cases = append(cases, artifactCase{
		name:         "chains-deployment-path",
		args:         []string{"chains", "deployment-path", "--output", "json"},
		artifactBase: "chains",
		jsonGolden:   "chains-deployment-path.golden.json",
		lootGolden:   "chains-deployment-path.golden.json",
		csvGolden:    "chains-deployment-path.golden.csv",
		tableGolden:  "chains-deployment-path.golden.table.txt",
	})
	cases = append(cases, artifactCase{
		name:         "chains-escalation-path",
		args:         []string{"chains", "escalation-path", "--output", "json"},
		artifactBase: "chains",
		jsonGolden:   "chains-escalation-path.golden.json",
		lootGolden:   "chains-escalation-path.golden.json",
		csvGolden:    "chains-escalation-path.golden.csv",
		tableGolden:  "chains-escalation-path.golden.table.txt",
	})
	cases = append(cases, artifactCase{
		name:         "chains-compute-control",
		args:         []string{"chains", "compute-control", "--output", "json"},
		artifactBase: "chains",
		jsonGolden:   "chains-compute-control.golden.json",
		lootGolden:   "chains-compute-control.golden.json",
		csvGolden:    "chains-compute-control.golden.csv",
		tableGolden:  "chains-compute-control.golden.table.txt",
	})
	cases = append(cases, artifactCase{
		name:         "chains-persistence-path",
		args:         []string{"chains", "persistence-path", "--output", "json"},
		artifactBase: "chains",
		jsonGolden:   "chains-persistence-path.golden.json",
		lootGolden:   "chains-persistence-path.golden.json",
		csvGolden:    "chains-persistence-path.golden.csv",
		tableGolden:  "chains-persistence-path.golden.table.txt",
	})
	cases = append(cases, artifactCase{
		name:         "persistence-automation",
		args:         []string{"persistence", "automation", "--output", "json"},
		artifactBase: "persistence",
		jsonGolden:   "persistence-automation.golden.json",
		lootGolden:   "persistence-automation.golden.json",
		csvGolden:    "persistence-automation.golden.csv",
		tableGolden:  "persistence-automation.golden.table.txt",
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

func TestGlobalFlagsMustFollowCommand(t *testing.T) {
	app := newTestApp()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := app.Run([]string{"--output", "json", "inventory"}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d with stderr %q", exitCode, stderr.String())
	}
	if !strings.Contains(stderr.String(), "command must come first; use `ho-azure <command> [flags]`") {
		t.Fatalf("expected command-first guidance, got stderr %q", stderr.String())
	}
}

func TestGlobalFlagsWorkAfterCommand(t *testing.T) {
	stdout, _ := runSuccess(t, "inventory", "--output", "json")
	assertSchemaVersion(t, stdout)
}

func TestGlobalDebugFlagIsAccepted(t *testing.T) {
	stdout, _ := runSuccess(t, "inventory", "--debug", "--output", "json")
	assertSchemaVersion(t, stdout)
}

func TestRootHelpFlagNormalizesToHelp(t *testing.T) {
	app := newTestApp()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := app.Run([]string{"--help"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "HO-Azure Help") {
		t.Fatalf("expected root help output, got %q", stdout.String())
	}
}

func TestCommandHelpFlagNormalizesToCustomHelp(t *testing.T) {
	app := newTestApp()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := app.Run([]string{"whoami", "--help"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "HO-Azure Help :: whoami") {
		t.Fatalf("expected command help output, got %q", stdout.String())
	}
}

func TestSectionHelpTopicIsAvailable(t *testing.T) {
	app := newTestApp()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := app.Run([]string{"help", "identity"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "HO-Azure Help :: identity") {
		t.Fatalf("expected section help output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "whoami:") {
		t.Fatalf("expected identity commands in section help, got %q", stdout.String())
	}
}

func TestWorkflowSectionHelpTopicIsAvailable(t *testing.T) {
	app := newTestApp()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := app.Run([]string{"help", "workflow"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "HO-Azure Help :: workflow") {
		t.Fatalf("expected workflow section help output, got %q", stdout.String())
	}
	for _, commandName := range []string{"logic-apps:", "event-grid:", "azure-ml:"} {
		if !strings.Contains(stdout.String(), commandName) {
			t.Fatalf("expected workflow section help to include %q, got %q", commandName, stdout.String())
		}
	}
}

func TestRootHelpListsWorkflowSection(t *testing.T) {
	app := newTestApp()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := app.Run([]string{"help"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "workflow: Review workflow and event-driven surfaces") {
		t.Fatalf("expected root help to list workflow section, got %q", stdout.String())
	}
}

func TestRootHelpListsOnlyImplementedCommands(t *testing.T) {
	app := newTestApp()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := app.Run([]string{"help"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}
	if strings.Contains(stdout.String(), "Placeholder contract carried forward from AzureFox for faithful migration.") {
		t.Fatalf("expected root help to omit placeholder commands, got %q", stdout.String())
	}
}

func TestCommandHelpShowsTruthfulStatus(t *testing.T) {
	app := newTestApp()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := app.Run([]string{"help", "inventory"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Status: implemented command.") {
		t.Fatalf("expected truthful implemented status in help, got %q", stdout.String())
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

func TestChainsHelpMatchesOverviewJSON(t *testing.T) {
	overview, _ := runSuccess(t, "chains", "--output", "json")
	helpView, _ := runSuccess(t, "chains", "help", "--output", "json")
	assertMatchesGolden(t, overview, "chains.golden.json")
	if overview != helpView {
		t.Fatalf("expected chains help JSON to match overview\noverview:\n%s\nhelp:\n%s", overview, helpView)
	}
}

func TestPersistenceHelpMatchesOverviewJSON(t *testing.T) {
	overview, _ := runSuccess(t, "persistence", "--output", "json")
	helpView, _ := runSuccess(t, "persistence", "help", "--output", "json")
	assertMatchesGolden(t, overview, "persistence.golden.json")
	if overview != helpView {
		t.Fatalf("expected persistence help JSON to match overview\noverview:\n%s\nhelp:\n%s", overview, helpView)
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
