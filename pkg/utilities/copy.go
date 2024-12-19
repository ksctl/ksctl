package utilities

import (
	"reflect"
)

func DeepCopyMap[K comparable, V string | any](src map[K]V) (dest map[K]V) {
	dest = make(map[K]V)

	for k, v := range src {
		dest[k] = v
	}

	return
}

func CopySrcToDestPreservingDestVals[K comparable, V any](dest, src map[K]V) {
	for key, srcVal := range src {
		if destVal, exists := dest[key]; exists {

			srcMap, srcOk := any(srcVal).(map[K]V)
			destMap, destOk := any(destVal).(map[K]V)

			if srcOk && destOk {
				CopySrcToDestPreservingDestVals(destMap, srcMap)
				dest[key] = any(destMap).(V)
			} else if reflect.TypeOf(srcVal).Kind() == reflect.Slice && reflect.TypeOf(destVal).Kind() == reflect.Slice {
				mergedSlice := deduplicateAndMergeSlices(destVal, srcVal)
				dest[key] = mergedSlice
			} else {
				dest[key] = srcVal
			}

		} else {
			dest[key] = srcVal
		}
	}
}

func contains(slice reflect.Value, elem interface{}) bool {
	for i := 0; i < slice.Len(); i++ {
		if reflect.DeepEqual(slice.Index(i).Interface(), elem) {
			return true
		}
	}
	return false
}

func deduplicateAndMergeSlices[V any](destVal, srcVal V) V {
	destSlice := reflect.ValueOf(destVal)
	srcSlice := reflect.ValueOf(srcVal)
	mergedSlice := reflect.MakeSlice(destSlice.Type(), destSlice.Len(), destSlice.Len()+srcSlice.Len())

	reflect.Copy(mergedSlice, destSlice)

	for i := 0; i < srcSlice.Len(); i++ {
		elem := srcSlice.Index(i).Interface()
		if !contains(mergedSlice, elem) {
			mergedSlice = reflect.Append(mergedSlice, srcSlice.Index(i))
		}
	}

	return mergedSlice.Interface().(V)
}

func DeepCopySlice[T any](src []T) (dest []T) {
	dest = make([]T, len(src))
	_ = copy(dest, src)

	return
}
