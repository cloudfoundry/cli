package v7action

import (
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3"
)

func (actor Actor) PollUploadBuildpackJob(jobURL ccv3.JobURL) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.PollJob(jobURL)
	return Warnings(warnings), err
}
