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

func commandArtifactCase(contract contracts.CommandContract) artifactCase {
	return artifactCase{
		name:         contract.Name,
		args:         []string{contract.Name, "--output", "json"},
		artifactBase: contract.Name,
		jsonGolden:   contract.Name + ".golden.json",
		lootGolden:   contract.Name + ".golden.json",
		csvGolden:    contract.Name + ".golden.csv",
		tableGolden:  contract.Name + ".golden.table.txt",
	}
}

func explicitArtifactCase(name string, args []string, artifactBase string) artifactCase {
	return artifactCase{
		name:         name,
		args:         append([]string{}, args...),
		artifactBase: artifactBase,
		jsonGolden:   name + ".golden.json",
		lootGolden:   name + ".golden.json",
		csvGolden:    name + ".golden.csv",
		tableGolden:  name + ".golden.table.txt",
	}
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
		artifact := commandArtifactCase(contract)
		if override, ok := overrides[contract.Name]; ok {
			if override.lootGolden != "" {
				artifact.lootGolden = override.lootGolden
			}
			artifact.tableGolden = override.tableGolden
			artifact.tableContains = append([]string{}, override.tableContains...)
		}
		cases = append(cases, artifact)
	}

	for _, extra := range []artifactCase{
		explicitArtifactCase("role-trusts-full", []string{"role-trusts", "--mode", "full", "--output", "json"}, "role-trusts"),
		explicitArtifactCase("chains-credential-path", []string{"chains", "credential-path", "--output", "json"}, "chains"),
		explicitArtifactCase("chains-deployment-path", []string{"chains", "deployment-path", "--output", "json"}, "chains"),
		explicitArtifactCase("chains-escalation-path", []string{"chains", "escalation-path", "--output", "json"}, "chains"),
		explicitArtifactCase("chains-compute-control", []string{"chains", "compute-control", "--output", "json"}, "chains"),
		explicitArtifactCase("persistence-automation", []string{"persistence", "automation", "--output", "json"}, "persistence"),
		explicitArtifactCase("persistence-app-service", []string{"persistence", "app-service", "--output", "json"}, "persistence"),
		explicitArtifactCase("persistence-azure-ml", []string{"persistence", "azure-ml", "--output", "json"}, "persistence"),
		explicitArtifactCase("persistence-container-apps-jobs", []string{"persistence", "container-apps-jobs", "--output", "json"}, "persistence"),
		explicitArtifactCase("persistence-vm-extensions", []string{"persistence", "vm-extensions", "--output", "json"}, "persistence"),
		explicitArtifactCase("persistence-logic-apps", []string{"persistence", "logic-apps", "--output", "json"}, "persistence"),
		explicitArtifactCase("persistence-functions", []string{"persistence", "functions", "--output", "json"}, "persistence"),
		explicitArtifactCase("persistence-webjobs", []string{"persistence", "webjobs", "--output", "json"}, "persistence"),
		explicitArtifactCase("evasion-appinsights", []string{"evasion", "appinsights", "--output", "json"}, "evasion"),
		explicitArtifactCase("evasion-dcr", []string{"evasion", "dcr", "--output", "json"}, "evasion"),
		explicitArtifactCase("evasion-diagnostic-settings", []string{"evasion", "diagnostic-settings", "--output", "json"}, "evasion"),
		explicitArtifactCase("resourcehijacking-api-mgmt", []string{"resourcehijacking", "api-mgmt", "--output", "json"}, "resourcehijacking"),
		explicitArtifactCase("resourcehijacking-automation", []string{"resourcehijacking", "automation", "--output", "json"}, "resourcehijacking"),
		explicitArtifactCase("resourcehijacking-logic-apps", []string{"resourcehijacking", "logic-apps", "--output", "json"}, "resourcehijacking"),
		explicitArtifactCase("pathmasking-api-mgmt", []string{"pathmasking", "api-mgmt", "--output", "json"}, "pathmasking"),
		explicitArtifactCase("pathmasking-logic-apps", []string{"pathmasking", "logic-apps", "--output", "json"}, "pathmasking"),
		explicitArtifactCase("pathmasking-relay", []string{"pathmasking", "relay", "--output", "json"}, "pathmasking"),
	} {
		cases = append(cases, extra)
	}

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

func TestEvasionHelpMatchesOverviewJSON(t *testing.T) {
	overview, _ := runSuccess(t, "evasion", "--output", "json")
	helpView, _ := runSuccess(t, "evasion", "help", "--output", "json")
	assertMatchesGolden(t, overview, "evasion.golden.json")
	if overview != helpView {
		t.Fatalf("expected evasion help JSON to match overview\noverview:\n%s\nhelp:\n%s", overview, helpView)
	}
}

func TestResourceHijackingHelpMatchesOverviewJSON(t *testing.T) {
	overview, _ := runSuccess(t, "resourcehijacking", "--output", "json")
	helpView, _ := runSuccess(t, "resourcehijacking", "help", "--output", "json")
	assertMatchesGolden(t, overview, "resourcehijacking.golden.json")
	if overview != helpView {
		t.Fatalf("expected resourcehijacking help JSON to match overview\noverview:\n%s\nhelp:\n%s", overview, helpView)
	}
}

func TestPathMaskingHelpMatchesOverviewJSON(t *testing.T) {
	overview, _ := runSuccess(t, "pathmasking", "--output", "json")
	helpView, _ := runSuccess(t, "pathmasking", "help", "--output", "json")
	assertMatchesGolden(t, overview, "pathmasking.golden.json")
	if overview != helpView {
		t.Fatalf("expected pathmasking help JSON to match overview\noverview:\n%s\nhelp:\n%s", overview, helpView)
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

func TestDefaultArtifactWorkspaceIsCurrentDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	app := newTestApp()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := app.Run([]string{"rbac", "--output", "json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}

	if !strings.Contains(stdout.String(), `"artifact_context"`) {
		t.Fatalf("expected stdout JSON to carry artifact context")
	}
	for _, path := range []string{
		filepath.Join("json", "rbac.json"),
		filepath.Join("loot", "rbac.json"),
		filepath.Join("csv", "rbac.csv"),
		filepath.Join("table", "rbac.txt"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected default artifact %s: %v", path, err)
		}
	}
}

func runSuccess(t *testing.T, args ...string) (string, string) {
	t.Helper()
	defer cleanupDefaultArtifacts(t)

	app := newTestApp()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := app.Run(args, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}

	return stdout.String(), stderr.String()
}

func cleanupDefaultArtifacts(t *testing.T) {
	t.Helper()
	for _, dir := range []string{"csv", "json", "loot", "table"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("cleanup generated artifact directory %s: %v", dir, err)
		}
	}
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
