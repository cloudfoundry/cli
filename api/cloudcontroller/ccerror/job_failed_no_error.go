package ccerror

import (
	"fmt"
)

// V3JobFailedError represents a failed Cloud Controller Job. It wraps the error
// returned back from the Cloud Controller.
type JobFailedNoErrorError struct {
	JobGUID string
}

func (e JobFailedNoErrorError) Error() string {
	return fmt.Sprintf("Job (%s) failed with no error. This is unexpected, contact your operator for details.", e.JobGUID)
}
