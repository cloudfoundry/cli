package v7action

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/resources"
)

func (actor Actor) DiffSpaceManifest(spaceGUID string, rawManifest []byte) (resources.ManifestDiff, Warnings, error) {
	diff, warnings, err := actor.CloudControllerClient.GetSpaceManifestDiff(
		spaceGUID,
		rawManifest,
	)

	return diff, Warnings(warnings), err
}

func (actor Actor) SetSpaceManifest(spaceGUID string, rawManifest []byte) (Warnings, error) {
	var allWarnings Warnings
	jobURL, applyManifestWarnings, err := actor.CloudControllerClient.UpdateSpaceApplyManifest(
		spaceGUID,
		rawManifest,
	)
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
	return allWarnings, nil
}
