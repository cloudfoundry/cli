package ccerror

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// UserNotFoundError is returned when a role does not exist.
type UserNotFoundError struct {
	Username string
	Origin   string
}

func (e UserNotFoundError) Error() string {
	if e.Origin != "" && e.Origin != constant.DefaultOriginUaa {
		return fmt.Sprintf("User '%s' with origin '%s' does not exist.", e.Username, e.Origin)
	}

	return fmt.Sprintf("User '%s' does not exist.", e.Username)
}
