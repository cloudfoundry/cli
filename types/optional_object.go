package types

import "encoding/json"

// OptionalObject is for situations where we want to differentiate between an
// empty object, and the object not having been set. An example would be an
// optional command line option where we want to tell the difference between
// it being set to an empty object, and it not being specified at all.
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
