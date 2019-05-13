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

func (actor *Actor) UpdateApplicationLabelsByApplicationName(appName string, spaceGUID string, labels map[string]types.NullString) (Warnings, error) {
	app, appWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return appWarnings, err
	}
	app.Metadata = &Metadata{Labels: labels}
	_, updateWarnings, err := actor.UpdateApplication(app)
	return append(appWarnings, updateWarnings...), err
}

func (actor *Actor) UpdateOrganizationLabelsByOrganizationName(orgName string, labels map[string]types.NullString) (Warnings, error) {
	org, warnings, err := actor.GetOrganizationByName(orgName)
	if err != nil {
		return warnings, err
	}
	org.Metadata = &Metadata{Labels: labels}
	_, updateWarnings, err := actor.UpdateOrganization(org)
	return append(warnings, updateWarnings...), err
}

func (actor *Actor) UpdateSpaceLabelsBySpaceName(spaceName string, orgGUID string, labels map[string]types.NullString) (Warnings, error) {
	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	if err != nil {
		return warnings, err
	}
	space.Metadata = &ccv3.Metadata{Labels: labels}
	_, updateWarnings, err := actor.CloudControllerClient.UpdateSpace(ccv3.Space(space))
	return append(warnings, updateWarnings...), err
}
