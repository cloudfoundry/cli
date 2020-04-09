package truncate

import (
	"reflect"
)

// Value recursively walks through the value provided by `v` and truncates
// any strings longer than `maxLength`.
// Example:
//	 type foobar struct{A string; B string}
//   truncate.Value(foobar{A:"foo",B:"bar"}, 20) == foobar{A:"foo",B:"bar"}
//   truncate.Value(foobar{A:strings.Repeat("a", 25),B:"bar"}, 20) == foobar{A:"aaaaaaaa-(truncated)",B:"bar"}
func Value(v interface{}, maxLength int) interface{} {
	rv := reflect.ValueOf(v)
	tv := truncateValue(rv, maxLength)
	if rv != tv {
		return tv.Interface()
	}
	return v
}

func truncateValue(rv reflect.Value, maxLength int) reflect.Value {
	if maxLength <= 0 {
		return rv
	}

	switch rv.Kind() {
	case reflect.Interface:
		return truncateInterface(rv, maxLength)
	case reflect.Ptr:
		return truncatePtr(rv, maxLength)
	case reflect.Struct:
		return truncateStruct(rv, maxLength)
	case reflect.Map:
		return truncateMap(rv, maxLength)
	case reflect.Array:
		return truncateArray(rv, maxLength)
	case reflect.Slice:
		return truncateSlice(rv, maxLength)
	case reflect.String:
		return truncateString(rv, maxLength)
	}
	return rv
}

func truncateInterface(rv reflect.Value, maxLength int) reflect.Value {
	tv := truncateValue(rv.Elem(), maxLength)
	if tv != rv.Elem() {
		return tv
	}
	return rv
}

func truncatePtr(rv reflect.Value, maxLength int) reflect.Value {
	tv := truncateValue(rv.Elem(), maxLength)
	if rv.Elem() != tv {
		tvp := reflect.New(rv.Elem().Type())
		tvp.Elem().Set(tv)
		return tvp
	}
	return rv
}

func truncateStruct(rv reflect.Value, maxLength int) reflect.Value {
	numFields := rv.NumField()
	fields := make([]reflect.Value, numFields)
	changed := false
	for i := 0; i < numFields; i++ {
		fv := rv.Field(i)
		tv := truncateValue(fv, maxLength)
		if fv != tv {
			changed = true
		}
		fields[i] = tv
	}
	if changed {
		nv := reflect.New(rv.Type()).Elem()
		for i, fv := range fields {
			nv.Field(i).Set(fv)
		}
		return nv
	}
	return rv
}

func truncateMap(rv reflect.Value, maxLength int) reflect.Value {
	keys := rv.MapKeys()
	truncatedMap := make(map[reflect.Value]reflect.Value)
	changed := false
	for _, key := range keys {
		mapV := rv.MapIndex(key)
		tv := truncateValue(mapV, maxLength)
		if mapV != tv {
			changed = true
		}
		truncatedMap[key] = tv
	}
	if changed {
		nv := reflect.MakeMap(rv.Type())
		for k, v := range truncatedMap {
			nv.SetMapIndex(k, v)
		}
		return nv
	}
	return rv

}

func truncateArray(rv reflect.Value, maxLength int) reflect.Value {
	return truncateList(rv, maxLength, func(size int) reflect.Value {
		arrayType := reflect.ArrayOf(size, rv.Index(0).Type())
		return reflect.New(arrayType).Elem()
	})
}

func truncateSlice(rv reflect.Value, maxLength int) reflect.Value {
	return truncateList(rv, maxLength, func(size int) reflect.Value {
		return reflect.MakeSlice(rv.Type(), size, size)
	})
}

func truncateList(rv reflect.Value, maxLength int, newList func(size int) reflect.Value) reflect.Value {
	size := rv.Len()
	truncatedValues := make([]reflect.Value, size)
	changed := false
	for i := 0; i < size; i++ {
		elemV := rv.Index(i)
		tv := truncateValue(elemV, maxLength)
		if elemV != tv {
			changed = true
		}
		truncatedValues[i] = tv
	}
	if changed {
		nv := newList(size)
		for i, v := range truncatedValues {
			nv.Index(i).Set(v)
		}
		return nv
	}
	return rv
}

func truncateString(rv reflect.Value, maxLength int) reflect.Value {
	s := String(rv.String(), maxLength)
	if s != rv.String() {
		return reflect.ValueOf(s)
	}
	return rv

}

const truncated = "-(truncated)"
const lenTruncated = len(truncated)

// String truncates long strings from the middle, but leaves strings shorter
// than `maxLength` untouched.
// If the string is shorter than the string "-(truncated)" and the string
// exceeds `maxLength`, the output will not be truncated.
// Example:
//   truncate.String(strings.Repeat("a", 25), 20) == "aaaaaaaa-(truncated)"
//   truncate.String("foobar", 20) == "foobar"
//   truncate.String("foobar", 5) == "foobar"
func String(s string, maxLength int) string {
	if maxLength <= 0 || len(s) < lenTruncated || len(s) <= maxLength {
		return s
	}

	strBytes := []byte(s)
	truncatedBytes := []byte(truncated)
	prefixLength := maxLength - lenTruncated
	prefix := strBytes[0:prefixLength]
	return string(append(prefix, truncatedBytes...))
}
