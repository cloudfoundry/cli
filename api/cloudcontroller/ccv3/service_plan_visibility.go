package ccv3

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"

type VisibilityType string

type VisibilityDetail struct {
	// Name is the organization name
	Name string `json:"name,omitempty"`
	// GUID of the organization
	GUID string `json:"guid"`
}

// ServicePlanVisibility represents a Cloud Controller V3 Service Plan Visibility.
type ServicePlanVisibility struct {
	// Type is one of 'public', 'organization', 'space' or 'admin'
	Type VisibilityType `json:"type"`

	// Organizations list of organizations for the service plan
	Organizations []VisibilityDetail `json:"organizations,omitempty"`

	// Space that the plan is visible in
	Space *VisibilityDetail `json:"space,omitempty"`
}

func (client *Client) GetServicePlanVisibility(servicePlanGUID string) (ServicePlanVisibility, Warnings, error) {
	var result ServicePlanVisibility

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetServicePlanVisibilityRequest,
		URIParams:    internal.Params{"service_plan_guid": servicePlanGUID},
		ResponseBody: &result,
	})

	return result, warnings, err
}

func (client *Client) UpdateServicePlanVisibility(servicePlanGUID string, planVisibility ServicePlanVisibility) (ServicePlanVisibility, Warnings, error) {
	var result ServicePlanVisibility

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostServicePlanVisibilityRequest,
		URIParams:    internal.Params{"service_plan_guid": servicePlanGUID},
		RequestBody:  planVisibility,
		ResponseBody: &result,
	})

	return result, warnings, err
}

func (client *Client) DeleteServicePlanVisibility(servicePlanGUID, organizationGUID string) (Warnings, error) {

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteServicePlanVisibilityRequest,
		URIParams:   internal.Params{"service_plan_guid": servicePlanGUID, "organization_guid": organizationGUID},
	})

	return warnings, err
}
