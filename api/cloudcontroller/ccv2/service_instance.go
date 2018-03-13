package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServiceInstance represents a Cloud Controller Service Instance.
type ServiceInstance struct {
	// GUID is the unique service instance identifier.
	GUID string

	// Name is the name given to the service instance.
	Name string

	// SpaceGUID is the unique identifier of the space that this service instance
	// belongs to.
	SpaceGUID string

	// ServiceGUID is the unique identifier of the service that this service
	// instance belongs to.
	ServiceGUID string

	// ServicePlanGUID is the unique identifier of the service plan that this
	// service instance belongs to.
	ServicePlanGUID string

	// Type is the type of service instance.
	Type constant.ServiceInstanceType

	// Tags is a list of all tags for the service instance.
	Tags []string

	// DashboardURL is the service-broker provided URL to access administrative
	// features of the service instance.
	DashboardURL string

	// LastOperation is the status of the last operation requested on the service
	// instance.
	LastOperation LastOperation
}

type LastOperation struct {
	// Type is the type of operation that was last performed or currently being
	// performed on the service instance.
	Type string

	// State is the status of the last operation or current operation being
	// performed on the service instance.
	State string

	// Description is the service broker-provided description of the operation.
	Description string

	// UpdatedAt is the timestamp that the Cloud Controller last checked the
	// service instance state from the broker.
	UpdatedAt string

	// CreatedAt is the timestamp that the Cloud Controller created the service
	// instance from the broker.
	CreatedAt string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Service Instance response.
func (serviceInstance *ServiceInstance) UnmarshalJSON(data []byte) error {
	var ccServiceInstance struct {
		Metadata internal.Metadata
		Entity   struct {
			Name            string   `json:"name"`
			SpaceGUID       string   `json:"space_guid"`
			ServiceGUID     string   `json:"service_guid"`
			ServicePlanGUID string   `json:"service_plan_guid"`
			Type            string   `json:"type"`
			Tags            []string `json:"tags"`
			DashboardURL    string   `json:"dashboard_url"`
			LastOperation   struct {
				Type        string `json:"type"`
				State       string `json:"state"`
				Description string `json:"description"`
				UpdatedAt   string `json:"updated_at"`
				CreatedAt   string `json:"created_at"`
			} `json:"last_operation"`
		}
	}
	err := cloudcontroller.DecodeJSON(data, &ccServiceInstance)
	if err != nil {
		return err
	}

	serviceInstance.GUID = ccServiceInstance.Metadata.GUID
	serviceInstance.Name = ccServiceInstance.Entity.Name
	serviceInstance.SpaceGUID = ccServiceInstance.Entity.SpaceGUID
	serviceInstance.ServiceGUID = ccServiceInstance.Entity.ServiceGUID
	serviceInstance.ServicePlanGUID = ccServiceInstance.Entity.ServicePlanGUID
	serviceInstance.Type = constant.ServiceInstanceType(ccServiceInstance.Entity.Type)
	serviceInstance.Tags = ccServiceInstance.Entity.Tags
	serviceInstance.DashboardURL = ccServiceInstance.Entity.DashboardURL
	serviceInstance.LastOperation = LastOperation(ccServiceInstance.Entity.LastOperation)
	return nil
}

// UserProvided returns true if the Service Instance is a user provided
// service.
func (serviceInstance ServiceInstance) UserProvided() bool {
	return serviceInstance.Type == constant.ServiceInstanceTypeUserProvidedService
}

// Managed returns true if the Service Instance is a managed service.
func (serviceInstance ServiceInstance) Managed() bool {
	return serviceInstance.Type == constant.ServiceInstanceTypeManagedService
}

// GetServiceInstance returns the service instance with the given GUID. This
// service can be either a managed or user provided.
func (client *Client) GetServiceInstance(serviceInstanceGUID string) (ServiceInstance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceInstanceRequest,
		URIParams:   Params{"service_instance_guid": serviceInstanceGUID},
	})
	if err != nil {
		return ServiceInstance{}, nil, err
	}

	var serviceInstance ServiceInstance
	response := cloudcontroller.Response{
		Result: &serviceInstance,
	}

	err = client.connection.Make(request, &response)
	return serviceInstance, response.Warnings, err
}

// GetServiceInstances returns back a list of *user provided* Service Instances based
// off of the provided queries.
func (client *Client) GetUserProvidedServiceInstances(filters ...Filter) ([]ServiceInstance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetUserProvidedServiceInstancesRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullInstancesList []ServiceInstance
	warnings, err := client.paginate(request, ServiceInstance{}, func(item interface{}) error {
		if instance, ok := item.(ServiceInstance); ok {
			fullInstancesList = append(fullInstancesList, instance)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceInstance{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullInstancesList, warnings, err
}

// GetServiceInstances returns back a list of *managed* Service Instances based
// off of the provided filters.
func (client *Client) GetServiceInstances(filters ...Filter) ([]ServiceInstance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceInstancesRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullInstancesList []ServiceInstance
	warnings, err := client.paginate(request, ServiceInstance{}, func(item interface{}) error {
		if instance, ok := item.(ServiceInstance); ok {
			fullInstancesList = append(fullInstancesList, instance)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceInstance{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullInstancesList, warnings, err
}

// GetSpaceServiceInstances returns back a list of Service Instances based off
// of the space and filters provided. User provided services will be included
// if includeUserProvidedServices is set to true.
func (client *Client) GetSpaceServiceInstances(spaceGUID string, includeUserProvidedServices bool, filters ...Filter) ([]ServiceInstance, Warnings, error) {
	query := ConvertFilterParameters(filters)

	if includeUserProvidedServices {
		query.Add("return_user_provided_service_instances", "true")
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceServiceInstancesRequest,
		URIParams:   map[string]string{"guid": spaceGUID},
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullInstancesList []ServiceInstance
	warnings, err := client.paginate(request, ServiceInstance{}, func(item interface{}) error {
		if instance, ok := item.(ServiceInstance); ok {
			fullInstancesList = append(fullInstancesList, instance)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceInstance{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullInstancesList, warnings, err
}
