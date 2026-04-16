package providers

import (
	"fmt"
	"os"
	"strings"
)

const providerModeEnv = "AZUREFOX_PROVIDER"

func NewProviderFromEnvironment() (Provider, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv(providerModeEnv)))
	switch mode {
	case "", "azure":
		return NewAzureProvider(), nil
	case "static":
		return NewStaticProvider(), nil
	default:
		return nil, fmt.Errorf(
			"invalid %s value %q; valid values: azure, static",
			providerModeEnv,
			mode,
		)
	}
}
