package property

import (
	"reflect"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

// BuildMap creates a new string keyed map from an interface{}-keyed map, erroring if a key is not a string
func BuildMap(rawProperties map[interface{}]interface{}) (Map, error) {
	result := make(Map, len(rawProperties))

	for name, val := range rawProperties {
		nameStr, ok := name.(string)
		if !ok {
			return result, bosherr.Errorf("Map contains non-string key %#v", name)
		}

		convertedVal, err := Build(val)
		if err != nil {
			return result, err
		}

		result[nameStr] = convertedVal
	}

	return result, nil
}

// BuildList creates a new property List from an slice of interface{}, erroring if any elements are maps with non-string keys.
// Slices in the property List are converted to property Lists. Maps in the property List are converted to property Maps.
func BuildList(rawProperties []interface{}) (List, error) {
	result := make(List, len(rawProperties), len(rawProperties))

	for i, val := range rawProperties {
		convertedVal, err := Build(val)
		if err != nil {
			return result, err
		}

		result[i] = convertedVal
	}

	return result, nil
}

// Build creates a generic property that may be a Map, List or primitive.
// If it is a Map or List it will be built using the appropriate builder and constraints.
func Build(val interface{}) (Property, error) {
	if val == nil {
		return nil, nil
	}

	switch reflect.TypeOf(val).Kind() {
	case reflect.Map:
		valMap, ok := val.(map[interface{}]interface{})
		if !ok {
			return nil, bosherr.Errorf("Converting map %#v", val)
		}

		return BuildMap(valMap)

	case reflect.Slice:
		valSlice, ok := val.([]interface{})
		if !ok {
			return nil, bosherr.Errorf("Converting slice %#v", val)
		}

		return BuildList(valSlice)

	default:
		return val, nil
	}
}
