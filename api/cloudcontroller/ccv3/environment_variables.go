package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

// EnvironmentVariables represents the environment variables that can be set on
// an application by the user.
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

// GetEnvironmentVariableGroup gets the values of a particular environment variable group.
func (client *Client) GetEnvironmentVariableGroup(group constant.EnvironmentVariableGroupName) (EnvironmentVariables, Warnings, error) {
	var responseBody EnvironmentVariables

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetEnvironmentVariableGroupRequest,
		URIParams:    internal.Params{"group_name": string(group)},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UpdateApplicationEnvironmentVariables adds/updates the user provided
// environment variables on an application. A restart is required for changes
// to take effect.
func (client *Client) UpdateApplicationEnvironmentVariables(appGUID string, envVars EnvironmentVariables) (EnvironmentVariables, Warnings, error) {
	var responseBody EnvironmentVariables

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchApplicationEnvironmentVariablesRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		RequestBody:  envVars,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) UpdateEnvironmentVariableGroup(group constant.EnvironmentVariableGroupName, envVars EnvironmentVariables) (EnvironmentVariables, Warnings, error) {
	var responseBody EnvironmentVariables

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchEnvironmentVariableGroupRequest,
		URIParams:    internal.Params{"group_name": string(group)},
		RequestBody:  envVars,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
