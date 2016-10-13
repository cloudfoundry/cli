package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

const USER_PROVIDED_SERVICE = "user_provided_service_instance"
const MANAGED_SERVICE = "managed_service_instance"

type ServiceInstance struct {
	GUID string
	Name string
	Type string
}

func (serviceInstance *ServiceInstance) UnmarshalJSON(data []byte) error {
	var ccServiceInstance struct {
		Metadata internal.Metadata
		Entity   struct {
			Name string
			Type string
		}
	}
	err := json.Unmarshal(data, &ccServiceInstance)
	if err != nil {
		return err
	}

	serviceInstance.GUID = ccServiceInstance.Metadata.GUID
	serviceInstance.Name = ccServiceInstance.Entity.Name
	serviceInstance.Type = ccServiceInstance.Entity.Type
	return nil
}

func (serviceInstance ServiceInstance) UserProvided() bool {
	return serviceInstance.Type == USER_PROVIDED_SERVICE
}

func (serviceInstance ServiceInstance) Managed() bool {
	return serviceInstance.Type == MANAGED_SERVICE
}

func (client *CloudControllerClient) GetServiceInstances(queries []Query) ([]ServiceInstance, Warnings, error) {
	request := cloudcontroller.Request{
		RequestName: internal.ServiceInstancesRequest,
		Query:       FormatQueryParameters(queries),
	}

	allServiceInstancesList := []ServiceInstance{}
	allWarningsList := Warnings{}

	for {
		var serviceInstances []ServiceInstance
		wrapper := PaginatedWrapper{
			Resources: &serviceInstances,
		}
		response := cloudcontroller.Response{
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
		request = cloudcontroller.Request{
			URI:    wrapper.NextURL,
			Method: "GET",
		}
	}

	return allServiceInstancesList, allWarningsList, nil
}

func (client *CloudControllerClient) GetSpaceServiceInstances(spaceGUID string, includeUserProvidedServices bool, queries []Query) ([]ServiceInstance, Warnings, error) {
	query := FormatQueryParameters(queries)

	if includeUserProvidedServices {
		query.Add("return_user_provided_service_instances", "true")
	}

	request := cloudcontroller.Request{
		RequestName: internal.SpaceServiceInstancesRequest,
		Params: map[string]string{
			"guid": spaceGUID,
		},
		Query: query,
	}

	allServiceInstancesList := []ServiceInstance{}
	allWarningsList := Warnings{}

	for {
		var serviceInstances []ServiceInstance
		wrapper := PaginatedWrapper{
			Resources: &serviceInstances,
		}
		response := cloudcontroller.Response{
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
		request = cloudcontroller.Request{
			URI:    wrapper.NextURL,
			Method: "GET",
		}
	}

	return allServiceInstancesList, allWarningsList, nil
}
