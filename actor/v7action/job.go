package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

func (actor Actor) PollUploadBuildpackJob(jobURL ccv3.JobURL) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.PollJob(jobURL)

	if err != nil {
		switch v := err.(type) {
		case ccerror.V3JobFailedError:
			switch v.Code {
			case 290000:
				return Warnings(warnings), ccerror.BuildpackAlreadyExistsForStackError{Message: v.Detail}
			case 290003:
				return Warnings(warnings), ccerror.BuildpackAlreadyExistsWithoutStackError{Message: v.Detail}
			case 390011:
				return Warnings(warnings), ccerror.BuildpackStacksDontMatchError{Message: v.Detail}
			case 390012:
				return Warnings(warnings), ccerror.BuildpackStackDoesNotExistError{Message: v.Detail}
			case 390013:
				return Warnings(warnings), ccerror.BuildpackZipInvalidError{Message: v.Detail}
			}
		}
	}

	return Warnings(warnings), err
}
