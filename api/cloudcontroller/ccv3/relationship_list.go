package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// RelationshipList represents a one to many relationship.
type RelationshipList struct {
	GUIDs []string
}

func (r RelationshipList) MarshalJSON() ([]byte, error) {
	var ccRelationship struct {
		Data []map[string]string `json:"data"`
	}

	for _, guid := range r.GUIDs {
		ccRelationship.Data = append(
			ccRelationship.Data,
			map[string]string{
				"guid": guid,
			})
	}

	return json.Marshal(ccRelationship)
}

func (r *RelationshipList) UnmarshalJSON(data []byte) error {
	var ccRelationships struct {
		Data []map[string]string `json:"data"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccRelationships)
	if err != nil {
		return err
	}

	for _, partner := range ccRelationships.Data {
		r.GUIDs = append(r.GUIDs, partner["guid"])
	}
	return nil
}

// EntitleIsolationSegmentToOrganizations will create a link between the
// isolation segment and the list of organizations provided.
func (client *Client) EntitleIsolationSegmentToOrganizations(isolationSegmentGUID string, organizationGUIDs []string) (RelationshipList, Warnings, error) {
	var responseBody RelationshipList

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostIsolationSegmentRelationshipOrganizationsRequest,
		URIParams:    internal.Params{"isolation_segment_guid": isolationSegmentGUID},
		RequestBody:  RelationshipList{GUIDs: organizationGUIDs},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// ShareServiceInstanceToSpaces will create a sharing relationship between
// the service instance and the shared-to space for each space provided.
func (client *Client) ShareServiceInstanceToSpaces(serviceInstanceGUID string, spaceGUIDs []string) (RelationshipList, Warnings, error) {
	var responseBody RelationshipList

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostServiceInstanceRelationshipsSharedSpacesRequest,
		URIParams:    internal.Params{"service_instance_guid": serviceInstanceGUID},
		RequestBody:  RelationshipList{GUIDs: spaceGUIDs},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
