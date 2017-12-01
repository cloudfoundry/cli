package ccv2

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

//go:generate counterfeiter . Reader

// Reader is an io.Reader.
type Reader interface {
	io.Reader
}

// JobStatus is the current state of a job.
type JobStatus string

const (
	// JobStatusFailed is when the job is no longer running due to a failure.
	JobStatusFailed JobStatus = "failed"

	// JobStatusFinished is when the job is no longer and it was successful.
	JobStatusFinished JobStatus = "finished"

	// JobStatusQueued is when the job is waiting to be run.
	JobStatusQueued JobStatus = "queued"

	// JobStatusRunning is when the job is running.
	JobStatusRunning JobStatus = "running"
)

// Job represents a Cloud Controller Job.
type Job struct {
	Error        string
	ErrorDetails struct {
		Description string
	}
	GUID   string
	Status JobStatus
}

// UnmarshalJSON helps unmarshal a Cloud Controller Job response.
func (job *Job) UnmarshalJSON(data []byte) error {
	var ccJob struct {
		Entity struct {
			Error        string `json:"error"`
			ErrorDetails struct {
				Description string `json:"description"`
			} `json:"error_details"`
			GUID   string `json:"guid"`
			Status string `json:"status"`
		} `json:"entity"`
		Metadata internal.Metadata `json:"metadata"`
	}
	if err := json.Unmarshal(data, &ccJob); err != nil {
		return err
	}

	job.Error = ccJob.Entity.Error
	job.ErrorDetails.Description = ccJob.Entity.ErrorDetails.Description
	job.GUID = ccJob.Entity.GUID
	job.Status = JobStatus(ccJob.Entity.Status)
	return nil
}

// Finished returns true when the job has completed successfully.
func (job Job) Finished() bool {
	return job.Status == JobStatusFinished
}

// Failed returns true when the job has completed with an error/failure.
func (job Job) Failed() bool {
	return job.Status == JobStatusFailed
}

// GetJob returns a job for the provided GUID.
func (client *Client) GetJob(jobGUID string) (Job, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetJobRequest,
		URIParams:   Params{"job_guid": jobGUID},
	})
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
func (client *Client) PollJob(job Job) (Warnings, error) {
	originalJobGUID := job.GUID

	var (
		err         error
		warnings    Warnings
		allWarnings Warnings
	)

	startTime := time.Now()
	for time.Now().Sub(startTime) < client.jobPollingTimeout {
		job, warnings, err = client.GetJob(job.GUID)
		allWarnings = append(allWarnings, Warnings(warnings)...)
		if err != nil {
			return allWarnings, err
		}

		if job.Failed() {
			return allWarnings, ccerror.JobFailedError{
				JobGUID: originalJobGUID,
				Message: job.ErrorDetails.Description,
			}
		}

		if job.Finished() {
			return allWarnings, nil
		}

		time.Sleep(client.jobPollingInterval)
	}

	return allWarnings, ccerror.JobTimeoutError{
		JobGUID: originalJobGUID,
		Timeout: client.jobPollingTimeout,
	}
}

// UploadApplicationPackage uploads the newResources and a list of existing
// resources to the cloud controller. A job that combines the requested/newly
// uploaded bits is returned. The function will act differently given the
// following Readers:
//   io.ReadSeeker: Will function properly on retry.
//   io.Reader: Will return a ccerror.PipeSeekError on retry.
//   nil: Will not add the "application" section to the request.
//   newResourcesLength is ignored in this case.
func (client *Client) UploadApplicationPackage(appGUID string, existingResources []Resource, newResources Reader, newResourcesLength int64) (Job, Warnings, error) {
	if existingResources == nil {
		return Job{}, nil, ccerror.NilObjectError{Object: "existingResources"}
	}

	if newResources == nil {
		return client.uploadExistingResourcesOnly(appGUID, existingResources)
	}

	return client.uploadNewAndExistingResources(appGUID, existingResources, newResources, newResourcesLength)
}

func (client *Client) UploadDroplet(appGUID string, droplet io.Reader, dropletLength int64) (Job, Warnings, error) {
	contentLength, err := client.calculateDropletRequestSize(dropletLength)
	if err != nil {
		return Job{}, nil, err
	}

	contentType, body, writeErrors := client.createMultipartBodyAndHeaderForDroplet(droplet)

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutDropletRequest,
		URIParams:   Params{"app_guid": appGUID},
		Body:        body,
	})
	if err != nil {
		return Job{}, nil, err
	}

	request.Header.Set("Content-Type", contentType)
	request.ContentLength = contentLength

	return client.uploadAsynchronously(request, writeErrors)
}

func (*Client) createMultipartBodyAndHeaderForAppBits(existingResources []Resource, newResources io.Reader, newResourcesLength int64) (string, io.ReadSeeker, <-chan error) {
	writerOutput, writerInput := cloudcontroller.NewPipeBomb()
	form := multipart.NewWriter(writerInput)

	writeErrors := make(chan error)

	go func() {
		defer close(writeErrors)
		defer writerInput.Close()

		jsonResources, err := json.Marshal(existingResources)
		if err != nil {
			writeErrors <- err
			return
		}

		err = form.WriteField("resources", string(jsonResources))
		if err != nil {
			writeErrors <- err
			return
		}

		writer, err := form.CreateFormFile("application", "application.zip")
		if err != nil {
			writeErrors <- err
			return
		}

		if newResourcesLength != 0 {
			_, err = io.Copy(writer, newResources)
			if err != nil {
				writeErrors <- err
				return
			}
		}

		err = form.Close()
		if err != nil {
			writeErrors <- err
		}
	}()

	return form.FormDataContentType(), writerOutput, writeErrors
}

