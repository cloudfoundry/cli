package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

func (actor Actor) GetServicePlanByNameOfferingAndBroker(servicePlanName, serviceOfferingName, serviceBrokerName string) (resources.ServicePlan, Warnings, error) {
	query := []ccv3.Query{{Key: ccv3.NameFilter, Values: []string{servicePlanName}}}
	if serviceBrokerName != "" {
		query = append(query, ccv3.Query{Key: ccv3.ServiceBrokerNamesFilter, Values: []string{serviceBrokerName}})
	}
	if serviceOfferingName != "" {
		query = append(query, ccv3.Query{Key: ccv3.ServiceOfferingNamesFilter, Values: []string{serviceOfferingName}})
	}

	servicePlans, warnings, err := actor.CloudControllerClient.GetServicePlans(query...)
	if err != nil {
		return resources.ServicePlan{}, Warnings(warnings), err
	}

	switch len(servicePlans) {
	case 0:
		return resources.ServicePlan{}, Warnings(warnings), actionerror.ServicePlanNotFoundError{PlanName: servicePlanName}
	case 1:
		return resources.ServicePlan(servicePlans[0]), Warnings(warnings), nil
	default:
		return resources.ServicePlan{}, Warnings(warnings), actionerror.DuplicateServicePlanError{
			Name:                servicePlanName,
			ServiceOfferingName: serviceOfferingName,
			ServiceBrokerName:   serviceBrokerName,
		}
	}
}
