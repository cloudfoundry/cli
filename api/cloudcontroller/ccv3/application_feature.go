package ccv3

import (
	"bytes"
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type ApplicationFeature struct {
	// Name of the application feature
	Name    string
	Enabled bool
	//Reason  string `json:omitempty`
}

type SSHEnabled struct {
	Enabled bool
	Reason  string
}

func (client *Client) GetAppFeature(appGUID string, featureName string) (ApplicationFeature, Warnings, error) {
	var responseBody ApplicationFeature

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.GetApplicationFeaturesRequest,
		URIParams:    internal.Params{"app_guid": appGUID, "name": featureName},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetSSHEnabled(appGUID string) (SSHEnabled, Warnings, error) {
	var responseBody SSHEnabled

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.GetSSHEnabled,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UpdateAppFeature enables/disables the ability to ssh for a given application.
func (client *Client) UpdateAppFeature(appGUID string, enabled bool, featureName string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchApplicationFeaturesRequest,
		Body:        bytes.NewReader([]byte(fmt.Sprintf(`{"enabled":%t}`, enabled))),
		URIParams:   map[string]string{"app_guid": appGUID, "name": featureName},
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
