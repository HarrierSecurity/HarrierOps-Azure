package render

import (
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

func monitoringSinksTable(payload models.MonitoringSinksOutput) string {
	rows := make([][]string, 0, len(payload.Sinks))
	for _, sink := range payload.Sinks {
		rows = append(rows, []string{
			sink.Name,
			sink.Kind,
			sink.VisibilitySource,
			fmt.Sprintf("%d", sink.ReferenceCount),
			monitoringSinkSentinelText(sink),
		})
	}
	output := renderListTable(
		"ho-azure monitoring-sinks",
		[]string{"sink", "kind", "visible from", "routes", "sentinel"},
		rows,
		[]string{"no visible monitoring sinks", "", "", "", ""},
		monitoringSinksTakeaway(payload),
	)
	output += "\nNot collected by default\n"
	output += renderAlignedPipeTable(
		[]string{"item", "classification", "reason"},
		[][]string{
			{"expected SOC baseline", "proof boundary", "Visible sinks and declared telemetry routes do not prove which sink defenders expect."},
			{"sink contents", "proof boundary", "The helper does not query Log Analytics, Storage, Event Hub, or partner sink contents."},
			{"detector wiring", "proof boundary", "Sentinel enablement is posture only; rule dependencies and alert behavior are not inspected."},
		},
	)
	return output
}

func monitoringSinkSentinelText(sink models.MonitoringSinkAsset) string {
	if sink.SentinelEnabled == nil {
		return "unknown"
	}
	if *sink.SentinelEnabled {
		return "visible"
	}
	return "not visible"
}

func monitoringSinksTakeaway(payload models.MonitoringSinksOutput) string {
	referenced := 0
	kinds := []string{}
	for _, sink := range payload.Sinks {
		if sink.ReferenceCount > 0 {
			referenced++
		}
		kinds = append(kinds, sink.Kind)
	}
	return fmt.Sprintf("%d visible or declared monitoring sink(s); %d referenced by DCR or diagnostic-settings routes; kinds: %s.", len(payload.Sinks), referenced, strings.Join(dedupeStrings(kinds), ", "))
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
