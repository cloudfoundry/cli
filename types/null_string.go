package types

import (
	"encoding/json"
	"fmt"
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
		return []byte(fmt.Sprintf(`"%s"`, n.Value)), nil
	}
	return []byte("null"), nil
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
