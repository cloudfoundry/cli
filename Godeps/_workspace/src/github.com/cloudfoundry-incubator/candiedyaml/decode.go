/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package candiedyaml

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

type Unmarshaler interface {
	UnmarshalYAML(tag string, value interface{}) error
}

// A Number represents a JSON number literal.
type Number string

// String returns the literal text of the number.
func (n Number) String() string { return string(n) }

// Float64 returns the number as a float64.
func (n Number) Float64() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

// Int64 returns the number as an int64.
func (n Number) Int64() (int64, error) {
	return strconv.ParseInt(string(n), 10, 64)
}

type Decoder struct {
	parser    yaml_parser_t
	event     yaml_event_t
	useNumber bool

	anchors map[string]interface{}
}

type ParserError struct {
	ErrorType   YAML_error_type_t
	Context     string
	ContextMark YAML_mark_t
	Problem     string
	ProblemMark YAML_mark_t
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("yaml: [%s] %s at line %d, column %d", e.Context, e.Problem, e.ProblemMark.line+1, e.ProblemMark.column+1)
}

type UnexpectedEventError struct {
	Value     string
	EventType yaml_event_type_t
	At        YAML_mark_t
}

func (e *UnexpectedEventError) Error() string {
	return fmt.Sprintf("yaml: Unexpect event [%d]: '%s' at line %d, column %d", e.EventType, e.Value, e.At.line+1, e.At.column+1)
}

func recovery(err *error) {
	if r := recover(); r != nil {
		if _, ok := r.(runtime.Error); ok {
			panic(r)
		}

		var tmpError error
		switch r := r.(type) {
		case error:
			tmpError = r
		case string:
			tmpError = errors.New(r)
		default:
			tmpError = errors.New("Unknown panic: " + reflect.TypeOf(r).String())
		}

		stackTrace := debug.Stack()
		*err = fmt.Errorf("%s\n%s", tmpError.Error(), string(stackTrace))
	}
}

func Unmarshal(data []byte, v interface{}) error {
	d := NewDecoder(bytes.NewBuffer(data))
	return d.Decode(v)
}

func NewDecoder(r io.Reader) *Decoder {
	d := &Decoder{
		anchors: make(map[string]interface{}),
	}
	yaml_parser_initialize(&d.parser)
	yaml_parser_set_input_reader(&d.parser, r)
	return d
}

func (d *Decoder) Decode(v interface{}) (err error) {
	defer recovery(&err)

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		rType := reflect.TypeOf(v)
		msg := "nil"
		if rType != nil {
			msg = rType.String()
		}
		return errors.New("Invalid type: " + msg)
	}

	if d.event.event_type == yaml_NO_EVENT {
		d.nextEvent()

		if d.event.event_type != yaml_STREAM_START_EVENT {
			return errors.New("Invalid stream")
		}

		d.nextEvent()
	}

	d.document(rv)
	return nil
}

func (d *Decoder) UseNumber() { d.useNumber = true }

func (d *Decoder) error(err error) {
	panic(err)
}

func (d *Decoder) nextEvent() {
	if d.event.event_type == yaml_STREAM_END_EVENT {
		d.error(errors.New("The stream is closed"))
	}

	if !yaml_parser_parse(&d.parser, &d.event) {
		yaml_event_delete(&d.event)

		d.error(&ParserError{
			ErrorType:   d.parser.error,
			Context:     d.parser.context,
			ContextMark: d.parser.context_mark,
			Problem:     d.parser.problem,
			ProblemMark: d.parser.problem_mark,
		})
	}
}

func (d *Decoder) document(rv reflect.Value) {
	if d.event.event_type != yaml_DOCUMENT_START_EVENT {
		d.error(fmt.Errorf("Expected document start - found %d", d.event.event_type))
	}

	d.nextEvent()
	d.parse(rv)

	if d.event.event_type != yaml_DOCUMENT_END_EVENT {
		d.error(fmt.Errorf("Expected document end - found %d", d.event.event_type))
	}

	d.nextEvent()
}

