package actionerror

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// SecurityGroupNotBoundToSpaceError is returned when a requested security group is
// not bound in the requested lifecycle phase to the requested space.
type SecurityGroupNotBoundToSpaceError struct {
	Lifecycle constant.SecurityGroupLifecycle
	Name      string
}

func (e SecurityGroupNotBoundToSpaceError) Error() string {
	return fmt.Sprintf("Security group '%s' not bound to this space for the %s lifecycle.", e.Name, e.Lifecycle)
}
