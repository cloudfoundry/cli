package ccv3

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"encoding/json"
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
	// Links are links to related resources.
	Links APILinks
}

// MarshalJSON converts a Package into a Cloud Controller Package.
func (b Buildpack) MarshalJSON() ([]byte, error) {
	var ccBuildpack struct {
		Name     string `json:"name,omitempty"`
		Stack    string `json:"stack,omitempty"`
		Position int    `json:"position,omitempty"`
		Enabled  bool   `json:"enabled,omitempty"`
		Locked   bool   `json:"locked,omitempty"`
	}

	ccBuildpack.Name = b.Name
	ccBuildpack.Stack = b.Stack
	ccBuildpack.Position = b.Position
	ccBuildpack.Enabled = b.Enabled
	ccBuildpack.Locked = b.Locked

	return json.Marshal(ccBuildpack)
}

func (b *Buildpack) UnmarshalJSON(data []byte) error {
	var ccBuildpack struct {
		Enabled  bool     `json:"enabled,omitempty"`
		Filename string   `json:"filename,omitempty"`
		GUID     string   `json:"guid,omitempty"`
		Locked   bool     `json:"locked,omitempty"`
		Name     string   `json:"name,omitempty"`
		Position int      `json:"position,omitempty"`
		Stack    string   `json:"stack,omitempty"`
		State    string   `json:"state,omitempty"`
		Links    APILinks `json:"links,omitempty"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccBuildpack)
	if err != nil {
		return err
	}

	b.Enabled = ccBuildpack.Enabled
	b.Filename = ccBuildpack.Filename
	b.GUID = ccBuildpack.GUID
	b.Locked = ccBuildpack.Locked
	b.Name = ccBuildpack.Name
	b.Position = ccBuildpack.Position
	b.Stack = ccBuildpack.Stack
	b.State = ccBuildpack.State
	b.Links = ccBuildpack.Links

	return nil
}

// CreateBuildpack creates a buildpack with the given settings, Type and the
// ApplicationRelationship must be set.
func (client *Client) CreateBuildpack(bp Buildpack) (Buildpack, Warnings, error) {
	bodyBytes, err := json.Marshal(bp)
	if err != nil {
		return Buildpack{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostBuildpackRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Buildpack{}, nil, err
	}

	var responseBuildpack Buildpack
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseBuildpack,
	}
	err = client.connection.Make(request, &response)

	return responseBuildpack, response.Warnings, err
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
