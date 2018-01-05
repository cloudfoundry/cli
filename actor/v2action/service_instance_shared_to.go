package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

type ServiceInstanceSharedTo ccv2.ServiceInstanceSharedTo

func (actor Actor) GetServiceInstanceSharedTosByServiceInstance(serviceInstanceGUID string) ([]ServiceInstanceSharedTo, Warnings, error) {
	sharedTos, warnings, err := actor.CloudControllerClient.GetServiceInstanceSharedTos(serviceInstanceGUID)
	if err != nil {
		return nil, Warnings(warnings), err
	}

	var serviceInstanceSharedTos = make([]ServiceInstanceSharedTo, len(sharedTos))
	for i, sharedTo := range sharedTos {
		serviceInstanceSharedTos[i] = ServiceInstanceSharedTo(sharedTo)
	}

	return serviceInstanceSharedTos, Warnings(warnings), nil
}
