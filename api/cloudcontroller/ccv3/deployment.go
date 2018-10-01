package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Deployment struct {
	GUID          string
	State         string
	Droplet       Droplet
	CreatedAt     string
	UpdatedAt     string
	Relationships Relationships
}

// MarshalJSON converts a Deployment into a Cloud Controller Deployment.
func (d Deployment) MarshalJSON() ([]byte, error) {
	var ccDeployment struct {
		Relationships Relationships `json:"relationships,omitempty"`
	}
	ccDeployment.Relationships = d.Relationships

	return json.Marshal(ccDeployment)
}

// UnmarshalJSON helps unmarshal a Cloud Controller Deployment response.
func (d *Deployment) UnmarshalJSON(data []byte) error {
	var ccDeployment struct {
		GUID          string        `json:"guid,omitempty"`
		CreatedAt     string        `json:"created_at,omitempty"`
		Relationships Relationships `json:"relationships,omitempty"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccDeployment)
	if err != nil {
		return err
	}

	d.GUID = ccDeployment.GUID
	d.CreatedAt = ccDeployment.CreatedAt
	d.Relationships = ccDeployment.Relationships

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

func (client *Client) CreateApplicationDeployment(appGUID string) (Warnings, error) {
	dep := Deployment{
		Relationships: Relationships{constant.RelationshipTypeApplication: Relationship{GUID: appGUID}},
	}
	bodyBytes, err := json.Marshal(dep)

	if err != nil {
		return nil, err
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationDeploymentRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return nil, err
	}

	var responseDeployment Deployment
	response := cloudcontroller.Response{
		Result: &responseDeployment,
	}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
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
	if err != nil {
		return nil, nil, err // untested
	}
	return deployments, warnings, nil
}
