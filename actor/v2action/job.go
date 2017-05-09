package v2action

import (
	"io"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type Job ccv2.Job

func (actor Actor) PollJob(job Job) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.PollJob(ccv2.Job(job))
	return Warnings(warnings), err
}

func (actor Actor) UploadApplicationPackage(appGUID string, existingResources []Resource, newResources io.Reader, newResourcesLength int64) (Job, Warnings, error) {
	job, warnings, err := actor.CloudControllerClient.UploadApplicationPackage(appGUID, actor.actorToCCResources(existingResources), newResources, newResourcesLength)
	return Job(job), Warnings(warnings), err
}
