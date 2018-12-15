package types

import (
	"encoding/json"
	"strings"
)

// FilteredString is a wrapper around string values that can be null/default or an
// actual value.  Use IsSet to check if the value is provided, instead of
// checking against the empty string.
type FilteredString struct {
	IsSet bool
	Value string
}

type FilteredStrings []FilteredString

func NewFilteredString(val string) *FilteredString {
	var result FilteredString

	result.ParseValue(val)

	return &result
}

// ParseValue is used to parse a user provided flag argument.
func (n *FilteredString) ParseValue(val string) {
	if val == "" {
		n.IsSet = false
		n.Value = ""
		return
	}

	n.IsSet = true

	switch val {
	case "null", "default":
		n.Value = ""
	default:
		n.Value = val
	}
}

func (n FilteredString) IsDefault() bool {
	return n.IsSet && n.Value == ""
}

func (n *FilteredString) UnmarshalJSON(rawJSON []byte) error {
	var value *string
	err := json.Unmarshal(rawJSON, &value)
	if err != nil {
		return err
	}

	if value != nil {
		n.Value = *value
		n.IsSet = true
		return nil
	}

	n.Value = ""
	n.IsSet = false
	return nil
}

// MarshalJSON marshals the value field if it's not empty, otherwise returns an
// null.
func (n FilteredString) MarshalJSON() ([]byte, error) {
	if n.Value != "" {
		return json.Marshal(n.Value)
	}

	return json.Marshal(new(json.RawMessage))
}

func (n FilteredString) String() string {
	if n.IsSet {
		return n.Value
	}

	return ""
}

func (n FilteredStrings) String() string {
	var ss []string

	for _, fs := range n {
		ss = append(ss, fs.Value)
	}

	return strings.Join(ss, ", ")
}
