package extract

import (
	"reflect"
	"strings"

	"code.cloudfoundry.org/cli/v9/util/unique"
)

type appender func(string)

func UniqueList(expr string, input interface{}) []string {
	return unique.StringSlice(List(expr, input))
}

func First(expr string, input interface{}) string {
	l := List(expr, input)
	if len(l) > 0 {
		return l[0]
	}
	return ""
}

func List(expr string, input interface{}) []string {
	var result []string
	app := func(v string) {
		result = append(result, v)
	}

	path := strings.Split(expr, ".")
	val := reflect.ValueOf(input)

	extract(app, path, val)
	return result
}

func extract(app appender, path []string, input reflect.Value) {
	switch input.Kind() {
	case reflect.Struct:
		extractStruct(app, path, input)
	case reflect.Slice:
		extractSlice(app, path, input)
	case reflect.Interface:
		extract(app, path, input.Elem())
	}
}

func extractSlice(app appender, path []string, input reflect.Value) {
	for i := 0; i < input.Len(); i++ {
		extract(app, path, input.Index(i))
	}
}

func extractStruct(app appender, path []string, input reflect.Value) {
	v := input.FieldByName(path[0])
	if v.IsValid() {
		switch v.Kind() {
		case reflect.String:
			app(v.String())
		default:
			extract(app, path[1:], v)
		}
	}
}
