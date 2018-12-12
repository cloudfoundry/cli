package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type ProcessInstance ccv3.ProcessInstance

// Running will return true if the instance is running.
func (instance ProcessInstance) Running() bool {
	return instance.State == constant.ProcessInstanceRunning
}

// StartTime returns the time that the instance started.
func (instance *ProcessInstance) StartTime() time.Time {
	uptimeDuration := time.Duration(instance.Uptime) * time.Second

	return time.Now().Add(-uptimeDuration)
}

type ProcessInstances []ccv3.ProcessInstance

func (pi ProcessInstances) AllCrashed() bool {
	for _, instance := range pi {
		if instance.State != constant.ProcessInstanceCrashed {
			return false
		}
	}
	return true
}

func (pi ProcessInstances) AnyRunning() bool {
	for _, instance := range pi {
		if instance.State == constant.ProcessInstanceRunning {
			return true
		}
	}
	return false
}

func (pi ProcessInstances) Empty() bool {
	return len(pi) == 0
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