func (*Client) createMultipartBodyAndHeaderForDroplet(droplet io.Reader) (string, io.ReadSeeker, <-chan error) {
	writerOutput, writerInput := cloudcontroller.NewPipeBomb()
	form := multipart.NewWriter(writerInput)

	writeErrors := make(chan error)

	go func() {
		defer close(writeErrors)
		defer writerInput.Close()

		writer, err := form.CreateFormFile("droplet", "droplet.tgz")
		if err != nil {
			writeErrors <- err
			return
		}

		_, err = io.Copy(writer, droplet)
		if err != nil {
			writeErrors <- err
			return
		}

		err = form.Close()
		if err != nil {
			writeErrors <- err
		}
	}()

	return form.FormDataContentType(), writerOutput, writeErrors
}

func (*Client) calculateAppBitsRequestSize(existingResources []Resource, newResourcesLength int64) (int64, error) {
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	jsonResources, err := json.Marshal(existingResources)
	if err != nil {
		return 0, err
	}
	err = form.WriteField("resources", string(jsonResources))
	if err != nil {
		return 0, err
	}
	_, err = form.CreateFormFile("application", "application.zip")
	if err != nil {
		return 0, err
	}
	err = form.Close()
	if err != nil {
		return 0, err
	}

	return int64(body.Len()) + newResourcesLength, nil
}

func (*Client) calculateDropletRequestSize(dropletSize int64) (int64, error) {
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	_, err := form.CreateFormFile("droplet", "droplet.tgz")
	if err != nil {
		return 0, err
	}

	err = form.Close()
	if err != nil {
		return 0, err
	}

	return int64(body.Len()) + dropletSize, nil
}

func (client *Client) uploadExistingResourcesOnly(appGUID string, existingResources []Resource) (Job, Warnings, error) {
	jsonResources, err := json.Marshal(existingResources)
	if err != nil {
		return Job{}, nil, err
	}

	body := bytes.NewBuffer(nil)
	form := multipart.NewWriter(body)
	err = form.WriteField("resources", string(jsonResources))
	if err != nil {
		return Job{}, nil, err
	}

	err = form.Close()
	if err != nil {
		return Job{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutAppBitsRequest,
		URIParams:   Params{"app_guid": appGUID},
		Query: url.Values{
			"async": {"true"},
		},
		Body: bytes.NewReader(body.Bytes()),
	})
	if err != nil {
		return Job{}, nil, err
	}

	request.Header.Set("Content-Type", form.FormDataContentType())

	var job Job
	response := cloudcontroller.Response{
		Result: &job,
	}

	err = client.connection.Make(request, &response)
	return job, response.Warnings, err
}

func (client *Client) uploadAsynchronously(request *cloudcontroller.Request, writeErrors <-chan error) (Job, Warnings, error) {
	var job Job
	response := cloudcontroller.Response{
		Result: &job,
	}

	httpErrors := make(chan error)

	go func() {
		defer close(httpErrors)

		err := client.connection.Make(request, &response)
		if err != nil {
			httpErrors <- err
		}
	}()

	// The following section makes the following assumptions:
	// 1) If an error occurs during file reading, an EOF is sent to the request
	// object. Thus ending the request transfer.
	// 2) If an error occurs during request transfer, an EOF is sent to the pipe.
	// Thus ending the writing routine.
	var firstError error
	var writeClosed, httpClosed bool

	for {
		select {
		case writeErr, ok := <-writeErrors:
			if !ok {
				writeClosed = true
				break // for select
			}
			if firstError == nil {
				firstError = writeErr
			}
		case httpErr, ok := <-httpErrors:
			if !ok {
				httpClosed = true
				break // for select
			}
			if firstError == nil {
				firstError = httpErr
			}
		}

		if writeClosed && httpClosed {
			break // for for
		}
	}

	return job, response.Warnings, firstError
}

func (client *Client) uploadNewAndExistingResources(appGUID string, existingResources []Resource, newResources Reader, newResourcesLength int64) (Job, Warnings, error) {
	contentLength, err := client.calculateAppBitsRequestSize(existingResources, newResourcesLength)
	if err != nil {
		return Job{}, nil, err
	}

	contentType, body, writeErrors := client.createMultipartBodyAndHeaderForAppBits(existingResources, newResources, newResourcesLength)

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutAppBitsRequest,
		URIParams:   Params{"app_guid": appGUID},
		Query: url.Values{
			"async": {"true"},
		},
		Body: body,
	})
	if err != nil {
		return Job{}, nil, err
	}

	request.Header.Set("Content-Type", contentType)
	request.ContentLength = contentLength

	return client.uploadAsynchronously(request, writeErrors)
}
