package helpers

import (
	"encoding/json"
)

func GetsDefaultEnvVarValue(stream []byte) string {
	envVariableJSON := struct {
		Var struct {
			SomeEnvVar string `json:"SOME_ENV_VAR"`
		} `json:"var"`
	}{}

	json.Unmarshal(stream, &envVariableJSON)

	return envVariableJSON.Var.SomeEnvVar
}
