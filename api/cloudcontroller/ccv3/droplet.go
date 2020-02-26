package ccv3

import (
	"io"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/cloudcontroller/uploads"
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

type DropletCreateRequest struct {
	Relationships Relationships `json:"relationships"`
}

// CreateDroplet creates a new droplet without a package for the app with
// the given guid.
func (client *Client) CreateDroplet(appGUID string) (Droplet, Warnings, error) {
	requestBody := DropletCreateRequest{
		Relationships: Relationships{
			constant.RelationshipTypeApplication: Relationship{GUID: appGUID},
		},
	}

	var responseBody Droplet

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostDropletRequest,
		RequestBody:  requestBody,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetApplicationDropletCurrent returns the current droplet for a given
// application.
func (client *Client) GetApplicationDropletCurrent(appGUID string) (Droplet, Warnings, error) {
	var responseBody Droplet

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetApplicationDropletCurrentRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetDroplet returns a droplet with the given GUID.
func (client *Client) GetDroplet(dropletGUID string) (Droplet, Warnings, error) {
	var responseBody Droplet

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetDropletRequest,
		URIParams:    internal.Params{"droplet_guid": dropletGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetDroplets lists droplets with optional filters.
func (client *Client) GetDroplets(query ...Query) ([]Droplet, Warnings, error) {
	var resources []Droplet

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetDropletsRequest,
		Query:        query,
		ResponseBody: Droplet{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Droplet))
			return nil
		},
	})

	return resources, warnings, err
}

// GetPackageDroplets returns the droplets that run the specified packages
func (client *Client) GetPackageDroplets(packageGUID string, query ...Query) ([]Droplet, Warnings, error) {
	var resources []Droplet

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetPackageDropletsRequest,
		URIParams:    internal.Params{"package_guid": packageGUID},
		Query:        query,
		ResponseBody: Droplet{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Droplet))
			return nil
		},
	})

	return resources, warnings, err
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
