package v2action

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// JobFailedError represents a failed Cloud Controller Job. It wraps the error
// returned back from the Cloud Controller.
type JobFailedError struct {
	JobGUID string
	Message string
}

func (e JobFailedError) Error() string {
	return fmt.Sprintf("Job (%s) failed: %s", e.JobGUID, e.Message)
}

// JobTimeoutError is returned from PollJob when the OverallPollingTimeout has
// been reached.
type JobTimeoutError struct {
	JobGUID string
	Timeout time.Duration
}

func (e JobTimeoutError) Error() string {
	return fmt.Sprintf("Job (%s) polling has reached the maximum timeout of %s seconds", e.JobGUID, e.Timeout)
}

// PollJob will keep polling the given job until the job has terminated, an
// error is encountered, or config.OverallPollingTimeout is reached. In the
// last case, a JobTimeoutError is returned.
func (actor Actor) PollJob(job ccv2.Job) (Warnings, error) {
	originalJobGUID := job.GUID

	var allWarnings ccv2.Warnings
	var err error

	finished := make(chan bool, 1)
	stopPolling := make(chan bool, 1)
	go func() {
	Polling:
		for {
			select {
			case <-stopPolling:
				break Polling
			default:
				var warnings ccv2.Warnings
				job, warnings, err = actor.CloudControllerClient.GetJob(job.GUID)
				allWarnings = append(allWarnings, warnings...)

				if err != nil || job.Terminated() {
					finished <- true
					break Polling
				}
				time.Sleep(actor.Config.PollingInterval())
			}
		}
	}()

	select {
	case <-time.After(actor.Config.OverallPollingTimeout()):
		stopPolling <- true
		return Warnings(allWarnings), JobTimeoutError{
			JobGUID: originalJobGUID,
			Timeout: actor.Config.OverallPollingTimeout(),
		}
	case <-finished:
		if err != nil {
			return Warnings(allWarnings), err
		}
		if job.Failed() {
			return Warnings(allWarnings), JobFailedError{
				JobGUID: originalJobGUID,
				Message: job.Error,
			}
		}
	}
	return Warnings(allWarnings), nil
}