func (d *Decoder) parse(rv reflect.Value) {
	if !rv.IsValid() {
		// skip ahead since we cannot store
		d.valueInterface()
		return
	}

	anchor := string(d.event.anchor)
	switch d.event.event_type {
	case yaml_SEQUENCE_START_EVENT:
		d.sequence(rv)
		d.anchor(anchor, rv)
	case yaml_MAPPING_START_EVENT:
		d.mapping(rv)
		d.anchor(anchor, rv)
	case yaml_SCALAR_EVENT:
		d.scalar(rv)
		d.anchor(anchor, rv)
	case yaml_ALIAS_EVENT:
		d.alias(rv)
	case yaml_DOCUMENT_END_EVENT:
	default:
		d.error(&UnexpectedEventError{
			Value:     string(d.event.value),
			EventType: d.event.event_type,
			At:        d.event.start_mark,
		})
	}
}

func (d *Decoder) anchor(anchor string, rv reflect.Value) {
	if anchor != "" {
		d.anchors[anchor] = rv.Interface()
	}
}

func (d *Decoder) indirect(v reflect.Value) (Unmarshaler, reflect.Value) {
	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() {
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(Unmarshaler); ok {
				var temp interface{}
				return u, reflect.ValueOf(&temp)
			}
		}

		v = v.Elem()
	}

	return nil, v
}

func (d *Decoder) sequence(v reflect.Value) {
	if d.event.event_type != yaml_SEQUENCE_START_EVENT {
		d.error(fmt.Errorf("Expected sequence start - found %d", d.event.event_type))
	}

	u, pv := d.indirect(v)
	if u != nil {
		defer func() {
			if err := u.UnmarshalYAML("!!seq", pv.Interface()); err != nil {
				d.error(err)
			}
		}()
		_, pv = d.indirect(pv)
	}

	v = pv

	// Check type of target.
	switch v.Kind() {
	case reflect.Interface:
		if v.NumMethod() == 0 {
			// Decoding into nil interface?  Switch to non-reflect code.
			v.Set(reflect.ValueOf(d.sequenceInterface()))
			return
		}
		// Otherwise it's invalid.
		fallthrough
	default:
		d.error(errors.New("sequence: invalid type: " + v.Type().String()))
	case reflect.Array:
	case reflect.Slice:
		break
	}

	d.nextEvent()

	i := 0
	for {
		if d.event.event_type == yaml_SEQUENCE_END_EVENT {
			break
		}

		// Get element of array, growing if necessary.
		if v.Kind() == reflect.Slice {
			// Grow slice if necessary
			if i >= v.Cap() {
				newcap := v.Cap() + v.Cap()/2
				if newcap < 4 {
					newcap = 4
				}
				newv := reflect.MakeSlice(v.Type(), v.Len(), newcap)
				reflect.Copy(newv, v)
				v.Set(newv)
			}
			if i >= v.Len() {
				v.SetLen(i + 1)
			}
		}

		if i < v.Len() {
			// Decode into element.
			d.parse(v.Index(i))
		} else {
			// Ran out of fixed array: skip.
			d.parse(reflect.Value{})
		}
		i++
	}

	if i < v.Len() {
		if v.Kind() == reflect.Array {
			// Array.  Zero the rest.
			z := reflect.Zero(v.Type().Elem())
			for ; i < v.Len(); i++ {
				v.Index(i).Set(z)
			}
		} else {
			v.SetLen(i)
		}
	}
	if i == 0 && v.Kind() == reflect.Slice {
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}

	d.nextEvent()
}

func (d *Decoder) mapping(v reflect.Value) {
	u, pv := d.indirect(v)
	if u != nil {
		defer func() {
			if err := u.UnmarshalYAML("!!map", pv.Interface()); err != nil {
				d.error(err)
			}
		}()
		_, pv = d.indirect(pv)
	}
	v = pv

	// Decoding into nil interface?  Switch to non-reflect code.
	if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
		v.Set(reflect.ValueOf(d.mappingInterface()))
		return
	}

	// Check type of target: struct or map[X]Y
	switch v.Kind() {
	case reflect.Struct:
		d.mappingStruct(v)
		return
	case reflect.Map:
	default:
		d.error(errors.New("mapping: invalid type: " + v.Type().String()))
	}

	mapt := v.Type()
	if v.IsNil() {
		v.Set(reflect.MakeMap(mapt))
	}

	d.nextEvent()

	keyt := mapt.Key()
	mapElemt := mapt.Elem()

	var mapElem reflect.Value
	for {
		if d.event.event_type == yaml_MAPPING_END_EVENT {
			break
		}
		key := reflect.New(keyt)
		d.parse(key.Elem())

		if !mapElem.IsValid() {
			mapElem = reflect.New(mapElemt).Elem()
		} else {
			mapElem.Set(reflect.Zero(mapElemt))
		}

		d.parse(mapElem)

		v.SetMapIndex(key.Elem(), mapElem)
	}

	d.nextEvent()
}

