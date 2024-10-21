package v2action

import (
	"code.cloudfoundry.org/cli/v7/actor/actionerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv2/constant"
)

// ServiceInstance represents an instance of a service.
type ServiceInstance ccv2.ServiceInstance
type MaintenanceInfo ccv2.MaintenanceInfo

// CreateServiceInstance creates a new service instance with the provided attributes.
func (actor Actor) CreateServiceInstance(spaceGUID, serviceName, servicePlanName, serviceInstanceName, brokerName string, params map[string]interface{}, tags []string) (ServiceInstance, Warnings, error) {
	var allWarnings Warnings
	plan, allWarnings, err := actor.getServicePlanForServiceInSpace(servicePlanName, serviceName, spaceGUID, brokerName)

	if err != nil {
		return ServiceInstance{}, allWarnings, err
	}

	instance, warnings, err := actor.CloudControllerClient.CreateServiceInstance(spaceGUID, plan.GUID, serviceInstanceName, params, tags)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ServiceInstance{}, allWarnings, err
	}

	return ServiceInstance(instance), allWarnings, nil
}

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
		ccv2.Filter{
			Type:     constant.NameFilter,
			Operator: constant.EqualOperator,
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
	bindings, apiWarnings, err := actor.CloudControllerClient.GetServiceBindings(ccv2.Filter{
		Type:     constant.AppGUIDFilter,
		Operator: constant.EqualOperator,
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

// UpdateServiceInstanceMaintenanceInfo requests that the service instance be updated to the specified `maintenance_info`
func (actor Actor) UpdateServiceInstanceMaintenanceInfo(guid string, maintenanceInfo MaintenanceInfo) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.UpdateServiceInstanceMaintenanceInfo(guid, ccv2.MaintenanceInfo(maintenanceInfo))
	return Warnings(warnings), err
}

// IsManaged returns true if the service instance is managed, otherwise false.
func (instance ServiceInstance) IsManaged() bool {
	return ccv2.ServiceInstance(instance).Managed()
}

// IsUserProvided returns true if the service instance is user provided, otherwise false.
func (instance ServiceInstance) IsUserProvided() bool {
	return ccv2.ServiceInstance(instance).UserProvided()
}
