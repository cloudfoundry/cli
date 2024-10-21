package v3action

import (
	"code.cloudfoundry.org/cli/v7/actor/actionerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccerror"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ManifestParser

type ManifestParser interface {
	AppNames() []string
	RawAppManifest(name string) ([]byte, error)
}

// ApplyApplicationManifest reads in the manifest from the path and provides it
// to the cloud controller.
func (actor Actor) ApplyApplicationManifest(parser ManifestParser, spaceGUID string) (Warnings, error) {
	var allWarnings Warnings

	for _, appName := range parser.AppNames() {
		rawManifest, err := parser.RawAppManifest(appName)
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
			if newErr, ok := err.(ccerror.V3JobFailedError); ok {
				return allWarnings, actionerror.ApplicationManifestError{Message: newErr.Detail}
			}
			return allWarnings, err
		}
	}

	return allWarnings, nil
}
