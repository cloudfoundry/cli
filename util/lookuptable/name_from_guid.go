package lookuptable

import "reflect"

func NameFromGUID(input interface{}) map[string]string {
	val := reflect.ValueOf(input)
	if val.Kind() != reflect.Slice || val.Type().Elem().Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]string)
	for i := 0; i < val.Len(); i++ {
		element := val.Index(i)
		guid := element.FieldByName("GUID")
		name := element.FieldByName("Name")
		result[guid.String()] = name.String()
	}

	return result
}
