package ccv3

import (
	"bytes"
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

	err := json.Unmarshal(data, &ccRelationships)
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
	body, err := json.Marshal(RelationshipList{GUIDs: organizationGUIDs})
	if err != nil {
		return RelationshipList{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostIsolationSegmentRelationshipOrganizationsRequest,
		URIParams:   internal.Params{"isolation_segment_guid": isolationSegmentGUID},
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return RelationshipList{}, nil, err
	}

	var relationships RelationshipList
	response := cloudcontroller.Response{
		Result: &relationships,
	}

	err = client.connection.Make(request, &response)
	return relationships, response.Warnings, err
}

// ShareServiceInstanceToSpaces will create a sharing relationship between
// the service instance and the shared-to space for each space provided.
func (client *Client) ShareServiceInstanceToSpaces(serviceInstanceGUID string, spaceGUIDs []string) (RelationshipList, Warnings, error) {
	body, err := json.Marshal(RelationshipList{GUIDs: spaceGUIDs})
	if err != nil {
		return RelationshipList{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostServiceInstanceRelationshipsSharedSpacesRequest,
		URIParams:   internal.Params{"service_instance_guid": serviceInstanceGUID},
		Body:        bytes.NewReader(body),
	})

	if err != nil {
		return RelationshipList{}, nil, err
	}

	var relationships RelationshipList
	response := cloudcontroller.Response{
		Result: &relationships,
	}

	err = client.connection.Make(request, &response)
	return relationships, response.Warnings, err
}
