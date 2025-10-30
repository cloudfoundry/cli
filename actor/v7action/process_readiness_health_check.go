package v7action

import (
	"sort"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
)

type ProcessReadinessHealthCheck struct {
	ProcessType       string
	HealthCheckType   constant.HealthCheckType
	Endpoint          string
	InvocationTimeout int64
	Interval          int64
}

type ProcessReadinessHealthChecks []ProcessReadinessHealthCheck

func (phs ProcessReadinessHealthChecks) Sort() {
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

func (actor Actor) GetApplicationProcessReadinessHealthChecksByNameAndSpace(appName string, spaceGUID string) ([]ProcessReadinessHealthCheck, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	ccv3Processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(app.GUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return nil, allWarnings, err
	}

	var processReadinessHealthChecks ProcessReadinessHealthChecks
	for _, ccv3Process := range ccv3Processes {
		processReadinessHealthCheck := ProcessReadinessHealthCheck{
			ProcessType:       ccv3Process.Type,
			HealthCheckType:   ccv3Process.ReadinessHealthCheckType,
			Endpoint:          ccv3Process.ReadinessHealthCheckEndpoint,
			InvocationTimeout: ccv3Process.ReadinessHealthCheckInvocationTimeout,
			Interval:          ccv3Process.ReadinessHealthCheckInterval,
		}
		processReadinessHealthChecks = append(processReadinessHealthChecks, processReadinessHealthCheck)
	}

	processReadinessHealthChecks.Sort()

	return processReadinessHealthChecks, allWarnings, nil
}
