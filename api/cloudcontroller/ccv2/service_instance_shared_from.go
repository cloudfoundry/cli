package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServiceInstanceSharedFrom represents a Cloud Controller relationship object
// that describes a service instance in another space (and possibly org) that
// this service instance is **shared from**.
type ServiceInstanceSharedFrom struct {
	// SpaceGUID is the unique identifier of the space that this service is
	// shared from.
	SpaceGUID string `json:"space_guid"`

	// SpaceName is the name of the space that this service is shared from.
	SpaceName string `json:"space_name"`

	// OrganizationName is the name of the organization that this service is
	// shared from.
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
