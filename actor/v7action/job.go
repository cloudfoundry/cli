package v7action

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
)

type JobState constant.JobState

const (
	JobPolling    = JobState(constant.JobPolling)
	JobComplete   = JobState(constant.JobComplete)
	JobFailed     = JobState(constant.JobFailed)
	JobProcessing = JobState(constant.JobProcessing)
)

type PollJobEvent struct {
	State    JobState
	Err      error
	Warnings Warnings
}

func (actor Actor) PollUploadBuildpackJob(jobURL ccv3.JobURL) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.PollJob(jobURL)
	return Warnings(warnings), err
}

func (actor Actor) PollJobToEventStream(jobURL ccv3.JobURL) chan PollJobEvent {
	input := actor.CloudControllerClient.PollJobToEventStream(jobURL)
	if input == nil {
		return nil
	}

	output := make(chan PollJobEvent)

	go func() {
		for event := range input {
			output <- PollJobEvent{
				State:    JobState(event.State),
				Err:      event.Err,
				Warnings: Warnings(event.Warnings),
			}
		}
		close(output)
	}()

	return output
}
