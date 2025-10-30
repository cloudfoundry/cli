package ccv3

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v8/resources"
)

func (client *Client) ContinueDeployment(deploymentGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.PostApplicationDeploymentActionContinueRequest,
		URIParams:   internal.Params{"deployment_guid": deploymentGUID},
	})

	return warnings, err
}

func (client *Client) CancelDeployment(deploymentGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.PostApplicationDeploymentActionCancelRequest,
		URIParams:   internal.Params{"deployment_guid": deploymentGUID},
	})

	return warnings, err
}

func (client *Client) CreateApplicationDeployment(dep resources.Deployment) (string, Warnings, error) {

	var responseBody resources.Deployment

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostApplicationDeploymentRequest,
		RequestBody:  dep,
		ResponseBody: &responseBody,
	})

	return responseBody.GUID, warnings, err
}

func (client *Client) GetDeployment(deploymentGUID string) (resources.Deployment, Warnings, error) {
	var responseBody resources.Deployment

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetDeploymentRequest,
		URIParams:    internal.Params{"deployment_guid": deploymentGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetDeployments(query ...Query) ([]resources.Deployment, Warnings, error) {
	var deployments []resources.Deployment

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetDeploymentsRequest,
		Query:        query,
		ResponseBody: resources.Deployment{},
		AppendToList: func(item interface{}) error {
			deployments = append(deployments, item.(resources.Deployment))
			return nil
		},
	})

	return deployments, warnings, err
}
