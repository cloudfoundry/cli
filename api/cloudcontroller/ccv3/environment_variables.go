package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

// EnvironmentVariableGroups represents all environment variables on an application
type EnvironmentVariableGroups struct {
	SystemProvided      map[string]interface{} `json:"system_env_json"`
	ApplicationProvided map[string]interface{} `json:"application_env_json"`
	UserProvided        map[string]interface{} `json:"environment_variables"`
	RunningGroup        map[string]interface{} `json:"running_env_json"`
	StagingGroup        map[string]interface{} `json:"staging_env_json"`
}

// EnvironmentVariables represents the environment variables that can be set on application by user
type EnvironmentVariables struct {
	Variables map[string]types.FilteredString `json:"var"`
}

// GetApplicationEnvironmentVariables fetches all the environment variables on
// an application by groups.
func (client *Client) GetApplicationEnvironmentVariables(appGUID string) (EnvironmentVariableGroups, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URIParams:   internal.Params{"app_guid": appGUID},
		RequestName: internal.GetApplicationEnvironmentVariables,
	})
	if err != nil {
		return EnvironmentVariableGroups{}, nil, err
	}

	var responseEnvVars EnvironmentVariableGroups
	response := cloudcontroller.Response{
		Result: &responseEnvVars,
	}
	err = client.connection.Make(request, &response)
	return responseEnvVars, response.Warnings, err
}

// PatchApplicationUserProvidedEnvironmentVariables updates the user provided environment
// variables on an applicaiton. A restart is required for changes to take effect.
func (client *Client) PatchApplicationUserProvidedEnvironmentVariables(appGUID string, envVars EnvironmentVariables) (EnvironmentVariables, Warnings, error) {
	bodyBytes, err := json.Marshal(envVars)
	if err != nil {
		return EnvironmentVariables{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		URIParams:   internal.Params{"app_guid": appGUID},
		RequestName: internal.PatchApplicationUserProvidedEnvironmentVariablesRequest,
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
