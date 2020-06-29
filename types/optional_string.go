package types

import (
	"encoding/json"
)

type OptionalString struct {
	IsSet bool
	Value string
}

func NewOptionalString(v string) OptionalString {
	return OptionalString{
		IsSet: true,
		Value: v,
	}
}

func (o *OptionalString) UnmarshalJSON(rawJSON []byte) error {
	o.IsSet = true
	return json.Unmarshal(rawJSON, &o.Value)
}

func (o OptionalString) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Value)
}

func (o OptionalString) OmitJSONry() bool {
	return !o.IsSet
}

func (o OptionalString) String() string {
	return o.Value
}
