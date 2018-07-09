package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServiceBroker represents a Cloud Controller Service Broker.
type ServiceBroker struct {
	// GUID is the unique Service Broker identifier.
	GUID string
	// Name is the name of the service broker.
	Name string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Service Broker response.
func (serviceBroker *ServiceBroker) UnmarshalJSON(data []byte) error {
	var ccServiceBroker struct {
		Metadata internal.Metadata
		Entity   struct {
			Name string `json:"name"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccServiceBroker)
	if err != nil {
		return err
	}

	serviceBroker.Name = ccServiceBroker.Entity.Name
	serviceBroker.GUID = ccServiceBroker.Metadata.GUID
	return nil
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
