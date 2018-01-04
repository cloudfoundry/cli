package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Droplet represents a cloud controller droplet's metadata. A droplet is a set of
// compiled bits for a given application.
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

// GetApplicationDropletCurrent returns the current droplet for a given
// application.
func (client *Client) GetApplicationDropletCurrent(appGUID string) (Droplet, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetApplicationDropletCurrentRequest,
		URIParams:   map[string]string{"app_guid": appGUID},
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

// GetDroplets lists droplets with optional filters.
func (client *Client) GetDroplets(query ...Query) ([]Droplet, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetDropletsRequest,
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

// GetDroplet returns a droplet with the given GUID.
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
