package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	ccv3internal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
	"code.cloudfoundry.org/cli/resources"
)

// EnvironmentVariables represents the environment variables that can be set on
// an application by the user.

// GetEnvironmentVariableGroup gets the values of a particular environment variable group.
func (client *Client) GetEnvironmentVariableGroup(group constant.EnvironmentVariableGroupName) (resources.EnvironmentVariables, Warnings, error) {
	var responseBody resources.EnvironmentVariables

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.GetEnvironmentVariableGroupRequest,
		URIParams:    internal.Params{"group_name": string(group)},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UpdateApplicationEnvironmentVariables adds/updates the user provided
// environment variables on an application. A restart is required for changes
// to take effect.
func (client *Client) UpdateApplicationEnvironmentVariables(appGUID string, envVars resources.EnvironmentVariables) (resources.EnvironmentVariables, Warnings, error) {
	var responseBody resources.EnvironmentVariables

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.PatchApplicationEnvironmentVariablesRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		RequestBody:  envVars,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) UpdateEnvironmentVariableGroup(group constant.EnvironmentVariableGroupName, envVars resources.EnvironmentVariables) (resources.EnvironmentVariables, Warnings, error) {
	var responseBody resources.EnvironmentVariables

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.PatchEnvironmentVariableGroupRequest,
		URIParams:    internal.Params{"group_name": string(group)},
		RequestBody:  envVars,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
