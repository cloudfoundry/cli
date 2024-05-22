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

type UpdateManagedServiceInstanceParams struct {
	ServiceInstanceName string
	ServicePlanName     string
	SpaceGUID           string
	Tags                types.OptionalStringSlice
	Parameters          types.OptionalObject
}

type CreateManagedServiceInstanceParams struct {
	ServiceOfferingName string
	ServicePlanName     string
	ServiceInstanceName string
	ServiceBrokerName   string
	SpaceGUID           string
	Tags                types.OptionalStringSlice
	Parameters          types.OptionalObject
}

func (actor Actor) GetServiceInstanceByNameAndSpace(serviceInstanceName string, spaceGUID string) (resources.ServiceInstance, Warnings, error) {
	serviceInstance, _, warnings, err := actor.getServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
	return serviceInstance, Warnings(warnings), err
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

func (actor Actor) CreateManagedServiceInstance(params CreateManagedServiceInstanceParams) (chan PollJobEvent, Warnings, error) {
	var (
		servicePlan resources.ServicePlan
		jobURL      ccv3.JobURL
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			var v7Warnings Warnings
			servicePlan, v7Warnings, err = actor.GetServicePlanByNameOfferingAndBroker(
				params.ServicePlanName,
				params.ServiceOfferingName,
				params.ServiceBrokerName,
			)
			return ccv3.Warnings(v7Warnings), err
		},
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance := resources.ServiceInstance{
				Type:            resources.ManagedServiceInstance,
				Name:            params.ServiceInstanceName,
				ServicePlanGUID: servicePlan.GUID,
				SpaceGUID:       params.SpaceGUID,
				Tags:            params.Tags,
				Parameters:      params.Parameters,
			}

			jobURL, warnings, err = actor.CloudControllerClient.CreateServiceInstance(serviceInstance)
			return
		},
	)
	switch e := err.(type) {
	case nil:
		return actor.PollJobToEventStream(jobURL), Warnings(warnings), nil
	case actionerror.DuplicateServicePlanError:
		return nil, Warnings(warnings), actionerror.ServiceBrokerNameRequiredError{
			ServiceOfferingName: e.ServiceOfferingName,
		}
	default:
		return nil, Warnings(warnings), err
	}
}

func (actor Actor) UpdateManagedServiceInstance(params UpdateManagedServiceInstanceParams) (chan PollJobEvent, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		serviceOffering resources.ServiceOffering
		serviceBroker   resources.ServiceBroker
		newPlanGUID     string
		jobURL          ccv3.JobURL
		stream          chan PollJobEvent
	)

	planChangeRequested := params.ServicePlanName != ""

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, serviceOffering, serviceBroker, warnings, err = actor.getServiceInstanceForUpdate(
				params.ServiceInstanceName,
				params.SpaceGUID,
				planChangeRequested,
			)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			err = assertServiceInstanceType(resources.ManagedServiceInstance, serviceInstance)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if planChangeRequested {
				newPlanGUID, warnings, err = actor.getPlanForInstanceUpdate(params.ServicePlanName, serviceOffering, serviceBroker)
			}
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = actor.updateManagedServiceInstance(serviceInstance, newPlanGUID, params)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			stream = actor.PollJobToEventStream(jobURL)
			return
		},
	)

	return stream, Warnings(warnings), err
}

func (actor Actor) UpgradeManagedServiceInstance(serviceInstanceName string, spaceGUID string) (chan PollJobEvent, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		servicePlan     resources.ServicePlan
		jobURL          ccv3.JobURL
		stream          chan PollJobEvent
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
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
			stream = actor.PollJobToEventStream(jobURL)
			return
		},
	)

	return stream, Warnings(warnings), err
}

func (actor Actor) RenameServiceInstance(currentServiceInstanceName, spaceGUID, newServiceInstanceName string) (Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		jobURL          ccv3.JobURL
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getServiceInstanceByNameAndSpace(currentServiceInstanceName, spaceGUID)
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
	)

	return Warnings(warnings), err
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

func (actor Actor) PurgeServiceInstance(serviceInstanceName, spaceGUID string) (Warnings, error) {
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
		return Warnings(warnings), nil
	case ccerror.ServiceInstanceNotFoundError:
		return Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	default:
		return Warnings(warnings), err
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

func (actor Actor) getServiceInstanceByNameAndSpace(serviceInstanceName string, spaceGUID string, query ...ccv3.Query) (resources.ServiceInstance, ccv3.IncludedResources, ccv3.Warnings, error) {
	serviceInstance, includedResources, warnings, err := actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID, query...)
	switch e := err.(type) {
	case ccerror.ServiceInstanceNotFoundError:
		return serviceInstance, ccv3.IncludedResources{}, warnings, actionerror.ServiceInstanceNotFoundError{Name: e.Name}
	default:
		return serviceInstance, includedResources, warnings, err
	}
}

func (actor Actor) getServiceInstanceForUpdate(serviceInstanceName string, spaceGUID string, includePlan bool) (resources.ServiceInstance, resources.ServiceOffering, resources.ServiceBroker, ccv3.Warnings, error) {
	var query []ccv3.Query
	if includePlan {
		query = append(
			query,
			ccv3.Query{Key: ccv3.FieldsServicePlanServiceOffering, Values: []string{"name", "guid"}},
			ccv3.Query{Key: ccv3.FieldsServicePlanServiceOfferingServiceBroker, Values: []string{"name"}},
		)
	}

	serviceInstance, includedResources, warnings, err := actor.getServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID, query...)

	var (
		serviceOffering resources.ServiceOffering
		serviceBroker   resources.ServiceBroker
	)
	if len(includedResources.ServiceOfferings) != 0 {
		serviceOffering = includedResources.ServiceOfferings[0]
	}
	if len(includedResources.ServiceBrokers) != 0 {
		serviceBroker = includedResources.ServiceBrokers[0]
	}

	return serviceInstance, serviceOffering, serviceBroker, warnings, err
}

func (actor Actor) getPlanForInstanceUpdate(planName string, serviceOffering resources.ServiceOffering, serviceBroker resources.ServiceBroker) (string, ccv3.Warnings, error) {
	plans, warnings, err := actor.CloudControllerClient.GetServicePlans([]ccv3.Query{
		{Key: ccv3.ServiceOfferingGUIDsFilter, Values: []string{serviceOffering.GUID}},
		{Key: ccv3.NameFilter, Values: []string{planName}},
	}...)

	switch {
	case err != nil:
		return "", warnings, err
	case len(plans) == 0:
		return "", warnings, actionerror.ServicePlanNotFoundError{
			PlanName:          planName,
			OfferingName:      serviceOffering.Name,
			ServiceBrokerName: serviceBroker.Name,
		}
	default:
		return plans[0].GUID, warnings, nil
	}
}

func (actor Actor) updateManagedServiceInstance(serviceInstance resources.ServiceInstance, newServicePlanGUID string, params UpdateManagedServiceInstanceParams) (ccv3.JobURL, ccv3.Warnings, error) {
	if newServicePlanGUID == serviceInstance.ServicePlanGUID {
		newServicePlanGUID = ""
	}

	update := resources.ServiceInstance{
		ServicePlanGUID: newServicePlanGUID,
		Tags:            params.Tags,
		Parameters:      params.Parameters,
	}

	if update.ServicePlanGUID == "" && !update.Tags.IsSet && !update.Parameters.IsSet {
		return "", nil, actionerror.ServiceInstanceUpdateIsNoop{}
	}

	return actor.CloudControllerClient.UpdateServiceInstance(serviceInstance.GUID, update)
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
