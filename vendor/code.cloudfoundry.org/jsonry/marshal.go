package jsonry

import (
	"encoding/json"
	"fmt"
	"reflect"

	"code.cloudfoundry.org/jsonry/internal/tree"

	"code.cloudfoundry.org/jsonry/internal/context"
	"code.cloudfoundry.org/jsonry/internal/path"
)

// Marshal converts the specified Go struct into JSON. The input must be a struct or a pointer to a struct.
// Where a field is optional, the suffix ",omitempty" can be specified. This will mean that the field will
// be omitted from the JSON output if it is a nil pointer or has zero value for the type.
// When a field is a slice or an array, a single list hint "[]" may be specified in the JSONry path so that the array
// is created at the correct position in the JSON output.
//
// If a field implements the json.Marshaler interface, then the MarshalJSON() method will be called.
//
// The field type can be string, bool, int*, uint*, float*, map, slice, array or struct. JSONry is recursive.
func Marshal(in interface{}) ([]byte, error) {
	iv := reflect.Indirect(reflect.ValueOf(in))

	if iv.Kind() != reflect.Struct {
		return nil, fmt.Errorf(`the input must be a struct, not "%s"`, iv.Kind())
	}

	m, err := marshalStruct(context.Context{}, iv)
	if err != nil {
		return nil, err
	}

	return json.Marshal(m)
}

func marshalStruct(ctx context.Context, in reflect.Value) (map[string]interface{}, error) {
	out := make(tree.Tree)
	t := in.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if public(f) {
			p := path.ComputePath(f)
			if !p.OmitEmpty || !in.Field(i).IsZero() {
				r, err := marshal(ctx.WithField(f.Name, f.Type), in.Field(i))
				if err != nil {
					return nil, err
				}

				out.Attach(p, r)
			}
		}
	}

	return out, nil
}

func marshal(ctx context.Context, in reflect.Value) (r interface{}, err error) {
	input := reflect.Indirect(in)
	kind := input.Kind()

	switch {
	case kind == reflect.Invalid:
		r = nil
	case input.Type().Implements(reflect.TypeOf((*json.Marshaler)(nil)).Elem()):
		r, err = marshalJSONMarshaler(ctx, input)
	case kind == reflect.Interface:
		r, err = marshal(ctx, input.Elem())
	case basicType(kind):
		r = in.Interface()
	case kind == reflect.Struct:
		r, err = marshalStruct(ctx, input)
	case kind == reflect.Slice || kind == reflect.Array:
		r, err = marshalList(ctx, input)
	case kind == reflect.Map:
		r, err = marshalMap(ctx, input)
	default:
		err = newUnsupportedTypeError(ctx, input.Type())
	}

	return
}

func marshalList(ctx context.Context, in reflect.Value) ([]interface{}, error) {
	var out []interface{}

	for i := 0; i < in.Len(); i++ {
		ctx := ctx.WithIndex(i, in.Type())
		r, err := marshal(ctx, in.Index(i))
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}

	return out, nil
}

func marshalMap(ctx context.Context, in reflect.Value) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	iter := in.MapRange()
	for iter.Next() {
		k := iter.Key()
		if k.Kind() != reflect.String {
			return nil, newUnsupportedKeyTypeError(ctx, in.Type())
		}

		ctx := ctx.WithKey(k.String(), k.Type())

		r, err := marshal(ctx, iter.Value())
		if err != nil {
			return nil, err
		}
		out[k.String()] = r
	}

	return out, nil
}

func marshalJSONMarshaler(ctx context.Context, in reflect.Value) (interface{}, error) {
	t := in.MethodByName("MarshalJSON").Call(nil)

	if err := checkForError(t[1]); err != nil {
		return nil, fmt.Errorf("error from MarshaJSON() call at %s: %w", ctx, err)
	}

	var r interface{}
	err := json.Unmarshal(t[0].Bytes(), &r)
	if err != nil {
		return nil, fmt.Errorf(`error parsing MarshaJSON() output "%s" at %s: %w`, t[0].Bytes(), ctx, err)
	}

	return r, nil
}
