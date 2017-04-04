package ccerror

import (
	"fmt"
	"time"
)

// JobTimeoutError is returned from PollJob when the OverallPollingTimeout has
// been reached.
type JobTimeoutError struct {
	JobGUID string
	Timeout time.Duration
}

func (e JobTimeoutError) Error() string {
	return fmt.Sprintf("Job (%s) polling has reached the maximum timeout of %s seconds", e.JobGUID, e.Timeout)
}
