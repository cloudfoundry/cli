package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

// Environment variables that will be provided to an app at runtime. It will
// include environment variables for Environment Variable Groups and Service
// Bindings.
type Environment struct {
	// Application contains basic application settings set by the user and CF
	// instance.
	Application map[string]interface{} `json:"application_env_json"`
	// EnvironmentVariables are user provided environment variables.
	EnvironmentVariables map[string]interface{} `json:"environment_variables"`
	//Running is the set of default environment variables available to running
	//apps.
	Running map[string]interface{} `json:"running_env_json"`
	//Staging is the set of default environment variables available during
	//staging.
	Staging map[string]interface{} `json:"staging_env_json"`
	// System contains information about bound services for the application. AKA
	// VCAP_SERVICES.
	System map[string]interface{} `json:"system_env_json"`
}

// EnvironmentVariables represents the environment variables that can be set on application by user
type EnvironmentVariables struct {
	Variables map[string]types.FilteredString `json:"var"`
}

// GetApplicationEnvironment fetches all the environment variables on
// an application by groups.
func (client *Client) GetApplicationEnvironment(appGUID string) (Environment, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URIParams:   internal.Params{"app_guid": appGUID},
		RequestName: internal.GetApplicationEnvRequest,
	})
	if err != nil {
		return Environment{}, nil, err
	}

	var responseEnvVars Environment
	response := cloudcontroller.Response{
		Result: &responseEnvVars,
	}
	err = client.connection.Make(request, &response)
	return responseEnvVars, response.Warnings, err
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
