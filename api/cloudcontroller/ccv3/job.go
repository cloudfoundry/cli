package ccv3

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// Job represents a Cloud Controller Job.
type Job struct {
	// RawErrors is a list of errors that occurred while processing the job.
	RawErrors []JobErrorDetails `json:"errors"`
	// GUID is a unique identifier for the job.
	GUID string `json:"guid"`
	// State is the state of the job.
	State constant.JobState `json:"state"`
	// Warnings are the warnings emitted by the job during its processing.
	Warnings []jobWarning `json:"warnings"`
}

// Errors returns back a list of
func (job Job) Errors() []error {
	var errs []error
	for _, errDetails := range job.RawErrors {
		switch errDetails.Code {
		case constant.JobErrorCodeBuildpackAlreadyExistsForStack:
			errs = append(errs, ccerror.BuildpackAlreadyExistsForStackError{Message: errDetails.Detail})
		case constant.JobErrorCodeBuildpackInvalid:
			errs = append(errs, ccerror.BuildpackInvalidError{Message: errDetails.Detail})
		case constant.JobErrorCodeBuildpackStacksDontMatch:
			errs = append(errs, ccerror.BuildpackStacksDontMatchError{Message: errDetails.Detail})
		case constant.JobErrorCodeBuildpackStackDoesNotExist:
			errs = append(errs, ccerror.BuildpackStackDoesNotExistError{Message: errDetails.Detail})
		case constant.JobErrorCodeBuildpackZipInvalid:
			errs = append(errs, ccerror.BuildpackZipInvalidError{Message: errDetails.Detail})
		default:
			errs = append(errs, ccerror.V3JobFailedError{
				JobGUID: job.GUID,
				Code:    errDetails.Code,
				Detail:  errDetails.Detail,
				Title:   errDetails.Title,
			})
		}
	}
	return errs
}

// HasFailed returns true when the job has completed with an error/failure.
func (job Job) HasFailed() bool {
	return job.State == constant.JobFailed
}

// IsComplete returns true when the job has completed successfully.
func (job Job) IsComplete() bool {
	return job.State == constant.JobComplete
}

type jobWarning struct {
	Detail string `json:"detail"`
}

// JobErrorDetails provides information regarding a job's error.
type JobErrorDetails struct {
	// Code is a numeric code for this error.
	Code constant.JobErrorCode `json:"code"`
	// Detail is a verbose description of the error.
	Detail string `json:"detail"`
	// Title is a short description of the error.
	Title string `json:"title"`
}

// GetJob returns a job for the provided GUID.
func (client *Client) GetJob(jobURL JobURL) (Job, Warnings, error) {
	var responseBody Job

	_, warnings, err := client.MakeRequest(RequestParams{
		URL:          string(jobURL),
		ResponseBody: &responseBody,
	})

	for _, jobWarning := range responseBody.Warnings {
		warnings = append(warnings, jobWarning.Detail)
	}

	return responseBody, warnings, err
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

	factor := client.jobPollingBackoffFactor
	backoff :=  client.jobPollingInterval

	startTime := client.clock.Now()
	for client.clock.Now().Sub(startTime) < client.jobPollingTimeout {
		job, warnings, err = client.GetJob(jobURL)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return allWarnings, err
		}

		if job.HasFailed() {
			firstError := job.Errors()[0]
			return allWarnings, firstError
		}

		if job.IsComplete() {
			return allWarnings, nil
		}

		backoff = time.Duration(backoff * factor) / time.Second
		time.Sleep(backoff)
	}

	return allWarnings, ccerror.JobTimeoutError{
		JobGUID: job.GUID,
		Timeout: client.jobPollingTimeout,
	}
}
