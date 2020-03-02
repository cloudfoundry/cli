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
