package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// SecurityGroupRule represents a Cloud Controller Security Group Role.
type SecurityGroupRule struct {
	// Description is a short message discribing the rule.
	Description string

	// Destination is the destination CIDR or range of IPs.
	Destination string

	// Ports is the port or port range.
	Ports string

	// Protocol can be tcp, icmp, udp, all.
	Protocol string
}

// SecurityGroup represents a Cloud Controller Security Group.
type SecurityGroup struct {
	// GUID is the unique Security Group identifier.
	GUID string
	// Name is the Security Group's name.
	Name string
	// Rules are the Security Group Rules associated with this Security Group.
	Rules []SecurityGroupRule
	// RunningDefault is true when this Security Group is applied to all running
	// apps in the CF instance.
	RunningDefault bool
	// StagingDefault is true when this Security Group is applied to all staging
	// apps in the CF instance.
	StagingDefault bool
}

// UnmarshalJSON helps unmarshal a Cloud Controller Security Group response
func (securityGroup *SecurityGroup) UnmarshalJSON(data []byte) error {
	var ccSecurityGroup struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			GUID  string `json:"guid"`
			Name  string `json:"name"`
			Rules []struct {
				Description string `json:"description"`
				Destination string `json:"destination"`
				Ports       string `json:"ports"`
				Protocol    string `json:"protocol"`
			} `json:"rules"`
			RunningDefault bool `json:"running_default"`
			StagingDefault bool `json:"staging_default"`
		} `json:"entity"`
	}

	if err := json.Unmarshal(data, &ccSecurityGroup); err != nil {
		return err
	}

	securityGroup.GUID = ccSecurityGroup.Metadata.GUID
	securityGroup.Name = ccSecurityGroup.Entity.Name
	securityGroup.Rules = make([]SecurityGroupRule, len(ccSecurityGroup.Entity.Rules))
	for i, ccRule := range ccSecurityGroup.Entity.Rules {
		securityGroup.Rules[i].Description = ccRule.Description
		securityGroup.Rules[i].Destination = ccRule.Destination
		securityGroup.Rules[i].Ports = ccRule.Ports
		securityGroup.Rules[i].Protocol = ccRule.Protocol
	}
	securityGroup.RunningDefault = ccSecurityGroup.Entity.RunningDefault
	securityGroup.StagingDefault = ccSecurityGroup.Entity.StagingDefault
	return nil
}

// UpdateSecurityGroupSpace associates a security group in the running phase
// for the lifecycle, specified by its GUID, from a space, which is also
// specified by its GUID.
func (client *Client) UpdateSecurityGroupSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
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

// UpdateSecurityGroupStagingSpace associates a security group in the staging
// phase for the lifecycle, specified by its GUID, from a space, which is also
// specified by its GUID.
func (client *Client) UpdateSecurityGroupStagingSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSecurityGroupStagingSpaceRequest,
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

// GetSecurityGroups returns a list of Security Groups based off the provided
// filters.
func (client *Client) GetSecurityGroups(filters ...Filter) ([]SecurityGroup, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSecurityGroupsRequest,
		Query:       ConvertFilterParameters(filters),
	})

	if err != nil {
		return nil, nil, err
	}

	var securityGroupsList []SecurityGroup
	warnings, err := client.paginate(request, SecurityGroup{}, func(item interface{}) error {
		if securityGroup, ok := item.(SecurityGroup); ok {
			securityGroupsList = append(securityGroupsList, securityGroup)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   SecurityGroup{},
				Unexpected: item,
			}
		}
		return nil
	})

	return securityGroupsList, warnings, err
}

// GetSpaceSecurityGroups returns the running Security Groups associated with
// the provided Space GUID.
func (client *Client) GetSpaceSecurityGroups(spaceGUID string, filters ...Filter) ([]SecurityGroup, Warnings, error) {
	return client.getSpaceSecurityGroupsBySpaceAndLifecycle(spaceGUID, internal.GetSpaceSecurityGroupsRequest, filters)
}

// GetSpaceStagingSecurityGroups returns the staging Security Groups
// associated with the provided Space GUID.
func (client *Client) GetSpaceStagingSecurityGroups(spaceGUID string, filters ...Filter) ([]SecurityGroup, Warnings, error) {
	return client.getSpaceSecurityGroupsBySpaceAndLifecycle(spaceGUID, internal.GetSpaceStagingSecurityGroupsRequest, filters)
}

func (client *Client) getSpaceSecurityGroupsBySpaceAndLifecycle(spaceGUID string, lifecycle string, filters []Filter) ([]SecurityGroup, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: lifecycle,
		URIParams:   map[string]string{"space_guid": spaceGUID},
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var securityGroupsList []SecurityGroup
	warnings, err := client.paginate(request, SecurityGroup{}, func(item interface{}) error {
		if securityGroup, ok := item.(SecurityGroup); ok {
			securityGroupsList = append(securityGroupsList, securityGroup)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   SecurityGroup{},
				Unexpected: item,
			}
		}
		return err
	})

	return securityGroupsList, warnings, err
}

// DeleteSecurityGroupSpace disassociates a security group in the running phase
// for the lifecycle, specified by its GUID, from a space, which is also
// specified by its GUID.
func (client *Client) DeleteSecurityGroupSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSecurityGroupSpaceRequest,
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

// DeleteSecurityGroupStagingSpace disassociates a security group in the
// staging phase fo the lifecycle, specified by its GUID, from a space, which
// is also specified by its GUID.
func (client *Client) DeleteSecurityGroupStagingSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSecurityGroupStagingSpaceRequest,
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
