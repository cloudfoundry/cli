package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

type Deployment struct {
	GUID          string
	State         constant.DeploymentState
	StatusValue   constant.DeploymentStatusValue
	StatusReason  constant.DeploymentStatusReason
	DropletGUID   string
	CreatedAt     string
	UpdatedAt     string
	Relationships resources.Relationships
	NewProcesses  []Process
}

// MarshalJSON converts a Deployment into a Cloud Controller Deployment.
func (d Deployment) MarshalJSON() ([]byte, error) {
	type Droplet struct {
		GUID string `json:"guid,omitempty"`
	}

	var ccDeployment struct {
		Droplet       *Droplet                `json:"droplet,omitempty"`
		Relationships resources.Relationships `json:"relationships,omitempty"`
	}

	if d.DropletGUID != "" {
		ccDeployment.Droplet = &Droplet{d.DropletGUID}
	}

	ccDeployment.Relationships = d.Relationships

	return json.Marshal(ccDeployment)
}

// UnmarshalJSON helps unmarshal a Cloud Controller Deployment response.
func (d *Deployment) UnmarshalJSON(data []byte) error {
	var ccDeployment struct {
		GUID          string                   `json:"guid,omitempty"`
		CreatedAt     string                   `json:"created_at,omitempty"`
		Relationships resources.Relationships  `json:"relationships,omitempty"`
		State         constant.DeploymentState `json:"state,omitempty"`
		Status        struct {
			Value  constant.DeploymentStatusValue  `json:"value"`
			Reason constant.DeploymentStatusReason `json:"reason"`
		} `json:"status"`
		Droplet      resources.Droplet `json:"droplet,omitempty"`
		NewProcesses []Process         `json:"new_processes,omitempty"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccDeployment)
	if err != nil {
		return err
	}

	d.GUID = ccDeployment.GUID
	d.CreatedAt = ccDeployment.CreatedAt
	d.Relationships = ccDeployment.Relationships
	d.State = ccDeployment.State
	d.StatusValue = ccDeployment.Status.Value
	d.StatusReason = ccDeployment.Status.Reason
	d.DropletGUID = ccDeployment.Droplet.GUID
	d.NewProcesses = ccDeployment.NewProcesses

	return nil
}

func (client *Client) CancelDeployment(deploymentGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.PostApplicationDeploymentActionCancelRequest,
		URIParams:   internal.Params{"deployment_guid": deploymentGUID},
	})

	return warnings, err
}

func (client *Client) CreateApplicationDeployment(appGUID string, dropletGUID string) (string, Warnings, error) {
	dep := Deployment{
		DropletGUID:   dropletGUID,
		Relationships: resources.Relationships{constant.RelationshipTypeApplication: resources.Relationship{GUID: appGUID}},
	}

	var responseBody Deployment

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostApplicationDeploymentRequest,
		RequestBody:  dep,
		ResponseBody: &responseBody,
	})

	return responseBody.GUID, warnings, err
}

func (client *Client) GetDeployment(deploymentGUID string) (Deployment, Warnings, error) {
	var responseBody Deployment

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetDeploymentRequest,
		URIParams:    internal.Params{"deployment_guid": deploymentGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetDeployments(query ...Query) ([]Deployment, Warnings, error) {
	var resources []Deployment

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetDeploymentsRequest,
		Query:        query,
		ResponseBody: Deployment{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Deployment))
			return nil
		},
	})

	return resources, warnings, err
}
