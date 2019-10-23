package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Role struct {
	// GUID is the unique identifier for the role.
	GUID string `json:"guid"`
	// Type is the type of the role.
	Type constant.RoleType `json:"type"`
	// UserGUID is the unique identifier of the user who has this role.
	UserGUID string
	// SpaceGUID is the unique identifier of the space where this role applies,
	// if it is a space role.
	SpaceGUID string
	// OrgGUID is the unique identifier of the org where this role applies,
	// if it is an org role.
	OrgGUID string
}

// MarshalJSON converts a Role into a Cloud Controller Application.
func (r Role) MarshalJSON() ([]byte, error) {
	var ccRole struct {
		GUID          string        `json:"guid,omitempty"`
		Type          string        `json:"type"`
		Relationships Relationships `json:"relationships"`
	}

	ccRole.GUID = r.GUID
	ccRole.Type = string(r.Type)
	ccRole.Relationships = Relationships{
		constant.RelationshipTypeUser: Relationship{GUID: r.UserGUID},
	}

	if r.OrgGUID != "" {
		ccRole.Relationships[constant.RelationshipTypeOrganization] = Relationship{GUID: r.OrgGUID}
	} else if r.SpaceGUID != "" {
		ccRole.Relationships[constant.RelationshipTypeSpace] = Relationship{GUID: r.SpaceGUID}
	}

	return json.Marshal(ccRole)
}

// UnmarshalJSON helps unmarshal a Cloud Controller Role response.
func (r *Role) UnmarshalJSON(data []byte) error {
	var ccRole struct {
		GUID          string `json:"guid"`
		Type          string `json:"type"`
		Relationships Relationships
	}

	err := cloudcontroller.DecodeJSON(data, &ccRole)
	if err != nil {
		return err
	}

	r.GUID = ccRole.GUID
	r.Type = constant.RoleType(ccRole.Type)
	if userRelationship, ok := ccRole.Relationships[constant.RelationshipTypeUser]; ok {
		r.UserGUID = userRelationship.GUID
	}
	if spaceRelationship, ok := ccRole.Relationships[constant.RelationshipTypeSpace]; ok {
		r.SpaceGUID = spaceRelationship.GUID
	}
	if orgRelationship, ok := ccRole.Relationships[constant.RelationshipTypeOrganization]; ok {
		r.OrgGUID = orgRelationship.GUID
	}

	return nil
}

func (client *Client) CreateRole(roleSpec Role) (Role, Warnings, error) {
	bodyBytes, err := json.Marshal(roleSpec)
	if err != nil {
		return Role{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostRoleRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Role{}, nil, err
	}

	var responseRole Role
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseRole,
	}
	err = client.connection.Make(request, &response)

	return responseRole, response.Warnings, err
}
