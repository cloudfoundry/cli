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

	ServicePlan                       ServicePlan
	Service                           Service
	ServiceInstanceSharingFeatureFlag bool
	ServiceInstanceShareType          ServiceInstanceShareType
	ServiceInstanceSharedFrom         ServiceInstanceSharedFrom
	ServiceInstanceSharedTos          []ServiceInstanceSharedTo
	BoundApplications                 []string
}

func (s ServiceInstanceSummary) IsShareable() bool {
	return s.ServiceInstanceSharingFeatureFlag && s.Service.Extra.Shareable
}

func (s ServiceInstanceSummary) IsNotShared() bool {
	return s.ServiceInstanceShareType == ServiceInstanceIsNotShared
}

func (s ServiceInstanceSummary) IsSharedFrom() bool {
	return s.ServiceInstanceShareType == ServiceInstanceIsSharedFrom
}

func (s ServiceInstanceSummary) IsSharedTo() bool {
	return s.ServiceInstanceShareType == ServiceInstanceIsSharedTo
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

	if serviceInstance.IsManaged() {
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

	// Part of determining if a service instance is shareable, we need to find
	// out if the service_instance_sharing feature flag is enabled
	featureFlags, featureFlagsWarnings, featureFlagsErr := actor.CloudControllerClient.GetConfigFeatureFlags()
	allWarnings := Warnings(featureFlagsWarnings)
	if featureFlagsErr != nil {
		return allWarnings, featureFlagsErr
	}

	for _, flag := range featureFlags {
		if flag.Name == string(ccv2.FeatureFlagServiceInstanceSharing) {
			summary.ServiceInstanceSharingFeatureFlag = flag.Enabled
		}
	}

	// Service instance is shared from if:
	// 1. the source space of the service instance is empty (API returns json null)
	// 2. the targeted space is not the same as the source space of the service instance AND
	//    we call the shared_from url and it returns a non-empty resource
	if summary.ServiceInstance.SpaceGUID == "" || summary.ServiceInstance.SpaceGUID != spaceGUID {
		summary.ServiceInstanceSharedFrom, warnings, err = actor.GetServiceInstanceSharedFromByServiceInstance(summary.ServiceInstance.GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			// if the API version does not support service instance sharing, ignore the 404
			if _, ok := err.(ccerror.ResourceNotFoundError); !ok {
				return allWarnings, err
			}
		}

		if summary.ServiceInstanceSharedFrom.SpaceGUID != "" {
			summary.ServiceInstanceShareType = ServiceInstanceIsSharedFrom
		} else {
			summary.ServiceInstanceShareType = ServiceInstanceIsNotShared
		}

		return allWarnings, nil
	}

	// Service instance is shared to if:
	// the targeted space is the same as the source space of the service instance AND
	// we call the shared_to url and get a non-empty list
	summary.ServiceInstanceSharedTos, warnings, err = actor.GetServiceInstanceSharedTosByServiceInstance(summary.ServiceInstance.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		// if the API version does not support service instance sharing, ignore the 404
		if _, ok := err.(ccerror.ResourceNotFoundError); !ok {
			return allWarnings, err
		}
	}

	if len(summary.ServiceInstanceSharedTos) > 0 {
		summary.ServiceInstanceShareType = ServiceInstanceIsSharedTo
	} else {
		summary.ServiceInstanceShareType = ServiceInstanceIsNotShared
	}

	return allWarnings, nil
}
