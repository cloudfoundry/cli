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

type ServiceInstanceUpdateManagedParams struct {
	ServicePlanName types.OptionalString
	Tags            types.OptionalStringSlice
	Parameters      types.OptionalObject
}

type ManagedServiceInstanceParams struct {
	ServiceOfferingName string
	ServicePlanName     string
	ServiceInstanceName string
	ServiceBrokerName   string
	SpaceGUID           string
	Tags                types.OptionalStringSlice
	Parameters          types.OptionalObject
}

type ServiceInstanceDeleteState int

const (
	ServiceInstanceUnknownState ServiceInstanceDeleteState = iota
	ServiceInstanceDidNotExist
	ServiceInstanceGone
	ServiceInstanceDeleteInProgress
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

func (actor Actor) CreateUserProvidedServiceInstance(serviceInstance resources.ServiceInstance) (Warnings, error) {
	serviceInstance.Type = resources.UserProvidedServiceInstance
	_, warnings, err := actor.CloudControllerClient.CreateServiceInstance(serviceInstance)
	return Warnings(warnings), err
}

func (actor Actor) UpdateUserProvidedServiceInstance(serviceInstanceName, spaceGUID string, serviceInstanceUpdates resources.ServiceInstance) (Warnings, error) {
	var original resources.ServiceInstance

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			original, _, warnings, err = actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			err = assertServiceInstanceType(resources.UserProvidedServiceInstance, original)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			_, warnings, err = actor.CloudControllerClient.UpdateServiceInstance(original.GUID, serviceInstanceUpdates)
			return
		},
	)

	return Warnings(warnings), err
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
		Parameters:      params.Parameters,
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

// TODO finish this refactor
func (actor Actor) UpdateManagedServiceInstance(serviceInstanceName, spaceGUID string, serviceInstanceUpdates ServiceInstanceUpdateManagedParams) (bool, Warnings, error) {
	var (
		jobURL      ccv3.JobURL
		allWarnings Warnings
		updates     managedServiceInstanceUpdate
	)

	updatesBuilder := &managedServiceInstanceUpdateBuilder{
		actor:               actor,
		requestedUpdates:    serviceInstanceUpdates,
		serviceInstanceName: serviceInstanceName,
		spaceGUID:           spaceGUID,
	}

	warnings, err := handleServiceInstanceErrors(railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			updates, warnings, err = updatesBuilder.build()
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if updatesBuilder.IsNoop() {
				return
			}
			jobURL, warnings, err = actor.CloudControllerClient.UpdateServiceInstance(updatesBuilder.serviceInstance.GUID, *updates.serviceInstance)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if updatesBuilder.IsNoop() {
				return
			}
			return actor.CloudControllerClient.PollJobForState(jobURL, constant.JobPolling)
		},
	))

	allWarnings = append(allWarnings, warnings...)

	return updatesBuilder.IsNoop(), allWarnings, err
}

func (actor Actor) UpgradeManagedServiceInstance(serviceInstanceName string, spaceGUID string) (Warnings, error) {
	var serviceInstance resources.ServiceInstance
	var servicePlan resources.ServicePlan
	var jobURL ccv3.JobURL

	return handleServiceInstanceErrors(railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if serviceInstance.UpgradeAvailable.Value != true {
				err = actionerror.ServiceInstanceUpgradeNotAvailableError{}
			}
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			servicePlan, warnings, err = actor.CloudControllerClient.GetServicePlanByGUID(serviceInstance.ServicePlanGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = actor.CloudControllerClient.UpdateServiceInstance(serviceInstance.GUID, resources.ServiceInstance{
				MaintenanceInfoVersion: servicePlan.MaintenanceInfoVersion,
			})
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			return actor.CloudControllerClient.PollJobForState(jobURL, constant.JobPolling)
		},
	))
}

func (actor Actor) RenameServiceInstance(currentServiceInstanceName, spaceGUID, newServiceInstanceName string) (Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		jobURL          ccv3.JobURL
	)

	return handleServiceInstanceErrors(railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(currentServiceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = actor.CloudControllerClient.UpdateServiceInstance(
				serviceInstance.GUID,
				resources.ServiceInstance{Name: newServiceInstanceName},
			)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			return actor.CloudControllerClient.PollJobForState(jobURL, constant.JobPolling)
		},
	))
}

func (actor Actor) DeleteServiceInstance(serviceInstanceName, spaceGUID string, wait bool) (ServiceInstanceDeleteState, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		jobURL          ccv3.JobURL
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = actor.CloudControllerClient.DeleteServiceInstance(serviceInstance.GUID)
			return
		},
		func() (ccv3.Warnings, error) {
			return actor.pollJob(jobURL, wait)
		},
	)

	switch err.(type) {
	case nil:
	case ccerror.ServiceInstanceNotFoundError:
		return ServiceInstanceDidNotExist, Warnings(warnings), nil
	default:
		return ServiceInstanceUnknownState, Warnings(warnings), err
	}

	if jobURL != "" && !wait {
		return ServiceInstanceDeleteInProgress, Warnings(warnings), nil
	}
	return ServiceInstanceGone, Warnings(warnings), nil
}

