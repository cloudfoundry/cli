package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServiceBroker ccv3.ServiceBroker

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
