package artifacts

import (
	"encoding/json"
	"os"
	"path/filepath"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/output"
)

type Paths struct {
	Loot  string
	JSON  string
	Table string
	CSV   string
}

func Write(command string, payload any, outDir string, context models.RenderContext) (Paths, error) {
	paths := Paths{
		Loot:  filepath.Join(outDir, "loot", command+".json"),
		JSON:  filepath.Join(outDir, "json", command+".json"),
		Table: filepath.Join(outDir, "table", command+".txt"),
		CSV:   filepath.Join(outDir, "csv", command+".csv"),
	}

	for _, path := range []string{paths.Loot, paths.JSON, paths.Table, paths.CSV} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return Paths{}, err
		}
	}

	lootPayload, err := buildLootPayload(payload)
	if err != nil {
		return Paths{}, err
	}
	lootBytes, err := json.MarshalIndent(lootPayload, "", "  ")
	if err != nil {
		return Paths{}, err
	}
	jsonContent, err := output.RenderWithContext(models.OutputJSON, command, payload, context)
	if err != nil {
		return Paths{}, err
	}
	tableContent, err := output.RenderWithContext(models.OutputTable, command, payload, context)
	if err != nil {
		return Paths{}, err
	}
	csvContent, err := output.RenderWithContext(models.OutputCSV, command, payload, context)
	if err != nil {
		return Paths{}, err
	}

	if err := os.WriteFile(paths.Loot, append(lootBytes, '\n'), 0o644); err != nil {
		return Paths{}, err
	}
	if err := os.WriteFile(paths.JSON, []byte(jsonContent), 0o644); err != nil {
		return Paths{}, err
	}
	if err := os.WriteFile(paths.Table, []byte(tableContent), 0o644); err != nil {
		return Paths{}, err
	}
	if err := os.WriteFile(paths.CSV, []byte(csvContent), 0o644); err != nil {
		return Paths{}, err
	}

	return paths, nil
}

func buildLootPayload(payload any) (any, error) {
	switch out := payload.(type) {
	case models.WhoAmIOutput:
		type lootMetadata struct {
			SchemaVersion string `json:"schema_version"`
			Command       string `json:"command"`
		}
		type whoAmILoot struct {
			EffectiveScopes []models.ScopeRef      `json:"effective_scopes"`
			Issues          []models.Issue         `json:"issues"`
			Metadata        lootMetadata           `json:"metadata"`
			Principal       models.Principal       `json:"principal"`
			Subscription    models.SubscriptionRef `json:"subscription"`
			TenantID        string                 `json:"tenant_id"`
		}
		return whoAmILoot{
			EffectiveScopes: out.EffectiveScopes,
			Issues:          out.Issues,
			Metadata: lootMetadata{
				SchemaVersion: out.Metadata.Metadata.SchemaVersion,
				Command:       out.Metadata.Metadata.Command,
			},
			Principal:    out.Principal,
			Subscription: out.Subscription,
			TenantID:     out.TenantID,
		}, nil
	default:
		return payload, nil
	}
}
