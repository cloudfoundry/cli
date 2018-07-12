package ccv2

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Buildpack represents a Cloud Controller Buildpack.
type Buildpack struct {
	Enabled  bool   `json:"enabled,omitempty"`
	GUID     string `json:"guid,omitempty"`
	Name     string `json:"name"`
	Position int    `json:"position,omitempty"`
}

func (buildpack *Buildpack) UnmarshalJSON(data []byte) error {
	var alias struct {
		Metadata struct {
			GUID string `json:"guid"`
		} `json:"metadata"`
		Entity struct {
			Name     string `json:"name"`
			Position int    `json:"position"`
			Enabled  bool   `json:"enabled"`
		} `json:"entity"`
	}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}

	buildpack.Enabled = alias.Entity.Enabled
	buildpack.GUID = alias.Metadata.GUID
	buildpack.Name = alias.Entity.Name
	buildpack.Position = alias.Entity.Position

	return nil
}

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
		Result: &createdBuildpack,
	}

	err = client.connection.Make(request, &response)
	return createdBuildpack, response.Warnings, err
}
