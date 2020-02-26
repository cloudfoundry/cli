package ccv3

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/cloudcontroller/uploads"
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
	// Metadata is used for custom tagging of API resources
	Metadata *Metadata
}

// MarshalJSON converts a Package into a Cloud Controller Package.
func (buildpack Buildpack) MarshalJSON() ([]byte, error) {
	ccBuildpack := struct {
		Name     string    `json:"name,omitempty"`
		Stack    string    `json:"stack,omitempty"`
		Position *int      `json:"position,omitempty"`
		Enabled  *bool     `json:"enabled,omitempty"`
		Locked   *bool     `json:"locked,omitempty"`
		Metadata *Metadata `json:"metadata,omitempty"`
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
		Metadata *Metadata      `json:"metadata"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccBuildpack)
	if err != nil {
		return err
	}

	buildpack.Enabled = ccBuildpack.Enabled
	buildpack.Filename = ccBuildpack.Filename
	buildpack.GUID = ccBuildpack.GUID
	buildpack.Locked = ccBuildpack.Locked
	buildpack.Name = ccBuildpack.Name
	buildpack.Position = ccBuildpack.Position
	buildpack.Stack = ccBuildpack.Stack
	buildpack.State = ccBuildpack.State
	buildpack.Links = ccBuildpack.Links
	buildpack.Metadata = ccBuildpack.Metadata

	return nil
}

// CreateBuildpack creates a buildpack with the given settings, Type and the
// ApplicationRelationship must be set.
func (client *Client) CreateBuildpack(bp Buildpack) (Buildpack, Warnings, error) {
	var responseBody Buildpack

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
func (client *Client) GetBuildpacks(query ...Query) ([]Buildpack, Warnings, error) {
	var resources []Buildpack

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetBuildpacksRequest,
		Query:        query,
		ResponseBody: Buildpack{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Buildpack))
			return nil
		},
	})

	return resources, warnings, err
}

func (client Client) UpdateBuildpack(buildpack Buildpack) (Buildpack, Warnings, error) {
	var responseBody Buildpack

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
