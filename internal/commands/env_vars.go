package commands

import (
	"context"
	"sort"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func envVarsHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.EnvVars(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		envVars := append([]models.EnvVarSummary{}, facts.EnvVars...)
		sort.SliceStable(envVars, func(i int, j int) bool {
			left := envVars[i]
			right := envVars[j]

			leftSensitivePlain := left.LooksSensitive && left.ValueType == "plain-text"
			rightSensitivePlain := right.LooksSensitive && right.ValueType == "plain-text"
			if leftSensitivePlain != rightSensitivePlain {
				return leftSensitivePlain
			}

			leftKeyVault := left.ValueType == "keyvault-ref"
			rightKeyVault := right.ValueType == "keyvault-ref"
			if leftKeyVault != rightKeyVault {
				return leftKeyVault
			}

			if left.AssetName != right.AssetName {
				return left.AssetName < right.AssetName
			}

			return left.SettingName < right.SettingName
		})

		return models.EnvVarsOutput{
			EnvVars:  envVars,
			Findings: buildEnvVarFindings(envVars),
			Issues:   facts.Issues,
			Metadata: commandMetadata("env-vars", now, request, facts.TenantID, facts.SubscriptionID, ""),
		}, nil
	}
}

func buildEnvVarFindings(envVars []models.EnvVarSummary) []models.EnvVarFinding {
	findings := make([]models.EnvVarFinding, 0, len(envVars))
	for _, envVar := range envVars {
		if envVar.LooksSensitive && envVar.ValueType == "plain-text" {
			findings = append(findings, models.EnvVarFinding{
				Description: envVar.AssetKind + " '" + envVar.AssetName + "' stores setting '" + envVar.SettingName + "' as plain-text management-plane config.",
				ID:          "env-var-plain-sensitive-" + envVar.AssetID + "-" + envVar.SettingName,
				RelatedIDs:  append([]string{}, envVar.RelatedIDs...),
				Severity:    "medium",
				Title:       "Sensitive-looking app setting is stored in plain text",
			})
		}

		if envVar.ValueType == "keyvault-ref" {
			description := envVar.AssetKind + " '" + envVar.AssetName + "' maps setting '" + envVar.SettingName + "' to Key Vault-backed configuration"
			if envVar.ReferenceTarget != nil && *envVar.ReferenceTarget != "" {
				description += " (" + *envVar.ReferenceTarget + ")"
			}
			description += "."

			findings = append(findings, models.EnvVarFinding{
				Description: description,
				ID:          "env-var-keyvault-ref-" + envVar.AssetID + "-" + envVar.SettingName,
				RelatedIDs:  append([]string{}, envVar.RelatedIDs...),
				Severity:    "low",
				Title:       "App setting references Key Vault",
			})
		}
	}
	return findings
}
