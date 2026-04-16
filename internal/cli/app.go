package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"harrierops-azure/internal/artifacts"
	"harrierops-azure/internal/commands"
	"harrierops-azure/internal/contracts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/output"
)

type App struct {
	registry *commands.Registry
}

type sectionHelpTopic struct {
	Summary      string
	OperatorGoal string
}

var (
	helpFlags = map[string]struct{}{
		"-h":     {},
		"--help": {},
	}
	sectionOrder      = []string{"identity", "core", "config", "storage", "secrets", "resource", "compute", "network", "ai"}
	sectionHelpTopics = map[string]sectionHelpTopic{
		"identity": {
			Summary:      "Review caller context, RBAC, principal trust, and tenant-wide identity control paths.",
			OperatorGoal: "Work out who can operate here now, who can influence them, and which identity pivots matter first.",
		},
		"core": {
			Summary:      "Build fast subscription scope and footprint context before deeper command work.",
			OperatorGoal: "Understand the visible Azure estate quickly enough to choose the next meaningful recon path.",
		},
		"config": {
			Summary:      "Review management-plane configuration history and exposed deployment context.",
			OperatorGoal: "Find config clues, output exposure, and environment detail that changes the next operator decision.",
		},
		"storage": {
			Summary:      "Review storage assets and adjacent posture that may expose accessible data paths.",
			OperatorGoal: "Spot the storage surfaces most worth follow-up for public reachability, trust, or data access clues.",
		},
		"secrets": {
			Summary:      "Review secret-bearing services and credential surfaces without inflating proof claims.",
			OperatorGoal: "Find the most useful token, credential, and secret-adjacent follow-ons from current scope.",
		},
		"resource": {
			Summary:      "Review Azure services that create control paths, trust boundaries, or privileged service context.",
			OperatorGoal: "Identify the services whose posture most changes what an operator can control next.",
		},
		"compute": {
			Summary:      "Review workload-bearing assets for identity, ingress, and runtime follow-up.",
			OperatorGoal: "Identify the workloads that most change reachable execution, identity pivot, or deployment paths.",
		},
		"network": {
			Summary:      "Review ingress, addressability, and network boundary context around visible assets.",
			OperatorGoal: "Understand what is reachable, what it belongs to, and which network clues deserve follow-up first.",
		},
		"ai": {
			Summary:      "Reserved for future coverage.",
			OperatorGoal: "Reserved for future coverage.",
		},
	}
)

func New(registry *commands.Registry) *App {
	return &App{registry: registry}
}

func (app *App) Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = io.WriteString(stdout, app.rootHelp())
		return 0
	}
	if isHelpFlag(args[0]) {
		_, _ = io.WriteString(stdout, app.rootHelp())
		return 0
	}

	if args[0] == "help" {
		if len(args) == 1 {
			_, _ = io.WriteString(stdout, app.rootHelp())
			return 0
		}
		help, ok := app.helpTopic(args[1])
		if !ok {
			fmt.Fprintf(stderr, "unknown help topic %q\n", args[1])
			return 2
		}
		_, _ = io.WriteString(stdout, help)
		return 0
	}

	if strings.HasPrefix(args[0], "-") {
		fmt.Fprintln(stderr, "command must come first; use `ho-azure <command> [flags]`")
		return 2
	}

	commandName := args[0]
	if len(args) >= 2 && isHelpFlag(args[1]) && app.isHelpTopic(commandName) {
		help, ok := app.helpTopic(commandName)
		if !ok {
			fmt.Fprintf(stderr, "unknown help topic %q\n", commandName)
			return 2
		}
		_, _ = io.WriteString(stdout, help)
		return 0
	}
	contract, ok := contracts.Command(commandName)
	if !ok {
		fmt.Fprintf(stderr, "unknown command %q\n", commandName)
		return 2
	}

	options, err := parseOptions(commandName, args[1:], stderr)
	if err != nil {
		fmt.Fprintf(stderr, "error: %s\n", err)
		return 2
	}

	response, err := app.registry.Run(context.Background(), commandName, commands.Request{
		Tenant:             options.Tenant,
		Subscription:       options.Subscription,
		DevOpsOrganization: options.DevOpsOrganization,
		ChainFamily:        options.ChainFamily,
		Output:             options.Output,
		RoleTrustsMode:     options.RoleTrustsMode,
		OutDir:             options.OutDir,
	})
	if err != nil {
		if contract.Status != contracts.StatusImplemented {
			fmt.Fprintf(stderr, "%s\n", err)
			return 2
		}
		fmt.Fprintf(stderr, "error: %s\n", err)
		return 1
	}

	if options.OutDir != "" {
		if _, err := artifacts.Write(commandName, response.Payload, options.OutDir, models.RenderContext{
			Tenant:       options.Tenant,
			Subscription: options.Subscription,
		}); err != nil {
			fmt.Fprintf(stderr, "error: %s\n", err)
			return 1
		}
	}

	rendered, err := output.RenderWithContext(options.Output, commandName, response.Payload, models.RenderContext{
		Tenant:       options.Tenant,
		Subscription: options.Subscription,
	})
	if err != nil {
		fmt.Fprintf(stderr, "error: %s\n", err)
		return 1
	}

	_, _ = io.WriteString(stdout, rendered)
	return 0
}

