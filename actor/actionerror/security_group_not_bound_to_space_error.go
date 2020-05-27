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
	Space     string
}

func (e SecurityGroupNotBoundToSpaceError) Error() string {
	return fmt.Sprintf("Security group %s not bound to space %s for lifecycle phase '%s'.", e.Name, e.Space, e.Lifecycle)
}
