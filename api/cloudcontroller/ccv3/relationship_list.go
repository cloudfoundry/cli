package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

// EntitleIsolationSegmentToOrganizations will create a link between the
// isolation segment and the list of organizations provided.
func (client *Client) EntitleIsolationSegmentToOrganizations(isolationSegmentGUID string, organizationGUIDs []string) (resources.RelationshipList, Warnings, error) {
	var responseBody resources.RelationshipList

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostIsolationSegmentRelationshipOrganizationsRequest,
		URIParams:    internal.Params{"isolation_segment_guid": isolationSegmentGUID},
		RequestBody:  resources.RelationshipList{GUIDs: organizationGUIDs},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
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

// GetServiceInstanceSharedSpaces will fetch relationships between
// a service instance and the shared-to spaces for that service.
func (client *Client) GetServiceInstanceSharedSpaces(serviceInstanceGUID string) ([]resources.Space, Warnings, error) {
	var relationships resources.RelationshipList

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetServiceInstanceRelationshipsSharedSpacesRequest,
		URIParams:    internal.Params{"service_instance_guid": serviceInstanceGUID},
		ResponseBody: &relationships,
	})

	return mapRelationshipsToSpaces(relationships), warnings, err
}

func mapRelationshipsToSpaces(relationships resources.RelationshipList) []resources.Space {
	spaces := make([]resources.Space, len(relationships.GUIDs))
	for i, g := range relationships.GUIDs {
		spaces[i] = resources.Space{GUID: g}
	}
	return spaces
}
