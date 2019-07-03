package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServiceBroker = ccv3.ServiceBroker

func (actor Actor) GetServiceBrokers() ([]ServiceBroker, Warnings, error) {
	serviceBrokers, warnings, err := actor.CloudControllerClient.GetServiceBrokers()
	if err != nil {
		return nil, Warnings(warnings), err
	}

	return serviceBrokers, Warnings(warnings), nil
}

func (actor Actor) CreateServiceBroker(name, username, password, url, spaceGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.CreateServiceBroker(name, username, password, url, spaceGUID)
	return Warnings(warnings), err
}
