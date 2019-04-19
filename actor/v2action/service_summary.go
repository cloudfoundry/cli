package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

type ServiceSummary struct {
	Service
	Plans []ServicePlanSummary
}

func (actor Actor) GetServicesSummaries() ([]ServiceSummary, Warnings, error) {
	services, warnings, err := actor.CloudControllerClient.GetServices()
	if err != nil {
		return []ServiceSummary{}, Warnings(warnings), err
	}

	summaries, serviceWarnings, err := actor.createServiceSummaries(services, "", "")
	warnings = append(warnings, serviceWarnings...)
	return summaries, Warnings(warnings), err
}

func (actor Actor) GetServicesSummariesForSpace(spaceGUID string, organizationGUID string) ([]ServiceSummary, Warnings, error) {
	services, warnings, err := actor.CloudControllerClient.GetSpaceServices(spaceGUID)
	if err != nil {
		return []ServiceSummary{}, Warnings(warnings), err
	}

	summaries, summaryWarnings, err := actor.createServiceSummaries(services, organizationGUID, spaceGUID)
	warnings = append(warnings, summaryWarnings...)
	return summaries, Warnings(warnings), err
}

func (actor Actor) GetServiceSummaryByName(serviceName string) (ServiceSummary, Warnings, error) {
	services, warnings, err := actor.CloudControllerClient.GetServices(ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	})
	if err != nil {
		return ServiceSummary{}, Warnings(warnings), err
	}

	if len(services) == 0 {
		return ServiceSummary{}, Warnings(warnings), actionerror.ServiceNotFoundError{Name: serviceName}
	}

	plans, planWarnings, err := actor.getPlansForOneService(services[0])
	warnings = append(warnings, planWarnings...)
	if err != nil {
		return ServiceSummary{}, Warnings(warnings), err
	}

	summary, summaryWarnings, err := actor.createServiceSummary(services[0], plans, "", "")
	warnings = append(warnings, summaryWarnings...)
	if err != nil {
		return ServiceSummary{}, Warnings(warnings), err
	}

	return summary, Warnings(warnings), err
}

func (actor Actor) GetServiceSummaryForSpaceByName(spaceGUID, serviceName, organizationGUID string) (ServiceSummary, Warnings, error) {
	services, warnings, err := actor.CloudControllerClient.GetSpaceServices(spaceGUID, ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	})
	if err != nil {
		return ServiceSummary{}, Warnings(warnings), err
	}

	if len(services) == 0 {
		return ServiceSummary{}, Warnings(warnings), actionerror.ServiceNotFoundError{Name: serviceName}
	}

	plans, planWarnings, err := actor.getPlansForOneService(services[0])
	warnings = append(warnings, planWarnings...)
	if err != nil {
		return ServiceSummary{}, Warnings(warnings), err
	}

	summary, summaryWarnings, err := actor.createServiceSummary(services[0], plans, organizationGUID, spaceGUID)
	warnings = append(warnings, summaryWarnings...)
	if err != nil {
		return ServiceSummary{}, Warnings(warnings), err
	}

	return summary, Warnings(warnings), err
}

func (actor Actor) createServiceSummaries(services []ccv2.Service, organizationGUID, spaceGUID string) ([]ServiceSummary, ccv2.Warnings, error) {
	var serviceSummaries []ServiceSummary
	var warnings ccv2.Warnings

	plans, planWarnings, err := actor.getPlansForManyServices(services)
	warnings = append(warnings, planWarnings...)
	if err != nil {
		return []ServiceSummary{}, warnings, err
	}

	for _, service := range services {
		plansForThatService := actor.getPlansForService(service, plans)

		summary, summaryWarnings, err := actor.createServiceSummary(service, plansForThatService, organizationGUID, spaceGUID)
		warnings = append(warnings, summaryWarnings...)
		if err != nil {
			return []ServiceSummary{}, warnings, err
		}

		if len(summary.Plans) > 0 {
			serviceSummaries = append(serviceSummaries, summary)
		}
	}

	return serviceSummaries, warnings, nil
}

func (actor Actor) getPlansForOneService(service ccv2.Service) ([]ccv2.ServicePlan, ccv2.Warnings, error) {
	return actor.CloudControllerClient.GetServicePlans(ccv2.Filter{
		Type:     constant.ServiceGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{service.GUID},
	})
}

