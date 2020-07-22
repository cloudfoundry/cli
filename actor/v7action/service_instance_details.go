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

type ServiceInstanceDetails struct {
	resources.ServiceInstance
	ServiceOffering   resources.ServiceOffering
	ServicePlanName   string
	ServiceBrokerName string
	Parameters        ServiceInstanceParameters
	SharedStatus      SharedStatus
}

func (actor Actor) GetServiceInstanceDetails(serviceInstanceName string, spaceGUID string) (ServiceInstanceDetails, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		included        ccv3.IncludedResources
		params          ServiceInstanceParameters
		sharedStatus    SharedStatus
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, included, warnings, err = actor.getServiceInstanceWithIncludes(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			params, warnings = actor.getServiceInstanceParameters(serviceInstance.GUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			sharedStatus, warnings, err = actor.getSharedStatus(serviceInstance)
			return
		},
	)
	if err != nil {
		return ServiceInstanceDetails{}, Warnings(warnings), err
	}

	result := ServiceInstanceDetails{
		ServiceInstance:   serviceInstance,
		ServicePlanName:   extractServicePlanName(included),
		ServiceOffering:   extractServiceOffering(included),
		ServiceBrokerName: extractServiceBrokerName(included),
		Parameters:        params,
		SharedStatus:      sharedStatus,
	}

	return result, Warnings(warnings), nil
}

func (actor Actor) getServiceInstanceWithIncludes(serviceInstanceName string, spaceGUID string) (resources.ServiceInstance, ccv3.IncludedResources, ccv3.Warnings, error) {
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
		return serviceInstance, included, warnings, nil
	case ccerror.ServiceInstanceNotFoundError:
		return resources.ServiceInstance{}, ccv3.IncludedResources{}, warnings, actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	default:
		return resources.ServiceInstance{}, ccv3.IncludedResources{}, warnings, err
	}
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

func (actor Actor) getSharedStatus(serviceInstance resources.ServiceInstance) (SharedStatus, ccv3.Warnings, error) {
	if serviceInstance.Type != resources.ManagedServiceInstance {
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
			offeringDisablesSharing, warnings, err = actor.getOfferingSharingDetails(serviceInstance.ServiceOfferingGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			sharedSpaces, warnings, err = actor.CloudControllerClient.GetServiceInstanceSharedSpaces(serviceInstance.GUID)
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

func extractServicePlanName(included ccv3.IncludedResources) string {
	if len(included.ServicePlans) == 1 {
		return included.ServicePlans[0].Name
	}

	return ""
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
