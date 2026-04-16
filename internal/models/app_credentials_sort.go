package models

import "sort"

func SortAppCredentialRows(rows []AppCredentialSummary) {
	sort.SliceStable(rows, func(i, j int) bool {
		left := rows[i]
		right := rows[j]

		if appCredentialRowClassRank(left.RowClass) != appCredentialRowClassRank(right.RowClass) {
			return appCredentialRowClassRank(left.RowClass) < appCredentialRowClassRank(right.RowClass)
		}
		if left.TargetObjectName != right.TargetObjectName {
			return left.TargetObjectName < right.TargetObjectName
		}
		if left.TargetObjectType != right.TargetObjectType {
			return left.TargetObjectType < right.TargetObjectType
		}
		return left.TargetObjectID < right.TargetObjectID
	})
}

func appCredentialRowClassRank(rowClass string) int {
	switch rowClass {
	case "directly_addable_federated_trust":
		return 0
	case "directly_addable":
		return 1
	case "federated_trust_present":
		return 2
	case "existing_credential":
		return 3
	case "control_context_only":
		return 4
	default:
		return 9
	}
}
