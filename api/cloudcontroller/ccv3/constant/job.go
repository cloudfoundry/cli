package constant

// JobState is the current state of a job.
type JobState string

const (
	// JobComplete is when the job is no longer and it was successful.
	JobComplete JobState = "COMPLETE"
	// JobFailed is when the job is no longer running due to a failure.
	JobFailed JobState = "FAILED"
	// JobProcessing is when the job is waiting to be run.
	JobProcessing JobState = "PROCESSING"
)

// JobErrorCode is the numeric code for a particular error.
type JobErrorCode int64

const (
	JobErrorCodeBuildpackAlreadyExistsForStack JobErrorCode = 290000
	JobErrorCodeBuildpackInvalid               JobErrorCode = 290003
	JobErrorCodeBuildpackStacksDontMatch       JobErrorCode = 390011
	JobErrorCodeBuildpackStackDoesNotExist     JobErrorCode = 390012
	JobErrorCodeBuildpackZipInvalid            JobErrorCode = 390013
)
