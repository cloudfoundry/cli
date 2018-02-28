package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Space represents a Cloud Controller Space.
type Space struct {
	// GUID is the unique space identifier.
	GUID string

	// OrganizationGUID is the unique identifier of the organization this space
	// belongs to.
	OrganizationGUID string

	// Name is the name given to the space.
	Name string

	// AllowSSH specifies whether SSH is enabled for this space.
	AllowSSH bool

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
	if err := json.Unmarshal(data, &ccSpace); err != nil {
		return err
	}

	space.GUID = ccSpace.Metadata.GUID
	space.Name = ccSpace.Entity.Name
	space.AllowSSH = ccSpace.Entity.AllowSSH
	space.SpaceQuotaDefinitionGUID = ccSpace.Entity.SpaceQuotaDefinitionGUID
	space.OrganizationGUID = ccSpace.Entity.OrganizationGUID
	return nil
}

//go:generate go run $GOPATH/src/code.cloudfoundry.org/cli/util/codegen/generate.go Space codetemplates/delete_async_by_guid.go.template delete_space.go
//go:generate go run $GOPATH/src/code.cloudfoundry.org/cli/util/codegen/generate.go Space codetemplates/delete_async_by_guid_test.go.template delete_space_test.go

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
