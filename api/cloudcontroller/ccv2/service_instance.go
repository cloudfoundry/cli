package ccv2

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServiceInstanceType is the type of the Service Instance.
type ServiceInstanceType string

const (
	// UserProvidedService is a Service Instance that is created by a user.
	UserProvidedService ServiceInstanceType = "user_provided_service_instance"

	// ManagedService is a Service Instance that is managed by a service broker.
	ManagedService ServiceInstanceType = "managed_service_instance"
)

// ServiceInstance represents a Cloud Controller Service Instance.
type ServiceInstance struct {
	GUID string
	Name string
	Type ServiceInstanceType
}

// UnmarshalJSON helps unmarshal a Cloud Controller Service Instance response.
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
	serviceInstance.Type = ServiceInstanceType(ccServiceInstance.Entity.Type)
	return nil
}

// UserProvidedService returns true if the Service Instance is a user provided
// service.
func (serviceInstance ServiceInstance) UserProvided() bool {
	return serviceInstance.Type == UserProvidedService
}

// Managed returns true if the Service Instance is a managed service.
func (serviceInstance ServiceInstance) Managed() bool {
	return serviceInstance.Type == ManagedService
}

// GetServiceInstances returns back a list of *managed* Service Instances based
// off of the provided queries.
func (client *Client) GetServiceInstances(queries []Query) ([]ServiceInstance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.ServiceInstancesRequest,
		Query:       FormatQueryParameters(queries),
	})
	if err != nil {
		return nil, nil, err
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
		request, err = client.newHTTPRequest(requestOptions{
			URI:    wrapper.NextURL,
			Method: http.MethodGet,
		})
		if err != nil {
			return nil, allWarningsList, err
		}
	}

	return allServiceInstancesList, allWarningsList, nil
}

// GetSpaceServiceInstances returns back a list of Service Instances based off
// of the space and queries provided. User provided services will be included
// if includeUserProvidedServices is set to true.
func (client *Client) GetSpaceServiceInstances(spaceGUID string, includeUserProvidedServices bool, queries []Query) ([]ServiceInstance, Warnings, error) {
	query := FormatQueryParameters(queries)

	if includeUserProvidedServices {
		query.Add("return_user_provided_service_instances", "true")
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.SpaceServiceInstancesRequest,
		URIParams:   map[string]string{"guid": spaceGUID},
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
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

		request, err = client.newHTTPRequest(requestOptions{
			URI:    wrapper.NextURL,
			Method: http.MethodGet,
		})
		if err != nil {
			return nil, allWarningsList, err
		}
	}

	return allServiceInstancesList, allWarningsList, nil
}
