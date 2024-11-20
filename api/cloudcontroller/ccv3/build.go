package ccv3

import (
	apiinternal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
	"code.cloudfoundry.org/cli/resources"
)

// CreateBuild creates the given build, requires Package GUID to be set on the
// build.
func (client *Client) CreateBuild(build resources.Build) (resources.Build, Warnings, error) {
	var responseBody resources.Build

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  apiinternal.PostBuildRequest,
		RequestBody:  build,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetBuild gets the build with the given GUID.
func (client *Client) GetBuild(guid string) (resources.Build, Warnings, error) {
	var responseBody resources.Build

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  apiinternal.GetBuildRequest,
		URIParams:    internal.Params{"build_guid": guid},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
