package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Droplet struct {
	GUID       string      `json:"guid"`
	Stack      string      `json:"stack,omitempty"`
	Buildpacks []Buildpack `json:"buildpacks,omitempty"`
}

type Buildpack struct {
	Name string `json:"name"`
}

// GetApplicationCurrentDroplet returns the Current Droplet for a given app
func (client *Client) GetApplicationCurrentDroplet(appGUID string) (Droplet, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppDropletCurrent,
		URIParams:   map[string]string{"guid": appGUID},
	})

	var responseDroplet Droplet
	response := cloudcontroller.Response{
		Result: &responseDroplet,
	}
	err = client.connection.Make(request, &response)

	return responseDroplet, response.Warnings, err
}
