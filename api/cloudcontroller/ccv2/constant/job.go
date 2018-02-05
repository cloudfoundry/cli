package constant

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
