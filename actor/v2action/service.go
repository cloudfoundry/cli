package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

type Service ccv2.Service

// ServicesWithPlans is an association between a Service and the plans it offers.
type ServicesWithPlans map[Service][]ServicePlan

type Filter ccv2.Filter

// GetService fetches a service by GUID.
func (actor Actor) GetService(serviceGUID string) (Service, Warnings, error) {
	service, warnings, err := actor.CloudControllerClient.GetService(serviceGUID)
	return Service(service), Warnings(warnings), err
}

// GetServiceByNameAndBrokerName returns services based on the name and broker provided.
// If a service broker name is provided, only services for that broker will be returned.
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

func (actor Actor) getServiceByNameForSpace(serviceName, spaceGUID, brokerGUID string) (Service, Warnings, error) {
	var filters []ccv2.Filter

	filters = append(filters, ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	})

	if brokerGUID != "" {
		filters = append(filters, ccv2.Filter{
			Type:     constant.ServiceBrokerGUIDFilter,
			Operator: constant.EqualOperator,
			Values:   []string{brokerGUID},
		})
	}

	services, warnings, err := actor.CloudControllerClient.GetSpaceServices(spaceGUID, filters...)

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

// GetServicesWithPlans returns a map of Services to ServicePlans for a particular broker.
// A particular service with associated plans from a broker can be fetched by additionally providing
// a service label filter.
func (actor Actor) GetServicesWithPlans(filters ...Filter) (ServicesWithPlans, Warnings, error) {
	ccv2Filters := []ccv2.Filter{}
	for _, f := range filters {
		ccv2Filters = append(ccv2Filters, ccv2.Filter(f))
	}

	var allWarnings Warnings

	services, warnings, err := actor.CloudControllerClient.GetServices(ccv2Filters...)
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

// ServiceExistsWithName returns true if there is an Organization with the
// provided name, otherwise false.
func (actor Actor) ServiceExistsWithName(serviceName string) (bool, Warnings, error) {
	services, warnings, err := actor.CloudControllerClient.GetServices(ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	})
	if err != nil {
		return false, Warnings(warnings), err
	}

	if len(services) == 0 {
		return false, Warnings(warnings), nil
	}

	return true, Warnings(warnings), nil
}

func (actor Actor) PurgeServiceOffering(service Service) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.DeleteService(service.GUID, true)
	if err != nil {
		return Warnings(warnings), err
	}

	return Warnings(warnings), nil
}
