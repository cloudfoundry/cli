package jsonry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"code.cloudfoundry.org/jsonry/internal/path"

	"code.cloudfoundry.org/jsonry/internal/context"
	"code.cloudfoundry.org/jsonry/internal/tree"
)

// Unmarshal parses the specified JSON into the specified Go struct receiver.
// The receiver must be a pointer to a Go struct containing only fields of the type:
// string, bool, int*, uint*, float*, map, slice or struct. JSONry is recursive.
//
// If a field implements the json.Unmarshaler interface, then the UnmarshalJSON() method will be called.
func Unmarshal(data []byte, receiver interface{}) error {
	target := reflect.ValueOf(receiver)

	if target.Kind() != reflect.Ptr {
		return errors.New("receiver must be a pointer to a struct, got a non-pointer")
	}

	target = target.Elem()
	if target.Kind() != reflect.Struct {
		return fmt.Errorf("receiver must be a pointer to a struct type, got: %s", target.Type())
	}

	var source map[string]interface{}

	d := json.NewDecoder(bytes.NewBuffer(data))
	d.UseNumber()
	if err := d.Decode(&source); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	return unmarshalIntoStruct(context.Context{}, target, true, source)
}

func unmarshalIntoStruct(ctx context.Context, target reflect.Value, found bool, source interface{}) error {
	if !found || source == nil {
		return nil
	}

	src, ok := source.(map[string]interface{})
	if !ok {
		return newConversionError(ctx, source)
	}

	target = allocateIfNeeded(target)

	for i := 0; i < target.NumField(); i++ {
		field := target.Type().Field(i)

		if public(field) {
			p := path.ComputePath(field)
			s, found := tree.Tree(src).Fetch(p)
			if err := unmarshal(ctx.WithField(field.Name, field.Type), target.Field(i), found, s); err != nil {
				return err
			}
		}
	}

	return nil
}

func unmarshal(ctx context.Context, target reflect.Value, found bool, source interface{}) error {
	kind := underlyingType(target).Kind()

	var err error
	switch {
	case reflect.PtrTo(target.Type()).Implements(reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()):
		err = unmarshalIntoJSONUnmarshaler(ctx, target, found, source)
	case basicType(kind), kind == reflect.Interface:
		err = unmarshalInfoLeaf(ctx, target, found, source)
	case kind == reflect.Struct:
		err = unmarshalIntoStruct(ctx, target, found, source)
	case kind == reflect.Slice:
		err = unmarshalIntoSlice(ctx, target, found, source)
	case kind == reflect.Map:
		err = unmarshalIntoMap(ctx, target, found, source)
	default:
		err = newUnsupportedTypeError(ctx, target.Type())
	}
	return err
}

func unmarshalInfoLeaf(ctx context.Context, target reflect.Value, found bool, source interface{}) error {
	if !found {
		return nil
	}

	switch target.Kind() {
	case reflect.Ptr:
		switch source {
		case nil:
			return setZeroValue(target)
		default:
			return unmarshalInfoLeaf(ctx, allocateIfNeeded(target), found, source)
		}
	case reflect.String:
		switch s := source.(type) {
		case string:
			target.SetString(s)
			return nil
		case nil:
			return nil
		}
	case reflect.Bool:
		switch b := source.(type) {
		case bool:
			target.SetBool(b)
			return nil
		case nil:
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch n := source.(type) {
		case json.Number:
			if i, err := strconv.ParseInt(n.String(), 10, 64); err == nil {
				target.SetInt(i)
				return nil
			}
		case nil:
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch n := source.(type) {
		case json.Number:
			if i, err := strconv.ParseUint(n.String(), 10, 64); err == nil {
				target.SetUint(i)
				return nil
			}
		case nil:
			return nil
		}
	case reflect.Float32, reflect.Float64:
		switch n := source.(type) {
		case json.Number:
			if f, err := strconv.ParseFloat(n.String(), 64); err == nil {
				target.SetFloat(f)
				return nil
			}
		case nil:
			return nil
		}
	case reflect.Interface:
		switch source {
		case nil:
			return setZeroValue(target)
		default:
			target.Set(reflect.ValueOf(convertNumbers(source)))
		}
		return nil
	}

	return newConversionError(ctx, source)
}

func unmarshalIntoSlice(ctx context.Context, target reflect.Value, found bool, source interface{}) error {
	if !found || source == nil {
		return nil
	}

	src, ok := source.([]interface{})
	if !ok {
		return newConversionError(ctx, source)
	}

	slice := reflect.MakeSlice(underlyingType(target), len(src), len(src))
	allocateIfNeeded(target).Set(slice)

	for i := range src {
		elem := slice.Index(i)
		if err := unmarshal(ctx.WithIndex(i, elem.Type()), elem, true, src[i]); err != nil {
			return err
		}
	}

	return nil
}

func unmarshalIntoMap(ctx context.Context, target reflect.Value, found bool, source interface{}) error {
	targetType := underlyingType(target)

	if targetType.Key() != reflect.TypeOf("") {
		return newUnsupportedKeyTypeError(ctx, targetType.Key())
	}

	if !found || source == nil {
		return nil
	}

	src, ok := source.(map[string]interface{})
	if !ok {
		return newConversionError(ctx, source)
	}

	m := reflect.MakeMap(targetType)
	allocateIfNeeded(target).Set(m)

	for k, v := range src {
		targetValue := reflect.New(targetType.Elem()).Elem()
		if err := unmarshal(ctx.WithKey(k, targetValue.Type()), targetValue, true, v); err != nil {
			return err
		}

		m.SetMapIndex(reflect.ValueOf(k), targetValue)
	}

	return nil
}

func unmarshalIntoJSONUnmarshaler(ctx context.Context, target reflect.Value, found bool, source interface{}) error {
	if !found {
		return nil
	}

	json, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("error creating JSON for UnmarshalJSON(): %w", err)
	}

	elem := reflect.New(target.Type())
	s := elem.MethodByName("UnmarshalJSON").Call([]reflect.Value{reflect.ValueOf(json)})

	if err := checkForError(s[0]); err != nil {
		return fmt.Errorf("error from UnmarshalJSON() call at %s: %w", ctx, err)
	}

	target.Set(elem.Elem())
	return nil
}

func setZeroValue(target reflect.Value) error {
	target.Set(reflect.Zero(target.Type()))
	return nil
}

func allocateIfNeeded(target reflect.Value) reflect.Value {
	if target.Kind() != reflect.Ptr {
		return target
	}

	n := reflect.New(target.Type().Elem())
	target.Set(n)
	return n.Elem()
}

func convertNumbers(input interface{}) interface{} {
	n, ok := input.(json.Number)
	if !ok {
		return input
	}

	if i, err := n.Int64(); err == nil {
		return int(i)
	}

	if f, err := n.Float64(); err == nil {
		return f
	}

	return n.String()
}

func underlyingType(v reflect.Value) reflect.Type {
	if v.Kind() == reflect.Ptr {
		return v.Type().Elem()
	}
	return v.Type()
}
