package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
)

func (actor Actor) CreateDeploymentByApplicationAndDroplet(appGUID string, dropletGUID string) (string, Warnings, error) {
	deploymentGUID, warnings, err := actor.CloudControllerClient.CreateApplicationDeployment(appGUID, dropletGUID)

	return deploymentGUID, Warnings(warnings), err
}

func (actor Actor) CreateDeploymentByApplicationAndRevision(appGUID string, revisionGUID string) (string, Warnings, error) {
	deploymentGUID, warnings, err := actor.CloudControllerClient.CreateApplicationDeploymentByRevision(appGUID, revisionGUID)

	return deploymentGUID, Warnings(warnings), err
}

func (actor Actor) GetLatestActiveDeploymentForApp(appGUID string) (resources.Deployment, Warnings, error) {
	ccDeployments, warnings, err := actor.CloudControllerClient.GetDeployments(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{appGUID}},
		ccv3.Query{Key: ccv3.StatusValueFilter, Values: []string{string(constant.DeploymentStatusValueActive)}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
	)

	if err != nil {
		return resources.Deployment{}, Warnings(warnings), err
	}

	if len(ccDeployments) == 0 {
		return resources.Deployment{}, Warnings(warnings), actionerror.ActiveDeploymentNotFoundError{}
	}

	return resources.Deployment(ccDeployments[0]), Warnings(warnings), nil
}

func (actor Actor) CancelDeployment(deploymentGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.CancelDeployment(deploymentGUID)
	return Warnings(warnings), err
}
