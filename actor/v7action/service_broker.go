package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServiceBroker ccv3.ServiceBroker

type ServiceBrokerCredentials ccv3.ServiceBrokerCredentials

func (actor Actor) GetServiceBrokers() ([]ServiceBroker, Warnings, error) {
	ccv3ServiceBrokers, warnings, err := actor.CloudControllerClient.GetServiceBrokers()
	if err != nil {
		return nil, Warnings(warnings), err
	}

	var serviceBrokers []ServiceBroker
	for _, broker := range ccv3ServiceBrokers {
		serviceBrokers = append(serviceBrokers, ServiceBroker(broker))
	}

	return serviceBrokers, Warnings(warnings), nil
}

func (actor Actor) CreateServiceBroker(credentials ServiceBrokerCredentials) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.CreateServiceBroker(ccv3.ServiceBrokerCredentials(credentials))
	return Warnings(warnings), err
}
