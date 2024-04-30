package types

import (
	"fmt"
	"strconv"

	"github.com/jessevdk/go-flags"
)

// NullFloat64 is a wrapper around float64 values that can be null or a
// float64. Use IsSet to check if the value is provided, instead of checking
// against 0.
type NullFloat64 struct {
	IsSet bool
	Value float64
}

// ParseStringValue is used to parse a user provided flag argument.
func (n *NullFloat64) ParseStringValue(val string) error {
	if val == "" {
		n.IsSet = false
		n.Value = 0
		return nil
	}

	// float64Val, err := strconv.Atoi(val)
	float64Val, err := strconv.ParseFloat(val, 64)
	if err != nil {
		n.IsSet = false
		n.Value = 0
		return &flags.Error{
			Type:    flags.ErrMarshal,
			Message: fmt.Sprintf("invalid float64 value `%s`", val),
		}
	}

	n.Value = float64Val
	n.IsSet = true

	return nil
}

// IsValidValue returns an error if the input value is not a float64.
func (n *NullFloat64) IsValidValue(val string) error {
	return n.ParseStringValue(val)
}

// ParseFloat64Value is used to parse a user provided *float64 argument.
func (n *NullFloat64) ParseFloat64Value(val *float64) {
	if val == nil {
		n.IsSet = false
		n.Value = 0
		return
	}

	n.Value = *val
	n.IsSet = true
}

func (n *NullFloat64) UnmarshalFlag(val string) error {
	return n.ParseStringValue(val)
}

func (n *NullFloat64) UnmarshalJSON(rawJSON []byte) error {
	stringValue := string(rawJSON)

	if stringValue == JsonNull {
		n.Value = 0
		n.IsSet = false
		return nil
	}

	return n.ParseStringValue(stringValue)
}

func (n NullFloat64) MarshalJSON() ([]byte, error) {
	if n.IsSet {
		return []byte(fmt.Sprint(n.Value)), nil
	}
	return []byte(JsonNull), nil
}
