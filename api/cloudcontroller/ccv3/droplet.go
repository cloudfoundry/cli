package ccv3

import (
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Droplet struct {
	GUID       string                `json:"guid"`
	State      constant.DropletState `json:"state"`
	CreatedAt  string                `json:"created_at"`
	Stack      string                `json:"stack,omitempty"`
	Buildpacks []DropletBuildpack    `json:"buildpacks,omitempty"`
	Image      string                `json:"image"`
}

type DropletBuildpack struct {
	Name         string `json:"name"`
	DetectOutput string `json:"detect_output"`
}

// GetApplicationDroplets returns the Droplets for a given app
func (client *Client) GetApplicationDroplets(appGUID string, query url.Values) ([]Droplet, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppDropletsRequest,
		URIParams:   map[string]string{"app_guid": appGUID},
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var responseDroplets []Droplet
	warnings, err := client.paginate(request, Droplet{}, func(item interface{}) error {
		if droplet, ok := item.(Droplet); ok {
			responseDroplets = append(responseDroplets, droplet)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Droplet{},
				Unexpected: item,
			}
		}
		return nil
	})

	return responseDroplets, warnings, err
}

func (client *Client) GetDroplet(dropletGUID string) (Droplet, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetDropletRequest,
		URIParams:   map[string]string{"droplet_guid": dropletGUID},
	})
	if err != nil {
		return Droplet{}, nil, err
	}

	var responseDroplet Droplet
	response := cloudcontroller.Response{
		Result: &responseDroplet,
	}
	err = client.connection.Make(request, &response)

	return responseDroplet, response.Warnings, err
}
