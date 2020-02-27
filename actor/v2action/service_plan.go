package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

type ServicePlan ccv2.ServicePlan

func (actor Actor) GetServicePlan(servicePlanGUID string) (ServicePlan, Warnings, error) {
	servicePlan, warnings, err := actor.CloudControllerClient.GetServicePlan(servicePlanGUID)
	return ServicePlan(servicePlan), Warnings(warnings), err
}

// GetServicePlansForService returns a list of plans associated with the service and the broker if provided
func (actor Actor) GetServicePlansForService(serviceName, brokerName string) ([]ServicePlan, Warnings, error) {
	service, allWarnings, err := actor.GetServiceByNameAndBrokerName(serviceName, brokerName)
	if err != nil {
		return []ServicePlan{}, allWarnings, err
	}

	servicePlans, warnings, err := actor.CloudControllerClient.GetServicePlans(ccv2.Filter{
		Type:     constant.ServiceGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{service.GUID},
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []ServicePlan{}, allWarnings, err
	}

	var plansToReturn []ServicePlan
	for _, plan := range servicePlans {
		plansToReturn = append(plansToReturn, ServicePlan(plan))
	}

	return plansToReturn, allWarnings, nil
}

func (actor Actor) getServicePlanForServiceInSpace(servicePlanName, serviceName, spaceGUID, brokerName string) (ServicePlan, Warnings, error) {
	services, allWarnings, err := actor.getServicesByNameForSpace(serviceName, spaceGUID)
	if err != nil {
		return ServicePlan{}, allWarnings, err
	}

	if len(services) == 0 {
		return ServicePlan{}, Warnings(allWarnings), actionerror.ServiceNotFoundError{Name: serviceName}
	}

	if len(services) > 1 && brokerName == "" {
		return ServicePlan{}, Warnings(allWarnings), actionerror.DuplicateServiceError{Name: serviceName}
	}

	service, err := findServiceByBrokerName(services, serviceName, brokerName)
	if err != nil {
		return ServicePlan{}, allWarnings, err
	}

	plans, warnings, err := actor.CloudControllerClient.GetServicePlans(ccv2.Filter{
		Type:     constant.ServiceGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{service.GUID},
	})

	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ServicePlan{}, allWarnings, err
	}

	for _, plan := range plans {
		if servicePlanName == plan.Name {
			return ServicePlan(plan), allWarnings, err
		}
	}

	return ServicePlan{}, allWarnings, actionerror.ServicePlanNotFoundError{PlanName: servicePlanName, OfferingName: serviceName}
}

func findServiceByBrokerName(services []Service, serviceName, brokerName string) (Service, error) {
	if brokerName == "" && len(services) == 1 {
		return services[0], nil
	}

	for _, s := range services {
		if s.ServiceBrokerName == brokerName {
			return s, nil
		}
	}

	return Service{}, actionerror.ServiceAndBrokerCombinationNotFoundError{ServiceName: serviceName, BrokerName: brokerName}
}
