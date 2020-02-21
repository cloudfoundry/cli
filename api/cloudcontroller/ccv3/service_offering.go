package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// ServiceOffering represents a Cloud Controller V3 Service Offering.
type ServiceOffering struct {
	// GUID is a unique service offering identifier.
	GUID string
	// Name is the name of the service offering.
	Name string
	// ServiceBrokerName is the name of the service broker
	ServiceBrokerName string

	Metadata *Metadata
}

func (so *ServiceOffering) UnmarshalJSON(data []byte) error {
	var response serviceOfferingResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	so.GUID = response.GUID
	so.Name = response.Name
	so.ServiceBrokerName = response.Relationships.ServiceBroker.Data.Name
	so.Metadata = response.Metadata

	return nil
}

// serviceOfferingResponse represents a Cloud Controller V3 Service Offering (when reading)
type serviceOfferingResponse struct {
	// GUID is a unique service offering identifier.
	GUID string `json:"guid"`
	// Name is the name of the service offering.
	Name string `json:"name"`
	// Relationships is the relationship for the service broker
	Relationships serviceOfferingRelationships `json:"relationships"`

	Metadata *Metadata
}

// serviceOfferingRelationships represents the relationships to other resources
type serviceOfferingRelationships struct {
	ServiceBroker serviceOfferingRelationshipBroker `json:"service_broker"`
}

type serviceOfferingRelationshipBroker struct {
	Data serviceOfferingRelationshipBrokerData `json:"data"`
}

type serviceOfferingRelationshipBrokerData struct {
	Name string `json:"name"`
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
