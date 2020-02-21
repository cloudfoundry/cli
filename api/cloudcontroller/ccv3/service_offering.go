package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// ServiceOffering represents a Cloud Controller V3 Service Offering.
type ServiceOffering struct {
	// GUID is a unique service offering identifier.
	GUID string `json:"guid"`
	// Name is the name of the service offering.
	Name string `json:"name"`

	Metadata *Metadata
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
