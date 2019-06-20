package v7action

import (
	"strconv"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

func (actor Actor) SetSpaceManifest(spaceGUID string, rawManifest []byte, noRoute bool) (Warnings, error) {
	var allWarnings Warnings
	jobURL, applyManifestWarnings, err := actor.CloudControllerClient.UpdateSpaceApplyManifest(
		spaceGUID,
		rawManifest,
		ccv3.Query{
			Key:    ccv3.NoRouteFilter,
			Values: []string{strconv.FormatBool(noRoute)},
		},
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
