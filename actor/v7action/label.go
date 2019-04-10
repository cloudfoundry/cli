package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
)

func (actor *Actor) UpdateApplicationLabelsByApplicationName(appName string, spaceGUID string, labels map[string]types.NullString) (Warnings, error) {
	app, appWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return appWarnings, err
	}
	app.Metadata.Labels = labels
	_, updateWarnings, err := actor.UpdateApplication(app)
	return append(appWarnings, updateWarnings...), err
}

func (actor *Actor) UpdateOrganizationLabelsByOrganizationName(orgName string, labels map[string]types.NullString) (Warnings, error) {
	org, warnings, err := actor.GetOrganizationByName(orgName)
	if err != nil {
		return warnings, err
	}
	org.Metadata.Labels = labels
	_, updateWarnings, err := actor.CloudControllerClient.UpdateOrganization(ccv3.Organization(org))
	return append(warnings, updateWarnings...), err
}
