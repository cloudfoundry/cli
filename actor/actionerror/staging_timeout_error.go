package actionerror

import (
	"fmt"
	"time"
)

// StagingTimeoutError is returned when staging timeout is reached waiting for
// an application to stage.
type StagingTimeoutError struct {
	AppName string
	Timeout time.Duration
}

func (e StagingTimeoutError) Error() string {
	return fmt.Sprintf("Timed out waiting for application '%s' to stage", e.AppName)
}
