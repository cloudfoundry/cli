package ccv2

import (
	"bytes"
	"encoding/json"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Space represents a Cloud Controller Space.
type Space struct {
	// AllowSSH specifies whether SSH is enabled for this space.
	AllowSSH bool

	// GUID is the unique space identifier.
	GUID string

	// Name is the name given to the space.
	Name string

	// OrganizationGUID is the unique identifier of the organization this space
	// belongs to.
	OrganizationGUID string

	// SpaceQuotaDefinitionGUID is the unique identifier of the space quota
	// defined for this space.
	SpaceQuotaDefinitionGUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Space response.
func (space *Space) UnmarshalJSON(data []byte) error {
	var ccSpace struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name                     string `json:"name"`
			AllowSSH                 bool   `json:"allow_ssh"`
			SpaceQuotaDefinitionGUID string `json:"space_quota_definition_guid"`
			OrganizationGUID         string `json:"organization_guid"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccSpace)
	if err != nil {
		return err
	}

	space.GUID = ccSpace.Metadata.GUID
	space.Name = ccSpace.Entity.Name
	space.AllowSSH = ccSpace.Entity.AllowSSH
	space.SpaceQuotaDefinitionGUID = ccSpace.Entity.SpaceQuotaDefinitionGUID
	space.OrganizationGUID = ccSpace.Entity.OrganizationGUID
	return nil
}

type createSpaceRequestBody struct {
	Name             string `json:"name"`
	OrganizationGUID string `json:"organization_guid"`
}

// CreateSpace creates a new space with the provided spaceName in the org with
// the provided orgGUID.
func (client *Client) CreateSpace(spaceName string, orgGUID string) (Space, Warnings, error) {
	requestBody := createSpaceRequestBody{
		Name:             spaceName,
		OrganizationGUID: orgGUID,
	}

	bodyBytes, _ := json.Marshal(requestBody)

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostSpaceRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return Space{}, nil, err
	}

	var space Space
	response := cloudcontroller.Response{
		Result: &space,
	}

	err = client.connection.Make(request, &response)

	return space, response.Warnings, err
}

// DeleteSpace deletes the Space associated with the provided
// GUID. It will return the Cloud Controller job that is assigned to the
// Space deletion.
func (client *Client) DeleteSpace(guid string) (Job, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSpaceRequest,
		URIParams:   Params{"space_guid": guid},
		Query: url.Values{
			"recursive": {"true"},
			"async":     {"true"},
		},
	})
	if err != nil {
		return Job{}, nil, err
	}

	var job Job
	response := cloudcontroller.Response{
		Result: &job,
	}

	err = client.connection.Make(request, &response)
	return job, response.Warnings, err
}

// GetSecurityGroupSpaces returns a list of Spaces based on the provided
// SecurityGroup GUID.
func (client *Client) GetSecurityGroupSpaces(securityGroupGUID string) ([]Space, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSecurityGroupSpacesRequest,
		URIParams:   map[string]string{"security_group_guid": securityGroupGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullSpacesList []Space
	warnings, err := client.paginate(request, Space{}, func(item interface{}) error {
		if space, ok := item.(Space); ok {
			fullSpacesList = append(fullSpacesList, space)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Space{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullSpacesList, warnings, err
}

// GetSecurityGroupStagingSpaces returns a list of Spaces based on the provided
// SecurityGroup GUID.
func (client *Client) GetSecurityGroupStagingSpaces(securityGroupGUID string) ([]Space, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSecurityGroupStagingSpacesRequest,
		URIParams:   map[string]string{"security_group_guid": securityGroupGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullSpacesList []Space
	warnings, err := client.paginate(request, Space{}, func(item interface{}) error {
		if space, ok := item.(Space); ok {
			fullSpacesList = append(fullSpacesList, space)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Space{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullSpacesList, warnings, err
}

// GetSpaces returns a list of Spaces based off of the provided filters.
func (client *Client) GetSpaces(filters ...Filter) ([]Space, Warnings, error) {
	params := ConvertFilterParameters(filters)
	params.Add("order-by", "name")
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpacesRequest,
		Query:       params,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullSpacesList []Space
	warnings, err := client.paginate(request, Space{}, func(item interface{}) error {
		if space, ok := item.(Space); ok {
			fullSpacesList = append(fullSpacesList, space)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Space{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullSpacesList, warnings, err
}

type updateRoleRequestBody struct {
	Username string `json:"username"`
}

// UpdateSpaceDeveloperByUsername grants the given username the space developer role.
func (client *Client) UpdateSpaceDeveloperByUsername(spaceGUID string, username string) (Warnings, error) {
	requestBody := updateRoleRequestBody{
		Username: username,
	}

	bodyBytes, _ := json.Marshal(requestBody)

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSpaceDeveloperByUsernameRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Warnings{}, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return Warnings(response.Warnings), err
}

// UpdateSpaceManagerByUsername grants the given username the space manager role.
func (client *Client) UpdateSpaceManagerByUsername(spaceGUID string, username string) (Warnings, error) {
	requestBody := updateRoleRequestBody{
		Username: username,
	}

	bodyBytes, _ := json.Marshal(requestBody)

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSpaceManagerByUsernameRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
