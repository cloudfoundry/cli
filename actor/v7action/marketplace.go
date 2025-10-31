package v7action

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
)

type ServiceOfferingWithPlans ccv3.ServiceOfferingWithPlans

type MarketplaceFilter struct {
	SpaceGUID, ServiceOfferingName, ServiceBrokerName string
	ShowUnavailable                                   bool
}

func (actor Actor) Marketplace(filter MarketplaceFilter) ([]ServiceOfferingWithPlans, Warnings, error) {
	query := []ccv3.Query{{Key: ccv3.PerPage, Values: []string{ccv3.MaxPerPage}}}

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

	if !filter.ShowUnavailable {
		query = append(query, ccv3.Query{
			Key:    ccv3.AvailableFilter,
			Values: []string{"true"},
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
