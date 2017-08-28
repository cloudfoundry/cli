package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// ServiceInstance represents an instance of a service.
type ServiceInstance ccv2.ServiceInstance

type ServiceInstanceNotFoundError struct {
	GUID string
	Name string
}

func (e ServiceInstanceNotFoundError) Error() string {
	if e.Name == "" {
		return fmt.Sprintf("Service instance (GUID: %s) not found.", e.GUID)
	}
	return fmt.Sprintf("Service instance '%s' not found.", e.Name)
}

func (actor Actor) GetServiceInstance(guid string) (ServiceInstance, Warnings, error) {
	instance, warnings, err := actor.CloudControllerClient.GetServiceInstance(guid)
	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return ServiceInstance{}, Warnings(warnings), ServiceInstanceNotFoundError{GUID: guid}
	}
	return ServiceInstance(instance), Warnings(warnings), err
}

func (actor Actor) GetServiceInstanceByNameAndSpace(name string, spaceGUID string) (ServiceInstance, Warnings, error) {
	serviceInstances, warnings, err := actor.CloudControllerClient.GetSpaceServiceInstances(
		spaceGUID,
		true,
		ccv2.Query{
			Filter:   ccv2.NameFilter,
			Operator: ccv2.EqualOperator,
			Values:   []string{name},
		})

	if err != nil {
		return ServiceInstance{}, Warnings(warnings), err
	}

	if len(serviceInstances) == 0 {
		return ServiceInstance{}, Warnings(warnings), ServiceInstanceNotFoundError{
			Name: name,
		}
	}

	return ServiceInstance(serviceInstances[0]), Warnings(warnings), nil
}

func (actor Actor) GetServiceInstancesByApplication(appGUID string) ([]ServiceInstance, Warnings, error) {
	var allWarnings Warnings
	bindings, apiWarnings, err := actor.CloudControllerClient.GetServiceBindings(ccv2.Query{
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
