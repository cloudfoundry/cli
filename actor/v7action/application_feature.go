package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

type ApplicationFeature ccv3.Buildpack

func (actor Actor) GetAppFeature(appGUID string, featureName string) (ccv3.ApplicationFeature, Warnings, error) {
	appFeature, warnings, err := actor.CloudControllerClient.GetAppFeature(appGUID, featureName)

	return appFeature, Warnings(warnings), err
}

func (actor Actor) GetSSHEnabled(appGUID string) (ccv3.SSHEnabled, Warnings, error) {
	sshEnabled, warnings, err := actor.CloudControllerClient.GetSSHEnabled(appGUID)
	return sshEnabled, Warnings(warnings), err
}

func (actor Actor) GetSSHEnabledByAppName(appName string, spaceGUID string) (ccv3.SSHEnabled, Warnings, error) {
	var allWarnings Warnings

	app, warnings, err := actor.CloudControllerClient.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ccv3.SSHEnabled{}, allWarnings, err
	}

	sshEnabled, warnings, err := actor.CloudControllerClient.GetSSHEnabled(app.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ccv3.SSHEnabled{}, allWarnings, err
	}

	return sshEnabled, allWarnings, nil
}

func (actor Actor) UpdateAppFeature(app resources.Application, enabled bool, featureName string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.UpdateAppFeature(app.GUID, enabled, featureName)
	return Warnings(warnings), err
}
