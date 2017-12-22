package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type ServiceInstanceSummary struct {
	ServiceInstance

	ServicePlan               ServicePlan
	Service                   Service
	ServiceInstanceSharedFrom ServiceInstanceSharedFrom
	BoundApplications         []string
}

func (actor Actor) GetServiceInstanceSummaryByNameAndSpace(name string, spaceGUID string) (ServiceInstanceSummary, Warnings, error) {
	serviceInstanceSummary := ServiceInstanceSummary{}

	serviceInstance, instanceWarnings, instanceErr := actor.GetServiceInstanceByNameAndSpace(name, spaceGUID)
	allWarnings := Warnings(instanceWarnings)
	if instanceErr != nil {
		return serviceInstanceSummary, allWarnings, instanceErr
	}
	serviceInstanceSummary.ServiceInstance = serviceInstance

	var (
		serviceBindings  []ServiceBinding
		bindingsWarnings Warnings
		bindingsErr      error
	)

	if ccv2.ServiceInstance(serviceInstance).Managed() {
		serviceInstanceSharedFrom, sharedFromWarnings, shareFromErr := actor.GetServiceInstanceSharedFromByServiceInstance(serviceInstance.GUID)
		allWarnings = append(allWarnings, sharedFromWarnings...)
		if shareFromErr != nil {
			if _, ok := shareFromErr.(ccerror.ResourceNotFoundError); ok {
				serviceInstanceSharedFrom = ServiceInstanceSharedFrom{}
			} else {
				return serviceInstanceSummary, allWarnings, shareFromErr
			}
		}
		serviceInstanceSummary.ServiceInstanceSharedFrom = serviceInstanceSharedFrom

		servicePlan, planWarnings, planErr := actor.GetServicePlan(serviceInstance.ServicePlanGUID)
		allWarnings = append(allWarnings, planWarnings...)
		if planErr != nil {
			return serviceInstanceSummary, allWarnings, planErr
		}
		serviceInstanceSummary.ServicePlan = servicePlan

		service, serviceWarnings, serviceErr := actor.GetService(servicePlan.ServiceGUID)
		allWarnings = append(allWarnings, serviceWarnings...)
		if serviceErr != nil {
			return serviceInstanceSummary, allWarnings, serviceErr
		}
		serviceInstanceSummary.Service = service

		serviceBindings, bindingsWarnings, bindingsErr = actor.GetServiceBindingsByServiceInstance(serviceInstance.GUID)
	} else {
		serviceBindings, bindingsWarnings, bindingsErr = actor.GetServiceBindingsByUserProvidedServiceInstance(serviceInstance.GUID)
	}

	allWarnings = append(allWarnings, bindingsWarnings...)
	if bindingsErr != nil {
		return serviceInstanceSummary, allWarnings, bindingsErr
	}

	for _, serviceBinding := range serviceBindings {
		app, appWarnings, appErr := actor.GetApplication(serviceBinding.AppGUID)
		allWarnings = append(allWarnings, appWarnings...)
		if appErr != nil {
			return serviceInstanceSummary, allWarnings, appErr
		}
		serviceInstanceSummary.BoundApplications = append(serviceInstanceSummary.BoundApplications, app.Name)
	}

	return serviceInstanceSummary, allWarnings, nil
}
