package helpers

import (
	"encoding/json"
)

func GetsEnablementValue(stream []byte) bool {
	enablementResponse := struct {
		Enabled bool `json:"enabled"`
	}{}

	json.Unmarshal(stream, &enablementResponse)

	return enablementResponse.Enabled
}
