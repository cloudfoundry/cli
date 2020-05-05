package jsonry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"code.cloudfoundry.org/cli/types"
)

var (
	jsonNumberType      = reflect.TypeOf(json.Number("0"))
	mapOfNullStringType = reflect.TypeOf(map[string]types.NullString{})
)

func Unmarshal(data []byte, store interface{}) error {
	storeValue, err := reflectOnAndCheckStructPointer(store)
	if err != nil {
		return err
	}

	var tree interface{}
	if err := unmarshalJSON(data, &tree); err != nil {
		return err
	}

	return unmarshal(storeValue, tree)
}

func reflectOnAndCheckStructPointer(store interface{}) (reflect.Value, error) {
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

func unmarshalJSON(data []byte, store interface{}) error {
	d := json.NewDecoder(bytes.NewBuffer(data))
	d.UseNumber()
	return d.Decode(store)
}

func unmarshal(storeValue reflect.Value, tree interface{}) error {
	storeValue = depointerify(storeValue)
	storeType := storeValue.Type()

	for i := 0; i < storeType.NumField(); i++ {
		field := storeType.Field(i)
		path := computePath(field)
		if value, ok := navigateAndFetch(path, tree); ok {
			if err := set(field, storeValue.Field(i), value); err != nil {
				return err
			}
		}
	}

	return nil
}

func depointerify(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr {
		return v
	}

	n := reflect.New(v.Type().Elem())
	v.Set(n)
	return n.Elem()
}

func navigateAndFetch(path jsonryPath, tree interface{}) (interface{}, bool) {
	branch := path.elements[0]

	node, ok := tree.(map[string]interface{})
	if !ok {
		return nil, false
	}

	val, ok := node[branch.name]
	if !ok {
		return nil, false
	}

	if path.len() > 1 {
		if vals, ok := val.([]interface{}); ok {
			return iterateAndFetch(path, vals), true
		}

		return navigateAndFetch(path.chop(), val)
	}

	return val, true
}

func iterateAndFetch(path jsonryPath, vals []interface{}) interface{} {
	var results []interface{}

	for _, v := range vals {
		r, _ := navigateAndFetch(path.chop(), v)
		results = append(results, r)
	}

	return results
}

func set(field reflect.StructField, fieldValue reflect.Value, value interface{}) error {
	switch actualKind(field.Type) {
	case reflect.Struct:
		return unmarshal(fieldValue, value)
	case reflect.Slice:
		return setSlice(field.Name, fieldValue, value)
	default:
		return setValue(field.Name, fieldValue, value)
	}
}

func setValue(fieldName string, store reflect.Value, value interface{}) error {
	vv := valueOfWithDenumberification(value)

	// Special case: map[string]types.NullString
	store = depointerify(store)
	if store.Type().AssignableTo(mapOfNullStringType) {
		if mv, ok := tryToConvertMapOfNullStrings(value); ok {
			vv = mv
		}
	}

	if vv.Type().AssignableTo(store.Type()) {
		store.Set(vv)
		return nil
	}
	if vv.Type().ConvertibleTo(store.Type()) {
		store.Set(vv.Convert(store.Type()))
		return nil
	}

	return fmt.Errorf(
		"could not convert value '%v' type '%s' to '%s' for field '%s'",
		value,
		reflect.TypeOf(value),
		store.Type(),
		fieldName,
	)
}

func setSlice(fieldName string, store reflect.Value, value interface{}) error {
	vs, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf(
			"could not convert value '%v' type '%s' to '%s' for field '%s' because it is not a list type",
			value,
			reflect.TypeOf(value),
			store.Type(),
			fieldName,
		)
	}

	store = depointerify(store)
	elemType := store.Type().Elem()
	arr := reflect.MakeSlice(reflect.SliceOf(elemType), len(vs), len(vs))
	for i, v := range vs {
		vv := valueOfWithDenumberification(v)
		if !vv.IsValid() {
			continue
		}

		if vv.Type().AssignableTo(elemType) {
			arr.Index(i).Set(vv)
			continue
		}

		if vv.Type().ConvertibleTo(elemType) {
			arr.Index(i).Set(vv.Convert(elemType))
			continue
		}

		if actualKind(elemType) == reflect.Struct && actualKind(vv.Type()) == reflect.Map {
			var structElement = reflect.New(elemType).Elem()
			err := unmarshal(structElement, v)
			if err == nil {
				arr.Index(i).Set(structElement)
				continue
			}
		}

		return fmt.Errorf(
			"could not convert value '%v' type '%s' to '%s' for field '%s' index %d",
			v,
			vv.Type(),
			elemType,
			fieldName,
			i,
		)
	}

	store.Set(arr.Slice(0, arr.Len()))
	return nil
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

func valueOfWithDenumberification(v interface{}) reflect.Value {
	vv := reflect.ValueOf(v)

	if vv.IsValid() && vv.Type() == jsonNumberType {
		num := json.Number(vv.String())
		// Extend to support other number types as needed
		if i64, err := num.Int64(); err == nil {
			return reflect.ValueOf(int(i64))
		}
		if f64, err := num.Float64(); err == nil {
			return reflect.ValueOf(float64(f64))
		}
	}

	return vv
}
