package types

import (
	"encoding/json"
)

type FilteredInterface struct {
	IsSet bool
	Value interface{}
}

func (n *FilteredInterface) UnmarshalJSON(rawJSON []byte) error {
	var value interface{}
	err := json.Unmarshal(rawJSON, &value)
	if err != nil {
		return err
	}

	n.Value = value
	n.IsSet = true
	return nil
}

func (n FilteredInterface) MarshalJSON() ([]byte, error) {
	if n.IsSet {
		return json.Marshal(n.Value)
	}

	return json.Marshal(new(json.RawMessage))
}
