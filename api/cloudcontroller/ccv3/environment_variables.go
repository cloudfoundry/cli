package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

// EnvironmentVariables represents the environment variables on an application
type EnvironmentVariables struct {
	Variables map[string]types.FilteredString `json:"var"`
}

// GetApplicationEnvironmentVariables fetches the enviornment variables on
// an application.
func (client *Client) GetApplicationEnvironmentVariables(appGUID string) (EnvironmentVariables, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URIParams:   internal.Params{"app_guid": appGUID},
		RequestName: internal.GetApplicationEnvironmentVariables,
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

// PatchApplicationEnvironmentVariables updates the environment variables on
// an applicaiton. A restart is required for changes to take effect.
func (client *Client) PatchApplicationEnvironmentVariables(appGUID string, envVars EnvironmentVariables) (EnvironmentVariables, Warnings, error) {
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
