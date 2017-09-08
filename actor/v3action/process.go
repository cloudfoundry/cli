package v3action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Process represents a V3 actor process.
type Process ccv3.Process

// ProcessNotFoundError is returned when the proccess type cannot be found
type ProcessNotFoundError struct {
	ProcessType string
}

func (e ProcessNotFoundError) Error() string {
	return fmt.Sprintf("Process %s not found", e.ProcessType)
}

func (actor Actor) ScaleProcessByApplication(appGUID string, process Process) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.CreateApplicationProcessScale(appGUID, ccv3.Process(process))
	allWarnings := Warnings(warnings)
	if err != nil {
		if _, ok := err.(ccerror.ProcessNotFoundError); ok {
			return allWarnings, ProcessNotFoundError{ProcessType: process.Type}
		}
		return allWarnings, err
	}

	return allWarnings, nil
}
