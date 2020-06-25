package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

// GetServiceInstances lists service instances with optional filters.
func (client *Client) GetServiceInstances(query ...Query) ([]resources.ServiceInstance, IncludedResources, Warnings, error) {
	var result []resources.ServiceInstance

	included, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServiceInstancesRequest,
		Query:        query,
		ResponseBody: resources.ServiceInstance{},
		AppendToList: func(item interface{}) error {
			result = append(result, item.(resources.ServiceInstance))
			return nil
		},
	})

	return result, included, warnings, err
}

func (client *Client) GetServiceInstanceByNameAndSpace(name, spaceGUID string, query ...Query) (resources.ServiceInstance, IncludedResources, Warnings, error) {
	query = append(query,
		Query{
			Key:    NameFilter,
			Values: []string{name},
		},
		Query{
			Key:    SpaceGUIDFilter,
			Values: []string{spaceGUID},
		},
	)

	instances, included, warnings, err := client.GetServiceInstances(query...)

	if err != nil {
		return resources.ServiceInstance{}, IncludedResources{}, warnings, err
	}

	if len(instances) == 0 {
		return resources.ServiceInstance{},
			IncludedResources{},
			warnings,
			ccerror.ServiceInstanceNotFoundError{
				Name:      name,
				SpaceGUID: spaceGUID,
			}
	}

	return instances[0], included, warnings, nil
}

func (client *Client) CreateServiceInstance(serviceInstance resources.ServiceInstance) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{
		RequestName: internal.PostServiceInstanceRequest,
		RequestBody: serviceInstance,
	})
}

func (client *Client) UpdateServiceInstance(serviceInstanceGUID string, serviceInstanceUpdates resources.ServiceInstance) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{
		RequestName: internal.PatchServiceInstanceRequest,
		URIParams:   internal.Params{"service_instance_guid": serviceInstanceGUID},
		RequestBody: serviceInstanceUpdates,
	})
}
