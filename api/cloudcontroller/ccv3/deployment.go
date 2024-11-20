package ccv3

import (
	ccv3internal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) ContinueDeployment(deploymentGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: ccv3internal.PostApplicationDeploymentActionContinueRequest,
		URIParams:   internal.Params{"deployment_guid": deploymentGUID},
	})

	return warnings, err
}

func (client *Client) CancelDeployment(deploymentGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: ccv3internal.PostApplicationDeploymentActionCancelRequest,
		URIParams:   internal.Params{"deployment_guid": deploymentGUID},
	})

	return warnings, err
}

func (client *Client) CreateApplicationDeployment(dep resources.Deployment) (string, Warnings, error) {

	var responseBody resources.Deployment

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.PostApplicationDeploymentRequest,
		RequestBody:  dep,
		ResponseBody: &responseBody,
	})

	return responseBody.GUID, warnings, err
}

func (client *Client) GetDeployment(deploymentGUID string) (resources.Deployment, Warnings, error) {
	var responseBody resources.Deployment

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.GetDeploymentRequest,
		URIParams:    internal.Params{"deployment_guid": deploymentGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetDeployments(query ...Query) ([]resources.Deployment, Warnings, error) {
	var deployments []resources.Deployment

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  ccv3internal.GetDeploymentsRequest,
		Query:        query,
		ResponseBody: resources.Deployment{},
		AppendToList: func(item interface{}) error {
			deployments = append(deployments, item.(resources.Deployment))
			return nil
		},
	})

	return deployments, warnings, err
}
