package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

type ApplicationInstanceState constant.ApplicationInstanceState

type ApplicationInstance ccv2.ApplicationInstance

func (instance ApplicationInstance) Crashed() bool {
	return instance.State == constant.ApplicationInstanceCrashed
}

func (instance ApplicationInstance) Flapping() bool {
	return instance.State == constant.ApplicationInstanceFlapping
}

func (instance ApplicationInstance) Running() bool {
	return instance.State == constant.ApplicationInstanceRunning
}

func (actor Actor) GetApplicationInstancesByApplication(guid string) (map[int]ApplicationInstance, Warnings, error) {
	ccAppInstances, warnings, err := actor.CloudControllerClient.GetApplicationApplicationInstances(guid)

	switch err.(type) {
	case ccerror.ResourceNotFoundError, ccerror.NotStagedError, ccerror.InstancesError:
		return nil, Warnings(warnings), actionerror.ApplicationInstancesNotFoundError{ApplicationGUID: guid}
	}

	appInstances := map[int]ApplicationInstance{}

	for id, applicationInstance := range ccAppInstances {
		appInstances[id] = ApplicationInstance(applicationInstance)
	}

	return appInstances, Warnings(warnings), err
}
