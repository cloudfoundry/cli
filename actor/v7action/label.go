package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
)

func (actor *Actor) GetApplicationLabels(appName string, spaceGUID string) (map[string]types.NullString, Warnings, error) {
	var labels map[string]types.NullString
	resource, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return labels, warnings, err
	}
	if resource.Metadata != nil {
		labels = resource.Metadata.Labels
	}
	return labels, warnings, nil
}

func (actor *Actor) GetOrganizationLabels(orgName string) (map[string]types.NullString, Warnings, error) {
	var labels map[string]types.NullString
	resource, warnings, err := actor.GetOrganizationByName(orgName)
	if err != nil {
		return labels, warnings, err
	}
	if resource.Metadata != nil {
		labels = resource.Metadata.Labels
	}
	return labels, warnings, nil
}

func (actor *Actor) GetSpaceLabels(spaceName string, orgGUID string) (map[string]types.NullString, Warnings, error) {
	var labels map[string]types.NullString
	resource, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	if err != nil {
		return labels, warnings, err
	}
	if resource.Metadata != nil {
		labels = resource.Metadata.Labels
	}
	return labels, warnings, nil
}

func (actor *Actor) UpdateApplicationLabelsByApplicationName(appName string, spaceGUID string, labels map[string]types.NullString) (Warnings, error) {
	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("app", app.GUID, ccv3.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateBuildpackLabelsByBuildpackNameAndStack(buildpackName string, stack string, labels map[string]types.NullString) (Warnings, error) {
	buildpack, warnings, err := actor.GetBuildpackByNameAndStack(buildpackName, stack)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("buildpack", buildpack.GUID, ccv3.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateOrganizationLabelsByOrganizationName(orgName string, labels map[string]types.NullString) (Warnings, error) {
	org, warnings, err := actor.GetOrganizationByName(orgName)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("org", org.GUID, ccv3.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) UpdateSpaceLabelsBySpaceName(spaceName string, orgGUID string, labels map[string]types.NullString) (Warnings, error) {
	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	if err != nil {
		return warnings, err
	}
	return actor.updateResourceMetadata("space", space.GUID, ccv3.Metadata{Labels: labels}, warnings)
}

func (actor *Actor) updateResourceMetadata(resourceType string, resourceGUID string, payload ccv3.Metadata, warnings Warnings) (Warnings, error) {
	_, updateWarnings, err := actor.CloudControllerClient.UpdateResourceMetadata(resourceType, resourceGUID, payload)
	return append(warnings, updateWarnings...), err
}