func (d *Decoder) mappingStruct(v reflect.Value) {

	structt := v.Type()
	fields := cachedTypeFields(structt)

	d.nextEvent()

	for {
		if d.event.event_type == yaml_MAPPING_END_EVENT {
			break
		}
		key := ""
		d.parse(reflect.ValueOf(&key))

		// Figure out field corresponding to key.
		var subv reflect.Value

		var f *field
		for i := range fields {
			ff := &fields[i]
			if ff.name == key {
				f = ff
				break
			}

			if f == nil && strings.EqualFold(ff.name, key) {
				f = ff
			}
		}

		if f != nil {
			subv = v
			for _, i := range f.index {
				if subv.Kind() == reflect.Ptr {
					if subv.IsNil() {
						subv.Set(reflect.New(subv.Type().Elem()))
					}
					subv = subv.Elem()
				}
				subv = subv.Field(i)
			}
		}
		d.parse(subv)
	}

	d.nextEvent()
}

func (d *Decoder) scalar(v reflect.Value) {
	u, pv := d.indirect(v)

	var tag string
	if u != nil {
		defer func() {
			if err := u.UnmarshalYAML(tag, pv.Interface()); err != nil {
				d.error(err)
			}
		}()

		_, pv = d.indirect(pv)
	}
	v = pv

	var err error
	tag, err = resolve(d.event, v, d.useNumber)
	if err != nil {
		d.error(err)
	}

	d.nextEvent()
}

func (d *Decoder) alias(rv reflect.Value) {
	if val, ok := d.anchors[string(d.event.anchor)]; ok {
		rv.Set(reflect.ValueOf(val))
	}

	d.nextEvent()
}

func (d *Decoder) valueInterface() interface{} {
	var v interface{}

	anchor := string(d.event.anchor)
	switch d.event.event_type {
	case yaml_SEQUENCE_START_EVENT:
		v = d.sequenceInterface()
	case yaml_MAPPING_START_EVENT:
		v = d.mappingInterface()
	case yaml_SCALAR_EVENT:
		v = d.scalarInterface()
	case yaml_ALIAS_EVENT:
		return d.aliasInterface()
	case yaml_DOCUMENT_END_EVENT:
		d.error(&UnexpectedEventError{
			Value:     string(d.event.value),
			EventType: d.event.event_type,
			At:        d.event.start_mark,
		})

	}

	d.anchorInterface(anchor, v)
	return v
}

func (d *Decoder) scalarInterface() interface{} {
	_, v := resolveInterface(d.event, d.useNumber)

	d.nextEvent()
	return v
}

func (d *Decoder) anchorInterface(anchor string, i interface{}) {
	if anchor != "" {
		d.anchors[anchor] = i
	}
}

func (d *Decoder) aliasInterface() interface{} {
	v := d.anchors[string(d.event.anchor)]

	d.nextEvent()
	return v
}

// arrayInterface is like array but returns []interface{}.
func (d *Decoder) sequenceInterface() []interface{} {
	var v = make([]interface{}, 0)

	d.nextEvent()
	for {
		if d.event.event_type == yaml_SEQUENCE_END_EVENT {
			break
		}

		v = append(v, d.valueInterface())
	}

	d.nextEvent()
	return v
}

// objectInterface is like object but returns map[string]interface{}.
func (d *Decoder) mappingInterface() map[interface{}]interface{} {
	m := make(map[interface{}]interface{})

	d.nextEvent()

	for {
		if d.event.event_type == yaml_MAPPING_END_EVENT {
			break
		}

		key := d.valueInterface()

		// Read value.
		m[key] = d.valueInterface()
	}

	d.nextEvent()
	return m
}
