package resources

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller"
	"code.cloudfoundry.org/cli/v8/types"
)

type EnvironmentVariables map[string]types.FilteredString

func (variables EnvironmentVariables) MarshalJSON() ([]byte, error) {
	ccEnvVars := struct {
		Var map[string]types.FilteredString `json:"var"`
	}{
		Var: variables,
	}

	return json.Marshal(ccEnvVars)
}

func (variables *EnvironmentVariables) UnmarshalJSON(data []byte) error {
	var ccEnvVars struct {
		Var map[string]types.FilteredInterface `json:"var"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccEnvVars)

	*variables = EnvironmentVariables{}

	for envVarName, envVarValue := range ccEnvVars.Var {
		var valueAsString string
		if str, ok := envVarValue.Value.(string); ok {
			valueAsString = str
		} else {
			bytes, err := json.Marshal(envVarValue.Value)
			if err != nil {
				return err
			}
			valueAsString = string(bytes)
		}

		(*variables)[envVarName] = types.FilteredString{Value: valueAsString, IsSet: true}
	}

	return err
}
