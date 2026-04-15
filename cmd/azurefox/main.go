package main

import (
	"fmt"
	"os"
	"time"

	"harrierops-azure/internal/cli"
	"harrierops-azure/internal/commands"
	"harrierops-azure/internal/providers"
)

func main() {
	provider, err := providers.NewProviderFromEnvironment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	app := cli.New(
		commands.NewRegistry(
			provider,
			func() time.Time { return time.Now().UTC() },
		),
	)
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
