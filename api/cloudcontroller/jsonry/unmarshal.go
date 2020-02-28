package jsonry

import (
	"bytes"
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/types"

	"encoding/json"
	"reflect"
	"strings"
)

var (
	jsonNumberType             = reflect.TypeOf(json.Number("0"))
	mapOfNullStringType        = reflect.TypeOf(map[string]types.NullString{})
	mapOfNullStringPointerType = reflect.PtrTo(mapOfNullStringType)
)

func Unmarshal(data []byte, store interface{}) error {
	storeValue, err := relectOnAndCheck(store)
	if err != nil {
		return err
	}

	var tree interface{}
	if err := unmarshalJSON(data, &tree); err != nil {
		return err
	}

	return unmarshal(storeValue, tree)
}

func computePath(field reflect.StructField) []string {
	if tag := field.Tag.Get("jsonry"); tag != "" {
		return strings.Split(tag, ".")
	}

	if tag := field.Tag.Get("json"); tag != "" {
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			return []string{parts[0]}
		}
	}

	return []string{strings.ToLower(field.Name)}
}

func fieldIsStruct(field reflect.Type) bool {
	kind := field.Kind()
	if kind == reflect.Ptr {
		kind = field.Elem().Kind()
	}

	return kind == reflect.Struct
}

func navigateAndFetch(path []string, tree interface{}) (interface{}, bool) {
	node, ok := tree.(map[string]interface{})
	if !ok {
		return nil, false
	}

	val, ok := node[path[0]]
	if !ok {
		return nil, false
	}

	if len(path) > 1 {
		return navigateAndFetch(path[1:], val)
	}

	return val, true
}

func relectOnAndCheck(store interface{}) (reflect.Value, error) {
	p := reflect.ValueOf(store)
	if kind := p.Kind(); kind != reflect.Ptr {
		return reflect.Value{}, errors.New("the storage object must be a pointer")
	}

	v := p.Elem()
	if kind := v.Kind(); kind != reflect.Struct {
		return reflect.Value{}, errors.New("the storage object pointer must point to a struct")
	}

	return v, nil
}

func setValue(fieldName string, store reflect.Value, value interface{}) error {
	// Special case: JSON numbers
	vv := reflect.ValueOf(value)
	if vv.Type() == jsonNumberType {
		if nv, ok := tryToConvertNumber(value.(json.Number)); ok {
			vv = nv
		}
	}

	// Special case: map[string]types.NullString
	if store.Type().AssignableTo(mapOfNullStringType) || store.Type().AssignableTo(mapOfNullStringPointerType) {
		if mv, ok := tryToConvertMapOfNullStrings(value); ok {
			vv = mv
		}
	}

	if store.Kind() == reflect.Ptr {
		n := reflect.New(vv.Type())
		if n.Type().AssignableTo(store.Type()) {
			n.Elem().Set(vv)
			store.Set(n)
			return nil
		}
		if n.Type().ConvertibleTo(store.Type()) {
			n.Elem().Set(vv)
			store.Set(n.Convert(store.Type()))
			return nil
		}
	} else {
		if vv.Type().AssignableTo(store.Type()) {
			store.Set(vv)
			return nil
		}
		if vv.Type().ConvertibleTo(store.Type()) {
			store.Set(vv.Convert(store.Type()))
			return nil
		}
	}

	return fmt.Errorf(
		"could not convert value '%v' type '%s' to '%s' for field '%s'",
		value,
		reflect.TypeOf(value),
		store.Type(),
		fieldName,
	)
}

func tryToConvertMapOfNullStrings(value interface{}) (reflect.Value, bool) {
	if source, ok := value.(map[string]interface{}); ok {
		destination := reflect.MakeMap(mapOfNullStringType)

		for k, v := range source {
			ns := types.NewNullString()
			if s, ok := v.(string); ok {
				ns = types.NewNullString(s)
			}
			destination.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(ns))
		}

		return destination, true
	}

	return reflect.Value{}, false
}

func tryToConvertNumber(num json.Number) (reflect.Value, bool) {
	// Extend to support other number types as needed
	if i64, err := num.Int64(); err == nil {
		return reflect.ValueOf(int(i64)), true
	}

	return reflect.Value{}, false
}

func unmarshal(storeValue reflect.Value, tree interface{}) error {
	if storeValue.Kind() == reflect.Ptr {
		n := reflect.New(storeValue.Type().Elem())
		storeValue.Set(n)
		storeValue = n.Elem()
	}

	storeType := storeValue.Type()
	for i := 0; i < storeType.NumField(); i++ {
		field := storeType.Field(i)
		path := computePath(field)
		if value, ok := navigateAndFetch(path, tree); ok {
			if fieldIsStruct(field.Type) {
				if err := unmarshal(storeValue.Field(i), value); err != nil {
					return err
				}
			} else {
				if err := setValue(field.Name, storeValue.Field(i), value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func unmarshalJSON(data []byte, store interface{}) error {
	d := json.NewDecoder(bytes.NewBuffer(data))
	d.UseNumber()
	return d.Decode(store)
}
