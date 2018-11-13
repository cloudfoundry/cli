package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Process represents a V3 actor process.
type Process ccv3.Process

// GetProcessByTypeAndApplication returns a process for the given application
// and type.
func (actor Actor) GetProcessByTypeAndApplication(processType string, appGUID string) (Process, Warnings, error) {
	process, warnings, err := actor.CloudControllerClient.GetApplicationProcessByType(appGUID, processType)
	if _, ok := err.(ccerror.ProcessNotFoundError); ok {
		return Process{}, Warnings(warnings), actionerror.ProcessNotFoundError{ProcessType: processType}
	}
	return Process(process), Warnings(warnings), err
}

func (actor Actor) ScaleProcessByApplication(appGUID string, process Process) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.CreateApplicationProcessScale(appGUID, ccv3.Process(process))
	allWarnings := Warnings(warnings)
	if err != nil {
		if _, ok := err.(ccerror.ProcessNotFoundError); ok {
			return allWarnings, actionerror.ProcessNotFoundError{ProcessType: process.Type}
		}
		return allWarnings, err
	}

	return allWarnings, nil
}
