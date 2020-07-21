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

type ServiceInstanceDetails struct {
	resources.ServiceInstance
	ServiceOffering         resources.ServiceOffering
	ServicePlanName         string
	ServiceBrokerName       string
	Parameters              types.OptionalObject
	ParametersMissingReason string
	SharedStatus            SharedStatus
}

func (actor Actor) GetServiceInstanceByNameAndSpace(serviceInstanceName string, spaceGUID string) (resources.ServiceInstance, Warnings, error) {
	serviceInstance, _, warnings, err := actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
	switch e := err.(type) {
	case ccerror.ServiceInstanceNotFoundError:
		return serviceInstance, Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: e.Name}
	default:
		return serviceInstance, Warnings(warnings), err
	}
}

func (actor Actor) GetServiceInstanceDetails(serviceInstanceName string, spaceGUID string) (ServiceInstanceDetails, Warnings, error) {
	query := []ccv3.Query{
		{
			Key:    ccv3.FieldsServicePlan,
			Values: []string{"name", "guid"},
		},
		{
			Key:    ccv3.FieldsServicePlanServiceOffering,
			Values: []string{"name", "guid", "description", "documentation_url"},
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
		return ServiceInstanceDetails{}, Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	default:
		return ServiceInstanceDetails{}, Warnings(warnings), err
	}

	result := ServiceInstanceDetails{
		ServiceInstance: serviceInstance,
	}

	params, paramsWarnings, err := actor.CloudControllerClient.GetServiceInstanceParameters(serviceInstance.GUID)
	warnings = append(warnings, paramsWarnings...)
	result.Parameters = params
	if err != nil {
		if e, ok := err.(ccerror.V3UnexpectedResponseError); ok && len(e.Errors) > 0 {
			result.ParametersMissingReason = e.Errors[0].Detail
		} else {
			result.ParametersMissingReason = err.Error()
		}
	}

	if len(included.ServicePlans) == 1 {
		result.ServicePlanName = included.ServicePlans[0].Name
	}

	if len(included.ServiceOfferings) == 1 {
		result.ServiceOffering = included.ServiceOfferings[0]
	}

	if len(included.ServiceBrokers) == 1 {
		result.ServiceBrokerName = included.ServiceBrokers[0].Name
	}

	if result.Type == resources.ManagedServiceInstance {
		sharedStatus, sharedWarnings, err := actor.getSharedStatus(result, serviceInstance)

		warnings = append(warnings, sharedWarnings...)

		if err != nil {
			return ServiceInstanceDetails{}, Warnings(warnings), err
		}

		result.SharedStatus = sharedStatus
	}

	return result, Warnings(warnings), nil
}

func (actor Actor) UnshareServiceInstanceByServiceInstanceAndSpace(serviceInstanceGUID string, sharedToSpaceGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.DeleteServiceInstanceRelationshipsSharedSpace(serviceInstanceGUID, sharedToSpaceGUID)
	return Warnings(warnings), err
}

func (actor Actor) CreateUserProvidedServiceInstance(serviceInstance resources.ServiceInstance) (Warnings, error) {
	serviceInstance.Type = resources.UserProvidedServiceInstance
	_, warnings, err := actor.CloudControllerClient.CreateServiceInstance(serviceInstance)
	return Warnings(warnings), err
}

func (actor Actor) UpdateUserProvidedServiceInstance(serviceInstanceName, spaceGUID string, serviceInstanceUpdates resources.ServiceInstance) (Warnings, error) {
	original, _, warnings, err := actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
	if err != nil {
		return Warnings(warnings), err
	}

	if original.Type != resources.UserProvidedServiceInstance {
		return Warnings(warnings), actionerror.ServiceInstanceTypeError{
			Name:         serviceInstanceName,
			RequiredType: resources.UserProvidedServiceInstance,
		}
	}

	_, updateWarnings, err := actor.CloudControllerClient.UpdateServiceInstance(original.GUID, serviceInstanceUpdates)
	warnings = append(warnings, updateWarnings...)
	if err != nil {
		return Warnings(warnings), err
	}

	return Warnings(warnings), nil
}

func (actor Actor) RenameServiceInstance(currentServiceInstanceName, spaceGUID, newServiceInstanceName string) (Warnings, error) {
	var serviceInstance resources.ServiceInstance
	serviceInstance, warnings, err := actor.GetServiceInstanceByNameAndSpace(currentServiceInstanceName, spaceGUID)
	if err != nil {
		return warnings, err
	}

	jobURL, updateWarnings, err := actor.CloudControllerClient.UpdateServiceInstance(
		serviceInstance.GUID,
		resources.ServiceInstance{Name: newServiceInstanceName},
	)
	warnings = append(warnings, updateWarnings...)
	if err != nil {
		return warnings, err
	}

	if jobURL != "" {
		pollWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
		warnings = append(warnings, pollWarnings...)
		if err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}

func (actor Actor) getSharedStatus(result ServiceInstanceDetails, serviceInstance resources.ServiceInstance) (SharedStatus, ccv3.Warnings, error) {
	var featureFlag resources.FeatureFlag
	var offeringDisablesSharing bool
	var sharedSpaces []resources.Space

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			featureFlag, warnings, err = actor.CloudControllerClient.GetFeatureFlag(featureFlagServiceInstanceSharing)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			offeringDisablesSharing, warnings, err = actor.getOfferingSharingDetails(result.ServiceOffering.Name, result.ServiceBrokerName)
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

func (actor Actor) getOfferingSharingDetails(serviceOfferingName string, brokerName string) (bool, ccv3.Warnings, error) {
	serviceOffering, serviceOfferingWarning, err :=
		actor.CloudControllerClient.GetServiceOfferingByNameAndBroker(serviceOfferingName, brokerName)

	switch err := err.(type) {
	case nil:
		return !serviceOffering.AllowsInstanceSharing, serviceOfferingWarning, nil
	case ccerror.ServiceOfferingNotFoundError:
		return false, serviceOfferingWarning, nil
	default:
		return false, serviceOfferingWarning, err
	}
}