func (actor Actor) getPlansForManyServices(services []ccv2.Service) ([]ccv2.ServicePlan, ccv2.Warnings, error) {
	serviceGUIDs := []string{}
	for _, service := range services {
		serviceGUIDs = append(serviceGUIDs, service.GUID)
	}

	return actor.CloudControllerClient.GetServicePlans(ccv2.Filter{
		Type:     constant.ServiceGUIDFilter,
		Operator: constant.InOperator,
		Values:   serviceGUIDs,
	})
}

func (actor Actor) createServiceSummary(service ccv2.Service, plans []ccv2.ServicePlan, organizationGUID, spaceGUID string) (ServiceSummary, ccv2.Warnings, error) {
	planSummaries, warnings, err := actor.getPlanSummariesForService(service, plans, organizationGUID, spaceGUID)
	if err != nil {
		return ServiceSummary{}, warnings, err
	}

	return ServiceSummary{Service: Service(service), Plans: planSummaries}, warnings, nil
}

func (actor Actor) getPlanSummariesForService(service ccv2.Service, plans []ccv2.ServicePlan, organizationGUID, spaceGUID string) ([]ServicePlanSummary, ccv2.Warnings, error) {
	var err error
	var warnings ccv2.Warnings

	nonPublicPlans := []string{}
	for _, plan := range plans {
		if !plan.Public {
			nonPublicPlans = append(nonPublicPlans, plan.GUID)
		}
	}

	var broker ServiceBroker
	if spaceGUID != "" && len(nonPublicPlans) > 0 {
		var brokerWarnings Warnings
		broker, brokerWarnings, err = actor.GetServiceBrokerByName(service.ServiceBrokerName)
		warnings = append(warnings, brokerWarnings...)
		if err != nil {
			return []ServicePlanSummary{}, warnings, err
		}
	}

	var visibilities []ccv2.ServicePlanVisibility

	if len(nonPublicPlans) > 0 {
		var visibilityWarnings ccv2.Warnings

		visibilities, visibilityWarnings, err = actor.getPlanVisibilitiesForOrg(nonPublicPlans, organizationGUID)
		warnings = append(warnings, visibilityWarnings...)
		if err != nil {
			return []ServicePlanSummary{}, warnings, err
		}
	}

	planSummaries := actor.getSummariesForVisiblePlans(plans, broker, visibilities, spaceGUID)

	return planSummaries, warnings, nil
}

func (actor Actor) getSummariesForVisiblePlans(plans []ccv2.ServicePlan, broker ServiceBroker, visibilities []ccv2.ServicePlanVisibility, spaceGUID string) []ServicePlanSummary {
	var planSummaries []ServicePlanSummary
	for _, plan := range plans {
		if plan.Public || (spaceGUID != "" && broker.SpaceGUID == spaceGUID) {
			planSummaries = append(planSummaries, ServicePlanSummary{ServicePlan: ServicePlan(plan)})
		} else {
			visibleInOrg := false
			for _, visibility := range visibilities {
				if visibility.ServicePlanGUID == plan.GUID {
					visibleInOrg = true
					break
				}
			}

			if visibleInOrg {
				planSummaries = append(planSummaries, ServicePlanSummary{ServicePlan: ServicePlan(plan)})
			}
		}
	}

	return planSummaries
}

func (actor Actor) getPlanVisibilitiesForOrg(plans []string, organizationGUID string) ([]ccv2.ServicePlanVisibility, ccv2.Warnings, error) {
	return actor.CloudControllerClient.GetServicePlanVisibilities(
		ccv2.Filter{
			Type:     constant.ServicePlanGUIDFilter,
			Operator: constant.InOperator,
			Values:   plans,
		},
		ccv2.Filter{
			Type:     constant.OrganizationGUIDFilter,
			Operator: constant.EqualOperator,
			Values:   []string{organizationGUID},
		},
	)
}

func (actor Actor) getPlansForService(service ccv2.Service, plans []ccv2.ServicePlan) []ccv2.ServicePlan {
	plansForThatService := []ccv2.ServicePlan{}
	for _, plan := range plans {
		if plan.ServiceGUID == service.GUID {
			plansForThatService = append(plansForThatService, plan)
		}
	}
	return plansForThatService
}
