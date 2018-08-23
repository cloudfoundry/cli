package v3action

import "github.com/code.cloudfoundry.org/cli/actor/v3action"

func (actor Actor) CreateApplicationDeployment(appGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.CreateApplicationDeployment(appGUID)

	return Warnings(warnings), err
}

func (actor Actor) ZdtPollStart(appGUID string, warningsChannel chan<- v3action.Warnings) error {
	return nil
}
