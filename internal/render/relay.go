package render

import (
	"fmt"
	"strings"

	"harrierops-azure/internal/models"
)

func relayTable(payload models.RelayOutput) string {
	rows := make([][]string, 0, len(payload.Namespaces))
	for _, namespace := range payload.Namespaces {
		rows = append(rows, []string{
			namespace.Name,
			namespace.ResourceGroup,
			intPtrString(namespace.HybridConnectionCount),
			intPtrString(namespace.AuthorizationRuleCount),
			relayListenerSummary(namespace),
			relayAppServiceAttachmentSummary(namespace),
			valueOrEmpty(namespace.ServiceBusEndpoint),
		})
	}
	output := renderListTable(
		"ho-azure relay",
		[]string{"namespace", "resource group", "hybrid connections", "auth rules", "listeners", "app attachments", "endpoint"},
		rows,
		[]string{"No Relay namespaces were visible from current scope.", "", "", "", "", "", ""},
		relayTakeaway(payload),
	)
	output += "\nNot collected by default:\n"
	output += "- authorization keys: recon safety; authorization rules are counted, but key material is not retrieved or printed\n"
	output += "- listener runtime state: proof boundary; listener counts do not prove a current listener process, host, or session\n"
	output += "- backend process and traffic contents: proof boundary; Relay posture does not identify the private backend process or inspect traffic payloads\n"
	return output
}

func relayListenerSummary(namespace models.RelayNamespaceAsset) string {
	total := 0
	known := false
	for _, connection := range namespace.HybridConnections {
		if connection.ListenerCount == nil {
			continue
		}
		known = true
		total += *connection.ListenerCount
	}
	if !known {
		return "unknown"
	}
	return fmt.Sprintf("%d", total)
}

func relayAppServiceAttachmentSummary(namespace models.RelayNamespaceAsset) string {
	values := []string{}
	for _, connection := range namespace.HybridConnections {
		for _, app := range connection.AppServiceAttachments {
			values = append(values, connection.Name+"->"+app)
		}
	}
	if len(values) == 0 {
		return "none visible"
	}
	return strings.Join(values, "; ")
}

func relayTakeaway(payload models.RelayOutput) string {
	if len(payload.Namespaces) == 0 {
		return "0 Relay namespaces visible; no Azure Relay pathmasking helper surface was confirmed from current scope."
	}
	return fmt.Sprintf("%d Relay namespace(s) visible; review Hybrid Connections for cloud rendezvous and private-path masking posture.", len(payload.Namespaces))
}
