package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServiceOfferingWithPlans ccv3.ServiceOfferingWithPlans

type MarketplaceFilter struct {
	SpaceGUID, ServiceOfferingName, ServiceBrokerName string
}

func (actor Actor) Marketplace(filter MarketplaceFilter) ([]ServiceOfferingWithPlans, Warnings, error) {
	var query []ccv3.Query

	if filter.SpaceGUID != "" {
		query = append(query, ccv3.Query{
			Key:    ccv3.SpaceGUIDFilter,
			Values: []string{filter.SpaceGUID},
		})
	}

	if filter.ServiceOfferingName != "" {
		query = append(query, ccv3.Query{
			Key:    ccv3.ServiceOfferingNamesFilter,
			Values: []string{filter.ServiceOfferingName},
		})
	}

	if filter.ServiceBrokerName != "" {
		query = append(query, ccv3.Query{
			Key:    ccv3.ServiceBrokerNamesFilter,
			Values: []string{filter.ServiceBrokerName},
		})
	}

	serviceOffering, warnings, err := actor.CloudControllerClient.GetServicePlansWithOfferings(query...)
	if err != nil {
		return nil, Warnings(warnings), err
	}

	result := make([]ServiceOfferingWithPlans, len(serviceOffering))
	for i := range serviceOffering {
		result[i] = ServiceOfferingWithPlans(serviceOffering[i])
	}

	return result, Warnings(warnings), nil
}
