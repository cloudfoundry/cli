package generic

import "reflect"

func IsSliceable(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Slice
}
