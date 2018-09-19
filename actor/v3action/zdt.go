package v3action

import (
	"errors"
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
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

func (actor Actor) CancelDeploymentByAppNameAndSpace(appName string, spaceGUID string) (Warnings, error) {
	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return warnings, err
	}

	deploymentGuid, deploymentWarnings, err := actor.GetCurrentDeployment(app.GUID)
	warnings = append(warnings, deploymentWarnings...)
	if err != nil {
		return warnings, err
	}

	apiWarnings, err := actor.CloudControllerClient.CancelDeployment(deploymentGuid)
	warnings = append(warnings, apiWarnings...)

	return warnings, err
}

func (actor Actor) GetCurrentDeployment(appGUID string) (string, Warnings, error) {
	var collectedWarnings Warnings
	deployments, warnings, err := actor.CloudControllerClient.GetDeployments(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{appGUID}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{"-created_at"}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
	)
	collectedWarnings = append(collectedWarnings, warnings...)
	if err != nil {
		return "", collectedWarnings, err
	}

	if len(deployments) < 1 {
		return "", collectedWarnings, errors.New("failed to find a deployment for that app")
	}

	return deployments[0].GUID, collectedWarnings, nil
}
