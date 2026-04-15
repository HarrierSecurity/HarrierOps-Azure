package commands

import "strings"

func normalizedLower(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func exposurePriorityRank(value string) int {
	switch normalizedLower(value) {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	default:
		return 9
	}
}
