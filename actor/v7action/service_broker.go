package v7action

import (
	"code.cloudfoundry.org/cli/v7/actor/actionerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v7/resources"
)

func (actor Actor) GetServiceBrokers() ([]resources.ServiceBroker, Warnings, error) {
	serviceBrokers, warnings, err := actor.CloudControllerClient.GetServiceBrokers()
	if err != nil {
		return nil, Warnings(warnings), err
	}

	return serviceBrokers, Warnings(warnings), nil
}

func (actor Actor) GetServiceBrokerByName(serviceBrokerName string) (resources.ServiceBroker, Warnings, error) {
	query := []ccv3.Query{
		{
			Key:    ccv3.NameFilter,
			Values: []string{serviceBrokerName},
		},
	}
	serviceBrokers, warnings, err := actor.CloudControllerClient.GetServiceBrokers(query...)
	if err != nil {
		return resources.ServiceBroker{}, Warnings(warnings), err
	}

	if len(serviceBrokers) == 0 {
		return resources.ServiceBroker{}, Warnings(warnings), actionerror.ServiceBrokerNotFoundError{Name: serviceBrokerName}
	}

	return serviceBrokers[0], Warnings(warnings), nil
}

func (actor Actor) CreateServiceBroker(model resources.ServiceBroker) (Warnings, error) {
	allWarnings := Warnings{}

	jobURL, warnings, err := actor.CloudControllerClient.CreateServiceBroker(model)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	warnings, err = actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, warnings...)
	return allWarnings, err
}

func (actor Actor) UpdateServiceBroker(serviceBrokerGUID string, model resources.ServiceBroker) (Warnings, error) {
	allWarnings := Warnings{}

	jobURL, warnings, err := actor.CloudControllerClient.UpdateServiceBroker(serviceBrokerGUID, model)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	warnings, err = actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, warnings...)
	return allWarnings, err
}

func (actor Actor) DeleteServiceBroker(serviceBrokerGUID string) (Warnings, error) {
	allWarnings := Warnings{}

	jobURL, warnings, err := actor.CloudControllerClient.DeleteServiceBroker(serviceBrokerGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	warnings, err = actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, warnings...)

	return allWarnings, err
}
