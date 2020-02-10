package v7action

import (
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
)

func (actor *Actor) GetApplicationLabels(appName string, spaceGUID string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	return actor.extractLabels(resource.Metadata, warnings, err)
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
	return actor.extractLabels(resource.Metadata, warnings, err)
}

func (actor Actor) GetServiceBrokerLabels(serviceBrokerName string) (map[string]types.NullString, Warnings, error) {
	serviceBroker, warnings, err := actor.GetServiceBrokerByName(serviceBrokerName)
	return actor.extractLabels(serviceBroker.Metadata, warnings, err)
}

func (actor Actor) GetServiceOfferingLabels(serviceOfferingName, serviceBrokerName string) (map[string]types.NullString, Warnings, error) {
	serviceOffering, warnings, err := actor.GetServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName)
	return actor.extractLabels(serviceOffering.Metadata, warnings, err)
}

func (actor *Actor) GetSpaceLabels(spaceName string, orgGUID string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	return actor.extractLabels(resource.Metadata, warnings, err)
}

func (actor *Actor) GetStackLabels(stackName string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetStackByName(stackName)
	return actor.extractLabels(resource.Metadata, warnings, err)
}

func (actor *Actor) GetBuildpackLabels(buildpackName string, buildpackStack string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetBuildpackByNameAndStack(buildpackName, buildpackStack)
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

func (actor *Actor) UpdateBuildpackLabelsByBuildpackNameAndStack(buildpackName string, stack string, labels map[string]types.NullString) (Warnings, error) {
	buildpack, warnings, err := actor.GetBuildpackByNameAndStack(buildpackName, stack)
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
	return actor.updateResourceMetadataAsync("service-broker", serviceBroker.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateServiceOfferingLabels(serviceOfferingName string, serviceBrokerName string, labels map[string]types.NullString) (Warnings, error) {
	serviceOffering, warnings, err := actor.GetServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("service-offering", serviceOffering.GUID, resources.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) updateResourceMetadata(resourceType string, resourceGUID string, payload resources.Metadata, warnings Warnings) (Warnings, error) {
	_, updateWarnings, err := actor.CloudControllerClient.UpdateResourceMetadata(resourceType, resourceGUID, payload)
	return append(warnings, updateWarnings...), err
}

func (actor *Actor) updateResourceMetadataAsync(resourceType string, resourceGUID string, payload resources.Metadata, warnings Warnings) (Warnings, error) {
	jobURL, updateWarnings, err := actor.CloudControllerClient.UpdateResourceMetadataAsync(resourceType, resourceGUID, payload)
	warnings = append(warnings, updateWarnings...)
	if err != nil {
		return warnings, err
	}
	pollWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
	warnings = append(warnings, pollWarnings...)
	return warnings, err
}
