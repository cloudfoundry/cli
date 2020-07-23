package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/railway"
)

func (actor Actor) GetServiceInstanceByNameAndSpace(serviceInstanceName string, spaceGUID string) (resources.ServiceInstance, Warnings, error) {
	serviceInstance, _, warnings, err := actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
	switch e := err.(type) {
	case ccerror.ServiceInstanceNotFoundError:
		return serviceInstance, Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: e.Name}
	default:
		return serviceInstance, Warnings(warnings), err
	}
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

func (actor Actor) UpdateManagedServiceInstance(serviceInstanceName, spaceGUID string, serviceInstanceUpdates resources.ServiceInstance) (Warnings, error) {
	var (
		original resources.ServiceInstance
		jobURL   ccv3.JobURL
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			original, _, warnings, err = actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if original.Type != resources.ManagedServiceInstance {
				err = actionerror.ServiceInstanceTypeError{
					Name:         serviceInstanceName,
					RequiredType: resources.ManagedServiceInstance,
				}
			}
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = actor.CloudControllerClient.UpdateServiceInstance(original.GUID, serviceInstanceUpdates)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if jobURL != "" {
				warnings, err = actor.CloudControllerClient.PollJob(jobURL)
			}
			return
		},
	)

	switch err.(type) {
	case nil:
	case ccerror.ServiceInstanceNotFoundError:
		return Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	default:
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

func (actor Actor) fetchServiceInstanceDetails(serviceInstanceName string, spaceGUID string) (resources.ServiceInstance, ccv3.IncludedResources, Warnings, error) {
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
		return resources.ServiceInstance{}, ccv3.IncludedResources{}, Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	default:
		return resources.ServiceInstance{}, ccv3.IncludedResources{}, Warnings(warnings), err
	}

	return serviceInstance, included, Warnings(warnings), nil
}

type ManagedServiceInstanceParams struct {
	ServiceOfferingName string
	ServicePlanName     string
	ServiceInstanceName string
	ServiceBrokerName   string
	SpaceGUID           string
	Tags                types.OptionalStringSlice
}

func (actor Actor) CreateManagedServiceInstance(params ManagedServiceInstanceParams) (Warnings, error) {
	allWarnings := Warnings{}

	servicePlan, warnings, err := actor.GetServicePlanByNameOfferingAndBroker(
		params.ServicePlanName,
		params.ServiceOfferingName,
		params.ServiceBrokerName,
	)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	serviceInstance := resources.ServiceInstance{
		Type:            resources.ManagedServiceInstance,
		Name:            params.ServiceInstanceName,
		ServicePlanGUID: servicePlan.GUID,
		SpaceGUID:       params.SpaceGUID,
		Tags:            params.Tags,
	}

	jobURL, clientWarnings, err := actor.CloudControllerClient.CreateServiceInstance(serviceInstance)
	allWarnings = append(allWarnings, clientWarnings...)
	if err != nil {
		return allWarnings, err
	}

	clientWarnings, err = actor.CloudControllerClient.PollJobForState(jobURL, constant.JobPolling)
	allWarnings = append(allWarnings, clientWarnings...)

	return allWarnings, err

}
