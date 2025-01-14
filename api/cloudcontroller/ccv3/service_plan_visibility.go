package ccv3

import (
	ccv3internal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) GetServicePlanVisibility(servicePlanGUID string) (resources.ServicePlanVisibility, Warnings, error) {
	var result resources.ServicePlanVisibility

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.GetServicePlanVisibilityRequest,
		URIParams:    internal.Params{"service_plan_guid": servicePlanGUID},
		ResponseBody: &result,
	})

	return result, warnings, err
}

func (client *Client) UpdateServicePlanVisibility(servicePlanGUID string, planVisibility resources.ServicePlanVisibility) (resources.ServicePlanVisibility, Warnings, error) {
	var result resources.ServicePlanVisibility

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.PostServicePlanVisibilityRequest,
		URIParams:    internal.Params{"service_plan_guid": servicePlanGUID},
		RequestBody:  planVisibility,
		ResponseBody: &result,
	})

	return result, warnings, err
}

func (client *Client) DeleteServicePlanVisibility(servicePlanGUID, organizationGUID string) (Warnings, error) {

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: ccv3internal.DeleteServicePlanVisibilityRequest,
		URIParams:   internal.Params{"service_plan_guid": servicePlanGUID, "organization_guid": organizationGUID},
	})

	return warnings, err
}
