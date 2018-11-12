package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

// EnableServiceForAllOrgs enables access for the given service in all orgs.
func (actor Actor) EnableServiceForAllOrgs(serviceName string) (Warnings, error) {
	servicePlans, allWarnings, err := actor.GetServicePlansForService(serviceName)
	if err != nil {
		return allWarnings, err
	}

	for _, plan := range servicePlans {
		warnings, err := actor.CloudControllerClient.UpdateServicePlan(plan.GUID, true)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return allWarnings, err
		}
	}

	return allWarnings, nil
}

// EnablePlanForAllOrgs enables access to a specific plan of the given service in all orgs.
func (actor Actor) EnablePlanForAllOrgs(serviceName, servicePlanName string) (Warnings, error) {
	servicePlans, allWarnings, err := actor.GetServicePlansForService(serviceName)
	if err != nil {
		return allWarnings, err
	}

	// We delete all service plan visibilities for the given Plan since the attribute public should function as a giant on/off
	// switch for all orgs. Thus we need to clean up any visibilities laying around so that they don't carry over.
	for _, plan := range servicePlans {
		if plan.Name == servicePlanName {
			warnings, err := actor.removeOrgLevelServicePlanVisibilities(plan.GUID)
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				return allWarnings, err
			}

			ccv2Warnings, err := actor.CloudControllerClient.UpdateServicePlan(plan.GUID, true)
			allWarnings = append(allWarnings, ccv2Warnings...)
			return allWarnings, err
		}
	}

	return allWarnings, actionerror.ServicePlanNotFoundError{PlanName: servicePlanName, ServiceName: serviceName}
}

// EnableServiceForOrg enables access for the given service in a specific org.
func (actor Actor) EnableServiceForOrg(serviceName, orgName string) (Warnings, error) {
	servicePlans, allWarnings, err := actor.GetServicePlansForService(serviceName)
	if err != nil {
		return allWarnings, err
	}

	org, orgWarnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, orgWarnings...)
	if err != nil {
		return allWarnings, err
	}

	for _, plan := range servicePlans {
		_, warnings, err := actor.CloudControllerClient.CreateServicePlanVisibility(plan.GUID, org.GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return allWarnings, err
		}
	}

	return allWarnings, nil
}

// EnablePlanForOrg enables access to a specific plan of the given service in a specific org.
func (actor Actor) EnablePlanForOrg(serviceName, servicePlanName, orgName string) (Warnings, error) {
	servicePlans, allWarnings, err := actor.GetServicePlansForService(serviceName)
	if err != nil {
		return allWarnings, err
	}

	org, orgWarnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, orgWarnings...)
	if err != nil {
		return allWarnings, err
	}

	for _, plan := range servicePlans {
		if plan.Name == servicePlanName {
			_, warnings, err := actor.CloudControllerClient.CreateServicePlanVisibility(plan.GUID, org.GUID)
			allWarnings = append(allWarnings, warnings...)
			return allWarnings, err
		}
	}

	return nil, fmt.Errorf("Service plan '%s' not found", servicePlanName)
}

func (actor Actor) removeOrgLevelServicePlanVisibilities(servicePlanGUID string) (Warnings, error) {
	var allWarnings Warnings

	visibilities, warnings, err := actor.CloudControllerClient.GetServicePlanVisibilities(ccv2.Filter{
		Type:     constant.ServicePlanGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{servicePlanGUID},
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	for _, visibility := range visibilities {
		warnings, err := actor.CloudControllerClient.DeleteServicePlanVisibility(visibility.GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return allWarnings, err
		}
	}

	return allWarnings, nil
}
