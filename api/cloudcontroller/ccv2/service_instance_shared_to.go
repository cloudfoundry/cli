package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServiceInstanceSharedTo represents a Cloud Controller relationship object
// that describes a service instance in another space (and possibly org) that
// this service is **shared to**.
type ServiceInstanceSharedTo struct {
	// SpaceGUID is the unique identifier of the space that this service is
	// shared to.
	SpaceGUID string `json:"space_guid"`

	// SpaceName is the name of the space that this service is shared to.
	SpaceName string `json:"space_name"`

	// OrganizationName is the name of the organization that this service is
	// shared to.
	OrganizationName string `json:"organization_name"`

	// BoundAppCount is the number of apps that are bound to the shared to
	// service instance.
	BoundAppCount int `json:"bound_app_count"`
}

// GetServiceInstanceSharedTos returns a list of ServiceInstanceSharedTo objects.
func (client *Client) GetServiceInstanceSharedTos(serviceInstanceGUID string) ([]ServiceInstanceSharedTo, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceInstanceSharedToRequest,
		URIParams:   Params{"service_instance_guid": serviceInstanceGUID},
	})

	if err != nil {
		return nil, nil, err
	}

	var fullSharedToList []ServiceInstanceSharedTo
	warnings, err := client.paginate(request, ServiceInstanceSharedTo{}, func(item interface{}) error {
		if instance, ok := item.(ServiceInstanceSharedTo); ok {
			fullSharedToList = append(fullSharedToList, instance)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceInstanceSharedTo{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullSharedToList, warnings, err
}
