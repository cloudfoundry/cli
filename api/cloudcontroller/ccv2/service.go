package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

type Service struct {
	GUID             string
	Label            string
	Description      string
	DocumentationURL string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Service response.
func (service *Service) UnmarshalJSON(data []byte) error {
	var ccService struct {
		Metadata internal.Metadata
		Entity   struct {
			Label            string `json:"label"`
			Description      string `json:"description"`
			DocumentationURL string `json:"documentation_url"`
		}
	}
	err := json.Unmarshal(data, &ccService)
	if err != nil {
		return err
	}

	service.GUID = ccService.Metadata.GUID
	service.Label = ccService.Entity.Label
	service.Description = ccService.Entity.Description
	service.DocumentationURL = ccService.Entity.DocumentationURL
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
		Result: &service,
	}

	err = client.connection.Make(request, &response)
	return service, response.Warnings, err
}
