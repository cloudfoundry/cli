package cloudcontroller

import (
	"bytes"
	"encoding/json"
)

// DecodeJSON unmarshals JSON into the given object with the appropriate
// settings.
func DecodeJSON(raw []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewBuffer(raw))
	decoder.UseNumber()
	return decoder.Decode(v)
}
