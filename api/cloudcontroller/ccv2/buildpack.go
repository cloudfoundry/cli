package ccv2

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/buildpacks"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"code.cloudfoundry.org/cli/types"
	"encoding/json"
	"io"
)

// Buildpack represents a Cloud Controller Buildpack.
type Buildpack struct {
	Locked   types.NullBool
	Enabled  types.NullBool
	GUID     string
	Name     string
	Position types.NullInt
	Stack    string
}

func (buildpack Buildpack) MarshalJSON() ([]byte, error) {
	ccBuildpack := struct {
		Locked   *bool  `json:"locked,omitempty"`
		Enabled  *bool  `json:"enabled,omitempty"`
		Name     string `json:"name"`
		Position *int   `json:"position,omitempty"`
		Stack    string `json:"stack,omitempty"`
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

func (buildpack *Buildpack) UnmarshalJSON(data []byte) error {
	var alias struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Locked   types.NullBool `json:"locked"`
			Enabled  types.NullBool `json:"enabled"`
			Name     string         `json:"name"`
			Position types.NullInt  `json:"position"`
			Stack    string         `json:"stack"`
		} `json:"entity"`
	}

	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}

	buildpack.Locked = alias.Entity.Locked
	buildpack.Enabled = alias.Entity.Enabled
	buildpack.GUID = alias.Metadata.GUID
	buildpack.Name = alias.Entity.Name
	buildpack.Position = alias.Entity.Position
	buildpack.Stack = alias.Entity.Stack
	return nil
}

// CreateBuildpack creates a new buildpack.
func (client *Client) CreateBuildpack(buildpack Buildpack) (Buildpack, Warnings, error) {
	body, err := json.Marshal(buildpack)
	if err != nil {
		return Buildpack{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostBuildpackRequest,
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Buildpack{}, nil, err
	}

	var createdBuildpack Buildpack
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &createdBuildpack,
	}

	err = client.connection.Make(request, &response)
	return createdBuildpack, response.Warnings, err
}

// GetBuildpacks searches for a buildpack with the given name and returns it if it exists.
func (client *Client) GetBuildpacks(filters ...Filter) ([]Buildpack, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetBuildpacksRequest,
		Query:       ConvertFilterParameters(filters),
	})

	if err != nil {
		return nil, nil, err
	}

	var buildpacks []Buildpack
	warnings, err := client.paginate(request, Buildpack{}, func(item interface{}) error {
		if buildpack, ok := item.(Buildpack); ok {
			buildpacks = append(buildpacks, buildpack)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Buildpack{},
				Unexpected: item,
			}
		}
		return nil
	})

	return buildpacks, warnings, err
}

// UpdateBuildpack updates the buildpack with the provided GUID and returns the
// updated buildpack. Note: Stack cannot be updated without uploading a new
// buildpack.
func (client *Client) UpdateBuildpack(buildpack Buildpack) (Buildpack, Warnings, error) {
	body, err := json.Marshal(buildpack)
	if err != nil {
		return Buildpack{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutBuildpackRequest,
		URIParams:   Params{"buildpack_guid": buildpack.GUID},
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Buildpack{}, nil, err
	}

	var updatedBuildpack Buildpack
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedBuildpack,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return Buildpack{}, response.Warnings, err
	}

	return updatedBuildpack, response.Warnings, nil
}

// UploadBuildpack uploads the contents of a buildpack zip to the server.
func (client *Client) UploadBuildpack(buildpackGUID string, buildpackPath string, buildpack io.Reader, buildpackLength int64) (Warnings, error) {

	contentLength, err := buildpacks.CalculateRequestSize(buildpackLength, buildpackPath)
	if err != nil {
		return nil, err
	}

	contentType, body, writeErrors := buildpacks.CreateMultipartBodyAndHeader(buildpack, buildpackPath)

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutBuildpackBitsRequest,
		URIParams:   Params{"buildpack_guid": buildpackGUID},
		Body:        body,
	})

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", contentType)
	request.ContentLength = contentLength

	_, warnings, err := client.uploadBuildpackAsynchronously(request, writeErrors)
	if err != nil {
		return warnings, err
	}
	return warnings, nil

}

func (client *Client) uploadBuildpackAsynchronously(request *cloudcontroller.Request, writeErrors <-chan error) (Buildpack, Warnings, error) {

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
	return buildpack, response.Warnings, firstError
}
