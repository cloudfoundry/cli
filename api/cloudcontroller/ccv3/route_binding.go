package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) CreateRouteBinding(binding resources.RouteBinding) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{
		RequestName: internal.PostRouteBindingRequest,
		RequestBody: binding,
	})
}
