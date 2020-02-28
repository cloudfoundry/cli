package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// NullBool is a wrapper around bool values that can be null or an bool.
// Use IsSet to check if the value is provided, instead of checking against false.
type NullBool struct {
	IsSet bool
	Value bool
}

// ParseStringValue is used to parse a user provided flag argument.
func (n *NullBool) ParseStringValue(val string) error {
	if val == "" {
		return nil
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}

	n.Value = boolVal
	n.IsSet = true

	return nil
}

// ParseBoolValue is used to parse a user provided *bool argument.
func (n *NullBool) ParseBoolValue(val *bool) {
	if val == nil {
		n.IsSet = false
		n.Value = false
		return
	}

	n.Value = *val
	n.IsSet = true
}

func (n *NullBool) UnmarshalJSON(rawJSON []byte) error {
	var value *bool
	err := json.Unmarshal(rawJSON, &value)
	if err != nil {
		return err
	}

	if value == nil {
		n.Value = false
		n.IsSet = false
		return nil
	}

	n.Value = *value
	n.IsSet = true

	return nil
}

func (n NullBool) MarshalJSON() ([]byte, error) {
	if n.IsSet {
		return []byte(fmt.Sprint(n.Value)), nil
	}
	return []byte(JsonNull), nil
}
