package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

type SecurityGroup struct {
	GUID string
	Name string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Security Group response
func (securityGroup *SecurityGroup) UnmarshalJSON(data []byte) error {
	var ccSecurityGroup struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			GUID string `json:"guid"`
			Name string `json:"name"`
		}
	}

	if err := json.Unmarshal(data, &ccSecurityGroup); err != nil {
		return err
	}

	securityGroup.GUID = ccSecurityGroup.Metadata.GUID
	securityGroup.Name = ccSecurityGroup.Entity.Name
	return nil
}

func (client *Client) AssociateSpaceWithSecurityGroup(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSecurityGroupSpaceRequest,
		URIParams: Params{
			"security_group_guid": securityGroupGUID,
			"space_guid":          spaceGUID,
		},
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

func (client *Client) GetSecurityGroups(queries []Query) ([]SecurityGroup, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSecurityGroupsRequest,
		Query:       FormatQueryParameters(queries),
	})

	if err != nil {
		return nil, nil, err
	}

	var securityGroupsList []SecurityGroup
	warnings, err := client.paginate(request, SecurityGroup{}, func(item interface{}) error {
		if securityGroup, ok := item.(SecurityGroup); ok {
			securityGroupsList = append(securityGroupsList, securityGroup)
		} else {
			return cloudcontroller.UnknownObjectInListError{
				Expected:   SecurityGroup{},
				Unexpected: item,
			}
		}
		return nil
	})

	return securityGroupsList, warnings, err
}
