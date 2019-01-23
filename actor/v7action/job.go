package v7action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

func (actor Actor) PollJob(jobURL ccv3.JobURL) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.PollJob(jobURL)

	return Warnings(warnings), err
}
