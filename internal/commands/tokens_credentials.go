package commands

import (
	"context"
	"sort"
	"strings"
	"time"
	"unicode"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func tokensCredentialsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.TokensCredentials(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		surfaces := append([]models.TokenCredentialSurfaceSummary{}, facts.Surfaces...)
		sort.SliceStable(surfaces, func(i int, j int) bool {
			left := surfaces[i]
			right := surfaces[j]

			if tokenCredentialPriorityRank(left.Priority) != tokenCredentialPriorityRank(right.Priority) {
				return tokenCredentialPriorityRank(left.Priority) < tokenCredentialPriorityRank(right.Priority)
			}
			if left.AssetName != right.AssetName {
				return left.AssetName < right.AssetName
			}
			if left.SurfaceType != right.SurfaceType {
				return left.SurfaceType < right.SurfaceType
			}
			return left.OperatorSignal < right.OperatorSignal
		})

		return models.TokensCredentialsOutput{
			Findings: buildTokensCredentialsFindings(surfaces),
			Issues:   facts.Issues,
			Metadata: scopedMetadata(now, request, facts.TenantID, facts.SubscriptionID, "tokens-credentials"),
			Surfaces: surfaces,
		}, nil
	}
}

func buildTokensCredentialsFindings(surfaces []models.TokenCredentialSurfaceSummary) []models.TokenCredentialFinding {
	findings := make([]models.TokenCredentialFinding, 0, len(surfaces))
	for _, surface := range surfaces {
		suffix := tokenCredentialFindingSuffix(surface)

		switch surface.SurfaceType {
		case models.TokenCredentialSurfacePlainTextSecret:
			findings = append(findings, models.TokenCredentialFinding{
				Description: surface.Summary,
				ID:          "tokens-credentials-plain-text-" + suffix,
				RelatedIDs:  append([]string{}, surface.RelatedIDs...),
				Severity:    "high",
				Title:       "Credential-like value is exposed in plain-text app settings",
			})
		case models.TokenCredentialSurfaceKeyVaultReference:
			findings = append(findings, models.TokenCredentialFinding{
				Description: surface.Summary,
				ID:          "tokens-credentials-keyvault-ref-" + suffix,
				RelatedIDs:  append([]string{}, surface.RelatedIDs...),
				Severity:    "low",
				Title:       "Workload setting depends on Key Vault-backed secret retrieval",
			})
		case models.TokenCredentialSurfaceManagedIdentityToken:
			title := "Workload can mint tokens with managed identity"
			severity := "medium"
			if surface.PubliclyReachable || tokenCredentialPriorityRank(surface.Priority) == 0 {
				title = "Publicly reachable workload can mint tokens with managed identity"
				severity = "high"
			}
			findings = append(findings, models.TokenCredentialFinding{
				Description: surface.Summary,
				ID:          "tokens-credentials-managed-identity-" + suffix,
				RelatedIDs:  append([]string{}, surface.RelatedIDs...),
				Severity:    severity,
				Title:       title,
			})
		case models.TokenCredentialSurfaceDeploymentOutput:
			findings = append(findings, models.TokenCredentialFinding{
				Description: surface.Summary,
				ID:          "tokens-credentials-deployment-output-" + suffix,
				RelatedIDs:  append([]string{}, surface.RelatedIDs...),
				Severity:    "medium",
				Title:       "Deployment history records output values",
			})
		case models.TokenCredentialSurfaceLinkedDeploymentAsset:
			findings = append(findings, models.TokenCredentialFinding{
				Description: surface.Summary,
				ID:          "tokens-credentials-linked-content-" + suffix,
				RelatedIDs:  append([]string{}, surface.RelatedIDs...),
				Severity:    "low",
				Title:       "Deployment history references remote template or parameter content",
			})
		default:
			findings = append(findings, models.TokenCredentialFinding{
				Description: surface.Summary,
				ID:          "tokens-credentials-unclassified-" + suffix,
				RelatedIDs:  append([]string{}, surface.RelatedIDs...),
				Severity:    "low",
				Title:       "Credential-bearing surface needs classification review",
			})
		}
	}
	return findings
}

func tokenCredentialPriorityRank(priority string) int {
	return exposurePriorityRank(priority)
}

func tokenCredentialFindingSuffix(surface models.TokenCredentialSurfaceSummary) string {
	parts := []string{
		surface.AssetID,
		surface.AccessPath,
		tokenCredentialSlug(surface.OperatorSignal),
	}
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, "-")
}

func tokenCredentialSlug(value string) string {
	var builder strings.Builder
	lastHyphen := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastHyphen = false
		case !lastHyphen:
			builder.WriteByte('-')
			lastHyphen = true
		}
	}
	return strings.Trim(builder.String(), "-")
}
