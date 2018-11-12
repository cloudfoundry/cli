package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

// EnableServiceForAllOrgs enables access for the given service in all orgs.
func (actor Actor) EnableServiceForAllOrgs(serviceName string) (Warnings, error) {
	var allWarnings Warnings

	services, warnings, err := actor.CloudControllerClient.GetServices(ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	})

	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	if len(services) == 0 {
		return allWarnings, actionerror.ServiceNotFoundError{Name: serviceName}
	}

	servicePlans, warnings, err := actor.CloudControllerClient.GetServicePlans(ccv2.Filter{
		Type:     constant.ServiceGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{services[0].GUID},
	})

	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	for _, plan := range servicePlans {
		warnings, err = actor.CloudControllerClient.UpdateServicePlan(plan.GUID, true)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return allWarnings, err
		}
	}

	return allWarnings, nil
}

// EnablePlanForAllOrgs enables access to a specific plan of the given service in all orgs.
func (actor Actor) EnablePlanForAllOrgs(serviceName, servicePlan string) (Warnings, error) {
	var allWarnings Warnings

	services, warnings, err := actor.CloudControllerClient.GetServices(ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	if len(services) == 0 {
		return allWarnings, actionerror.ServiceNotFoundError{Name: serviceName}
	}

	servicePlans, warnings, err := actor.CloudControllerClient.GetServicePlans(ccv2.Filter{
		Type:     constant.ServiceGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{services[0].GUID},
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	for _, plan := range servicePlans {
		if plan.Name == servicePlan {
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

	return allWarnings, actionerror.ServicePlanNotFoundError{Name: servicePlan, ServiceName: serviceName}
}

// EnableServiceForOrg enables access for the given service in a specific org.
func (actor Actor) EnableServiceForOrg(serviceName, organization string) (Warnings, error) {
	var allWarnings Warnings

	services, warnings, err := actor.CloudControllerClient.GetServices(ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	})
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return allWarnings, err
	}

	if len(services) == 0 {
		return nil, actionerror.ServiceNotFoundError{Name: serviceName}
	}

	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations(ccv2.Filter{
		Type:     constant.NameFilter,
		Operator: constant.EqualOperator,
		Values:   []string{organization},
	})

	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	if len(orgs) == 0 {
		return nil, actionerror.OrganizationNotFoundError{Name: organization}
	}

	servicePlans, warnings, err := actor.CloudControllerClient.GetServicePlans(ccv2.Filter{
		Type:     constant.ServiceGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{services[0].GUID},
	})

	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	for _, plan := range servicePlans {
		_, warnings, err := actor.CloudControllerClient.CreateServicePlanVisibility(plan.GUID, orgs[0].GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return allWarnings, err
		}
	}

	return allWarnings, nil
}

// EnablePlanForOrg enables access to a specific plan of the given service in a specific org.
func (actor Actor) EnablePlanForOrg(serviceName, servicePlanName, orgName string) (Warnings, error) {
	var allWarnings Warnings
	services, warnings, err := actor.CloudControllerClient.GetServices(ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	})

	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}
	if len(services) == 0 {
		return nil, actionerror.ServiceNotFoundError{Name: serviceName}
	}

	orgs, _, _ := actor.CloudControllerClient.GetOrganizations(ccv2.Filter{
		Type:     constant.NameFilter,
		Operator: constant.EqualOperator,
		Values:   []string{orgName},
	})

	if len(orgs) == 0 {
		return nil, actionerror.OrganizationNotFoundError{Name: orgName}
	}

	servicePlans, warnings, err := actor.CloudControllerClient.GetServicePlans(ccv2.Filter{
		Type:     constant.ServiceGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{services[0].GUID},
	})

	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	for _, plan := range servicePlans {
		if plan.Name == servicePlanName {
			_, warnings, err := actor.CloudControllerClient.CreateServicePlanVisibility(plan.GUID, orgs[0].GUID)
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
		actor.CloudControllerClient.DeleteServicePlanVisibility(visibility.GUID)
	}

	return allWarnings, nil
}