type Options struct {
	Tenant             string
	Subscription       string
	DevOpsOrganization string
	ChainFamily        string
	Output             models.OutputMode
	RoleTrustsMode     models.RoleTrustsMode
	OutDir             string
	Debug              bool
}

func parseOptions(commandName string, args []string, stderr io.Writer) (Options, error) {
	contract, ok := contracts.Command(commandName)
	if !ok {
		return Options{}, fmt.Errorf("unknown command %q", commandName)
	}

	options := Options{
		Output:             models.OutputTable,
		RoleTrustsMode:     models.RoleTrustsModeFast,
		DevOpsOrganization: strings.TrimSpace(os.Getenv("AZUREFOX_DEVOPS_ORG")),
	}

	flags := flag.NewFlagSet("ho-azure", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.Tenant, "tenant", "", "Azure tenant ID")
	flags.StringVar(&options.Subscription, "subscription", "", "Azure subscription ID")
	flags.StringVar(&options.DevOpsOrganization, "devops-organization", options.DevOpsOrganization, "Azure DevOps organization")
	flags.Func("output", "Output format: table, json, csv", func(value string) error {
		options.Output = models.OutputMode(strings.ToLower(value))
		if !options.Output.Valid() {
			return fmt.Errorf("invalid output %q; valid values: table, json, csv", value)
		}
		return nil
	})
	flags.StringVar(&options.OutDir, "outdir", "", "Output directory for emitted artifacts")
	flags.BoolVar(&options.Debug, "debug", false, "Enable verbose error output")

	for _, commandFlag := range contract.Flags {
		switch commandFlag.Name {
		case "mode":
			flags.Func(commandFlag.Name, commandFlag.Usage, func(value string) error {
				options.RoleTrustsMode = models.RoleTrustsMode(strings.ToLower(value))
				if !options.RoleTrustsMode.Valid() {
					return fmt.Errorf("invalid mode %q; valid values: fast, full, fast-old, full-old", value)
				}
				return nil
			})
		}
	}

	if commandName == "chains" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		if args[0] != "help" {
			options.ChainFamily = args[0]
		}
		args = args[1:]
	}

	if err := flags.Parse(args); err != nil {
		return Options{}, err
	}
	remainingArgs := flags.Args()
	if commandName == "chains" {
		switch len(remainingArgs) {
		case 0:
		case 1:
			if remainingArgs[0] != "help" {
				options.ChainFamily = remainingArgs[0]
			}
		default:
			return Options{}, fmt.Errorf("unexpected arguments: %s", strings.Join(remainingArgs, " "))
		}
	} else if len(remainingArgs) != 0 {
		return Options{}, fmt.Errorf("unexpected arguments: %s", strings.Join(remainingArgs, " "))
	}
	if options.OutDir != "" {
		options.OutDir = filepath.Clean(options.OutDir)
	}
	return options, nil
}

func (app *App) rootHelp() string {
	var builder strings.Builder
	builder.WriteString("HO-Azure Help\n\n")
	builder.WriteString("Attack-path-focused Azure recon with flat commands and scoped help.\n\n")
	builder.WriteString("Usage:\n")
	builder.WriteString("  ho-azure help\n")
	builder.WriteString("  ho-azure help <section>\n")
	builder.WriteString("  ho-azure help <command>\n")
	builder.WriteString("  ho-azure -h <section>\n")
	builder.WriteString("  ho-azure -h <command>\n")
	builder.WriteString("  ho-azure <command> --help\n\n")
	builder.WriteString("Global flags:\n")
	builder.WriteString("  --tenant string\n")
	builder.WriteString("  --subscription string\n")
	builder.WriteString("  --devops-organization string\n")
	builder.WriteString("  --output string\n")
	builder.WriteString("  --outdir string\n")
	builder.WriteString("  --debug\n\n")
	builder.WriteString("Sections:\n")
	for _, section := range sectionOrder {
		topic := sectionHelpTopics[section]
		builder.WriteString(fmt.Sprintf("  %s: %s\n", section, topic.Summary))
	}
	builder.WriteString("\nCommands:\n")
	for _, contract := range contracts.ImplementedCommands() {
		builder.WriteString(fmt.Sprintf("  %s: %s\n", contract.Name, contract.OperatorQuestion))
	}
	builder.WriteString("\nNotes:\n")
	builder.WriteString("  - Shared flags such as --tenant, --subscription, --output, and --outdir work before or after the command.\n")
	builder.WriteString("  - Grouped chains help stays available while additional chain families land.\n")
	builder.WriteString("  - Default output prefers exact claims when proven and bounded weaker claims when they stay honest and useful.\n")
	return builder.String()
}

