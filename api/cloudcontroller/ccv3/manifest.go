package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// GetApplicationManifest returns a (YAML) manifest for an application and its
// underlying processes.
func (client *Client) GetApplicationManifest(appGUID string) ([]byte, Warnings, error) {
	request, err := client.NewHTTPRequest(requestOptions{
		RequestName: internal.GetApplicationManifestRequest,
		URIParams:   internal.Params{"app_guid": appGUID},
	})
	if err != nil {
		return nil, nil, err
	}
	request.Header.Set("Accept", "application/x-yaml")

	var response cloudcontroller.Response
	err = client.Connection.Make(request, &response)

	return response.RawResponse, response.Warnings, err
}
