package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Build struct {
	CreatedAt   string
	GUID        string
	Error       string
	PackageGUID string
	State       constant.BuildState
	DropletGUID string
}

func (b Build) MarshalJSON() ([]byte, error) {
	var ccBuild struct {
		Package struct {
			GUID string `json:"guid"`
		} `json:"package"`
	}

	ccBuild.Package.GUID = b.PackageGUID

	return json.Marshal(ccBuild)
}

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

	if err := json.Unmarshal(data, &ccBuild); err != nil {
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
