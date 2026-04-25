package commands

import (
	"context"
	"sort"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func vmsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.VMs(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		vmAssets := append([]models.VmAsset{}, facts.VMAssets...)
		sort.SliceStable(vmAssets, func(i int, j int) bool {
			left := vmAssets[i]
			right := vmAssets[j]

			if len(left.PublicIPs) != len(right.PublicIPs) {
				return len(left.PublicIPs) > len(right.PublicIPs)
			}
			if len(left.IdentityIDs) != len(right.IdentityIDs) {
				return len(left.IdentityIDs) > len(right.IdentityIDs)
			}

			leftTypeRank := vmAssetTypeRank(left.VMType)
			rightTypeRank := vmAssetTypeRank(right.VMType)
			if leftTypeRank != rightTypeRank {
				return leftTypeRank < rightTypeRank
			}
			if left.Name != right.Name {
				return left.Name < right.Name
			}
			return left.ID < right.ID
		})

		return models.VmsOutput{
			Findings: buildVMFindings(vmAssets),
			Issues:   facts.Issues,
			Metadata: withArtifactContext(commandMetadata("vms", now, request, facts.TenantID, facts.SubscriptionID, facts.TokenSource), request, facts.CurrentPrincipal, facts.AuthMode, facts.TokenSource),
			VMAssets: vmAssets,
		}, nil
	}
}

func buildVMFindings(vmAssets []models.VmAsset) []models.VmsFinding {
	findings := make([]models.VmsFinding, 0, len(vmAssets))
	for _, vm := range vmAssets {
		if len(vm.PublicIPs) == 0 || len(vm.IdentityIDs) == 0 {
			continue
		}
		findings = append(findings, models.VmsFinding{
			Description: "Workload '" + vm.Name + "' has public IP exposure and one or more managed " +
				"identities. Validate identity privileges and ingress hardening.",
			ID:         "vm-public-identity-" + vm.ID,
			RelatedIDs: append([]string{vm.ID}, vm.IdentityIDs...),
			Severity:   "medium",
			Title:      "Public workload with attached identity",
		})
	}
	return findings
}

func vmAssetTypeRank(value string) int {
	switch strings.ToLower(value) {
	case "vm":
		return 0
	case "vmss":
		return 1
	default:
		return 9
	}
}
