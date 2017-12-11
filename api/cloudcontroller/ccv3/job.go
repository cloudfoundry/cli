package ccv3

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type ErrorDetails struct {
	Detail string `json:"detail"`
	Title  string `json:"title"`
	Code   int    `json:"code"`
}

// Job represents a Cloud Controller Job.
type Job struct {
	Errors []ErrorDetails    `json:"errors"`
	GUID   string            `json:"guid"`
	State  constant.JobState `json:"state"`
}

// Complete returns true when the job has completed successfully.
func (job Job) Complete() bool {
	return job.State == constant.JobComplete
}

// Failed returns true when the job has completed with an error/failure.
func (job Job) Failed() bool {
	return job.State == constant.JobFailed
}

// GetJob returns a job for the provided GUID.
func (client *Client) GetJob(jobURL string) (Job, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{URL: jobURL})
	if err != nil {
		return Job{}, nil, err
	}

	var job Job
	response := cloudcontroller.Response{
		Result: &job,
	}

	err = client.connection.Make(request, &response)
	return job, response.Warnings, err
}

// PollJob will keep polling the given job until the job has terminated, an
// error is encountered, or config.OverallPollingTimeout is reached. In the
// last case, a JobTimeoutError is returned.
func (client *Client) PollJob(jobURL string) (Warnings, error) {
	var (
		err         error
		warnings    Warnings
		allWarnings Warnings
		job         Job
	)

	startTime := time.Now()
	for time.Now().Sub(startTime) < client.jobPollingTimeout {
		job, warnings, err = client.GetJob(jobURL)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return allWarnings, err
		}

		if job.Failed() {
			return allWarnings, ccerror.JobFailedError{
				JobGUID: job.GUID,
				Message: job.Errors[0].Detail,
			}
		}

		if job.Complete() {
			return allWarnings, nil
		}

		time.Sleep(client.jobPollingInterval)
	}

	return allWarnings, ccerror.JobTimeoutError{
		JobGUID: job.GUID,
		Timeout: client.jobPollingTimeout,
	}
}
