package providers

import (
	"fmt"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

func monitoringSinkSummary(sink models.MonitoringSinkAsset) string {
	parts := []string{fmt.Sprintf("%s sink %q is visible through %s", sink.Kind, sink.Name, sink.VisibilitySource)}
	if sink.SentinelEnabled != nil {
		if *sink.SentinelEnabled {
			parts = append(parts, "Sentinel appears enabled")
		} else {
			parts = append(parts, "Sentinel enablement not visible")
		}
	}
	if sink.ReferenceCount > 0 {
		parts = append(parts, fmt.Sprintf("referenced by %d telemetry route(s)", sink.ReferenceCount))
	}
	return strings.Join(parts, "; ") + "."
}

func monitoringSinkSort(sinks []models.MonitoringSinkAsset) {
	sort.SliceStable(sinks, func(i, j int) bool {
		if monitoringSinkRank(sinks[i]) != monitoringSinkRank(sinks[j]) {
			return monitoringSinkRank(sinks[i]) > monitoringSinkRank(sinks[j])
		}
		if sinks[i].Kind != sinks[j].Kind {
			return sinks[i].Kind < sinks[j].Kind
		}
		return sinks[i].Name < sinks[j].Name
	})
}

func MonitoringSinksFromDCRReferences(dcrs []models.DCRAsset) []models.MonitoringSinkAsset {
	sinks := []models.MonitoringSinkAsset{}
	monitoringSinksEnsureDCRDestinations(&sinks, dcrs)
	monitoringSinksAttachDCRReferences(sinks, dcrs)
	return monitoringSinksFinalize(sinks)
}

func MonitoringSinksFromDiagnosticReferences(sources []models.DiagnosticSettingsSource) []models.MonitoringSinkAsset {
	sinks := []models.MonitoringSinkAsset{}
	monitoringSinksEnsureDiagnosticDestinations(&sinks, sources)
	monitoringSinksAttachDiagnosticReferences(sinks, sources)
	return monitoringSinksFinalize(sinks)
}

func monitoringSinksFinalize(sinks []models.MonitoringSinkAsset) []models.MonitoringSinkAsset {
	for index := range sinks {
		sinks[index].ReferenceCount = len(sinks[index].References)
		sinks[index].RelatedIDs = sortedUniqueStrings(append(sinks[index].RelatedIDs, monitoringSinkReferenceIDs(sinks[index].References)...))
		sinks[index].Summary = monitoringSinkSummary(sinks[index])
	}
	monitoringSinkSort(sinks)
	return sinks
}

func monitoringSinkRank(sink models.MonitoringSinkAsset) int {
	rank := sink.ReferenceCount
	switch sink.Kind {
	case "sentinel":
		rank += 5
	case "logAnalytics":
		rank += 4
	case "eventHubs":
		rank += 3
	case "storage":
		rank += 2
	}
	return rank
}
