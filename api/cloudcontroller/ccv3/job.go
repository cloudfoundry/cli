package ccv3

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// ErrorDetails provides information regarding a job's error.
type ErrorDetails struct {
	// Code is a numeric code for this error.
	Code int `json:"code"`
	// Detail is a verbose description of the error.
	Detail string `json:"detail"`
	// Title is a short description of the error.
	Title string `json:"title"`
}

// Job represents a Cloud Controller Job.
type Job struct {
	// Errors is a list of errors that occurred while processing the job.
	Errors []ErrorDetails `json:"errors"`
	// GUID is a unique identifier for the job.
	GUID string `json:"guid"`
	// State is the state of the job.
	State constant.JobState `json:"state"`
}

// HasFailed returns true when the job has completed with an error/failure.
func (job Job) HasFailed() bool {
	return job.State == constant.JobFailed
}

// IsComplete returns true when the job has completed successfully.
func (job Job) IsComplete() bool {
	return job.State == constant.JobComplete
}

// GetJob returns a job for the provided GUID.
func (client *Client) GetJob(jobURL JobURL) (Job, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{URL: string(jobURL)})
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
func (client *Client) PollJob(jobURL JobURL) (Warnings, error) {
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

		if job.HasFailed() {
			return allWarnings, ccerror.JobFailedError{
				JobGUID: job.GUID,
				Message: job.Errors[0].Detail,
			}
		}

		if job.IsComplete() {
			return allWarnings, nil
		}

		time.Sleep(client.jobPollingInterval)
	}

	return allWarnings, ccerror.JobTimeoutError{
		JobGUID: job.GUID,
		Timeout: client.jobPollingTimeout,
	}
}
