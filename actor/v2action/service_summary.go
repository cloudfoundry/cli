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

	summaries, serviceWarnings, err := actor.createServiceSummaries(services)
	warnings = append(warnings, serviceWarnings...)
	return summaries, Warnings(warnings), err
}

func (actor Actor) GetServicesSummariesForSpace(spaceGUID string) ([]ServiceSummary, Warnings, error) {
	services, warnings, err := actor.CloudControllerClient.GetSpaceServices(spaceGUID)
	if err != nil {
		return []ServiceSummary{}, Warnings(warnings), err
	}

	summaries, summaryWarnings, err := actor.createServiceSummaries(services)
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

	summary, summaryWarnings, err := actor.createServiceSummary(services[0])
	warnings = append(warnings, summaryWarnings...)
	if err != nil {
		return ServiceSummary{}, Warnings(warnings), err
	}

	return summary, Warnings(warnings), err
}

func (actor Actor) GetServiceSummaryForSpaceByName(spaceGUID, serviceName string) (ServiceSummary, Warnings, error) {
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

	summary, summaryWarnings, err := actor.createServiceSummary(services[0])
	warnings = append(warnings, summaryWarnings...)
	if err != nil {
		return ServiceSummary{}, Warnings(warnings), err
	}

	return summary, Warnings(warnings), err
}

func (actor Actor) createServiceSummaries(services []ccv2.Service) ([]ServiceSummary, ccv2.Warnings, error) {
	var serviceSummaries []ServiceSummary
	var warnings ccv2.Warnings

	for _, service := range services {
		summary, summaryWarnings, err := actor.createServiceSummary(service)
		warnings = append(warnings, summaryWarnings...)
		if err != nil {
			return []ServiceSummary{}, warnings, err
		}

		serviceSummaries = append(serviceSummaries, summary)
	}

	return serviceSummaries, warnings, nil
}

func (actor Actor) createServiceSummary(service ccv2.Service) (ServiceSummary, ccv2.Warnings, error) {
	planSummaries, warnings, err := actor.getPlanSummariesForService(service)
	if err != nil {
		return ServiceSummary{}, warnings, err
	}

	return ServiceSummary{Service: Service(service), Plans: planSummaries}, warnings, nil
}

func (actor Actor) getPlanSummariesForService(service ccv2.Service) ([]ServicePlanSummary, ccv2.Warnings, error) {
	plans, warnings, err := actor.CloudControllerClient.GetServicePlans(ccv2.Filter{
		Type:     constant.ServiceGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{service.GUID},
	})
	if err != nil {
		return []ServicePlanSummary{}, warnings, err
	}

	var planSummaries []ServicePlanSummary
	for _, plan := range plans {
		planSummaries = append(planSummaries, ServicePlanSummary{ServicePlan: ServicePlan(plan)})
	}

	return planSummaries, warnings, nil
}
