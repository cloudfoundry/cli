package ccv3

import (
	"io"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/cloudcontroller/uploads"
	"code.cloudfoundry.org/cli/resources"
)

type DropletCreateRequest struct {
	Relationships resources.Relationships `json:"relationships"`
}

// CreateDroplet creates a new droplet without a package for the app with
// the given guid.
func (client *Client) CreateDroplet(appGUID string) (resources.Droplet, Warnings, error) {
	requestBody := DropletCreateRequest{
		Relationships: resources.Relationships{
			constant.RelationshipTypeApplication: resources.Relationship{GUID: appGUID},
		},
	}

	var responseBody resources.Droplet

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostDropletRequest,
		RequestBody:  requestBody,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetApplicationDropletCurrent returns the current droplet for a given
// application.
func (client *Client) GetApplicationDropletCurrent(appGUID string) (resources.Droplet, Warnings, error) {
	var responseBody resources.Droplet

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetApplicationDropletCurrentRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetDroplet returns a droplet with the given GUID.
func (client *Client) GetDroplet(dropletGUID string) (resources.Droplet, Warnings, error) {
	var responseBody resources.Droplet

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetDropletRequest,
		URIParams:    internal.Params{"droplet_guid": dropletGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetDroplets lists droplets with optional filters.
func (client *Client) GetDroplets(query ...Query) ([]resources.Droplet, Warnings, error) {
	var droplets []resources.Droplet

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetDropletsRequest,
		Query:        query,
		ResponseBody: resources.Droplet{},
		AppendToList: func(item interface{}) error {
			droplets = append(droplets, item.(resources.Droplet))
			return nil
		},
	})

	return droplets, warnings, err
}

// GetPackageDroplets returns the droplets that run the specified packages
func (client *Client) GetPackageDroplets(packageGUID string, query ...Query) ([]resources.Droplet, Warnings, error) {
	var droplets []resources.Droplet

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetPackageDropletsRequest,
		URIParams:    internal.Params{"package_guid": packageGUID},
		Query:        query,
		ResponseBody: resources.Droplet{},
		AppendToList: func(item interface{}) error {
			droplets = append(droplets, item.(resources.Droplet))
			return nil
		},
	})

	return droplets, warnings, err
}

// UploadDropletBits asynchronously uploads bits from a .tgz file located at dropletPath to the
// droplet with guid dropletGUID. It returns a job URL pointing to the asynchronous upload job.
func (client *Client) UploadDropletBits(dropletGUID string, dropletPath string, droplet io.Reader, dropletLength int64) (JobURL, Warnings, error) {
	contentLength, err := uploads.CalculateRequestSize(dropletLength, dropletPath, "bits")
	if err != nil {
		return "", nil, err
	}

	contentType, body, writeErrors := uploads.CreateMultipartBodyAndHeader(droplet, dropletPath, "bits")

	responseLocation, warnings, err := client.MakeRequestUploadAsync(
		internal.PostDropletBitsRequest,
		internal.Params{"droplet_guid": dropletGUID},
		contentType,
		body,
		contentLength,
		nil,
		writeErrors,
	)

	return JobURL(responseLocation), warnings, err
}
