package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Buildpack represents a Cloud Controller V3 buildpack.
type Buildpack struct {
	// Enabled is true when the buildpack can be used for staging.
	Enabled bool
	// Filename is the uploaded filename of the buildpack.
	Filename string
	// GUID is the unique identifier for the buildpack.
	GUID string
	// Locked is true when the buildpack cannot be updated.
	Locked bool
	// Name is the name of the buildpack. To be used by app buildpack field.
	// (only alphanumeric characters)
	Name string
	// Position is the order in which the buildpacks are checked during buildpack
	// auto-detection.
	Position int
	// Stack is the name of the stack that the buildpack will use.
	Stack string
	// State is the current state of the buildpack.
	State string
}

// GetBuildpacks lists buildpacks with optional filters.
func (client *Client) GetBuildpacks(query ...Query) ([]Buildpack, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetBuildpacksRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullBuildpacksList []Buildpack
	warnings, err := client.paginate(request, Buildpack{}, func(item interface{}) error {
		if buildpack, ok := item.(Buildpack); ok {
			fullBuildpacksList = append(fullBuildpacksList, buildpack)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Buildpack{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullBuildpacksList, warnings, err
}
