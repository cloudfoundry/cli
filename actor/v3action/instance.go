package v3action

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Instance represents a V3 actor instance.
type Instance ccv3.Instance

// StartTime returns the time that the instance started.
func (instance *Instance) StartTime() time.Time {
	uptimeDuration := time.Duration(instance.Uptime) * time.Second

	return time.Now().Add(-uptimeDuration)
}

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
