package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Droplet represents a Cloud Controller droplet's metadata. A droplet is a set of
// compiled bits for a given application.
type Droplet struct {
	//Buildpacks are the detected buildpacks from the staging process.
	Buildpacks []DropletBuildpack `json:"buildpacks,omitempty"`
	// CreatedAt is the timestamp that the Cloud Controller created the droplet.
	CreatedAt string `json:"created_at"`
	// GUID is the unique droplet identifier.
	GUID string `json:"guid"`
	// Image is the Docker image name.
	Image string `json:"image"`
	// Stack is the root filesystem to use with the buildpack.
	Stack string `json:"stack,omitempty"`
	// State is the current state of the droplet.
	State constant.DropletState `json:"state"`
}

// DropletBuildpack is the name and output of a buildpack used to create a
// droplet.
type DropletBuildpack struct {
	// Name is the buildpack name.
	Name string `json:"name"`
	//DetectOutput is the output during buildpack detect process.
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
