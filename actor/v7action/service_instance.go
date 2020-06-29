package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

func (actor Actor) GetServiceInstanceByNameAndSpace(serviceInstanceName string, spaceGUID string) (resources.ServiceInstance, Warnings, error) {
	serviceInstances, warnings, err := actor.CloudControllerClient.GetServiceInstances(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceInstanceName}},
		ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
	)

	if err != nil {
		return resources.ServiceInstance{}, Warnings(warnings), err
	}

	if len(serviceInstances) == 0 {
		return resources.ServiceInstance{}, Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	}

	//Handle multiple serviceInstances being returned as GetServiceInstances arnt filtered by space
	return serviceInstances[0], Warnings(warnings), nil
}

func (actor Actor) UnshareServiceInstanceByServiceInstanceAndSpace(serviceInstanceGUID string, sharedToSpaceGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.DeleteServiceInstanceRelationshipsSharedSpace(serviceInstanceGUID, sharedToSpaceGUID)
	return Warnings(warnings), err
}

func (actor Actor) CreateUserProvidedServiceInstance(serviceInstance resources.ServiceInstance) (Warnings, error) {
	serviceInstance.Type = resources.UserProvidedServiceInstance
	_, warnings, err := actor.CloudControllerClient.CreateServiceInstance(serviceInstance)
	return Warnings(warnings), err
}

func (actor Actor) UpdateUserProvidedServiceInstance(serviceInstanceName, spaceGUID string, serviceInstanceUpdates resources.ServiceInstance) (Warnings, error) {
	original, warnings, err := actor.CloudControllerClient.GetServiceInstanceByNameAndSpaceGUID(serviceInstanceName, spaceGUID)
	if err != nil {
		return Warnings(warnings), err
	}

	if original.Type != resources.UserProvidedServiceInstance {
		return Warnings(warnings), actionerror.ServiceInstanceTypeError{
			Name:         serviceInstanceName,
			RequiredType: resources.UserProvidedServiceInstance,
		}
	}

	_, updateWarnings, err := actor.CloudControllerClient.UpdateServiceInstance(original.GUID, serviceInstanceUpdates)
	warnings = append(warnings, updateWarnings...)
	if err != nil {
		return Warnings(warnings), err
	}

	return Warnings(warnings), nil
}
