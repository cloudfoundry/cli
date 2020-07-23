package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/railway"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
)

const featureFlagServiceInstanceSharing string = "service_instance_sharing"

type SharedStatus struct {
	FeatureFlagIsDisabled   bool
	OfferingDisablesSharing bool
	IsShared                bool
}

type ServiceInstanceParameters struct {
	Value         types.OptionalObject
	MissingReason string
}

type ServiceInstanceUpgradeState int

type ServiceInstanceUpgradeStatus struct {
	State       ServiceInstanceUpgradeState
	Description string
}

const (
	ServiceInstanceUpgradeNotSupported ServiceInstanceUpgradeState = iota
	ServiceInstanceUpgradeAvailable
	ServiceInstanceUpgradeNotAvailable
)

type ServiceInstanceDetails struct {
	resources.ServiceInstance
	ServiceOffering   resources.ServiceOffering
	ServicePlan       resources.ServicePlan
	ServiceBrokerName string
	Parameters        ServiceInstanceParameters
	SharedStatus      SharedStatus
	UpgradeStatus     ServiceInstanceUpgradeStatus
}

func (actor Actor) GetServiceInstanceDetails(serviceInstanceName string, spaceGUID string) (ServiceInstanceDetails, Warnings, error) {
	var serviceInstanceDetails ServiceInstanceDetails

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstanceDetails, warnings, err = actor.getServiceInstanceDetails(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			serviceInstanceDetails.Parameters, warnings = actor.getServiceInstanceParameters(serviceInstanceDetails.GUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			serviceInstanceDetails.SharedStatus, warnings, err = actor.getServiceInstanceSharedStatus(serviceInstanceDetails)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			serviceInstanceDetails.UpgradeStatus, warnings, err = actor.getServiceInstanceUpgradeStatus(serviceInstanceDetails)
			return
		},
	)
	if err != nil {
		return ServiceInstanceDetails{}, Warnings(warnings), err
	}

	return serviceInstanceDetails, Warnings(warnings), nil
}

func (actor Actor) getServiceInstanceDetails(serviceInstanceName string, spaceGUID string) (ServiceInstanceDetails, ccv3.Warnings, error) {
	query := []ccv3.Query{
		{
			Key:    ccv3.FieldsServicePlan,
			Values: []string{"name", "guid"},
		},
		{
			Key:    ccv3.FieldsServicePlanServiceOffering,
			Values: []string{"name", "guid", "description", "tags", "documentation_url"},
		},
		{
			Key:    ccv3.FieldsServicePlanServiceOfferingServiceBroker,
			Values: []string{"name", "guid"},
		},
	}

	serviceInstance, included, warnings, err := actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID, query...)
	switch err.(type) {
	case nil:
	case ccerror.ServiceInstanceNotFoundError:
		return ServiceInstanceDetails{}, warnings, actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	default:
		return ServiceInstanceDetails{}, warnings, err
	}

	result := ServiceInstanceDetails{
		ServiceInstance:   serviceInstance,
		ServicePlan:       extractServicePlan(included),
		ServiceOffering:   extractServiceOffering(included),
		ServiceBrokerName: extractServiceBrokerName(included),
	}

	return result, warnings, nil
}

func (actor Actor) getServiceInstanceParameters(serviceInstanceGUID string) (ServiceInstanceParameters, ccv3.Warnings) {
	params, warnings, err := actor.CloudControllerClient.GetServiceInstanceParameters(serviceInstanceGUID)
	if err != nil {
		if e, ok := err.(ccerror.V3UnexpectedResponseError); ok && len(e.Errors) > 0 {
			return ServiceInstanceParameters{MissingReason: e.Errors[0].Detail}, warnings
		} else {
			return ServiceInstanceParameters{MissingReason: err.Error()}, warnings
		}
	}

	return ServiceInstanceParameters{Value: params}, warnings
}

func (actor Actor) getServiceInstanceSharedStatus(serviceInstanceDetails ServiceInstanceDetails) (SharedStatus, ccv3.Warnings, error) {
	if serviceInstanceDetails.Type != resources.ManagedServiceInstance {
		return SharedStatus{}, nil, nil
	}

	var (
		featureFlag             resources.FeatureFlag
		offeringDisablesSharing bool
		sharedSpaces            []resources.Space
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			featureFlag, warnings, err = actor.CloudControllerClient.GetFeatureFlag(featureFlagServiceInstanceSharing)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			offeringDisablesSharing, warnings, err = actor.getOfferingSharingDetails(serviceInstanceDetails.ServiceOffering.GUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			sharedSpaces, warnings, err = actor.CloudControllerClient.GetServiceInstanceSharedSpaces(serviceInstanceDetails.GUID)
			return
		},
	)
	if err != nil {
		return SharedStatus{}, warnings, err
	}

	sharedStatus := SharedStatus{
		IsShared:                len(sharedSpaces) > 0,
		OfferingDisablesSharing: offeringDisablesSharing,
		FeatureFlagIsDisabled:   !featureFlag.Enabled,
	}

	return sharedStatus, warnings, nil
}

func (actor Actor) getOfferingSharingDetails(serviceOfferingGUID string) (bool, ccv3.Warnings, error) {
	serviceOffering, serviceOfferingWarning, err :=
		actor.CloudControllerClient.GetServiceOfferingByGUID(serviceOfferingGUID)

	switch err := err.(type) {
	case nil:
		return !serviceOffering.AllowsInstanceSharing, serviceOfferingWarning, nil
	case ccerror.ServiceOfferingNotFoundError:
		return false, serviceOfferingWarning, nil
	default:
		return false, serviceOfferingWarning, err
	}
}

func (actor Actor) getServiceInstanceUpgradeStatus(serviceInstanceDetails ServiceInstanceDetails) (ServiceInstanceUpgradeStatus, ccv3.Warnings, error) {
	if !serviceInstanceDetails.UpgradeAvailable.Value {
		if serviceInstanceDetails.MaintenanceInfoVersion == "" {
			return ServiceInstanceUpgradeStatus{State: ServiceInstanceUpgradeNotSupported}, nil, nil
		}
		return ServiceInstanceUpgradeStatus{State: ServiceInstanceUpgradeNotAvailable}, nil, nil
	}

	servicePlan, warnings, err := actor.CloudControllerClient.GetServicePlanByGUID(serviceInstanceDetails.ServicePlanGUID)
	switch err.(type) {
	case nil:
		return ServiceInstanceUpgradeStatus{
			State:       ServiceInstanceUpgradeAvailable,
			Description: servicePlan.MaintenanceInfoDescription,
		}, warnings, nil
	case ccerror.ServicePlanNotFound:
		return ServiceInstanceUpgradeStatus{
			State:       ServiceInstanceUpgradeAvailable,
			Description: "No upgrade details where found",
		}, warnings, nil
	default:
		return ServiceInstanceUpgradeStatus{}, warnings, err
	}
}

func extractServicePlan(included ccv3.IncludedResources) resources.ServicePlan {
	if len(included.ServicePlans) == 1 {
		return included.ServicePlans[0]
	}

	return resources.ServicePlan{}
}

func extractServiceBrokerName(included ccv3.IncludedResources) string {
	if len(included.ServiceBrokers) == 1 {
		return included.ServiceBrokers[0].Name
	}

	return ""
}

func extractServiceOffering(included ccv3.IncludedResources) resources.ServiceOffering {
	if len(included.ServiceOfferings) == 1 {
		return included.ServiceOfferings[0]
	}

	return resources.ServiceOffering{}
}
