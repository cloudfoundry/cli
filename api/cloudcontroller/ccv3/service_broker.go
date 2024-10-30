package ccv3

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v8/resources"
)

// CreateServiceBroker registers a new service broker.
func (client *Client) CreateServiceBroker(serviceBroker resources.ServiceBroker) (JobURL, Warnings, error) {
	serviceBroker.CredentialsType = resources.ServiceBrokerBasicCredentials

	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.PostServiceBrokerRequest,
		RequestBody: serviceBroker,
	})

	return jobURL, warnings, err
}

// DeleteServiceBroker deletes a named service broker
func (client *Client) DeleteServiceBroker(serviceBrokerGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteServiceBrokerRequest,
		URIParams:   internal.Params{"service_broker_guid": serviceBrokerGUID},
	})

	return jobURL, warnings, err
}

// GetServiceBrokers lists service brokers.
func (client *Client) GetServiceBrokers(query ...Query) ([]resources.ServiceBroker, Warnings, error) {
	var result []resources.ServiceBroker

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServiceBrokersRequest,
		Query:        query,
		ResponseBody: resources.ServiceBroker{},
		AppendToList: func(item interface{}) error {
			result = append(result, item.(resources.ServiceBroker))
			return nil
		},
	})

	return result, warnings, err
}

// UpdateServiceBroker updates an existing service broker.
func (client *Client) UpdateServiceBroker(serviceBrokerGUID string, serviceBroker resources.ServiceBroker) (JobURL, Warnings, error) {
	if (serviceBroker.Username == "" && serviceBroker.Password != "") || (serviceBroker.Username != "" && serviceBroker.Password == "") {
		return "", nil, errors.New("Incorrect usage: both username and password must be defined in order to do an update")
	}

	if serviceBroker.Username != "" && serviceBroker.Password != "" {
		serviceBroker.CredentialsType = resources.ServiceBrokerBasicCredentials
	}

	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.PatchServiceBrokerRequest,
		URIParams:   internal.Params{"service_broker_guid": serviceBrokerGUID},
		RequestBody: serviceBroker,
	})

	return jobURL, warnings, err
}
