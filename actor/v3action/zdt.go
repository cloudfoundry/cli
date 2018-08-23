package v3action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"regexp"
	"time"
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

func (actor Actor) CreateDeployment(appGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.CreateApplicationDeployment(appGUID)

	return Warnings(warnings), err
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
