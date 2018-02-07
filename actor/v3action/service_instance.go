package v3action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServiceInstance ccv3.ServiceInstance

func (actor Actor) UnshareServiceInstanceFromSpace(serviceInstanceName string, sourceSpaceGUID string, sharedToSpaceGUID string) (Warnings, error) {
	serviceInstance, allWarnings, err := actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, sourceSpaceGUID)

	if _, ok := err.(actionerror.ServiceInstanceNotFoundError); ok == true {
		return allWarnings, actionerror.SharedServiceInstanceNotFoundError{}
	}

	if err != nil {
		return allWarnings, err
	}

	apiWarnings, err := actor.CloudControllerClient.UnshareServiceInstanceFromSpace(serviceInstance.GUID, sharedToSpaceGUID)
	allWarnings = append(allWarnings, apiWarnings...)

	return allWarnings, err
}

func (actor Actor) GetServiceInstanceByNameAndSpace(serviceInstanceName string, spaceGUID string) (ServiceInstance, Warnings, error) {
	serviceInstances, warnings, err := actor.CloudControllerClient.GetServiceInstances(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceInstanceName}},
		ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
	)

	if err != nil {
		return ServiceInstance{}, Warnings(warnings), err
	}

	if len(serviceInstances) == 0 {
		return ServiceInstance{}, Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	}

	//Handle multiple serviceInstances being returned as GetServiceInstances arnt filtered by space
	return ServiceInstance(serviceInstances[0]), Warnings(warnings), nil
}
