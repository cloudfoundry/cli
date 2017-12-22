package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

type ServiceInstanceSharedFrom ccv2.ServiceInstanceSharedFrom

func (actor Actor) GetServiceInstanceSharedFromByServiceInstance(serviceInstanceGUID string) (ServiceInstanceSharedFrom, Warnings, error) {
	serviceInstanceSharedFrom, warnings, err := actor.CloudControllerClient.GetServiceInstanceSharedFrom(serviceInstanceGUID)
	return ServiceInstanceSharedFrom(serviceInstanceSharedFrom), Warnings(warnings), err
}
