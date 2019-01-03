package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Service represents a Cloud Controller Service.
type Service struct {
	// GUID is the unique Service identifier.
	GUID string
	// Label is the name of the service.
	Label string
	// Description is a short blurb describing the service.
	Description string
	// DocumentationURL is a url that points to a documentation page for the
	// service.
	DocumentationURL string
	// ServiceBrokerName is the name of the broker providing this service.
	ServiceBrokerName string
	// Extra is a field with extra data pertaining to the service.
	Extra ServiceExtra
}

// UnmarshalJSON helps unmarshal a Cloud Controller Service response.
func (service *Service) UnmarshalJSON(data []byte) error {
	var ccService struct {
		Metadata internal.Metadata
		Entity   struct {
			Label             string `json:"label"`
			Description       string `json:"description"`
			DocumentationURL  string `json:"documentation_url"`
			ServiceBrokerName string `json:"service_broker_name"`
			Extra             string `json:"extra"`
		}
	}

	err := cloudcontroller.DecodeJSON(data, &ccService)
	if err != nil {
		return err
	}

	service.GUID = ccService.Metadata.GUID
	service.Label = ccService.Entity.Label
	service.Description = ccService.Entity.Description
	service.DocumentationURL = ccService.Entity.DocumentationURL
	service.ServiceBrokerName = ccService.Entity.ServiceBrokerName

	// We explicitly unmarshal the Extra field to type string because CC returns
	// a stringified JSON object ONLY for the 'extra' key (see test stub JSON
	// response). This unmarshal strips escaped quotes, at which time we can then
	// unmarshal into the ServiceExtra object.
	// If 'extra' is null or not provided, this means sharing is NOT enabled
	if len(ccService.Entity.Extra) != 0 {
		extra := ServiceExtra{}
		err = json.Unmarshal([]byte(ccService.Entity.Extra), &extra)
		if err != nil {
			return err
		}
		service.Extra.Shareable = extra.Shareable
		if service.DocumentationURL == "" {
			service.DocumentationURL = extra.DocumentationURL
		}
	}

	return nil
}

// GetService returns the service with the given GUID.
func (client *Client) GetService(serviceGUID string) (Service, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceRequest,
		URIParams:   Params{"service_guid": serviceGUID},
	})
	if err != nil {
		return Service{}, nil, err
	}

	var service Service
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &service,
	}

	err = client.connection.Make(request, &response)
	return service, response.Warnings, err
}

// GetServices returns a list of Services given the provided filters.
func (client *Client) GetServices(filters ...Filter) ([]Service, Warnings, error) {
	opts := requestOptions{
		RequestName: internal.GetServicesRequest,
		Query:       ConvertFilterParameters(filters),
	}

	return client.makeServicesRequest(opts)
}

func (client *Client) GetSpaceServices(spaceGUID string, filters ...Filter) ([]Service, Warnings, error) {
	opts := requestOptions{
		RequestName: internal.GetSpaceServicesRequest,
		Query:       ConvertFilterParameters(filters),
		URIParams:   Params{"space_guid": spaceGUID},
	}

	return client.makeServicesRequest(opts)
}

func (client *Client) makeServicesRequest(opts requestOptions) ([]Service, Warnings, error) {
	request, err := client.newHTTPRequest(opts)

	if err != nil {
		return nil, nil, err
	}

	var fullServicesList []Service
	warnings, err := client.paginate(request, Service{}, func(item interface{}) error {
		if service, ok := item.(Service); ok {
			fullServicesList = append(fullServicesList, service)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Service{},
				Unexpected: item,
			}
		}
		return nil
	})
	return fullServicesList, warnings, err
}
