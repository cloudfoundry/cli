package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

//go:generate counterfeiter . ManifestParser

type ManifestParser interface {
	AppNames() []string
	RawManifest(name string) ([]byte, error)
}

// ApplyApplicationManifest reads in the manifest from the path and provides it
// to the cloud controller.
func (actor Actor) ApplyApplicationManifest(parser ManifestParser, spaceGUID string) (Warnings, error) {
	var allWarnings Warnings

	for _, appName := range parser.AppNames() {
		rawManifest, err := parser.RawManifest(appName)
		if err != nil {
			return allWarnings, err
		}

		app, getAppWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)

		allWarnings = append(allWarnings, getAppWarnings...)
		if err != nil {
			return allWarnings, err
		}

		jobURL, applyManifestWarnings, err := actor.CloudControllerClient.UpdateApplicationApplyManifest(app.GUID, rawManifest)
		allWarnings = append(allWarnings, applyManifestWarnings...)
		if err != nil {
			return allWarnings, err
		}

		pollWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
		allWarnings = append(allWarnings, pollWarnings...)
		if err != nil {
			if newErr, ok := err.(ccerror.JobFailedError); ok {
				return allWarnings, actionerror.ApplicationManifestError{Message: newErr.Message}
			}
			return allWarnings, err
		}
	}

	return allWarnings, nil
}
