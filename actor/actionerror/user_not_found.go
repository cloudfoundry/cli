package actionerror

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// UserNotFoundError is an error wrapper that represents the case
// when the user is not found in UAA.
type UserNotFoundError struct {
	Username string
	Origin   string
}

// Error method to display the error message.
func (e UserNotFoundError) Error() string {
	if e.Origin != "" && e.Origin != constant.DefaultOriginUaa {
		return fmt.Sprintf("User '%s' with origin '%s' does not exist.", e.Username, e.Origin)
	}

	return fmt.Sprintf("User '%s' does not exist.", e.Username)
}
