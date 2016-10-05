package cloudcontrollerv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontrollerv2/internal"
)

type ServiceInstance struct {
	GUID string
	Name string
}

func (serviceInstance *ServiceInstance) UnmarshalJSON(data []byte) error {
	var ccServiceInstance struct {
		Metadata internal.Metadata
		Entity   struct {
			Name string
		}
	}
	err := json.Unmarshal(data, &ccServiceInstance)
	if err != nil {
		return err
	}

	serviceInstance.GUID = ccServiceInstance.Metadata.GUID
	serviceInstance.Name = ccServiceInstance.Entity.Name
	return nil
}

func (client *CloudControllerClient) GetServiceInstances(queries []Query) ([]ServiceInstance, Warnings, error) {
	request := Request{
		RequestName: ServiceInstancesRequest,
		Query:       FormatQueryParameters(queries),
	}

	allServiceInstancesList := []ServiceInstance{}
	allWarningsList := Warnings{}

	for {
		var serviceInstances []ServiceInstance
		wrapper := PaginatedWrapper{
			Resources: &serviceInstances,
		}
		response := Response{
			Result: &wrapper,
		}

		err := client.connection.Make(request, &response)
		allWarningsList = append(allWarningsList, response.Warnings...)
		if err != nil {
			return nil, allWarningsList, err
		}

		allServiceInstancesList = append(allServiceInstancesList, serviceInstances...)

		if wrapper.NextURL == "" {
			break
		}
		request = Request{
			URI:    wrapper.NextURL,
			Method: "GET",
		}
	}

	return allServiceInstancesList, allWarningsList, nil
}
