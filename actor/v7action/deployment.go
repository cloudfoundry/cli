package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type Deployment ccv3.Deployment

func (actor Actor) CreateDeployment(appGUID string, dropletGUID string) (string, Warnings, error) {
	deploymentGUID, warnings, err := actor.CloudControllerClient.CreateApplicationDeployment(appGUID, dropletGUID)

	return deploymentGUID, Warnings(warnings), err
}

func (actor Actor) GetLatestDeploymentForApp(appGUID string) (Deployment, Warnings, error) {
	ccDeployments, warnings, err := actor.CloudControllerClient.GetDeployments(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{appGUID}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{"-created_at"}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
	)

	if err != nil {
		return Deployment{}, Warnings(warnings), err
	}

	if len(ccDeployments) == 0 {
		return Deployment{}, Warnings(warnings), actionerror.DeploymentNotFoundError{}
	}

	return Deployment(ccDeployments[0]), Warnings(warnings), nil
}
func (actor Actor) CancelDeployment(deploymentGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.CancelDeployment(deploymentGUID)
	return Warnings(warnings), err
}
