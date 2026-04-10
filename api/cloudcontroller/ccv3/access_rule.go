package ccv3

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v9/resources"
)

// CreateAccessRule creates an access rule for a route
func (client *Client) CreateAccessRule(accessRule resources.AccessRule) (resources.AccessRule, Warnings, error) {
	var responseBody resources.AccessRule

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostAccessRuleRequest,
		RequestBody:  accessRule,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetAccessRules lists access rules
func (client *Client) GetAccessRules(query ...Query) ([]resources.AccessRule, IncludedResources, Warnings, error) {
	var accessRules []resources.AccessRule

	includedResources, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetAccessRulesRequest,
		Query:        query,
		ResponseBody: resources.AccessRule{},
		AppendToList: func(item interface{}) error {
			accessRules = append(accessRules, item.(resources.AccessRule))
			return nil
		},
	})

	return accessRules, includedResources, warnings, err
}

// GetAccessRule gets a single access rule by GUID
func (client *Client) GetAccessRule(guid string) (resources.AccessRule, Warnings, error) {
	var responseBody resources.AccessRule

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetAccessRuleRequest,
		URIParams:    internal.Params{"access_rule_guid": guid},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// DeleteAccessRule deletes an access rule
func (client *Client) DeleteAccessRule(guid string) (JobURL, Warnings, error) {
	jobURLString, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteAccessRuleRequest,
		URIParams:   internal.Params{"access_rule_guid": guid},
	})

	return JobURL(jobURLString), warnings, err
}
