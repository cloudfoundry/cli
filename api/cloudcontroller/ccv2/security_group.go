package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// SecurityGroupLifecycle represents the lifecycle phase of a security group
// binding.
type SecurityGroupLifecycle string

const (
	// SecurityGroupLifecycleRunning indicates the lifecycle phase running.
	SecurityGroupLifecycleRunning SecurityGroupLifecycle = "running"

	// SecurityGroupLifecycleStaging indicates the lifecycle phase staging.
	SecurityGroupLifecycleStaging SecurityGroupLifecycle = "staging"
)

type SecurityGroupRule struct {
	Description string
	Destination string
	Ports       string
	Protocol    string
}

type SecurityGroup struct {
	GUID           string
	Name           string
	Rules          []SecurityGroupRule
	RunningDefault bool
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
