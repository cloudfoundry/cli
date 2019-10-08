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
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetApplicationFeaturesRequest,
		URIParams:   map[string]string{"app_guid": appGUID, "name": featureName},
	})

	if err != nil {
		return ApplicationFeature{}, nil, err
	}

	var applicationFeature ApplicationFeature
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &applicationFeature,
	}

	err = client.connection.Make(request, &response)

	return applicationFeature, response.Warnings, err
}

func (client *Client) GetSSHEnabled(appGUID string) (SSHEnabled, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSSHEnabled,
		URIParams:   map[string]string{"app_guid": appGUID},
	})

	if err != nil {
		return SSHEnabled{}, nil, err
	}

	var sshEnabled SSHEnabled
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &sshEnabled,
	}

	err = client.connection.Make(request, &response)

	return sshEnabled, response.Warnings, err
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
