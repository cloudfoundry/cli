package v7action

import (
	"sort"
)

type ServicePlanAccess struct {
	BrokerName          string
	ServiceOfferingName string
	ServicePlanName     string
	VisibilityType      string
	VisibilityDetails   []string
}

func (actor *Actor) GetServiceAccess(broker, service, organization string) ([]ServicePlanAccess, Warnings, error) {
	allWarnings := Warnings{}

	type offeringDetails struct{ offeringName, brokerName string }
	offerings := make(map[string]offeringDetails)

	serviceOfferings, warnings, err := actor.CloudControllerClient.GetServiceOfferings()
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	for _, o := range serviceOfferings {
		offerings[o.GUID] = offeringDetails{
			offeringName: o.Name,
			brokerName:   o.ServiceBrokerName,
		}
	}

	plans, warnings, err := actor.CloudControllerClient.GetServicePlans()
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	var result []ServicePlanAccess
	for _, plan := range plans {
		if offering, ok := offerings[plan.ServiceOfferingGUID]; ok {

			visibilityDetails, warnings, err := actor.getServicePlanVisibilityDetails(ServicePlan(plan))
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				return nil, allWarnings, err
			}

			result = append(result, ServicePlanAccess{
				ServicePlanName:     plan.Name,
				VisibilityType:      plan.VisibilityType,
				VisibilityDetails:   visibilityDetails,
				ServiceOfferingName: offering.offeringName,
				BrokerName:          offering.brokerName,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].BrokerName != result[j].BrokerName {
			return result[i].BrokerName < result[j].BrokerName
		}
		if result[i].ServiceOfferingName != result[j].ServiceOfferingName {
			return result[i].ServiceOfferingName < result[j].ServiceOfferingName
		}
		return result[i].ServicePlanName < result[j].ServicePlanName
	})

	return result, allWarnings, err
}

func (actor *Actor) getServicePlanVisibilityDetails(plan ServicePlan) (names []string, warnings Warnings, err error) {
	if plan.VisibilityType == "organization" {
		result, vwarn, err := actor.CloudControllerClient.GetServicePlanVisibility(plan.GUID)
		warnings = Warnings(vwarn)
		if err != nil {
			return nil, warnings, err
		}

		for _, organization := range result.Organizations {
			names = append(names, organization.Name)
		}
	}

	return names, warnings, nil
}
