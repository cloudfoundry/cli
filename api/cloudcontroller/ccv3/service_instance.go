package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
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

func (client Client) GetServiceInstanceParameters(serviceInstanceGUID string) (types.OptionalObject, Warnings, error) {
	var receiver map[string]interface{}

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetServiceInstanceParametersRequest,
		URIParams:    internal.Params{"service_instance_guid": serviceInstanceGUID},
		ResponseBody: &receiver,
	})

	if err != nil {
		return types.OptionalObject{}, warnings, err
	}

	return types.NewOptionalObject(receiver), warnings, nil
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

func (client *Client) DeleteServiceInstance(serviceInstanceGUID string, query ...Query) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{
		RequestName: internal.DeleteServiceInstanceRequest,
		URIParams:   internal.Params{"service_instance_guid": serviceInstanceGUID},
		Query:       query,
	})
}

// ShareServiceInstanceToSpaces will create a sharing relationship between
// the service instance and the shared-to space for each space provided.
func (client *Client) ShareServiceInstanceToSpaces(serviceInstanceGUID string, spaceGUIDs []string) (resources.RelationshipList, Warnings, error) {
	var responseBody resources.RelationshipList

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostServiceInstanceRelationshipsSharedSpacesRequest,
		URIParams:    internal.Params{"service_instance_guid": serviceInstanceGUID},
		RequestBody:  resources.RelationshipList{GUIDs: spaceGUIDs},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UnshareServiceInstanceFromSpace will delete the sharing relationship
// between the service instance and the shared-to space provided.
func (client *Client) UnshareServiceInstanceFromSpace(serviceInstanceGUID string, spaceGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteServiceInstanceRelationshipsSharedSpaceRequest,
		URIParams:   internal.Params{"service_instance_guid": serviceInstanceGUID, "space_guid": spaceGUID},
	})

	return warnings, err
}

// GetServiceInstanceSharedSpaces will fetch relationships between
// a service instance and the shared-to spaces for that service.
func (client *Client) GetServiceInstanceSharedSpaces(serviceInstanceGUID string) ([]SpaceWithOrganization, Warnings, error) {
	var responseBody resources.SharedToSpacesListWrapper
	query := []Query{
		{
			Key:    FieldsSpace,
			Values: []string{"guid", "name", "relationships.organization"},
		},
		{
			Key:    FieldsSpaceOrganization,
			Values: []string{"guid", "name"},
		},
	}
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetServiceInstanceRelationshipsSharedSpacesRequest,
		URIParams:    internal.Params{"service_instance_guid": serviceInstanceGUID},
		Query:        query,
		ResponseBody: &responseBody,
	})
	return mapRelationshipsToSpaces(responseBody), warnings, err
}

func (client *Client) GetServiceInstanceUsageSummary(serviceInstanceGUID string) ([]resources.ServiceInstanceUsageSummary, Warnings, error) {
	var result resources.ServiceInstanceUsageSummaryList

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetServiceInstanceSharedSpacesUsageSummaryRequest,
		URIParams:    internal.Params{"service_instance_guid": serviceInstanceGUID},
		ResponseBody: &result,
	})
	return result.UsageSummary, warnings, err
}

type SpaceWithOrganization struct {
	SpaceGUID        string
	SpaceName        string
	OrganizationName string
}

func mapRelationshipsToSpaces(sharedToSpaces resources.SharedToSpacesListWrapper) []SpaceWithOrganization {
	var spacesToReturn []SpaceWithOrganization

	guidToOrgNameMap := make(map[string]string)

	for _, o := range sharedToSpaces.Organizations {
		guidToOrgNameMap[o.GUID] = o.Name
	}

	for _, s := range sharedToSpaces.Spaces {
		org := s.Relationships[constant.RelationshipTypeOrganization]
		space := SpaceWithOrganization{SpaceGUID: s.GUID, SpaceName: s.Name, OrganizationName: guidToOrgNameMap[org.GUID]}
		spacesToReturn = append(spacesToReturn, space)
	}

	return spacesToReturn
}
