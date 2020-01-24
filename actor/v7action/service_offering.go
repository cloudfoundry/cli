package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServiceOffering ccv3.ServiceOffering

func (actor Actor) GetServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName string) (ServiceOffering, Warnings, error) {
	query := []ccv3.Query{{Key: ccv3.NameFilter, Values: []string{serviceOfferingName}}}
	if serviceBrokerName != "" {
		query = append(query, ccv3.Query{Key: ccv3.ServiceBrokerNamesFilter, Values: []string{serviceBrokerName}})
	}

	serviceOfferings, warnings, err := actor.CloudControllerClient.GetServiceOfferings(query...)
	if err != nil {
		return ServiceOffering{}, Warnings(warnings), err
	}

	switch len(serviceOfferings) {
	case 0:
		return ServiceOffering{}, Warnings(warnings), actionerror.ServiceNotFoundError{Name: serviceOfferingName, Broker: serviceBrokerName}
	case 1:
		return ServiceOffering(serviceOfferings[0]), Warnings(warnings), nil
	default:
		return ServiceOffering{}, Warnings(warnings), actionerror.DuplicateServiceError{Name: serviceOfferingName}
	}
}
