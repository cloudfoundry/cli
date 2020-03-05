package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) CreateSecurityGroup(securityGroup resources.SecurityGroup) (resources.SecurityGroup, Warnings, error) {
	var responseBody resources.SecurityGroup

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostSecurityGroupRequest,
		RequestBody:  securityGroup,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetSecurityGroups(queries ...Query) ([]resources.SecurityGroup, Warnings, error) {
	var securityGroups []resources.SecurityGroup

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetSecurityGroupsRequest,
		Query:        queries,
		ResponseBody: resources.SecurityGroup{},
		AppendToList: func(item interface{}) error {
			securityGroups = append(securityGroups, item.(resources.SecurityGroup))
			return nil
		},
	})

	return securityGroups, warnings, err
}
