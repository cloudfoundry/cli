package v7action

import (
	"sort"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type ProcessHealthCheck struct {
	ProcessType       string
	HealthCheckType   string
	Endpoint          string
	InvocationTimeout int
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
			ProcessType:       ccv3Process.Type,
			HealthCheckType:   ccv3Process.HealthCheckType,
			Endpoint:          ccv3Process.HealthCheckEndpoint,
			InvocationTimeout: ccv3Process.HealthCheckInvocationTimeout,
		}
		processHealthChecks = append(processHealthChecks, processHealthCheck)
	}

	processHealthChecks.Sort()

	return processHealthChecks, allWarnings, nil
}
