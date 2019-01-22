package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

type Service ccv2.Service

func (actor Actor) GetService(serviceGUID string) (Service, Warnings, error) {
	service, warnings, err := actor.CloudControllerClient.GetService(serviceGUID)
	return Service(service), Warnings(warnings), err
}

// GetServicesByNameAndBrokerName returns services based on the name provided.
// If there are no services, a ServiceNotFoundError will be returned.
func (actor Actor) GetServiceByNameAndBrokerName(serviceName, serviceBrokerName string) (Service, Warnings, error) {
	filters := []ccv2.Filter{ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	}}

	if serviceBrokerName != "" {
		serviceBroker, warnings, err := actor.GetServiceBrokerByName(serviceBrokerName)
		if err != nil {
			return Service{}, warnings, err
		}

		brokerFilter := ccv2.Filter{
			Type:     constant.ServiceBrokerGUIDFilter,
			Operator: constant.EqualOperator,
			Values:   []string{serviceBroker.GUID},
		}
		filters = append(filters, brokerFilter)
	}

	services, warnings, err := actor.CloudControllerClient.GetServices(filters...)
	if err != nil {
		return Service{}, Warnings(warnings), err
	}

	if len(services) == 0 {
		return Service{}, Warnings(warnings), actionerror.ServiceNotFoundError{Name: serviceName}
	}

	if len(services) > 1 {
		return Service{}, Warnings(warnings), actionerror.DuplicateServiceError{Name: serviceName}
	}

	return Service(services[0]), Warnings(warnings), nil
}

func (actor Actor) getServiceByNameForSpace(serviceName, spaceGUID string) (Service, Warnings, error) {
	services, warnings, err := actor.CloudControllerClient.GetSpaceServices(spaceGUID, ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	})
	if err != nil {
		return Service{}, Warnings(warnings), err
	}

	if len(services) == 0 {
		return Service{}, Warnings(warnings), actionerror.ServiceNotFoundError{Name: serviceName}
	}

	return Service(services[0]), Warnings(warnings), nil
}

type ServicesWithPlans map[Service][]ServicePlan

func (actor Actor) GetServicesWithPlansForBroker(brokerGUID string) (ServicesWithPlans, Warnings, error) {
	var allWarnings Warnings
	services, warnings, err := actor.CloudControllerClient.GetServices(ccv2.Filter{
		Type:     constant.ServiceBrokerGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{brokerGUID},
	})
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return nil, allWarnings, err
	}

	servicesWithPlans := ServicesWithPlans{}
	for _, service := range services {
		servicePlans, warnings, err := actor.CloudControllerClient.GetServicePlans(ccv2.Filter{
			Type:     constant.ServiceGUIDFilter,
			Operator: constant.EqualOperator,
			Values:   []string{service.GUID},
		})
		allWarnings = append(allWarnings, Warnings(warnings)...)
		if err != nil {
			return nil, allWarnings, err
		}

		plansToReturn := []ServicePlan{}
		for _, plan := range servicePlans {
			plansToReturn = append(plansToReturn, ServicePlan(plan))
		}

		servicesWithPlans[Service(service)] = plansToReturn
	}

	return servicesWithPlans, allWarnings, nil
}
