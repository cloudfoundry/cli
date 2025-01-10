package ccv3

import (
	ccv3internal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) CreateSecurityGroup(securityGroup resources.SecurityGroup) (resources.SecurityGroup, Warnings, error) {
	var responseBody resources.SecurityGroup

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.PostSecurityGroupRequest,
		RequestBody:  securityGroup,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetSecurityGroups(queries ...Query) ([]resources.SecurityGroup, Warnings, error) {
	var securityGroups []resources.SecurityGroup

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  ccv3internal.GetSecurityGroupsRequest,
		Query:        queries,
		ResponseBody: resources.SecurityGroup{},
		AppendToList: func(item interface{}) error {
			securityGroups = append(securityGroups, item.(resources.SecurityGroup))
			return nil
		},
	})

	return securityGroups, warnings, err
}

func (client *Client) GetRunningSecurityGroups(spaceGUID string, queries ...Query) ([]resources.SecurityGroup, Warnings, error) {
	var securityGroups []resources.SecurityGroup

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  ccv3internal.GetSpaceRunningSecurityGroupsRequest,
		URIParams:    internal.Params{"space_guid": spaceGUID},
		Query:        queries,
		ResponseBody: resources.SecurityGroup{},
		AppendToList: func(item interface{}) error {
			securityGroups = append(securityGroups, item.(resources.SecurityGroup))
			return nil
		},
	})

	return securityGroups, warnings, err
}

func (client *Client) GetStagingSecurityGroups(spaceGUID string, queries ...Query) ([]resources.SecurityGroup, Warnings, error) {
	var securityGroups []resources.SecurityGroup

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  ccv3internal.GetSpaceStagingSecurityGroupsRequest,
		URIParams:    internal.Params{"space_guid": spaceGUID},
		Query:        queries,
		ResponseBody: resources.SecurityGroup{},
		AppendToList: func(item interface{}) error {
			securityGroups = append(securityGroups, item.(resources.SecurityGroup))
			return nil
		},
	})

	return securityGroups, warnings, err
}

func (client *Client) UnbindSecurityGroupRunningSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: ccv3internal.DeleteSecurityGroupRunningSpaceRequest,
		URIParams: internal.Params{
			"security_group_guid": securityGroupGUID,
			"space_guid":          spaceGUID,
		},
	})

	return warnings, err
}

func (client *Client) UnbindSecurityGroupStagingSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: ccv3internal.DeleteSecurityGroupStagingSpaceRequest,
		URIParams: internal.Params{
			"security_group_guid": securityGroupGUID,
			"space_guid":          spaceGUID,
		},
	})

	return warnings, err
}

func (client *Client) UpdateSecurityGroupRunningSpace(securityGroupGUID string, spaceGUIDs []string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: ccv3internal.PostSecurityGroupRunningSpaceRequest,
		URIParams:   internal.Params{"security_group_guid": securityGroupGUID},
		RequestBody: resources.RelationshipList{
			GUIDs: spaceGUIDs,
		},
	})

	return warnings, err
}

func (client *Client) UpdateSecurityGroupStagingSpace(securityGroupGUID string, spaceGUIDs []string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: ccv3internal.PostSecurityGroupStagingSpaceRequest,
		URIParams:   internal.Params{"security_group_guid": securityGroupGUID},
		RequestBody: resources.RelationshipList{
			GUIDs: spaceGUIDs,
		},
	})

	return warnings, err
}

func (client *Client) UpdateSecurityGroup(securityGroup resources.SecurityGroup) (resources.SecurityGroup, Warnings, error) {
	var responseBody resources.SecurityGroup

	securityGroupGUID := securityGroup.GUID
	securityGroup.GUID = ""
	securityGroup.Name = ""

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.PatchSecurityGroupRequest,
		URIParams:    internal.Params{"security_group_guid": securityGroupGUID},
		RequestBody:  securityGroup,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) DeleteSecurityGroup(securityGroupGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: ccv3internal.DeleteSecurityGroupRequest,
		URIParams:   internal.Params{"security_group_guid": securityGroupGUID},
	})

	return jobURL, warnings, err
}
