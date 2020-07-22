package types

import (
	"encoding/json"
)

type OptionalBoolean struct {
	IsSet bool
	Value bool
}

func NewOptionalBoolean(value bool) OptionalBoolean {
	return OptionalBoolean{
		IsSet: true,
		Value: value,
	}
}

func (o *OptionalBoolean) UnmarshalJSON(rawJSON []byte) error {
	if err := json.Unmarshal(rawJSON, &o.Value); err != nil {
		return err
	}

	o.IsSet = true
	return nil
}

func (o OptionalBoolean) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Value)
}

func (o OptionalBoolean) OmitJSONry() bool {
	return !o.IsSet
}