func (app *App) helpTopic(name string) (string, bool) {
	if _, ok := contracts.Command(name); ok {
		return app.commandHelp(name), true
	}
	if _, ok := sectionHelpTopics[name]; ok {
		return app.sectionHelp(name), true
	}
	return "", false
}

func (app *App) commandHelp(name string) string {
	contract, ok := contracts.Command(name)
	if !ok {
		return fmt.Sprintf("unknown command %q\n", name)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("HO-Azure Help :: %s\n\n", contract.Name))
	builder.WriteString(contract.OperatorQuestion + "\n\n")
	builder.WriteString(fmt.Sprintf("Status: %s.\n", commandStatusLabel(contract.Status)))
	builder.WriteString(fmt.Sprintf("Section: %s\n", contract.Section))
	builder.WriteString(fmt.Sprintf("Model: %s\n", contract.Model))
	builder.WriteString("Output highlights:\n")
	for _, field := range contract.TopLevelFields {
		builder.WriteString(fmt.Sprintf("  %s\n", field))
	}
	if len(contract.Flags) > 0 {
		builder.WriteString("\nCommand flags:\n")
		for _, commandFlag := range contract.Flags {
			builder.WriteString(fmt.Sprintf("  --%s string   %s\n", commandFlag.Name, commandFlag.Usage))
		}
	}
	builder.WriteString("\nExample:\n")
	builder.WriteString(fmt.Sprintf("  %s\n", commandExample(contract.Name)))
	if name == "chains" {
		builder.WriteString("\nUsage:\n  ho-azure chains [family|help] [flags]\n")
		builder.WriteString("\nCurrent families:\n")
		for _, familyName := range contracts.FamilyNames() {
			family, ok := contracts.Family(familyName)
			if !ok {
				continue
			}
			builder.WriteString(fmt.Sprintf("  %s: %s\n", family.Name, family.Summary))
		}
	}
	builder.WriteString("\nNotes:\n")
	builder.WriteString("  - Shared flags such as --tenant, --subscription, --output, and --outdir work before or after the command.\n")
	builder.WriteString("  - Table output is the primary parity surface; JSON may carry extra durable fields when that adds value without spamming the operator.\n")
	return builder.String()
}

func (app *App) sectionHelp(name string) string {
	topic, ok := sectionHelpTopics[name]
	if !ok {
		return fmt.Sprintf("unknown section %q\n", name)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("HO-Azure Help :: %s\n\n", name))
	builder.WriteString(topic.Summary + "\n\n")
	builder.WriteString(fmt.Sprintf("Operator goal: %s\n\n", topic.OperatorGoal))
	builder.WriteString("Implemented commands:\n")

	count := 0
	for _, contract := range contracts.ImplementedCommands() {
		if contract.Section != name {
			continue
		}
		builder.WriteString(fmt.Sprintf("  %s: %s\n", contract.Name, contract.OperatorQuestion))
		count++
	}
	if count == 0 {
		builder.WriteString("  none yet\n")
	}

	builder.WriteString("\nExamples:\n")
	builder.WriteString(fmt.Sprintf("  ho-azure help %s\n", name))
	return builder.String()
}

func (app *App) isHelpTopic(token string) bool {
	if _, ok := sectionHelpTopics[token]; ok {
		return true
	}
	_, ok := contracts.Command(token)
	return ok
}

func commandExample(name string) string {
	switch name {
	case "devops":
		return "ho-azure --devops-organization contoso devops --output table"
	case "role-trusts":
		return "ho-azure role-trusts --mode full --output table"
	case "chains":
		return "ho-azure chains deployment-path --output table"
	default:
		return fmt.Sprintf("ho-azure %s --output table", name)
	}
}

func isHelpFlag(arg string) bool {
	_, ok := helpFlags[arg]
	return ok
}

func commandStatusLabel(status string) string {
	switch status {
	case contracts.StatusImplemented:
		return "implemented command"
	case contracts.StatusPlaceholder:
		return "placeholder contract"
	default:
		return status
	}
}
