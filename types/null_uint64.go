package types

import (
	"encoding/json"
	"strconv"
)

// NullUint64 is a wrapper around uint64 values that can be null or an unint64.
// Use IsSet to check if the value is provided, instead of checking against 0.
type NullUint64 struct {
	IsSet bool
	Value uint64
}

// ParseStringValue is used to parse a user provided flag argument.
func (n *NullUint64) ParseStringValue(val string) error {
	if val == "" {
		return nil
	}

	uint64Val, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return err
	}

	n.Value = uint64Val
	n.IsSet = true

	return nil
}

func (n *NullUint64) UnmarshalJSON(rawJSON []byte) error {
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

	valueInt, err := strconv.ParseUint(value.String(), 10, 64)
	if err != nil {
		return err
	}

	n.Value = valueInt
	n.IsSet = true

	return nil
}
