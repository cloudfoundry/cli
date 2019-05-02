package ccv2

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServiceBroker represents a Cloud Controller Service Broker.
type ServiceBroker struct {
	// AuthUsername is the HTTP Basic Auth username of the service broker.
	AuthUsername string
	// BrokerURL is the URL of the service broker.
	BrokerURL string
	// GUID is the unique Service Broker identifier.
	GUID string
	// Name is the name of the service broker.
	Name string
	// SpaceGUID is the space guid of the serivce broker. Only applies to space scoped service brokers.
	SpaceGUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Service Broker response.
func (serviceBroker *ServiceBroker) UnmarshalJSON(data []byte) error {
	var ccServiceBroker struct {
		Metadata internal.Metadata
		Entity   struct {
			Name         string `json:"name"`
			BrokerURL    string `json:"broker_url"`
			AuthUsername string `json:"auth_username"`
			SpaceGUID    string `json:"space_guid"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccServiceBroker)
	if err != nil {
		return err
	}

	serviceBroker.Name = ccServiceBroker.Entity.Name
	serviceBroker.GUID = ccServiceBroker.Metadata.GUID
	serviceBroker.BrokerURL = ccServiceBroker.Entity.BrokerURL
	serviceBroker.AuthUsername = ccServiceBroker.Entity.AuthUsername
	serviceBroker.SpaceGUID = ccServiceBroker.Entity.SpaceGUID
	return nil
}

type createServiceBrokerRequestBody struct {
	Name         string `json:"name"`
	BrokerURL    string `json:"broker_url"`
	AuthUsername string `json:"auth_username"`
	AuthPassword string `json:"auth_password"`
	SpaceGUID    string `json:"space_guid,omitempty"`
}

// CreateServiceBroker posts a service broker resource with the provided
// attributes to the api and returns the result.
func (client *Client) CreateServiceBroker(brokerName, username, password, url, spaceGUID string) (ServiceBroker, Warnings, error) {
	requestBody := createServiceBrokerRequestBody{
		Name:         brokerName,
		BrokerURL:    url,
		AuthUsername: username,
		AuthPassword: password,
		SpaceGUID:    spaceGUID,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return ServiceBroker{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostServiceBrokerRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return ServiceBroker{}, nil, err
	}

	var serviceBroker ServiceBroker
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &serviceBroker,
	}

	err = client.connection.Make(request, &response)

	return serviceBroker, response.Warnings, err
}

// GetServiceBroker returns the service broker with the given GUID.
func (client *Client) GetServiceBroker(serviceBrokerGUID string) (ServiceBroker, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceBrokerRequest,
		URIParams:   Params{"service_broker_guid": serviceBrokerGUID},
	})
	if err != nil {
		return ServiceBroker{}, nil, err
	}

	var serviceBroker ServiceBroker
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &serviceBroker,
	}

	err = client.connection.Make(request, &response)
	return serviceBroker, response.Warnings, err
}

// GetServiceBrokers returns back a list of Service Brokers given the provided
// filters.
func (client *Client) GetServiceBrokers(filters ...Filter) ([]ServiceBroker, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceBrokersRequest,
		Query:       ConvertFilterParameters(filters),
	})

	if err != nil {
		return nil, nil, err
	}

	var fullBrokersList []ServiceBroker
	warnings, err := client.paginate(request, ServiceBroker{}, func(item interface{}) error {
		if broker, ok := item.(ServiceBroker); ok {
			fullBrokersList = append(fullBrokersList, broker)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceBroker{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullBrokersList, warnings, err
}
