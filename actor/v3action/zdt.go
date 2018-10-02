package v3action

import (
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
)

func (actor Actor) ZeroDowntimePollStart(appGUID string, warningsChannel chan<- Warnings) error {
	processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(appGUID)
	warningsChannel <- Warnings(warnings)

	if err != nil {
		return err
	}

	timeout := time.Now().Add(actor.Config.StartupTimeout())
	for time.Now().Before(timeout) {
		deployingProcess := getDeployingProcess(processes)
		if deployingProcess == nil {
			return nil
		}
		ready, err := actor.processStatus(*deployingProcess, warningsChannel)
		if err != nil {
			return err
		}

		if ready {
			return nil
		}

		time.Sleep(actor.Config.PollingInterval())
	}

	return actionerror.StartupTimeoutError{}
}

func (actor Actor) CreateDeployment(appGUID string, dropletGUID string) (string, Warnings, error) {
	deploymentGUID, warnings, err := actor.CloudControllerClient.CreateApplicationDeployment(appGUID, dropletGUID)

	return deploymentGUID, Warnings(warnings), err
}

func getDeployingProcess(processes []ccv3.Process) *ccv3.Process {
	deployingMatcher, _ := regexp.Compile("web-deployment-.*")
	for _, process := range processes {
		if deployingMatcher.MatchString(process.Type) {
			return &process
		}
	}
	return nil
}

func (actor Actor) GetDeploymentState(deploymentGUID string) (constant.DeploymentState, Warnings, error) {
	deployment, warnings, err := actor.CloudControllerClient.GetDeployment(deploymentGUID)
	if err != nil {
		return "", Warnings(warnings), err
	}
	return deployment.State, Warnings(warnings), nil
}

func (actor Actor) PollDeployment(deploymentGUID string, warningsChannel chan<- Warnings) error {
	timeout := time.Now().Add(actor.Config.StartupTimeout())
	for time.Now().Before(timeout) {
		deploymentState, warnings, err := actor.GetDeploymentState(deploymentGUID)
		warningsChannel <- Warnings(warnings)
		if err != nil {
			return err
		}
		switch deploymentState {
		case constant.DeploymentDeployed:
			return nil
		case constant.DeploymentCanceled:
			return errors.New("Deployment has been canceled")
		case constant.DeploymentDeploying:
			time.Sleep(actor.Config.PollingInterval())
		}
	}

	return actionerror.StartupTimeoutError{}

}
