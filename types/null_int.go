package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// NullInt is a wrapper around integer values that can be null or an integer.
// Use IsSet to check if the value is provided, instead of checking against 0.
type NullInt struct {
	IsSet bool
	Value int
}

// ParseStringValue is used to parse a user provided flag argument.
func (n *NullInt) ParseStringValue(val string) error {
	if val == "" {
		return nil
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return err
	}

	n.Value = intVal
	n.IsSet = true

	return nil
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

func (n *NullInt) UnmarshalJSON(rawJSON []byte) error {
	var value json.Number
	err := json.Unmarshal(rawJSON, &value)
	if err != nil {
		return err
	}

	if value.String() == "" {
		n.Value = 0
		n.IsSet = false
		return nil
	}

	valueInt, err := strconv.Atoi(value.String())
	if err != nil {
		return err
	}

	n.Value = valueInt
	n.IsSet = true

	return nil
}

func (n NullInt) MarshalJSON() ([]byte, error) {
	if n.IsSet {
		return []byte(fmt.Sprint(n.Value)), nil
	}
	return []byte("null"), nil
}
