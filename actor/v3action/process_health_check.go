package v3action

import (
	"sort"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type ProcessHealthCheck struct {
	ProcessType     string
	HealthCheckType string
	Endpoint        string
}

type ProcessHealthChecks []ProcessHealthCheck

func (phs ProcessHealthChecks) Sort() {
	sort.Slice(phs, func(i int, j int) bool {
		var iScore int
		var jScore int

		switch phs[i].ProcessType {
		case constant.ProcessTypeWeb:
			iScore = 0
		default:
			iScore = 1
		}

		switch phs[j].ProcessType {
		case constant.ProcessTypeWeb:
			jScore = 0
		default:
			jScore = 1
		}

		if iScore == 1 && jScore == 1 {
			return phs[i].ProcessType < phs[j].ProcessType
		}
		return iScore < jScore
	})
}

func (actor Actor) GetApplicationProcessHealthChecksByNameAndSpace(appName string, spaceGUID string) ([]ProcessHealthCheck, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	ccv3Processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(app.GUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return nil, allWarnings, err
	}

	var processHealthChecks ProcessHealthChecks
	for _, ccv3Process := range ccv3Processes {
		processHealthCheck := ProcessHealthCheck{
			ProcessType:     ccv3Process.Type,
			HealthCheckType: ccv3Process.HealthCheck.Type,
			Endpoint:        ccv3Process.HealthCheck.Data.Endpoint,
		}
		processHealthChecks = append(processHealthChecks, processHealthCheck)
	}

	processHealthChecks.Sort()

	return processHealthChecks, allWarnings, nil
}

func (actor Actor) SetApplicationProcessHealthCheckTypeByNameAndSpace(appName string, spaceGUID string, healthCheckType string, httpEndpoint string, processType string) (Application, Warnings, error) {
	if healthCheckType != "http" {
		if httpEndpoint == "/" {
			httpEndpoint = ""
		} else {
			return Application{}, nil, actionerror.HTTPHealthCheckInvalidError{}
		}
	}

	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return Application{}, allWarnings, err
	}

	process, warnings, err := actor.CloudControllerClient.GetApplicationProcessByType(app.GUID, processType)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		if _, ok := err.(ccerror.ProcessNotFoundError); ok {
			return Application{}, allWarnings, actionerror.ProcessNotFoundError{ProcessType: processType}
		}
		return Application{}, allWarnings, err
	}

	_, warnings, err = actor.CloudControllerClient.PatchApplicationProcessHealthCheck(
		process.GUID,
		healthCheckType,
		httpEndpoint,
	)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return Application{}, allWarnings, err
	}

	return app, allWarnings, nil
}
