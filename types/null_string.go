package types

import (
	"encoding/json"
)

type NullString struct {
	Value string
	IsSet bool
}

func NewNullString(optionalValue ...string) NullString {
	switch len(optionalValue) {
	case 0:
		return NullString{IsSet: false}
	case 1:
		return NullString{Value: optionalValue[0], IsSet: true}
	default:
		panic("Too many strings passed to nullable string constructor")
	}
}

func (n NullString) MarshalJSON() ([]byte, error) {
	if n.IsSet {
		return json.Marshal(n.Value)
	}
	return json.Marshal(nil)
}

func (n *NullString) UnmarshalJSON(rawJSON []byte) error {
	var value *string
	err := json.Unmarshal(rawJSON, &value)
	if err != nil {
		return err
	}

	if value == nil {
		n.Value = ""
		n.IsSet = false
		return nil
	}

	n.Value = *value
	n.IsSet = true

	return nil
}
