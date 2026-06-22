package v7action

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/types"
)

func (actor *Actor) GetApplicationLabels(appName string, spaceGUID string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	return actor.extractLabels((*resources.Metadata)(resource.Metadata), warnings, err)
}

func (actor *Actor) GetDomainLabels(domainName string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetDomainByName(domainName)
	return actor.extractLabels(resource.Metadata, warnings, err)
}

func (actor *Actor) GetOrganizationLabels(orgName string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetOrganizationByName(orgName)
	return actor.extractLabels(resource.Metadata, warnings, err)
}

func (actor *Actor) GetRouteLabels(routeName string, spaceGUID string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetRoute(routeName, spaceGUID)
	return actor.extractLabels((*resources.Metadata)(resource.Metadata), warnings, err)
}

func (actor *Actor) GetServiceBrokerLabels(serviceBrokerName string) (map[string]types.NullString, Warnings, error) {
	serviceBroker, warnings, err := actor.GetServiceBrokerByName(serviceBrokerName)
	return actor.extractLabels(serviceBroker.Metadata, warnings, err)
}

func (actor *Actor) GetServiceInstanceLabels(serviceInstanceName, spaceGUID string) (map[string]types.NullString, Warnings, error) {
	serviceInstance, warnings, err := actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
	return actor.extractLabels(serviceInstance.Metadata, warnings, err)
}

func (actor *Actor) GetServiceOfferingLabels(serviceOfferingName, serviceBrokerName string) (map[string]types.NullString, Warnings, error) {
	serviceOffering, warnings, err := actor.CloudControllerClient.GetServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName)
	return actor.extractLabels(serviceOffering.Metadata, Warnings(warnings), actionerror.EnrichAPIErrors(err))
}

func (actor *Actor) GetServicePlanLabels(servicePlanName, serviceOfferingName, serviceBrokerName string) (map[string]types.NullString, Warnings, error) {
	servicePlan, warnings, err := actor.GetServicePlanByNameOfferingAndBroker(servicePlanName, serviceOfferingName, serviceBrokerName)
	return actor.extractLabels(servicePlan.Metadata, warnings, err)
}

func (actor *Actor) GetSpaceLabels(spaceName string, orgGUID string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	return actor.extractLabels(resource.Metadata, warnings, err)
}

func (actor *Actor) GetStackLabels(stackName string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetStackByName(stackName)
	return actor.extractLabels(resource.Metadata, warnings, err)
}

func (actor *Actor) GetBuildpackLabels(buildpackName string, buildpackStack string, buildpackLifecycle string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetBuildpackByNameAndStackAndLifecycle(buildpackName, buildpackStack, buildpackLifecycle)
	return actor.extractLabels(resource.Metadata, warnings, err)
}

func (actor *Actor) extractLabels(metadata *resources.Metadata, warnings Warnings, err error) (map[string]types.NullString, Warnings, error) {
	var labels map[string]types.NullString

	if err != nil {
		return labels, warnings, err
	}
	if metadata != nil {
		labels = metadata.Labels
	}
	return labels, warnings, nil
}

