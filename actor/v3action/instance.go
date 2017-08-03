package v3action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

// ProcessInstanceNotFoundError is returned when the proccess type or process instance cannot be found
type ProcessInstanceNotFoundError struct {
	ProcessType   string
	InstanceIndex int
}

func (e ProcessInstanceNotFoundError) Error() string {
	return fmt.Sprintf("Instance %d for process %s not found", e.InstanceIndex, e.ProcessType)
}

func (actor Actor) DeleteInstanceByApplicationNameSpaceProcessTypeAndIndex(appName string, spaceGUID string, processType string, instanceIndex int) (Warnings, error) {
	var allWarnings Warnings
	app, appWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, appWarnings...)
	if err != nil {
		return allWarnings, err
	}

	deleteWarnings, err := actor.CloudControllerClient.DeleteApplicationProcessInstance(app.GUID, processType, instanceIndex)
	allWarnings = append(allWarnings, deleteWarnings...)

	if err != nil {
		switch err.(type) {
		case ccerror.ProcessNotFoundError:
			return allWarnings, ProcessNotFoundError{
				ProcessType: processType,
			}
		case ccerror.InstanceNotFoundError:
			return allWarnings, ProcessInstanceNotFoundError{
				ProcessType:   processType,
				InstanceIndex: instanceIndex,
			}
		default:
			return allWarnings, err
		}
	}

	return allWarnings, nil
}
