package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/cloudcontroller/jsonry"
)

// ServicePlan represents a Cloud Controller V3 Service Plan.
type ServicePlan struct {
	// GUID is a unique service plan identifier.
	GUID string
	// Name is the name of the service plan.
	Name string
	// VisibilityType can be "public", "admin", "organization" or "space"
	VisibilityType VisibilityType `json:"visibility_type"`
	// ServicePlanGUID is the GUID of the service offering
	ServiceOfferingGUID string `jsonry:"relationships.service_offering.data.guid"`

	Metadata *Metadata
}

func (sp *ServicePlan) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, sp)
}

// GetServicePlans lists service plan with optional filters.
func (client *Client) GetServicePlans(query ...Query) ([]ServicePlan, Warnings, error) {
	var resources []ServicePlan

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServicePlansRequest,
		Query:        query,
		ResponseBody: ServicePlan{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(ServicePlan))
			return nil
		},
	})

	return resources, warnings, err
}
