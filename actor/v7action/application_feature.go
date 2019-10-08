package v7action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

type ApplicationFeature ccv3.Buildpack

func (actor Actor) GetAppFeature(appGUID string, featureName string) (ccv3.ApplicationFeature, Warnings, error) {
	appFeature, warnings, err := actor.CloudControllerClient.GetAppFeature(appGUID, "ssh")

	return appFeature, Warnings(warnings), err
}

func (actor Actor) GetSSHEnabled(appGUID string) (ccv3.SSHEnabled, Warnings, error) {
	sshEnabled, warnings, err := actor.CloudControllerClient.GetSSHEnabled(appGUID)
	return sshEnabled, Warnings(warnings), err
}

func (actor Actor) UpdateAppFeature(app Application, enabled bool, featureName string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.UpdateAppFeature(app.GUID, enabled, "ssh")
	return Warnings(warnings), err
}
