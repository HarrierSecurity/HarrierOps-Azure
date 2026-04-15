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

func New(registry *commands.Registry) *App {
	return &App{registry: registry}
}

func (app *App) Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = io.WriteString(stdout, app.rootHelp())
		return 0
	}

	if args[0] == "help" {
		if len(args) == 1 {
			_, _ = io.WriteString(stdout, app.rootHelp())
			return 0
		}
		_, _ = io.WriteString(stdout, app.commandHelp(args[1]))
		return 0
	}

	if strings.HasPrefix(args[0], "-") {
		fmt.Fprintln(stderr, "command must come first; use `azurefox <command> [flags]`")
		return 2
	}

	commandName := args[0]
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
		if _, err := artifacts.Write(commandName, response.Payload, options.OutDir); err != nil {
			fmt.Fprintf(stderr, "error: %s\n", err)
			return 1
		}
	}

	rendered, err := output.Render(options.Output, commandName, response.Payload)
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
	Output             models.OutputMode
	RoleTrustsMode     models.RoleTrustsMode
	OutDir             string
}

func parseOptions(commandName string, args []string, stderr io.Writer) (Options, error) {
	contract, ok := contracts.Command(commandName)
	if !ok {
		return Options{}, fmt.Errorf("unknown command %q", commandName)
	}

	options := Options{
		Output:         models.OutputTable,
		RoleTrustsMode: models.RoleTrustsModeFast,
	}

	flags := flag.NewFlagSet("azurefox", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.Tenant, "tenant", "", "Azure tenant ID")
	flags.StringVar(&options.Subscription, "subscription", "", "Azure subscription ID")
	flags.Func("output", "Output format: table, json, csv", func(value string) error {
		options.Output = models.OutputMode(strings.ToLower(value))
		if !options.Output.Valid() {
			return fmt.Errorf("invalid output %q; valid values: table, json, csv", value)
		}
		return nil
	})

	for _, commandFlag := range contract.Flags {
		switch commandFlag.Name {
		case "devops-organization":
			options.DevOpsOrganization = strings.TrimSpace(os.Getenv("AZUREFOX_DEVOPS_ORG"))
			flags.StringVar(&options.DevOpsOrganization, commandFlag.Name, "", commandFlag.Usage)
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
	flags.StringVar(&options.OutDir, "outdir", "", "Output directory for emitted artifacts")

	if err := flags.Parse(args); err != nil {
		return Options{}, err
	}
	if len(flags.Args()) != 0 {
		return Options{}, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if options.OutDir != "" {
		options.OutDir = filepath.Clean(options.OutDir)
	}
	return options, nil
}

func (app *App) rootHelp() string {
	var builder strings.Builder
	builder.WriteString("azurefox\n\n")
	builder.WriteString("Faithful Go rewrite scaffold for AzureFox.\n\n")
	builder.WriteString("Usage:\n  azurefox <command> [flags]\n\n")
	builder.WriteString("Global flags:\n")
	builder.WriteString("  --tenant string\n")
	builder.WriteString("  --subscription string\n")
	builder.WriteString("  --output string\n")
	builder.WriteString("  --outdir string\n")
	builder.WriteString("\n")
	builder.WriteString("Commands:\n")
	for _, contract := range app.registry.Commands() {
		builder.WriteString(fmt.Sprintf("  %-20s %-12s %s\n", contract.Name, contract.Status, contract.Section))
	}
	return builder.String()
}

func (app *App) commandHelp(name string) string {
	contract, ok := contracts.Command(name)
	if !ok {
		return fmt.Sprintf("unknown command %q\n", name)
	}
	help := fmt.Sprintf(
		"azurefox %s\n\nSection: %s\nStatus: %s\nModel: %s\nQuestion: %s\nTop-level fields: %s\n",
		contract.Name,
		contract.Section,
		contract.Status,
		contract.Model,
		contract.OperatorQuestion,
		strings.Join(contract.TopLevelFields, ", "),
	)
	if name == "role-trusts" {
	}
	if len(contract.Flags) > 0 {
		help += "\nCommand flags:\n"
		for _, commandFlag := range contract.Flags {
			help += fmt.Sprintf("  --%s string   %s\n", commandFlag.Name, commandFlag.Usage)
		}
	}
	return help
}
