package constant

// JobState is the current state of a job.
type JobState string

const (
	// JobFailed is when the job is no longer running due to a failure.
	JobFailed JobState = "FAILED"
	// JobComplete is when the job is no longer and it was successful.
	JobComplete JobState = "COMPLETE"
	// JobProcessing is when the job is waiting to be run.
	JobProcessing JobState = "PROCESSING"
)
