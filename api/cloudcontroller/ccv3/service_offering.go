package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/cloudcontroller/jsonry"
)

// ServiceOffering represents a Cloud Controller V3 Service Offering.
type ServiceOffering struct {
	// GUID is a unique service offering identifier.
	GUID string
	// Name is the name of the service offering.
	Name string
	// ServiceBrokerName is the name of the service broker
	ServiceBrokerName string `jsonry:"relationships.service_broker.data.name"`

	Metadata *Metadata
}

func (so *ServiceOffering) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, so)
}

// GetServiceOffering lists service offering with optional filters.
func (client *Client) GetServiceOfferings(query ...Query) ([]ServiceOffering, Warnings, error) {
	var resources []ServiceOffering

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServiceOfferingsRequest,
		Query:        query,
		ResponseBody: ServiceOffering{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(ServiceOffering))
			return nil
		},
	})

	return resources, warnings, err
}
