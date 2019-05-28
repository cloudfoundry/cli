package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// ServiceBroker represents a Cloud Controller V3 Service Broker.
type ServiceBroker struct {
	// GUID is a unique service broker identifier.
	GUID string `json:"guid"`
	// Name is the name of the service broker.
	Name string `json:"name"`
	// URL is the url of the service broker.
	URL string `json:"url"`
}

// ServiceBrokerCredentials represents a data structure required to create a
// new Cloud Controller V3 Service Broker.
type ServiceBrokerCredentials struct {
	// Name is the name of the service broker.
	Name string `json:"name"`
	// URL is the url of the service broker.
	URL string `json:"url"`
	// Username is the Basic Auth username for the service broker.
	Username string `json:"username"`
	// Password is the Basic Auth password for the service broker.
	Password string `json:"password"`
	// SpaceGUID references which space this service broker belongs to. Empty if
	// not space-scoped.
	SpaceGUID string `json:"space_guid,omitempty"`
}

// CreateServiceBroker registers a new service broker.
func (client *Client) CreateServiceBroker(credentials ServiceBrokerCredentials) (Warnings, error) {
	bodyBytes, err := json.Marshal(credentials)
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostServiceBrokerRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// GetServiceBrokers lists service brokers.
func (client *Client) GetServiceBrokers() ([]ServiceBroker, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceBrokersRequest,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullList []ServiceBroker
	warnings, err := client.paginate(request, ServiceBroker{}, func(item interface{}) error {
		if serviceBroker, ok := item.(ServiceBroker); ok {
			fullList = append(fullList, serviceBroker)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceBroker{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullList, warnings, err
}
