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

// FIXME: please, put me in a single parameter object.
func (actor Actor) CreateServiceBroker(name, username, password, url, spaceGUID string) (Warnings, error) {
	allWarnings := Warnings{}

	jobURL, warnings, err := actor.CloudControllerClient.CreateServiceBroker(name, username, password, url, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	warnings, err = actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, warnings...)
	return allWarnings, err
}
