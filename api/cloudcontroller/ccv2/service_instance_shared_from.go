package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServiceInstanceSharedFrom is the struct representation of a share_from
// object in Cloud Controller.
type ServiceInstanceSharedFrom struct {
	SpaceGUID        string `json:"space_guid"`
	SpaceName        string `json:"space_name"`
	OrganizationName string `json:"organization_name"`
}

// GetServiceInstanceSharedFrom returns back a ServiceInstanceSharedFrom
// object.
func (client *Client) GetServiceInstanceSharedFrom(serviceInstanceGUID string) (ServiceInstanceSharedFrom, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceInstanceSharedFromRequest,
		URIParams:   Params{"service_instance_guid": serviceInstanceGUID},
	})
	if err != nil {
		return ServiceInstanceSharedFrom{}, nil, err
	}

	var serviceInstanceSharedFrom ServiceInstanceSharedFrom
	response := cloudcontroller.Response{
		Result: &serviceInstanceSharedFrom,
	}

	err = client.connection.Make(request, &response)
	return serviceInstanceSharedFrom, response.Warnings, err
}
