package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type ServiceInstance struct {
	GUID string `json:"guid"`
	Name string `json:"name"`
}

// GetServiceInstances lists ServiceInstances with optional filters.
func (client *Client) GetServiceInstances(query ...Query) ([]ServiceInstance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
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
