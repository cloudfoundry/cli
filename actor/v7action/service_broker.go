package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServiceBroker = ccv3.ServiceBroker
type ServiceBrokerModel = ccv3.ServiceBrokerModel

func (actor Actor) GetServiceBrokers() ([]ServiceBroker, Warnings, error) {
	serviceBrokers, warnings, err := actor.CloudControllerClient.GetServiceBrokers()
	if err != nil {
		return nil, Warnings(warnings), err
	}

	return serviceBrokers, Warnings(warnings), nil
}

func (actor Actor) GetServiceBrokerByName(serviceBrokerName string) (ServiceBroker, Warnings, error) {
	query := []ccv3.Query{
		{
			Key:    ccv3.NameFilter,
			Values: []string{serviceBrokerName},
		},
	}
	serviceBrokers, warnings, err := actor.CloudControllerClient.GetServiceBrokers(query...)
	if err != nil {
		return ServiceBroker{}, Warnings(warnings), err
	}

	if len(serviceBrokers) == 0 {
		return ServiceBroker{}, Warnings(warnings), actionerror.ServiceBrokerNotFoundError{Name: serviceBrokerName}
	}

	return serviceBrokers[0], Warnings(warnings), nil
}

func (actor Actor) CreateServiceBroker(model ServiceBrokerModel) (Warnings, error) {
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

func (actor Actor) UpdateServiceBroker(serviceBrokerGUID string, model ServiceBrokerModel) (Warnings, error) {
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
