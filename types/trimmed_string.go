package types

import (
	"encoding/json"
	"strings"
)

type TrimmedString struct {
	Value string
}

func NewTrimmedString(v string) TrimmedString {
	return TrimmedString{
		Value: strings.TrimSpace(v),
	}
}

func (o *TrimmedString) UnmarshalJSON(rawJSON []byte) error {
	var str string
	if err := json.Unmarshal(rawJSON, &str); err != nil {
		return err
	}
	o.Value = strings.TrimSpace(str)
	return nil
}

func (o TrimmedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Value)
}

func (o TrimmedString) String() string {
	return o.Value
}
