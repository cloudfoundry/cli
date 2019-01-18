package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

type ServiceBroker ccv2.ServiceBroker

// CreateServiceBroker returns a ServiceBroker and any warnings or errors
func (actor Actor) CreateServiceBroker(serviceBrokerName, username, password, brokerURI, spaceGUID string) (ServiceBroker, Warnings, error) {
	serviceBroker, warnings, err := actor.CloudControllerClient.CreateServiceBroker(serviceBrokerName, username, password, brokerURI, spaceGUID)
	return ServiceBroker(serviceBroker), Warnings(warnings), err
}

func (actor Actor) GetServiceBrokers() ([]ServiceBroker, Warnings, error) {
	brokers, warnings, err := actor.CloudControllerClient.GetServiceBrokers()
	if err != nil {
		return nil, Warnings(warnings), err
	}

	var brokersToReturn []ServiceBroker
	for _, b := range brokers {
		brokersToReturn = append(brokersToReturn, ServiceBroker(b))
	}

	return brokersToReturn, Warnings(warnings), nil
}

// GetServiceBrokerByName returns a ServiceBroker and any warnings or errors
func (actor Actor) GetServiceBrokerByName(brokerName string) (ServiceBroker, Warnings, error) {
	serviceBrokers, warnings, err := actor.CloudControllerClient.GetServiceBrokers(ccv2.Filter{
		Type:     constant.NameFilter,
		Operator: constant.EqualOperator,
		Values:   []string{brokerName},
	})

	if err != nil {
		return ServiceBroker{}, Warnings(warnings), err
	}

	if len(serviceBrokers) == 0 {
		return ServiceBroker{}, Warnings(warnings), actionerror.ServiceBrokerNotFoundError{Name: brokerName}
	}

	return ServiceBroker(serviceBrokers[0]), Warnings(warnings), err
}
