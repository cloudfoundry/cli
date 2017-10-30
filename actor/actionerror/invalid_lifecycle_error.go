package actionerror

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// InvalidLifecycleError is returned when the lifecycle specified is neither
// running nor staging.
type InvalidLifecycleError struct {
	Lifecycle ccv2.SecurityGroupLifecycle
}

func (e InvalidLifecycleError) Error() string {
	return fmt.Sprintf("Invalid lifecycle: %s", e.Lifecycle)
}
