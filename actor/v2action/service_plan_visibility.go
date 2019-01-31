package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

type ServicePlanVisibility ccv2.ServicePlanVisibility

// GetServicePlanVisibilities fetches service plan visibilities for a plan by GUID.
func (actor *Actor) GetServicePlanVisibilities(planGUID string) ([]ServicePlanVisibility, Warnings, error) {
	visibilities, warnings, err := actor.CloudControllerClient.GetServicePlanVisibilities(ccv2.Filter{
		Type:     constant.ServicePlanGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{planGUID},
	})
	if err != nil {
		return nil, Warnings(warnings), err
	}

	var visibilitesToReturn []ServicePlanVisibility
	for _, v := range visibilities {
		visibilitesToReturn = append(visibilitesToReturn, ServicePlanVisibility(v))
	}

	return visibilitesToReturn, Warnings(warnings), nil
}
