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
	// UserName is the name of the user who has this role, e.g. "admin", "user@example.com"
	UserName string
	// Origin is the identity server, default "uaa". Active Directory can also be an origin
	Origin string
	// OrgGUID is the unique identifier of the org where this role applies,
	// if it is an org role.
	OrgGUID string
	// SpaceGUID is the unique identifier of the space where this role applies,
	// if it is a space role.
	SpaceGUID string
}

// MarshalJSON converts a Role into a Cloud Controller Application.
func (r Role) MarshalJSON() ([]byte, error) {
	type data struct {
		GUID string `json:"guid"`
	}

	type orgOrSpaceJSON struct {
		Data data `json:"data"`
	}
	var ccRole struct {
		GUID          string `json:"guid,omitempty"`
		Type          string `json:"type"`
		Relationships struct {
			Organization *orgOrSpaceJSON `json:"organization,omitempty"`
			Space        *orgOrSpaceJSON `json:"space,omitempty"`
			User         struct {
				Data struct {
					GUID   string `json:"guid,omitempty"`
					Name   string `json:"name,omitempty"`
					Origin string `json:"origin,omitempty"`
				} `json:"data"`
			} `json:"user"`
		} `json:"relationships"`
	}

	ccRole.GUID = r.GUID
	ccRole.Type = string(r.Type)
	if r.OrgGUID != "" {
		ccRole.Relationships.Organization = &orgOrSpaceJSON{
			Data: data{GUID: r.OrgGUID},
		}
	}
	if r.SpaceGUID != "" {
		ccRole.Relationships.Space = &orgOrSpaceJSON{
			Data: data{GUID: r.SpaceGUID},
		}
	}
	if r.UserName != "" {
		ccRole.Relationships.User.Data.Name = r.UserName
		ccRole.Relationships.User.Data.Origin = r.Origin
	} else {
		ccRole.Relationships.User.Data.GUID = r.UserGUID
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
