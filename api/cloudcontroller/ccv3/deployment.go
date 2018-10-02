package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Deployment struct {
	GUID          string
	State         constant.DeploymentState
	DropletGUID   string
	CreatedAt     string
	UpdatedAt     string
	Relationships Relationships
}

// MarshalJSON converts a Deployment into a Cloud Controller Deployment.
func (d Deployment) MarshalJSON() ([]byte, error) {
	var ccDeployment struct {
		Droplet struct {
			GUID string `json:"guid"`
		} `json:"droplet"`
		Relationships Relationships `json:"relationships,omitempty"`
	}
	ccDeployment.Droplet.GUID = d.DropletGUID
	ccDeployment.Relationships = d.Relationships

	return json.Marshal(ccDeployment)
}

// UnmarshalJSON helps unmarshal a Cloud Controller Deployment response.
func (d *Deployment) UnmarshalJSON(data []byte) error {
	var ccDeployment struct {
		GUID          string                   `json:"guid,omitempty"`
		CreatedAt     string                   `json:"created_at,omitempty"`
		Relationships Relationships            `json:"relationships,omitempty"`
		State         constant.DeploymentState `json:"state,omitempty"`
		Droplet       Droplet                  `json:"droplet,omitempty"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccDeployment)
	if err != nil {
		return err
	}

	d.GUID = ccDeployment.GUID
	d.CreatedAt = ccDeployment.CreatedAt
	d.Relationships = ccDeployment.Relationships
	d.State = ccDeployment.State
	d.DropletGUID = ccDeployment.Droplet.GUID

	return nil
}

func (client *Client) CreateApplicationDeployment(appGUID string, dropletGUID string) (string, Warnings, error) {
	dep := Deployment{
		DropletGUID:   dropletGUID,
		Relationships: Relationships{constant.RelationshipTypeApplication: Relationship{GUID: appGUID}},
	}
	bodyBytes, err := json.Marshal(dep)

	if err != nil {
		return "", nil, err
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationDeploymentRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return "", nil, err
	}

	var responseDeployment Deployment
	response := cloudcontroller.Response{
		Result: &responseDeployment,
	}
	err = client.connection.Make(request, &response)

	return responseDeployment.GUID, response.Warnings, err
}

func (client *Client) GetDeployment(deploymentGUID string) (Deployment, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetDeploymentRequest,
		URIParams:   internal.Params{"deployment_guid": deploymentGUID},
	})
	if err != nil {
		return Deployment{}, nil, err
	}

	var responseDeployment Deployment
	response := cloudcontroller.Response{
		Result: &responseDeployment,
	}
	err = client.connection.Make(request, &response)

	return responseDeployment, response.Warnings, err
}
