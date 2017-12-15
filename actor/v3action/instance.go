package v3action

import (
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// Instance represents a V3 process instance.
type Instance ccv3.ProcessInstance

// Running will return true if the instance is running.
func (instance Instance) Running() bool {
	return instance.State == constant.ProcessInstanceRunning
}

// StartTime returns the time that the instance started.
func (instance *Instance) StartTime() time.Time {
	uptimeDuration := time.Duration(instance.Uptime) * time.Second

	return time.Now().Add(-uptimeDuration)
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
			return allWarnings, actionerror.ProcessNotFoundError{
				ProcessType: processType,
			}
		case ccerror.InstanceNotFoundError:
			return allWarnings, actionerror.ProcessInstanceNotFoundError{
				ProcessType:   processType,
				InstanceIndex: uint(instanceIndex),
			}
		default:
			return allWarnings, err
		}
	}

	return allWarnings, nil
}
