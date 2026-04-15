package commands

import "sort"

func sortedByLess[T any](items []T, less func(T, T) bool) []T {
	sorted := append([]T{}, items...)
	sort.SliceStable(sorted, func(i int, j int) bool {
		return less(sorted[i], sorted[j])
	})
	return sorted
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
