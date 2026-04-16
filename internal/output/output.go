package output

import (
	"fmt"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/render"
)

func Render(mode models.OutputMode, command string, payload any) (string, error) {
	return RenderWithContext(mode, command, payload, models.RenderContext{})
}

func RenderWithContext(mode models.OutputMode, command string, payload any, context models.RenderContext) (string, error) {
	switch mode {
	case models.OutputJSON:
		return render.JSON(payload)
	case models.OutputCSV:
		return render.CSV(command, payload)
	case models.OutputTable:
		return render.Table(command, payload, context)
	default:
		return "", fmt.Errorf("unsupported output mode %q", mode)
	}
}