func (actor Actor) UnshareServiceInstanceByServiceInstanceAndSpace(serviceInstanceGUID string, sharedToSpaceGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.DeleteServiceInstanceRelationshipsSharedSpace(serviceInstanceGUID, sharedToSpaceGUID)
	return Warnings(warnings), err
}

func (actor Actor) pollJob(jobURL ccv3.JobURL, wait bool) (ccv3.Warnings, error) {
	switch {
	case jobURL == "":
		return ccv3.Warnings{}, nil
	case wait:
		return actor.CloudControllerClient.PollJob(jobURL)
	default:
		return actor.CloudControllerClient.PollJobForState(jobURL, constant.JobPolling)
	}
}

func assertServiceInstanceType(requiredType resources.ServiceInstanceType, instance resources.ServiceInstance) error {
	if instance.Type != requiredType {
		return actionerror.ServiceInstanceTypeError{
			Name:         instance.Name,
			RequiredType: requiredType,
		}
	}

	return nil
}

func handleServiceInstanceErrors(warnings ccv3.Warnings, err error) (Warnings, error) {
	switch e := err.(type) {
	case nil:
		return Warnings(warnings), nil
	case ccerror.ServiceInstanceNotFoundError:
		return Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: e.Name}
	default:
		return Warnings(warnings), err
	}
}

type managedServiceInstanceUpdateBuilder struct {
	actor               Actor
	requestedUpdates    ServiceInstanceUpdateManagedParams
	serviceInstanceName string
	spaceGUID           string

	serviceInstance     resources.ServiceInstance
	serviceOfferingName string
	serviceBrokerName   string
	updates             *managedServiceInstanceUpdate
}

type managedServiceInstanceUpdate struct {
	serviceInstance *resources.ServiceInstance
	isNoop          bool
}

func (builder *managedServiceInstanceUpdateBuilder) build() (managedServiceInstanceUpdate, ccv3.Warnings, error) {
	builder.updates = &managedServiceInstanceUpdate{
		serviceInstance: &resources.ServiceInstance{
			Tags:       builder.requestedUpdates.Tags,
			Parameters: builder.requestedUpdates.Parameters,
		},
		isNoop: false,
	}

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			return builder.GetServiceInstanceToUpdate()
		},
		func() (warnings ccv3.Warnings, err error) {
			err = assertServiceInstanceType(resources.ManagedServiceInstance, builder.serviceInstance)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			return builder.GetServicePlan()
		},
	)
	if err != nil {
		return managedServiceInstanceUpdate{}, warnings, err
	}
	return *builder.updates, warnings, err
}

func (builder *managedServiceInstanceUpdateBuilder) GetServicePlan() (ccv3.Warnings, error) {
	var actorWarnings Warnings

	if !builder.requestedUpdates.ServicePlanName.IsSet {
		return nil, nil
	}

	plan, actorWarnings, err := builder.actor.GetServicePlanByNameOfferingAndBroker(
		builder.requestedUpdates.ServicePlanName.Value,
		builder.serviceOfferingName,
		builder.serviceBrokerName,
	)
	if err == nil {
		builder.updates.serviceInstance.ServicePlanGUID = plan.GUID
	}
	return ccv3.Warnings(actorWarnings), err
}

func (builder *managedServiceInstanceUpdateBuilder) GetServiceInstanceToUpdate() (warnings ccv3.Warnings, err error) {
	var serviceInstanceQuery []ccv3.Query

	if builder.requestedUpdates.ServicePlanName.IsSet {
		serviceInstanceQuery = []ccv3.Query{
			{
				Key:    ccv3.FieldsServicePlanServiceOffering,
				Values: []string{"name"},
			},
			{
				Key:    ccv3.FieldsServicePlanServiceOfferingServiceBroker,
				Values: []string{"name"},
			},
		}
	}
	var includedResources ccv3.IncludedResources

	builder.serviceInstance, includedResources, warnings, err = builder.actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(
		builder.serviceInstanceName,
		builder.spaceGUID,
		serviceInstanceQuery...,
	)

	if len(includedResources.ServiceBrokers) > 0 {
		builder.serviceBrokerName = includedResources.ServiceBrokers[0].Name
	}
	if len(includedResources.ServiceOfferings) > 0 {
		builder.serviceOfferingName = includedResources.ServiceOfferings[0].Name
	}

	return warnings, err
}

func (builder *managedServiceInstanceUpdateBuilder) IsNoop() bool {
	if builder.updates.serviceInstance.ServicePlanGUID == builder.serviceInstance.ServicePlanGUID {
		if !(builder.updates.serviceInstance.Tags.IsSet || builder.updates.serviceInstance.Parameters.IsSet) {
			return true
		}
	}
	return false
}
