package utilities

import "slices"

func Contains[T comparable](dictionary []T, item T) bool {
	return slices.Contains(dictionary, item)
}
