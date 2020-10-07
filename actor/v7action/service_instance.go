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

func (actor Actor) UpdateManagedServiceInstance(serviceInstanceName, spaceGUID string, serviceInstanceUpdates ServiceInstanceUpdateManagedParams) (bool, Warnings, error) {
	var (
		jobURL                ccv3.JobURL
		allWarnings           Warnings
		serviceInstanceUpdate managedServiceInstanceUpdate
	)

	warnings, err := handleServiceInstanceErrors(railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			validator := managedServiceInstanceValidator{
				actor: actor,
			}
			serviceInstanceUpdate, warnings, err = validator.validateManagedInstanceUpdate(serviceInstanceName, spaceGUID, serviceInstanceUpdates)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if serviceInstanceUpdate.isNoop {
				return
			}
			jobURL, warnings, err = actor.CloudControllerClient.UpdateServiceInstance(serviceInstanceUpdate.serviceInstanceGUID, serviceInstanceUpdate.updateRequest)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if serviceInstanceUpdate.isNoop {
				return
			}
			return actor.CloudControllerClient.PollJobForState(jobURL, constant.JobPolling)
		},
	))

	allWarnings = append(allWarnings, warnings...)

	return serviceInstanceUpdate.isNoop, allWarnings, err
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

func (actor Actor) DeleteServiceInstance(serviceInstanceName, spaceGUID string) (chan PollJobEvent, Warnings, error) {
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
	)

	switch err.(type) {
	case nil:
	case ccerror.ServiceInstanceNotFoundError:
		return nil, Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	default:
		return nil, Warnings(warnings), err
	}

	switch jobURL {
	case "":
		return nil, Warnings(warnings), nil
	default:
		return actor.PollJobToEventStream(jobURL), Warnings(warnings), nil
	}
}

func (actor Actor) PurgeServiceInstance(serviceInstanceName, spaceGUID string) (ServiceInstanceDeleteState, Warnings, error) {
	var serviceInstance resources.ServiceInstance

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			_, warnings, err = actor.CloudControllerClient.DeleteServiceInstance(
				serviceInstance.GUID,
				ccv3.Query{Key: ccv3.Purge, Values: []string{"true"}},
			)
			return
		},
	)

	switch err.(type) {
	case nil:
		return ServiceInstanceGone, Warnings(warnings), nil
	case ccerror.ServiceInstanceNotFoundError:
		return ServiceInstanceDidNotExist, Warnings(warnings), nil
	default:
		return ServiceInstanceUnknownState, Warnings(warnings), err
	}
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

type managedServiceInstanceUpdate struct {
	serviceInstanceGUID string
	updateRequest       resources.ServiceInstance
	isNoop              bool
}

func (m *managedServiceInstanceUpdate) updateIsNoop(originalInstanceDetails ServiceInstanceDetails) {
	if originalInstanceDetails.ServicePlanGUID == m.updateRequest.ServicePlanGUID {
		if !(m.updateRequest.Tags.IsSet || m.updateRequest.Parameters.IsSet) {
			m.isNoop = true
		}
	}
}

type managedServiceInstanceValidator struct {
	actor Actor
}

func (v *managedServiceInstanceValidator) validateManagedInstanceUpdate(serviceInstanceName, spaceGUID string, requestedUpdates ServiceInstanceUpdateManagedParams) (
	managedServiceInstanceUpdate, ccv3.Warnings, error) {
	var originalInstanceDetails ServiceInstanceDetails

	m := &managedServiceInstanceUpdate{
		updateRequest: resources.ServiceInstance{
			Tags:       requestedUpdates.Tags,
			Parameters: requestedUpdates.Parameters,
		},
	}

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			originalInstanceDetails, warnings, err = v.getServiceInstanceToUpdate(serviceInstanceName, spaceGUID, requestedUpdates.ServicePlanName)
			if err == nil {
				m.serviceInstanceGUID = originalInstanceDetails.GUID
			}
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			err = assertServiceInstanceType(resources.ManagedServiceInstance, originalInstanceDetails.ServiceInstance)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			m.updateRequest.ServicePlanGUID, warnings, err = v.getTargetServicePlan(
				requestedUpdates.ServicePlanName,
				originalInstanceDetails.ServiceOffering.Name,
				originalInstanceDetails.ServiceBrokerName,
			)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			m.updateIsNoop(originalInstanceDetails)
			return
		},
	)
	if err != nil {
		return managedServiceInstanceUpdate{}, warnings, err
	}
	return *m, warnings, err
}

func (v *managedServiceInstanceValidator) getServiceInstanceToUpdate(serviceInstanceName, spaceGUID string, targetPlanName types.OptionalString) (
	serviceInstanceDetails ServiceInstanceDetails, warnings ccv3.Warnings, err error) {
	var serviceInstanceQuery []ccv3.Query
	if targetPlanName.IsSet {
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

	serviceInstanceDetails.ServiceInstance, includedResources, warnings, err = v.actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(
		serviceInstanceName,
		spaceGUID,
		serviceInstanceQuery...,
	)

	if len(includedResources.ServiceBrokers) > 0 {
		serviceInstanceDetails.ServiceBrokerName = includedResources.ServiceBrokers[0].Name
	}
	if len(includedResources.ServiceOfferings) > 0 {
		serviceInstanceDetails.ServiceOffering.Name = includedResources.ServiceOfferings[0].Name
	}

	return serviceInstanceDetails, warnings, err
}

func (v *managedServiceInstanceValidator) getTargetServicePlan(servicePlanName types.OptionalString, offeringName, brokerName string) (string, ccv3.Warnings, error) {
	var actorWarnings Warnings
	targetPlanGUID := ""

	if !servicePlanName.IsSet {
		return targetPlanGUID, nil, nil
	}

	plan, actorWarnings, err := v.actor.GetServicePlanByNameOfferingAndBroker(
		servicePlanName.Value,
		offeringName,
		brokerName,
	)
	if err == nil {
		targetPlanGUID = plan.GUID
	}
	return targetPlanGUID, ccv3.Warnings(actorWarnings), err
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
