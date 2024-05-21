package helpers

import (
	"encoding/json"

	. "github.com/onsi/gomega"
)

func GetsEnablementValue(stream []byte) bool {
	enablementResponse := struct {
		Enabled bool `json:"enabled"`
	}{}

	Expect(json.Unmarshal(stream, &enablementResponse)).To(Succeed())

	return enablementResponse.Enabled
}
