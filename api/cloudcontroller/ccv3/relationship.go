package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Relationship represents a one to one relationship.
// An empty GUID will be marshaled as `null`.
type Relationship struct {
	GUID string
}

func (r Relationship) MarshalJSON() ([]byte, error) {
	if r.GUID == "" {
		var emptyCCRelationship struct {
			Data interface{} `json:"data"`
		}
		return json.Marshal(emptyCCRelationship)
	}

	var ccRelationship struct {
		Data struct {
			GUID string `json:"guid"`
		} `json:"data"`
	}

	ccRelationship.Data.GUID = r.GUID
	return json.Marshal(ccRelationship)
}

func (r *Relationship) UnmarshalJSON(data []byte) error {
	var ccRelationship struct {
		Data struct {
			GUID string `json:"guid"`
		} `json:"data"`
	}

	err := json.Unmarshal(data, &ccRelationship)
	if err != nil {
		return err
	}

	r.GUID = ccRelationship.Data.GUID
	return nil
}

// AssignSpaceToIsolationSegment assigns an isolation segment to a space and
// returns the relationship.
func (client *Client) AssignSpaceToIsolationSegment(spaceGUID string, isolationSegmentGUID string) (Relationship, Warnings, error) {
	body, err := json.Marshal(Relationship{GUID: isolationSegmentGUID})
	if err != nil {
		return Relationship{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchSpaceRelationshipIsolationSegmentRequest,
		URIParams:   internal.Params{"space_guid": spaceGUID},
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Relationship{}, nil, err
	}

	var relationship Relationship
	response := cloudcontroller.Response{
		Result: &relationship,
	}

	err = client.connection.Make(request, &response)
	return relationship, response.Warnings, err
}

// GetSpaceIsolationSegment returns the relationship between a space and it's
// isolation segment.
func (client *Client) GetSpaceIsolationSegment(spaceGUID string) (Relationship, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceRelationshipIsolationSegmentRequest,
		URIParams:   internal.Params{"space_guid": spaceGUID},
	})
	if err != nil {
		return Relationship{}, nil, err
	}

	var relationship Relationship
	response := cloudcontroller.Response{
		Result: &relationship,
	}

	err = client.connection.Make(request, &response)
	return relationship, response.Warnings, err
}

// RevokeIsolationSegmentFromOrganization will delete the relationship between
// the isolation segment and the organization provided.
func (client *Client) RevokeIsolationSegmentFromOrganization(isolationSegmentGUID string, orgGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteIsolationSegmentRelationshipOrganizationRequest,
		URIParams:   internal.Params{"isolation_segment_guid": isolationSegmentGUID, "organization_guid": orgGUID},
	})
	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// GetOrganizationDefaultIsolationSegment returns the relationship between an
// organization and it's default isolation segment.
func (client *Client) GetOrganizationDefaultIsolationSegment(orgGUID string) (Relationship, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationDefaultIsolationSegmentRequest,
		URIParams:   internal.Params{"organization_guid": orgGUID},
	})
	if err != nil {
		return Relationship{}, nil, err
	}

	var relationship Relationship
	response := cloudcontroller.Response{
		Result: &relationship,
	}

	err = client.connection.Make(request, &response)
	return relationship, response.Warnings, err
}

// PatchOrganizationDefaultIsolationSegment sets the default isolation segment
// for an organization on the controller.
// If isoSegGuid is empty it will reset the default isolation segment.
func (client *Client) PatchOrganizationDefaultIsolationSegment(orgGUID string, isoSegGUID string) (Relationship, Warnings, error) {
	body, err := json.Marshal(Relationship{GUID: isoSegGUID})
	if err != nil {
		return Relationship{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchOrganizationDefaultIsolationSegmentRequest,
		Body:        bytes.NewReader(body),
		URIParams:   internal.Params{"organization_guid": orgGUID},
	})
	if err != nil {
		return Relationship{}, nil, err
	}

	var relationship Relationship
	response := cloudcontroller.Response{
		Result: &relationship,
	}
	err = client.connection.Make(request, &response)
	return relationship, response.Warnings, err
}

// UnshareServiceInstanceFromSpace will delete the sharing relationship between
// the service instance and the space provided.
func (client *Client) UnshareServiceInstanceFromSpace(serviceInstanceGUID string, spaceGUID string) (Warnings, error) {
	request, _ := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteServiceInstanceRelationshipSharedSpacesRequest,
		URIParams:   internal.Params{"service_instance_guid": serviceInstanceGUID, "space_guid": spaceGUID},
	})

	response := cloudcontroller.Response{}

	err := client.connection.Make(request, &response)

	return response.Warnings, err
}
