package types

import (
	"fmt"
	"strconv"

	"github.com/jessevdk/go-flags"
)

const JsonNull = "null"

// NullInt is a wrapper around integer values that can be null or an integer.
// Use IsSet to check if the value is provided, instead of checking against 0.
type NullInt struct {
	IsSet bool
	Value int
}

// ParseStringValue is used to parse a user provided flag argument.
func (n *NullInt) ParseStringValue(val string) error {
	if val == "" {
		n.IsSet = false
		n.Value = 0
		return nil
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		n.IsSet = false
		n.Value = 0
		return &flags.Error{
			Type:    flags.ErrMarshal,
			Message: "invalid integer value",
		}
	}

	n.Value = intVal
	n.IsSet = true

	return nil
}

// IsValidValue returns an error if the input value is not an integer.
func (n *NullInt) IsValidValue(val string) error {
	_, err := strconv.Atoi(val)
	return err
}

// ParseIntValue is used to parse a user provided *int argument.
func (n *NullInt) ParseIntValue(val *int) {
	if val == nil {
		n.IsSet = false
		n.Value = 0
		return
	}

	n.Value = *val
	n.IsSet = true
}

func (n *NullInt) UnmarshalFlag(val string) error {
	return n.ParseStringValue(val)
}

func (n *NullInt) UnmarshalJSON(rawJSON []byte) error {
	stringValue := string(rawJSON)

	if stringValue == JsonNull {
		n.Value = 0
		n.IsSet = false
		return nil
	}

	return n.ParseStringValue(stringValue)
}

func (n NullInt) MarshalJSON() ([]byte, error) {
	if n.IsSet {
		return []byte(fmt.Sprint(n.Value)), nil
	}
	return []byte(JsonNull), nil
}
