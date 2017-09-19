package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// EnvironmentVariables represents the environment variables on an application
type EnvironmentVariables struct {
	Variables map[string]string `json:"var"`
}

// UpdateApplicationEnvironmentVariables updates the environment variables on
// an applicaiton. A restart is required for changes to take effect.
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
