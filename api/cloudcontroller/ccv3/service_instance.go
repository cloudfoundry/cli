package ccv3

import (
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
	var resources []ServiceInstance

	_, warnings, err := client.makeListRequest(requestParams{
		RequestName:  internal.GetServiceInstancesRequest,
		Query:        query,
		ResponseBody: ServiceInstance{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(ServiceInstance))
			return nil
		},
	})

	return resources, warnings, err
}
