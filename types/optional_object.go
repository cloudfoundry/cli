package types

import "encoding/json"

type OptionalObject struct {
	IsSet bool
	Value map[string]interface{}
}

func NewOptionalObject(v map[string]interface{}) OptionalObject {
	if v == nil {
		// This ensures that when IsSet==true, we always have an empty map as the value which
		// marshals to `{}` and not a nil map which marshals to `null`
		v = make(map[string]interface{})
	}

	return OptionalObject{
		IsSet: true,
		Value: v,
	}
}

func (o *OptionalObject) UnmarshalJSON(rawJSON []byte) error {
	var receiver map[string]interface{}
	if err := json.Unmarshal(rawJSON, &receiver); err != nil {
		return err
	}

	*o = NewOptionalObject(receiver)
	return nil
}

func (o OptionalObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Value)
}

func (o OptionalObject) OmitJSONry() bool {
	return !o.IsSet
}
