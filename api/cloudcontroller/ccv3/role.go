package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
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
	// Username is the name of the user who has this role, e.g. "admin", "user@example.com"
	Username string
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
					GUID     string `json:"guid,omitempty"`
					Username string `json:"username,omitempty"`
					Origin   string `json:"origin,omitempty"`
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
	if r.Username != "" {
		ccRole.Relationships.User.Data.Username = r.Username
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
		IncludedUsers IncludedUsers
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

	if includedUsers, ok := ccRole.IncludedUsers[constant.IncludedTypeUsers]; ok {
		r.Username = includedUsers.Username
	}
	return nil
}

func (client *Client) CreateRole(roleSpec Role) (Role, Warnings, error) {
	var responseBody Role

	warnings, err := client.makeCreateRequest(
		internal.PostRoleRequest,
		roleSpec,
		&responseBody,
	)

	return responseBody, warnings, err
}

func (client *Client) DeleteRole(roleGUID string) (JobURL, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URIParams: map[string]string{
			"role_guid": roleGUID,
		},
		RequestName: internal.DeleteRoleRequest,
	})
	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return JobURL(response.ResourceLocationURL), response.Warnings, err
}

// GetRoles lists roles with optional filters & includes.
func (client *Client) GetRoles(query ...Query) ([]Role, IncludedResources, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRolesRequest,
		Query:       query,
	})
	if err != nil {
		return nil, IncludedResources{}, nil, err
	}

	var rolesList []Role
	includes, warnings, err := client.paginateWithIncludes(request, Role{}, func(item interface{}) error {
		if role, ok := item.(Role); ok {
			rolesList = append(rolesList, role)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Role{},
				Unexpected: item,
			}
		}
		return nil
	})

	return rolesList, includes, warnings, err
}
