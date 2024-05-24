package helpers

import (
	"encoding/json"

	. "github.com/onsi/gomega"
)

func GetsDefaultEnvVarValue(stream []byte) string {
	envVariableJSON := struct {
		Var struct {
			SomeEnvVar string `json:"SOME_ENV_VAR"`
		} `json:"var"`
	}{}

	Expect(json.Unmarshal(stream, &envVariableJSON)).To(Succeed())

	return envVariableJSON.Var.SomeEnvVar
}
