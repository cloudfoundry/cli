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
	"encoding/base64"
	"errors"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var byteSliceType = reflect.TypeOf([]byte(nil))

var bool_values map[string]bool
var null_values map[string]bool

var signs = []byte{'-', '+'}
var nulls = []byte{'~', 'n', 'N'}
var bools = []byte{'t', 'T', 'f', 'F', 'y', 'Y', 'n', 'N', 'o', 'O'}

var timestamp_regexp *regexp.Regexp
var ymd_regexp *regexp.Regexp

func init() {
	bool_values = make(map[string]bool)
	bool_values["y"] = true
	bool_values["yes"] = true
	bool_values["n"] = false
	bool_values["no"] = false
	bool_values["true"] = true
	bool_values["false"] = false
	bool_values["on"] = true
	bool_values["off"] = false

	null_values = make(map[string]bool)
	null_values["~"] = true
	null_values["null"] = true
	null_values["Null"] = true
	null_values["NULL"] = true
	null_values[""] = true

	timestamp_regexp = regexp.MustCompile("^([0-9][0-9][0-9][0-9])-([0-9][0-9]?)-([0-9][0-9]?)(?:(?:[Tt]|[ \t]+)([0-9][0-9]?):([0-9][0-9]):([0-9][0-9])(?:\\.([0-9]*))?(?:[ \t]*(?:Z|([-+][0-9][0-9]?)(?::([0-9][0-9])?)?))?)?$")
	ymd_regexp = regexp.MustCompile("^([0-9][0-9][0-9][0-9])-([0-9][0-9]?)-([0-9][0-9]?)$")
}

