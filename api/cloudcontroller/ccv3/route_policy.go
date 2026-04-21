package ccv3

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v9/resources"
)

// CreateRoutePolicy creates a route policy for a route
func (client *Client) CreateRoutePolicy(routePolicy resources.RoutePolicy) (resources.RoutePolicy, Warnings, error) {
	var responseBody resources.RoutePolicy

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostRoutePolicyRequest,
		RequestBody:  routePolicy,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetRoutePolicies lists route policies
func (client *Client) GetRoutePolicies(query ...Query) ([]resources.RoutePolicy, IncludedResources, Warnings, error) {
	var routePolicies []resources.RoutePolicy

	includedResources, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetRoutePoliciesRequest,
		Query:        query,
		ResponseBody: resources.RoutePolicy{},
		AppendToList: func(item interface{}) error {
			routePolicies = append(routePolicies, item.(resources.RoutePolicy))
			return nil
		},
	})

	return routePolicies, includedResources, warnings, err
}

// GetRoutePolicy gets a single route policy by GUID
func (client *Client) GetRoutePolicy(guid string) (resources.RoutePolicy, Warnings, error) {
	var responseBody resources.RoutePolicy

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetRoutePolicyRequest,
		URIParams:    internal.Params{"route_policy_guid": guid},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// DeleteRoutePolicy deletes a route policy
func (client *Client) DeleteRoutePolicy(guid string) (JobURL, Warnings, error) {
	jobURLString, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteRoutePolicyRequest,
		URIParams:   internal.Params{"route_policy_guid": guid},
	})

	return JobURL(jobURLString), warnings, err
}
