package actionerror

import (
	"fmt"

	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv2/constant"
)

// SecurityGroupNotBoundError is returned when a requested security group is
// not bound in the requested lifecycle phase to the requested space.
type SecurityGroupNotBoundError struct {
	Lifecycle constant.SecurityGroupLifecycle
	Name      string
}

func (e SecurityGroupNotBoundError) Error() string {
	return fmt.Sprintf("Security group %s not bound to this space for lifecycle phase %s.", e.Name, e.Lifecycle)
}
