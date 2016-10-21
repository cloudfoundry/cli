package ccv2

import (
	"encoding/json"

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
func (client *CloudControllerClient) GetServiceInstances(queries []Query) ([]ServiceInstance, Warnings, error) {
	request := cloudcontroller.NewRequest(
		internal.ServiceInstancesRequest,
		nil,
		nil,
		FormatQueryParameters(queries),
	)

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
		request = cloudcontroller.NewRequestFromURI(
			wrapper.NextURL,
			"GET",
			nil,
		)
	}

	return allServiceInstancesList, allWarningsList, nil
}

// GetSpaceServiceInstances returns back a list of Service Instances based off
// of the space and queries provided. User provided services will be included
// if includeUserProvidedServices is set to true.
func (client *CloudControllerClient) GetSpaceServiceInstances(spaceGUID string, includeUserProvidedServices bool, queries []Query) ([]ServiceInstance, Warnings, error) {
	query := FormatQueryParameters(queries)

	if includeUserProvidedServices {
		query.Add("return_user_provided_service_instances", "true")
	}

	request := cloudcontroller.NewRequest(
		internal.SpaceServiceInstancesRequest,
		map[string]string{
			"guid": spaceGUID,
		},
		nil,
		query,
	)

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
		request = cloudcontroller.NewRequestFromURI(
			wrapper.NextURL,
			"GET",
			nil,
		)
	}

	return allServiceInstancesList, allWarningsList, nil
}
