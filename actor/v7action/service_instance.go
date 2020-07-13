package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

type SharedStatus interface {
	IsShared() bool
}

type ServiceIsShared struct {
}

type ServiceIsNotShared struct {
}

func (_ ServiceIsShared) IsShared() bool {
	return true
}

func (_ ServiceIsNotShared) IsShared() bool {
	return false
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
		sharedSpaces, shareWarnings, err := actor.CloudControllerClient.GetServiceInstanceSharedSpaces(serviceInstance.GUID)

		warnings = append(warnings, shareWarnings...)

		if err != nil {
			return ServiceInstanceWithRelationships{}, Warnings(warnings), err
		}

		if len(sharedSpaces) == 0 {
			result.SharedStatus = ServiceIsNotShared{}
		} else {
			result.SharedStatus = ServiceIsShared{}
		}
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
