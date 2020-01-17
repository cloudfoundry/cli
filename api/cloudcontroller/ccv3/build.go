package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Build represent the process of staging an application package.
type Build struct {
	// CreatedAt is the time with zone when the build was created.
	CreatedAt string
	// DropletGUID is the unique identifier for the resulting droplet from the
	// staging process.
	DropletGUID string
	// Error describes errors during the build process.
	Error string
	// GUID is the unique build identifier.
	GUID string
	// PackageGUID is the unique identifier for package that is the input to the
	// staging process.
	PackageGUID string
	// State is the state of the build.
	State constant.BuildState
}

// MarshalJSON converts a Build into a Cloud Controller Application.
func (b Build) MarshalJSON() ([]byte, error) {
	var ccBuild struct {
		Package struct {
			GUID string `json:"guid"`
		} `json:"package"`
	}

	ccBuild.Package.GUID = b.PackageGUID

	return json.Marshal(ccBuild)
}

// UnmarshalJSON helps unmarshal a Cloud Controller Build response.
func (b *Build) UnmarshalJSON(data []byte) error {
	var ccBuild struct {
		CreatedAt string `json:"created_at,omitempty"`
		GUID      string `json:"guid,omitempty"`
		Error     string `json:"error"`
		Package   struct {
			GUID string `json:"guid"`
		} `json:"package"`
		State   constant.BuildState `json:"state,omitempty"`
		Droplet struct {
			GUID string `json:"guid"`
		} `json:"droplet"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccBuild)
	if err != nil {
		return err
	}

	b.GUID = ccBuild.GUID
	b.CreatedAt = ccBuild.CreatedAt
	b.Error = ccBuild.Error
	b.PackageGUID = ccBuild.Package.GUID
	b.State = ccBuild.State
	b.DropletGUID = ccBuild.Droplet.GUID

	return nil
}

// CreateBuild creates the given build, requires Package GUID to be set on the
// build.
func (client *Client) CreateBuild(build Build) (Build, Warnings, error) {
	var responseBody Build

	warnings, err := client.makeCreateRequest(
		internal.PostBuildRequest,
		build,
		&responseBody,
	)

	return responseBody, warnings, err
}

// GetBuild gets the build with the given GUID.
func (client *Client) GetBuild(guid string) (Build, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetBuildRequest,
		URIParams:   internal.Params{"build_guid": guid},
	})
	if err != nil {
		return Build{}, nil, err
	}

	var responseBuild Build
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseBuild,
	}
	err = client.connection.Make(request, &response)

	return responseBuild, response.Warnings, err
}
