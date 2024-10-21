package ccv3

import (
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v7/resources"
)

type SSHEnabled struct {
	Enabled bool
	Reason  string
}

func (client *Client) GetAppFeature(appGUID string, featureName string) (resources.ApplicationFeature, Warnings, error) {
	var responseBody resources.ApplicationFeature

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetApplicationFeaturesRequest,
		URIParams:    internal.Params{"app_guid": appGUID, "name": featureName},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetSSHEnabled(appGUID string) (SSHEnabled, Warnings, error) {
	var responseBody SSHEnabled

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetSSHEnabled,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UpdateAppFeature enables/disables the ability to ssh for a given application.
func (client *Client) UpdateAppFeature(appGUID string, enabled bool, featureName string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.PatchApplicationFeaturesRequest,
		RequestBody: struct {
			Enabled bool `json:"enabled"`
		}{Enabled: enabled},
		URIParams: internal.Params{"app_guid": appGUID, "name": featureName},
	})

	return warnings, err
}
