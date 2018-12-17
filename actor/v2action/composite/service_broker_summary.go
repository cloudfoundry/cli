package composite

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

//go:generate counterfeiter . ServiceActor
type ServiceActor interface {
	GetServicesWithPlansForBroker(brokerGUID string) (v2action.ServicesWithPlans, v2action.Warnings, error)
}

//go:generate counterfeiter . OrganizationActor
type OrganizationActor interface {
	GetOrganization(organizationGUID string) (v2action.Organization, v2action.Warnings, error)
}

type ServiceBrokerSummaryCompositeActor struct {
	CloudControllerClient v2action.CloudControllerClient
	ServiceActor          ServiceActor
	OrgActor              OrganizationActor
}

func (w *ServiceBrokerSummaryCompositeActor) GetServiceBrokerSummaries(broker string, service string, organization string) ([]v2action.ServiceBrokerSummary, v2action.Warnings, error) {
	var filters []ccv2.Filter
	if broker != "" {
		filters = append(filters, ccv2.Filter{
			Type:     constant.NameFilter,
			Operator: constant.EqualOperator,
			Values:   []string{broker},
		})
	}
	brokers, ccWarnings, brokersErr := w.CloudControllerClient.GetServiceBrokers(filters...)

	warnings := v2action.Warnings(ccWarnings)
	var brokerSummaries []v2action.ServiceBrokerSummary
	for _, broker := range brokers {
		brokerSummary, brokerWarnings, err := w.fetchBrokerSummary(broker)
		warnings = append(warnings, brokerWarnings...)
		if err != nil {
			return nil, warnings, err
		}
		brokerSummaries = append(brokerSummaries, brokerSummary)
	}

	return brokerSummaries, warnings, brokersErr
}

func (w *ServiceBrokerSummaryCompositeActor) fetchBrokerSummary(broker ccv2.ServiceBroker) (v2action.ServiceBrokerSummary, v2action.Warnings, error) {
	servicesWithPlans, warnings, err := w.ServiceActor.GetServicesWithPlansForBroker(broker.GUID)
	if err != nil {
		return v2action.ServiceBrokerSummary{}, warnings, err
	}

	var services []v2action.ServiceSummary
	for service, servicePlans := range servicesWithPlans {
		serviceSummary, serviceWarnings, err := w.fetchServiceSummary(service, servicePlans)
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

func (w *ServiceBrokerSummaryCompositeActor) fetchServiceSummary(service v2action.Service, servicePlans []v2action.ServicePlan) (v2action.ServiceSummary, v2action.Warnings, error) {
	var warnings v2action.Warnings
	var servicePlanSummaries []v2action.ServicePlanSummary
	for _, plan := range servicePlans {
		var visibleTo []string
		if !plan.Public {
			serviceVisibilities, visibilityWarnings, visErr := w.CloudControllerClient.GetServicePlanVisibilities(ccv2.Filter{
				Type:     constant.ServicePlanGUIDFilter,
				Operator: constant.EqualOperator,
				Values:   []string{plan.GUID},
			})
			warnings = append(warnings, visibilityWarnings...)

			if visErr != nil {
				return v2action.ServiceSummary{}, v2action.Warnings(warnings), visErr
			}

			for _, serviceVisibility := range serviceVisibilities {
				org, orgWarnings, orgsErr := w.OrgActor.GetOrganization(serviceVisibility.OrganizationGUID)
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
