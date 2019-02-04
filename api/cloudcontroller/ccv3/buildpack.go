package ccv3

import (
	"bytes"
	"encoding/json"
	"io"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/buildpacks"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

// Buildpack represents a Cloud Controller V3 buildpack.
type Buildpack struct {
	// Enabled is true when the buildpack can be used for staging.
	Enabled types.NullBool
	// Filename is the uploaded filename of the buildpack.
	Filename string
	// GUID is the unique identifier for the buildpack.
	GUID string
	// Locked is true when the buildpack cannot be updated.
	Locked types.NullBool
	// Name is the name of the buildpack. To be used by app buildpack field.
	// (only alphanumeric characters)
	Name string
	// Position is the order in which the buildpacks are checked during buildpack
	// auto-detection.
	Position types.NullInt
	// Stack is the name of the stack that the buildpack will use.
	Stack string
	// State is the current state of the buildpack.
	State string
	// Links are links to related resources.
	Links APILinks
}

// MarshalJSON converts a Package into a Cloud Controller Package.
func (buildpack Buildpack) MarshalJSON() ([]byte, error) {
	ccBuildpack := struct {
		Name     string `json:"name,omitempty"`
		Stack    string `json:"stack,omitempty"`
		Position *int   `json:"position,omitempty"`
		Enabled  *bool  `json:"enabled,omitempty"`
		Locked   *bool  `json:"locked,omitempty"`
	}{
		Name:  buildpack.Name,
		Stack: buildpack.Stack,
	}

	if buildpack.Position.IsSet {
		ccBuildpack.Position = &buildpack.Position.Value
	}
	if buildpack.Enabled.IsSet {
		ccBuildpack.Enabled = &buildpack.Enabled.Value
	}
	if buildpack.Locked.IsSet {
		ccBuildpack.Locked = &buildpack.Locked.Value
	}

	return json.Marshal(ccBuildpack)
}

func (b *Buildpack) UnmarshalJSON(data []byte) error {
	var ccBuildpack struct {
		GUID     string         `json:"guid,omitempty"`
		Links    APILinks       `json:"links,omitempty"`
		Name     string         `json:"name,omitempty"`
		Filename string         `json:"filename,omitempty"`
		Stack    string         `json:"stack,omitempty"`
		State    string         `json:"state,omitempty"`
		Enabled  types.NullBool `json:"enabled"`
		Locked   types.NullBool `json:"locked"`
		Position types.NullInt  `json:"position"`
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

// Delete a buildpack by guid
func (client Client) DeleteBuildpack(buildpackGUID string) (JobURL, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteBuildpackRequest,
		URIParams: map[string]string{
			"buildpack_guid": buildpackGUID,
		},
	})
	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return JobURL(response.ResourceLocationURL), response.Warnings, err
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

func (client Client) UpdateBuildpack(buildpack Buildpack) (Buildpack, Warnings, error) {
	bodyBytes, err := json.Marshal(buildpack)
	if err != nil {
		return Buildpack{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchBuildpackRequest,
		Body:        bytes.NewReader(bodyBytes),
		URIParams:   map[string]string{"buildpack_guid": buildpack.GUID},
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

// UploadBuildpack uploads the contents of a buildpack zip to the server.
func (client *Client) UploadBuildpack(buildpackGUID string, buildpackPath string, buildpack io.Reader, buildpackLength int64) (JobURL, Warnings, error) {

	contentLength, err := buildpacks.CalculateRequestSize(buildpackLength, buildpackPath, "bits")
	if err != nil {
		return "", nil, err
	}

	contentType, body, writeErrors := buildpacks.CreateMultipartBodyAndHeader(buildpack, buildpackPath, "bits")

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostBuildpackBitsRequest,
		URIParams:   internal.Params{"buildpack_guid": buildpackGUID},
		Body:        body,
	})

	if err != nil {
		return "", nil, err
	}

	request.ContentLength = contentLength
	request.Header.Set("Content-Type", contentType)

	jobURL, warnings, err := client.uploadBuildpackAsynchronously(request, writeErrors)
	if err != nil {
		return "", warnings, err
	}
	return jobURL, warnings, nil
}

func (client *Client) uploadBuildpackAsynchronously(request *cloudcontroller.Request, writeErrors <-chan error) (JobURL, Warnings, error) {

	var buildpack Buildpack
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &buildpack,
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
