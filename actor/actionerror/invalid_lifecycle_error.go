package actionerror

import (
	"fmt"
)

// InvalidLifecycleError is returned when the lifecycle specified is neither
// running nor staging.
type InvalidLifecycleError struct {
	Lifecycle string
}

func (e InvalidLifecycleError) Error() string {
	return fmt.Sprintf("Invalid lifecycle: %s", e.Lifecycle)
}
