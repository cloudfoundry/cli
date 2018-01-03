package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Space represents a Cloud Controller Space.
type Space struct {
	GUID                     string
	OrganizationGUID         string
	Name                     string
	AllowSSH                 bool
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

// GetSpaces returns a list of Spaces based off of the provided queries.
func (client *Client) GetSpaces(queries ...QQuery) ([]Space, Warnings, error) {
	params := FormatQueryParameters(queries)
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

// GetStagingSpacesBySecurityGroup returns a list of Spaces based on the provided
// SecurityGroup GUID.
func (client *Client) GetStagingSpacesBySecurityGroup(securityGroupGUID string) ([]Space, Warnings, error) {
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

// GetRunningSpacesBySecurityGroup returns a list of Spaces based on the provided
// SecurityGroup GUID.
func (client *Client) GetRunningSpacesBySecurityGroup(securityGroupGUID string) ([]Space, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSecurityGroupRunningSpacesRequest,
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
