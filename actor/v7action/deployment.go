package v7action

func (actor Actor) CreateDeployment(appGUID string, dropletGUID string) (string, Warnings, error) {
	deploymentGUID, warnings, err := actor.CloudControllerClient.CreateApplicationDeployment(appGUID, dropletGUID)

	return deploymentGUID, Warnings(warnings), err
}
