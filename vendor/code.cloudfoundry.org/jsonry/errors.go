package jsonry

import (
	"encoding/json"
	"fmt"
	"reflect"

	"code.cloudfoundry.org/jsonry/internal/context"
)

type unsupportedType struct {
	context context.Context
	typ     reflect.Type
}

func newUnsupportedTypeError(ctx context.Context, t reflect.Type) error {
	return &unsupportedType{
		context: ctx,
		typ:     t,
	}
}

func (u unsupportedType) Error() string {
	return fmt.Sprintf(`unsupported type "%s" at %s`, u.typ, u.context)
}

type unsupportedKeyType struct {
	context context.Context
	typ     reflect.Type
}

func newUnsupportedKeyTypeError(ctx context.Context, t reflect.Type) error {
	return &unsupportedKeyType{
		context: ctx,
		typ:     t,
	}
}

func (u unsupportedKeyType) Error() string {
	return fmt.Sprintf(`maps must only have string keys for "%s" at %s`, u.typ, u.context)
}

type conversionError struct {
	context context.Context
	value   interface{}
}

func newConversionError(ctx context.Context, value interface{}) error {
	return &conversionError{
		context: ctx,
		value:   value,
	}
}

func (c conversionError) Error() string {
	var t string
	switch c.value.(type) {
	case nil:
	case json.Number:
		t = "number"
	default:
		t = reflect.TypeOf(c.value).String()
	}

	msg := fmt.Sprintf(`cannot unmarshal "%+v" `, c.value)

	if t != "" {
		msg = fmt.Sprintf(`%stype "%s" `, msg, t)
	}

	return msg + "into " + c.context.String()
}
