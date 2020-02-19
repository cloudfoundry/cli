package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) CreateSecurityGroup(securityGroup resources.SecurityGroup) (resources.SecurityGroup, Warnings, error) {
	bodyBytes, err := json.Marshal(securityGroup)
	if err != nil {
		return resources.SecurityGroup{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostSecurityGroupRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return resources.SecurityGroup{}, nil, err
	}

	var responseSecurityGroup resources.SecurityGroup
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseSecurityGroup,
	}
	err = client.connection.Make(request, &response)

	return responseSecurityGroup, response.Warnings, err
}
