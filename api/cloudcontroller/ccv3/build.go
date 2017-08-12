package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type BuildState string

const (
	BuildStateFailed  BuildState = "FAILED"
	BuildStateStaged  BuildState = "STAGED"
	BuildStateStaging BuildState = "STAGING"
)

type Build struct {
	GUID    string     `json:"guid,omitempty"`
	Error   string     `json:"error"`
	Package Package    `json:"package"`
	State   BuildState `json:"state,omitempty"`
	Droplet Droplet    `json:"droplet"`
}

func (b Build) MarshalJSON() ([]byte, error) {
	var ccBuild struct {
		Package struct {
			GUID string `json:"guid"`
		} `json:"package"`
	}

	ccBuild.Package.GUID = b.Package.GUID

	return json.Marshal(ccBuild)
}

// CreateBuild creates the given build, requires Package GUID to be set on the
// build.
func (client *Client) CreateBuild(build Build) (Build, Warnings, error) {
	bodyBytes, err := json.Marshal(build)
	if err != nil {
		return Build{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostBuildRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Build{}, nil, err
	}

	var responseBuild Build
	response := cloudcontroller.Response{
		Result: &responseBuild,
	}
	err = client.connection.Make(request, &response)

	return responseBuild, response.Warnings, err
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
		Result: &responseBuild,
	}
	err = client.connection.Make(request, &response)

	return responseBuild, response.Warnings, err
}
