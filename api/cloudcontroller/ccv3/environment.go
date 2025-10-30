package ccv3

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/internal"
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
	// Running is the set of default environment variables available to running
	// apps.
	Running map[string]interface{} `json:"running_env_json"`
	// Staging is the set of default environment variables available during
	// staging.
	Staging map[string]interface{} `json:"staging_env_json"`
	// System contains information about bound services for the application. AKA
	// VCAP_SERVICES.
	System map[string]interface{} `json:"system_env_json"`
}

// GetApplicationEnvironment fetches all the environment variables on
// an application by groups.
func (client *Client) GetApplicationEnvironment(appGUID string) (Environment, Warnings, error) {
	var responseBody Environment

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetApplicationEnvRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
