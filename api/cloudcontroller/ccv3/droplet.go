package ccv3

import (
	"io"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
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

	_, warnings, err := client.makeRequest(requestParams{
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

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.GetApplicationDropletCurrentRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetDroplet returns a droplet with the given GUID.
func (client *Client) GetDroplet(dropletGUID string) (Droplet, Warnings, error) {
	var responseBody Droplet

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.GetDropletRequest,
		URIParams:    internal.Params{"droplet_guid": dropletGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetDroplets lists droplets with optional filters.
func (client *Client) GetDroplets(query ...Query) ([]Droplet, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetDropletsRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var responseDroplets []Droplet
	warnings, err := client.paginate(request, Droplet{}, func(item interface{}) error {
		if droplet, ok := item.(Droplet); ok {
			responseDroplets = append(responseDroplets, droplet)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Droplet{},
				Unexpected: item,
			}
		}
		return nil
	})

	return responseDroplets, warnings, err
}

// GetPackageDroplets returns the droplets that run the specified packages
func (client *Client) GetPackageDroplets(packageGUID string, query ...Query) ([]Droplet, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetPackageDropletsRequest,
		URIParams:   map[string]string{"package_guid": packageGUID},
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var responseDroplets []Droplet
	warnings, err := client.paginate(request, Droplet{}, func(item interface{}) error {
		if droplet, ok := item.(Droplet); ok {
			responseDroplets = append(responseDroplets, droplet)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Droplet{},
				Unexpected: item,
			}
		}
		return nil
	})

	return responseDroplets, warnings, err
}

// UploadDropletBits asynchronously uploads bits from a .tgz file located at dropletPath to the
// droplet with guid dropletGUID. It returns a job URL pointing to the asynchronous upload job.
func (client *Client) UploadDropletBits(dropletGUID string, dropletPath string, droplet io.Reader, dropletLength int64) (JobURL, Warnings, error) {
	contentLength, err := uploads.CalculateRequestSize(dropletLength, dropletPath, "bits")
	if err != nil {
		return "", nil, err
	}

	contentType, body, writeErrors := uploads.CreateMultipartBodyAndHeader(droplet, dropletPath, "bits")

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostDropletBitsRequest,
		URIParams:   internal.Params{"droplet_guid": dropletGUID},
		Body:        body,
	})
	if err != nil {
		return "", nil, err
	}

	request.ContentLength = contentLength
	request.Header.Set("Content-Type", contentType)

	jobURL, warnings, err := client.uploadDropletAsynchronously(request, writeErrors)
	if err != nil {
		return "", warnings, err
	}

	return jobURL, warnings, nil
}

func (client *Client) uploadDropletAsynchronously(request *cloudcontroller.Request, writeErrors <-chan error) (JobURL, Warnings, error) {
	var droplet Droplet
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &droplet,
	}

	httpErrors := make(chan error)

	go func() {
		defer close(httpErrors)

		err := client.connection.Make(request, &response)
		if err != nil {
			httpErrors <- err
		}
	}()

	// The following section makes the following assumptions:
	// 1) If an error occurs during file reading, an EOF is sent to the request
	// object. Thus ending the request transfer.
	// 2) If an error occurs during request transfer, an EOF is sent to the pipe.
	// Thus ending the writing routine.
	var firstError error
	var writeClosed, httpClosed bool

	for {
		select {
		case writeErr, ok := <-writeErrors:
			if !ok {
				writeClosed = true
				break // for select
			}
			if firstError == nil {
				firstError = writeErr
			}
		case httpErr, ok := <-httpErrors:
			if !ok {
				httpClosed = true
				break // for select
			}
			if firstError == nil {
				firstError = httpErr
			}
		}

		if writeClosed && httpClosed {
			break // for for
		}
	}

	return JobURL(response.ResourceLocationURL), response.Warnings, firstError
}
