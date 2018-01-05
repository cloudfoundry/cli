package v2action

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// ServiceInstance represents an instance of a service.
type ServiceInstance ccv2.ServiceInstance

func (actor Actor) GetServiceInstance(guid string) (ServiceInstance, Warnings, error) {
	instance, warnings, err := actor.CloudControllerClient.GetServiceInstance(guid)
	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return ServiceInstance{}, Warnings(warnings), actionerror.ServiceInstanceNotFoundError{GUID: guid}
	}
	return ServiceInstance(instance), Warnings(warnings), err
}

func (actor Actor) GetServiceInstanceByNameAndSpace(name string, spaceGUID string) (ServiceInstance, Warnings, error) {
	serviceInstances, warnings, err := actor.CloudControllerClient.GetSpaceServiceInstances(
		spaceGUID,
		true,
		ccv2.QQuery{
			Filter:   ccv2.NameFilter,
			Operator: ccv2.EqualOperator,
			Values:   []string{name},
		})

	if err != nil {
		return ServiceInstance{}, Warnings(warnings), err
	}

	if len(serviceInstances) == 0 {
		return ServiceInstance{}, Warnings(warnings), actionerror.ServiceInstanceNotFoundError{
			Name: name,
		}
	}

	return ServiceInstance(serviceInstances[0]), Warnings(warnings), nil
}

func (actor Actor) GetServiceInstancesByApplication(appGUID string) ([]ServiceInstance, Warnings, error) {
	var allWarnings Warnings
	bindings, apiWarnings, err := actor.CloudControllerClient.GetServiceBindings(ccv2.QQuery{
		Filter:   ccv2.AppGUIDFilter,
		Operator: ccv2.EqualOperator,
		Values:   []string{appGUID},
	})
	allWarnings = append(allWarnings, apiWarnings...)

	if err != nil {
		return nil, allWarnings, err
	}

	var serviceInstances []ServiceInstance
	for _, binding := range bindings {
		instance, warnings, instanceErr := actor.GetServiceInstance(binding.ServiceInstanceGUID)
		allWarnings = append(allWarnings, warnings...)
		if instanceErr != nil {
			return nil, allWarnings, instanceErr
		}
		serviceInstances = append(serviceInstances, ServiceInstance(instance))
	}

	return serviceInstances, allWarnings, err
}

func (actor Actor) GetServiceInstancesBySpace(spaceGUID string) ([]ServiceInstance, Warnings, error) {
	ccv2ServiceInstances, warnings, err := actor.CloudControllerClient.GetSpaceServiceInstances(spaceGUID, true)

	if err != nil {
		return []ServiceInstance{}, Warnings(warnings), err
	}

	serviceInstances := make([]ServiceInstance, len(ccv2ServiceInstances))
	for i, ccv2ServiceInstance := range ccv2ServiceInstances {
		serviceInstances[i] = ServiceInstance(ccv2ServiceInstance)
	}

	return serviceInstances, Warnings(warnings), nil
}

func (actor Actor) GetSharedToSpaceGUID(serviceInstanceName string, sourceSpaceGUID string, sharedToOrgName string, sharedToSpaceName string) (string, Warnings, error) {
	serviceInstance, allWarnings, err := actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, sourceSpaceGUID)

	if err != nil {
		return "", allWarnings, err
	}

	sharedTos, warnings, err := actor.GetServiceInstanceSharedTosByServiceInstance(serviceInstance.GUID)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return "", allWarnings, err
	}

	for _, sharedTo := range sharedTos {
		if strings.EqualFold(sharedTo.SpaceName, sharedToSpaceName) && strings.EqualFold(sharedTo.OrganizationName, sharedToOrgName) {
			return sharedTo.SpaceGUID, allWarnings, nil
		}
	}

	return "", allWarnings, actionerror.ServiceInstanceNotSharedToSpaceError{ServiceInstanceName: serviceInstanceName}
}
