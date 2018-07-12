package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

type ServiceBrokerSummary struct {
	ServiceBroker
	Services []ServiceSummary
}

func (actor Actor) GetServiceBrokerSummaries(broker string, service string, organization string) ([]ServiceBrokerSummary, Warnings, error) {
	var filters []ccv2.Filter
	if broker != "" {
		filters = append(filters, ccv2.Filter{
			Type:     constant.NameFilter,
			Operator: constant.EqualOperator,
			Values:   []string{broker},
		})
	}
	brokers, ccWarnings, brokersErr := actor.CloudControllerClient.GetServiceBrokers(filters...)

	warnings := Warnings(ccWarnings)
	var brokerSummaries []ServiceBrokerSummary
	for _, broker := range brokers {
		brokerSummary, brokerWarnings, err := actor.fetchBrokerSummary(broker)
		warnings = append(warnings, brokerWarnings...)
		if err != nil {
			return nil, warnings, err
		}
		brokerSummaries = append(brokerSummaries, brokerSummary)
	}

	return brokerSummaries, warnings, brokersErr
}

func (actor Actor) fetchBrokerSummary(broker ccv2.ServiceBroker) (ServiceBrokerSummary, Warnings, error) {
	ccServices, ccWarnings, servicesErr := actor.CloudControllerClient.GetServices(ccv2.Filter{
		Type:     constant.ServiceBrokerGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{broker.GUID},
	})

	warnings := Warnings(ccWarnings)

	if servicesErr != nil {
		return ServiceBrokerSummary{}, warnings, servicesErr
	}

	var services []ServiceSummary
	for _, service := range ccServices {
		serviceSummary, serviceWarnings, err := actor.fetchServiceSummary(service)
		warnings = append(warnings, serviceWarnings...)
		if err != nil {
			return ServiceBrokerSummary{}, warnings, err
		}

		services = append(services, serviceSummary)
	}

	return ServiceBrokerSummary{
		ServiceBroker: ServiceBroker(broker),
		Services:      services,
	}, warnings, servicesErr
}

func (actor Actor) fetchServiceSummary(service ccv2.Service) (ServiceSummary, Warnings, error) {
	ccServicePlans, warnings, plansErr := actor.CloudControllerClient.GetServicePlans(ccv2.Filter{
		Type:     constant.ServiceGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{service.GUID},
	})

	if plansErr != nil {
		return ServiceSummary{}, Warnings(warnings), plansErr
	}

	var servicePlanSummaries []ServicePlanSummary
	for _, plan := range ccServicePlans {
		var visibleTo []string
		if !plan.Public {
			serviceVisibilities, visibilityWarnings, visErr := actor.CloudControllerClient.GetServicePlanVisibilities(ccv2.Filter{
				Type:     constant.ServicePlanGUIDFilter,
				Operator: constant.EqualOperator,
				Values:   []string{plan.GUID},
			})
			warnings = append(warnings, visibilityWarnings...)

			if visErr != nil {
				return ServiceSummary{}, Warnings(warnings), visErr
			}

			for _, serviceVisibility := range serviceVisibilities {
				org, orgWarnings, orgsErr := actor.GetOrganization(serviceVisibility.OrganizationGUID)
				warnings = append(warnings, orgWarnings...)

				if orgsErr != nil {
					return ServiceSummary{}, Warnings(warnings), orgsErr
				}

				visibleTo = append(visibleTo, org.Name)
			}
		}

		servicePlanSummaries = append(servicePlanSummaries,
			ServicePlanSummary{
				ServicePlan: ServicePlan(plan),
				VisibleTo:   visibleTo,
			})
	}

	return ServiceSummary{
		Service: Service(service),
		Plans:   servicePlanSummaries,
	}, Warnings(warnings), plansErr
}