func (actor *Actor) UpdateApplicationLabelsByApplicationName(appName string, spaceGUID string, labels map[string]types.NullString) (Warnings, error) {
	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("app", app.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateBuildpackLabelsByBuildpackNameAndStackAndLifecycle(buildpackName string, stack string, lifecycle string, labels map[string]types.NullString) (Warnings, error) {
	buildpack, warnings, err := actor.GetBuildpackByNameAndStackAndLifecycle(buildpackName, stack, lifecycle)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("buildpack", buildpack.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateDomainLabelsByDomainName(domainName string, labels map[string]types.NullString) (Warnings, error) {
	domain, warnings, err := actor.GetDomainByName(domainName)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("domain", domain.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateOrganizationLabelsByOrganizationName(orgName string, labels map[string]types.NullString) (Warnings, error) {
	org, warnings, err := actor.GetOrganizationByName(orgName)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("org", org.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateRouteLabels(routeName string, spaceGUID string, labels map[string]types.NullString) (Warnings, error) {
	route, warnings, err := actor.GetRoute(routeName, spaceGUID)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("route", route.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateSpaceLabelsBySpaceName(spaceName string, orgGUID string, labels map[string]types.NullString) (Warnings, error) {
	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("space", space.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateStackLabelsByStackName(stackName string, labels map[string]types.NullString) (Warnings, error) {
	stack, warnings, err := actor.GetStackByName(stackName)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("stack", stack.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateServiceBrokerLabelsByServiceBrokerName(serviceBrokerName string, labels map[string]types.NullString) (Warnings, error) {
	serviceBroker, warnings, err := actor.GetServiceBrokerByName(serviceBrokerName)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("service-broker", serviceBroker.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateServiceInstanceLabels(serviceInstanceName, spaceGUID string, labels map[string]types.NullString) (Warnings, error) {
	serviceInstance, warnings, err := actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("service-instance", serviceInstance.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateServiceOfferingLabels(serviceOfferingName string, serviceBrokerName string, labels map[string]types.NullString) (Warnings, error) {
	serviceOffering, warnings, err := actor.CloudControllerClient.GetServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName)
	if err != nil {
		return Warnings(warnings), actionerror.EnrichAPIErrors(err)
	}
	return actor.updateResourceMetadata("service-offering", serviceOffering.GUID, resources.Metadata{Labels: labels}, Warnings(warnings))
}

func (actor *Actor) UpdateServicePlanLabels(servicePlanName string, serviceOfferingName string, serviceBrokerName string, labels map[string]types.NullString) (Warnings, error) {
	servicePlan, warnings, err := actor.GetServicePlanByNameOfferingAndBroker(servicePlanName, serviceOfferingName, serviceBrokerName)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("service-plan", servicePlan.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) updateResourceMetadata(resourceType string, resourceGUID string, payload resources.Metadata, warnings Warnings) (Warnings, error) {
	jobURL, updateWarnings, err := actor.CloudControllerClient.UpdateResourceMetadata(resourceType, resourceGUID, payload)
	warnings = append(warnings, updateWarnings...)
	if err != nil {
		return warnings, err
	}

	if jobURL != "" {
		pollWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
		warnings = append(warnings, pollWarnings...)
		if err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}

// resolveRoutePolicyGUID finds the GUID of the route policy to operate on.
// If source is non-empty, it finds the policy with that source.
// If source is empty, it requires exactly one policy (returns ambiguity or not-found error).
func (actor *Actor) resolveRoutePolicyGUID(routeURL, spaceGUID, source string) (string, *resources.Metadata, Warnings, error) {
	route, routeWarnings, err := actor.GetRoute(routeURL, spaceGUID)
	if err != nil {
		return "", nil, routeWarnings, err
	}

	routePolicies, _, policyWarnings, err := actor.CloudControllerClient.GetRoutePolicies(
		ccv3.Query{Key: ccv3.RouteGUIDFilter, Values: []string{route.GUID}},
	)
	allWarnings := append(routeWarnings, Warnings(policyWarnings)...)
	if err != nil {
		return "", nil, allWarnings, err
	}

	if source != "" {
		for _, p := range routePolicies {
			if p.Source == source {
				return p.GUID, p.Metadata, allWarnings, nil
			}
		}
		return "", nil, allWarnings, actionerror.RoutePolicyNotFoundError{Source: source}
	}

	if len(routePolicies) == 0 {
		return "", nil, allWarnings, actionerror.RoutePolicyNotFoundError{Source: ""}
	}
	if len(routePolicies) > 1 {
		return "", nil, allWarnings, actionerror.RoutePolicyAmbiguityError{RouteURL: routeURL, Count: len(routePolicies)}
	}
	p := routePolicies[0]
	return p.GUID, p.Metadata, allWarnings, nil
}

func (actor *Actor) GetRoutePolicyLabels(routeURL, spaceGUID, source string) (map[string]types.NullString, Warnings, error) {
	_, metadata, warnings, err := actor.resolveRoutePolicyGUID(routeURL, spaceGUID, source)
	if err != nil {
		return nil, warnings, err
	}
	return actor.extractLabels(metadata, warnings, nil)
}

func (actor *Actor) UpdateRoutePolicyLabels(routeURL, spaceGUID, source string, labels map[string]types.NullString) (Warnings, error) {
	policyGUID, _, warnings, err := actor.resolveRoutePolicyGUID(routeURL, spaceGUID, source)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("route-policy", policyGUID, resources.Metadata{Labels: labels}, warnings)
}
