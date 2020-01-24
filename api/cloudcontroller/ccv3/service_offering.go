package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
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
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceOfferingsRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullServiceOfferingList []ServiceOffering
	warnings, err := client.paginate(request, ServiceOffering{}, func(item interface{}) error {
		if serviceOffering, ok := item.(ServiceOffering); ok {
			fullServiceOfferingList = append(fullServiceOfferingList, serviceOffering)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceOffering{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullServiceOfferingList, warnings, err
}
