package v7action

import (
	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/railway"
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
		{Key: ccv3.NameFilter, Values: []string{serviceBrokerName}},
		{Key: ccv3.PerPage, Values: []string{"1"}},
		{Key: ccv3.Page, Values: []string{"1"}},
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
	return actor.performAndPoll(func() (ccv3.JobURL, ccv3.Warnings, error) {
		return actor.CloudControllerClient.CreateServiceBroker(model)
	})
}

func (actor Actor) UpdateServiceBroker(serviceBrokerGUID string, model resources.ServiceBroker) (Warnings, error) {
	return actor.performAndPoll(func() (ccv3.JobURL, ccv3.Warnings, error) {
		return actor.CloudControllerClient.UpdateServiceBroker(serviceBrokerGUID, model)
	})
}

func (actor Actor) DeleteServiceBroker(serviceBrokerGUID string) (Warnings, error) {
	return actor.performAndPoll(func() (ccv3.JobURL, ccv3.Warnings, error) {
		return actor.CloudControllerClient.DeleteServiceBroker(serviceBrokerGUID)
	})
}

func (actor Actor) performAndPoll(callback func() (ccv3.JobURL, ccv3.Warnings, error)) (Warnings, error) {
	var jobURL ccv3.JobURL

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = callback()
			return
		},
		func() (ccv3.Warnings, error) {
			return actor.CloudControllerClient.PollJob(jobURL)
		},
	)
	return Warnings(warnings), err
}
