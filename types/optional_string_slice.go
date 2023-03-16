package types

import (
	"encoding/json"
	"strings"
)

type OptionalStringSlice struct {
	IsSet bool
	Value []string
}

func NewOptionalStringSlice(s ...string) OptionalStringSlice {
	return OptionalStringSlice{
		IsSet: true,
		Value: s,
	}
}

func (o *OptionalStringSlice) UnmarshalJSON(rawJSON []byte) error {
	var receiver []string
	if err := json.Unmarshal(rawJSON, &receiver); err != nil {
		return err
	}

	// This ensures that the empty state is always a nil slice, not an allocated (but empty) slice
	switch len(receiver) {
	case 0:
		o.Value = nil
	default:
		o.Value = receiver
	}

	o.IsSet = true
	return nil
}

func (o OptionalStringSlice) MarshalJSON() ([]byte, error) {
	if len(o.Value) > 0 {
		return json.Marshal(o.Value)
	}

	return []byte(`[]`), nil
}

func (o OptionalStringSlice) OmitJSONry() bool {
	return !o.IsSet
}

func (o OptionalStringSlice) String() string {
	return strings.Join(o.Value, ", ")
}
