package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Deployment struct {
	GUID          string
	State         constant.DeploymentState
	StatusValue   constant.DeploymentStatusValue
	StatusReason  constant.DeploymentStatusReason
	DropletGUID   string
	CreatedAt     string
	UpdatedAt     string
	Relationships Relationships
	NewProcesses  []Process
}

// MarshalJSON converts a Deployment into a Cloud Controller Deployment.
func (d Deployment) MarshalJSON() ([]byte, error) {
	type Droplet struct {
		GUID string `json:"guid,omitempty"`
	}

	var ccDeployment struct {
		Droplet       *Droplet      `json:"droplet,omitempty"`
		Relationships Relationships `json:"relationships,omitempty"`
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
		Relationships Relationships            `json:"relationships,omitempty"`
		State         constant.DeploymentState `json:"state,omitempty"`
		Status        struct {
			Value  constant.DeploymentStatusValue  `json:"value"`
			Reason constant.DeploymentStatusReason `json:"reason"`
		} `json:"status"`
		Droplet      Droplet   `json:"droplet,omitempty"`
		NewProcesses []Process `json:"new_processes,omitempty"`
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
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationDeploymentActionCancelRequest,
		URIParams:   map[string]string{"deployment_guid": deploymentGUID},
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

func (client *Client) CreateApplicationDeployment(appGUID string, dropletGUID string) (string, Warnings, error) {
	dep := Deployment{
		DropletGUID:   dropletGUID,
		Relationships: Relationships{constant.RelationshipTypeApplication: Relationship{GUID: appGUID}},
	}

	var responseBody Deployment

	warnings, err := client.makeCreateRequest(
		internal.PostApplicationDeploymentRequest,
		dep,
		&responseBody,
	)

	return responseBody.GUID, warnings, err
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
		DecodeJSONResponseInto: &responseDeployment,
	}
	err = client.connection.Make(request, &response)

	return responseDeployment, response.Warnings, err
}

func (client *Client) GetDeployments(query ...Query) ([]Deployment, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetDeploymentsRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err // untested
	}
	var deployments []Deployment
	warnings, err := client.paginate(request, Deployment{}, func(item interface{}) error {
		if deployment, ok := item.(Deployment); ok {
			deployments = append(deployments, deployment)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Deployment{},
				Unexpected: item,
			}
		}
		return nil
	})

	return deployments, warnings, err
}
