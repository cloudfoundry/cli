package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

const featureFlagServiceInstanceSharing string = "service_instance_sharing"

type SharedStatus struct {
	FeatureFlagIsDisabled   bool
	OfferingDisablesSharing bool
	IsShared                bool
}

type ServiceInstanceWithRelationships struct {
	resources.ServiceInstance
	ServiceOffering   resources.ServiceOffering
	ServicePlanName   string
	ServiceBrokerName string
	SharedStatus      SharedStatus
}

func (actor Actor) GetServiceInstanceByNameAndSpace(serviceInstanceName string, spaceGUID string) (resources.ServiceInstance, Warnings, error) {
	serviceInstance, _, warnings, err := actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
	return serviceInstance, Warnings(warnings), err
}

func (actor Actor) GetServiceInstanceDetails(serviceInstanceName string, spaceGUID string) (ServiceInstanceWithRelationships, Warnings, error) {
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
		return ServiceInstanceWithRelationships{}, Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	default:
		return ServiceInstanceWithRelationships{}, Warnings(warnings), err
	}

	result := ServiceInstanceWithRelationships{
		ServiceInstance: serviceInstance,
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
			return ServiceInstanceWithRelationships{}, Warnings(warnings), err
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

func (actor Actor) getSharedStatus(result ServiceInstanceWithRelationships, serviceInstance resources.ServiceInstance) (SharedStatus, ccv3.Warnings, error) {
	var warnings ccv3.Warnings

	featureFlag, featureFlagWarning, err := actor.CloudControllerClient.GetFeatureFlag(featureFlagServiceInstanceSharing)
	warnings = append(warnings, featureFlagWarning...)
	if err != nil {
		return SharedStatus{}, warnings, err
	}

	offeringDisablesSharing, serviceOfferingWarnings, err := actor.getOfferingSharingDetails(result.ServiceOffering.Name, result.ServiceBrokerName)
	warnings = append(warnings, serviceOfferingWarnings...)
	if err != nil {
		return SharedStatus{}, warnings, err
	}

	sharedSpaces, shareWarnings, err := actor.CloudControllerClient.GetServiceInstanceSharedSpaces(serviceInstance.GUID)
	warnings = append(warnings, shareWarnings...)
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
