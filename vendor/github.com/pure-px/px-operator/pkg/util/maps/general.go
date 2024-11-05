package maps

import (
	"cmp"
	"sort"
)

// Keys returns all keys in the map, in unsorted order.
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// KeysSorted returns all keys in the map, in sorted order.
// - note, may not work on all types of maps/keys
// - note2, needs Golang 1.21 to compile due to cmp.Ordered
func KeysSorted[K cmp.Ordered, V any](m map[K]V) []K {
	keys := Keys(m)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}