func resolve(event yaml_event_t, v reflect.Value, useNumber bool) (string, error) {
	val := string(event.value)

	if null_values[val] {
		v.Set(reflect.Zero(v.Type()))
		return "!!null", nil
	}

	switch v.Kind() {
	case reflect.String:
		if useNumber && v.Type() == numberType {
			tag, i := resolveInterface(event, useNumber)
			if n, ok := i.(Number); ok {
				v.Set(reflect.ValueOf(n))
				return tag, nil
			} else {
				return "", errors.New("Not a Number: " + reflect.TypeOf(i).String())
			}
		} else {
			v.SetString(val)
		}
	case reflect.Bool:
		return resolve_bool(val, v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return resolve_int(val, v, useNumber)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return resolve_uint(val, v, useNumber)
	case reflect.Float32, reflect.Float64:
		return resolve_float(val, v, useNumber)
	case reflect.Interface:
		_, i := resolveInterface(event, useNumber)
		v.Set(reflect.ValueOf(i))
	case reflect.Struct:
		return resolve_time(val, v)
	case reflect.Slice:
		if v.Type() != byteSliceType {
			return "", errors.New("Cannot resolve into " + v.Type().String())
		}
		b := make([]byte, base64.StdEncoding.DecodedLen(len(event.value)))
		n, err := base64.StdEncoding.Decode(b, event.value)
		if err != nil {
			return "", err
		}

		v.Set(reflect.ValueOf(b[0:n]))
	default:
		return "", errors.New("Resolve failed for " + v.Kind().String())
	}

	return "!!str", nil
}

func resolve_bool(val string, v reflect.Value) (string, error) {
	b, found := bool_values[strings.ToLower(val)]
	if !found {
		return "", errors.New("Invalid boolean: " + val)
	}

	v.SetBool(b)
	return "!!bool", nil
}

func resolve_int(val string, v reflect.Value, useNumber bool) (string, error) {
	original := val
	val = strings.Replace(val, "_", "", -1)
	var value int64

	isNumberValue := v.Type() == numberType

	sign := int64(1)
	if val[0] == '-' {
		sign = -1
		val = val[1:]
	} else if val[0] == '+' {
		val = val[1:]
	}

	base := 10
	if val == "0" {
		if isNumberValue {
			v.SetString("0")
		} else {
			v.Set(reflect.Zero(v.Type()))
		}

		return "!!int", nil
	}

	if strings.HasPrefix(val, "0b") {
		base = 2
		val = val[2:]
	} else if strings.HasPrefix(val, "0x") {
		base = 16
		val = val[2:]
	} else if val[0] == '0' {
		base = 8
		val = val[1:]
	} else if strings.Contains(val, ":") {
		digits := strings.Split(val, ":")
		bes := int64(1)
		for j := len(digits) - 1; j >= 0; j-- {
			n, err := strconv.ParseInt(digits[j], 10, 64)
			n *= bes
			if err != nil {
				return "", errors.New("Integer: " + original)
			}
			value += n
			bes *= 60
		}

		value *= sign

		if isNumberValue {
			v.SetString(strconv.FormatInt(value, 10))
		} else {
			if v.OverflowInt(value) {
				return "", errors.New("Integer: " + original)
			}

			v.SetInt(value)
		}
		return "!!int", nil
	}

	value, err := strconv.ParseInt(val, base, 64)
	if err != nil {
		return "", errors.New("Integer: " + original)
	}
	value *= sign

	if isNumberValue {
		v.SetString(strconv.FormatInt(value, 10))
	} else {
		if v.OverflowInt(value) {
			return "", errors.New("Integer: " + original)
		}

		v.SetInt(value)
	}

	return "!!int", nil
}

func resolve_uint(val string, v reflect.Value, useNumber bool) (string, error) {
	original := val
	val = strings.Replace(val, "_", "", -1)
	var value uint64

	isNumberValue := v.Type() == numberType

	if val[0] == '-' {
		return "", errors.New("Unsigned int with negative value: " + original)
	}

	if val[0] == '+' {
		val = val[1:]
	}

	base := 10
	if val == "0" {
		if isNumberValue {
			v.SetString("0")
		} else {
			v.Set(reflect.Zero(v.Type()))
		}

		return "!!int", nil
	}

	if strings.HasPrefix(val, "0b") {
		base = 2
		val = val[2:]
	} else if strings.HasPrefix(val, "0x") {
		base = 16
		val = val[2:]
	} else if val[0] == '0' {
		base = 8
		val = val[1:]
	} else if strings.Contains(val, ":") {
		digits := strings.Split(val, ":")
		bes := uint64(1)
		for j := len(digits) - 1; j >= 0; j-- {
			n, err := strconv.ParseUint(digits[j], 10, 64)
			n *= bes
			if err != nil {
				return "", errors.New("Unsigned Integer: " + original)
			}
			value += n
			bes *= 60
		}

		if isNumberValue {
			v.SetString(strconv.FormatUint(value, 10))
		} else {
			if v.OverflowUint(value) {
				return "", errors.New("Unsigned Integer: " + original)
			}

			v.SetUint(value)
		}
		return "!!int", nil
	}

	value, err := strconv.ParseUint(val, base, 64)
	if err != nil {
		return "", errors.New("Unsigned Integer: " + val)
	}

	if isNumberValue {
		v.SetString(strconv.FormatUint(value, 10))
	} else {
		if v.OverflowUint(value) {
			return "", errors.New("Unsigned Integer: " + val)
		}

		v.SetUint(value)
	}

	return "!!int", nil
}

func resolve_float(val string, v reflect.Value, useNumber bool) (string, error) {
	val = strings.Replace(val, "_", "", -1)
	var value float64

	isNumberValue := v.Type() == numberType
	typeBits := 64
	if !isNumberValue {
		typeBits = v.Type().Bits()
	}

	sign := 1
	if val[0] == '-' {
		sign = -1
		val = val[1:]
	} else if val[0] == '+' {
		val = val[1:]
	}

	valLower := strings.ToLower(val)
	if valLower == ".inf" {
		value = math.Inf(sign)
	} else if valLower == ".nan" {
		value = math.NaN()
	} else if strings.Contains(val, ":") {
		digits := strings.Split(val, ":")
		bes := float64(1)
		for j := len(digits) - 1; j >= 0; j-- {
			n, err := strconv.ParseFloat(digits[j], typeBits)
			n *= bes
			if err != nil {
				return "", errors.New("Float: " + val)
			}
			value += n
			bes *= 60
		}
		value *= float64(sign)
	} else {
		var err error
		value, err = strconv.ParseFloat(val, typeBits)
		value *= float64(sign)

		if err != nil {
			return "", errors.New("Float: " + val)
		}
	}

	if isNumberValue {
		v.SetString(strconv.FormatFloat(value, 'g', -1, typeBits))
	} else {
		if v.OverflowFloat(value) {
			return "", errors.New("Float: " + val)
		}

		v.SetFloat(value)
	}

	return "!!float", nil
}

func resolve_time(val string, v reflect.Value) (string, error) {
	var parsedTime time.Time
	matches := ymd_regexp.FindStringSubmatch(val)
	if len(matches) > 0 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		parsedTime = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	} else {
		matches = timestamp_regexp.FindStringSubmatch(val)
		if len(matches) == 0 {
			return "", errors.New("Unexpected timestamp: " + val)
		}

		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		hour, _ := strconv.Atoi(matches[4])
		min, _ := strconv.Atoi(matches[5])
		sec, _ := strconv.Atoi(matches[6])

		nsec := 0
		if matches[7] != "" {
			millis, _ := strconv.Atoi(matches[7])
			nsec = int(time.Duration(millis) * time.Millisecond)
		}

		loc := time.UTC
		if matches[8] != "" {
			sign := matches[8][0]
			hr, _ := strconv.Atoi(matches[8][1:])
			min := 0
			if matches[9] != "" {
				min, _ = strconv.Atoi(matches[9])
			}

			zoneOffset := (hr*60 + min) * 60
			if sign == '-' {
				zoneOffset = -zoneOffset
			}

			loc = time.FixedZone("", zoneOffset)
		}
		parsedTime = time.Date(year, time.Month(month), day, hour, min, sec, nsec, loc)
	}

	v.Set(reflect.ValueOf(parsedTime))
	return "", nil
}

