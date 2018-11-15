package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

type ServiceBroker ccv2.ServiceBroker

func (actor Actor) CreateServiceBroker(serviceBrokerName, username, password, brokerURI, spaceGUID string) (ServiceBroker, Warnings, error) {
	serviceBroker, warnings, err := actor.CloudControllerClient.CreateServiceBroker(serviceBrokerName, username, password, brokerURI, spaceGUID)
	return ServiceBroker(serviceBroker), Warnings(warnings), err
}
