package render

import (
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

type familySurfaceTableConfig struct {
	Title             string
	EmptyHeaders      []string
	EmptyRow          []string
	EmptyTakeaway     string
	CapabilityTitle   string
	CapabilitySteps   []models.FamilyCapabilityStep
	MultiTargetNote   string
	TargetCount       int
	Explanation       string
	ReducedVisibility string
	InventoryTitle    string
	InventoryHeaders  []string
	InventoryRows     [][]string
	BoundaryNotes     []models.FamilyBoundaryNote
}

func renderFamilySurfaceTable(config familySurfaceTableConfig) string {
	if config.TargetCount == 0 {
		return renderListTable(
			config.Title,
			config.EmptyHeaders,
			nil,
			config.EmptyRow,
			config.EmptyTakeaway,
		)
	}

	lines := []string{
		config.CapabilityTitle,
		"",
		renderPersistenceSectionTable(
			[]string{"action", "api surface", "status"},
			familyCapabilityRows(config.CapabilitySteps),
		),
	}
	if config.TargetCount > 1 && config.MultiTargetNote != "" {
		lines = append(lines, config.MultiTargetNote)
	}
	explanation := config.Explanation
	if !familyHasActionableCapability(config.CapabilitySteps) && config.ReducedVisibility != "" {
		explanation = config.ReducedVisibility
	}
	lines = append(lines,
		explanation,
		"",
		config.InventoryTitle,
		renderAlignedPipeTable(config.InventoryHeaders, config.InventoryRows),
		"",
		"Not collected by default",
		familyBoundaryRows(config.BoundaryNotes),
	)
	return strings.Join(lines, "\n")
}

func familyCapabilityRows(steps []models.FamilyCapabilityStep) [][]string {
	rows := make([][]string, 0, len(steps))
	for _, step := range steps {
		rows = append(rows, []string{step.Action, step.APISurface, step.Status})
	}
	return rows
}

func familyBoundaryRows(notes []models.FamilyBoundaryNote) string {
	if len(notes) == 0 {
		return "none"
	}
	rows := make([][]string, 0, len(notes))
	for _, note := range notes {
		rows = append(rows, []string{note.Name, note.Classification, note.Reason})
	}
	return renderAlignedPipeTable([]string{"item", "classification", "reason"}, rows)
}

func familyRoleSummary(context *models.FamilyRoleContext) string {
	if context == nil {
		return "current identity context not visible"
	}
	return context.Summary
}

func familyRoleControlLabel(context *models.FamilyRoleContext) string {
	if context == nil {
		return "not visible"
	}
	if context.ControlLabel == "" {
		return "not proven"
	}
	return context.ControlLabel
}

func familyHasActionableCapability(steps []models.FamilyCapabilityStep) bool {
	for _, step := range steps {
		if step.CanAct {
			return true
		}
	}
	return false
}

func familyReducedVisibilityExplanation(surface string, family string, path string, context *models.FamilyRoleContext) string {
	lines := []string{
		"",
		"Operator read",
		fmt.Sprintf("Current identity can see %s posture, but no change-control path is proven from current evidence.", surface),
		fmt.Sprintf("Higher permissions are required to use this as a %s path.", path),
		"Current identity: " + familyRoleSummary(context),
		fmt.Sprintf("First boundary: this is %s posture only; the walkthrough stops before operator actions that require write/control permissions.", family),
	}
	return strings.Join(lines, "\n")
}
