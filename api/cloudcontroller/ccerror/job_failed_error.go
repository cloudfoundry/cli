package ccerror

import "fmt"

// JobFailedError represents a failed Cloud Controller Job. It wraps the error
// returned back from the Cloud Controller.
type JobFailedError struct {
	JobGUID string
	Message string
}

func (e JobFailedError) Error() string {
	return fmt.Sprintf("Job (%s) failed: %s", e.JobGUID, e.Message)
}
