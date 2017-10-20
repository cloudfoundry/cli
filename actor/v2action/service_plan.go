package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

type ServicePlan ccv2.ServicePlan

func (actor Actor) GetServicePlan(servicePlanGUID string) (ServicePlan, Warnings, error) {
	servicePlan, warnings, err := actor.CloudControllerClient.GetServicePlan(servicePlanGUID)
	return ServicePlan(servicePlan), Warnings(warnings), err
}
