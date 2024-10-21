package ccv3

import (
	"io"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/uploads"
	"code.cloudfoundry.org/cli/v9/resources"
)

// CreateBuildpack creates a buildpack with the given settings, Type and the
// ApplicationRelationship must be set.
func (client *Client) CreateBuildpack(bp resources.Buildpack) (resources.Buildpack, Warnings, error) {
	var responseBody resources.Buildpack

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostBuildpackRequest,
		RequestBody:  bp,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// DeleteBuildpack deletes the buildpack with the provided guid.
func (client Client) DeleteBuildpack(buildpackGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteBuildpackRequest,
		URIParams:   internal.Params{"buildpack_guid": buildpackGUID},
	})

	return jobURL, warnings, err
}

// GetBuildpacks lists buildpacks with optional filters.
func (client *Client) GetBuildpacks(query ...Query) ([]resources.Buildpack, Warnings, error) {
	var buildpacks []resources.Buildpack

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetBuildpacksRequest,
		Query:        query,
		ResponseBody: resources.Buildpack{},
		AppendToList: func(item interface{}) error {
			buildpacks = append(buildpacks, item.(resources.Buildpack))
			return nil
		},
	})

	return buildpacks, warnings, err
}

func (client Client) UpdateBuildpack(buildpack resources.Buildpack) (resources.Buildpack, Warnings, error) {
	var responseBody resources.Buildpack

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchBuildpackRequest,
		URIParams:    internal.Params{"buildpack_guid": buildpack.GUID},
		RequestBody:  buildpack,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UploadBuildpack uploads the contents of a buildpack zip to the server.
func (client *Client) UploadBuildpack(buildpackGUID string, buildpackPath string, buildpack io.Reader, buildpackLength int64) (JobURL, Warnings, error) {

	contentLength, err := uploads.CalculateRequestSize(buildpackLength, buildpackPath, "bits")
	if err != nil {
		return "", nil, err
	}

	contentType, body, writeErrors := uploads.CreateMultipartBodyAndHeader(buildpack, buildpackPath, "bits")

	responseLocation, warnings, err := client.MakeRequestUploadAsync(
		internal.PostBuildpackBitsRequest,
		internal.Params{"buildpack_guid": buildpackGUID},
		contentType,
		body,
		contentLength,
		nil,
		writeErrors,
	)

	return JobURL(responseLocation), warnings, err
}
