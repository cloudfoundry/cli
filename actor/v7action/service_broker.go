package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServiceBrokerCredentials = ccv3.ServiceBrokerCredentials
type ServiceBrokerCredentialsData = ccv3.ServiceBrokerCredentialsData
type ServiceBroker = ccv3.ServiceBroker

func (actor Actor) GetServiceBrokers() ([]ServiceBroker, Warnings, error) {
	serviceBrokers, warnings, err := actor.CloudControllerClient.GetServiceBrokers()
	if err != nil {
		return nil, Warnings(warnings), err
	}

	return serviceBrokers, Warnings(warnings), nil
}

func (actor Actor) CreateServiceBroker(serviceBroker ServiceBroker) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.CreateServiceBroker(serviceBroker)
	return Warnings(warnings), err
}
