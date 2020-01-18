package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// ServiceInstance represents a Cloud Controller V3 Service Instance.
type ServiceInstance struct {
	// GUID is a unique service instance identifier.
	GUID string `json:"guid"`
	// Name is the name of the service instance.
	Name string `json:"name"`
}

// GetServiceInstances lists service instances with optional filters.
func (client *Client) GetServiceInstances(query ...Query) ([]ServiceInstance, Warnings, error) {
	request, err := client.NewHTTPRequest(requestOptions{
		RequestName: internal.GetServiceInstancesRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullServiceInstanceList []ServiceInstance
	warnings, err := client.paginate(request, ServiceInstance{}, func(item interface{}) error {
		if serviceInstance, ok := item.(ServiceInstance); ok {
			fullServiceInstanceList = append(fullServiceInstanceList, serviceInstance)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceInstance{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullServiceInstanceList, warnings, err
}
