package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

// EnvironmentVariables represents the environment variables that can be set on application by user
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
		Var map[string]types.FilteredString `json:"var"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccEnvVars)
	*variables = EnvironmentVariables(ccEnvVars.Var)

	return err
}

// UpdateApplicationEnvironmentVariables updates the user provided environment
// variables on an applicaiton. A restart is required for changes to take
// effect.
func (client *Client) UpdateApplicationEnvironmentVariables(appGUID string, envVars EnvironmentVariables) (EnvironmentVariables, Warnings, error) {
	bodyBytes, err := json.Marshal(envVars)
	if err != nil {
		return EnvironmentVariables{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		URIParams:   internal.Params{"app_guid": appGUID},
		RequestName: internal.PatchApplicationEnvironmentVariablesRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return EnvironmentVariables{}, nil, err
	}

	var responseEnvVars EnvironmentVariables
	response := cloudcontroller.Response{
		Result: &responseEnvVars,
	}
	err = client.connection.Make(request, &response)
	return responseEnvVars, response.Warnings, err
}
