package ccv3

import (
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
