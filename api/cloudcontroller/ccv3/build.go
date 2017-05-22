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
	Package Package    `json:"package"`
	State   BuildState `json:"state,omitempty"`
}

func (client *Client) CreateBuild(build Build) (Build, Warnings, error) {
	bodyBytes, err := json.Marshal(build)
	if err != nil {
		return Build{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostBuildRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	var responseBuild Build
	response := cloudcontroller.Response{
		Result: &responseBuild,
	}
	err = client.connection.Make(request, &response)

	return responseBuild, response.Warnings, err
}

func (client *Client) GetBuild(guid string) (Build, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetBuildRequest,
		URIParams:   internal.Params{"guid": guid},
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
