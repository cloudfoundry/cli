package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
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
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServicePlansRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullServicePlans []ServicePlan
	warnings, err := client.paginate(request, ServicePlan{}, func(item interface{}) error {
		if servicePlan, ok := item.(ServicePlan); ok {
			fullServicePlans = append(fullServicePlans, servicePlan)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServicePlan{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullServicePlans, warnings, err
}
