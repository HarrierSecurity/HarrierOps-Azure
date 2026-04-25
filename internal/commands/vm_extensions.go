package commands

import (
	"context"
	"sort"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func vmExtensionsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.VMExtensions(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		extensions := append([]models.VMExtensionAsset{}, facts.VMExtensions...)
		sort.SliceStable(extensions, func(i int, j int) bool {
			return vmExtensionLess(extensions[i], extensions[j])
		})

		return models.VMExtensionsOutput{
			Findings:     []models.Finding{},
			Issues:       facts.Issues,
			Metadata:     withRuntimeArtifactContext(runtimeCommandMetadata("vm-extensions", now, facts.TenantID, facts.SubscriptionID), request, facts.CurrentPrincipal, facts.AuthMode, facts.TokenSource),
			VMExtensions: extensions,
		}, nil
	}
}

func vmExtensionLess(left, right models.VMExtensionAsset) bool {
	leftCustomScript := vmExtensionIsCustomScript(left)
	rightCustomScript := vmExtensionIsCustomScript(right)
	if leftCustomScript != rightCustomScript {
		return leftCustomScript
	}
	leftCommand := stringPtrValue(left.CommandClue) != ""
	rightCommand := stringPtrValue(right.CommandClue) != ""
	if leftCommand != rightCommand {
		return leftCommand
	}
	leftSources := len(left.SourceClues) + len(left.FileURIHosts)
	rightSources := len(right.SourceClues) + len(right.FileURIHosts)
	if leftSources != rightSources {
		return leftSources > rightSources
	}
	if left.TargetKind != right.TargetKind {
		return left.TargetKind < right.TargetKind
	}
	if left.TargetName != right.TargetName {
		return left.TargetName < right.TargetName
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func vmExtensionIsCustomScript(extension models.VMExtensionAsset) bool {
	extensionType := strings.ToLower(strings.TrimSpace(stringPtrValue(extension.ExtensionType)))
	return strings.Contains(extensionType, "customscript")
}
