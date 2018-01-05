package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type ServiceInstanceShareType string

const (
	ServiceInstanceIsSharedFrom ServiceInstanceShareType = "SharedFrom"
	ServiceInstanceIsSharedTo   ServiceInstanceShareType = "SharedTo"
	ServiceInstanceIsNotShared  ServiceInstanceShareType = "NotShared"
)

type ServiceInstanceSummary struct {
	ServiceInstance

	ServicePlan               ServicePlan
	Service                   Service
	ServiceInstanceShareType  ServiceInstanceShareType
	ServiceInstanceSharedFrom ServiceInstanceSharedFrom
	ServiceInstanceSharedTos  []ServiceInstanceSharedTo
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
		sharedWarnings, sharedErr := actor.getAndSetSharedInformation(&serviceInstanceSummary, spaceGUID)
		allWarnings = append(allWarnings, sharedWarnings...)
		if sharedErr != nil {
			return serviceInstanceSummary, allWarnings, sharedErr
		}

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

// getAndSetSharedInformation gets a service instance's shared from or shared to information,
func (actor Actor) getAndSetSharedInformation(summary *ServiceInstanceSummary, spaceGUID string) (Warnings, error) {
	var (
		warnings Warnings
		err      error
	)

	// Service instance is shared from if:
	// 1. the source space of the service instance is empty (API returns json null)
	// 2. the targeted space is not the same as the source space of the service instance AND
	//    we call the shared_from url and it returns a non-empty resource
	if summary.ServiceInstance.SpaceGUID == "" || summary.ServiceInstance.SpaceGUID != spaceGUID {
		summary.ServiceInstanceSharedFrom, warnings, err = actor.GetServiceInstanceSharedFromByServiceInstance(summary.ServiceInstance.GUID)
		if err != nil {
			// if the API version does not support service instance sharing, ignore the 404
			if _, ok := err.(ccerror.ResourceNotFoundError); !ok {
				return warnings, err
			}
		}

		if summary.ServiceInstanceSharedFrom.SpaceGUID != "" {
			summary.ServiceInstanceShareType = ServiceInstanceIsSharedFrom
		} else {
			summary.ServiceInstanceShareType = ServiceInstanceIsNotShared
		}

		return warnings, nil
	}

	// Service instance is shared to if:
	// the targeted space is the same as the source space of the service instance AND
	// we call the shared_to url and get a non-empty list
	summary.ServiceInstanceSharedTos, warnings, err = actor.GetServiceInstanceSharedTosByServiceInstance(summary.ServiceInstance.GUID)
	if err != nil {
		// if the API version does not support service instance sharing, ignore the 404
		if _, ok := err.(ccerror.ResourceNotFoundError); !ok {
			return warnings, err
		}
	}

	if len(summary.ServiceInstanceSharedTos) > 0 {
		summary.ServiceInstanceShareType = ServiceInstanceIsSharedTo
	} else {
		summary.ServiceInstanceShareType = ServiceInstanceIsNotShared
	}

	return warnings, nil
}
