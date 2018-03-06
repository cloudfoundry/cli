package cloudcontroller

import (
	"bytes"
	"encoding/json"
)

// NewJSONDecoder is the JSON decoder that should be used for all Unmarshal
// operations.
func NewJSONDecoder(raw []byte) *json.Decoder {
	decoder := json.NewDecoder(bytes.NewBuffer(raw))
	decoder.UseNumber()
	return decoder
}
