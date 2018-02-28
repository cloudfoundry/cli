package v2action

import (
	"sort"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/util/sorting"
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

func (actor Actor) GetServiceInstancesSummaryBySpace(spaceGUID string) ([]ServiceInstanceSummary, Warnings, error) {
	serviceInstances, warnings, err := actor.CloudControllerClient.GetSpaceServiceInstances(
		spaceGUID,
		true)
	allWarnings := Warnings(warnings)

	var summaryInstances []ServiceInstanceSummary
	for _, instance := range serviceInstances {
		serviceInstanceSummary, summaryInfoWarnings, summaryInfoErr := actor.getSummaryInfoCompositeForInstance(
			spaceGUID,
			ServiceInstance(instance),
			false)
		allWarnings = append(allWarnings, summaryInfoWarnings...)
		if summaryInfoErr != nil {
			return nil, allWarnings, summaryInfoErr
		}
		summaryInstances = append(summaryInstances, serviceInstanceSummary)
	}

	sort.Slice(summaryInstances, func(i, j int) bool {
		return sorting.LessIgnoreCase(summaryInstances[i].Name, summaryInstances[j].Name)
	})

	return summaryInstances, allWarnings, err
}

func (actor Actor) GetServiceInstanceSummaryByNameAndSpace(name string, spaceGUID string) (ServiceInstanceSummary, Warnings, error) {
	serviceInstance, instanceWarnings, err := actor.GetServiceInstanceByNameAndSpace(name, spaceGUID)
	allWarnings := Warnings(instanceWarnings)
	if err != nil {
		return ServiceInstanceSummary{}, allWarnings, err
	}

	serviceInstanceSummary, warnings, err := actor.getSummaryInfoCompositeForInstance(spaceGUID, serviceInstance, true)
	return serviceInstanceSummary, append(allWarnings, warnings...), err
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

func (actor Actor) getSummaryInfoCompositeForInstance(spaceGUID string, serviceInstance ServiceInstance,
	retrieveSharedInfo bool) (ServiceInstanceSummary, Warnings, error) {
	serviceInstanceSummary := ServiceInstanceSummary{ServiceInstance: serviceInstance}
	var (
		serviceBindings []ServiceBinding
		allWarnings     Warnings
	)

	if serviceInstance.IsManaged() {
		if retrieveSharedInfo {
			sharedWarnings, err := actor.getAndSetSharedInformation(&serviceInstanceSummary, spaceGUID)
			allWarnings = Warnings(sharedWarnings)
			if err != nil {
				return serviceInstanceSummary, allWarnings, err
			}
		}

		servicePlan, planWarnings, err := actor.GetServicePlan(serviceInstance.ServicePlanGUID)
		allWarnings = append(allWarnings, planWarnings...)
		if err != nil {
			return serviceInstanceSummary, allWarnings, err
		}
		serviceInstanceSummary.ServicePlan = servicePlan

		service, serviceWarnings, err := actor.GetService(servicePlan.ServiceGUID)
		allWarnings = append(allWarnings, serviceWarnings...)
		if err != nil {
			return serviceInstanceSummary, allWarnings, err
		}
		serviceInstanceSummary.Service = service

		var bindingsWarnings Warnings
		serviceBindings, bindingsWarnings, err = actor.GetServiceBindingsByServiceInstance(serviceInstance.GUID)
		allWarnings = append(allWarnings, bindingsWarnings...)
		if err != nil {
			return serviceInstanceSummary, allWarnings, err
		}
	} else {
		var bindingsWarnings Warnings
		var err error
		serviceBindings, bindingsWarnings, err = actor.GetServiceBindingsByUserProvidedServiceInstance(serviceInstance.GUID)
		allWarnings = append(allWarnings, bindingsWarnings...)
		if err != nil {
			return serviceInstanceSummary, allWarnings, err
		}
	}

	for _, serviceBinding := range serviceBindings {
		app, appWarnings, err := actor.GetApplication(serviceBinding.AppGUID)
		allWarnings = append(allWarnings, appWarnings...)
		if err != nil {
			return serviceInstanceSummary, allWarnings, err
		}
		serviceInstanceSummary.BoundApplications = append(serviceInstanceSummary.BoundApplications, app.Name)
	}

	sort.Slice(serviceInstanceSummary.BoundApplications,
		sorting.SortAlphabeticFunc(serviceInstanceSummary.BoundApplications))

	return serviceInstanceSummary, allWarnings, nil
}