func resolveInterface(event yaml_event_t, useNumber bool) (string, interface{}) {
	if len(event.value) == 0 {
		return "", nil
	}

	val := string(event.value)
	if len(event.tag) == 0 && !event.implicit {
		return "", val
	}

	var result interface{}

	sign := false
	c := val[0]
	switch {
	case bytes.IndexByte(signs, c) != -1:
		sign = true
		fallthrough
	case c >= '0' && c <= '9':
		i := int64(0)
		result = &i
		if useNumber {
			var n Number
			result = &n
		}

		v := reflect.ValueOf(result).Elem()
		if _, err := resolve_int(val, v, useNumber); err == nil {
			return "!!int", v.Interface()
		}

		f := float64(0)
		result = &f
		if useNumber {
			var n Number
			result = &n
		}

		v = reflect.ValueOf(result).Elem()
		if _, err := resolve_float(val, v, useNumber); err == nil {
			return "!!float", v.Interface()
		}

		if !sign {
			t := time.Time{}
			if _, err := resolve_time(val, reflect.ValueOf(&t).Elem()); err == nil {
				return "", t
			}
		}
	case bytes.IndexByte(nulls, c) != -1:
		if null_values[val] {
			return "!!null", nil
		}
		b := false
		if _, err := resolve_bool(val, reflect.ValueOf(&b).Elem()); err == nil {
			return "!!bool", b
		}
	case c == '.':
		f := float64(0)
		result = &f
		if useNumber {
			var n Number
			result = &n
		}

		v := reflect.ValueOf(result).Elem()
		if _, err := resolve_float(val, v, useNumber); err == nil {
			return "!!float", v.Interface()
		}
	case bytes.IndexByte(bools, c) != -1:
		b := false
		if _, err := resolve_bool(val, reflect.ValueOf(&b).Elem()); err == nil {
			return "!!bool", b
		}
	}

	return "!!str", string(event.value)
}
