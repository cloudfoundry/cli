package composite

import (
	"code.cloudfoundry.org/cli/actor/v2action"
)

//go:generate counterfeiter . ServiceActor

type ServiceActor interface {
	GetServicesWithPlansForBroker(brokerGUID string) (v2action.ServicesWithPlans, v2action.Warnings, error)
}

//go:generate counterfeiter . BrokerActor

type BrokerActor interface {
	GetServiceBrokerByName(brokerName string) (v2action.ServiceBroker, v2action.Warnings, error)
	GetServiceBrokers() ([]v2action.ServiceBroker, v2action.Warnings, error)
}

//go:generate counterfeiter . OrganizationActor

type OrganizationActor interface {
	GetOrganization(organizationGUID string) (v2action.Organization, v2action.Warnings, error)
}

//go:generate counterfeiter . VisibilityActor

type VisibilityActor interface {
	GetServicePlanVisibilities(planGUID string) ([]v2action.ServicePlanVisibility, v2action.Warnings, error)
}

type ServiceBrokerSummaryCompositeActor struct {
	ServiceActor    ServiceActor
	BrokerActor     BrokerActor
	OrgActor        OrganizationActor
	VisibilityActor VisibilityActor
}

func (c *ServiceBrokerSummaryCompositeActor) GetServiceBrokerSummaries(brokerName string, service string, organization string) ([]v2action.ServiceBrokerSummary, v2action.Warnings, error) {
	var (
		err      error
		warnings v2action.Warnings
		brokers  []v2action.ServiceBroker
	)

	if brokerName != "" {
		var broker v2action.ServiceBroker

		broker, warnings, err = c.BrokerActor.GetServiceBrokerByName(brokerName)
		brokers = append(brokers, broker)
	} else {
		brokers, warnings, err = c.BrokerActor.GetServiceBrokers()
	}
	if err != nil {
		return nil, warnings, err
	}

	brokerSummaries, brokerWarnings, err := c.fetchBrokerSummaries(brokers)
	warnings = append(warnings, brokerWarnings...)
	if err != nil {
		return nil, warnings, err
	}

	return brokerSummaries, warnings, nil
}

func (c *ServiceBrokerSummaryCompositeActor) fetchBrokerSummaries(brokers []v2action.ServiceBroker) ([]v2action.ServiceBrokerSummary, v2action.Warnings, error) {
	var (
		brokerSummaries []v2action.ServiceBrokerSummary
		warnings        v2action.Warnings
	)

	for _, broker := range brokers {
		brokerSummary, brokerWarnings, err := c.fetchBrokerSummary(v2action.ServiceBroker(broker))
		warnings = append(warnings, brokerWarnings...)
		if err != nil {
			return nil, warnings, err
		}
		brokerSummaries = append(brokerSummaries, brokerSummary)
	}

	return brokerSummaries, warnings, nil
}

func (c *ServiceBrokerSummaryCompositeActor) fetchBrokerSummary(broker v2action.ServiceBroker) (v2action.ServiceBrokerSummary, v2action.Warnings, error) {
	servicesWithPlans, warnings, err := c.ServiceActor.GetServicesWithPlansForBroker(broker.GUID)
	if err != nil {
		return v2action.ServiceBrokerSummary{}, warnings, err
	}

	var services []v2action.ServiceSummary
	for service, servicePlans := range servicesWithPlans {
		serviceSummary, serviceWarnings, err := c.fetchServiceSummary(service, servicePlans)
		warnings = append(warnings, serviceWarnings...)
		if err != nil {
			return v2action.ServiceBrokerSummary{}, warnings, err
		}

		services = append(services, serviceSummary)
	}

	return v2action.ServiceBrokerSummary{
		ServiceBroker: v2action.ServiceBroker(broker),
		Services:      services,
	}, warnings, nil
}

func (c *ServiceBrokerSummaryCompositeActor) fetchServiceSummary(service v2action.Service, servicePlans []v2action.ServicePlan) (v2action.ServiceSummary, v2action.Warnings, error) {
	var warnings v2action.Warnings
	var servicePlanSummaries []v2action.ServicePlanSummary
	for _, plan := range servicePlans {
		var visibleTo []string
		if !plan.Public {
			serviceVisibilities, visibilityWarnings, visErr := c.VisibilityActor.GetServicePlanVisibilities(plan.GUID)
			warnings = append(warnings, visibilityWarnings...)
			if visErr != nil {
				return v2action.ServiceSummary{}, v2action.Warnings(warnings), visErr
			}

			for _, serviceVisibility := range serviceVisibilities {
				org, orgWarnings, orgsErr := c.OrgActor.GetOrganization(serviceVisibility.OrganizationGUID)
				warnings = append(warnings, orgWarnings...)
				if orgsErr != nil {
					return v2action.ServiceSummary{}, v2action.Warnings(warnings), orgsErr
				}

				visibleTo = append(visibleTo, org.Name)
			}
		}

		servicePlanSummaries = append(servicePlanSummaries,
			v2action.ServicePlanSummary{
				ServicePlan: v2action.ServicePlan(plan),
				VisibleTo:   visibleTo,
			})
	}

	return v2action.ServiceSummary{
		Service: v2action.Service(service),
		Plans:   servicePlanSummaries,
	}, v2action.Warnings(warnings), nil
}
