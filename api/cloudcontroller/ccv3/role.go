package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) CreateRole(roleSpec resources.Role) (resources.Role, Warnings, error) {
	var responseBody resources.Role

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostRoleRequest,
		RequestBody:  roleSpec,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) DeleteRole(roleGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteRoleRequest,
		URIParams:   internal.Params{"role_guid": roleGUID},
	})

	return jobURL, warnings, err
}

// GetRoles lists roles with optional filters & includes.
func (client *Client) GetRoles(query ...Query) ([]resources.Role, IncludedResources, Warnings, error) {
	var roles []resources.Role

	includedResources, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetRolesRequest,
		Query:        query,
		ResponseBody: resources.Role{},
		AppendToList: func(item interface{}) error {
			roles = append(roles, item.(resources.Role))
			return nil
		},
	})

	return roles, includedResources, warnings, err
}
