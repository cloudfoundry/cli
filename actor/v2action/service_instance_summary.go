package v2action

import (
	"fmt"
	"sort"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/util/sorting"
	log "github.com/sirupsen/logrus"
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
	BoundApplications                 []BoundApplication
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

type BoundApplication struct {
	AppName            string
	ServiceBindingName string
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

func (actor Actor) GetServiceInstancesSummaryBySpace(spaceGUID string) ([]ServiceInstanceSummary, Warnings, error) {
	serviceInstances, warnings, err := actor.CloudControllerClient.GetSpaceServiceInstances(
		spaceGUID,
		true)
	allWarnings := Warnings(warnings)

	log.WithField("number_of_service_instances", len(serviceInstances)).Info("listing number of service instances")

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
		if flag.Name == string(constant.FeatureFlagServiceInstanceSharing) {
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

func (actor Actor) getSummaryInfoCompositeForInstance(spaceGUID string, serviceInstance ServiceInstance, retrieveSharedInfo bool) (ServiceInstanceSummary, Warnings, error) {
	log.WithField("GUID", serviceInstance.GUID).Info("looking up service instance info")

	serviceInstanceSummary := ServiceInstanceSummary{ServiceInstance: serviceInstance}
	var (
		serviceBindings []ServiceBinding
		allWarnings     Warnings
	)

	if serviceInstance.IsManaged() {
		log.Debug("service is managed")
		if retrieveSharedInfo {
			sharedWarnings, err := actor.getAndSetSharedInformation(&serviceInstanceSummary, spaceGUID)
			allWarnings = Warnings(sharedWarnings)
			if err != nil {
				log.WithField("GUID", serviceInstance.GUID).Errorln("looking up share info:", err)
				return serviceInstanceSummary, allWarnings, err
			}
		}

		servicePlan, planWarnings, err := actor.GetServicePlan(serviceInstance.ServicePlanGUID)
		allWarnings = append(allWarnings, planWarnings...)
		if err != nil {
			log.WithField("service_plan_guid", serviceInstance.ServicePlanGUID).Errorln("looking up service plan:", err)
			if _, ok := err.(ccerror.ForbiddenError); !ok {
				return serviceInstanceSummary, allWarnings, err
			}
			log.Warning("Forbidden Error - ignoring and continue")
			allWarnings = append(allWarnings, fmt.Sprintf("This org is not authorized to view necessary data about this service plan. Contact your administrator regarding service GUID %s.", serviceInstance.ServicePlanGUID))
		}
		serviceInstanceSummary.ServicePlan = servicePlan

		service, serviceWarnings, err := actor.GetService(serviceInstance.ServiceGUID)
		allWarnings = append(allWarnings, serviceWarnings...)
		if err != nil {
			log.WithField("service_guid", serviceInstance.ServiceGUID).Errorln("looking up service:", err)
			if _, ok := err.(ccerror.ForbiddenError); !ok {
				return serviceInstanceSummary, allWarnings, err
			}
			log.Warning("Forbidden Error - ignoring and continue")
			allWarnings = append(allWarnings, fmt.Sprintf("This org is not authorized to view necessary data about this service. Contact your administrator regarding service GUID %s.", serviceInstance.ServiceGUID))
		}
		serviceInstanceSummary.Service = service

		var bindingsWarnings Warnings
		serviceBindings, bindingsWarnings, err = actor.GetServiceBindingsByServiceInstance(serviceInstance.GUID)
		allWarnings = append(allWarnings, bindingsWarnings...)
		if err != nil {
			log.WithField("GUID", serviceInstance.GUID).Errorln("looking up service binding:", err)
			return serviceInstanceSummary, allWarnings, err
		}
	} else {
		log.Debug("service is user provided")
		var bindingsWarnings Warnings
		var err error
		serviceBindings, bindingsWarnings, err = actor.GetServiceBindingsByUserProvidedServiceInstance(serviceInstance.GUID)
		allWarnings = append(allWarnings, bindingsWarnings...)
		if err != nil {
			log.WithField("service_instance_guid", serviceInstance.GUID).Errorln("looking up service bindings:", err)
			return serviceInstanceSummary, allWarnings, err
		}
	}

	for _, serviceBinding := range serviceBindings {
		log.WithFields(log.Fields{
			"app_guid":             serviceBinding.AppGUID,
			"service_binding_guid": serviceBinding.GUID,
		}).Debug("application lookup")

		app, appWarnings, err := actor.GetApplication(serviceBinding.AppGUID)
		allWarnings = append(allWarnings, appWarnings...)
		if err != nil {
			log.WithFields(log.Fields{
				"app_guid":             serviceBinding.AppGUID,
				"service_binding_guid": serviceBinding.GUID,
			}).Errorln("looking up application:", err)
			return serviceInstanceSummary, allWarnings, err
		}

		serviceInstanceSummary.BoundApplications = append(serviceInstanceSummary.BoundApplications, BoundApplication{AppName: app.Name, ServiceBindingName: serviceBinding.Name})
	}

	sort.Slice(
		serviceInstanceSummary.BoundApplications,
		func(i, j int) bool {
			return sorting.LessIgnoreCase(serviceInstanceSummary.BoundApplications[i].AppName, serviceInstanceSummary.BoundApplications[j].AppName)
		})

	return serviceInstanceSummary, allWarnings, nil
}
