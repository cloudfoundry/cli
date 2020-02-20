package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// ServicePlan represents a Cloud Controller V3 Service Plan.
type ServicePlan struct {
	// GUID is a unique service plan identifier.
	GUID string `json:"guid,required"`
	// Name is the name of the service plan.
	Name string `json:"name"`

	Metadata *Metadata
}

// GetServicePlans lists service plan with optional filters.
func (client *Client) GetServicePlans(query ...Query) ([]ServicePlan, Warnings, error) {
	var resources []ServicePlan

	_, warnings, err := client.makeListRequest(requestParams{
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
