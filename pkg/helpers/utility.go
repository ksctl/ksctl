package helpers

func DeepCopyMap[K comparable, V string | any](src map[K]V) (dest map[K]V) {
	dest = make(map[K]V)

	for k, v := range src {
		dest[k] = v
	}

	return
}

func DeepCopySlice[T any](src []T) (dest []T) {
	dest = make([]T, len(src))
	_ = copy(dest, src)

	return
}
